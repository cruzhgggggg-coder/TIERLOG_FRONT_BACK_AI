<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Attributes\Fillable;
use Illuminate\Database\Eloquent\Relations\BelongsTo;

#[Fillable(['consultation_log_id', 'content', 'category', 'status'])]
class FeedbackItem extends Model
{
    public function consultationLog(): BelongsTo
    {
        return $this->belongsTo(ConsultationLog::class);
    }
}
