<script setup lang="ts">
defineProps<{
  title: string;
  subtitle?: string;
}>();
</script>

<template>
  <div class="space-y-1">
    <h1 class="font-[family-name:var(--font-display)] text-2xl font-semibold text-stone-100 tracking-tight">
      {{ title }}
    </h1>
    <p v-if="subtitle" class="text-[14px] text-stone-500 mt-1">
      {{ subtitle }}
    </p>
  </div>
</template>
