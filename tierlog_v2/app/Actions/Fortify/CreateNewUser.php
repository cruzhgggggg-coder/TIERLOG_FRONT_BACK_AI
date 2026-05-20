<?php

namespace App\Actions\Fortify;

use App\Concerns\PasswordValidationRules;
use App\Concerns\ProfileValidationRules;
use App\Models\User;
use Illuminate\Support\Facades\Validator;
use Laravel\Fortify\Contracts\CreatesNewUsers;

class CreateNewUser implements CreatesNewUsers
{
    use PasswordValidationRules, ProfileValidationRules;

    /**
     * Validate and create a newly registered user.
     *
     * @param  array<string, string>  $input
     */
    public function create(array $input): User
    {
        Validator::make($input, [
            'name' => ['required', 'string', 'max:255'],
            'email' => ['required', 'string', 'email', 'max:255', 'unique:users'],
            'password' => $this->passwordRules(),
            'role' => ['required', 'string', 'in:student,lecturer'],
            'nip' => ['nullable', 'required_if:role,lecturer', 'string', 'max:20'],
            'nim' => ['nullable', 'required_if:role,student', 'string', 'max:20'],
            'prodi' => ['nullable', 'required_if:role,student', 'string', 'max:100'],
        ])->validate();

        return \Illuminate\Support\Facades\DB::transaction(function () use ($input) {
            $user = User::create([
                'name' => $input['name'],
                'email' => $input['email'],
                'password' => $input['password'],
                'role' => $input['role'],
            ]);

            // Create associated profile based on role
            if ($user->role === 'student') {
                // Find latest available lecturer as a better fallback during testing
                $lecturer = \App\Models\Lecturer::latest()->first();
                
                if (!$lecturer) {
                    $systemLecturer = User::create([
                        'name' => 'System Lecturer',
                        'email' => 'lecturer@system.com',
                        'password' => \Illuminate\Support\Facades\Hash::make('password123'),
                        'role' => 'lecturer',
                    ]);
                    
                    $lecturer = $systemLecturer->lecturer()->create([
                        'name' => 'System Lecturer',
                        'nip' => '0000000000',
                        'faculty' => 'Technology',
                    ]);
                }

                $user->student()->create([
                    'name' => $user->name,
                    'nim' => $input['nim'] ?? 'STUDENT-' . str_pad($user->id, 5, '0', STR_PAD_LEFT),
                    'prodi' => $input['prodi'] ?? null,
                    'lecturer_id' => $lecturer->id,
                ]);
            } else {
                $user->lecturer()->create([
                    'name' => $user->name,
                    'nip' => $input['nip'] ?? 'NIP-' . str_pad($user->id, 8, '0', STR_PAD_LEFT),
                    'faculty' => $input['faculty'] ?? null,
                ]);
            }

            return $user;
        });
    }
}
