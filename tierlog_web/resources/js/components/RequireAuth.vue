<script setup lang="ts">
import { watch } from 'vue';
import { router } from '@inertiajs/vue3';
import { useAuthStore } from '@/stores/auth';

const auth = useAuthStore();

watch(
  () => auth.booting,
  (isBooting) => {
    if (!isBooting && !auth.isAuthenticated) {
      router.visit('/login');
    }
  },
  { immediate: true }
);
</script>

<template>
  <div
    v-if="auth.booting"
    class="flex min-h-screen items-center justify-center bg-[#0a0a0b]"
  >
    <div class="flex flex-col items-center gap-4">
      <div class="h-6 w-6 animate-spin rounded-full border-2 border-amber-500 border-t-transparent" />
      <p class="text-xs text-stone-500">Loading...</p>
    </div>
  </div>
  <slot v-else-if="auth.isAuthenticated" />
</template>
