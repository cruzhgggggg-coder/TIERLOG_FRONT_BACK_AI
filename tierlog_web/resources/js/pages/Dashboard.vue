<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue';
import { router } from '@inertiajs/vue3';
import { useAuthStore } from '@/stores/auth';
import type { DashboardStats, ConsultationLog, StudentProfile } from '@/types';
import RequireAuth from '@/components/RequireAuth.vue';
import NavBar from '@/components/NavBar.vue';
import UiPage from '@/components/UiPage.vue';
import UiHeading from '@/components/UiHeading.vue';
import UiCard from '@/components/UiCard.vue';
import UiStatCard from '@/components/UiStatCard.vue';
import UiBadge from '@/components/UiBadge.vue';
import { CheckCircleIcon, ClockIcon, AlertIcon, ProfileIcon, BellIcon } from '@/components/icons';

const auth = useAuthStore();
const stats = ref<DashboardStats | null>(null);
const consultations = ref<ConsultationLog[]>([]);
const students = ref<StudentProfile[]>([]);
const loading = ref(true);
const error = ref('');

interface Toast {
  id: number;
  message: string;
  type: 'success' | 'info' | 'warning' | 'error';
}

const toasts = ref<Toast[]>([]);
let toastIdCounter = 0;

function showToast(message: string, type: 'success' | 'info' | 'warning' | 'error' = 'info') {
  const id = toastIdCounter++;
  toasts.value.push({ id, message, type });
  setTimeout(() => {
    removeToast(id);
  }, 4000);
}

function removeToast(id: number) {
  toasts.value = toasts.value.filter((t) => t.id !== id);
}

const isLecturer = computed(() => auth.user?.role === 'lecturer');
const isStudent = computed(() => auth.user?.role === 'student');

const statCards = computed(() => {
  if (!stats.value) return [];
  if (isLecturer.value) {
    return [
      { label: 'Total Consultations', value: stats.value.total_consultations },
      { label: 'Validation Queue', value: stats.value.pending_feedback },
      { label: 'Completion Rate', value: `${stats.value.completion_rate}%` },
      { label: 'Active Students', value: stats.value.student_count ?? 0 },
    ];
  }
  return [
    { label: 'Total Sessions', value: stats.value.total_consultations },
    { label: 'Pending Revisions', value: stats.value.pending_feedback },
    { label: 'Completion Rate', value: `${stats.value.completion_rate}%` },
    { label: 'Draft Count', value: stats.value.draft_count },
  ];
});

function getStudentStatus(studentId: number): string {
  const studentLogs = consultations.value.filter(c => c.student_id === studentId);
  if (!studentLogs.length) return 'NO SUBMISSIONS';
  let hasPending = false;
  studentLogs.forEach(log => {
    if (log.feedback_items?.some(item => item.status === 'Pending' || item.status === 'Fixed')) {
      hasPending = true;
    }
  });
  return hasPending ? 'NEW SUBMISSIONS' : 'ALL CLEAR';
}

function getStudentStatusColor(studentId: number): string {
  const status = getStudentStatus(studentId);
  if (status === 'NEW SUBMISSIONS') return '#f59e0b';
  if (status === 'ALL CLEAR') return '#22c55e';
  return '#71717a';
}

function studentInitials(name: string): string {
  return name.split(' ').map(w => w[0]).slice(0, 2).join('').toUpperCase();
}

let ws: WebSocket | null = null;

function connectWebSocket() {
  const apiUrl = import.meta.env.VITE_API_URL || 'http://127.0.0.1:8080';
  const token = auth.accessToken;
  if (!token) return;
  const wsUrl = apiUrl.replace(/^http/, 'ws') + `/ws?token=${token}`;
  try {
    ws = new WebSocket(wsUrl);
    ws.onopen = () => {
      consultations.value.forEach((log) => {
        ws?.send(JSON.stringify({ action: 'subscribe', room: `consultation.${log.id}` }));
      });
    };
    ws.onmessage = (event) => {
      try {
        const payload = JSON.parse(event.data);
        if (payload.event === 'feedback.new') {
          showToast('New feedback received', 'info');
          fetchAllData();
        } else if (payload.event === 'feedback.status-updated') {
          showToast('Feedback status updated', 'success');
          fetchAllData();
        }
      } catch {}
    };
    ws.onclose = () => {
      setTimeout(connectWebSocket, 5000);
    };
    ws.onerror = () => {
      ws?.close();
    };
  } catch {}
}

async function fetchAllData() {
  loading.value = true;
  error.value = '';

  try {
    const statsRes = await auth.api<DashboardStats>('/dashboard/stats');
    if (statsRes.ok && statsRes.data) {
      stats.value = statsRes.data;
    } else if (statsRes.error) {
      error.value = statsRes.error;
    }

    if (isStudent.value) {
      const consRes = await auth.api<{ data: ConsultationLog[] }>('/consultations');
      if (consRes.ok && consRes.data) {
        consultations.value = consRes.data.data ?? [];
      }
    }

    if (isLecturer.value) {
      const [stuRes, consRes] = await Promise.all([
        auth.api<{ data: StudentProfile[] }>('/lecturer/students'),
        auth.api<{ data: ConsultationLog[] }>('/lecturer/consultations'),
      ]);
      if (stuRes.ok && stuRes.data) {
        students.value = stuRes.data.data ?? [];
      }
      if (consRes.ok && consRes.data) {
        consultations.value = consRes.data.data ?? [];
      }
    }
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to load data';
  }

  loading.value = false;
}

onMounted(() => {
  fetchAllData();
  connectWebSocket();
});

onUnmounted(() => {
  ws?.close();
  ws = null;
});
</script>

<template>
  <RequireAuth>
    <!-- Real-time Toast Notifications -->
    <transition-group
      name="toast"
      tag="div"
      class="fixed top-6 right-6 z-[100] flex flex-col gap-3 max-w-sm w-full pointer-events-none px-4 sm:px-0"
    >
      <div
        v-for="toast in toasts"
        :key="toast.id"
        class="pointer-events-auto flex items-center justify-between gap-3.5 rounded-xl border p-4 shadow-2xl backdrop-blur-md transition-all duration-300 toast-banner"
        :class="`toast-${toast.type}`"
      >
        <div class="flex items-center gap-3">
          <div class="toast-icon-wrapper shrink-0">
            <CheckCircleIcon v-if="toast.type === 'success'" :size="18" />
            <BellIcon v-else-if="toast.type === 'info'" :size="18" />
            <AlertIcon v-else :size="18" />
          </div>
          <p class="text-sm font-medium text-stone-100">{{ toast.message }}</p>
        </div>
        <button
          @click="removeToast(toast.id)"
          class="text-stone-500 hover:text-stone-300 transition-colors duration-150 text-xs font-semibold px-1 py-0.5 rounded focus:outline-none cursor-pointer"
        >
          ✕
        </button>
      </div>
    </transition-group>

    <NavBar />
    <UiPage>
      <div class="mesh-gradient dot-grid space-y-8">
        <UiHeading
          class="font-[family-name:var(--font-display)]"
          :title="isLecturer ? 'Evaluation Portal' : 'Dashboard'"
          :subtitle="isLecturer
            ? 'Overview of supervised students, revision validations, and thesis progress.'
            : 'Track your consultation sessions, revision tasks, and academic progress.'"
        />

        <div v-if="error" class="flex items-center gap-2.5 rounded-lg border border-red-500/20 bg-red-500/5 px-4 py-3">
          <AlertIcon :size="16" color="#ef4444" />
          <span class="text-sm text-red-400">{{ error }}</span>
        </div>

        <div v-if="loading" class="flex items-center justify-center py-20">
          <div class="h-6 w-6 animate-spin rounded-full border-2 border-amber-500 border-t-transparent" />
        </div>

        <template v-else>
          <div class="grid grid-cols-2 gap-3 lg:grid-cols-4">
            <UiStatCard
              v-for="card in statCards"
              :key="card.label"
              :label="card.label"
              :value="card.value"
            />
          </div>

          <template v-if="isStudent">
            <UiCard>
              <div class="p-5">
                <div class="mb-4 flex items-center justify-between">
                  <h3 class="text-sm font-semibold text-stone-200">Your Revision Tasks</h3>
                  <UiBadge
                    v-if="stats?.upcoming_quests?.length"
                    :text="`${stats.upcoming_quests.length} active`"
                    color="bg-amber-500/10 text-amber-400"
                  />
                </div>

                <div v-if="stats?.upcoming_quests?.length" class="space-y-2">
                  <div
                    v-for="quest in stats.upcoming_quests"
                    :key="quest.id"
                    class="flex items-start gap-3 rounded-lg border border-white/[0.04] bg-[#1c1c1e]/50 p-3.5"
                  >
                    <div class="mt-0.5 shrink-0">
                      <CheckCircleIcon v-if="quest.status === 'Validated'" :size="14" color="#22c55e" />
                      <ClockIcon v-else-if="quest.status === 'Fixed'" :size="14" color="#3b82f6" />
                      <ClockIcon v-else :size="14" color="#f59e0b" />
                    </div>
                    <div class="min-w-0 flex-1">
                      <div class="mb-1 flex items-center gap-1.5">
                        <UiBadge :text="quest.category" :color="quest.category === 'Major' ? '#ef4444' : '#71717a'" />
                        <UiBadge
                          :text="quest.status === 'Validated' ? 'Approved' : quest.status === 'Fixed' ? 'Submitted' : 'Pending'"
                        />
                      </div>
                      <p class="text-sm text-stone-300">{{ quest.content }}</p>
                    </div>
                  </div>
                </div>

                <div v-else class="flex flex-col items-center gap-2.5 py-12">
                  <CheckCircleIcon :size="24" color="#22c55e" />
                  <p class="text-sm text-stone-600">No active revision tasks.</p>
                </div>
              </div>
            </UiCard>
          </template>

          <template v-if="isLecturer">
            <UiCard>
              <div class="p-5">
                <div class="mb-4 flex items-center justify-between">
                  <h3 class="text-sm font-semibold text-stone-200">Supervised Students</h3>
                  <span class="text-xs text-stone-600">{{ students.length }} students</span>
                </div>

                <div v-if="students.length" class="grid grid-cols-1 gap-3 sm:grid-cols-2">
                  <div
                    v-for="student in students"
                    :key="student.id"
                    class="cursor-pointer rounded-lg border border-white/[0.04] bg-[#1c1c1e]/50 p-4 transition-colors duration-150 hover:border-white/[0.08]"
                    @click="router.visit('/lecturer-dashboard')"
                  >
                    <div class="mb-3 flex items-start justify-between">
                      <div class="flex items-center gap-2.5">
                        <div class="flex h-8 w-8 shrink-0 items-center justify-center rounded-md bg-zinc-800 text-xs font-semibold text-stone-500">
                          {{ studentInitials(student.name) }}
                        </div>
                        <div>
                          <p class="text-sm font-medium text-stone-100">{{ student.name }}</p>
                          <p class="text-[11px] text-stone-600">{{ student.nim }} &middot; {{ student.prodi }}</p>
                        </div>
                      </div>
                      <UiBadge :text="getStudentStatus(student.id)" :color="getStudentStatusColor(student.id)" />
                    </div>
                    <p class="line-clamp-2 text-xs leading-relaxed text-stone-600">
                      {{ student.thesis_title || 'Research title not registered yet.' }}
                    </p>
                  </div>
                </div>

                <div v-else class="flex flex-col items-center gap-2.5 py-12">
                  <ProfileIcon :size="24" color="#52525b" />
                  <p class="text-sm text-stone-600">No supervised students assigned.</p>
                </div>
              </div>
            </UiCard>
          </template>
        </template>
      </div>
    </UiPage>
  </RequireAuth>
</template>

<style scoped>
.toast-banner {
  box-shadow: 0 10px 30px -10px rgba(0, 0, 0, 0.7);
}

.toast-success {
  background-color: hsla(142, 76%, 8%, 0.95);
  border-color: hsla(142, 70%, 45%, 0.35);
}
.toast-success .toast-icon-wrapper {
  color: hsl(142, 70%, 55%);
}

.toast-info {
  background-color: hsla(217, 90%, 8%, 0.95);
  border-color: hsla(217, 90%, 50%, 0.35);
}
.toast-info .toast-icon-wrapper {
  color: hsl(217, 90%, 65%);
}

.toast-warning {
  background-color: hsla(38, 90%, 6%, 0.95);
  border-color: hsla(38, 90%, 50%, 0.35);
}
.toast-warning .toast-icon-wrapper {
  color: hsl(38, 90%, 60%);
}

.toast-error {
  background-color: hsla(0, 85%, 8%, 0.95);
  border-color: hsla(0, 85%, 50%, 0.35);
}
.toast-error .toast-icon-wrapper {
  color: hsl(0, 85%, 65%);
}

/* Toast animations */
.toast-enter-active,
.toast-leave-active {
  transition: all 0.35s cubic-bezier(0.21, 1.02, 0.43, 1.01);
}

.toast-enter-from {
  opacity: 0;
  transform: translateY(-20px) scale(0.95);
}

.toast-leave-to {
  opacity: 0;
  transform: scale(0.95);
}

.toast-move {
  transition: transform 0.3s ease;
}
</style>
