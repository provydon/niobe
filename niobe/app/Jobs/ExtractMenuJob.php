<?php

namespace App\Jobs;

use App\Models\Waitress;
use App\Services\MenuExtractionService;
use Illuminate\Contracts\Queue\ShouldQueue;
use Illuminate\Foundation\Queue\Queueable;
use Illuminate\Http\UploadedFile;
use Illuminate\Support\Facades\Storage;

class ExtractMenuJob implements ShouldQueue
{
    use Queueable;

    /**
     * @param  array<int, string>  $storagePaths  Relative paths under storage (e.g. waitresses/1/menu-extraction/file.jpg)
     */
    public function __construct(
        public int $waitressId,
        public array $storagePaths,
    ) {}

    public function handle(MenuExtractionService $menuExtractor): void
    {
        $waitress = Waitress::find($this->waitressId);

        if (! $waitress) {
            $this->deleteStoredFiles();

            return;
        }

        $files = $this->buildUploadedFiles();

        if (! empty($files)) {
            $menu = $menuExtractor->extractMenuFromFiles($files);
            $items = $menu['items'] ?? [];
            foreach (array_values($items) as $position => $item) {
                $waitress->menuItems()->create([
                    'name' => $item['name'] ?? '',
                    'category' => $item['category'] ?? 'Other',
                    'unit_price' => (float) ($item['unit_price'] ?? 0),
                    'position' => $position,
                ]);
            }
            $updates = ['menu_image_paths' => $this->storagePaths];
            if (array_key_exists('currency', $menu) && $menu['currency'] !== null && $menu['currency'] !== '') {
                $updates['menu_currency'] = $menu['currency'];
            }
            $waitress->update($updates);
        }
    }

    /**
     * @return array<int, UploadedFile>
     */
    private function buildUploadedFiles(): array
    {
        $out = [];

        foreach ($this->storagePaths as $relativePath) {
            $fullPath = Storage::path($relativePath);

            if (! is_file($fullPath) || ! is_readable($fullPath)) {
                continue;
            }

            $name = basename($relativePath);
            $mime = @mime_content_type($fullPath) ?: 'image/jpeg';

            $out[] = new UploadedFile(
                $fullPath,
                $name,
                $mime,
                \UPLOAD_ERR_OK,
                true
            );
        }

        return $out;
    }

    private function deleteStoredFiles(): void
    {
        foreach ($this->storagePaths as $path) {
            Storage::disk('local')->delete($path);
        }
    }
}
