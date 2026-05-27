<?php

use Illuminate\Support\Facades\Broadcast;

Broadcast::channel('App.Models.User.{id}', function ($user, $id) {
    return (int) $user->id === (int) $id;
});

Broadcast::channel('consultation.{logId}', function ($user, $logId) {
    $log = \App\Models\ConsultationLog::with('student.lecturer')->find($logId);
    if (!$log || !$log->student) {
        return false;
    }
    if ($user->role === 'student') {
        return (int) $user->id === (int) $log->student->user_id;
    }
    if ($user->role === 'lecturer') {
        return $log->student->lecturer && (int) $user->id === (int) $log->student->lecturer->user_id;
    }
    return $user->role === 'admin';
});
