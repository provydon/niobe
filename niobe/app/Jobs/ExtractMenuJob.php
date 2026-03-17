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

    /**
     * @return array<int, UploadedFile>
     */
    private function buildUploadedFiles(): array
    {
        $out = [];
        $disk = Storage::disk();

        foreach ($this->storagePaths as $relativePath) {
            $relativePath = is_string($relativePath) ? trim($relativePath) : '';
            if ($relativePath === '') {
                continue;
            }
            if (! $disk->exists($relativePath)) {
                continue;
            }

            $content = $disk->get($relativePath);
            $tmpPath = (string) tempnam(sys_get_temp_dir(), 'menu-extract-');
            file_put_contents($tmpPath, $content);

            $name = basename($relativePath);
            $mime = $disk->mimeType($relativePath) ?: 'image/jpeg';

            $out[] = new UploadedFile(
                $tmpPath,
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
            $path = is_string($path) ? trim($path) : '';
            if ($path !== '') {
                Storage::delete($path);
            }
        }
    }

    public function handle(MenuExtractionService $menuExtractor): void
    {
        $waitress = Waitress::find($this->waitressId);

        if (! $waitress) {
            $this->deleteStoredFiles();

            return;
        }

        $files = $this->buildUploadedFiles();

        try {
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
        } finally {
            foreach ($files as $file) {
                $path = $file->getRealPath();
                if ($path !== false && is_file($path) && str_starts_with($path, sys_get_temp_dir())) {
                    @unlink($path);
                }
            }
        }
    }
}
