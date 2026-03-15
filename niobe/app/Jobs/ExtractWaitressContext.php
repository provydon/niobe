<?php

namespace App\Jobs;

use App\Models\Waitress;
use App\Services\ContextExtractionService;
use Illuminate\Contracts\Queue\ShouldQueue;
use Illuminate\Foundation\Queue\Queueable;
use Illuminate\Support\Facades\Storage;

class ExtractWaitressContext implements ShouldQueue
{
    use Queueable;

    /**
     * @param  array<int, array{disk: string, path: string, filename: string}>  $files
     */
    public function __construct(
        public int $waitressId,
        public array $files,
    ) {}

    public function handle(ContextExtractionService $extractor): void
    {
        $waitress = Waitress::find($this->waitressId);

        if (! $waitress) {
            $this->deleteStoredFiles();

            return;
        }

        $results = $extractor->extractFromStoredFiles($this->files);

        $waitress->update([
            'extracted_context' => $results,
        ]);

        $this->deleteStoredFiles();
    }

    private function deleteStoredFiles(): void
    {
        foreach ($this->files as $file) {
            Storage::disk($file['disk'])->delete($file['path']);
        }
    }
}
