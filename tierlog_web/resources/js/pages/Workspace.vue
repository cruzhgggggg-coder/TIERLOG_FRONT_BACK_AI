<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, nextTick, watch } from 'vue';
import { useAuthStore } from '@/stores/auth';
import { useWorkspaceStore } from '@/stores/workspace';
import type { StudentProfile, ConsultationLog, FeedbackItem, DashboardStats, SessionInfo } from '@/types';
import RequireAuth from '@/components/RequireAuth.vue';
import NavBar from '@/components/NavBar.vue';
import WorkspaceWindow from '@/components/WorkspaceWindow.vue';
import RejectionDialog from '@/components/RejectionDialog.vue';
import UiBadge from '@/components/UiBadge.vue';
import UiButton from '@/components/UiButton.vue';
import { AlertIcon, ConsultationIcon, ProfileIcon, CheckCircleIcon, ClockIcon, BellIcon } from '@/components/icons';

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
const store = useWorkspaceStore();

const loading = ref(true);
const error = ref('');
const students = ref<StudentProfile[]>([]);
const consultations = ref<ConsultationLog[]>([]);
const directMessages = ref<DirectMessage[]>([]);
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

const canvasRef = ref<HTMLDivElement | null>(null);
const canvasWidth = ref(1200);
const canvasHeight = ref(650);

// Computed variables mapping from Pinia store
const panels = computed(() => store.panels);
const selectedStudentId = computed({
  get: () => store.selectedStudentId,
  set: (val) => store.setSelectedStudentId(val),
});

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

const filteredRevisionItems = computed(() => {
  const allItems = selectedStudentLogs.value.flatMap((l) => l.feedback_items ?? []);
  if (sessionFilter.value === null) return allItems;
  const targetLog = consultations.value.find((l) => l.id === sessionFilter.value);
  return targetLog?.feedback_items ?? [];
});

// Canvas resizing tracker
function updateCanvasSize() {
  if (canvasRef.value) {
    canvasWidth.value = canvasRef.value.clientWidth;
    canvasHeight.value = canvasRef.value.clientHeight;
  }
}

function handleTilePanels() {
  updateCanvasSize();
  store.tilePanels(canvasWidth.value, canvasHeight.value);
}

// ── Toasts system ──
function addToast(message: string, type: Toast['type'] = 'info') {
  const id = ++toastCounter;
  toasts.value.push({ id, message, type });
  setTimeout(() => {
    toasts.value = toasts.value.filter((t) => t.id !== id);
  }, 4500);
}

// ── Data Fetching ──
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

function selectStudent(id: number) {
  selectedStudentId.value = id;
  store.raisePanel('roster');
}

// ── Validation Queue Operations ──
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
      comment: comment,
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

// ── Direct Chat ──
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
    nextTick(scrollChatToBottom);
  }
  chatLoading.value = false;
}

const chatContainer = ref<HTMLDivElement | null>(null);
function scrollChatToBottom() {
  if (chatContainer.value) chatContainer.value.scrollTop = chatContainer.value.scrollHeight;
}

// ── WebSocket Connection ──
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

// ── Session Deletion ──
async function deleteSession(logId: number) {
  if (!confirm('Are you sure you want to delete this session? This will physically clear files from the disk and cannot be undone.')) return;
  const res = await auth.api(`/consultations/${logId}`, { method: 'DELETE' });
  if (res.ok) {
    addToast('Session deleted successfully', 'success');
    await fetchAllData();
  } else {
    addToast(res.error || 'Failed to delete session', 'error');
  }
}

watch(selectedStudentId, (id) => {
  if (id) {
    sessionFilter.value = null;
    fetchStudentSessions(id);
    const log = consultations.value.find((l) => l.student_id === id);
    if (log) {
      loadDirectMessages(log.id);
      store.activeLogId = log.id;
    } else {
      directMessages.value = [];
      store.activeLogId = null;
    }
  }
});

onMounted(() => {
  fetchAllData();
  connectWebSocket();
  updateCanvasSize();
  window.addEventListener('resize', updateCanvasSize);
});

onUnmounted(() => {
  ws?.close();
  ws = null;
  if (stopConsultationWatcher) {
    stopConsultationWatcher();
    stopConsultationWatcher = null;
  }
  window.removeEventListener('resize', updateCanvasSize);
});
</script>

<template>
  <RequireAuth>
    <NavBar />
    
    <!-- Real-time Toast Notifications -->
    <div class="fixed top-20 right-6 z-[999] flex flex-col gap-3 max-w-sm w-full pointer-events-none px-4 sm:px-0">
      <transition-group name="toast">
        <div
          v-for="toast in toasts"
          :key="toast.id"
          class="pointer-events-auto flex items-center justify-between gap-3.5 rounded-xl border p-4 shadow-2xl backdrop-blur-md transition-all duration-300"
          :class="[
            toast.type === 'success' ? 'bg-[#064e3b]/90 border-[#10b981]/30 text-emerald-400' :
            toast.type === 'warning' ? 'bg-[#78350f]/90 border-[#d97706]/30 text-amber-400' :
            toast.type === 'error' ? 'bg-[#7f1d1d]/90 border-[#ef4444]/30 text-red-400' :
            'bg-[#1e3a8a]/90 border-[#3b82f6]/30 text-blue-400'
          ]"
        >
          <div class="flex items-center gap-3">
            <span class="shrink-0">
              <CheckCircleIcon v-if="toast.type === 'success'" :size="18" />
              <AlertIcon v-else-if="toast.type === 'error' || toast.type === 'warning'" :size="18" />
              <BellIcon v-else :size="18" />
            </span>
            <p class="text-sm font-medium text-stone-100">{{ toast.message }}</p>
          </div>
        </div>
      </transition-group>
    </div>

    <!-- Workspace View -->
    <div class="flex flex-col h-[calc(100vh-56px)] bg-[#070708] overflow-hidden select-none">
      
      <!-- Toolbar Menu -->
      <div class="flex shrink-0 items-center justify-between border-b border-white/[0.04] bg-[#0c0c0d]/90 px-6 py-2">
        <div class="flex items-center gap-3">
          <span class="text-xs font-semibold uppercase tracking-wider text-stone-500 font-[family-name:var(--font-display)]">Workspace Panels</span>
          <div class="flex items-center gap-1 bg-[#141415] rounded-lg p-0.5 border border-white/[0.02]">
            <button
              v-for="panel in panels"
              :key="panel.id"
              :class="[
                'rounded px-2.5 py-1 text-[11px] font-medium transition-all duration-150',
                panel.visible
                  ? 'bg-amber-500/10 text-amber-400 border border-amber-500/20'
                  : 'text-stone-500 hover:text-stone-300 border border-transparent'
              ]"
              @click="store.togglePanel(panel.id)"
            >
              {{ panel.title }}
            </button>
          </div>
        </div>
        
        <div class="flex items-center gap-2">
          <UiButton
            title="Tile Windows"
            tone="secondary"
            class="!py-1 !px-3 !text-xs"
            @click="handleTilePanels"
          >
            <template #default>
              <div class="flex items-center gap-1.5">
                <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <rect x="3" y="3" width="7" height="9" />
                  <rect x="14" y="3" width="7" height="5" />
                  <rect x="14" y="12" width="7" height="9" />
                  <rect x="3" y="16" width="7" height="5" />
                </svg>
                Tile Windows
              </div>
            </template>
          </UiButton>
        </div>
      </div>

      <!-- Main Absolute Canvas -->
      <div ref="canvasRef" class="flex-1 relative overflow-hidden bg-[#070708] dot-grid">
        
        <div v-if="loading" class="flex items-center justify-center absolute inset-0 bg-[#070708]/50 z-50">
          <div class="h-7 w-7 animate-spin rounded-full border-[1.5px] border-amber-500 border-t-transparent" />
        </div>

        <div v-if="error" class="absolute top-4 left-4 right-4 z-50 flex items-center gap-2.5 rounded-xl border border-red-500/15 bg-red-500/5 px-4 py-3">
          <AlertIcon :size="16" color="#ef4444" />
          <span class="text-sm text-red-400">{{ error }}</span>
        </div>

        <!-- Render Workspace Windows -->
        <template v-if="!loading">
          
          <!-- PANEL: Student Roster -->
          <WorkspaceWindow
            v-if="panels.find(p => p.id === 'roster')?.visible"
            id="roster"
            title="Student Roster"
            :x="panels.find(p => p.id === 'roster')!.x"
            :y="panels.find(p => p.id === 'roster')!.y"
            :w="panels.find(p => p.id === 'roster')!.w"
            :h="panels.find(p => p.id === 'roster')!.h"
            :z-index="panels.find(p => p.id === 'roster')!.zIndex"
            :is-maximized="panels.find(p => p.id === 'roster')!.isMaximized"
            @update:position="store.updatePanelPosition('roster', $event.x, $event.y)"
            @update:size="store.updatePanelSize('roster', $event.w, $event.h)"
            @close="store.togglePanel('roster')"
            @maximize="store.maximizePanel('roster')"
            @focus="store.raisePanel('roster')"
          >
            <div class="flex flex-col h-full bg-[#111112]">
              <div class="border-b border-white/[0.04] p-3 flex items-center justify-between">
                <span class="text-xs font-semibold text-stone-400">Total Supervised</span>
                <span class="text-[11px] bg-white/5 text-stone-400 px-1.5 py-0.5 rounded">{{ students.length }} students</span>
              </div>
              <div class="flex-1 overflow-y-auto">
                <div
                  v-for="student in students"
                  :key="student.id"
                  :class="[
                    'cursor-pointer border-b border-white/[0.04] p-3 transition-colors duration-100',
                    selectedStudentId === student.id ? 'bg-amber-500/5 border-amber-500/20' : 'hover:bg-white/[0.02]',
                  ]"
                  @click="selectStudent(student.id)"
                >
                  <div class="flex items-center gap-3">
                    <div
                      :class="[
                        'flex h-8 w-8 shrink-0 items-center justify-center rounded-lg text-xs font-bold',
                        selectedStudentId === student.id ? 'bg-amber-500/15 text-amber-400' : 'bg-[#1c1c1e] text-stone-500',
                      ]"
                    >
                      {{ studentInitials(student.name) }}
                    </div>
                    <div class="min-w-0 flex-1">
                      <p class="truncate text-xs font-semibold text-stone-200">{{ student.name }}</p>
                      <p class="text-[10px] text-stone-600 font-mono">{{ student.nim }}</p>
                    </div>
                    <UiBadge :text="studentStatusLabel(student.id)" :color="studentStatusColor(student.id)" />
                  </div>
                  <div class="mt-2 flex items-center justify-between text-[9px] text-stone-600 pl-11">
                    <span>{{ getStudentStats(student.id).sessions }} sessions</span>
                    <span class="text-amber-400/80">{{ getStudentStats(student.id).pending }} pending</span>
                    <span class="text-emerald-400/80">{{ getStudentStats(student.id).validated }} done</span>
                  </div>
                </div>
              </div>
            </div>
          </WorkspaceWindow>

          <!-- PANEL: Session History -->
          <WorkspaceWindow
            v-if="panels.find(p => p.id === 'history')?.visible"
            id="history"
            title="Session History"
            :x="panels.find(p => p.id === 'history')!.x"
            :y="panels.find(p => p.id === 'history')!.y"
            :w="panels.find(p => p.id === 'history')!.w"
            :h="panels.find(p => p.id === 'history')!.h"
            :z-index="panels.find(p => p.id === 'history')!.zIndex"
            :is-maximized="panels.find(p => p.id === 'history')!.isMaximized"
            @update:position="store.updatePanelPosition('history', $event.x, $event.y)"
            @update:size="store.updatePanelSize('history', $event.w, $event.h)"
            @close="store.togglePanel('history')"
            @maximize="store.maximizePanel('history')"
            @focus="store.raisePanel('history')"
          >
            <div class="flex flex-col h-full bg-[#111112]">
              <div v-if="!selectedStudent" class="flex-1 flex items-center justify-center p-6 text-center text-stone-600">
                <p class="text-xs">Select a student from roster to see history</p>
              </div>
              <template v-else>
                <div class="border-b border-white/[0.04] p-3 flex items-center justify-between">
                  <span class="text-xs font-semibold text-stone-400 truncate max-w-[200px]">{{ selectedStudent.name }}'s Logs</span>
                  <span class="text-[10px] text-stone-600 font-mono">{{ selectedStudentLogs.length }} sessions</span>
                </div>
                <div class="flex-1 overflow-y-auto p-3 space-y-3">
                  <div
                    v-for="log in selectedStudentLogs"
                    :key="log.id"
                    class="rounded-xl border border-white/[0.04] bg-[#141415] p-3.5 hover:border-white/[0.08] transition-colors relative group"
                  >
                    <div class="flex items-start justify-between gap-2">
                      <div class="flex items-center gap-2">
                        <ConsultationIcon :size="14" color="#f59e0b" />
                        <span class="text-xs font-semibold text-stone-200 truncate max-w-[180px]">{{ log.paper_filename }}</span>
                        <a
                          :href="`${API_URL}/storage/paper/${log.paper_filename}`"
                          target="_blank"
                          download
                          class="inline-flex h-5 w-5 items-center justify-center rounded bg-[#1c1c1e] text-stone-400 hover:bg-[#2c2c2e] hover:text-stone-200 transition-colors"
                          title="Download Draft"
                        >
                          <svg xmlns="http://www.w3.org/2000/svg" width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                            <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
                            <polyline points="7 10 12 15 17 10" />
                            <line x1="12" y1="15" x2="12" y2="3" />
                          </svg>
                        </a>
                      </div>
                      <span class="text-[10px] text-stone-600 font-mono">{{ formatDate(log.created_at) }}</span>
                    </div>
                    <p class="mt-2 rounded bg-[#0a0a0b] p-2.5 text-[11px] leading-relaxed text-stone-500 font-mono">
                      {{ truncate(log.transcript_text, 140) }}
                    </p>
                    <div class="mt-2.5 flex items-center justify-between border-t border-white/[0.02] pt-2">
                      <UiBadge :text="`${log.feedback_items?.length ?? 0} feedback`" color="bg-[#1c1c1e] text-stone-500" />
                      <button
                        class="text-[10px] text-red-500 hover:text-red-400 transition-colors opacity-0 group-hover:opacity-100 flex items-center gap-1"
                        @click="deleteSession(log.id)"
                      >
                        Delete
                      </button>
                    </div>
                  </div>
                  <div v-if="!selectedStudentLogs.length" class="text-center py-10 text-xs text-stone-600">
                    No sessions submitted yet.
                  </div>
                </div>
              </template>
            </div>
          </WorkspaceWindow>

          <!-- PANEL: Feedback Composer -->
          <WorkspaceWindow
            v-if="panels.find(p => p.id === 'feedback')?.visible"
            id="feedback"
            title="Feedback Composer"
            :x="panels.find(p => p.id === 'feedback')!.x"
            :y="panels.find(p => p.id === 'feedback')!.y"
            :w="panels.find(p => p.id === 'feedback')!.w"
            :h="panels.find(p => p.id === 'feedback')!.h"
            :z-index="panels.find(p => p.id === 'feedback')!.zIndex"
            :is-maximized="panels.find(p => p.id === 'feedback')!.isMaximized"
            @update:position="store.updatePanelPosition('feedback', $event.x, $event.y)"
            @update:size="store.updatePanelSize('feedback', $event.w, $event.h)"
            @close="store.togglePanel('feedback')"
            @maximize="store.maximizePanel('feedback')"
            @focus="store.raisePanel('feedback')"
          >
            <div class="flex flex-col h-full bg-[#111112]">
              <div v-if="!selectedStudent" class="flex-1 flex items-center justify-center p-6 text-center text-stone-600">
                <p class="text-xs">Select a student from roster to add feedback</p>
              </div>
              <template v-else>
                <div class="border-b border-white/[0.04] p-3 flex flex-wrap gap-2 items-center justify-between">
                  <span class="text-xs font-semibold text-stone-400">Revisions Feedback</span>
                  <div v-if="studentSessions.length > 1" class="flex items-center gap-1.5 max-w-full">
                    <select
                      v-model="sessionFilter"
                      class="bg-[#1c1c1e] text-[10px] text-stone-300 border border-white/[0.06] rounded px-1.5 py-0.5 focus:outline-none"
                    >
                      <option :value="null">All Sessions</option>
                      <option v-for="s in studentSessions" :key="s.log_id" :value="s.log_id">
                        Session #{{ s.session_number }}
                      </option>
                    </select>
                  </div>
                </div>

                <div class="flex-1 overflow-y-auto p-3 space-y-3">
                  <!-- Quick composer box -->
                  <div class="rounded-xl border border-white/[0.04] bg-[#141415] p-3">
                    <textarea
                      v-model="feedbackText"
                      rows="2"
                      placeholder="Compose revision instruction..."
                      class="w-full resize-none rounded-lg border border-white/[0.04] bg-[#0a0a0b] px-3 py-2 text-xs text-stone-200 placeholder-zinc-500 focus:border-white/[0.12] focus:outline-none"
                    />
                    <div class="mt-2 flex items-center justify-end">
                      <UiButton
                        title="Dispatch"
                        :disabled="!feedbackText.trim() || !latestLog"
                        class="!py-1 !px-3 !text-xs"
                        @click="handleAddFeedback"
                      />
                    </div>
                  </div>

                  <!-- Revision Checklist -->
                  <div class="space-y-2">
                    <div
                      v-for="item in filteredRevisionItems"
                      :key="item.id"
                      class="rounded-xl border border-white/[0.04] bg-[#141415] p-3.5 hover:border-white/[0.08]"
                    >
                      <div class="flex items-start justify-between gap-3">
                        <p class="flex-1 text-xs leading-relaxed text-stone-300 font-medium">{{ item.content }}</p>
                        <div class="flex shrink-0 items-center gap-1.5">
                          <UiBadge
                            :text="item.category === 'Major' ? 'HOC' : 'LOC'"
                            :color="item.category === 'Major' ? 'bg-red-500/15 text-red-400 ring-red-500/30' : 'bg-stone-400/15 text-stone-400 ring-stone-500/30'"
                          />
                          <UiBadge
                            :text="item.status"
                            :color="
                              item.status === 'Validated' ? 'bg-emerald-500/15 text-emerald-400 ring-emerald-500/30' :
                              item.status === 'Fixed' ? 'bg-zinc-400/15 text-stone-500 ring-zinc-500/30' :
                              item.status === 'Rejected' ? 'bg-red-500/15 text-red-400 ring-red-500/30' :
                              'bg-amber-500/15 text-amber-400 ring-amber-500/30'
                            "
                          />
                        </div>
                      </div>

                      <div class="mt-2.5 flex items-center gap-1.5 border-t border-white/[0.02] pt-2">
                        <button
                          v-if="item.status === 'Fixed'"
                          class="inline-flex items-center gap-1 rounded bg-emerald-500/10 px-2 py-1 text-[10px] font-medium text-emerald-400 hover:bg-emerald-500/15"
                          :disabled="validatingId === item.id"
                          @click="handleValidate(item.id)"
                        >
                          <CheckCircleIcon :size="11" />
                          Approve
                        </button>
                        <button
                          v-if="item.status === 'Fixed'"
                          class="inline-flex items-center gap-1 rounded bg-red-500/10 px-2 py-1 text-[10px] font-medium text-red-400 hover:bg-red-500/15"
                          :disabled="validatingId === item.id"
                          @click="openRejectionDialog(item.id)"
                        >
                          <AlertIcon :size="11" />
                          Reject
                        </button>
                        <button
                          v-if="item.status === 'Validated'"
                          class="inline-flex items-center gap-1 rounded bg-[#1c1c1e] px-2 py-1 text-[10px] font-medium text-stone-500 hover:bg-[#2c2c2e]"
                          :disabled="validatingId === item.id"
                          @click="handleUndo(item.id)"
                        >
                          Undo
                        </button>
                      </div>
                    </div>
                    <div v-if="!filteredRevisionItems.length" class="text-center py-6 text-xs text-stone-600">
                      No feedback instructions generated.
                    </div>
                  </div>
                </div>
              </template>
            </div>
          </WorkspaceWindow>

          <!-- PANEL: Advisor Chat -->
          <WorkspaceWindow
            v-if="panels.find(p => p.id === 'chat')?.visible"
            id="chat"
            title="Advisor Chat"
            :x="panels.find(p => p.id === 'chat')!.x"
            :y="panels.find(p => p.id === 'chat')!.y"
            :w="panels.find(p => p.id === 'chat')!.w"
            :h="panels.find(p => p.id === 'chat')!.h"
            :z-index="panels.find(p => p.id === 'chat')!.zIndex"
            :is-maximized="panels.find(p => p.id === 'chat')!.isMaximized"
            @update:position="store.updatePanelPosition('chat', $event.x, $event.y)"
            @update:size="store.updatePanelSize('chat', $event.w, $event.h)"
            @close="store.togglePanel('chat')"
            @maximize="store.maximizePanel('chat')"
            @focus="store.raisePanel('chat')"
          >
            <div class="flex flex-col h-full bg-[#111112]">
              <div v-if="!selectedStudent" class="flex-1 flex items-center justify-center p-6 text-center text-stone-600">
                <p class="text-xs">Select a student from roster to start chatting</p>
              </div>
              <template v-else>
                <div class="border-b border-white/[0.04] p-3">
                  <span class="text-xs font-semibold text-stone-400">Direct Chat Room</span>
                </div>
                
                <!-- Messages List -->
                <div ref="chatContainer" class="flex-1 overflow-y-auto p-3 space-y-2">
                  <div
                    v-for="msg in sortedDirectMessages"
                    :key="msg.id"
                    :class="[
                      'flex flex-col max-w-[85%] rounded-xl p-2.5 text-xs',
                      msg.sender_role === 'lecturer'
                        ? 'ml-auto bg-amber-500/10 border border-amber-500/20 text-stone-100'
                        : 'bg-[#1c1c1e] text-stone-300'
                    ]"
                  >
                    <p class="leading-relaxed whitespace-pre-wrap select-text">{{ msg.content }}</p>
                    <span class="text-[8px] text-stone-600 mt-1 self-end font-mono">
                      {{ new Date(msg.created_at).toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' }) }}
                    </span>
                  </div>
                  <div v-if="!sortedDirectMessages.length" class="text-center py-10 text-xs text-stone-600">
                    No messages yet. Send a message to start conversation.
                  </div>
                </div>

                <!-- Input Box -->
                <div class="p-3 border-t border-white/[0.04] bg-[#0c0c0d] flex gap-2">
                  <input
                    v-model="chatInput"
                    type="text"
                    placeholder="Type message..."
                    class="flex-1 bg-[#141415] border border-white/[0.04] rounded-lg px-3 py-2 text-xs text-stone-100 placeholder-stone-600 focus:outline-none focus:border-amber-500/30"
                    @keyup.enter="sendDirectMessage"
                  />
                  <UiButton
                    title="Send"
                    :disabled="!chatInput.trim() || chatLoading"
                    class="!py-1.5 !px-3.5 !text-xs shrink-0"
                    @click="sendDirectMessage"
                  />
                </div>
              </template>
            </div>
          </WorkspaceWindow>

          <!-- PANEL: Validation Queue -->
          <WorkspaceWindow
            v-if="panels.find(p => p.id === 'queue')?.visible"
            id="queue"
            title="Validation Queue"
            :x="panels.find(p => p.id === 'queue')!.x"
            :y="panels.find(p => p.id === 'queue')!.y"
            :w="panels.find(p => p.id === 'queue')!.w"
            :h="panels.find(p => p.id === 'queue')!.h"
            :z-index="panels.find(p => p.id === 'queue')!.zIndex"
            :is-maximized="panels.find(p => p.id === 'queue')!.isMaximized"
            @update:position="store.updatePanelPosition('queue', $event.x, $event.y)"
            @update:size="store.updatePanelSize('queue', $event.w, $event.h)"
            @close="store.togglePanel('queue')"
            @maximize="store.maximizePanel('queue')"
            @focus="store.raisePanel('queue')"
          >
            <div class="flex flex-col h-full bg-[#111112]">
              <div class="border-b border-white/[0.04] p-3 flex items-center justify-between">
                <span class="text-xs font-semibold text-stone-400">Global Awaiting Queue</span>
                <span class="text-[10px] bg-amber-500/10 text-amber-400 px-1.5 py-0.5 rounded font-mono">{{ allValidationQueue.length }} tasks</span>
              </div>
              <div class="flex-1 overflow-y-auto p-3 space-y-2">
                <div
                  v-for="item in allValidationQueue"
                  :key="item.id"
                  class="rounded-xl border border-white/[0.04] bg-[#141415] p-3 hover:border-white/[0.08]"
                >
                  <div class="flex items-start justify-between gap-3">
                    <div>
                      <span class="text-[9px] uppercase tracking-wider text-amber-500 font-mono">Student Revision</span>
                      <p class="text-xs leading-relaxed text-stone-300 font-medium mt-0.5">{{ item.content }}</p>
                    </div>
                    <div class="flex shrink-0 items-center gap-1.5">
                      <UiBadge :text="item.category === 'Major' ? 'HOC' : 'LOC'" />
                    </div>
                  </div>
                  <div class="mt-2.5 flex items-center justify-end gap-1.5 border-t border-white/[0.02] pt-2">
                    <button
                      class="inline-flex items-center gap-1 rounded bg-emerald-500/10 px-2 py-1 text-[10px] font-medium text-emerald-400 hover:bg-emerald-500/15"
                      :disabled="validatingId === item.id"
                      @click="handleValidate(item.id)"
                    >
                      Approve
                    </button>
                    <button
                      class="inline-flex items-center gap-1 rounded bg-red-500/10 px-2 py-1 text-[10px] font-medium text-red-400 hover:bg-red-500/15"
                      :disabled="validatingId === item.id"
                      @click="openRejectionDialog(item.id)"
                    >
                      Reject
                    </button>
                  </div>
                </div>
                <div v-if="!allValidationQueue.length" class="text-center py-8 text-xs text-stone-600">
                  Validation queue is clear. Good job!
                </div>
              </div>
            </div>
          </WorkspaceWindow>

        </template>
      </div>
    </div>

    <!-- Rejection Dialog -->
    <RejectionDialog
      :open="rejectionDialogOpen"
      title="Reject Revision"
      message="Are you sure you want to reject this revision? Please provide an explanation for the student."
      confirm-label="Reject Revision"
      cancel-label="Cancel"
      @confirm="handleRejectConfirm"
      @cancel="rejectionDialogOpen = false"
    />
  </RequireAuth>
</template>

<style scoped>
.dot-grid {
  background-image: radial-gradient(rgba(255, 255, 255, 0.05) 1px, transparent 1px);
  background-size: 20px 20px;
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
