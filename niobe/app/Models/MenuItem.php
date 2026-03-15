<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Relations\BelongsTo;

class MenuItem extends Model
{
    protected $fillable = [
        'waitress_id',
        'name',
        'category',
        'unit_price',
        'position',
    ];

    protected function casts(): array
    {
        return [
            'unit_price' => 'decimal:2',
        ];
    }

    public function waitress(): BelongsTo
    {
        return $this->belongsTo(Waitress::class);
    }
}
