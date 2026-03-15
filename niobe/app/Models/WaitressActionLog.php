<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Relations\BelongsTo;

class WaitressActionLog extends Model
{
    protected $table = 'waitress_action_logs';

    protected $fillable = [
        'waitress_id',
        'tool_name',
        'tool_type',
        'display_name',
        'target',
        'status',
        'arguments',
        'result',
        'error_message',
        'queued_at',
        'started_at',
        'completed_at',
    ];

    protected function casts(): array
    {
        return [
            'arguments' => 'array',
            'result' => 'array',
            'queued_at' => 'datetime',
            'started_at' => 'datetime',
            'completed_at' => 'datetime',
        ];
    }

    public function waitress(): BelongsTo
    {
        return $this->belongsTo(Waitress::class);
    }
}
