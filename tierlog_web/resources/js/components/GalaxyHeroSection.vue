<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue';
import { Link } from '@inertiajs/vue3';
import GalaxyNavbar from '@/components/GalaxyNavbar.vue';
import HeroSplineBackground from '@/components/HeroSplineBackground.vue';
import HeroContent from '@/components/HeroContent.vue';

const heroContentEl = ref<HTMLDivElement | null>(null);

function handleScroll() {
  requestAnimationFrame(() => {
    const scrollY = window.pageYOffset;
    const maxScroll = 400;
    const opacity = 1 - Math.min(scrollY / maxScroll, 1);
    if (heroContentEl.value) {
      heroContentEl.value.style.opacity = opacity.toString();
    }
  });
}

onMounted(() => window.addEventListener('scroll', handleScroll, { passive: true }));
onUnmounted(() => window.removeEventListener('scroll', handleScroll));

const features = [
  {
    icon: `<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" /></svg>`,
    title: 'Upload & Transcribe',
    description: 'Drop your thesis draft in any format. TierLog extracts every comment, suggestion, and correction — even handwritten margin notes.',
  },
  {
    icon: `<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" /></svg>`,
    title: 'AI Analysis',
    description: 'AI clusters feedback by topic, detects conflicts between advisors, and prioritizes what matters most for your defense.',
  },
  {
    icon: `<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>`,
    title: 'Track & Validate',
    description: 'Mark revisions done, tag which advisor approved them, and generate a validation report — so nobody asks "did you fix that?" ever again.',
  },
  {
    icon: `<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z" /></svg>`,
    title: 'Advisor Dashboard',
    description: 'Advisors see real-time progress across all their students. No more chasing emails or guessing who submitted what.',
  },
  {
    icon: `<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253" /></svg>`,
    title: 'Version History',
    description: 'Every revision is tracked with full diff comparison. See exactly what changed between drafts and who requested each edit.',
  },
  {
    icon: `<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M13 10V3L4 14h7v7l9-11h-7z" /></svg>`,
    title: 'Consultation Scheduler',
    description: 'Book and manage advisor meetings directly inside TierLog. Attach context to each session so no time is wasted on catch-up.',
  },
];

const stats = [
  { value: '12x', label: 'Faster revision cycles' },
  { value: '89%', label: 'Fewer missed comments' },
  { value: '3,200+', label: 'Theses completed' },
  { value: '4.9', label: 'Average user rating' },
];
</script>

<template>
  <div class="relative">
    <GalaxyNavbar />

    <div class="relative min-h-screen">
      <div class="absolute inset-0 z-0 pointer-events-auto">
        <HeroSplineBackground />
      </div>

      <div
        ref="heroContentEl"
        class="absolute top-0 left-0 w-full z-10 pointer-events-none flex items-center"
        style="height: 100vh;"
      >
        <div class="container mx-auto">
          <HeroContent />
        </div>
      </div>
    </div>

    <div class="bg-[#0a0a0b] relative z-20">

      <!-- Features Section -->
      <section id="features" class="container mx-auto px-4 sm:px-6 lg:px-8 py-20 sm:py-28">
        <div class="text-center mb-16">
          <p class="text-amber-400 text-sm font-semibold tracking-widest uppercase mb-3" style="font-family: var(--font-display)">
            Built for thesis workflows
          </p>
          <h2 class="text-3xl sm:text-4xl md:text-5xl font-bold text-stone-100 mb-4" style="font-family: var(--font-display)">
            Everything you need to finish
          </h2>
          <p class="text-stone-500 max-w-2xl mx-auto text-base sm:text-lg">
            No more scattered docs, lost emails, or guesswork. TierLog is the single source of truth for your revision process.
          </p>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          <div
            v-for="(feature, i) in features"
            :key="i"
            class="group relative rounded-2xl border border-white/[0.04] bg-[#111112] p-6 sm:p-8 transition-all duration-300 hover:border-amber-500/20 hover:bg-[#151517]"
          >
            <div class="w-10 h-10 rounded-xl bg-amber-500/10 flex items-center justify-center text-amber-400 mb-4 transition-colors duration-300 group-hover:bg-amber-500/20">
              <span v-html="feature.icon" />
            </div>
            <h3 class="text-lg font-semibold text-stone-100 mb-2" style="font-family: var(--font-display)">
              {{ feature.title }}
            </h3>
            <p class="text-stone-500 text-sm leading-relaxed">
              {{ feature.description }}
            </p>
          </div>
        </div>
      </section>

      <!-- Stats Section -->
      <section class="border-y border-white/[0.04] bg-[#0d0d0e]">
        <div class="container mx-auto px-4 sm:px-6 lg:px-8 py-16 sm:py-20">
          <div class="grid grid-cols-2 md:grid-cols-4 gap-8 md:gap-12">
            <div v-for="(stat, i) in stats" :key="i" class="text-center">
              <p class="text-3xl sm:text-4xl md:text-5xl font-black text-amber-400 mb-2" style="font-family: var(--font-display)">
                {{ stat.value }}
              </p>
              <p class="text-stone-500 text-sm sm:text-base">
                {{ stat.label }}
              </p>
            </div>
          </div>
        </div>
      </section>

      <!-- CTA Section -->
      <section class="container mx-auto px-4 sm:px-6 lg:px-8 py-20 sm:py-28 text-center">
        <h2 class="text-3xl sm:text-4xl md:text-5xl font-bold text-stone-100 mb-4" style="font-family: var(--font-display)">
          Ready to finish your thesis?
        </h2>
        <p class="text-stone-500 max-w-xl mx-auto mb-8 text-base sm:text-lg">
          Join thousands of students who stopped drowning in feedback and started finishing.
        </p>
        <div class="flex flex-col sm:flex-row items-center justify-center gap-4">
          <Link
            href="/register"
            class="rounded-xl bg-gradient-to-r from-amber-400 to-amber-500 px-8 py-4 text-sm font-semibold text-black transition-all duration-300 hover:from-amber-500 hover:to-amber-400 hover:shadow-[0_0_32px_rgba(251,191,36,0.25)] w-full sm:w-auto text-center"
          >
            Get Started Free &rarr;
          </Link>
          <Link
            href="#features"
            class="rounded-xl border border-white/[0.06] bg-[#1c1c1e] px-8 py-4 text-sm font-medium text-zinc-300 transition-all duration-200 hover:border-white/10 hover:bg-[#242426] hover:text-white w-full sm:w-auto text-center"
          >
            Learn More
          </Link>
        </div>
      </section>

      <!-- Footer -->
      <footer class="border-t border-white/[0.04] bg-[#08080a]">
        <div class="container mx-auto px-4 sm:px-6 lg:px-8 py-12 sm:py-16">
          <div class="grid grid-cols-2 md:grid-cols-4 gap-8 mb-12">
            <div class="col-span-2 md:col-span-1">
              <Link href="/" class="flex items-center gap-2 text-white font-black tracking-tight mb-4" style="font-family: var(--font-display)">
                <span class="text-amber-400 text-xs">&#x25CF;</span>
                TierLog
              </Link>
              <p class="text-stone-600 text-sm leading-relaxed">
                The smart revision workflow for thesis students and advisors.
              </p>
            </div>

            <div>
              <h4 class="text-stone-300 text-sm font-semibold mb-4">Product</h4>
              <ul class="space-y-2">
                <li><Link href="#features" class="text-stone-600 hover:text-stone-300 text-sm transition-colors">Features</Link></li>
                <li><Link href="/register" class="text-stone-600 hover:text-stone-300 text-sm transition-colors">Pricing</Link></li>
                <li><Link href="#features" class="text-stone-600 hover:text-stone-300 text-sm transition-colors">Integrations</Link></li>
                <li><Link href="/settings/ai-gateway" class="text-stone-600 hover:text-stone-300 text-sm transition-colors">AI Gateway</Link></li>
              </ul>
            </div>

            <div>
              <h4 class="text-stone-300 text-sm font-semibold mb-4">Resources</h4>
              <ul class="space-y-2">
                <li><Link href="/consultations" class="text-stone-600 hover:text-stone-300 text-sm transition-colors">Consultations</Link></li>
                <li><Link href="/archive" class="text-stone-600 hover:text-stone-300 text-sm transition-colors">Archive</Link></li>
                <li><Link href="#" class="text-stone-600 hover:text-stone-300 text-sm transition-colors">Documentation</Link></li>
                <li><Link href="#" class="text-stone-600 hover:text-stone-300 text-sm transition-colors">Changelog</Link></li>
              </ul>
            </div>

            <div>
              <h4 class="text-stone-300 text-sm font-semibold mb-4">Company</h4>
              <ul class="space-y-2">
                <li><Link href="#" class="text-stone-600 hover:text-stone-300 text-sm transition-colors">About</Link></li>
                <li><Link href="#" class="text-stone-600 hover:text-stone-300 text-sm transition-colors">Blog</Link></li>
                <li><Link href="#" class="text-stone-600 hover:text-stone-300 text-sm transition-colors">Privacy</Link></li>
                <li><Link href="#" class="text-stone-600 hover:text-stone-300 text-sm transition-colors">Terms</Link></li>
              </ul>
            </div>
          </div>

          <div class="border-t border-white/[0.04] pt-8 flex flex-col sm:flex-row items-center justify-between gap-4">
            <p class="text-stone-700 text-xs">
              &copy; {{ new Date().getFullYear() }} TierLog. All rights reserved.
            </p>
            <div class="flex items-center gap-4">
              <Link href="#" class="text-stone-700 hover:text-stone-400 transition-colors" aria-label="GitHub">
                <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 24 24"><path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/></svg>
              </Link>
              <Link href="#" class="text-stone-700 hover:text-stone-400 transition-colors" aria-label="Twitter">
                <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 24 24"><path d="M18.244 2.25h3.308l-7.227 8.26 8.502 11.24H16.17l-5.214-6.817L4.99 21.75H1.68l7.73-8.835L1.254 2.25H8.08l4.713 6.231zm-1.161 17.52h1.833L7.084 4.126H5.117z"/></svg>
              </Link>
            </div>
          </div>
        </div>
      </footer>
    </div>
  </div>
</template>
