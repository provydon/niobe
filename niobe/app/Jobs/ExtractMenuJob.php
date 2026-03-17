<?php

namespace App\Jobs;

use App\Models\Waitress;
use App\Services\MenuExtractionService;
use Illuminate\Contracts\Queue\ShouldQueue;
use Illuminate\Foundation\Queue\Queueable;
use Illuminate\Http\UploadedFile;
use Illuminate\Support\Facades\Log;
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
                Log::debug('ExtractMenuJob skipping empty path', ['waitress_id' => $this->waitressId]);

                continue;
            }
            if (! $disk->exists($relativePath)) {
                Log::warning('ExtractMenuJob path not found on disk', ['waitress_id' => $this->waitressId, 'path' => $relativePath]);

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

        Log::info('ExtractMenuJob built uploaded files', ['waitress_id' => $this->waitressId, 'paths_given' => count($this->storagePaths), 'files_ready' => count($out)]);

        return $out;
    }

    private function deleteStoredFiles(): void
    {
        Log::info('ExtractMenuJob deleting stored files (waitress missing)', ['waitress_id' => $this->waitressId, 'paths' => $this->storagePaths]);
        foreach ($this->storagePaths as $path) {
            $path = is_string($path) ? trim($path) : '';
            if ($path !== '') {
                Storage::delete($path);
            }
        }
    }

    public function handle(MenuExtractionService $menuExtractor): void
    {
        Log::info('ExtractMenuJob started', ['waitress_id' => $this->waitressId, 'storage_paths' => $this->storagePaths]);

        $waitress = Waitress::find($this->waitressId);

        if (! $waitress) {
            Log::warning('ExtractMenuJob waitress not found', ['waitress_id' => $this->waitressId]);
            $this->deleteStoredFiles();

            return;
        }

        $files = $this->buildUploadedFiles();

        try {
            if (! empty($files)) {
                Log::info('ExtractMenuJob running extraction', ['waitress_id' => $this->waitressId, 'file_count' => count($files)]);
                $menu = $menuExtractor->extractMenuFromFiles($files);
                $items = $menu['items'] ?? [];
                $currency = $menu['currency'] ?? null;
                Log::info('ExtractMenuJob extraction result', ['waitress_id' => $this->waitressId, 'item_count' => count($items), 'currency' => $currency]);

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
                Log::info('ExtractMenuJob completed', ['waitress_id' => $this->waitressId, 'menu_items_created' => count($items)]);
            } else {
                Log::info('ExtractMenuJob no files to extract', ['waitress_id' => $this->waitressId]);
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
