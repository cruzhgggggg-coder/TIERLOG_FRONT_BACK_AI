<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::create('feedback_items', function (Blueprint $table) {
            $table->id();
            $table->foreignId('consultation_log_id')->constrained()->onDelete('cascade');
            $table->text('content');
            $table->enum('category', ['Minor', 'Major']);
            $table->enum('status', ['Fixed', 'Pending'])->default('Pending');
            $table->timestamps();
            $table->softDeletes();
        });
    }

    public function down(): void
    {
        Schema::dropIfExists('feedback_items');
    }
};
