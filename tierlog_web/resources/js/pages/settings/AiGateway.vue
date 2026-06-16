<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue';
import { useAuthStore } from '@/stores/auth';
import RequireAuth from '@/components/RequireAuth.vue';
import NavBar from '@/components/NavBar.vue';
import UiPage from '@/components/UiPage.vue';
import UiHeading from '@/components/UiHeading.vue';
import UiCard from '@/components/UiCard.vue';
import UiButton from '@/components/UiButton.vue';
import UiBadge from '@/components/UiBadge.vue';
import { CheckCircleIcon } from '@/components/icons';

type ProviderKey = 'groq' | 'gemini' | 'openai' | 'anthropic' | 'nvidia' | 'custom';

interface ModelOption {
  id: string;
  label: string;
  speed?: string;
}

const PROVIDER_MODELS: Record<string, ModelOption[]> = {
  groq: [
    { id: 'groq:llama-3.3-70b-versatile', label: 'Llama 3.3 70B', speed: 'Fast' },
    { id: 'groq:llama-3.1-8b-instant', label: 'Llama 3.1 8B', speed: 'Instant' },
    { id: 'groq:gemma2-9b-it', label: 'Gemma 2 9B', speed: 'Fast' },
    { id: 'groq:mixtral-8x7b-32768', label: 'Mixtral 8x7B', speed: 'Fast' },
  ],
  gemini: [
    { id: 'gemini:gemini-2.5-flash', label: 'Gemini 2.5 Flash', speed: 'Fast' },
    { id: 'gemini:gemini-2.5-pro', label: 'Gemini 2.5 Pro', speed: 'Advanced' },
    { id: 'gemini:gemini-2.0-flash', label: 'Gemini 2.0 Flash', speed: 'Instant' },
  ],
  openai: [
    { id: 'openai:gpt-4o', label: 'GPT-4o', speed: 'Balanced' },
    { id: 'openai:gpt-4o-mini', label: 'GPT-4o Mini', speed: 'Fast' },
  ],
  anthropic: [
    { id: 'anthropic:claude-3-5-sonnet-20241022', label: 'Claude 3.5 Sonnet', speed: 'Advanced' },
    { id: 'anthropic:claude-3-5-haiku-20241022', label: 'Claude 3.5 Haiku', speed: 'Fast' },
  ],
};

const auth = useAuthStore();
const activeTab = ref<ProviderKey>('groq');
const selectedModel = ref('');
const customModelId = ref('');
const nvidiaModels = ref<ModelOption[]>([]);
const nvidiaLoading = ref(false);
const saving = ref(false);
const saveSuccess = ref(false);

const openaiKey = ref('');
const geminiKey = ref('');
const anthropicKey = ref('');
const nvidiaKey = ref('');
const groqKey = ref('');

// Pagination & Filtering for NVIDIA models
const currentPage = ref(1);
const itemsPerPage = 12;

const isProviderKeyActive = computed(() => {
  if (activeTab.value === 'groq') return groqKey.value.trim().length > 0;
  if (activeTab.value === 'gemini') return geminiKey.value.trim().length > 0;
  if (activeTab.value === 'openai') return openaiKey.value.trim().length > 0;
  if (activeTab.value === 'anthropic') return anthropicKey.value.trim().length > 0;
  if (activeTab.value === 'nvidia') return nvidiaKey.value.trim().length > 0;
  if (activeTab.value === 'custom') return true;
  return false;
});

const activeProviderKeyRef = computed({
  get() {
    if (activeTab.value === 'groq') return groqKey.value;
    if (activeTab.value === 'gemini') return geminiKey.value;
    if (activeTab.value === 'openai') return openaiKey.value;
    if (activeTab.value === 'anthropic') return anthropicKey.value;
    if (activeTab.value === 'nvidia') return nvidiaKey.value;
    return '';
  },
  set(val: string) {
    if (activeTab.value === 'groq') groqKey.value = val;
    else if (activeTab.value === 'gemini') geminiKey.value = val;
    else if (activeTab.value === 'openai') openaiKey.value = val;
    else if (activeTab.value === 'anthropic') anthropicKey.value = val;
    else if (activeTab.value === 'nvidia') nvidiaKey.value = val;
  }
});

const providerLabel = computed(() => {
  const found = tabs.find(t => t.key === activeTab.value);
  return found ? found.label : '';
});

const providerPlaceholder = computed(() => {
  if (activeTab.value === 'groq') return 'gsk_...';
  if (activeTab.value === 'gemini') return 'AIza...';
  if (activeTab.value === 'openai') return 'sk-...';
  if (activeTab.value === 'anthropic') return 'sk-ant-...';
  if (activeTab.value === 'nvidia') return 'nvapi-...';
  return '';
});

const nvidiaVisionModels = computed(() => {
  return nvidiaModels.value.filter(m => m.speed === 'Vision & Analysis').slice(0, 10);
});

const nvidiaOtherModels = computed(() => {
  return nvidiaModels.value.filter(m => m.speed !== 'Vision & Analysis');
});

const totalPages = computed(() => {
  return Math.max(1, Math.ceil(nvidiaOtherModels.value.length / itemsPerPage));
});

const paginatedNvidiaOtherModels = computed(() => {
  const start = (currentPage.value - 1) * itemsPerPage;
  const end = start + itemsPerPage;
  return nvidiaOtherModels.value.slice(start, end);
});

const visiblePages = computed(() => {
  const total = totalPages.value;
  const current = currentPage.value;
  const delta = 1;
  const range = [];
  const rangeWithDots: (number | string)[] = [];
  let l: number | undefined;

  for (let i = 1; i <= total; i++) {
    if (i === 1 || i === total || (i >= current - delta && i <= current + delta)) {
      range.push(i);
    }
  }

  for (const i of range) {
    if (l !== undefined) {
      if (i - l === 2) {
        rangeWithDots.push(l + 1);
      } else if (i - l > 2) {
        rangeWithDots.push('...');
      }
    }
    rangeWithDots.push(i);
    l = i;
  }

  return rangeWithDots;
});

function setPage(page: number | string) {
  if (typeof page === 'number') {
    currentPage.value = page;
  }
}

watch(nvidiaModels, () => {
  currentPage.value = 1;
});

const currentModels = computed(() => {
  if (activeTab.value === 'custom') return [];
  if (activeTab.value === 'nvidia') return nvidiaModels.value;
  return PROVIDER_MODELS[activeTab.value] ?? [];
});

const activeModelLabel = computed(() => {
  if (!selectedModel.value) return 'None selected';
  if (activeTab.value === 'custom') return customModelId.value || 'Custom';
  const flat = [...Object.values(PROVIDER_MODELS).flat(), ...nvidiaModels.value];
  const found = flat.find((m) => m.id === selectedModel.value);
  return found?.label ?? selectedModel.value;
});

const tabs: { key: ProviderKey; label: string }[] = [
  { key: 'groq', label: 'Groq' },
  { key: 'gemini', label: 'Gemini' },
  { key: 'openai', label: 'OpenAI' },
  { key: 'anthropic', label: 'Anthropic' },
  { key: 'nvidia', label: 'NVIDIA' },
  { key: 'custom', label: 'Custom' },
];

const sectionTab = ref<'provider' | 'credentials'>('provider');

function keyStatus(key: string | undefined): { text: string; color: string } {
  if (key && key.length > 0) {
    return { text: 'Connected', color: 'bg-emerald-500/10 text-emerald-400' };
  }
  return { text: 'Not set', color: 'bg-zinc-800 text-stone-700' };
}

function populateFromUser() {
  const u = auth.user;
  if (!u) return;
  openaiKey.value = u.openai_key ?? '';
  geminiKey.value = u.gemini_key ?? '';
  anthropicKey.value = u.anthropic_key ?? '';
  nvidiaKey.value = u.nvidia_key ?? '';
  groqKey.value = u.groq_key ?? '';
  selectedModel.value = u.preferred_model ?? '';

  if (u.preferred_model) {
    if (u.preferred_model.startsWith('groq:')) activeTab.value = 'groq';
    else if (u.preferred_model.startsWith('gemini:')) activeTab.value = 'gemini';
    else if (u.preferred_model.startsWith('openai:')) activeTab.value = 'openai';
    else if (u.preferred_model.startsWith('anthropic:')) activeTab.value = 'anthropic';
    else if (u.preferred_model.startsWith('nvidia:')) activeTab.value = 'nvidia';
    else activeTab.value = 'custom';
    selectedModel.value = u.preferred_model;
  }
}

async function fetchNvidiaModels() {
  if (!nvidiaKey.value || nvidiaKey.value.length < 8) {
    nvidiaModels.value = [];
    return;
  }
  nvidiaLoading.value = true;
  const res = await auth.api<{ models: string[] }>(
    `/api/ai/models?provider=nvidia&api_key=${encodeURIComponent(nvidiaKey.value)}`
  );
  if (res.ok && res.data?.models) {
    nvidiaModels.value = res.data.models.map((m) => {
      let label = m;
      if (m.includes('/')) label = m.split('/').pop()!;
      label = label.replace(/-/g, ' ').replace(/\b\w/g, (c) => c.toUpperCase()).trim();

      const supportsVision = m.toLowerCase().includes('vision') || 
                             m.toLowerCase().includes('vl') || 
                             m.toLowerCase().includes('neva') || 
                             m.toLowerCase().includes('multimodal');

      return { 
        id: `nvidia:${m}`, 
        label, 
        speed: supportsVision ? 'Vision & Analysis' : 'NVIDIA' 
      };
    });
    nvidiaModels.value.sort((a, b) => {
      const aVision = a.speed === 'Vision & Analysis';
      const bVision = b.speed === 'Vision & Analysis';
      if (aVision && !bVision) return -1;
      if (!aVision && bVision) return 1;
      return a.label.localeCompare(b.label);
    });
  } else {
    nvidiaModels.value = [];
  }
  nvidiaLoading.value = false;
}

async function selectModel(modelId: string) {
  selectedModel.value = modelId;
  saving.value = true;
  await auth.api('/settings/ai-gateway', {
    method: 'PATCH',
    body: JSON.stringify({ preferred_model: modelId }),
  });
  if (auth.user) auth.user.preferred_model = modelId;
  saving.value = false;
}

function applyCustomModel() {
  if (customModelId.value.trim()) {
    selectModel(customModelId.value.trim());
  }
}

async function saveAllKeys() {
  saving.value = true;
  const body = {
    groq_key: groqKey.value,
    gemini_key: geminiKey.value,
    openai_key: openaiKey.value,
    anthropic_key: anthropicKey.value,
    nvidia_key: nvidiaKey.value,
  };

  const res = await auth.api('/settings/ai-gateway', {
    method: 'PATCH',
    body: JSON.stringify(body),
  });

  if (res.ok && auth.user) {
    if (groqKey.value) (auth.user as Record<string, unknown>).groq_key = groqKey.value;
    if (geminiKey.value) (auth.user as Record<string, unknown>).gemini_key = geminiKey.value;
    if (openaiKey.value) (auth.user as Record<string, unknown>).openai_key = openaiKey.value;
    if (anthropicKey.value) (auth.user as Record<string, unknown>).anthropic_key = anthropicKey.value;
    if (nvidiaKey.value) (auth.user as Record<string, unknown>).nvidia_key = nvidiaKey.value;
    saveSuccess.value = true;
    setTimeout(() => { saveSuccess.value = false; }, 3000);

    if (activeTab.value === 'nvidia') {
      fetchNvidiaModels();
    }
  }
  saving.value = false;
}

const keyFields = [
  { field: 'groq_key', label: 'Groq', placeholder: 'gsk_...', key: groqKey, purpose: 'audio' },
  { field: 'gemini_key', label: 'Gemini', placeholder: 'AIza...', key: geminiKey, purpose: 'llm' },
  { field: 'openai_key', label: 'OpenAI', placeholder: 'sk-...', key: openaiKey, purpose: 'llm' },
  { field: 'anthropic_key', label: 'Anthropic', placeholder: 'sk-ant-...', key: anthropicKey, purpose: 'llm' },
  { field: 'nvidia_key', label: 'NVIDIA NIM', placeholder: 'nvapi-...', key: nvidiaKey, purpose: 'llm' },
];

const hasGroqKey = computed(() => groqKey.value.length > 0);
const hasLLMKey = computed(() =>
  geminiKey.value.length > 0 ||
  openaiKey.value.length > 0 ||
  anthropicKey.value.length > 0 ||
  nvidiaKey.value.length > 0 ||
  groqKey.value.length > 0
);

watch(activeTab, (tab) => {
  if (tab === 'nvidia' && nvidiaModels.value.length === 0 && nvidiaKey.value.length >= 8) {
    fetchNvidiaModels();
  }
});

watch(nvidiaKey, (newVal) => {
  if (newVal && newVal.length >= 8 && activeTab.value === 'nvidia') {
    fetchNvidiaModels();
  } else if (!newVal || newVal.length < 8) {
    nvidiaModels.value = [];
  }
});

onMounted(() => {
  populateFromUser();
});
</script>

<template>
  <RequireAuth>
    <NavBar />
    <UiPage>
      <div class="space-y-6">
        <UiHeading title="AI Gateway" subtitle="Configure your AI providers and preferred model." title-class="font-[family-name:var(--font-display)]" />

        <UiCard>
          <div class="space-y-0 p-5 sm:p-6">
            <!-- Section Tabs -->
            <div class="flex gap-0.5 rounded-lg bg-[#1c1c1e] p-1 mb-5">
              <button
                :class="[
                  'flex-1 rounded-md px-4 py-2 text-[13px] font-medium transition-colors duration-100',
                  sectionTab === 'provider'
                    ? 'bg-[#2c2c2e] text-stone-100'
                    : 'text-stone-600 hover:text-stone-300',
                ]"
                @click="sectionTab = 'provider'"
              >
                Model Provider
              </button>
              <button
                :class="[
                  'flex-1 rounded-md px-4 py-2 text-[13px] font-medium transition-colors duration-100',
                  sectionTab === 'credentials'
                    ? 'bg-[#2c2c2e] text-stone-100'
                    : 'text-stone-600 hover:text-stone-300',
                ]"
                @click="sectionTab = 'credentials'"
              >
                API Credentials
              </button>
            </div>

            <!-- Provider Section -->
            <div v-if="sectionTab === 'provider'" class="space-y-5">
              <div class="border-b border-white/[0.04] pb-4">
                <h3 class="text-sm font-semibold text-stone-200">Select a provider and model for AI-powered features</h3>
              </div>

              <div class="flex flex-wrap gap-1 rounded-lg bg-[#1c1c1e] p-1">
                <button
                  v-for="tab in tabs"
                  :key="tab.key"
                  :class="[
                    'rounded-md px-3 py-1.5 text-[13px] font-medium transition-colors duration-100',
                    activeTab === tab.key
                      ? 'bg-[#2c2c2e] text-stone-100'
                      : 'text-stone-600 hover:text-stone-300',
                  ]"
                  @click="activeTab = tab.key"
                >
                  {{ tab.label }}
                </button>
              </div>

              <div class="flex items-center gap-2 rounded-lg bg-[#1c1c1e]/50 px-3.5 py-2">
                <span class="text-[12px] text-stone-600">Active:</span>
                <span class="text-[12px] font-medium text-amber-400">{{ activeModelLabel }}</span>
              </div>

              <template v-if="activeTab !== 'custom'">
                <!-- Inline API Key Input when not active -->
                <div v-if="!isProviderKeyActive" class="flex flex-col items-center justify-center p-6 text-center bg-[#1c1c1e]/30 rounded-xl border border-white/[0.04] space-y-4">
                  <div class="p-3 rounded-full bg-amber-500/10 text-amber-400">
                    <svg class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                      <path stroke-linecap="round" stroke-linejoin="round" d="M15 7a2 2 0 012 2m-9 5a3 3 0 11-6 0 3 3 0 016 0zm6 3l2 2m6-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                    </svg>
                  </div>
                  <div>
                    <h4 class="text-sm font-semibold text-stone-200">API Key Required</h4>
                    <p class="text-xs text-stone-500 mt-1 max-w-sm">
                      Please enter your {{ providerLabel }} API key below to unlock and automatically show models.
                    </p>
                  </div>
                  <div class="flex flex-col sm:flex-row gap-2 w-full max-w-md">
                    <input
                      v-model="activeProviderKeyRef"
                      type="password"
                      :placeholder="providerPlaceholder"
                      class="flex-1 rounded-lg border border-white/[0.04] bg-[#1c1c1e] px-3.5 py-2 text-sm text-stone-100 placeholder-stone-700 transition-colors duration-150 focus:border-amber-500/50 focus:outline-none focus:ring-1 focus:ring-amber-500/20"
                    />
                    <UiButton 
                      title="Save & Connect" 
                      tone="primary" 
                      :loading="saving" 
                      @click="saveAllKeys"
                    />
                  </div>
                </div>

                <!-- Models list when active -->
                <div v-else>
                  <div v-if="activeTab === 'nvidia' && nvidiaLoading" class="flex items-center gap-2 py-4">
                    <div class="h-4 w-4 animate-spin rounded-full border-2 border-amber-500 border-t-transparent" />
                    <span class="text-xs text-stone-600">Loading models...</span>
                  </div>

                  <template v-else-if="activeTab === 'nvidia'">
                    <div class="space-y-6">
                      <!-- Vision Models Section -->
                      <div class="space-y-3">
                        <div class="flex items-center gap-2 border-b border-white/[0.02] pb-2">
                          <span class="text-[11px] font-semibold text-purple-400 uppercase tracking-wider">Vision & Analysis Models (Recommended)</span>
                          <span class="text-[10px] bg-purple-500/10 text-purple-400 px-1.5 py-0.5 rounded font-medium border border-purple-500/20">Needs Vision</span>
                        </div>
                        <div v-if="nvidiaVisionModels.length === 0" class="text-xs text-stone-600 italic">No Vision & Analysis models found.</div>
                        <div v-else class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-2">
                          <button
                            v-for="model in nvidiaVisionModels"
                            :key="model.id"
                            :class="[
                              'flex items-center gap-2 rounded-lg border px-3 py-2 text-left transition-all duration-150',
                              selectedModel === model.id
                                ? 'border-amber-500/30 bg-amber-500/5 ring-1 ring-amber-500/10'
                                : 'border-white/[0.04] bg-[#1c1c1e]/50 hover:border-white/[0.12]',
                            ]"
                            @click="selectModel(model.id)"
                          >
                            <CheckCircleIcon
                              v-if="selectedModel === model.id"
                              :size="14"
                              color="#f59e0b"
                            />
                            <div v-else class="h-3.5 w-3.5 rounded-full border border-zinc-600" />
                            <span
                              :class="[
                                'text-[13px] font-medium truncate',
                                selectedModel === model.id ? 'text-amber-300' : 'text-stone-300',
                              ]"
                              :title="model.label"
                            >{{ model.label }}</span>
                          </button>
                        </div>
                      </div>

                      <!-- Other Models Section with Pagination -->
                      <div class="space-y-3">
                        <div class="flex items-center justify-between border-b border-white/[0.02] pb-2">
                          <span class="text-[11px] font-semibold text-stone-400 uppercase tracking-wider">Other Models</span>
                          <span class="text-[10px] bg-zinc-800 text-stone-400 px-1.5 py-0.5 rounded font-medium">{{ nvidiaOtherModels.length }} models</span>
                        </div>
                        
                        <div v-if="nvidiaOtherModels.length === 0" class="text-xs text-stone-600 italic">No other models loaded.</div>
                        <template v-else>
                          <div class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-2">
                            <button
                              v-for="model in paginatedNvidiaOtherModels"
                              :key="model.id"
                              :class="[
                                'flex items-center gap-2 rounded-lg border px-3 py-2 text-left transition-all duration-150',
                                selectedModel === model.id
                                  ? 'border-amber-500/30 bg-amber-500/5 ring-1 ring-amber-500/10'
                                  : 'border-white/[0.04] bg-[#1c1c1e]/50 hover:border-white/[0.12]',
                              ]"
                              @click="selectModel(model.id)"
                            >
                              <CheckCircleIcon
                                v-if="selectedModel === model.id"
                                :size="14"
                                color="#f59e0b"
                              />
                              <div v-else class="h-3.5 w-3.5 rounded-full border border-zinc-600" />
                              <span
                                :class="[
                                  'text-[13px] font-medium truncate',
                                  selectedModel === model.id ? 'text-amber-300' : 'text-stone-300',
                                ]"
                                :title="model.label"
                              >{{ model.label }}</span>
                            </button>
                          </div>

                          <!-- Pagination Controls -->
                          <div v-if="totalPages > 1" class="mt-4 flex items-center justify-center gap-1.5 py-2">
                            <button
                              type="button"
                              :disabled="currentPage === 1"
                              class="px-2.5 py-1.5 rounded-lg text-xs font-medium text-stone-400 hover:text-stone-100 disabled:opacity-30 disabled:hover:text-stone-400 bg-[#1c1c1e] border border-white/[0.02] hover:border-white/[0.08] transition-colors"
                              @click="currentPage > 1 && currentPage--"
                            >
                              Prev
                            </button>
                            
                            <button
                              v-for="(page, idx) in visiblePages"
                              :key="idx"
                              type="button"
                              :disabled="page === '...'"
                              :class="[
                                'px-3 py-1.5 rounded-lg text-xs font-medium transition-all',
                                page === currentPage
                                  ? 'bg-amber-500/20 text-amber-300 border border-amber-500/30 font-semibold'
                                  : page === '...'
                                  ? 'text-stone-600 cursor-default'
                                  : 'text-stone-400 hover:text-stone-100 bg-[#1c1c1e] border border-white/[0.02] hover:border-white/[0.08]'
                              ]"
                              @click="setPage(page)"
                            >
                              {{ page }}
                            </button>

                            <button
                              type="button"
                              :disabled="currentPage === totalPages"
                              class="px-2.5 py-1.5 rounded-lg text-xs font-medium text-stone-400 hover:text-stone-100 disabled:opacity-30 disabled:hover:text-stone-400 bg-[#1c1c1e] border border-white/[0.02] hover:border-white/[0.08] transition-colors"
                              @click="currentPage < totalPages && currentPage++"
                            >
                              Next
                            </button>
                          </div>
                        </template>
                      </div>
                    </div>
                  </template>

                  <!-- Standard Model Grid for Other Providers -->
                  <div v-else class="flex flex-wrap gap-2">
                    <button
                      v-for="model in currentModels"
                      :key="model.id"
                      :class="[
                        'flex items-center gap-2 rounded-lg border px-3 py-2 text-left transition-all duration-150',
                        selectedModel === model.id
                          ? 'border-amber-500/30 bg-amber-500/5 ring-1 ring-amber-500/10'
                          : 'border-white/[0.04] bg-[#1c1c1e]/50 hover:border-white/[0.12]',
                      ]"
                      @click="selectModel(model.id)"
                    >
                      <CheckCircleIcon
                        v-if="selectedModel === model.id"
                        :size="14"
                        color="#f59e0b"
                      />
                      <div v-else class="h-3.5 w-3.5 rounded-full border border-zinc-600" />
                      <span
                        :class="[
                          'text-[13px] font-medium',
                          selectedModel === model.id ? 'text-amber-300' : 'text-stone-300',
                        ]"
                      >{{ model.label }}</span>
                      <span 
                        v-if="model.speed" 
                        :class="[
                          'rounded px-1.5 py-0.5 text-[10px]',
                          model.speed.includes('Vision') 
                            ? 'bg-purple-500/10 text-purple-400 font-semibold border border-purple-500/20' 
                            : 'bg-zinc-800 text-stone-600'
                        ]"
                      >
                        {{ model.speed }}
                      </span>
                    </button>
                  </div>

                  <div v-if="activeTab === 'nvidia' && !nvidiaLoading" class="mt-4 flex items-center justify-between border-t border-white/[0.02] pt-4">
                    <span class="text-xs text-stone-600">
                      {{ nvidiaModels.length > 0 ? `${nvidiaModels.length} models loaded` : 'No models loaded' }}
                    </span>
                    <button
                      type="button"
                      class="rounded-lg bg-zinc-800/80 hover:bg-zinc-700 px-3 py-1.5 text-xs font-semibold text-stone-300 hover:text-stone-100 border border-white/[0.04] hover:border-white/[0.12] transition-colors"
                      @click="fetchNvidiaModels"
                    >
                      Load/Refresh Models
                    </button>
                  </div>
                </div>
              </template>

              <template v-if="activeTab === 'custom'">
                <div class="space-y-2">
                  <p class="text-xs font-medium text-stone-500">Custom Model ID</p>
                  <div class="flex gap-2">
                    <input
                      v-model="customModelId"
                      type="text"
                      placeholder="provider:model-name"
                      class="flex-1 rounded-lg border border-white/[0.04] bg-[#1c1c1e] px-3.5 py-2 text-sm text-stone-100 placeholder-stone-700 transition-colors duration-150 focus:border-amber-500/50 focus:outline-none focus:ring-1 focus:ring-amber-500/20"
                    />
                    <UiButton title="Apply" tone="secondary" @click="applyCustomModel" />
                  </div>
                </div>
              </template>
            </div>

            <!-- Credentials Section -->
            <div v-if="sectionTab === 'credentials'" class="space-y-4">
              <div class="border-b border-white/[0.04] pb-4">
                <h3 class="text-sm font-semibold text-stone-200">Enter your personal API keys — stored securely per account</h3>
              </div>

              <div class="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
                <div v-for="kf in keyFields" :key="kf.field" class="space-y-1.5">
                  <div class="flex items-center justify-between">
                    <label class="text-[12px] font-medium text-stone-500">{{ kf.label }}</label>
                    <UiBadge :text="keyStatus(kf.key.value).text" :color="keyStatus(kf.key.value).color" />
                  </div>
                  <input
                    v-model="kf.key.value"
                    type="password"
                    :placeholder="kf.placeholder"
                    class="w-full rounded-lg border border-white/[0.04] bg-[#1c1c1e] px-3 py-2 text-[13px] text-stone-100 placeholder-stone-700 transition-colors duration-150 focus:border-amber-500/50 focus:outline-none focus:ring-1 focus:ring-amber-500/20"
                  />
                </div>
              </div>

              <div class="flex items-center justify-between border-t border-white/[0.04] pt-4">
                <transition name="fade">
                  <div
                    v-if="saveSuccess"
                    class="flex items-center gap-2 text-[13px] text-emerald-400"
                  >
                    <svg class="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                      <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14" />
                      <polyline points="22 4 12 14.01 9 11.01" />
                    </svg>
                    API keys saved successfully!
                  </div>
                  <span v-else class="text-[12px] text-stone-700">Keys are encrypted before storage</span>
                </transition>

                <UiButton
                  id="save-api-keys-btn"
                  title="Save API Keys"
                  :loading="saving"
                  @click="saveAllKeys"
                />
              </div>
            </div>
          </div>
        </UiCard>

        <div v-if="!hasGroqKey || !hasLLMKey" class="space-y-3">
          <div
            v-if="!hasGroqKey"
            class="flex items-start gap-3 rounded-xl border border-amber-500/20 bg-amber-500/[0.04] px-4 py-3"
          >
            <svg class="mt-0.5 h-4 w-4 shrink-0 text-amber-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z" />
              <line x1="12" y1="9" x2="12" y2="13" />
              <line x1="12" y1="17" x2="12.01" y2="17" />
            </svg>
            <div>
              <p class="text-[13px] font-medium text-amber-300">Groq API key required for audio transcription</p>
              <p class="mt-0.5 text-[12px] leading-relaxed text-stone-600">
                Without a Groq API key, audio-to-text transcription will not work.
                Students won't be able to upload audio recordings for automatic transcription.
              </p>
            </div>
          </div>

          <div
            v-if="!hasLLMKey"
            class="flex items-start gap-3 rounded-xl border border-amber-500/20 bg-amber-500/[0.04] px-4 py-3"
          >
            <svg class="mt-0.5 h-4 w-4 shrink-0 text-amber-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z" />
              <line x1="12" y1="9" x2="12" y2="13" />
              <line x1="12" y1="17" x2="12.01" y2="17" />
            </svg>
            <div>
              <p class="text-[13px] font-medium text-amber-300">LLM API key required for AI features</p>
              <p class="mt-0.5 text-[12px] leading-relaxed text-stone-600">
                Without at least one LLM provider key (Groq, Gemini, OpenAI, Anthropic, or NVIDIA),
                the following features will not work:
              </p>
              <ul class="mt-1.5 space-y-0.5 text-[12px] text-stone-600">
                <li class="flex items-center gap-1.5">
                  <span class="h-1 w-1 shrink-0 rounded-full bg-zinc-600" />
                  AI Oracle consultation chat
                </li>
                <li class="flex items-center gap-1.5">
                  <span class="h-1 w-1 shrink-0 rounded-full bg-zinc-600" />
                  Automatic feedback classification (HOC / LOC)
                </li>
                <li class="flex items-center gap-1.5">
                  <span class="h-1 w-1 shrink-0 rounded-full bg-zinc-600" />
                  Quick AI revision suggestions
                </li>
              </ul>
            </div>
          </div>
        </div>

        <div v-else class="flex items-center gap-3 rounded-xl border border-emerald-500/20 bg-emerald-500/[0.04] px-4 py-3">
          <svg class="h-4 w-4 shrink-0 text-emerald-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14" />
            <polyline points="22 4 12 14.01 9 11.01" />
          </svg>
          <p class="text-[13px] text-emerald-400">All AI features are enabled — both audio transcription and LLM-powered tools are ready.</p>
        </div>
      </div>
    </UiPage>
  </RequireAuth>
</template>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
