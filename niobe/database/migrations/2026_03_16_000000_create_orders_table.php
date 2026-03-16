<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::create('orders', function (Blueprint $table) {
            $table->id();
            $table->foreignId('waitress_id')->constrained()->cascadeOnDelete();
            $table->foreignId('waitress_action_log_id')->nullable()->constrained('waitress_action_logs')->nullOnDelete();
            $table->text('order_summary');
            $table->string('sent_to'); // display name of the tool that ran (e.g. "Place order")
            $table->timestamp('sent_at');
            $table->string('table_number')->nullable(); // optional: which table the order is for
            $table->timestamps();
        });

        Schema::table('orders', function (Blueprint $table) {
            $table->index(['waitress_id', 'sent_at']);
        });
    }

    public function down(): void
    {
        Schema::dropIfExists('orders');
    }
};
