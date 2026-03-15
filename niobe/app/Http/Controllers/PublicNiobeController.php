<?php

namespace App\Http\Controllers;

use App\Models\Waitress;
use Illuminate\Support\Facades\Storage;
use Inertia\Inertia;
use Inertia\Response;
use Symfony\Component\HttpFoundation\Response as SymfonyResponse;

class PublicNiobeController extends Controller
{
    /**
     * Public page for a waitress (Niobe app). Customers land here and click Start to talk to the AI.
     */
    public function show(string $slug): Response
    {
        $waitress = Waitress::where('slug', $slug)->firstOrFail();

        return Inertia::render('Niobe/Show', [
            'niobe' => [
                'name' => $waitress->name,
                'context' => $waitress->context,
                'talk_url' => $waitress->talk_url,
            ],
        ]);
    }

    public function talk(string $slug): Response
    {
        $waitress = Waitress::with('menuItems')->where('slug', $slug)->firstOrFail();

        $menuImagePaths = $waitress->menu_image_paths ?? [];
        $menuImageUrls = [];
        foreach (array_values($menuImagePaths) as $i => $relativePath) {
            $menuImageUrls[] = route('niobe.menu-image', ['slug' => $waitress->slug, 'index' => $i]);
        }

        $menuItems = $waitress->menuItems->map(fn ($item) => [
            'name' => $item->name,
            'category' => $item->category,
            'unit_price' => (string) $item->unit_price,
        ])->all();

        return Inertia::render('Niobe/Talk', [
            'niobe' => [
                'name' => $waitress->name,
                'context' => $waitress->context,
                'menu' => $waitress->full_context,
                'share_url' => $waitress->share_url,
                'menu_items' => $menuItems,
                'menu_image_urls' => $menuImageUrls,
                'menu_currency' => $waitress->menu_currency,
            ],
            'voiceAgentWebsocketUrl' => $this->voiceAgentWebsocketUrl($waitress->slug),
        ]);
    }

    /**
     * Serve a menu image for the public talk page (owner-uploaded preview).
     */
    public function menuImage(string $slug, string $index): SymfonyResponse
    {
        $waitress = Waitress::where('slug', $slug)->firstOrFail();

        $paths = array_values($waitress->menu_image_paths ?? []);
        $i = (int) $index;
        if ($i < 0 || $i >= count($paths)) {
            abort(404);
        }

        $relativePath = $paths[$i];
        if (! Storage::disk('local')->exists($relativePath)) {
            abort(404);
        }

        $fullPath = Storage::disk('local')->path($relativePath);
        $filename = basename($relativePath);

        return response()->file($fullPath, [
            'Content-Disposition' => 'inline; filename="'.addslashes($filename).'"',
        ]);
    }

    private function voiceAgentWebsocketUrl(string $slug): string
    {
        $voiceAgentUrl = rtrim(config('app.voice_agent_url', 'http://localhost:9000'), '/');
        $websocketBase = preg_replace('/^http/i', 'ws', $voiceAgentUrl) ?? $voiceAgentUrl;

        return $websocketBase.'/live?niobe='.urlencode($slug);
    }
}
