<script setup lang="ts">
import { Link } from '@inertiajs/vue3';

withDefaults(defineProps<{
  title?: string;
  subtitle?: string;
  primaryCta?: string;
  secondaryCta?: string;
}>(), {
  title: 'Your thesis, on track — finally.',
  subtitle: 'Stop drowning in scattered emails and messy revision notes. TierLog turns advisor feedback into an AI-powered checklist so you finish faster, with fewer back-and-forth cycles.',
  primaryCta: 'Get Started Free',
  secondaryCta: 'Watch Demo',
});
</script>

<template>
  <div class="text-left text-stone-100 pt-16 sm:pt-24 md:pt-32 px-4 max-w-3xl">
    <h1
      class="text-3xl sm:text-5xl md:text-7xl font-bold mb-4 leading-tight tracking-tight"
      style="font-family: var(--font-display)"
    >
      {{ title }}
    </h1>
    <p class="text-base sm:text-lg md:text-xl mb-6 sm:mb-8 text-stone-400 max-w-xl">
      {{ subtitle }}
    </p>
    <div class="flex pointer-events-auto flex-col sm:flex-row items-start space-y-3 sm:space-y-0 sm:space-x-3">
      <Link
        href="/register"
        class="state-layer elevation-1 group relative rounded-xl bg-gradient-to-r from-amber-400 to-amber-500 px-7 py-3.5 text-sm font-semibold text-black transition-all duration-300 hover:from-amber-500 hover:to-amber-400 hover:shadow-[0_0_32px_rgba(251,191,36,0.25)] focus-ring w-full sm:w-auto text-center"
      >
        {{ primaryCta }}
        <span class="ml-1 inline-block transition-transform duration-200 group-hover:translate-x-0.5">&rarr;</span>
      </Link>
      <Link
        href="#features"
        class="state-layer elevation-1 rounded-xl border border-white/[0.06] bg-[#1c1c1e] px-7 py-3.5 text-sm font-medium text-zinc-300 transition-all duration-200 hover:border-white/10 hover:bg-[#242426] hover:text-white focus-ring w-full sm:w-auto text-center"
      >
        <svg class="w-4 h-4 sm:w-5 sm:h-5 mr-2 inline-block" fill="currentColor" viewBox="0 0 20 20" xmlns="http://www.w3.org/2000/svg">
          <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM9.555 7.168A1 1 0 008 8v4a1 1 0 001.555.832l3-2a1 1 0 000-1.664l-3-2z" clip-rule="evenodd" />
        </svg>
        {{ secondaryCta }}
      </Link>
    </div>
  </div>
</template>
