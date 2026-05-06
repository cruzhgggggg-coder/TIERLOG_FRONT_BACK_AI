<?php

use Illuminate\Support\Facades\Route;
use Laravel\Fortify\Features;

use App\Http\Controllers\ConsultationController;

Route::inertia('/', 'Welcome', [
    'canRegister' => Features::enabled(Features::registration()),
])->name('home');

Route::middleware(['auth', 'verified'])->group(function () {
    Route::inertia('dashboard', 'Dashboard')->name('dashboard');
    
    Route::get('consultation', [ConsultationController::class, 'index'])->name('consultation.index');
    Route::get('logs', [ConsultationController::class, 'archive'])->name('consultation.archive');
    Route::post('consultation', [ConsultationController::class, 'store'])->name('consultation.store');
    Route::post('consultation/chat', [ConsultationController::class, 'chat'])->name('consultation.chat');
    Route::get('dashboard/stats', [ConsultationController::class, 'stats'])->name('dashboard.stats');
    Route::put('consultation/feedback/{id}/status', [ConsultationController::class, 'updateFeedbackStatus'])->name('consultation.feedback.status');
    Route::get('lecturer/consultations', [ConsultationController::class, 'lecturerConsultations'])->name('lecturer.consultations');
    Route::get('lecturer/students', [ConsultationController::class, 'lecturerStudents'])->name('lecturer.students');
});

require __DIR__.'/settings.php';
