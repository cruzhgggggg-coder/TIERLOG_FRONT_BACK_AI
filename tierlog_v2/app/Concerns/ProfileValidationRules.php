<?php

namespace App\Concerns;

use App\Models\User;
use Illuminate\Contracts\Validation\ValidationRule;
use Illuminate\Validation\Rule;

trait ProfileValidationRules
{
    /**
     * Get the validation rules used to validate user profiles.
     *
     * @return array<string, array<int, ValidationRule|array<mixed>|string>>
     */
    protected function profileRules(?int $userId = null, ?string $role = null): array
    {
        $rules = [
            'name' => $this->nameRules(),
            'email' => $this->emailRules($userId),
        ];

        // Determine role: prioritize passed $role, then fallback to current user
        $currentRole = $role;
        if (!$currentRole && method_exists($this, 'user') && $this->user()) {
            $currentRole = $this->user()->role;
        }

        if ($currentRole === 'student') {
            $rules['nim'] = ['required', 'string', 'max:20'];
            $rules['prodi'] = ['required', 'string', 'max:100'];
            $rules['thesis_title'] = ['nullable', 'string'];
            $rules['lecturer_id'] = ['nullable', 'integer'];
        } elseif ($currentRole === 'lecturer') {
            $rules['nip'] = ['required', 'string', 'max:20'];
            $rules['faculty'] = ['required', 'string', 'max:100'];
        }

        return $rules;
    }

    /**
     * Get the validation rules used to validate user names.
     *
     * @return array<int, ValidationRule|array<mixed>|string>
     */
    protected function nameRules(): array
    {
        return ['required', 'string', 'max:255'];
    }

    /**
     * Get the validation rules used to validate user emails.
     *
     * @return array<int, ValidationRule|array<mixed>|string>
     */
    protected function emailRules(?int $userId = null): array
    {
        return [
            'required',
            'string',
            'email',
            'max:255',
            $userId === null
                ? Rule::unique(User::class)
                : Rule::unique(User::class)->ignore($userId),
        ];
    }
}
