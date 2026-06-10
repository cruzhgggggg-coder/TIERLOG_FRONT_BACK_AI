<script setup lang="ts">
import { ref, onMounted } from 'vue';

const isMounted = ref(false);
const sceneLoaded = ref(false);
const sceneError = ref(false);
const SCENE_URL = 'https://prod.spline.design/us3ALejTXl6usHZ7/scene.splinecode';

onMounted(async () => {
  isMounted.value = true;
  try {
    await import('@splinetool/viewer');
    // Small delay to ensure the custom element is registered
    await new Promise((r) => setTimeout(r, 100));
    sceneLoaded.value = true;
  } catch {
    sceneError.value = true;
  }
});
</script>

<template>
  <div class="relative w-full h-screen overflow-hidden pointer-events-auto">
    <!-- Spline 3D Scene (client-only via onMounted gate) -->
    <div v-if="isMounted && sceneLoaded && !sceneError" class="absolute inset-0 w-full h-screen pointer-events-auto">
      <spline-viewer
        :url="SCENE_URL"
        class="w-full h-full"
        loading-anim-type="spinner-big-dark"
        events-target="global"
      />
    </div>

    <!-- Loading fallback (shown while Spline loads) -->
    <div v-if="isMounted && !sceneLoaded && !sceneError" class="absolute inset-0 w-full h-screen">
      <div class="absolute inset-0 bg-gradient-to-br from-[#0a0a0b] via-amber-950/10 to-[#0a0a0b]">
        <div class="absolute inset-0 animate-pulse bg-gradient-to-t from-[#0a0a0b] via-transparent to-transparent opacity-60" />
      </div>
    </div>

    <!-- Error fallback -->
    <div v-if="isMounted && sceneError" class="absolute inset-0 w-full h-screen">
      <div class="absolute inset-0 bg-gradient-to-br from-[#0a0a0b] via-amber-950/10 to-[#0a0a0b]" />
      <div class="absolute inset-0 mesh-gradient opacity-50" />
    </div>

    <!-- Gradient overlays (match TierLog dark palette) -->
    <div
      class="absolute inset-0 pointer-events-none"
      :style="{
        background: `linear-gradient(to right, rgba(10,10,11,0.5), transparent 25%, transparent 75%, rgba(10,10,11,0.5)), linear-gradient(to bottom, transparent 60%, rgba(10,10,11,0.6))`
      }"
    />
  </div>
</template>
