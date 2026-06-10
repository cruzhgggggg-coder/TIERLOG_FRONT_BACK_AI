<?php

use Illuminate\Support\Facades\Route;
use Inertia\Inertia;

Route::get('/', fn () => Inertia::render('GalaxyWelcome'))->name('home');
Route::get('/login', fn () => Inertia::render('Login'))->name('login');
Route::get('/register', fn () => Inertia::render('Register'))->name('register');
Route::get('/dashboard', fn () => Inertia::render('Dashboard'))->name('dashboard');
Route::get('/consultations', fn () => Inertia::render('Consultations'))->name('consultations');
Route::get('/archive', fn () => Inertia::render('Archive'))->name('archive');
Route::get('/lecturer-dashboard', fn () => Inertia::render('LecturerDashboard'))->name('lecturer-dashboard');
Route::get('/settings/profile', fn () => Inertia::render('settings/Profile'))->name('settings.profile');
Route::get('/settings/security', fn () => Inertia::render('settings/Security'))->name('settings.security');
Route::get('/settings/ai-gateway', fn () => Inertia::render('settings/AiGateway'))->name('settings.ai-gateway');
