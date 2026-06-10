<script setup lang="ts">
import { computed } from 'vue';

const props = defineProps<{
  text: string;
  color?: string;
}>();

const isHex = computed(() => props.color?.startsWith('#') ?? false);

const presetClass = computed(() => {
  if (isHex || !props.color) {
    const lower = props.text.toLowerCase();
    if (lower.includes('fixed') || lower.includes('validated') || lower.includes('approved'))
      return 'bg-emerald-500/10 text-emerald-400';
    if (lower.includes('pending') || lower.includes('revision') || lower.includes('not set'))
      return 'bg-amber-500/10 text-amber-400';
    if (lower.includes('major'))
      return 'bg-red-500/10 text-red-400';
    if (lower.includes('minor'))
      return 'bg-stone-400/10 text-stone-400';
    return 'bg-stone-400/10 text-stone-400';
  }
  return props.color;
});

const inlineStyle = computed(() => {
  if (!isHex.value || !props.color) return {};
  return {
    backgroundColor: props.color + '15',
    color: props.color,
  };
});
</script>

<template>
  <span
    :class="[
      'inline-flex items-center rounded-full px-3 py-1 text-[11px] font-medium tracking-[0.5px] uppercase',
      !isHex ? presetClass : '',
    ]"
    :style="isHex ? inlineStyle : {}"
  >
    {{ text }}
  </span>
</template>
