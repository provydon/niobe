<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Relations\BelongsTo;

class Order extends Model
{
    protected $table = 'orders';

    protected $fillable = [
        'waitress_id',
        'waitress_action_log_id',
        'order_summary',
        'sent_to',
        'sent_at',
        'table_number',
        'customer_name',
    ];

    protected function casts(): array
    {
        return [
            'sent_at' => 'datetime',
        ];
    }

    public function waitress(): BelongsTo
    {
        return $this->belongsTo(Waitress::class);
    }

    public function waitressActionLog(): BelongsTo
    {
        return $this->belongsTo(WaitressActionLog::class, 'waitress_action_log_id');
    }
}
