<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Attributes\Fillable;
use Illuminate\Database\Eloquent\Relations\BelongsTo;
use Illuminate\Database\Eloquent\Relations\HasMany;

#[Fillable(['student_id', 'lecturer_id', 'transcript_text', 'paper_path', 'processed'])]
class ConsultationLog extends Model
{
    public function student(): BelongsTo
    {
        return $this->belongsTo(Student::class);
    }

    public function feedbackItems(): HasMany
    {
        return $this->hasMany(FeedbackItem::class);
    }
}
