<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Relations\BelongsTo;
use Illuminate\Database\Eloquent\Relations\HasMany;
use Illuminate\Support\Str;

class Waitress extends Model
{
    protected $table = 'waitresses';

    protected $appends = [
        'share_url',
        'talk_url',
    ];

    protected $fillable = [
        'name',
        'slug',
        'context',
        'extracted_context',
        'menu_image_paths',
        'menu_currency',
        'tools',
    ];

    protected function casts(): array
    {
        return [
            'tools' => 'array',
            'extracted_context' => 'array',
            'menu_image_paths' => 'array',
        ];
    }

    public function user(): BelongsTo
    {
        return $this->belongsTo(User::class);
    }

    public function actionLogs(): HasMany
    {
        return $this->hasMany(WaitressActionLog::class);
    }

    public function menuItems(): HasMany
    {
        return $this->hasMany(MenuItem::class)->orderBy('position')->orderBy('id');
    }

    protected static function booted(): void
    {
        static::creating(function (Waitress $waitress) {
            if (empty($waitress->slug)) {
                $waitress->slug = Str::slug($waitress->name);
            }
            $waitress->slug = static::uniqueSlug($waitress->slug, $waitress->id ?? 0);
        });
    }

    protected static function uniqueSlug(string $base, int $excludeId = 0): string
    {
        $slug = $base;
        $n = 0;
        while (true) {
            $q = static::where('slug', $slug)->when($excludeId > 0, fn ($q) => $q->where('id', '!=', $excludeId));
            if (! $q->exists()) {
                return $slug;
            }
            $slug = $base.'-'.(++$n);
        }
    }

    public function getShareUrlAttribute(): string
    {
        return route('niobe.show', ['slug' => $this->slug]);
    }

    public function getTalkUrlAttribute(): string
    {
        return route('niobe.talk', ['slug' => $this->slug]);
    }

    /** @return array<int, array{type: string, name: string, url?: string}> */
    public function getToolsList(): array
    {
        return $this->tools ?? [];
    }

    /**
     * Full context for the voice agent: menu (if set), written context, and structured data from uploaded files.
     */
    public function getFullContextAttribute(): string
    {
        $parts = [];

        $menuItems = $this->menuItems;
        if ($menuItems->isNotEmpty()) {
            $lines = ['Menu (items, category, unit_price):'];
            if (! empty($this->menu_currency)) {
                $lines[] = 'Currency: '.$this->menu_currency;
            }
            foreach ($menuItems as $item) {
                $lines[] = "- {$item->name} | {$item->category} | {$item->unit_price}";
            }
            $parts[] = implode("\n", $lines);
        }

        $parts[] = trim($this->context);

        $extracted = $this->extracted_context ?? [];
        if (! empty($extracted)) {
            $parts[] = "\n\n--- Structured data from uploaded documents ---\n";
            foreach ($extracted as $item) {
                $filename = $item['filename'] ?? 'document';
                $data = $item['data'] ?? null;
                $error = $item['error'] ?? null;
                if ($error) {
                    $parts[] = "({$filename}: {$error})";
                } elseif (is_array($data)) {
                    $parts[] = 'From '.$filename.': '.json_encode($data, JSON_PRETTY_PRINT | JSON_UNESCAPED_SLASHES);
                }
            }
        }

        return implode("\n\n", array_filter($parts));
    }
}
