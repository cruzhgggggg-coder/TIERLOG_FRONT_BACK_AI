<script setup lang="ts">
import { computed } from 'vue';

const props = withDefaults(
  defineProps<{
    title: string;
    tone?: 'primary' | 'secondary' | 'danger' | 'success' | 'warning';
    disabled?: boolean;
    loading?: boolean;
  }>(),
  {
    tone: 'primary',
    disabled: false,
    loading: false,
  },
);

const emit = defineEmits<{
  click: [e: MouseEvent];
}>();

const isDisabled = computed(() => props.disabled || props.loading);

const toneClass = computed(() => {
  const map: Record<string, string> = {
    primary:
      'bg-gradient-to-r from-amber-400 to-amber-500 text-black font-semibold rounded-xl px-6 py-2.5 elevation-1 state-layer motion-standard hover:elevation-2',
    secondary:
      'bg-[#1c1c1e] text-stone-300 border border-white/[0.08] rounded-xl state-layer motion-standard',
    danger:
      'bg-red-500/10 text-red-400 border border-red-500/20 rounded-xl state-layer',
    success:
      'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 rounded-xl state-layer',
    warning:
      'bg-amber-500/10 text-amber-400 border border-amber-500/20 rounded-xl state-layer',
  };
  return map[props.tone] ?? map.primary;
});
</script>

<template>
  <button
    :class="[
      'inline-flex items-center justify-center gap-2 text-[13px] font-semibold tracking-[0.1px] transition-all duration-200',
      toneClass,
      isDisabled
        ? 'opacity-40 cursor-not-allowed pointer-events-none'
        : 'cursor-pointer',
    ]"
    :disabled="isDisabled"
    @click="emit('click', $event)"
  >
    <svg
      v-if="loading"
      class="h-3.5 w-3.5 animate-spin"
      viewBox="0 0 24 24"
      fill="none"
    >
      <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
      <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
    </svg>
    {{ title }}
  </button>
</template>
