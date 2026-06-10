<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { useAuthStore } from '@/stores/auth';
import type { ConsultationLog } from '@/types';
import RequireAuth from '@/components/RequireAuth.vue';
import NavBar from '@/components/NavBar.vue';
import UiPage from '@/components/UiPage.vue';
import UiHeading from '@/components/UiHeading.vue';
import UiCard from '@/components/UiCard.vue';
import UiBadge from '@/components/UiBadge.vue';
import { CheckCircleIcon, ClockIcon, ChevronRightIcon } from '@/components/icons';

const auth = useAuthStore();
const logs = ref<ConsultationLog[]>([]);
const loading = ref(true);
const activeFilter = ref<'all' | 'approved' | 'revision'>('all');
const expandedLogId = ref<number | null>(null);
const activeTab = ref<'feedback' | 'transcript' | 'annotations'>('feedback');

const filteredLogs = computed(() => {
  if (activeFilter.value === 'approved') {
    return logs.value.filter((log) => {
      const pending = log.feedback_items?.filter((f) => f.status === 'Pending').length ?? 0;
      return pending === 0;
    });
  }
  if (activeFilter.value === 'revision') {
    return logs.value.filter((log) => {
      const pending = log.feedback_items?.filter((f) => f.status === 'Pending').length ?? 0;
      return pending > 0;
    });
  }
  return logs.value;
});

const filterCounts = computed(() => {
  const all = logs.value.length;
  const approved = logs.value.filter((log) => {
    const pending = log.feedback_items?.filter((f) => f.status === 'Pending').length ?? 0;
    return pending === 0;
  }).length;
  return { all, approved, revision: all - approved };
});

function timelineStatus(log: ConsultationLog): string {
  const pending = log.feedback_items?.filter((f) => f.status === 'Pending').length ?? 0;
  return pending > 0 ? 'Revision Required' : 'Approved';
}

function toggleExpand(logId: number) {
  if (expandedLogId.value === logId) {
    expandedLogId.value = null;
  } else {
    expandedLogId.value = logId;
    activeTab.value = 'feedback';
  }
}

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString('id-ID', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  });
}

async function fetchLogs() {
  loading.value = true;
  const res = await auth.api<{ data: ConsultationLog[] }>('/logs');
  if (res.ok && res.data) {
    logs.value = res.data.data ?? [];
  }
  loading.value = false;
}

onMounted(() => {
  fetchLogs();
});
</script>

<template>
  <RequireAuth>
    <NavBar />
    <UiPage>
      <div class="space-y-8">
        <UiHeading class="font-[family-name:var(--font-display)]" title="Session Archive" subtitle="Browse and review past consultation sessions." />

        <div v-if="loading" class="flex items-center justify-center py-20">
          <div class="h-6 w-6 animate-spin rounded-full border-2 border-amber-500 border-t-transparent" />
        </div>

        <template v-else>
          <div class="flex gap-1.5 rounded-lg bg-[#141415] p-1">
            <button
              :class="[
                'rounded-md px-3.5 py-1.5 text-[13px] font-medium transition-colors duration-100',
                activeFilter === 'all'
                  ? 'bg-[#1c1c1e] text-stone-100'
                  : 'text-stone-600 hover:text-stone-300',
              ]"
              @click="activeFilter = 'all'"
            >
              All
              <span class="ml-1 text-[11px] text-stone-700">{{ filterCounts.all }}</span>
            </button>
            <button
              :class="[
                'rounded-md px-3.5 py-1.5 text-[13px] font-medium transition-colors duration-100',
                activeFilter === 'approved'
                  ? 'bg-[#1c1c1e] text-stone-100'
                  : 'text-stone-600 hover:text-stone-300',
              ]"
              @click="activeFilter = 'approved'"
            >
              Approved
              <span class="ml-1 text-[11px] text-stone-700">{{ filterCounts.approved }}</span>
            </button>
            <button
              :class="[
                'rounded-md px-3.5 py-1.5 text-[13px] font-medium transition-colors duration-100',
                activeFilter === 'revision'
                  ? 'bg-[#1c1c1e] text-stone-100'
                  : 'text-stone-600 hover:text-stone-300',
              ]"
              @click="activeFilter = 'revision'"
            >
              Revision Required
              <span class="ml-1 text-[11px] text-stone-700">{{ filterCounts.revision }}</span>
            </button>
          </div>

          <div v-if="filteredLogs.length === 0" class="py-16 text-center">
            <p class="text-sm text-stone-700">No sessions found for the selected filter.</p>
          </div>

          <div v-else class="noise-overlay relative">
            <div class="absolute left-[15px] top-0 bottom-0 w-px bg-white/[0.04]" />

            <div v-for="log in filteredLogs" :key="log.id" class="relative pl-10">
              <div
                :class="[
                  'absolute left-2.5 top-5 z-10 h-2.5 w-2.5 rounded-full border-2',
                  timelineStatus(log) === 'Approved'
                    ? 'border-emerald-500 bg-emerald-500/30'
                    : 'border-amber-500 bg-amber-500/30',
                ]"
              />

              <div class="py-2.5">
                <UiCard>
                  <div
                    class="cursor-pointer p-4 transition-colors duration-100 hover:bg-white/[0.02]"
                    @click="toggleExpand(log.id)"
                  >
                    <div class="flex items-start justify-between gap-3">
                      <div class="min-w-0 flex-1">
                        <div class="flex flex-wrap items-center gap-2">
                          <p class="text-sm font-medium text-stone-100">{{ log.paper_filename }}</p>
                          <UiBadge :text="timelineStatus(log)" />
                        </div>
                        <p class="mt-1 text-[11px] text-stone-700">{{ formatDate(log.created_at) }}</p>
                      </div>
                      <div class="flex items-center gap-2 shrink-0">
                        <span class="text-[11px] text-stone-700">{{ log.feedback_items?.length ?? 0 }} items</span>
                        <ChevronRightIcon
                          :size="14"
                          :class="[
                            'text-stone-700 transition-transform duration-150',
                            expandedLogId === log.id ? 'rotate-90' : '',
                          ]"
                        />
                      </div>
                    </div>
                  </div>

                  <div v-if="expandedLogId === log.id" class="border-t border-white/[0.04]">
                    <div class="flex gap-0.5 border-b border-white/[0.04] px-4 pt-3">
                      <button
                        :class="[
                          'rounded-md px-3 py-1.5 text-[13px] font-medium transition-colors duration-100',
                          activeTab === 'feedback'
                            ? 'bg-[#1c1c1e] text-stone-100'
                            : 'text-stone-600 hover:text-stone-300',
                        ]"
                        @click="activeTab = 'feedback'"
                      >
                        Revision Items
                      </button>
                      <button
                        :class="[
                          'rounded-md px-3 py-1.5 text-[13px] font-medium transition-colors duration-100',
                          activeTab === 'transcript'
                            ? 'bg-[#1c1c1e] text-stone-100'
                            : 'text-stone-600 hover:text-stone-300',
                        ]"
                        @click="activeTab = 'transcript'"
                      >
                        Transcript
                      </button>
                      <button
                        :class="[
                          'rounded-md px-3 py-1.5 text-[13px] font-medium transition-colors duration-100',
                          activeTab === 'annotations'
                            ? 'bg-[#1c1c1e] text-stone-100'
                            : 'text-stone-600 hover:text-stone-300',
                        ]"
                        @click="activeTab = 'annotations'"
                      >
                        Annotations
                      </button>
                    </div>

                    <div class="p-4">
                      <div v-if="activeTab === 'feedback'" class="space-y-2">
                        <div
                          v-for="item in log.feedback_items"
                          :key="item.id"
                          class="flex items-start gap-3 rounded-lg border border-white/[0.04] bg-[#1c1c1e]/50 p-3.5"
                        >
                          <div class="mt-0.5 shrink-0">
                            <component
                              :is="item.status === 'Fixed' || item.status === 'Validated' ? CheckCircleIcon : item.status === 'Pending' ? ClockIcon : ClockIcon"
                              :size="14"
                              :color="item.status === 'Fixed' || item.status === 'Validated' ? '#22c55e' : item.status === 'Pending' ? '#f59e0b' : '#ef4444'"
                            />
                          </div>
                          <div class="min-w-0 flex-1">
                            <p class="text-sm text-stone-300">{{ item.content }}</p>
                          </div>
                          <div class="flex shrink-0 flex-col gap-1">
                            <UiBadge :text="item.category" :color="item.category === 'Major' ? '#ef4444' : '#71717a'" />
                            <UiBadge :text="item.status" />
                          </div>
                        </div>
                        <div v-if="!log.feedback_items?.length" class="py-8 text-center">
                          <p class="text-sm text-stone-700">No feedback items recorded.</p>
                        </div>
                      </div>

                      <div v-if="activeTab === 'transcript'" class="rounded-lg border border-white/[0.04] bg-[#1c1c1e]/50 p-4">
                        <div class="max-h-72 scroll-area">
                          <p v-if="log.transcript_text" class="whitespace-pre-wrap text-sm leading-relaxed text-stone-500">
                            {{ log.transcript_text }}
                          </p>
                          <p v-else class="text-sm text-stone-700">No transcript available for this session.</p>
                        </div>
                      </div>

                      <div v-if="activeTab === 'annotations'" class="space-y-2">
                        <div
                          v-for="annotation in log.revision_annotations"
                          :key="annotation.id"
                          class="rounded-lg border border-white/[0.04] bg-[#1c1c1e]/50 p-4"
                        >
                          <div class="mb-3 flex items-center justify-between">
                            <div class="flex items-center gap-2.5">
                              <div class="flex h-7 w-7 items-center justify-center rounded-md bg-zinc-800">
                                <span class="text-[10px] font-bold uppercase text-stone-500">
                                  {{ annotation.file_type === 'image' ? 'IMG' : 'DOC' }}
                                </span>
                              </div>
                              <div>
                                <p class="text-sm font-medium text-stone-200">{{ annotation.filename }}</p>
                                <p class="text-[11px] text-stone-700">{{ annotation.file_type.toUpperCase() }}</p>
                              </div>
                            </div>
                            <p class="text-[11px] text-stone-700">{{ formatDate(annotation.created_at) }}</p>
                          </div>
                          <div v-if="annotation.extracted_text" class="rounded-md bg-[#0a0a0b] p-3">
                            <p class="mb-1 text-[10px] font-medium uppercase tracking-wider text-stone-700">OCR Text</p>
                            <p class="whitespace-pre-wrap text-sm text-stone-600">{{ annotation.extracted_text }}</p>
                          </div>
                        </div>
                        <div v-if="!log.revision_annotations?.length" class="py-8 text-center">
                          <p class="text-sm text-stone-700">No correction notes available.</p>
                        </div>
                      </div>
                    </div>
                  </div>
                </UiCard>
              </div>
            </div>
          </div>
        </template>
      </div>
    </UiPage>
  </RequireAuth>
</template>
