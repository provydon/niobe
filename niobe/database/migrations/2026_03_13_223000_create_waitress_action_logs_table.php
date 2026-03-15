<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::create('waitress_action_logs', function (Blueprint $table) {
            $table->id();
            $table->foreignId('waitress_id')->constrained()->cascadeOnDelete();
            $table->string('tool_name');
            $table->string('tool_type');
            $table->string('display_name');
            $table->string('target')->nullable();
            $table->string('status');
            $table->json('arguments')->nullable();
            $table->json('result')->nullable();
            $table->text('error_message')->nullable();
            $table->timestamp('queued_at')->nullable();
            $table->timestamp('started_at')->nullable();
            $table->timestamp('completed_at')->nullable();
            $table->timestamps();
        });
    }

    public function down(): void
    {
        Schema::dropIfExists('waitress_action_logs');
    }
};
