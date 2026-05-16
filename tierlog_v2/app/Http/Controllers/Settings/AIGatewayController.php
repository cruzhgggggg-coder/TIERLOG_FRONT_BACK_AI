<?php

namespace App\Http\Controllers\Settings;

use App\Http\Controllers\Controller;
use App\Models\User;
use Illuminate\Http\RedirectResponse;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\Auth;
use Illuminate\Support\Facades\Http;
use Inertia\Inertia;
use Inertia\Response;

class AIGatewayController extends Controller
{
    /**
     * Show the AI Gateway settings page.
     */
    public function edit(Request $request): Response
    {
        return Inertia::render('settings/AIGateway', [
            'status' => session('status'),
        ]);
    }

    /**
     * Update the user's AI keys and model preference.
     */
    public function update(Request $request): RedirectResponse
    {
        $user = $request->user();

        $validated = $request->validate([
            'openai_key' => ['nullable', 'string', 'max:255'],
            'gemini_key' => ['nullable', 'string', 'max:255'],
            'anthropic_key' => ['nullable', 'string', 'max:255'],
            'nvidia_key' => ['nullable', 'string', 'max:255'],
            'preferred_model' => ['nullable', 'string', 'max:50'],
        ]);

        $user->update($validated);

        return back()->with('status', 'ai-gateway-updated');
    }

    /**
     * Redeem a code to activate the AI Gateway.
     */
    public function redeem(Request $request): RedirectResponse
    {
        $request->validate([
            'code' => ['required', 'string'],
        ]);

        $goBackendUrl = config('services.go_backend.url', 'http://localhost:8080');

        $response = Http::post("{$goBackendUrl}/api/settings/redeem", [
            'user_id' => Auth::id(),
            'code' => $request->code,
        ]);

        if ($response->successful()) {
            // Force refresh user data to show gateway as active
            Auth::user()->refresh();
            return back()->with('status', 'gateway-activated');
        }

        return back()->withErrors(['code' => $response->json('error') ?? 'Gagal mengaktifkan kode redeem.']);
    }
}
