<script setup lang="ts">
import { router } from '@inertiajs/vue3';
import { useAuthStore } from '@/stores/auth';
import GalaxyHeroSection from '@/components/GalaxyHeroSection.vue';

const auth = useAuthStore();

if (auth.isAuthenticated) {
  router.visit('/dashboard');
}
</script>

<template>
  <main class="bg-[#0a0a0b] relative w-screen">
    <GalaxyHeroSection />
  </main>
</template>
