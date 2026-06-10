<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue';

const screenshotRef = ref<HTMLDivElement | null>(null);

function handleScroll() {
  requestAnimationFrame(() => {
    if (screenshotRef.value) {
      const scrollY = window.pageYOffset;
      screenshotRef.value.style.transform = `translateY(-${scrollY * 0.5}px)`;
    }
  });
}

onMounted(() => {
  window.addEventListener('scroll', handleScroll, { passive: true });
});

onUnmounted(() => {
  window.removeEventListener('scroll', handleScroll);
});

defineExpose({ screenshotRef });
</script>

<template>
  <section class="relative z-10 container mx-auto px-4 md:px-6 lg:px-8 mt-11 md:mt-12">
    <div
      ref="screenshotRef"
      class="bg-[#141415] rounded-xl overflow-hidden elevation-4 border border-white/[0.04] w-full md:w-[80%] lg:w-[70%] mx-auto"
    >
      <img
        src="https://images.unsplash.com/photo-1618005182384-a83a8bd57fbe?w=1200&q=80"
        alt="App Screenshot"
        class="w-full h-auto block rounded-lg mx-auto"
        loading="lazy"
      />
    </div>
  </section>
</template>
