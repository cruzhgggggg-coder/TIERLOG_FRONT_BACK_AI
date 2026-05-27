<?php

namespace App\Http\Controllers;

use Illuminate\Http\Request;
use Illuminate\Support\Facades\Http;
use Illuminate\Support\Facades\Storage;
use Illuminate\Support\Facades\Auth;
use Inertia\Inertia;
use App\Events\FeedbackStatusUpdated;
use App\Events\AIChatMessageBroadcasted;

class ConsultationController extends Controller
{
    protected string $goBackendUrl;

    public function __construct()
    {
        $this->goBackendUrl = env('GO_BACKEND_URL', 'http://localhost:8080');
    }

    /**
     * Show the consultation list.
     */
    public function index()
    {
        $userId = Auth::id();
        // Fetch data from Go Backend with user filter
        $response = Http::get("{$this->goBackendUrl}/api/consultation", [
            'user_id' => $userId
        ]);
        
        $consultations = $response->successful() ? $response->json() : [];

        return Inertia::render('Consultation/Index', [
            'consultations' => $consultations
        ]);
    }

    /**
     * Show the consultation history (Arsip Log).
     */
    public function archive()
    {
        $userId = Auth::id();
        $response = Http::get("{$this->goBackendUrl}/api/consultation", [
            'user_id' => $userId
        ]);
        
        $consultations = $response->successful() ? $response->json() : [];

        return Inertia::render('Consultation/Archive', [
            'consultations' => $consultations
        ]);
    }

    /**
     * Handle the upload and proxy to Go Backend.
     */
    public function store(Request $request)
    {
        $userId = Auth::id();
        $request->validate([
            'paper' => 'required|file|mimes:docx|max:10240',
            'audio' => 'required|file|mimes:mp3,wav,m4a|max:20480',
        ]);

        // Proxying the multipart request to Go using stream resources to prevent high memory usage
        $response = Http::attach(
            'paper', 
            fopen($request->file('paper')->getRealPath(), 'r'), 
            $request->file('paper')->getClientOriginalName()
        )->attach(
            'audio',
            fopen($request->file('audio')->getRealPath(), 'r'),
            $request->file('audio')->getClientOriginalName()
        )->timeout(120)->post("{$this->goBackendUrl}/api/consultation", [
            'user_id' => $userId,
        ]);

        if ($response->successful()) {
            return redirect()->back()->with('success', 'Konsultasi berhasil diproses oleh Go Backend.');
        }

        return redirect()->back()->with('error', 'Gagal menghubungi Go Backend: ' . $response->body());
    }

    /**
     * Handle AI chat assistance via Go Backend.
     */
    public function chat(Request $request)
    {
        $request->validate([
            'log_id' => 'required',
            'query' => 'required|string',
        ]);

        $user = Auth::user();
        
        // Security Check
        if ($user->role === 'admin') {
            // Admin can access all logs
            $logExists = \App\Models\ConsultationLog::where('id', $request->log_id)->exists();
        } else {
            // Students can access their own logs
            // Lecturers can access logs of students they are assigned to
            $logExists = \App\Models\ConsultationLog::where('id', $request->log_id)
                ->where(function($query) use ($user) {
                    $query->whereHas('student', function($q) use ($user) {
                        $q->where('user_id', $user->id);
                    })
                    ->orWhereHas('student.lecturer', function($q) use ($user) {
                        $q->where('user_id', $user->id);
                    });
                })->exists();
        }

        if (!$logExists) {
            return response()->json(['error' => 'Akses ditolak atau Log tidak ditemukan.'], 403);
        }

        // Broadcast student query live
        try {
            broadcast(new AIChatMessageBroadcasted((int) $request->log_id, 'user', $request->input('query')))->toOthers();
        } catch (\Exception $e) {
            \Log::error("Failed to broadcast AIChatMessageBroadcasted student event: " . $e->getMessage());
        }

        $response = Http::timeout(120)->post("{$this->goBackendUrl}/api/ai/assist", [
            'log_id' => (int) $request->log_id,
            'query' => $request->input('query'),
            'model' => $request->input('model'),
        ]);

        if ($response->successful()) {
            $responseData = $response->json();
            // Broadcast AI response live
            if (isset($responseData['ai_response'])) {
                try {
                    broadcast(new AIChatMessageBroadcasted((int) $request->log_id, 'ai', $responseData['ai_response']))->toOthers();
                } catch (\Exception $e) {
                    \Log::error("Failed to broadcast AIChatMessageBroadcasted AI event: " . $e->getMessage());
                }
            }
            return response()->json($responseData);
        }

        $errorData = $response->json();
        $errorMessage = $errorData['error'] ?? ($errorData['message'] ?? 'Gagal menghubungi Go AI Service');
        
        return response()->json(['error' => $errorMessage], $response->status());
    }

    /**
     * Get statistics from Go Backend.
     */
    public function stats()
    {
        $userId = Auth::id();
        $response = Http::get("{$this->goBackendUrl}/api/stats", [
            'user_id' => $userId
        ]);
        
        if ($response->successful()) {
            $data = $response->json();
            // Calculate a more dynamic completion rate
            $total = $data['total_feedback'] ?? 1;
            $pending = $data['pending_feedback'] ?? 0;
            $rate = $total > 0 ? round((($total - $pending) / $total) * 100) : 0;

            $user = Auth::user();
            $lecturerName = 'Belum Ditentukan';
            if ($user->role === 'student' && $user->student && $user->student->lecturer) {
                $lecturerName = $user->student->lecturer->name;
            }

            return response()->json([
                'total_consultations' => $data['total_logs'] ?? 0,
                'total_feedback' => $data['total_feedback'] ?? 0,
                'pending_feedback' => $data['pending_feedback'] ?? 0,
                'major_feedback' => $data['major_feedback'] ?? 0,
                'completion_rate' => $rate,
                'draft_count' => $data['total_logs'] ?? 0, 
                'upcoming_quests' => $data['upcoming_quests'] ?? [],
                'lecturer_name' => $lecturerName,
            ]);
        }
    }

    /**
     * Update feedback status via Go Backend.
     */
    public function updateFeedbackStatus(Request $request, string $id)
    {
        $request->validate([
            'status' => 'required|string|in:Pending,Fixed,Validated',
        ]);

        $user = Auth::user();

        // Fetch FeedbackItem and perform BOLA ownership check
        $feedback = \App\Models\FeedbackItem::with(['consultationLog.student.lecturer', 'consultationLog.student'])->find($id);
        if (!$feedback || !$feedback->consultationLog || !$feedback->consultationLog->student) {
            return response()->json(['error' => 'Feedback item tidak ditemukan atau log tidak valid.'], 404);
        }

        $student = $feedback->consultationLog->student;

        // Security Check: Only student assigned to log or lecturer assigned to student can update status
        if ($user->role === 'student') {
            if ($student->user_id !== $user->id) {
                return response()->json(['error' => 'Akses ditolak. Anda hanya dapat mengubah status feedback milik sendiri.'], 403);
            }
        } elseif ($user->role === 'lecturer') {
            if (!$student->lecturer || $student->lecturer->user_id !== $user->id) {
                return response()->json(['error' => 'Akses ditolak. Anda hanya dapat mengubah status feedback untuk mahasiswa bimbingan Anda.'], 403);
            }
        } elseif ($user->role !== 'admin') {
            return response()->json(['error' => 'Akses ditolak.'], 403);
        }

        // Security Check: Only lecturer can set status to Validated
        if ($request->status === 'Validated' && $user->role !== 'lecturer') {
            return response()->json(['error' => 'Hanya Dosen yang dapat memvalidasi feedback.'], 403);
        }

        // Security Check: Student can only set status to Fixed or back to Pending (undo)
        if (in_array($request->status, ['Fixed', 'Pending']) && $user->role !== 'student') {
            return response()->json(['error' => 'Hanya Mahasiswa yang dapat mengubah status ini.'], 403);
        }

        $response = Http::put("{$this->goBackendUrl}/api/feedback/{$id}/status", [
            'status' => $request->status,
        ]);

        if ($response->successful()) {
            // Get log_id from Go response or from request
            $goData = $response->json();
            $logId = $goData['data']['log_id'] 
                  ?? $goData['log_id'] 
                  ?? $request->input('log_id');

            // Broadcast only if we have a log_id to build the channel name
            if ($logId) {
                try {
                    broadcast(new FeedbackStatusUpdated(
                        feedbackId: (int) $id,
                        logId: (int) $logId,
                        newStatus: $request->status,
                        updatedByRole: $user->role,
                    ))->toOthers();
                } catch (\Exception $e) {
                    \Log::error("Failed to broadcast FeedbackStatusUpdated event: " . $e->getMessage());
                }
            }
            return response()->json(['message' => 'Status feedback berhasil diperbarui.']);
        }

        return response()->json(['error' => 'Gagal memperbarui status feedback.'], $response->status());
    }

    public function lecturerConsultations()
    {
        $user = Auth::user();
        
        if ($user->role !== 'lecturer') {
            return response()->json(['error' => 'Unauthorized. Only lecturers can access this data.'], 403);
        }

        // Fetch the Lecturer profile to get the Profile ID
        $lecturer = $user->lecturer;
        if (!$lecturer) {
            return response()->json(['error' => 'Profil Dosen tidak ditemukan.'], 404);
        }

        $response = Http::get("{$this->goBackendUrl}/api/lecturer/{$lecturer->id}/consultations");
        return $response->json();
    }

    public function lecturerStudents()
    {
        $user = Auth::user();
        
        if ($user->role !== 'lecturer') {
            return response()->json(['error' => 'Unauthorized.'], 403);
        }

        $lecturer = $user->lecturer;
        if (!$lecturer) {
            return response()->json(['error' => 'Profil Dosen tidak ditemukan.'], 404);
        }

        $response = Http::get("{$this->goBackendUrl}/api/lecturer/{$lecturer->id}/students");
        return $response->json();
    }
}
