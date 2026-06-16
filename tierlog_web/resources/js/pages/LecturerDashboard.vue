<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, nextTick, watch } from 'vue';
import { useAuthStore } from '@/stores/auth';
import type { StudentProfile, ConsultationLog, FeedbackItem, DashboardStats, SessionInfo } from '@/types';
import RequireAuth from '@/components/RequireAuth.vue';
import NavBar from '@/components/NavBar.vue';
import UiPage from '@/components/UiPage.vue';
import UiHeading from '@/components/UiHeading.vue';
import UiCard from '@/components/UiCard.vue';
import UiStatCard from '@/components/UiStatCard.vue';
import UiBadge from '@/components/UiBadge.vue';
import UiButton from '@/components/UiButton.vue';
import { AlertIcon, ConsultationIcon, ProfileIcon, CheckCircleIcon, ClockIcon } from '@/components/icons';
import RejectionDialog from '@/components/RejectionDialog.vue';

interface DirectMessage {
  id: number;
  sender_role: string;
  content: string;
  created_at: string;
}

interface Toast {
  id: number;
  message: string;
  type: 'info' | 'success' | 'error' | 'warning';
}

const API_URL = import.meta.env.VITE_API_URL || 'http://127.0.0.1:8080';
const auth = useAuthStore();
const loading = ref(true);
const error = ref('');
const students = ref<StudentProfile[]>([]);
const selectedStudentId = ref<number | null>(null);
const consultations = ref<ConsultationLog[]>([]);
const directMessages = ref<DirectMessage[]>([]);
const activeTab = ref<'overview' | 'revisions' | 'sessions' | 'chat'>('overview');
const feedbackText = ref('');
const chatInput = ref('');
const chatLoading = ref(false);
const validatingId = ref<number | null>(null);
const rejectionDialogOpen = ref(false);
const rejectionFeedbackId = ref<number | null>(null);
const toasts = ref<Toast[]>([]);
const stats = ref<DashboardStats | null>(null);
const sessionFilter = ref<number | null>(null); // null = "All Sessions"
const studentSessions = ref<SessionInfo[]>([]);
let toastCounter = 0;

const selectedStudent = computed(() =>
  students.value.find((s) => s.id === selectedStudentId.value) ?? null,
);

const selectedStudentLogs = computed(() =>
  consultations.value
    .filter((l) => l.student_id === selectedStudentId.value)
    .sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime()),
);

const latestLog = computed(() => selectedStudentLogs.value[0] ?? null);

const allValidationQueue = computed(() =>
  consultations.value.flatMap((l) =>
    (l.feedback_items ?? []).filter((f) => f.status === 'Fixed'),
  ),
);

const sortedDirectMessages = computed(() =>
  [...directMessages.value].sort(
    (a, b) => new Date(a.created_at).getTime() - new Date(b.created_at).getTime(),
  ),
);

// ── Session-based filtering for revisions ──

const filteredRevisionItems = computed(() => {
  const allItems = selectedStudentLogs.value.flatMap(l => l.feedback_items ?? []);
  if (sessionFilter.value === null) return allItems;
  const targetLog = consultations.value.find(l => l.id === sessionFilter.value);
  return targetLog?.feedback_items ?? [];
});

async function fetchStudentSessions(studentId: number) {
  const res = await auth.api<{ data: SessionInfo[] }>(
    `/lecturer/student/${studentId}/sessions`,
  );
  if (res.ok && res.data) {
    studentSessions.value = res.data.data ?? [];
  }
}

function getStudentStats(studentId: number) {
  const studentLogs = consultations.value.filter((l) => l.student_id === studentId);
  let pending = 0,
    fixed = 0,
    validated = 0,
    rejected = 0;
  studentLogs.forEach((log) => {
    (log.feedback_items ?? []).forEach((f) => {
      if (f.status === 'Pending') pending++;
      else if (f.status === 'Fixed') fixed++;
      else if (f.status === 'Validated') validated++;
      else if (f.status === 'Rejected') rejected++;
    });
  });
  return { sessions: studentLogs.length, pending, fixed, validated, rejected };
}

function studentStatusColor(studentId: number): string {
  const { pending, sessions } = getStudentStats(studentId);
  if (sessions === 0) return '#71717a';
  if (pending > 0) return '#f59e0b';
  return '#22c55e';
}

function studentStatusLabel(studentId: number): string {
  const { pending, sessions } = getStudentStats(studentId);
  if (sessions === 0) return 'NO SESSIONS';
  if (pending > 0) return `${pending} PENDING`;
  return 'ALL CLEAR';
}

function completionPercent(studentId: number): number {
  const { pending, fixed, validated } = getStudentStats(studentId);
  const total = pending + fixed + validated;
  if (total === 0) return 0;
  return Math.round((validated / total) * 100);
}

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString('en-US', {
    weekday: 'short',
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  });
}

function formatTime(dateStr: string): string {
  return new Date(dateStr).toLocaleTimeString('en-US', {
    hour: '2-digit',
    minute: '2-digit',
  });
}

function truncate(text: string, len: number): string {
  if (!text) return '';
  return text.length > len ? text.slice(0, len) + '...' : text;
}

function studentInitials(name: string): string {
  return name
    .split(' ')
    .map((w) => w[0])
    .slice(0, 2)
    .join('')
    .toUpperCase();
}

function addToast(message: string, type: Toast['type'] = 'info') {
  const id = ++toastCounter;
  toasts.value.push({ id, message, type });
  setTimeout(() => {
    toasts.value = toasts.value.filter((t) => t.id !== id);
  }, 4500);
}

function selectStudent(id: number) {
  selectedStudentId.value = id;
  activeTab.value = 'overview';
}

async function handleValidate(feedbackId: number) {
  validatingId.value = feedbackId;
  const parentLog = consultations.value.find((l) =>
    (l.feedback_items ?? []).some((f) => f.id === feedbackId),
  );
  const res = await auth.api(`/consultations/feedback/${feedbackId}/status`, {
    method: 'PUT',
    body: JSON.stringify({ status: 'Validated', log_id: parentLog?.id ?? 0 }),
  });
  if (res.ok) {
    await fetchAllData();
    addToast('Feedback validated', 'success');
  } else {
    addToast(res.error || 'Failed to validate', 'error');
  }
  validatingId.value = null;
}

function openRejectionDialog(feedbackId: number) {
  rejectionFeedbackId.value = feedbackId;
  rejectionDialogOpen.value = true;
}

async function handleRejectConfirm(comment: string) {
  if (rejectionFeedbackId.value === null) return;
  const feedbackId = rejectionFeedbackId.value;
  rejectionDialogOpen.value = false;
  
  validatingId.value = feedbackId;
  const parentLog = consultations.value.find((l) =>
    (l.feedback_items ?? []).some((f) => f.id === feedbackId),
  );
  const res = await auth.api(`/consultations/feedback/${feedbackId}/status`, {
    method: 'PUT',
    body: JSON.stringify({ 
      status: 'Rejected', 
      log_id: parentLog?.id ?? 0,
      comment: comment
    }),
  });
  if (res.ok) {
    await fetchAllData();
    addToast('Feedback rejected', 'warning');
  } else {
    addToast(res.error || 'Failed to reject', 'error');
  }
  validatingId.value = null;
  rejectionFeedbackId.value = null;
}

async function handleUndo(feedbackId: number) {
  validatingId.value = feedbackId;
  const parentLog = consultations.value.find((l) =>
    (l.feedback_items ?? []).some((f) => f.id === feedbackId),
  );
  const res = await auth.api(`/consultations/feedback/${feedbackId}/status`, {
    method: 'PUT',
    body: JSON.stringify({ status: 'Pending', log_id: parentLog?.id ?? 0 }),
  });
  if (res.ok) {
    await fetchAllData();
    addToast('Feedback returned to pending', 'info');
  } else {
    addToast(res.error || 'Failed to undo', 'error');
  }
  validatingId.value = null;
}

async function handleAddFeedback() {
  if (!latestLog.value || !feedbackText.value.trim()) return;
  const res = await auth.api<{ data: FeedbackItem }>(
    `/consultations/${latestLog.value.id}/add-feedback`,
    {
      method: 'POST',
      body: JSON.stringify({ content: feedbackText.value.trim() }),
    },
  );
  if (res.ok) {
    await fetchAllData();
    feedbackText.value = '';
    addToast('Feedback dispatched', 'success');
  } else {
    addToast(res.error || 'Failed to add feedback', 'error');
  }
}

async function loadDirectMessages(logId: number) {
  const res = await auth.api<{ data: DirectMessage[] }>(`/consultations/${logId}/direct-messages`);
  if (res.ok && res.data) {
    directMessages.value = res.data.data ?? [];
  }
}

async function sendDirectMessage() {
  if (!latestLog.value || !chatInput.value.trim() || chatLoading.value) return;
  const draft = chatInput.value;
  chatInput.value = '';
  chatLoading.value = true;
  const res = await auth.api<{ data: DirectMessage }>(
    `/consultations/${latestLog.value.id}/direct-messages`,
    { method: 'POST', body: JSON.stringify({ content: draft }) },
  );
  if (res.ok && res.data) {
    const msg = res.data.data ?? res.data;
    if (!directMessages.value.some((m) => m.id === (msg as DirectMessage).id)) {
      directMessages.value.push(msg as DirectMessage);
    }
  }
  chatLoading.value = false;
}

const chatContainer = ref<HTMLDivElement | null>(null);
function scrollChatToBottom() {
  if (chatContainer.value) chatContainer.value.scrollTop = chatContainer.value.scrollHeight;
}

let ws: WebSocket | null = null;
let stopConsultationWatcher: (() => void) | null = null;
function connectWebSocket() {
  const token = auth.accessToken;
  if (!token) return;
  const wsUrl = API_URL.replace(/^http/, 'ws') + `/ws?token=${token}`;

  if (stopConsultationWatcher) {
    stopConsultationWatcher();
    stopConsultationWatcher = null;
  }

  const subscribedRooms = new Set<string>();
  const subscribeToRooms = () => {
    const activeWs = ws;
    if (activeWs && activeWs.readyState === WebSocket.OPEN) {
      consultations.value.forEach((log) => {
        const roomName = `consultation.${log.id}`;
        if (!subscribedRooms.has(roomName)) {
          activeWs.send(JSON.stringify({ action: 'subscribe', room: roomName }));
          subscribedRooms.add(roomName);
        }
      });
    }
  };

  try {
    ws = new WebSocket(wsUrl);
    ws.onopen = () => {
      subscribeToRooms();
    };
    stopConsultationWatcher = watch(
      consultations,
      () => {
        subscribeToRooms();
      },
      { deep: true }
    );
    ws.onmessage = (event) => {
      try {
        const payload = JSON.parse(event.data);
        if (payload.event === 'feedback.new' || payload.event === 'feedback.status-updated') {
          fetchAllData();
          if (payload.event === 'feedback.status-updated' && payload.data?.status === 'Fixed') {
            addToast('Student submitted a revision fix', 'info');
          }
        }
        if (payload.event === 'chat.direct-message' && payload.data?.sender_role === 'student') {
          if (latestLog.value && payload.data.log_id === latestLog.value.id) {
            directMessages.value.push(payload.data);
            nextTick(scrollChatToBottom);
          }
          addToast('New message from student', 'info');
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
    const [s, st, lg] = await Promise.all([
      auth.api<DashboardStats>('/dashboard/stats'),
      auth.api<{ data: StudentProfile[] }>('/lecturer/students'),
      auth.api<{ data: ConsultationLog[] }>('/lecturer/consultations'),
    ]);
    if (s.ok && s.data) stats.value = s.data;
    if (st.ok && st.data) students.value = st.data.data ?? [];
    if (lg.ok && lg.data) consultations.value = lg.data.data ?? [];
    if (!selectedStudentId.value && students.value.length > 0) {
      selectedStudentId.value = students.value[0].id;
    }
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to load data';
  }
  loading.value = false;
}

watch(selectedStudentId, (id) => {
  if (id) {
    sessionFilter.value = null;
    fetchStudentSessions(id);
    const log = consultations.value.find((l) => l.student_id === id);
    if (log) loadDirectMessages(log.id);
    else directMessages.value = [];
  }
});

watch(activeTab, (tab) => {
  if (tab === 'chat') nextTick(scrollChatToBottom);
});

onMounted(() => {
  fetchAllData();
  connectWebSocket();
});

onUnmounted(() => {
  ws?.close();
  ws = null;
  if (stopConsultationWatcher) {
    stopConsultationWatcher();
    stopConsultationWatcher = null;
  }
});
</script>

<template>
  <RequireAuth>
    <NavBar />
    <UiPage>
      <div class="mesh-gradient dot-grid space-y-6">
        <UiHeading
          title="Supervisor Portal"
          subtitle="Monitor student guidance progress, validate revisions, and dispatch structured feedback."
        />

        <div v-if="loading" class="flex items-center justify-center py-20">
          <div class="h-7 w-7 animate-spin rounded-full border-[1.5px] border-amber-500 border-t-transparent" />
        </div>

        <template v-else>
          <div v-if="error" class="flex items-center gap-2.5 rounded-xl border border-red-500/15 bg-red-500/5 px-4 py-3">
            <AlertIcon :size="16" color="#ef4444" />
            <span class="text-sm text-red-400">{{ error }}</span>
          </div>

          <div class="grid grid-cols-2 gap-3 lg:grid-cols-4">
            <UiStatCard label="Active Students" :value="stats?.student_count ?? students.length" />
            <UiStatCard label="Pending Revisions" :value="stats?.pending_feedback ?? 0" />
            <UiStatCard
              label="Awaiting Validation"
              :value="allValidationQueue.length"
            />
            <UiStatCard
              label="Avg. Completion"
              :value="stats ? `${stats.completion_rate}%` : '0%'"
            />
          </div>

          <div class="flex gap-5 lg:h-[calc(100vh-340px)]">
            <UiCard class="flex flex-col h-full lg:w-[300px] shrink-0" :padding="false">
              <div class="border-b border-white/[0.04] p-4">
                <div class="flex items-center justify-between">
                  <h3 class="text-sm font-semibold text-stone-100 font-[family-name:var(--font-display)]">Student Roster</h3>
                  <span class="text-[11px] text-stone-600">{{ students.length }} students</span>
                </div>
              </div>
              <div class="flex-1 scroll-area">
                <div
                  v-for="student in students"
                  :key="student.id"
                  :class="[
                    'cursor-pointer border-b border-white/[0.04] p-3.5 transition-colors duration-100',
                    selectedStudentId === student.id
                      ? 'bg-amber-500/5 border-amber-500/20'
                      : 'hover:bg-white/[0.02]',
                  ]"
                  @click="selectStudent(student.id)"
                >
                  <div class="flex items-center gap-3">
                    <div
                      :class="[
                        'flex h-9 w-9 shrink-0 items-center justify-center rounded-lg text-xs font-bold',
                        selectedStudentId === student.id
                          ? 'bg-amber-500/15 text-amber-400'
                          : 'bg-[#1c1c1e] text-stone-500',
                      ]"
                    >
                      {{ studentInitials(student.name) }}
                    </div>
                    <div class="min-w-0 flex-1">
                      <p class="truncate text-sm font-medium text-stone-200">{{ student.name }}</p>
                      <p class="text-[11px] text-stone-600">{{ student.nim }}</p>
                    </div>
                    <UiBadge :text="studentStatusLabel(student.id)" :color="studentStatusColor(student.id)" />
                  </div>
                  <div class="mt-2 flex items-center gap-3 pl-12">
                    <span class="text-[10px] text-stone-600">
                      <span class="text-stone-300">{{ getStudentStats(student.id).sessions }}</span> ses.
                    </span>
                    <span class="text-[10px] text-amber-400/80">
                      {{ getStudentStats(student.id).pending }} pend
                    </span>
                    <span class="text-[10px] text-stone-500">
                      {{ getStudentStats(student.id).fixed }} val
                    </span>
                    <span class="text-[10px] text-emerald-400/80">
                      {{ getStudentStats(student.id).validated }} done
                    </span>
                  </div>
                </div>
                <div v-if="!students.length" class="p-6 text-center">
                  <p class="text-sm text-stone-600">No students assigned</p>
                </div>
              </div>
            </UiCard>

            <div class="min-w-0 flex-1 flex flex-col h-full">
              <div v-if="!selectedStudent" class="flex flex-1 items-center justify-center rounded-2xl border border-white/[0.04] bg-[#141415] p-6 h-full">
                <div class="text-center">
                  <div class="mx-auto mb-3 flex h-12 w-12 items-center justify-center rounded-xl bg-[#141415]">
                    <ProfileIcon :size="22" color="#52525b" />
                  </div>
                  <p class="text-sm text-stone-600">Select a student to view details</p>
                </div>
              </div>

              <template v-else>
                <UiCard>
                  <div class="p-5">
                    <div class="flex items-start gap-4">
                      <div class="flex h-14 w-14 shrink-0 items-center justify-center rounded-xl bg-amber-500/10 text-lg font-bold text-amber-400">
                        {{ studentInitials(selectedStudent.name) }}
                      </div>
                      <div class="min-w-0 flex-1">
                        <div class="flex items-center gap-2.5">
                          <h3 class="text-base font-semibold text-stone-100 font-[family-name:var(--font-display)]">{{ selectedStudent.name }}</h3>
                          <UiBadge :text="studentStatusLabel(selectedStudent.id)" :color="studentStatusColor(selectedStudent.id)" />
                        </div>
                        <p class="mt-0.5 text-[13px] text-stone-500">{{ selectedStudent.nim }} &middot; {{ selectedStudent.prodi }}</p>
                        <p class="mt-2 text-sm leading-relaxed text-stone-300">
                          {{ selectedStudent.thesis_title }}
                        </p>
                      </div>
                    </div>
                  </div>
                </UiCard>

                <div class="mt-3 flex gap-0.5 rounded-xl bg-[#141415] p-1 shrink-0">
                  <button
                    v-for="tab in (['overview', 'revisions', 'sessions', 'chat'] as const)"
                    :key="tab"
                    :class="[
                      'flex-1 rounded-lg px-4 py-2 text-[13px] font-medium transition-all duration-100',
                      activeTab === tab
                        ? 'bg-amber-500/10 text-amber-400'
                        : 'text-stone-600 hover:text-stone-300',
                    ]"
                    @click="activeTab = tab"
                  >
                    {{ tab === 'overview' ? 'Overview' : tab === 'revisions' ? 'Revisions' : tab === 'sessions' ? 'Sessions' : 'Advisor Chat' }}
                  </button>
                </div>

                <div class="flex-1 min-h-0 mt-3">
                  <div v-if="activeTab === 'overview'" class="h-full scroll-area pr-1 space-y-3">
                    <div class="grid grid-cols-4 gap-3">
                      <div class="rounded-xl border border-white/[0.04] bg-[#141415] p-4 text-center">
                        <p class="text-2xl font-bold text-stone-100">{{ getStudentStats(selectedStudent.id).sessions }}</p>
                        <p class="mt-1 text-[11px] text-stone-600">Sessions</p>
                      </div>
                      <div class="rounded-xl border border-white/[0.04] bg-[#141415] p-4 text-center">
                        <p class="text-2xl font-bold text-amber-400">{{ getStudentStats(selectedStudent.id).pending }}</p>
                        <p class="mt-1 text-[11px] text-stone-600">Pending</p>
                      </div>
                      <div class="rounded-xl border border-white/[0.04] bg-[#141415] p-4 text-center">
                        <p class="text-2xl font-bold text-stone-300">{{ getStudentStats(selectedStudent.id).fixed }}</p>
                        <p class="mt-1 text-[11px] text-stone-600">Awaiting</p>
                      </div>
                      <div class="rounded-xl border border-white/[0.04] bg-[#141415] p-4 text-center">
                        <p class="text-2xl font-bold text-emerald-400">{{ getStudentStats(selectedStudent.id).validated }}</p>
                        <p class="mt-1 text-[11px] text-stone-600">Validated</p>
                      </div>
                    </div>

                    <div class="rounded-xl border border-white/[0.04] bg-[#141415] p-4">
                      <div class="mb-2.5 flex items-center justify-between">
                        <p class="text-sm font-medium text-stone-300">Completion Progress</p>
                        <p class="text-sm font-bold text-amber-400">{{ completionPercent(selectedStudent.id) }}%</p>
                      </div>
                      <div class="h-2 overflow-hidden rounded-full bg-[#1c1c1e]">
                        <div
                          class="h-full rounded-full bg-amber-500 transition-all duration-700"
                          :style="{ width: `${completionPercent(selectedStudent.id)}%` }"
                        />
                      </div>
                    </div>

                    <div
                      v-if="selectedStudentLogs.length > 0"
                      class="rounded-xl border border-white/[0.04] bg-[#141415] p-4"
                    >
                      <p class="mb-3 text-sm font-medium text-stone-300">Latest Session</p>
                      <div>
                        <div
                          v-for="log in selectedStudentLogs.slice(0, 1)"
                          :key="log.id"
                        >
                          <div class="flex items-center justify-between">
                            <div class="flex items-center gap-2">
                              <p class="text-sm font-medium text-stone-200">{{ log.paper_filename }}</p>
                              <a
                                :href="`${API_URL}/storage/paper/${log.paper_filename}`"
                                target="_blank"
                                download
                                class="inline-flex h-6 w-6 items-center justify-center rounded bg-[#1c1c1e] text-stone-400 hover:bg-[#2c2c2e] hover:text-stone-200 transition-colors"
                                title="Download Initial Draft"
                              >
                                <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                                  <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
                                  <polyline points="7 10 12 15 17 10" />
                                  <line x1="12" y1="15" x2="12" y2="3" />
                                </svg>
                              </a>
                            </div>
                            <p class="text-[11px] text-stone-600">{{ formatDate(log.created_at) }}</p>
                          </div>
                          <p class="mt-2 rounded-lg bg-[#0a0a0b] p-3 text-xs leading-relaxed text-stone-500 font-mono">
                            {{ truncate(log.transcript_text, 300) }}
                          </p>
                          
                          <!-- Drafts / Final Drafts section in overview -->
                          <div v-if="log.revision_annotations && log.revision_annotations.length > 0" class="mt-3 space-y-2">
                            <p class="text-[10px] font-semibold uppercase tracking-wider text-stone-600">Drafts / Final Documents</p>
                            <div class="grid grid-cols-1 gap-1.5">
                              <div
                                v-for="ann in log.revision_annotations"
                                :key="ann.id"
                                class="flex items-center justify-between rounded-lg bg-[#0a0a0b] p-2 text-xs border border-white/[0.02]"
                              >
                                <span class="text-stone-300 truncate max-w-[200px]">{{ ann.filename }}</span>
                                <a
                                  :href="`${API_URL}/storage/annotations/${ann.filename}`"
                                  target="_blank"
                                  download
                                  class="inline-flex items-center gap-1 text-[11px] text-amber-400 hover:text-amber-300 transition-colors"
                                  title="Download Draft"
                                >
                                  <svg xmlns="http://www.w3.org/2000/svg" width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                                    <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
                                    <polyline points="7 10 12 15 17 10" />
                                    <line x1="12" y1="15" x2="12" y2="3" />
                                  </svg>
                                  Download
                                </a>
                              </div>
                            </div>
                          </div>
                        </div>
                      </div>
                    </div>
                  </div>

                  <div v-if="activeTab === 'revisions'" class="h-full scroll-area pr-1 space-y-3">
                    <div class="rounded-xl border border-white/[0.04] bg-[#141415] p-4">
                      <p class="mb-3 text-sm font-medium text-stone-300">Compose Feedback</p>
                      <textarea
                        v-model="feedbackText"
                        rows="3"
                        placeholder="Write feedback for this student..."
                        class="w-full resize-none rounded-xl border border-white/[0.04] bg-[#0a0a0b] px-4 py-3 text-sm text-stone-200 placeholder-zinc-500 transition-colors focus:border-white/[0.12] focus:outline-none focus:ring-1 focus:ring-white/10"
                      />
                      <div class="mt-3 flex items-center justify-end">
                        <UiButton title="Dispatch" :disabled="!feedbackText.trim()" @click="handleAddFeedback" />
                      </div>
                    </div>

                    <!-- Session filter pills -->
                    <div v-if="studentSessions.length > 1" class="flex flex-wrap gap-1.5">
                      <button
                        :class="[
                          'rounded-lg px-3 py-1.5 text-xs font-medium transition-colors',
                          sessionFilter === null
                            ? 'bg-amber-500/15 text-amber-400'
                            : 'bg-[#1c1c1e] text-stone-500 hover:text-stone-300',
                        ]"
                        @click="sessionFilter = null"
                      >
                        All Sessions
                        <span class="ml-1 text-[10px] opacity-60">
                          {{ selectedStudentLogs.flatMap(l => l.feedback_items ?? []).length }}
                        </span>
                      </button>
                      <button
                        v-for="s in studentSessions"
                        :key="s.log_id"
                        :class="[
                          'rounded-lg px-3 py-1.5 text-xs font-medium transition-colors',
                          sessionFilter === s.log_id
                            ? 'bg-amber-500/15 text-amber-400'
                            : 'bg-[#1c1c1e] text-stone-500 hover:text-stone-300',
                        ]"
                        @click="sessionFilter = s.log_id"
                      >
                        Session #{{ s.session_number }}
                        <span class="ml-1 text-[10px] opacity-60">
                          ({{ new Date(s.created_at).toLocaleDateString('en-US', { month: 'short', day: 'numeric' }) }}) [{{ s.feedback_count }}]
                        </span>
                      </button>
                    </div>

                    <div class="space-y-2">
                      <div
                        v-for="item in filteredRevisionItems"
                        :key="item.id"
                        class="rounded-xl border border-white/[0.04] bg-[#141415] p-4 transition-colors hover:border-white/[0.08]"
                      >
                        <div class="flex items-start justify-between gap-3">
                          <p class="flex-1 text-sm leading-relaxed text-stone-300">{{ item.content }}</p>
                          <div class="flex shrink-0 items-center gap-1.5">
                            <UiBadge
                              :text="item.category === 'Major' ? 'HOC' : 'LOC'"
                              :color="item.category === 'Major' ? 'bg-red-500/15 text-red-400 ring-red-500/30' : 'bg-stone-400/15 text-stone-400 ring-stone-500/30'"
                            />
                            <UiBadge
                              :text="item.status"
                              :color="
                                item.status === 'Validated'
                                  ? 'bg-emerald-500/15 text-emerald-400 ring-emerald-500/30'
                                  : item.status === 'Fixed'
                                    ? 'bg-zinc-400/15 text-stone-500 ring-zinc-500/30'
                                    : item.status === 'Rejected'
                                      ? 'bg-red-500/15 text-red-400 ring-red-500/30'
                                      : 'bg-amber-500/15 text-amber-400 ring-amber-500/30'
                              "
                            />
                          </div>
                        </div>
                        <div class="mt-3 flex items-center gap-2">
                          <button
                            v-if="item.status === 'Fixed'"
                            class="inline-flex items-center gap-1.5 rounded-md bg-emerald-500/10 px-3 py-1.5 text-xs font-medium text-emerald-400 transition-colors hover:bg-emerald-500/15"
                            :disabled="validatingId === item.id"
                            @click="handleValidate(item.id)"
                          >
                            <CheckCircleIcon :size="13" />
                            Approve
                          </button>
                          <button
                            v-if="item.status === 'Fixed'"
                            class="inline-flex items-center gap-1.5 rounded-md bg-red-500/10 px-3 py-1.5 text-xs font-medium text-red-400 transition-colors hover:bg-red-500/15"
                            :disabled="validatingId === item.id"
                            @click="openRejectionDialog(item.id)"
                          >
                            <AlertIcon :size="13" />
                            Reject
                          </button>
                          <button
                            v-if="item.status === 'Validated'"
                            class="inline-flex items-center gap-1.5 rounded-md bg-[#1c1c1e] px-3 py-1.5 text-xs font-medium text-stone-500 transition-colors hover:bg-[#2c2c2e]"
                            :disabled="validatingId === item.id"
                            @click="handleUndo(item.id)"
                          >
                            Undo
                          </button>
                        </div>
                      </div>
                      <div v-if="!filteredRevisionItems.length" class="py-8 text-center">
                        <p class="text-sm text-stone-600">No feedback items yet</p>
                      </div>
                    </div>
                  </div>

                  <div v-if="activeTab === 'sessions'" class="h-full scroll-area pr-1 space-y-2">
                    <div class="space-y-2">
                      <div
                        v-for="log in selectedStudentLogs"
                        :key="log.id"
                        class="rounded-xl border border-white/[0.04] bg-[#141415] p-4 transition-colors hover:border-white/[0.08]"
                      >
                        <div class="flex items-center justify-between">
                          <div class="flex items-center gap-3">
                            <div class="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-[#1c1c1e]">
                              <ConsultationIcon :size="16" color="#f59e0b" />
                            </div>
                            <div>
                              <div class="flex items-center gap-2">
                                <p class="text-sm font-medium text-stone-200">{{ log.paper_filename }}</p>
                                <a
                                  :href="`${API_URL}/storage/paper/${log.paper_filename}`"
                                  target="_blank"
                                  download
                                  class="inline-flex h-5 w-5 items-center justify-center rounded bg-[#1c1c1e] text-stone-400 hover:bg-[#2c2c2e] hover:text-stone-200 transition-colors"
                                  title="Download Initial Draft"
                                >
                                  <svg xmlns="http://www.w3.org/2000/svg" width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                                    <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
                                    <polyline points="7 10 12 15 17 10" />
                                    <line x1="12" y1="15" x2="12" y2="3" />
                                  </svg>
                                </a>
                              </div>
                              <p class="text-[11px] text-stone-600">{{ formatDate(log.created_at) }}</p>
                            </div>
                          </div>
                          <UiBadge
                            :text="`${log.feedback_items?.length ?? 0} feedback`"
                            color="bg-[#1c1c1e] text-stone-500"
                          />
                        </div>
                        <p class="mt-3 rounded-lg bg-[#0a0a0b] p-3 text-xs leading-relaxed text-stone-500 font-mono">
                          {{ truncate(log.transcript_text, 200) }}
                        </p>
                        <div class="mt-3 flex flex-wrap gap-1.5">
                          <span
                            v-if="log.audio_filename"
                            class="inline-flex items-center rounded-md bg-[#1c1c1e] px-2.5 py-1 text-[11px] font-medium text-stone-500"
                          >
                            Audio
                          </span>
                          <span
                            v-if="log.transcript_filename"
                            class="inline-flex items-center rounded-md bg-[#1c1c1e] px-2.5 py-1 text-[11px] font-medium text-stone-500"
                          >
                            Transcript
                          </span>
                          <a
                            v-for="ann in log.revision_annotations ?? []"
                            :key="ann.id"
                            :href="`${API_URL}/storage/annotations/${ann.filename}`"
                            target="_blank"
                            download
                            class="inline-flex items-center gap-1 rounded-md bg-amber-500/10 px-2.5 py-1 text-[11px] font-medium text-amber-400 hover:bg-amber-500/20 transition-colors"
                            title="Download Final Draft/Revision"
                          >
                            <svg xmlns="http://www.w3.org/2000/svg" width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                              <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
                              <polyline points="7 10 12 15 17 10" />
                              <line x1="12" y1="15" x2="12" y2="3" />
                            </svg>
                            {{ ann.filename }}
                          </a>
                        </div>
                      </div>
                      <div v-if="!selectedStudentLogs.length" class="py-8 text-center">
                        <p class="text-sm text-stone-600">No sessions recorded</p>
                      </div>
                    </div>
                  </div>

                  <div v-if="activeTab === 'chat'" class="h-full">
                    <div class="flex flex-col h-full overflow-hidden rounded-xl border border-white/[0.04] bg-[#141415]">
                      <div class="border-b border-white/[0.04] px-4 py-3">
                        <p class="text-sm font-semibold text-stone-200">
                          Chat with {{ selectedStudent.name }}
                        </p>
                      </div>
                      <div ref="chatContainer" class="flex-1 space-y-2.5 scroll-area p-4">
                        <div
                          v-for="msg in sortedDirectMessages"
                          :key="msg.id"
                          :class="[
                            'flex',
                            msg.sender_role === 'lecturer' ? 'justify-end' : 'justify-start',
                          ]"
                        >
                          <div
                            :class="[
                              'max-w-[75%] rounded-2xl px-4 py-2.5 text-sm leading-relaxed',
                              msg.sender_role === 'lecturer'
                                ? 'rounded-br-md bg-amber-500/10 text-amber-100'
                                : 'rounded-bl-md bg-[#1c1c1e] text-stone-300',
                            ]"
                          >
                            <p>{{ msg.content }}</p>
                            <p class="mt-1 text-[10px] text-stone-700">{{ formatTime(msg.created_at) }}</p>
                          </div>
                        </div>
                        <div v-if="!sortedDirectMessages.length" class="flex h-full items-center justify-center">
                          <p class="text-sm text-stone-600">No messages yet.</p>
                        </div>
                      </div>
                      <div class="border-t border-white/[0.04] p-3">
                        <div class="flex gap-2">
                          <input
                            v-model="chatInput"
                            type="text"
                            placeholder="Type a message..."
                            class="flex-1 rounded-xl border border-white/[0.04] bg-[#0a0a0b] px-4 py-2.5 text-sm text-stone-200 placeholder-zinc-500 transition-colors focus:border-white/[0.12] focus:outline-none focus:ring-1 focus:ring-white/10"
                            @keyup.enter="sendDirectMessage"
                          />
                          <button
                            :disabled="!chatInput.trim() || chatLoading"
                            class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-amber-500 text-[#0a0a0b] transition-all hover:bg-amber-400 disabled:opacity-40"
                            @click="sendDirectMessage"
                          >
                            <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                              <path d="M22 2L11 13" />
                              <path d="M22 2L15 22L11 13L2 9L22 2Z" />
                            </svg>
                          </button>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              </template>
            </div>
          </div>

          <div v-if="allValidationQueue.length" class="space-y-4">
            <div class="flex items-center justify-between">
              <h3 class="text-sm font-semibold text-stone-100 font-[family-name:var(--font-display)]">Validation Queue</h3>
              <UiBadge
                :text="`${allValidationQueue.length} items`"
                color="bg-zinc-400/15 text-stone-500"
              />
            </div>
            <div class="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
              <div
                v-for="item in allValidationQueue"
                :key="item.id"
                class="rounded-xl border border-white/[0.04] bg-[#141415] p-4 transition-colors hover:border-white/[0.08]"
              >
                <p class="text-sm leading-relaxed text-stone-300">{{ item.content }}</p>
                <div class="mt-3 flex items-center justify-between">
                  <div class="flex items-center gap-1.5">
                    <UiBadge
                      :text="item.category === 'Major' ? 'HOC' : 'LOC'"
                      :color="item.category === 'Major' ? 'bg-red-500/15 text-red-400 ring-red-500/30' : 'bg-stone-400/15 text-stone-400 ring-stone-500/30'"
                    />
                    <UiBadge text="Fixed" color="bg-zinc-400/15 text-stone-500 ring-zinc-500/30" />
                  </div>
                  <div class="flex items-center gap-1.5">
                    <button
                      class="inline-flex items-center gap-1 rounded-md bg-emerald-500/10 px-2.5 py-1.5 text-[11px] font-medium text-emerald-400 transition-colors hover:bg-emerald-500/15"
                      :disabled="validatingId === item.id"
                      @click="handleValidate(item.id)"
                    >
                      <CheckCircleIcon :size="12" />
                      Approve
                    </button>
                    <button
                      class="inline-flex items-center gap-1 rounded-md bg-red-500/10 px-2.5 py-1.5 text-[11px] font-medium text-red-400 transition-colors hover:bg-red-500/15"
                      :disabled="validatingId === item.id"
                      @click="openRejectionDialog(item.id)"
                    >
                      <AlertIcon :size="12" />
                      Reject
                    </button>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </template>
      </div>

      <div class="pointer-events-none fixed right-4 top-20 z-[100] flex flex-col gap-2">
        <TransitionGroup
          enter-active-class="transition-all duration-300 ease-out"
          leave-active-class="transition-all duration-200 ease-in"
          enter-from-class="translate-x-full opacity-0"
          enter-to-class="translate-x-0 opacity-100"
          leave-from-class="translate-x-0 opacity-100"
          leave-to-class="translate-x-full opacity-0"
        >
          <div
            v-for="toast in toasts"
            :key="toast.id"
            :class="[
              'pointer-events-auto w-80 rounded-xl border px-4 py-3 shadow-xl backdrop-blur-xl',
              toast.type === 'success'
                ? 'border-emerald-500/15 bg-emerald-500/10'
                : toast.type === 'error'
                  ? 'border-red-500/15 bg-red-500/10'
                  : toast.type === 'warning'
                    ? 'border-amber-500/15 bg-amber-500/10'
                    : 'border-amber-500/15 bg-amber-500/10',
            ]"
          >
            <p
              :class="[
                'text-sm font-medium',
                toast.type === 'success'
                  ? 'text-emerald-400'
                  : toast.type === 'error'
                    ? 'text-red-400'
                    : toast.type === 'warning'
                      ? 'text-amber-400'
                      : 'text-amber-400',
              ]"
            >
              {{ toast.message }}
            </p>
          </div>
        </TransitionGroup>
      </div>
      <RejectionDialog
        :open="rejectionDialogOpen"
        title="Reject Revision"
        message="Are you sure you want to reject this revision? Please provide an explanation for the student."
        confirm-label="Reject Revision"
        cancel-label="Cancel"
        @confirm="handleRejectConfirm"
        @cancel="rejectionDialogOpen = false"
      />
    </UiPage>
  </RequireAuth>
</template>
