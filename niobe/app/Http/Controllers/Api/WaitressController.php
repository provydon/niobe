<?php

namespace App\Http\Controllers\Api;

use App\Http\Controllers\Controller;
use App\Jobs\ExtractMenuJob;
use App\Models\Waitress;
use App\Services\ContextExtractionService;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\File;
use Illuminate\Support\Facades\Log;
use Illuminate\Support\Facades\Storage;
use Illuminate\Support\Str;
use Illuminate\Validation\Rule;

class WaitressController extends Controller
{
    private const ACTION_TYPES = [
        'send_email',
        'send_webhook_event',
        'send_whatsapp_message',
    ];

    public function __construct(
        private ContextExtractionService $extractor,
    ) {}

    public function index(Request $request): JsonResponse
    {
        $waitresses = $request->user()
            ->waitresses()
            ->withCount('menuItems')
            ->latest()
            ->paginate(10);

        return response()->json($waitresses);
    }

    public function actionTypes(): JsonResponse
    {
        return response()->json([
            'actionTypes' => [
                ['value' => 'send_email', 'label' => 'Send email', 'targetLabel' => 'Email address', 'targetPlaceholder' => 'orders@example.com', 'hint' => 'Use the inbox or recipient that should get the message.'],
                ['value' => 'send_webhook_event', 'label' => 'Send webhook event', 'targetLabel' => 'Webhook URL', 'targetPlaceholder' => 'https://example.com/webhooks/niobe', 'hint' => 'Niobe will send a payload to this endpoint.'],
                ['value' => 'send_whatsapp_message', 'label' => 'Send WhatsApp message', 'targetLabel' => 'WhatsApp number', 'targetPlaceholder' => '+15551234567', 'hint' => 'Use an international number or WhatsApp destination.'],
            ],
        ]);
    }

    public function show(Request $request, Waitress $waitress): JsonResponse
    {
        $this->authorize('update', $waitress);
        $waitress->load('menuItems');

        return response()->json($waitress);
    }

    public function store(Request $request): JsonResponse
    {
        $actionsInput = $request->input('actions');
        if (is_string($actionsInput)) {
            $decoded = json_decode($actionsInput, true);
            $request->merge(['actions' => is_array($decoded) ? $decoded : []]);
        }
        $request->merge([
            'tables_count' => is_numeric($request->input('tables_count')) ? (int) $request->input('tables_count') : null,
        ]);
        $validated = $request->validate([
            'name' => ['required', 'string', 'max:255'],
            'actions' => ['required', 'array', 'min:1'],
            'actions.*.type' => ['required', 'string', Rule::in(self::ACTION_TYPES)],
            'actions.*.name' => ['required', 'string', 'max:255'],
            'actions.*.target' => ['required', 'string', 'max:2048'],
            'menu_files' => ['sometimes', 'nullable', 'array', 'max:10'],
            'menu_files.*' => ['required', 'file', 'max:10240', 'mimes:jpg,jpeg,png,gif,webp'],
            'tables_count' => ['nullable', 'integer', 'min:0', 'max:9999'],
        ]);

        $waitress = $request->user()->waitresses()->create([
            'name' => $validated['name'],
            'slug' => Str::slug($validated['name']),
            'context' => '',
            'tables_count' => isset($validated['tables_count']) && $validated['tables_count'] > 0 ? (int) $validated['tables_count'] : 2,
            'tools' => array_map(
                fn (array $action) => [
                    'type' => $action['type'],
                    'name' => $action['name'],
                    'target' => $action['target'],
                ],
                $validated['actions']
            ),
        ]);

        $paths = [];
        $menuFiles = $request->file('menu_files', []);
        $hasMenuFiles = is_array($menuFiles) && count($menuFiles) > 0;

        Log::info('Menu upload started', ['waitress_id' => $waitress->id, 'user_uploaded_files' => $hasMenuFiles, 'file_count' => $hasMenuFiles ? count($menuFiles) : 0]);

        try {
            if ($hasMenuFiles) {
                $prefix = "waitresses/{$waitress->id}/menu-extraction";
                foreach ($menuFiles as $file) {
                    $path = $file->store($prefix);
                    if (is_string($path) && $path !== '') {
                        $paths[] = $path;
                        Log::info('Menu file stored', ['waitress_id' => $waitress->id, 'path' => $path]);
                    } else {
                        Log::warning('Menu file store() returned no path', ['waitress_id' => $waitress->id, 'original_name' => $file->getClientOriginalName()]);
                        report(new \RuntimeException('Menu file store() returned no path. Check S3/AWS config.'));
                    }
                }
            } else {
                $defaultPath = public_path('menus/jays.jpeg');
                if (File::isFile($defaultPath)) {
                    $path = Storage::putFileAs("waitresses/{$waitress->id}/menu-extraction", new \Illuminate\Http\File($defaultPath), 'jays.jpeg');
                    if (is_string($path) && $path !== '') {
                        $paths[] = $path;
                        Log::info('Default menu image stored', ['waitress_id' => $waitress->id, 'path' => $path]);
                    }
                } else {
                    Log::info('No menu files and no default image', ['waitress_id' => $waitress->id]);
                }
            }
        } catch (\Throwable $e) {
            Log::error('Menu upload failed', ['waitress_id' => $waitress->id, 'exception' => $e->getMessage()]);
            report($e);
            return response()->json(['message' => __('Menu file upload failed.'), 'errors' => ['menu_files' => [__('Menu file upload failed.')]]], 422);
        }

        $paths = array_values(array_filter($paths, fn ($p) => is_string($p) && $p !== ''));
        Log::info('Menu upload paths after filter', ['waitress_id' => $waitress->id, 'paths' => $paths]);

        if ($hasMenuFiles && count($paths) === 0) {
            Log::warning('User provided menu files but none could be stored', ['waitress_id' => $waitress->id]);
            return response()->json(['message' => __('Menu file upload failed.'), 'errors' => ['menu_files' => [__('Menu file upload failed.')]]], 422);
        }
        if (count($paths) > 0) {
            ExtractMenuJob::dispatch($waitress->id, $paths);
            Log::info('ExtractMenuJob dispatched', ['waitress_id' => $waitress->id, 'paths' => $paths]);
        }

        return response()->json([
            'message' => __('Waitress created. Menu is being extracted and will appear shortly.'),
            'waitress' => ['id' => $waitress->id],
        ], 201);
    }

    public function update(Request $request, Waitress $waitress): JsonResponse
    {
        $this->authorize('update', $waitress);

        $request->merge([
            'tables_count' => is_numeric($request->input('tables_count')) ? (int) $request->input('tables_count') : null,
        ]);
        $validated = $request->validate([
            'name' => ['required', 'string', 'max:255'],
            'actions' => ['required', 'array', 'min:1'],
            'actions.*.type' => ['required', 'string', Rule::in(self::ACTION_TYPES)],
            'actions.*.name' => ['required', 'string', 'max:255'],
            'actions.*.target' => ['required', 'string', 'max:2048'],
            'tables_count' => ['nullable', 'integer', 'min:0', 'max:9999'],
        ]);

        $waitress->update([
            'name' => $validated['name'],
            'tables_count' => isset($validated['tables_count']) && $validated['tables_count'] > 0 ? (int) $validated['tables_count'] : 2,
            'tools' => array_map(
                fn (array $action) => [
                    'type' => $action['type'],
                    'name' => $action['name'],
                    'target' => $action['target'],
                ],
                $validated['actions']
            ),
        ]);

        return response()->json(['message' => __('Waitress updated.')]);
    }

    public function destroy(Waitress $waitress): JsonResponse
    {
        $this->authorize('delete', $waitress);
        $waitress->delete();

        return response()->json(['message' => __('Waitress deleted.')]);
    }

    public function extractContext(Request $request): JsonResponse
    {
        $request->validate([
            'context_files' => ['required', 'array', 'max:10'],
            'context_files.*' => ['required', 'file', 'max:10240', 'mimes:jpg,jpeg,png,gif,webp,pdf,doc,docx,txt,csv'],
        ]);

        $results = $this->extractor->extractFromFiles($request->file('context_files'));
        $parts = [];
        foreach ($results as $item) {
            $filename = $item['filename'] ?? 'document';
            $data = $item['data'] ?? null;
            $error = $item['error'] ?? null;
            if ($error) {
                $parts[] = "({$filename}: {$error})";
            } elseif (is_array($data)) {
                $parts[] = "--- {$filename} ---\n".json_encode($data, JSON_PRETTY_PRINT | JSON_UNESCAPED_UNICODE);
            }
        }
        $context = implode("\n\n", $parts);

        return response()->json(['context' => $context]);
    }
}
