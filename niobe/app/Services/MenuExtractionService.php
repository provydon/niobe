<?php

namespace App\Services;

use Illuminate\Http\UploadedFile;
use Illuminate\Support\Facades\Log;
use Laravel\Ai\AnonymousAgent;
use Laravel\Ai\Enums\Lab;

class MenuExtractionService
{
    private const INSTRUCTIONS = <<<'TEXT'
You are a menu analyst. You will receive one or more images of restaurant/food menus.

Merge all items from every image into a single menu. Detect the currency used on the menu (one currency for the whole menu). Return ONLY valid JSON in this exact shape:

{
  "currency": "USD",
  "items": [
    {
      "name": "Item name as shown",
      "category": "Category e.g. Starters, Mains, Drinks, Sides",
      "unit_price": 1234
    }
  ]
}

Rules:
- currency: ISO 4217 code (e.g. USD, EUR, GBP, NGN, JPY) or common symbol code. One value for the entire menu. If no currency is visible, use "USD".
- unit_price must be a number (integer or float). If the menu shows "15,610" or "15.61", use 15610 or 15.61 as appropriate (no commas).
- Use a single category per item. Normalize categories (e.g. "Main Course" -> "Mains").
- If price is in a currency like ₦ or $, still use the numeric value only for unit_price and set "currency" accordingly.
- Return ONLY the JSON object. No markdown, no explanation, no text before or after.
- If you cannot read a menu or images are empty, return {"currency": "USD", "items": []}.
TEXT;

    /**
     * Extract and merge menu from uploaded images using Gemini. Returns structured items with category and unit_price.
     *
     * @param  array<int, UploadedFile>  $files
     * @return array{currency: string|null, items: array<int, array{name: string, category: string, unit_price: int|float}>}
     */
    public function extractMenuFromFiles(array $files): array
    {
        $valid = array_values(array_filter($files, fn ($f) => $f instanceof UploadedFile && $f->isValid()));
        if (count($valid) === 0) {
            Log::info('MenuExtractionService: no valid files', ['received' => count($files)]);

            return ['currency' => null, 'items' => []];
        }

        Log::info('MenuExtractionService: calling Gemini', ['file_count' => count($valid)]);

        $agent = new AnonymousAgent(self::INSTRUCTIONS, [], []);
        $prompt = 'Merge all menu items from these images into one list. Return the JSON object only.';
        $attachments = $valid;

        try {
            $response = $agent->prompt($prompt, $attachments, Lab::Gemini, null, 90);
            $text = trim($response->text);

            if (preg_match('/^```(?:json)?\s*(.*)\s*```$/s', $text, $m)) {
                $text = trim($m[1]);
            }

            $decoded = json_decode($text, true);
            if (json_last_error() !== JSON_ERROR_NONE || ! is_array($decoded)) {
                Log::warning('MenuExtractionService: Gemini response not valid JSON', ['json_error' => json_last_error_msg()]);

                return ['currency' => null, 'items' => []];
            }

            $currency = isset($decoded['currency']) ? trim((string) $decoded['currency']) : null;
            if ($currency === '') {
                $currency = null;
            }
            $items = $decoded['items'] ?? [];
            if (! is_array($items)) {
                return ['currency' => $currency, 'items' => []];
            }

            $normalized = [];
            foreach ($items as $row) {
                if (! is_array($row)) {
                    continue;
                }
                $name = isset($row['name']) ? trim((string) $row['name']) : '';
                $category = isset($row['category']) ? trim((string) $row['category']) : 'Other';
                $price = $row['unit_price'] ?? $row['price'] ?? 0;
                if (is_string($price)) {
                    $price = (float) preg_replace('/[^\d.]/', '', $price) ?: 0;
                }
                $normalized[] = [
                    'name' => $name,
                    'category' => $category,
                    'unit_price' => is_numeric($price) ? (float) $price : 0,
                ];
            }

            Log::info('MenuExtractionService: parsed menu', ['item_count' => count($normalized), 'currency' => $currency]);

            return ['currency' => $currency, 'items' => $normalized];
        } catch (\Throwable $e) {
            Log::error('MenuExtractionService: extraction failed', ['exception' => $e->getMessage()]);

            return ['currency' => null, 'items' => []];
        }
    }
}
