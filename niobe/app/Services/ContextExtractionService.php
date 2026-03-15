<?php

namespace App\Services;

use Illuminate\Http\UploadedFile;
use Laravel\Ai\Files\Document;
use Laravel\Ai\AnonymousAgent;
use Laravel\Ai\Enums\Lab;

class ContextExtractionService
{
    private const EXTRACT_INSTRUCTIONS = <<<'TEXT'
You are a document analyst. You extract structured data from documents and images and return it as valid JSON only.

Rules:
- For restaurant/food menus: return an object with "menu" containing an array of items. Each item has "name", "category" (e.g. Starters, Mains, Drinks), and "price" (string or number). Include "currency" if visible.
- For company org charts or structure: return an object with "structure" describing departments, roles, and reporting (nested or flat as appropriate).
- For other documents: extract the main information (lists, tables, key-value data) into a clear JSON structure. Use keys like "sections", "items", "data" as appropriate.
- Return ONLY valid JSON. No markdown code fences, no explanation, no text before or after the JSON.
- If the document is unclear or empty, return {"error": "Could not extract structured data"}.
TEXT;

    /**
     * Extract structured JSON from one or more uploaded files using Gemini.
     *
     * @param  array<int, UploadedFile>  $files
     * @return array<int, array{filename: string, data: array<mixed>|null, error?: string}>
     */
    public function extractFromFiles(array $files): array
    {
        $results = [];

        foreach ($files as $file) {
            if (! $file instanceof UploadedFile || ! $file->isValid()) {
                continue;
            }

            $agent = new AnonymousAgent(self::EXTRACT_INSTRUCTIONS, [], []);
            $prompt = 'Extract all structured data from this document or image. Return only valid JSON.';
            $attachments = [$file];

            $results[] = $this->extractAttachment(
                $agent,
                $prompt,
                $attachments,
                $file->getClientOriginalName(),
            );
        }

        return $results;
    }

    /**
     * Extract structured JSON from one or more stored files using Gemini.
     *
     * @param  array<int, array{disk: string, path: string, filename: string}>  $files
     * @return array<int, array{filename: string, data: array<mixed>|null, error?: string}>
     */
    public function extractFromStoredFiles(array $files): array
    {
        $results = [];

        foreach ($files as $file) {
            $agent = new AnonymousAgent(self::EXTRACT_INSTRUCTIONS, [], []);
            $prompt = 'Extract all structured data from this document or image. Return only valid JSON.';
            $attachment = Document::fromStorage($file['path'], $file['disk'])->as($file['filename']);

            $results[] = $this->extractAttachment(
                $agent,
                $prompt,
                [$attachment],
                $file['filename'],
            );
        }

        return $results;
    }

    /**
     * @param  array<int, mixed>  $attachments
     * @return array{filename: string, data: array<mixed>|null, error?: string}
     */
    private function extractAttachment(
        AnonymousAgent $agent,
        string $prompt,
        array $attachments,
        string $filename,
    ): array {
        try {
            $response = $agent->prompt($prompt, $attachments, Lab::Gemini, null, 60);
            $text = trim($response->text);

            // Strip markdown code block if present
            if (preg_match('/^```(?:json)?\s*(.*)\s*```$/s', $text, $m)) {
                $text = trim($m[1]);
            }

            $decoded = json_decode($text, true);
            if (json_last_error() !== JSON_ERROR_NONE) {
                return [
                    'filename' => $filename,
                    'data' => null,
                    'error' => 'Invalid JSON from AI: '.json_last_error_msg(),
                ];
            }

            return [
                'filename' => $filename,
                'data' => $decoded,
            ];
        } catch (\Throwable $e) {
            return [
                'filename' => $filename,
                'data' => null,
                'error' => $e->getMessage(),
            ];
        }
    }
}
