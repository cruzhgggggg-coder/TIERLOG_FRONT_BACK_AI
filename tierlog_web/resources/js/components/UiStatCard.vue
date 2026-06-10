<script setup lang="ts">
withDefaults(
  defineProps<{
    label: string;
    value: string | number;
    color?: string;
  }>(),
  {
    color: '#fbbf24',
  },
);
</script>

<template>
  <div class="relative overflow-hidden rounded-2xl bg-[#141415] p-5 elevation-1">
    <div class="absolute inset-x-0 top-0 h-px bg-gradient-to-r from-transparent via-amber-500/30 to-transparent" />
    <p class="text-xs font-medium uppercase tracking-[0.5px] text-stone-600">
      {{ label }}
    </p>
    <p
      class="mt-2 text-2xl font-bold tracking-tight"
      :style="{ color }"
    >
      {{ value }}
    </p>
  </div>
</template>
