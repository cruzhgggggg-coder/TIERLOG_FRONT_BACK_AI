<?php

namespace Database\Seeders;

use App\Models\ConsultationLog;
use App\Models\FeedbackItem;
use App\Models\Lecturer;
use App\Models\Student;
use App\Models\User;
use Illuminate\Database\Seeder;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Hash;

class DemoDataSeeder extends Seeder
{
    /**
     * Seed demo data for TierLog presentation.
     * Run with: php artisan db:seed --class=DemoDataSeeder
     */
    public function run(): void
    {
        DB::statement('SET FOREIGN_KEY_CHECKS=0;');
        
        // Bersihkan data log dan feedback lama agar tidak bentrok
        FeedbackItem::truncate();
        ConsultationLog::truncate();

        DB::statement('SET FOREIGN_KEY_CHECKS=1;');

        // =========================================================
        // 1. CARI ATAU BUAT AKUN DOSEN
        // =========================================================
        // Prioritas 1: zaky@gmail.com (Dosen Zaky Manggala dari database saat ini)
        $dosenUser = User::where('email', 'zaky@gmail.com')->first();
        if ($dosenUser) {
            $dosenUser->update(['password' => Hash::make('password')]);
            $lecturer = Lecturer::where('user_id', $dosenUser->id)->first();
            if (!$lecturer) {
                $lecturer = Lecturer::create([
                    'user_id' => $dosenUser->id,
                    'nip'     => '123456789',
                    'name'    => 'Zaky Manggala',
                    'faculty' => 'Teknik Informatika',
                ]);
            }
        } else {
            // Prioritas 2: dosen1@university.ac.id
            $dosenUser = User::where('email', 'dosen1@university.ac.id')->first();
            if ($dosenUser) {
                $dosenUser->update(['password' => Hash::make('password')]);
                $lecturer = Lecturer::where('user_id', $dosenUser->id)->first();
                if (!$lecturer) {
                    $lecturer = Lecturer::create([
                        'user_id' => $dosenUser->id,
                        'nip'     => '198001012005011001',
                        'name'    => $dosenUser->name ?? 'Dr. Arsitek Go, M.Kom',
                        'faculty' => 'Informatika',
                    ]);
                }
            } else {
                // Fallback: dosen@tierlog.demo
                $dosenUser = User::where('email', 'dosen@tierlog.demo')->first();
                if (!$dosenUser) {
                    $dosenUser = User::create([
                        'name'     => 'Dr. Budi Santoso, M.Kom',
                        'email'    => 'dosen@tierlog.demo',
                        'password' => Hash::make('password'),
                        'role'     => 'lecturer',
                    ]);
                } else {
                    $dosenUser->update(['password' => Hash::make('password')]);
                }

                $lecturer = Lecturer::where('user_id', $dosenUser->id)->first();
                if (!$lecturer) {
                    $lecturer = Lecturer::create([
                        'user_id' => $dosenUser->id,
                        'nip'     => '198512012010121001',
                        'name'    => 'Dr. Budi Santoso, M.Kom',
                        'faculty' => 'Teknik Informatika',
                    ]);
                }
            }
        }

        // =========================================================
        // 2. CARI ATAU BUAT AKUN MAHASISWA (3 mahasiswa untuk demo yang kaya)
        // =========================================================

        // Mahasiswa 1 — utama untuk demo realtime (budi@gmail.com prioritas utama, lalu mhs1, lalu fallback)
        $mhs1User = User::where('email', 'budi@gmail.com')->first();
        if ($mhs1User) {
            $mhs1User->update(['password' => Hash::make('password')]);
            $student1 = Student::where('user_id', $mhs1User->id)->first();
            if (!$student1) {
                $student1 = Student::create([
                    'user_id'     => $mhs1User->id,
                    'lecturer_id' => $lecturer->id,
                    'nim'         => '123456987',
                    'name'        => 'Budi',
                    'prodi'       => 'Informatika',
                    'thesis_title'=> 'Web jual beli',
                ]);
            } else {
                $student1->update(['lecturer_id' => $lecturer->id]);
            }
        } else {
            $mhs1User = User::where('email', 'mhs1@university.ac.id')->first();
            if ($mhs1User) {
                $mhs1User->update(['password' => Hash::make('password')]);
                $student1 = Student::where('user_id', $mhs1User->id)->first();
                if (!$student1) {
                    $student1 = Student::create([
                        'user_id'     => $mhs1User->id,
                        'lecturer_id' => $lecturer->id,
                        'nim'         => '2200010001',
                        'name'        => 'Budi Mahasiswa',
                        'prodi'       => 'Teknik Informatika',
                        'thesis_title'=> 'Implementasi Microservices pada Sistem Log Bimbingan',
                    ]);
                } else {
                    $student1->update(['lecturer_id' => $lecturer->id]);
                }
            } else {
                $mhs1User = User::where('email', 'mahasiswa@tierlog.demo')->first();
                if (!$mhs1User) {
                    $mhs1User = User::create([
                        'name'     => 'Andi Pratama',
                        'email'    => 'mahasiswa@tierlog.demo',
                        'password' => Hash::make('password'),
                        'role'     => 'student',
                    ]);
                } else {
                    $mhs1User->update(['password' => Hash::make('password')]);
                }

                $student1 = Student::where('user_id', $mhs1User->id)->first();
                if (!$student1) {
                    $student1 = Student::create([
                        'user_id'     => $mhs1User->id,
                        'lecturer_id' => $lecturer->id,
                        'nim'         => '2021001',
                        'name'        => 'Andi Pratama',
                        'prodi'       => 'Teknik Informatika',
                        'thesis_title'=> 'Implementasi Sistem Log Bimbingan Skripsi Berbasis Kecerdasan Buatan Menggunakan Go dan Laravel',
                    ]);
                } else {
                    $student1->update(['lecturer_id' => $lecturer->id]);
                }
            }
        }

        // Mahasiswa 2 — variasi data
        $mhs2User = User::where('email', 'mahasiswa2@tierlog.demo')->first();
        if (!$mhs2User) {
            $mhs2User = User::create([
                'name'     => 'Sari Dewi',
                'email'    => 'mahasiswa2@tierlog.demo',
                'password' => Hash::make('password'),
                'role'     => 'student',
            ]);
        } else {
            $mhs2User->update(['password' => Hash::make('password')]);
        }

        $student2 = Student::where('user_id', $mhs2User->id)->first();
        if (!$student2) {
            $student2 = Student::create([
                'user_id'     => $mhs2User->id,
                'lecturer_id' => $lecturer->id,
                'nim'         => '2021002',
                'name'        => 'Sari Dewi',
                'prodi'       => 'Sistem Informasi',
                'thesis_title'=> 'Analisis dan Perancangan Sistem Manajemen Inventaris Berbasis Web menggunakan Framework Laravel',
            ]);
        }

        // =========================================================
        // 3. BUAT LOG KONSULTASI UNTUK MAHASISWA 1 (3 sesi)
        // =========================================================

        // --- Sesi 1: Sudah selesai, semua divalidasi ---
        $log1 = ConsultationLog::create([
            'student_id'          => $student1->id,
            'audio_filename'      => 'demo_session_1.mp3',
            'transcript_filename' => 'demo_session_1.txt',
            'transcript_text'     => 'Dosen: Selamat datang Andi, kita akan membahas progres bab pertama skripsimu. Andi: Terima kasih pak, saya sudah menyelesaikan kerangka bab 1 tapi masih ada beberapa bagian yang kurang jelas. Dosen: Baik, saya lihat draftnya. Ada beberapa catatan yang perlu kamu perbaiki. Pertama, latar belakang penelitian kamu terlalu umum, perlu lebih spesifik mengenai masalah yang ingin dipecahkan. Kedua, rumusan masalah harus lebih terukur dan jelas. Ketiga, daftar pustaka kamu masih menggunakan format yang tidak konsisten, sebaiknya gunakan format APA yang baku.',
            'paper_filename'      => 'draft_skripsi_andi_v1.docx',
            'created_at'          => now()->subDays(30),
            'updated_at'          => now()->subDays(30),
        ]);

        FeedbackItem::insert([
            [
                'consultation_log_id' => $log1->id,
                'log_id'              => $log1->id,
                'content'  => 'Latar belakang penelitian terlalu umum. Perlu dipersempit dan difokuskan pada masalah spesifik yang melatarbelakangi penelitian ini, seperti ketidakefisienan proses bimbingan manual saat ini.',
                'category' => 'Major',
                'status'   => 'Fixed',
                'created_at' => now()->subDays(30),
                'updated_at' => now()->subDays(25),
            ],
            [
                'consultation_log_id' => $log1->id,
                'log_id'              => $log1->id,
                'content'  => 'Rumusan masalah nomor 2 dan 3 masih kurang terukur. Gunakan kata kerja yang konkret seperti "bagaimana merancang", bukan "apakah sistem dapat".',
                'category' => 'Major',
                'status'   => 'Fixed',
                'created_at' => now()->subDays(30),
                'updated_at' => now()->subDays(24),
            ],
            [
                'consultation_log_id' => $log1->id,
                'log_id'              => $log1->id,
                'content'  => 'Format daftar pustaka tidak konsisten. Beberapa menggunakan APA, beberapa IEEE. Standarisasi seluruhnya ke format APA edisi ke-7.',
                'category' => 'Minor',
                'status'   => 'Fixed',
                'created_at' => now()->subDays(30),
                'updated_at' => now()->subDays(23),
            ],
        ]);

        // --- Sesi 2: Revisi sebagian sudah selesai ---
        $log2 = ConsultationLog::create([
            'student_id'          => $student1->id,
            'audio_filename'      => 'demo_session_2.mp3',
            'transcript_filename' => 'demo_session_2.txt',
            'transcript_text'     => 'Dosen: Bab 1 sudah jauh lebih baik Andi, sekarang kita bahas Bab 2 tentang landasan teori. Andi: Pak, saya sudah mengumpulkan referensi tapi bingung cara menyusunnya. Dosen: Landasan teori kamu perlu menyertakan teori-teori utama yang mendukung penelitian. Jangan lupa setiap konsep harus didukung minimal 2 referensi. Andi: Baik pak, apakah saya perlu menambahkan bagian penelitian terdahulu? Dosen: Ya, penelitian terdahulu sangat penting. Tambahkan setidaknya 5 penelitian terdahulu yang relevan dalam bentuk tabel perbandingan agar lebih terstruktur. Juga diagram arsitektur sistem kamu masih terlalu sederhana.',
            'paper_filename'      => 'draft_skripsi_andi_v2.docx',
            'created_at'          => now()->subDays(15),
            'updated_at'          => now()->subDays(15),
        ]);

        FeedbackItem::insert([
            [
                'consultation_log_id' => $log2->id,
                'log_id'              => $log2->id,
                'content'  => 'Landasan teori perlu menyertakan minimal 2 referensi akademik yang valid untuk setiap konsep yang disebutkan. Pastikan referensi yang digunakan adalah jurnal atau buku yang terbit dalam 10 tahun terakhir.',
                'category' => 'Major',
                'status'   => 'Fixed',
                'created_at' => now()->subDays(15),
                'updated_at' => now()->subDays(10),
            ],
            [
                'consultation_log_id' => $log2->id,
                'log_id'              => $log2->id,
                'content'  => 'Tambahkan bagian Penelitian Terdahulu berisi minimal 5 penelitian relevan dalam format tabel perbandingan (Penulis, Tahun, Judul, Metode, Hasil, Perbedaan dengan penelitian ini).',
                'category' => 'Major',
                'status'   => 'Pending',
                'created_at' => now()->subDays(15),
                'updated_at' => now()->subDays(15),
            ],
            [
                'consultation_log_id' => $log2->id,
                'log_id'              => $log2->id,
                'content'  => 'Diagram arsitektur sistem terlalu sederhana. Gambarkan interaksi antar komponen (Go Backend, Laravel Frontend, MySQL, AI Provider) secara lebih detail menggunakan diagram komponen UML.',
                'category' => 'Major',
                'status'   => 'Pending',
                'created_at' => now()->subDays(15),
                'updated_at' => now()->subDays(15),
            ],
            [
                'consultation_log_id' => $log2->id,
                'log_id'              => $log2->id,
                'content'  => 'Typo pada halaman 12: "implementsi" seharusnya "implementasi". Cek ulang ejaan secara menyeluruh sebelum pengumpulan berikutnya.',
                'category' => 'Minor',
                'status'   => 'Fixed',
                'created_at' => now()->subDays(15),
                'updated_at' => now()->subDays(12),
            ],
        ]);

        // --- Sesi 3: TERBARU — semua masih Pending (untuk demo Mark as Fixed) ---
        $log3 = ConsultationLog::create([
            'student_id'          => $student1->id,
            'audio_filename'      => 'demo_session_3.mp3',
            'transcript_filename' => 'demo_session_3.txt',
            'transcript_text'     => 'Dosen: Andi, saya sudah review bab 3 kamu tentang metodologi. Ada beberapa hal penting yang perlu diperbaiki. Andi: Siap pak, saya catat. Dosen: Pertama, metode pengembangan sistem yang kamu pilih, yaitu Waterfall, kurang cocok untuk proyek penelitian yang bersifat iteratif seperti ini. Saya sarankan gunakan metodologi Agile atau Scrum. Kedua, diagram alir penelitian kamu masih belum mencakup fase pengujian sistem secara eksplisit. Tambahkan tahap User Acceptance Testing. Ketiga, instrumen penelitian seperti kuesioner atau checklist pengujian belum terlampir. Andi: Baik pak, saya akan perbaiki segera. Dosen: Bagus, saya tunggu revisinya minggu depan. Jangan lupa juga tambahkan penjelasan tentang populasi dan sampel penelitian jika ada.',
            'paper_filename'      => 'draft_skripsi_andi_v3.docx',
            'created_at'          => now()->subDays(3),
            'updated_at'          => now()->subDays(3),
        ]);

        FeedbackItem::insert([
            [
                'consultation_log_id' => $log3->id,
                'log_id'              => $log3->id,
                'content'  => 'Metode pengembangan sistem Waterfall tidak cocok untuk penelitian iteratif ini. Ganti dengan metodologi Agile/Scrum dan jelaskan tahapan sprint yang akan digunakan beserta target deliverable setiap sprint.',
                'category' => 'Major',
                'status'   => 'Pending',
                'created_at' => now()->subDays(3),
                'updated_at' => now()->subDays(3),
            ],
            [
                'consultation_log_id' => $log3->id,
                'log_id'              => $log3->id,
                'content'  => 'Diagram alir penelitian belum mencakup fase pengujian (User Acceptance Testing/UAT). Tambahkan tahap UAT yang menjelaskan siapa penguji, skenario pengujian, dan kriteria penerimaan sistem.',
                'category' => 'Major',
                'status'   => 'Pending',
                'created_at' => now()->subDays(3),
                'updated_at' => now()->subDays(3),
            ],
            [
                'consultation_log_id' => $log3->id,
                'log_id'              => $log3->id,
                'content'  => 'Instrumen penelitian (kuesioner, checklist pengujian fungsional) belum terlampir di bagian lampiran. Buat dan lampirkan minimal 1 instrumen pengujian yang sesuai dengan tujuan penelitian.',
                'category' => 'Major',
                'status'   => 'Pending',
                'created_at' => now()->subDays(3),
                'updated_at' => now()->subDays(3),
            ],
            [
                'consultation_log_id' => $log3->id,
                'log_id'              => $log3->id,
                'content'  => 'Tambahkan sub-bab yang menjelaskan populasi dan sampel penelitian, termasuk teknik pengambilan sampel yang digunakan (purposive sampling, random sampling, dll).',
                'category' => 'Minor',
                'status'   => 'Pending',
                'created_at' => now()->subDays(3),
                'updated_at' => now()->subDays(3),
            ],
        ]);

        // =========================================================
        // 4. BUAT LOG KONSULTASI UNTUK MAHASISWA 2 (1 sesi)
        // =========================================================
        $log4 = ConsultationLog::create([
            'student_id'          => $student2->id,
            'audio_filename'      => 'demo_session_sari_1.mp3',
            'transcript_filename' => 'demo_session_sari_1.txt',
            'transcript_text'     => 'Dosen: Sari, saya sudah baca proposal skripsi kamu tentang sistem inventaris. Secara umum sudah cukup baik, tapi ada beberapa catatan. Sari: Terima kasih pak, saya senang mendengarnya. Apa saja yang perlu diperbaiki? Dosen: Pertama, judul penelitian kamu terlalu panjang. Sederhanakan menjadi tidak lebih dari 20 kata namun tetap informatif. Kedua, pada bagian analisis sistem yang berjalan, kamu perlu menggunakan diagram UML yang lebih lengkap, setidaknya Use Case Diagram, Activity Diagram, dan Sequence Diagram. Ketiga, untuk framework Laravel yang kamu gunakan, jelaskan alasan pemilihan framework ini dibanding alternatif lain seperti CodeIgniter atau Symfony.',
            'paper_filename'      => 'draft_skripsi_sari_v1.docx',
            'created_at'          => now()->subDays(7),
            'updated_at'          => now()->subDays(7),
        ]);

        FeedbackItem::insert([
            [
                'consultation_log_id' => $log4->id,
                'log_id'              => $log4->id,
                'content'  => 'Judul penelitian terlalu panjang (38 kata). Sederhanakan menjadi maksimal 20 kata yang tetap mencerminkan isi penelitian secara informatif dan padat.',
                'category' => 'Minor',
                'status'   => 'Pending',
                'created_at' => now()->subDays(7),
                'updated_at' => now()->subDays(7),
            ],
            [
                'consultation_log_id' => $log4->id,
                'log_id'              => $log4->id,
                'content'  => 'Analisis sistem yang berjalan membutuhkan diagram UML yang lengkap: Use Case Diagram, Activity Diagram, dan Sequence Diagram. Diagram harus dibuat menggunakan tools standar (Lucidchart, Draw.io) bukan sketsa tangan.',
                'category' => 'Major',
                'status'   => 'Pending',
                'created_at' => now()->subDays(7),
                'updated_at' => now()->subDays(7),
            ],
            [
                'consultation_log_id' => $log4->id,
                'log_id'              => $log4->id,
                'content'  => 'Tambahkan justifikasi pemilihan framework Laravel dibandingkan alternatif lain (CodeIgniter, Symfony, Lumen). Jelaskan kelebihan Laravel yang relevan dengan kebutuhan proyek ini.',
                'category' => 'Minor',
                'status'   => 'Fixed',
                'created_at' => now()->subDays(7),
                'updated_at' => now()->subDays(5),
            ],
        ]);

        $this->command->info('');
        $this->command->info('✅ Demo data berhasil disinkronkan!');
        $this->command->info('');
        $this->command->info('📋 Kredensial Login Demo (Prioritas Akun Aktif / SQL Dump):');
        $this->command->info('');
        $this->command->info('  🎓 DOSEN (Lecturer)');
        $this->command->info('     Email   : zaky@gmail.com (atau dosen1@university.ac.id / dosen@tierlog.demo)');
        $this->command->info('     Password: password');
        $this->command->info('');
        $this->command->info('  👨‍🎓 MAHASISWA 1 (Student - Budi / Andi Pratama)');
        $this->command->info('     Email   : budi@gmail.com (atau mhs1@university.ac.id / mahasiswa@tierlog.demo)');
        $this->command->info('     Password: password');
        $this->command->info('');
        $this->command->info('  👩‍🎓 MAHASISWA 2 (Student - Sari Dewi)');
        $this->command->info('     Email   : mahasiswa2@tierlog.demo');
        $this->command->info('     Password: password');
        $this->command->info('');
        $this->command->info('📂 Data yang telah dibuat:');
        $this->command->info('   - Log Bimbingan dan revisi berhasil dipetakan ke mahasiswa di atas!');
        $this->command->info('   - 4 Log Bimbingan & 14 Feedback Item ditambahkan.');
        $this->command->info('');
        $this->command->info('💡 Untuk chat Oracle, login sebagai Mahasiswa dan pilih Sesi 3 (Metodologi)');
        $this->command->info('   yang memiliki 4 feedback Pending baru dari 3 hari lalu.');
    }
}
