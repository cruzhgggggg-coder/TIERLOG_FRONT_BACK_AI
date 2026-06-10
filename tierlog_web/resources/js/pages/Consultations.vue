<script setup lang="ts">
import { ref, reactive, computed, watch, onMounted, onUnmounted, nextTick } from 'vue';
import { useAuthStore } from '@/stores/auth';
import type { ConsultationLog, FeedbackItem } from '@/types';
import RequireAuth from '@/components/RequireAuth.vue';
import NavBar from '@/components/NavBar.vue';
import UiPage from '@/components/UiPage.vue';
import UiHeading from '@/components/UiHeading.vue';
import UiCard from '@/components/UiCard.vue';
import UiBadge from '@/components/UiBadge.vue';
import UiButton from '@/components/UiButton.vue';
import {
  CheckCircleIcon,
  ClockIcon,
  AlertIcon,
  ChevronRightIcon,
  ConsultationIcon,
} from '@/components/icons';

interface ChatMessage {
  id: string;
  role: 'user' | 'assistant' | 'advisor';
  content: string;
  timestamp: string;
}

interface DraftItem {
  id: number;
  consultation_log_id: number;
  version: number;
  filename: string;
  created_at: string;
  notes: string;
}

interface StudentSubmission {
  id: number;
  student_name: string;
  nim: string;
  paper_filename: string;
  created_at: string;
  status: string;
  feedback_count: number;
  pending_count: number;
}

interface ToastItem {
  id: string;
  message: string;
  type: 'success' | 'error' | 'info';
  visible: boolean;
}

const auth = useAuthStore();
const isLecturer = computed(() => auth.user?.role === 'lecturer');
const isStudent = computed(() => auth.user?.role === 'student');

const loading = ref(true);
const consultations = ref<ConsultationLog[]>([]);
const selectedConsultation = ref<ConsultationLog | null>(null);
const studentSubmissions = ref<StudentSubmission[]>([]);

const activeCenterTab = ref<'feedback' | 'transcript' | 'annotations' | 'drafts'>('feedback');
const chatMode = ref<'oracle' | 'advisor'>('oracle');
const chatMessages = ref<ChatMessage[]>([]);
const chatInput = ref('');
const chatLoading = ref(false);
const drafts = ref<DraftItem[]>([]);
const toasts = ref<ToastItem[]>([]);

const uploadForm = reactive({
  paper: null as File | null,
  audio: null as File | null,
  annotations: null as File | null,
});
const uploading = ref(false);

const feedbackItems = ref<FeedbackItem[]>([]);
const evaluatingFeedbackId = ref<number | null>(null);
const evalContent = ref('');
const evalCategory = ref<'Major' | 'Minor'>('Major');

const archiveDropdownOpen = ref(false);
const messagesEndRef = ref<HTMLElement | null>(null);
const feedbackCategoryFilter = ref<'all' | 'Major' | 'Minor'>('all');

const selectedConsultationLabel = computed(() => {
  if (!selectedConsultation.value) return 'Select a session';
  return `#${selectedConsultation.value.id} — ${selectedConsultation.value.paper_filename}`;
});

const filteredFeedbackItems = computed(() => {
  if (feedbackCategoryFilter.value === 'all') return feedbackItems.value;
  return feedbackItems.value.filter((f) => f.category === feedbackCategoryFilter.value);
});

const feedbackCounts = computed(() => {
  const total = feedbackItems.value.length;
  const major = feedbackItems.value.filter((f) => f.category === 'Major').length;
  const minor = total - major;
  return { total, major, minor };
});

function categoryColor(cat: string): string {
  return cat === 'Major'
    ? 'bg-red-500/15 text-red-400 ring-red-500/30'
    : 'bg-amber-500/15 text-amber-400 ring-amber-500/30';
}

function statusColor(status: string): string {
  const lower = status.toLowerCase();
  if (lower === 'fixed' || lower === 'validated')
    return 'bg-emerald-500/15 text-emerald-400 ring-emerald-500/30';
  if (lower === 'pending')
    return 'bg-amber-500/15 text-amber-400 ring-amber-500/30';
  return 'bg-indigo-500/15 text-indigo-400 ring-indigo-500/30';
}

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString('id-ID', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
}

function formatTime(dateStr: string): string {
  return new Date(dateStr).toLocaleTimeString('id-ID', {
    hour: '2-digit',
    minute: '2-digit',
  });
}

let toastIdCounter = 0;
function showToast(message: string, type: ToastItem['type'] = 'info') {
  const id = String(++toastIdCounter);
  const toast: ToastItem = { id, message, type, visible: false };
  toasts.value.push(toast);
  requestAnimationFrame(() => {
    const t = toasts.value.find((t) => t.id === id);
    if (t) t.visible = true;
  });
  setTimeout(() => {
    const t = toasts.value.find((t) => t.id === id);
    if (t) t.visible = false;
    setTimeout(() => {
      toasts.value = toasts.value.filter((t) => t.id !== id);
    }, 350);
  }, 4000);
}

function scrollToChatBottom() {
  nextTick(() => {
    messagesEndRef.value?.scrollIntoView({ behavior: 'smooth' });
  });
}

watch(chatMessages, scrollToChatBottom, { deep: true });

async function fetchConsultations() {
  loading.value = true;
  const res = await auth.api<{ data: ConsultationLog[] }>('/consultations');
  if (res.ok && res.data) {
    consultations.value = res.data.data ?? [];
    if (consultations.value.length > 0 && !selectedConsultation.value) {
      selectConsultation(consultations.value[0]);
    }
  }
  loading.value = false;
}

async function fetchStudentSubmissions() {
  const res = await auth.api<{ data: StudentSubmission[] }>('/lecturer/submissions');
  if (res.ok && res.data) {
    studentSubmissions.value = res.data.data ?? [];
  }
}

function selectConsultation(log: ConsultationLog) {
  selectedConsultation.value = log;
  feedbackItems.value = log.feedback_items ? [...log.feedback_items] : [];
  archiveDropdownOpen.value = false;
  activeCenterTab.value = 'feedback';
  feedbackCategoryFilter.value = 'all';
  fetchDrafts(log.id);
}

async function fetchDrafts(logId: number) {
  const res = await auth.api<{ data: DraftItem[] }>(`/consultations/${logId}/drafts`);
  if (res.ok && res.data) {
    drafts.value = res.data.data ?? [];
  }
}

async function loadDirectMessages(logId: number) {
  const res = await auth.api<{ data: any[] }>(`/consultations/${logId}/direct-messages`);
  if (res.ok && res.data) {
    chatMessages.value = (res.data.data ?? []).map((msg: any) => ({
      id: String(msg.id),
      role: msg.sender_role === 'student' ? 'user' : 'advisor',
      content: msg.content,
      timestamp: msg.created_at,
    }));
  } else {
    chatMessages.value = [];
  }
}

async function loadAIChats(logId: number) {
  const res = await auth.api<{ data: any[] }>(`/consultations/${logId}/ai-chats`);
  if (res.ok && res.data) {
    chatMessages.value = (res.data.data ?? []).map((msg: any) => ({
      id: String(msg.id),
      role: msg.role === 'user' ? 'user' : 'assistant',
      content: msg.content,
      timestamp: msg.created_at,
    }));
  } else {
    chatMessages.value = [];
  }
}

watch(
  [chatMode, () => selectedConsultation.value?.id],
  async ([mode, logId]) => {
    if (!isStudent.value) return;
    if (!logId) {
      chatMessages.value = [];
      return;
    }
    chatLoading.value = true;
    try {
      if (mode === 'oracle') {
        await loadAIChats(Number(logId));
      } else {
        await loadDirectMessages(Number(logId));
      }
    } catch {
      showToast('Failed to load chat history', 'error');
    } finally {
      chatLoading.value = false;
    }
  },
  { immediate: true }
);

async function handleUpload() {
  if (!uploadForm.paper) {
    showToast('Paper file is required', 'error');
    return;
  }
  if (!uploadForm.audio && !uploadForm.annotations) {
    showToast('Please provide audio or annotation images', 'error');
    return;
  }
  uploading.value = true;
  const formData = new FormData();
  formData.append('paper', uploadForm.paper);
  if (uploadForm.audio) formData.append('audio', uploadForm.audio);
  if (uploadForm.annotations) formData.append('annotations', uploadForm.annotations);

  try {
    const apiUrl = import.meta.env.VITE_API_URL || 'http://127.0.0.1:8080';
    const res = await fetch(`${apiUrl}/consultations`, {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${auth.accessToken}`,
      },
      body: formData,
    });
    if (res.ok) {
      showToast('Session uploaded successfully', 'success');
      uploadForm.paper = null;
      uploadForm.audio = null;
      uploadForm.annotations = null;
      await fetchConsultations();
    } else {
      const data = await res.json().catch(() => ({}));
      showToast(data.error || 'Upload failed', 'error');
    }
  } catch {
    showToast('Upload failed — network error', 'error');
  }
  uploading.value = false;
}

// ── Image Compression ──

function compressImage(
  file: File,
  maxWidth = 1600,
  maxHeight = 1600,
  quality = 0.75,
): Promise<File> {
  return new Promise((resolve) => {
    if (!file.type.startsWith('image/')) {
      resolve(file);
      return;
    }
    const reader = new FileReader();
    reader.onload = (event) => {
      const img = new Image();
      img.onload = () => {
        let { width, height } = img;
        if (width > height) {
          if (width > maxWidth) {
            height = Math.round((height * maxWidth) / width);
            width = maxWidth;
          }
        } else {
          if (height > maxHeight) {
            width = Math.round((width * maxHeight) / height);
            height = maxHeight;
          }
        }
        const canvas = document.createElement('canvas');
        canvas.width = width;
        canvas.height = height;
        const ctx = canvas.getContext('2d');
        ctx?.drawImage(img, 0, 0, width, height);
        canvas.toBlob((blob) => {
          if (!blob) {
            resolve(file);
            return;
          }
          const nameWithoutExt = file.name.substring(0, file.name.lastIndexOf('.'));
          resolve(new File([blob], `${nameWithoutExt}_compressed.jpg`, { type: 'image/jpeg' }));
        }, 'image/jpeg', quality);
      };
      img.src = event.target?.result as string;
    };
    reader.readAsDataURL(file);
  });
}

function formatFileSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(0)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

async function onFileSelect(field: 'paper' | 'audio' | 'annotations', event: Event) {
  const input = event.target as HTMLInputElement;
  if (input.files && input.files.length > 0) {
    let file = input.files[0];
    // Compress annotation images before storing
    if (field === 'annotations' && file.type.startsWith('image/')) {
      const originalSize = file.size;
      file = await compressImage(file);
      const savedPercent = Math.round(((originalSize - file.size) / originalSize) * 100);
      showToast(
        `Image compressed: ${formatFileSize(originalSize)} → ${formatFileSize(file.size)} (${savedPercent}% saved)`,
        'info',
      );
    }
    uploadForm[field] = file;
  }
}

// ── Session Deletion ──

async function deleteSession(logId: number) {
  if (!confirm('Are you sure you want to delete this session? This action cannot be undone.')) {
    return;
  }
  const res = await auth.api(`/consultations/${logId}`, { method: 'DELETE' });
  if (res.ok) {
    showToast('Session deleted', 'success');
    if (selectedConsultation.value?.id === logId) {
      selectedConsultation.value = null;
    }
    await fetchConsultations();
  } else {
    showToast(res.error || 'Failed to delete session', 'error');
  }
}

async function toggleFeedbackStatus(item: FeedbackItem) {
  const newStatus = item.status === 'Fixed' ? 'Pending' : 'Fixed';
  const res = await auth.api(`/consultations/feedback/${item.id}/status`, {
    method: 'PUT',
    body: JSON.stringify({ status: newStatus, log_id: item.consultation_log_id }),
  });
  if (res.ok) {
    item.status = newStatus;
    showToast(`Marked as ${newStatus}`, 'success');
  } else {
    showToast('Failed to update status', 'error');
  }
}

async function quickAIRevision(item: FeedbackItem) {
  if (!selectedConsultation.value) return;
  chatMode.value = 'oracle';
  const prompt = `Provide concrete solutions and textual revision improvements for the following feedback item:\n"${item.content}"`;
  chatMessages.value.push({
    id: `u-${Date.now()}`,
    role: 'user',
    content: prompt,
    timestamp: new Date().toISOString(),
  });
  chatLoading.value = true;
  await nextTick();
  scrollToChatBottom();
  const res = await auth.api<{ ai_response: string }>(
    '/consultations/chat',
    {
      method: 'POST',
      body: JSON.stringify({ log_id: selectedConsultation.value.id, query: prompt }),
    },
  );
  chatLoading.value = false;
  if (res.ok && res.data) {
    chatMessages.value.push({
      id: `a-${Date.now()}`,
      role: 'assistant',
      content: res.data.ai_response,
      timestamp: new Date().toISOString(),
    });
  } else {
    chatMessages.value.push({
      id: `a-${Date.now()}`,
      role: 'assistant',
      content: `AI Oracle error: ${res.error || 'Failed to respond'}. Check your API key in AI Gateway settings.`,
      timestamp: new Date().toISOString(),
    });
  }
  await nextTick();
  scrollToChatBottom();
}

async function classifyFeedback() {
  if (!selectedConsultation.value) return;
  showToast('Running AI classification...', 'info');
  const res = await auth.api(
    `/consultations/${selectedConsultation.value.id}/classify-feedback`,
    { method: 'POST' },
  );
  if (res.ok) {
    showToast('Feedback classified successfully', 'success');
    await fetchConsultations();
  } else {
    showToast('Classification failed', 'error');
  }
}

async function sendChat() {
  if (!chatInput.value.trim() || chatLoading.value) return;
  const content = chatInput.value.trim();

  if (chatMode.value === 'advisor' && !selectedConsultation.value) {
    showToast('Select a consultation first', 'error');
    return;
  }

  chatInput.value = '';

  const tempId = chatMode.value === 'oracle' ? `u-${Date.now()}` : `dm-${Date.now()}`;
  const userMsg: ChatMessage = {
    id: tempId,
    role: 'user',
    content,
    timestamp: new Date().toISOString(),
  };
  chatMessages.value.push(userMsg);
  chatLoading.value = true;

  if (chatMode.value === 'oracle') {
    const logId = selectedConsultation.value?.id;
    const res = await auth.api<{ ai_response: string }>('/consultations/chat', {
      method: 'POST',
      body: JSON.stringify({ query: content, log_id: logId }),
    });
    chatLoading.value = false;
    if (res.ok && res.data) {
      chatMessages.value.push({
        id: `a-${Date.now()}`,
        role: 'assistant',
        content: res.data.ai_response,
        timestamp: new Date().toISOString(),
      });
    } else {
      showToast('AI Oracle failed to respond', 'error');
    }
  } else {
    const res = await auth.api<{ data: unknown }>(
      `/consultations/${selectedConsultation.value!.id}/direct-messages`,
      { method: 'POST', body: JSON.stringify({ content }) },
    );
    chatLoading.value = false;
    if (res.ok) {
      showToast('Message sent', 'success');
    } else {
      chatMessages.value = chatMessages.value.filter((m) => m.id !== tempId);
      showToast('Failed to send message', 'error');
    }
  }
}


async function generateTranscript() {
  if (!selectedConsultation.value) return;
  showToast('Generating transcript...', 'info');
  const res = await auth.api(
    `/consultations/${selectedConsultation.value.id}/transcribe`,
    { method: 'POST' },
  );
  if (res.ok) {
    showToast('Transcript generated', 'success');
    await fetchConsultations();
  } else {
    showToast('Transcription failed', 'error');
  }
}

async function analyzeFeedback() {
  if (!selectedConsultation.value) return;
  showToast('Analyzing feedback...', 'info');
  const res = await auth.api(
    `/consultations/${selectedConsultation.value.id}/analyze`,
    { method: 'POST' },
  );
  if (res.ok) {
    showToast('Feedback analysis complete', 'success');
    await fetchConsultations();
  } else {
    showToast('Analysis failed', 'error');
  }
}

async function validateFeedback(item: FeedbackItem) {
  const res = await auth.api(`/consultations/feedback/${item.id}/status`, {
    method: 'PUT',
    body: JSON.stringify({ status: 'Validated', log_id: item.consultation_log_id }),
  });
  if (res.ok) {
    item.status = 'Validated';
    showToast('Feedback validated', 'success');
  } else {
    showToast('Validation failed', 'error');
  }
}

async function rejectFeedback(item: FeedbackItem) {
  const res = await auth.api(`/consultations/feedback/${item.id}/status`, {
    method: 'PUT',
    body: JSON.stringify({ status: 'Pending', log_id: item.consultation_log_id }),
  });
  if (res.ok) {
    item.status = 'Pending';
    showToast('Feedback returned to pending', 'success');
  } else {
    showToast('Rejection failed', 'error');
  }
}

async function undoFeedbackAction(item: FeedbackItem) {
  const res = await auth.api(`/consultations/feedback/${item.id}/status`, {
    method: 'PUT',
    body: JSON.stringify({ status: 'Pending', log_id: item.consultation_log_id }),
  });
  if (res.ok) {
    item.status = 'Pending';
    showToast('Action undone', 'success');
  } else {
    showToast('Undo failed', 'error');
  }
}

function startEvaluate(item: FeedbackItem) {
  evaluatingFeedbackId.value = item.id;
  evalContent.value = item.content;
  evalCategory.value = item.category;
}

async function submitEvaluation() {
  if (!evaluatingFeedbackId.value || !selectedConsultation.value) return;
  const res = await auth.api(
    `/consultations/feedback/${evaluatingFeedbackId.value}/status`,
    {
      method: 'PUT',
      body: JSON.stringify({ status: evalCategory.value === 'Major' ? 'Pending' : 'Fixed', log_id: selectedConsultation.value.id }),
    },
  );
  if (res.ok) {
    const item = feedbackItems.value.find((f) => f.id === evaluatingFeedbackId.value);
    if (item) {
      item.content = evalContent.value;
      item.category = evalCategory.value;
    }
    evaluatingFeedbackId.value = null;
    evalContent.value = '';
    showToast('Feedback updated', 'success');
  } else {
    showToast('Update failed', 'error');
  }
}

function submissionStatusColor(status: string): string {
  const lower = status.toLowerCase();
  if (lower === 'approved') return 'bg-emerald-500/15 text-emerald-400 ring-emerald-500/30';
  if (lower === 'pending') return 'bg-amber-500/15 text-amber-400 ring-amber-500/30';
  return 'bg-indigo-500/15 text-indigo-400 ring-indigo-500/30';
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
      if (isLecturer.value) {
        studentSubmissions.value.forEach((sub) => {
          ws?.send(JSON.stringify({ action: 'subscribe', room: `consultation.${sub.id}` }));
        });
      }
    };
    ws.onmessage = (event) => {
      try {
        const payload = JSON.parse(event.data);
        if (payload.event === 'feedback.new' || payload.event === 'feedback.status-updated') {
          fetchConsultations();
          if (payload.event === 'feedback.new') {
            showToast('New feedback received', 'info');
          }
        }
        if (payload.event === 'chat.direct-message') {
          const msgData = payload.data;
          if (!msgData) return;
          const isFromAdvisor = msgData.sender_role === 'lecturer';
          
          if (selectedConsultation.value?.id === msgData.log_id && chatMode.value === 'advisor') {
            const tempIndex = chatMessages.value.findIndex(
              (m) => m.id.startsWith('dm-') && m.content === msgData.content
            );
            if (tempIndex !== -1) {
              chatMessages.value[tempIndex].id = String(msgData.id);
              chatMessages.value[tempIndex].timestamp = msgData.created_at;
            } else {
              const exists = chatMessages.value.some((m) => m.id === String(msgData.id));
              if (!exists) {
                chatMessages.value.push({
                  id: String(msgData.id),
                  role: msgData.sender_role === 'student' ? 'user' : 'advisor',
                  content: msgData.content,
                  timestamp: msgData.created_at,
                });
              }
            }
          }
          if (isFromAdvisor && (chatMode.value !== 'advisor' || selectedConsultation.value?.id !== msgData.log_id)) {
            showToast('New message from advisor', 'info');
          }
        }
        if (payload.event === 'chat.message') {
          const msgData = payload.data;
          if (!msgData) return;
          
          if (selectedConsultation.value?.id === msgData.log_id && chatMode.value === 'oracle') {
            const role = msgData.role === 'user' ? 'user' : 'assistant';
            const tempPrefix = role === 'user' ? 'u-' : 'a-';
            
            const tempIndex = chatMessages.value.findIndex(
              (m) => m.id.startsWith(tempPrefix) && m.content === msgData.content
            );
            if (tempIndex !== -1) {
              chatMessages.value[tempIndex].id = String(msgData.id);
              chatMessages.value[tempIndex].timestamp = msgData.created_at;
            } else {
              const exists = chatMessages.value.some((m) => m.id === String(msgData.id));
              if (!exists) {
                chatMessages.value.push({
                  id: String(msgData.id),
                  role: role,
                  content: msgData.content,
                  timestamp: msgData.created_at,
                });
              }
            }
          }
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

watch(
  () => consultations.value,
  (newLogs) => {
    if (ws && ws.readyState === WebSocket.OPEN) {
      newLogs.forEach((log) => {
        ws?.send(JSON.stringify({ action: 'subscribe', room: `consultation.${log.id}` }));
      });
    }
  },
  { deep: true }
);

watch(
  () => studentSubmissions.value,
  (newSubs) => {
    if (isLecturer.value && ws && ws.readyState === WebSocket.OPEN) {
      newSubs.forEach((sub) => {
        ws?.send(JSON.stringify({ action: 'subscribe', room: `consultation.${sub.id}` }));
      });
    }
  },
  { deep: true }
);

function closeArchiveDropdown(e: Event) {
  const target = e.target as HTMLElement;
  if (!target.closest('[data-archive-dropdown]')) {
    archiveDropdownOpen.value = false;
  }
}

onMounted(() => {
  fetchConsultations();
  if (isLecturer.value) {
    fetchStudentSubmissions();
  }
  connectWebSocket();
  document.addEventListener('click', closeArchiveDropdown);
});

onUnmounted(() => {
  ws?.close();
  ws = null;
  document.removeEventListener('click', closeArchiveDropdown);
});
</script>

<template>
  <RequireAuth>
    <NavBar />
    <UiPage>
      <div class="mesh-gradient space-y-6">
        <UiHeading
          :title="isLecturer ? 'Lecturer Consultation Workspace' : 'Student Consultation Workspace'"
          subtitle="Upload, review, and collaborate on consultation sessions"
        />

        <div v-if="loading" class="flex items-center justify-center py-20">
          <div class="h-7 w-7 animate-spin rounded-full border-[1.5px] border-amber-500 border-t-transparent" />
        </div>

        <template v-else>
          <div v-if="isStudent" class="flex flex-col gap-4 lg:flex-row lg:gap-5 lg:h-[calc(100vh-220px)]">
            <div class="flex w-full flex-col gap-4 lg:w-80 lg:shrink-0">
              <UiCard>
                <div class="p-5">
                  <div class="mb-4 flex items-center gap-2.5">
                    <div class="flex h-8 w-8 items-center justify-center rounded-lg bg-amber-500/10">
                      <svg class="h-4 w-4 text-amber-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
                        <path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4" />
                        <polyline points="17 8 12 3 7 8" />
                        <line x1="12" y1="3" x2="12" y2="15" />
                      </svg>
                    </div>
                    <div>
                      <h3 class="text-sm font-semibold text-stone-100 font-[family-name:var(--font-display)]">Upload Session</h3>
                      <p class="text-[11px] text-stone-600">Submit your consultation files</p>
                    </div>
                  </div>
                  <form class="space-y-3" @submit.prevent="handleUpload">
                    <div>
                      <label class="mb-1.5 block text-xs font-medium text-stone-500">
                        Paper <span class="text-red-400">*</span>
                      </label>
                      <div class="relative">
                        <input
                          type="file"
                          accept=".docx"
                          class="peer w-full cursor-pointer rounded-lg border border-white/[0.04] bg-[#0a0a0b] px-3.5 py-2.5 text-xs text-stone-300 transition-colors file:mr-3 file:cursor-pointer file:rounded-md file:border-0 file:bg-amber-500/10 file:px-3 file:py-1.5 file:text-xs file:font-medium file:text-amber-400 hover:file:bg-amber-500/15 focus:border-amber-500/40 focus:outline-none focus:ring-1 focus:ring-amber-500/20"
                          @change="onFileSelect('paper', $event)"
                        />
                      </div>
                    </div>
                    <div>
                      <label class="mb-1.5 block text-xs font-medium text-stone-500">Audio</label>
                      <input
                        type="file"
                        accept="audio/*"
                        class="w-full cursor-pointer rounded-lg border border-white/[0.04] bg-[#0a0a0b] px-3.5 py-2.5 text-xs text-stone-300 transition-colors file:mr-3 file:cursor-pointer file:rounded-md file:border-0 file:bg-zinc-500/10 file:px-3 file:py-1.5 file:text-xs file:font-medium file:text-stone-500 hover:file:bg-zinc-500/15 focus:border-white/20 focus:outline-none focus:ring-1 focus:ring-white/10"
                        @change="onFileSelect('audio', $event)"
                      />
                    </div>
                    <div>
                      <label class="mb-1.5 block text-xs font-medium text-stone-500">Annotations</label>
                      <input
                        type="file"
                        accept="image/*,.docx"
                        class="w-full cursor-pointer rounded-lg border border-white/[0.04] bg-[#0a0a0b] px-3.5 py-2.5 text-xs text-stone-300 transition-colors file:mr-3 file:cursor-pointer file:rounded-md file:border-0 file:bg-zinc-500/10 file:px-3 file:py-1.5 file:text-xs file:font-medium file:text-stone-500 hover:file:bg-zinc-500/15 focus:border-white/20 focus:outline-none focus:ring-1 focus:ring-white/10"
                        @change="onFileSelect('annotations', $event)"
                      />
                    </div>
                    <UiButton
                      title="Upload Session"
                      :disabled="!uploadForm.paper || uploading"
                      @click="handleUpload"
                    />
                  </form>
                </div>
              </UiCard>

              <div class="relative" data-archive-dropdown>
                <button
                  class="flex w-full items-center justify-between rounded-xl border border-white/[0.04] bg-[#141415] px-4 py-3 text-sm text-stone-300 transition-colors hover:border-white/[0.12]"
                  @click="archiveDropdownOpen = !archiveDropdownOpen"
                >
                  <span class="truncate">{{ selectedConsultationLabel }}</span>
                  <ChevronRightIcon
                    :size="14"
                    :class="['text-stone-600 transition-transform duration-200', archiveDropdownOpen ? 'rotate-90' : '']"
                  />
                </button>
                <Transition
                  enter-active-class="transition duration-200 ease-out"
                  enter-from-class="opacity-0 -translate-y-1"
                  enter-to-class="opacity-100 translate-y-0"
                  leave-active-class="transition duration-150 ease-in"
                  leave-from-class="opacity-100 translate-y-0"
                  leave-to-class="opacity-0 -translate-y-1"
                >
                  <div
                    v-if="archiveDropdownOpen"
                    class="absolute z-30 mt-1 max-h-64 w-full scroll-area rounded-xl border border-white/[0.04] bg-[#141415] shadow-2xl shadow-black/60"
                  >
                    <div
                      v-for="log in consultations"
                      :key="log.id"
                      :class="[
                        'cursor-pointer border-b border-white/[0.04] px-4 py-3 transition-colors last:border-b-0',
                        selectedConsultation?.id === log.id
                          ? 'bg-amber-500/5'
                          : 'hover:bg-white/[0.03]',
                      ]"
                      @click="selectConsultation(log)"
                    >
                      <div class="flex items-center justify-between">
                        <p class="truncate text-sm font-medium text-stone-200">{{ log.paper_filename }}</p>
                        <span class="ml-2 shrink-0 text-[10px] text-stone-700">#{{ log.id }}</span>
                      </div>
                      <p class="mt-0.5 text-[11px] text-stone-600">{{ formatDate(log.created_at) }}</p>
                    </div>
                    <div v-if="consultations.length === 0" class="px-4 py-6 text-center text-xs text-stone-600">
                      No sessions found
                    </div>
                  </div>
                </Transition>
              </div>

              <UiCard v-if="selectedConsultation" class="flex-1">
                <div class="p-5">
                  <h3 class="mb-3 text-[11px] font-semibold uppercase tracking-wider text-stone-600 font-[family-name:var(--font-display)]">Session Info</h3>
                  <div class="space-y-2.5">
                    <div class="flex items-center justify-between text-sm">
                      <span class="text-stone-600">Session</span>
                      <span class="font-medium text-stone-300">#{{ selectedConsultation.id }}</span>
                    </div>
                    <div class="border-t border-white/[0.04]" />
                    <div class="flex items-center justify-between text-sm">
                      <span class="text-stone-600">Paper</span>
                      <span class="max-w-[140px] truncate text-stone-300">{{ selectedConsultation.paper_filename }}</span>
                    </div>
                    <div class="border-t border-white/[0.04]" />
                    <div class="flex items-center justify-between text-sm">
                      <span class="text-stone-600">Feedback</span>
                      <span class="font-medium text-stone-300">{{ selectedConsultation.feedback_items?.length ?? 0 }}</span>
                    </div>
                    <div class="border-t border-white/[0.04]" />
                    <div class="flex items-center justify-between text-sm">
                      <span class="text-stone-600">Date</span>
                      <span class="text-stone-300">{{ formatDate(selectedConsultation.created_at) }}</span>
                    </div>
                  </div>
                  <div class="mt-4 border-t border-white/[0.04] pt-4">
                    <button
                      class="inline-flex w-full items-center justify-center gap-1.5 rounded-lg bg-red-500/10 px-3 py-2 text-xs font-medium text-red-400 transition-colors hover:bg-red-500/15"
                      @click="deleteSession(selectedConsultation.id)"
                    >
                      <svg class="h-3.5 w-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                        <polyline points="3 6 5 6 21 6" />
                        <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2" />
                      </svg>
                      Delete Session
                    </button>
                  </div>
                </div>
              </UiCard>
            </div>

            <div class="flex flex-[1.4] flex-col">
              <UiCard class="flex flex-1 flex-col overflow-hidden" :padding="false">
                <div class="border-b border-white/[0.04]">
                  <div class="flex gap-0.5 px-4 pt-3">
                    <button
                      v-for="tab in (['feedback', 'transcript', 'annotations', 'drafts'] as const)"
                      :key="tab"
                      :class="[
                        'rounded-lg px-4 py-2 text-[13px] font-medium transition-colors duration-100',
                        activeCenterTab === tab
                          ? 'bg-[#1c1c1e] text-stone-100'
                          : 'text-stone-600 hover:text-stone-300',
                      ]"
                      @click="activeCenterTab = tab"
                    >
                      {{ tab.charAt(0).toUpperCase() + tab.slice(1) }}
                    </button>
                  </div>
                </div>

                <div class="flex-1 scroll-area p-5">
                  <div v-if="!selectedConsultation" class="flex h-full flex-col items-center justify-center">
                    <div class="flex h-12 w-12 items-center justify-center rounded-xl bg-[#141415]">
                      <ConsultationIcon :size="22" color="#52525b" />
                    </div>
                    <p class="mt-3 text-sm text-stone-600">Select a session from the archive</p>
                  </div>

                  <div v-if="selectedConsultation && activeCenterTab === 'feedback'" class="space-y-3">
                    <div class="flex items-center justify-between">
                      <h4 class="text-sm font-semibold text-stone-200 font-[family-name:var(--font-display)]">Feedback Items</h4>
                      <button
                        class="inline-flex items-center gap-1.5 rounded-md bg-[#1c1c1e] px-3 py-1.5 text-xs font-medium text-stone-300 transition-colors hover:bg-[#2c2c2e]"
                        @click="classifyFeedback"
                      >
                        <svg class="h-3.5 w-3.5 text-amber-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                          <polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2" />
                        </svg>
                        AI Classify
                      </button>
                    </div>

                    <div class="flex gap-1 rounded-lg bg-[#0a0a0b] p-1">
                      <button
                        v-for="filter in (['all', 'Major', 'Minor'] as const)"
                        :key="filter"
                        :class="[
                          'rounded-md px-3 py-1.5 text-[12px] font-medium transition-colors duration-100',
                          feedbackCategoryFilter === filter
                            ? 'bg-[#1c1c1e] text-stone-100'
                            : 'text-stone-600 hover:text-stone-300',
                        ]"
                        @click="feedbackCategoryFilter = filter"
                      >
                        {{ filter === 'all' ? 'All' : filter }}
                        <span class="ml-0.5 text-[10px] text-stone-700">
                          {{ filter === 'all' ? feedbackCounts.total : filter === 'Major' ? feedbackCounts.major : feedbackCounts.minor }}
                        </span>
                      </button>
                    </div>

                    <div
                      v-for="item in filteredFeedbackItems"
                      :key="item.id"
                      class="rounded-xl border border-white/[0.04] bg-[#141415] p-4 transition-colors hover:border-white/[0.08]"
                    >
                      <div class="flex items-start gap-3">
                        <label class="mt-0.5 flex shrink-0 cursor-pointer items-center">
                          <input
                            type="checkbox"
                            :checked="item.status === 'Fixed'"
                            class="h-4 w-4 rounded border-zinc-600 bg-[#1c1c1e] text-amber-500 focus:ring-amber-500/40"
                            @change="toggleFeedbackStatus(item)"
                          />
                        </label>
                        <div class="min-w-0 flex-1">
                          <p class="text-sm leading-relaxed text-stone-300">{{ item.content }}</p>
                          <div class="mt-2.5 flex flex-wrap items-center gap-1.5">
                            <UiBadge :text="item.category" :color="categoryColor(item.category)" />
                            <UiBadge :text="item.status" :color="statusColor(item.status)" />
                          </div>
                        </div>
                      </div>
                      <div class="mt-3 flex gap-2 border-t border-white/[0.04] pt-3">
                        <button
                          class="inline-flex items-center gap-1.5 rounded-md bg-[#1c1c1e] px-3 py-1.5 text-xs font-medium text-stone-300 transition-colors hover:bg-[#2c2c2e]"
                          @click="quickAIRevision(item)"
                        >
                          <svg class="h-3.5 w-3.5 text-stone-500" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                            <polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2" />
                          </svg>
                          AI Revision
                        </button>
                      </div>
                    </div>

                    <div v-if="filteredFeedbackItems.length === 0" class="py-12 text-center">
                      <p class="text-sm text-stone-600">No feedback items for this filter.</p>
                    </div>
                  </div>

                  <div v-if="activeCenterTab === 'transcript' && selectedConsultation">
                    <div class="rounded-xl border border-white/[0.04] bg-[#0a0a0b] p-5">
                      <div v-if="selectedConsultation.transcript_text" class="max-h-[500px] scroll-area">
                        <p class="whitespace-pre-wrap text-sm leading-[1.8] text-stone-500">
                          {{ selectedConsultation.transcript_text }}
                        </p>
                      </div>
                      <div v-else class="flex flex-col items-center justify-center py-12">
                        <div class="flex h-10 w-10 items-center justify-center rounded-xl bg-[#141415]">
                          <ClockIcon :size="18" color="#52525b" />
                        </div>
                        <p class="mt-3 text-sm text-stone-600">No transcript available.</p>
                      </div>
                    </div>
                  </div>

                  <div v-if="activeCenterTab === 'annotations' && selectedConsultation" class="space-y-3">
                    <div
                      v-for="annotation in selectedConsultation.revision_annotations"
                      :key="annotation.id"
                      class="rounded-xl border border-white/[0.04] bg-[#141415] p-4"
                    >
                      <div class="mb-3 flex items-center justify-between">
                        <div class="flex items-center gap-3">
                          <div class="flex h-8 w-8 items-center justify-center rounded-lg bg-[#1c1c1e]">
                            <span class="text-[10px] font-bold uppercase text-stone-500">
                              {{ annotation.file_type === 'image' ? 'IMG' : 'DOC' }}
                            </span>
                          </div>
                          <div>
                            <p class="text-sm font-medium text-stone-200">{{ annotation.filename }}</p>
                            <p class="text-[11px] text-stone-600">{{ annotation.file_type.toUpperCase() }}</p>
                          </div>
                        </div>
                        <p class="text-[11px] text-stone-700">{{ formatDate(annotation.created_at) }}</p>
                      </div>
                      <div
                        v-if="annotation.file_type === 'image'"
                        class="mb-3 overflow-hidden rounded-lg border border-white/[0.04]"
                      >
                        <div class="flex h-28 items-center justify-center bg-[#0a0a0b]">
                          <p class="text-xs text-stone-700">Image Preview</p>
                        </div>
                      </div>
                      <div v-if="annotation.extracted_text" class="rounded-lg bg-[#0a0a0b] p-3.5">
                        <p class="mb-1.5 text-[10px] font-semibold uppercase tracking-wider text-stone-600">OCR Text</p>
                        <p class="whitespace-pre-wrap text-sm leading-relaxed text-stone-500">
                          {{ annotation.extracted_text }}
                        </p>
                      </div>
                    </div>
                    <div v-if="!selectedConsultation.revision_annotations?.length" class="py-12 text-center">
                      <p class="text-sm text-stone-600">No annotations for this session.</p>
                    </div>
                  </div>

                  <div v-if="activeCenterTab === 'drafts'" class="space-y-2">
                    <div
                      v-for="draft in drafts"
                      :key="draft.id"
                      class="flex items-center gap-3 rounded-xl border border-white/[0.04] bg-[#141415] p-4 transition-colors hover:border-white/[0.08]"
                    >
                      <div class="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-[#1c1c1e]">
                        <span class="text-xs font-bold text-stone-500">v{{ draft.version }}</span>
                      </div>
                      <div class="min-w-0 flex-1">
                        <p class="text-sm font-medium text-stone-200">{{ draft.filename }}</p>
                        <p class="text-[11px] text-stone-600">{{ formatDate(draft.created_at) }}</p>
                      </div>
                      <ChevronRightIcon :size="14" class="text-stone-700" />
                    </div>
                    <div v-if="drafts.length === 0" class="py-12 text-center">
                      <p class="text-sm text-stone-600">No drafts yet.</p>
                    </div>
                  </div>
                </div>
              </UiCard>
            </div>

            <div class="flex flex-[2.2] flex-col">
              <UiCard class="flex flex-1 flex-col overflow-hidden" :padding="false">
                <div class="border-b border-white/[0.04]">
                  <div class="flex gap-1 px-4 pt-3">
                    <button
                      :class="[
                        'rounded-lg px-5 py-2.5 text-sm font-semibold transition-all duration-200',
                        chatMode === 'oracle'
                          ? 'bg-amber-500/10 text-amber-400'
                          : 'text-stone-500 hover:text-stone-300 hover:bg-white/[0.03]',
                      ]"
                      @click="chatMode = 'oracle'"
                    >
                      AI Oracle
                    </button>
                    <button
                      :class="[
                        'rounded-lg px-5 py-2.5 text-sm font-semibold transition-all duration-200',
                        chatMode === 'advisor'
                          ? 'bg-amber-500/10 text-amber-400'
                          : 'text-stone-500 hover:text-stone-300 hover:bg-white/[0.03]',
                      ]"
                      @click="chatMode = 'advisor'"
                    >
                      Advisor
                    </button>
                  </div>
                </div>

                <div class="flex-1 scroll-area p-4">
                  <div v-if="chatMessages.length === 0" class="flex h-full flex-col items-center justify-center">
                    <div class="flex h-11 w-11 items-center justify-center rounded-xl bg-[#141415]">
                      <svg v-if="chatMode === 'oracle'" class="h-5 w-5 text-stone-600" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
                        <circle cx="12" cy="12" r="10" />
                        <path d="M12 16v-4" />
                        <path d="M12 8h.01" />
                      </svg>
                      <svg v-else class="h-5 w-5 text-stone-600" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
                        <path d="M21 15a2 2 0 01-2 2H7l-4 4V5a2 2 0 012-2h14a2 2 0 012 2z" />
                      </svg>
                    </div>
                    <p class="mt-3 text-sm text-stone-600">
                      {{ chatMode === 'oracle' ? 'Ask the AI Oracle about your session' : 'Start a discussion with your advisor' }}
                    </p>
                    <p class="mt-1 text-[11px] text-stone-700">Messages are sent in real-time</p>
                  </div>

                  <div v-else class="space-y-2.5">
                    <div
                      v-for="msg in chatMessages"
                      :key="msg.id"
                      :class="['flex', msg.role === 'user' ? 'justify-end' : 'justify-start']"
                    >
                      <div
                        :class="[
                          'max-w-[85%] rounded-2xl px-4 py-2.5 text-sm leading-relaxed',
                          msg.role === 'user'
                            ? 'rounded-br-md bg-amber-500/10 text-amber-100'
                            : msg.role === 'assistant'
                              ? 'rounded-bl-md bg-[#1c1c1e] text-stone-300'
                              : 'rounded-bl-md bg-[#1c1c1e] text-stone-300',
                        ]"
                      >
                        <p class="whitespace-pre-wrap">{{ msg.content }}</p>
                        <p class="mt-1 text-[10px] text-stone-700">{{ formatTime(msg.timestamp) }}</p>
                      </div>
                    </div>

                    <div v-if="chatLoading" class="flex justify-start">
                      <div class="rounded-2xl rounded-bl-md bg-[#1c1c1e] px-4 py-3">
                        <div class="flex items-center gap-1">
                          <span class="h-1.5 w-1.5 animate-pulse rounded-full bg-zinc-500" style="animation-delay: 0ms" />
                          <span class="h-1.5 w-1.5 animate-pulse rounded-full bg-zinc-500" style="animation-delay: 200ms" />
                          <span class="h-1.5 w-1.5 animate-pulse rounded-full bg-zinc-500" style="animation-delay: 400ms" />
                        </div>
                      </div>
                    </div>

                    <div ref="messagesEndRef" />
                  </div>
                </div>

                <div class="border-t border-white/[0.04] p-4">
                  <form class="flex gap-3" @submit.prevent="sendChat">
                    <input
                      v-model="chatInput"
                      type="text"
                      :placeholder="chatMode === 'oracle' ? 'Ask the AI Oracle about your thesis...' : 'Message your advisor...'"
                      class="flex-1 rounded-xl border border-white/[0.06] bg-[#0a0a0b] px-4 py-3 text-sm text-stone-200 placeholder-stone-600 transition-colors focus:border-amber-500/30 focus:outline-none focus:ring-1 focus:ring-amber-500/10"
                    />
                    <button
                      type="submit"
                      :disabled="!chatInput.trim() || chatLoading"
                      class="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-amber-500 text-[#0a0a0b] font-semibold transition-all hover:bg-amber-400 disabled:opacity-30"
                    >
                      <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                        <path d="M22 2L11 13" />
                        <path d="M22 2L15 22L11 13L2 9L22 2Z" />
                      </svg>
                    </button>
                  </form>
                </div>
              </UiCard>
            </div>
          </div>

          <div v-if="isLecturer" class="flex flex-col gap-4 lg:flex-row lg:gap-5 lg:h-[calc(100vh-220px)]">
            <div class="w-80 shrink-0">
              <UiCard class="flex h-full flex-col overflow-hidden" :padding="false">
                <div class="border-b border-white/[0.04] p-4">
                  <div class="flex items-center justify-between">
                    <h3 class="text-sm font-semibold text-stone-100 font-[family-name:var(--font-display)]">Submission Queue</h3>
                    <span class="text-[11px] text-stone-600">{{ studentSubmissions.length }} items</span>
                  </div>
                </div>
                <div class="flex-1 scroll-area">
                  <div
                    v-for="sub in studentSubmissions"
                    :key="sub.id"
                    :class="[
                      'cursor-pointer border-b border-white/[0.04] p-4 transition-colors duration-100',
                      selectedConsultation?.id === sub.id
                        ? 'bg-amber-500/5 border-l-2 border-l-amber-500'
                        : 'hover:bg-white/[0.02]',
                    ]"
                    @click="
                      selectedConsultation = {
                        id: sub.id,
                        student_id: 0,
                        audio_filename: '',
                        transcript_filename: '',
                        transcript_text: '',
                        paper_filename: sub.paper_filename,
                        created_at: sub.created_at,
                        feedback_items: [],
                      };
                      activeCenterTab = 'feedback';
                    "
                  >
                    <div class="flex items-start justify-between gap-2">
                      <div class="min-w-0">
                        <p class="truncate text-sm font-medium text-stone-200">{{ sub.student_name }}</p>
                        <p class="mt-0.5 text-[11px] text-stone-600">{{ sub.nim }}</p>
                      </div>
                      <UiBadge :text="sub.status" :color="submissionStatusColor(sub.status)" />
                    </div>
                    <p class="mt-1.5 truncate text-xs text-stone-500">{{ sub.paper_filename }}</p>
                    <p class="mt-1 text-[11px] text-stone-700">{{ sub.pending_count }} pending</p>
                  </div>
                  <div v-if="studentSubmissions.length === 0" class="py-12 text-center">
                    <p class="text-sm text-stone-600">No submissions</p>
                  </div>
                </div>
              </UiCard>
            </div>

            <div class="flex-[1.4]">
              <UiCard class="flex h-full flex-col overflow-hidden" :padding="false">
                <div v-if="!selectedConsultation" class="flex flex-1 flex-col items-center justify-center p-5">
                  <div class="flex h-12 w-12 items-center justify-center rounded-xl bg-[#1c1c1e]">
                    <ConsultationIcon :size="22" color="#52525b" />
                  </div>
                  <p class="mt-3 text-sm text-stone-600">Select a student submission</p>
                </div>

                <template v-else>
                  <div class="border-b border-white/[0.04] p-4 flex items-center justify-between gap-3 shrink-0">
                    <h4 class="text-sm font-semibold text-stone-100 font-[family-name:var(--font-display)] truncate">
                      {{ selectedConsultation.paper_filename }}
                    </h4>
                    <div class="flex gap-2 shrink-0">
                      <button
                        class="inline-flex items-center gap-1.5 rounded-lg bg-[#1c1c1e] px-3 py-1.5 text-xs font-medium text-stone-300 transition-colors hover:bg-[#2c2c2e]"
                        @click="generateTranscript"
                      >
                        <svg class="h-3.5 w-3.5 text-stone-500" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                          <polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2" />
                        </svg>
                        Generate Transcript
                      </button>
                      <button
                        class="inline-flex items-center gap-1.5 rounded-lg bg-amber-500/10 px-3 py-1.5 text-xs font-medium text-amber-400 transition-colors hover:bg-amber-500/15"
                        @click="analyzeFeedback"
                      >
                        <svg class="h-3.5 w-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                          <circle cx="11" cy="11" r="8" />
                          <line x1="21" y1="21" x2="16.65" y2="16.65" />
                        </svg>
                        Analyze Feedback
                      </button>
                    </div>
                  </div>

                  <div class="flex-1 scroll-area p-5 space-y-4">
                    <div class="rounded-xl border border-white/[0.04] bg-[#0a0a0b] p-4">
                      <div class="flex items-center gap-3">
                        <div class="h-2 w-2 rounded-full bg-amber-500" />
                        <div class="h-1.5 flex-1 overflow-hidden rounded-full bg-[#1c1c1e]">
                          <div class="h-full w-[35%] rounded-full bg-amber-500/60" />
                        </div>
                        <span class="text-[11px] text-stone-600">0:42 / 2:05</span>
                      </div>
                      <div class="mt-3 flex items-center justify-center gap-4">
                        <button class="rounded-full p-2 text-stone-500 transition-colors hover:text-stone-200">
                          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                            <polygon points="19 20 9 12 19 4 19 20" />
                            <line x1="5" y1="19" x2="5" y2="5" />
                          </svg>
                        </button>
                        <button class="rounded-full bg-amber-500 p-3 text-[#0a0a0b] transition-colors hover:bg-amber-400">
                          <svg width="18" height="18" viewBox="0 0 24 24" fill="currentColor" stroke="currentColor" stroke-width="0">
                            <polygon points="5 3 19 12 5 21 5 3" />
                          </svg>
                        </button>
                        <button class="rounded-full p-2 text-stone-500 transition-colors hover:text-stone-200">
                          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                            <polygon points="5 4 15 12 5 20 5 4" />
                            <line x1="19" y1="5" x2="19" y2="19" />
                          </svg>
                        </button>
                      </div>
                    </div>

                    <div class="rounded-xl border border-white/[0.04] bg-[#0a0a0b] p-4">
                      <h5 class="mb-2 text-[11px] font-semibold uppercase tracking-wider text-stone-600 font-[family-name:var(--font-display)]">Transcript</h5>
                      <div class="max-h-44 scroll-area">
                        <p
                          v-if="selectedConsultation.transcript_text"
                          class="whitespace-pre-wrap text-sm leading-[1.8] text-stone-500"
                        >
                          {{ selectedConsultation.transcript_text }}
                        </p>
                        <p v-else class="py-6 text-center text-sm text-stone-600">No transcript generated yet.</p>
                      </div>
                    </div>

                    <div>
                      <h5 class="mb-3 text-[11px] font-semibold uppercase tracking-wider text-stone-600 font-[family-name:var(--font-display)]">Feedback Items</h5>
                      <div class="space-y-2">
                        <div
                          v-for="item in feedbackItems"
                          :key="item.id"
                          class="rounded-xl border border-white/[0.04] bg-[#141415] p-3.5"
                        >
                          <p class="text-sm text-stone-300">{{ item.content }}</p>
                          <div class="mt-2 flex items-center gap-1.5">
                            <UiBadge :text="item.category" :color="categoryColor(item.category)" />
                            <UiBadge :text="item.status" :color="statusColor(item.status)" />
                          </div>
                        </div>
                        <div v-if="feedbackItems.length === 0" class="py-8 text-center">
                          <p class="text-sm text-stone-600">No feedback items.</p>
                        </div>
                      </div>
                    </div>
                  </div>
                </template>
              </UiCard>
            </div>

            <div class="flex-[1.4]">
              <UiCard class="flex h-full flex-col overflow-hidden" :padding="false">
                <div class="border-b border-white/[0.04] p-4">
                  <h3 class="text-sm font-semibold text-stone-100 font-[family-name:var(--font-display)]">Feedback Evaluation</h3>
                  <p class="mt-0.5 text-[11px] text-stone-600">Review, edit, and validate feedback items</p>
                </div>

                <div class="flex-1 scroll-area p-5">
                  <div v-if="feedbackItems.length === 0" class="flex flex-col items-center justify-center py-12">
                    <div class="flex h-10 w-10 items-center justify-center rounded-xl bg-[#1c1c1e]">
                      <ClockIcon :size="18" color="#52525b" />
                    </div>
                    <p class="mt-3 text-sm text-stone-600">Select a session with feedback items</p>
                  </div>

                  <template v-else>
                    <div class="mb-4">
                      <label class="mb-1.5 block text-xs font-medium text-stone-500">Select Item</label>
                      <select
                        class="w-full rounded-xl border border-white/[0.04] bg-[#0a0a0b] px-4 py-2.5 text-sm text-stone-200 transition-colors focus:border-white/[0.12] focus:outline-none focus:ring-1 focus:ring-white/10"
                        @change="startEvaluate(feedbackItems[($event.target as HTMLSelectElement).selectedIndex])"
                      >
                        <option value="" disabled selected>Choose a feedback item</option>
                        <option v-for="item in feedbackItems" :key="item.id" :value="item.id">
                          #{{ item.id }} — {{ item.content.substring(0, 40) }}...
                        </option>
                      </select>
                    </div>

                    <div v-if="evaluatingFeedbackId" class="space-y-4">
                      <div>
                        <label class="mb-1.5 block text-xs font-medium text-stone-500">Content</label>
                        <textarea
                          v-model="evalContent"
                          rows="4"
                          class="w-full resize-none rounded-xl border border-white/[0.04] bg-[#0a0a0b] px-4 py-3 text-sm text-stone-200 placeholder-zinc-500 transition-colors focus:border-white/[0.12] focus:outline-none focus:ring-1 focus:ring-white/10"
                        />
                      </div>

                      <div>
                        <label class="mb-1.5 block text-xs font-medium text-stone-500">Category</label>
                        <div class="flex gap-2">
                          <button
                            :class="[
                              'rounded-lg px-4 py-2 text-xs font-medium transition-colors',
                              evalCategory === 'Major'
                                ? 'bg-red-500/10 text-red-400 ring-1 ring-red-500/30'
                                : 'bg-[#1c1c1e] text-stone-500 hover:text-stone-300',
                            ]"
                            @click="evalCategory = 'Major'"
                          >
                            Major
                          </button>
                          <button
                            :class="[
                              'rounded-lg px-4 py-2 text-xs font-medium transition-colors',
                              evalCategory === 'Minor'
                                ? 'bg-amber-500/10 text-amber-400 ring-1 ring-amber-500/30'
                                : 'bg-[#1c1c1e] text-stone-500 hover:text-stone-300',
                            ]"
                            @click="evalCategory = 'Minor'"
                          >
                            Minor
                          </button>
                        </div>
                      </div>

                      <div class="flex gap-2">
                        <UiButton
                          title="Validate"
                          tone="success"
                          :disabled="!evaluatingFeedbackId"
                          @click="validateFeedback(feedbackItems.find((f) => f.id === evaluatingFeedbackId)!)!"
                        />
                        <UiButton
                          title="Reject"
                          tone="danger"
                          :disabled="!evaluatingFeedbackId"
                          @click="rejectFeedback(feedbackItems.find((f) => f.id === evaluatingFeedbackId)!)!"
                        />
                        <UiButton
                          title="Undo"
                          tone="warning"
                          :disabled="!evaluatingFeedbackId"
                          @click="undoFeedbackAction(feedbackItems.find((f) => f.id === evaluatingFeedbackId)!)!"
                        />
                      </div>

                      <UiButton
                        title="Update Feedback"
                        :disabled="!evaluatingFeedbackId"
                        @click="submitEvaluation"
                      />
                    </div>
                  </template>
                </div>
              </UiCard>
            </div>
          </div>
        </template>
      </div>

      <div class="pointer-events-none fixed right-4 top-20 z-[100] flex flex-col gap-2">
        <TransitionGroup
          enter-active-class="transition duration-300 ease-out"
          enter-from-class="translate-x-full opacity-0"
          enter-to-class="translate-x-0 opacity-100"
          leave-active-class="transition duration-300 ease-in"
          leave-from-class="translate-x-0 opacity-100"
          leave-to-class="translate-x-full opacity-0"
        >
          <div
            v-for="toast in toasts"
            :key="toast.id"
            :class="[
              'pointer-events-auto flex items-center gap-3 rounded-xl border px-4 py-3 shadow-2xl backdrop-blur-xl',
              toast.type === 'success'
                ? 'border-emerald-500/20 bg-emerald-500/10 text-emerald-300'
                : toast.type === 'error'
                  ? 'border-red-500/20 bg-red-500/10 text-red-300'
                  : 'border-amber-500/20 bg-amber-500/10 text-amber-300',
            ]"
          >
            <CheckCircleIcon v-if="toast.type === 'success'" :size="15" />
            <AlertIcon v-else-if="toast.type === 'error'" :size="15" />
            <ClockIcon v-else :size="15" />
            <span class="text-sm font-medium">{{ toast.message }}</span>
          </div>
        </TransitionGroup>
      </div>
    </UiPage>
  </RequireAuth>
</template>
