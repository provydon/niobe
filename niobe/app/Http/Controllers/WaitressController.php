<?php

namespace App\Http\Controllers;

use App\Jobs\ExtractMenuJob;
use App\Jobs\ExtractWaitressContext;
use App\Models\MenuItem;
use App\Models\Waitress;
use App\Services\ContextExtractionService;
use App\Services\MenuExtractionService;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\RedirectResponse;
use Illuminate\Http\Request;
use Illuminate\Support\Str;
use Illuminate\Validation\Rule;
use Inertia\Inertia;
use Inertia\Response;

class WaitressController extends Controller
{
    private const ACTION_TYPES = [
        'send_email',
        'send_webhook_event',
        'send_whatsapp_message',
    ];

    /**
     * @return array<int, array{value: string, label: string, targetLabel: string, targetPlaceholder: string, hint: string}>
     */
    private function actionTypes(): array
    {
        return [
            [
                'value' => 'send_email',
                'label' => 'Send email',
                'targetLabel' => 'Email address',
                'targetPlaceholder' => 'orders@example.com',
                'hint' => 'Use the inbox or recipient that should get the message.',
            ],
            [
                'value' => 'send_webhook_event',
                'label' => 'Send webhook event',
                'targetLabel' => 'Webhook URL',
                'targetPlaceholder' => 'https://example.com/webhooks/niobe',
                'hint' => 'Niobe will send a payload to this endpoint.',
            ],
            [
                'value' => 'send_whatsapp_message',
                'label' => 'Send WhatsApp message',
                'targetLabel' => 'WhatsApp number',
                'targetPlaceholder' => '+15551234567',
                'hint' => 'Use an international number or WhatsApp destination.',
            ],
        ];
    }

    public function __construct(
        private ContextExtractionService $extractor,
        private MenuExtractionService $menuExtractor,
    ) {}

    public function index(Request $request): Response
    {
        $waitresses = $request->user()
            ->waitresses()
            ->withCount('menuItems')
            ->latest()
            ->paginate(10);

        return Inertia::render('Waitresses/Index', [
            'waitresses' => $waitresses,
        ]);
    }

    public function create(): Response
    {
        return Inertia::render('Waitresses/Create', [
            'actionTypes' => $this->actionTypes(),
        ]);
    }

    /**
     * Extract context from uploaded files via AI. Returns a string to pre-fill the context field.
     */
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

    public function store(Request $request): RedirectResponse|JsonResponse
    {
        $validated =         $request->merge([
            'tables_count' => is_numeric($request->input('tables_count')) ? (int) $request->input('tables_count') : null,
        ]);
        $validated = $request->validate([
            'name' => ['required', 'string', 'max:255'],
            'actions' => ['required', 'array', 'min:1'],
            'actions.*.type' => ['required', 'string', Rule::in(self::ACTION_TYPES)],
            'actions.*.name' => ['required', 'string', 'max:255'],
            'actions.*.target' => ['required', 'string', 'max:2048'],
            'menu_files' => ['required', 'array', 'min:1', 'max:10'],
            'menu_files.*' => ['required', 'file', 'max:10240', 'mimes:jpg,jpeg,png,gif,webp'],
            'tables_count' => ['nullable', 'integer', 'min:0', 'max:9999'],
        ]);

        $waitress = $request->user()->waitresses()->create([
            'name' => $validated['name'],
            'slug' => Str::slug($validated['name']),
            'context' => '',
            'tables_count' => isset($validated['tables_count']) && $validated['tables_count'] > 0 ? (int) $validated['tables_count'] : null,
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
        foreach ($request->file('menu_files') as $file) {
            $paths[] = $file->store("waitresses/{$waitress->id}/menu-extraction", 'local');
        }

        ExtractMenuJob::dispatch($waitress->id, $paths);

        $message = __('Waitress created. Menu is being extracted and will appear shortly.');

        if ($request->expectsJson()) {
            return response()->json([
                'message' => $message,
                'redirect' => route('waitresses.index'),
                'waitress' => [
                    'id' => $waitress->id,
                ],
            ], 201);
        }

        return redirect()->route('waitresses.edit', $waitress)->with('success', $message);
    }

    public function edit(Waitress $waitress): Response
    {
        $this->authorize('update', $waitress);

        $waitress->load('menuItems');

        return Inertia::render('Waitresses/Edit', [
            'waitress' => $waitress,
            'actionTypes' => $this->actionTypes(),
        ]);
    }

    public function update(Request $request, Waitress $waitress): RedirectResponse
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
            'tables_count' => isset($validated['tables_count']) && $validated['tables_count'] > 0 ? (int) $validated['tables_count'] : null,
            'tools' => array_map(
                fn (array $action) => [
                    'type' => $action['type'],
                    'name' => $action['name'],
                    'target' => $action['target'],
                ],
                $validated['actions']
            ),
        ]);

        return redirect()->route('waitresses.edit', $waitress)->with('success', __('Waitress updated.'));
    }

    public function destroy(Waitress $waitress): RedirectResponse
    {
        $this->authorize('delete', $waitress);
        $waitress->delete();

        return redirect()->route('waitresses.index')->with('success', __('Waitress deleted.'));
    }
}
