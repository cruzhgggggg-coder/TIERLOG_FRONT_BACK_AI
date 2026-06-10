<script setup lang="ts">
withDefaults(
  defineProps<{
    padding?: boolean;
    elevated?: boolean;
  }>(),
  {
    padding: true,
    elevated: false,
  },
);
</script>

<template>
  <div
    :class="[
      'rounded-2xl border border-white/[0.04] transition-all duration-200',
      elevated ? 'bg-[#1c1c1e] elevation-2' : 'bg-[#141415]',
      'state-layer',
    ]"
  >
    <div v-if="padding" class="p-6">
      <slot />
    </div>
    <slot v-else />
  </div>
</template>
