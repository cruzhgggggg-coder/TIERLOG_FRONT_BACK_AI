<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { Link, router } from '@inertiajs/vue3';
import { useAuthStore } from '@/stores/auth';

const auth = useAuthStore();

if (auth.isAuthenticated) {
  router.visit('/dashboard');
}

const ready = ref(false);

onMounted(() => {
  setTimeout(() => { ready.value = true; }, 50);
});
</script>

<template>
  <div class="relative min-h-screen bg-[#0a0a0b] overflow-hidden">
    <div class="noise-overlay absolute inset-0 pointer-events-none" />
    <div class="mesh-gradient absolute inset-0 pointer-events-none" />

    <nav class="relative z-10 flex items-center justify-between px-6 py-5 md:px-12 max-w-7xl mx-auto w-full">
      <span class="text-lg font-black tracking-tight text-[#fafaf9]" style="font-family: var(--font-display)">TierLog</span>
      <Link
        href="/login"
        class="state-layer relative rounded-lg px-4 py-2 text-sm font-medium text-[#a8a29e] transition-colors duration-200 hover:text-[#fafaf9] focus-ring"
      >
        Sign In
      </Link>
    </nav>

    <main class="relative z-10">
      <section class="flex min-h-[calc(100vh-140px)] flex-col items-center justify-center px-6">
        <div class="flex max-w-3xl flex-col items-center text-center">
          <div
            :class="[
              'delay-1 flex items-center gap-2 rounded-full border border-[#fbbf24]/20 bg-[#fbbf24]/6 px-4 py-1.5 transition-all duration-500',
              ready ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-4',
            ]"
            style="transition-timing-function: cubic-bezier(0.2, 0, 0, 1)"
          >
            <span class="h-1.5 w-1.5 rounded-full bg-[#fbbf24] animate-pulse" />
            <span class="text-xs font-medium tracking-wide text-[#fbbf24]">Trusted by 2,000+ researchers</span>
          </div>

          <h1
            :class="[
              'delay-2 mt-8 text-5xl font-semibold tracking-tight text-[#fafaf9] md:text-7xl leading-[1.05] transition-all duration-500',
              ready ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-6',
            ]"
            style="font-family: var(--font-display); transition-timing-function: cubic-bezier(0.2, 0, 0, 1)"
          >
            Stop losing feedback<br />
            <span class="gradient-text">Start finishing your thesis</span>
          </h1>

          <p
            :class="[
              'delay-3 mt-7 max-w-lg text-body-large leading-relaxed text-[#a8a29e] transition-all duration-500',
              ready ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-4',
            ]"
            style="transition-timing-function: cubic-bezier(0.2, 0, 0, 1)"
          >
            TierLog transforms scattered advisor comments into a clear, trackable revision workflow.
          </p>

          <div
            :class="[
              'delay-4 mt-10 flex items-center gap-4 transition-all duration-500',
              ready ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-4',
            ]"
            style="transition-timing-function: cubic-bezier(0.2, 0, 0, 1)"
          >
            <Link
              href="/register"
              class="state-layer elevation-1 group relative rounded-xl bg-linear-to-r from-[#fbbf24] to-[#f59e0b] px-7 py-3.5 text-sm font-semibold text-black transition-all duration-300 hover:from-[#f59e0b] hover:to-[#fbbf24] hover:shadow-[0_0_32px_rgba(251,191,36,0.25)] focus-ring"
            >
              Start Free
              <span class="ml-1 inline-block transition-transform duration-200 group-hover:translate-x-0.5">&rarr;</span>
            </Link>
            <Link
              href="#features"
              class="state-layer elevation-1 rounded-xl border border-white/6 bg-[#1c1c1e] px-7 py-3.5 text-sm font-medium text-zinc-300 transition-all duration-200 hover:border-white/10 hover:bg-[#242426] hover:text-white focus-ring"
            >
              See How It Works
            </Link>
          </div>
        </div>

        <div
          id="features"
          :class="[
            'delay-5 mt-24 grid w-full max-w-4xl grid-cols-1 gap-4 md:grid-cols-3 transition-all duration-500',
            ready ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-4',
          ]"
          style="transition-timing-function: cubic-bezier(0.2, 0, 0, 1)"
        >
          <div class="elevation-1 state-layer group rounded-2xl bg-[#141415] p-6 transition-all duration-300 hover:bg-[#1a1a1c]">
            <div class="flex h-10 w-10 items-center justify-center rounded-xl bg-[#fbbf24]/10">
              <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" class="text-[#fbbf24]">
                <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
                <polyline points="17 8 12 3 7 8" />
                <line x1="12" y1="3" x2="12" y2="15" />
              </svg>
            </div>
            <h3 class="mt-4 title-medium text-[#fafaf9]" style="font-family: var(--font-display)">Upload &amp; Transcribe</h3>
            <p class="mt-2 body-medium text-[#78716c]">
              Drop your draft and audio recording. We transcribe, extract, and organize everything automatically.
            </p>
          </div>

          <div class="elevation-1 state-layer group rounded-2xl bg-[#141415] p-6 transition-all duration-300 hover:bg-[#1a1a1c]">
            <div class="flex h-10 w-10 items-center justify-center rounded-xl bg-[#fbbf24]/10">
              <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" class="text-[#fbbf24]">
                <path d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" />
              </svg>
            </div>
            <h3 class="mt-4 title-medium text-[#fafaf9]" style="font-family: var(--font-display)">AI-Powered Analysis</h3>
            <p class="mt-2 body-medium text-[#78716c]">
              Get instant feedback classification. HOC for structural issues, LOC for granular fixes. Never miss a critical revision.
            </p>
          </div>

          <div class="elevation-1 state-layer group rounded-2xl bg-[#141415] p-6 transition-all duration-300 hover:bg-[#1a1a1c]">
            <div class="flex h-10 w-10 items-center justify-center rounded-xl bg-[#fbbf24]/10">
              <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" class="text-[#fbbf24]">
                <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14" />
                <polyline points="22 4 12 14.01 9 11.01" />
              </svg>
            </div>
            <h3 class="mt-4 title-medium text-[#fafaf9]" style="font-family: var(--font-display)">Track &amp; Validate</h3>
            <p class="mt-2 body-medium text-[#78716c]">
              Mark revisions as fixed. Your advisor validates. Every change tracked with full audit history.
            </p>
          </div>
        </div>
      </section>

      <section class="relative z-10 border-t border-white/4 bg-[#0a0a0b]/80 backdrop-blur-sm">
        <div class="mx-auto max-w-5xl px-6 py-20">
          <p
            :class="[
              'delay-6 text-center label-medium uppercase tracking-[0.2em] text-[#78716c] transition-all duration-500',
              ready ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-4',
            ]"
            style="transition-timing-function: cubic-bezier(0.2, 0, 0, 1)"
          >
            Trusted by researchers across 12 universities
          </p>

          <div
            :class="[
              'delay-7 mt-12 grid grid-cols-2 gap-8 md:grid-cols-4 transition-all duration-500',
              ready ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-4',
            ]"
            style="transition-timing-function: cubic-bezier(0.2, 0, 0, 1)"
          >
            <div class="text-center">
              <p class="headline-medium text-[#fbbf24]" style="font-family: var(--font-display)">2,000+</p>
              <p class="mt-1.5 label-medium text-[#78716c]">Drafts Reviewed</p>
            </div>
            <div class="text-center">
              <p class="headline-medium text-[#fbbf24]" style="font-family: var(--font-display)">12</p>
              <p class="mt-1.5 label-medium text-[#78716c]">Universities</p>
            </div>
            <div class="text-center">
              <p class="headline-medium text-[#fbbf24]" style="font-family: var(--font-display)">94%</p>
              <p class="mt-1.5 label-medium text-[#78716c]">Completion Rate</p>
            </div>
            <div class="text-center">
              <p class="headline-medium text-[#fbbf24]" style="font-family: var(--font-display)">&lt; 2min</p>
              <p class="mt-1.5 label-medium text-[#78716c]">Avg Response</p>
            </div>
          </div>
        </div>
      </section>

      <footer class="relative z-10 border-t border-white/4 py-8">
        <div class="mx-auto max-w-5xl px-6 text-center">
          <p class="body-small text-[#52525b]">
            Built with care for academic excellence. &copy; 2026 TierLog
          </p>
        </div>
      </footer>
    </main>
  </div>
</template>

<style scoped>
.delay-1 { transition-delay: 0ms; }
.delay-2 { transition-delay: 100ms; }
.delay-3 { transition-delay: 200ms; }
.delay-4 { transition-delay: 300ms; }
.delay-5 { transition-delay: 400ms; }
.delay-6 { transition-delay: 500ms; }
.delay-7 { transition-delay: 600ms; }

.elevation-1 {
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.3), 0 1px 3px 1px rgba(0, 0, 0, 0.15);
}

.elevation-2 {
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.3), 0 2px 6px 2px rgba(0, 0, 0, 0.15);
}

.state-layer {
  position: relative;
}

.state-layer::before {
  content: '';
  position: absolute;
  inset: 0;
  border-radius: inherit;
  background: currentColor;
  opacity: 0;
  transition: opacity 200ms cubic-bezier(0.2, 0, 0, 1);
  pointer-events: none;
}

.state-layer:hover::before {
  opacity: 0.08;
}

.state-layer:active::before {
  opacity: 0.12;
}

.focus-ring:focus-visible {
  outline: 2px solid #fbbf24;
  outline-offset: 2px;
}

.headline-medium {
  font-size: 2.25rem;
  line-height: 2.5rem;
  font-weight: 600;
  letter-spacing: -0.025em;
}

.display-medium {
  font-size: 3rem;
  line-height: 3.5rem;
  font-weight: 600;
  letter-spacing: -0.025em;
}

.title-medium {
  font-size: 0.875rem;
  line-height: 1.25rem;
  font-weight: 600;
}

.body-large {
  font-size: 1.125rem;
  line-height: 1.75rem;
}

.body-medium {
  font-size: 0.875rem;
  line-height: 1.25rem;
}

.body-small {
  font-size: 0.75rem;
  line-height: 1rem;
}

.label-medium {
  font-size: 0.75rem;
  line-height: 1rem;
  font-weight: 500;
}

.gradient-text {
  background: linear-gradient(135deg, #fbbf24, #f59e0b);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}
</style>
