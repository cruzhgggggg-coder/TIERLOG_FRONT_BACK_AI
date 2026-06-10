<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { Link, router } from '@inertiajs/vue3';
import { useAuthStore, translateError } from '@/stores/auth';
import UiField from '@/components/UiField.vue';
import UiButton from '@/components/UiButton.vue';

const auth = useAuthStore();

if (auth.isAuthenticated) {
  router.visit('/dashboard');
}

const email = ref('');
const password = ref('');
const error = ref('');
const loading = ref(false);

const visible = ref(false);
onMounted(() => {
  setTimeout(() => { visible.value = true; }, 50);
});

async function handleLogin() {
  error.value = '';
  loading.value = true;
  const result = await auth.login(email.value, password.value);
  loading.value = false;
  if (result.ok) {
    router.visit('/dashboard');
  } else {
    error.value = translateError(result.error);
  }
}
</script>

<template>
  <div class="flex min-h-screen bg-[#0a0a0b]">
    <div
      class="relative hidden w-[60%] items-center justify-center overflow-hidden lg:flex noise-overlay mesh-gradient"
    >
      <div
        class="pointer-events-none absolute left-[10%] top-[20%] h-[500px] w-[500px] rounded-full opacity-[0.08]"
        style="background: radial-gradient(circle, #fbbf24, transparent 70%); animation: float 8s ease-in-out infinite;"
      />
      <div
        class="pointer-events-none absolute bottom-[15%] right-[15%] h-[350px] w-[350px] rounded-full opacity-[0.05]"
        style="background: radial-gradient(circle, #f59e0b, transparent 70%); animation: float 10s ease-in-out infinite 2s;"
      />

      <div class="relative z-10 flex flex-col items-start px-20">
        <span
          :class="[
            'display-medium font-black tracking-tight transition-all duration-500',
            visible ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-6',
          ]"
          style="font-family: var(--font-display); transition-timing-function: cubic-bezier(0.2, 0, 0, 1)"
        >
          TierLog
        </span>
        <p
          :class="[
            'mt-4 headline-small font-light leading-relaxed transition-all duration-500',
            visible ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-4',
          ]"
          style="transition-delay: 150ms; transition-timing-function: cubic-bezier(0.2, 0, 0, 1)"
        >
          <span class="gradient-text">Where thesis guidance becomes graduation</span>
        </p>

        <div
          :class="[
            'mt-10 flex flex-wrap gap-3 transition-all duration-500',
            visible ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-4',
          ]"
          style="transition-delay: 300ms; transition-timing-function: cubic-bezier(0.2, 0, 0, 1)"
        >
          <span class="rounded-full border border-white/[0.06] bg-[#1c1c1e] px-4 py-2 label-medium text-[#a8a29e]">
            AI-Powered
          </span>
          <span class="rounded-full border border-white/[0.06] bg-[#1c1c1e] px-4 py-2 label-medium text-[#a8a29e]">
            Real-Time
          </span>
          <span class="rounded-full border border-white/[0.06] bg-[#1c1c1e] px-4 py-2 label-medium text-[#a8a29e]">
            Collaboration
          </span>
        </div>
      </div>
    </div>

    <div
      :class="[
        'flex w-full flex-1 items-center justify-center px-6 py-12 transition-all duration-500 lg:w-[40%]',
        visible ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-4',
      ]"
      style="background: #141415; transition-timing-function: cubic-bezier(0.2, 0, 0, 1)"
    >
      <div class="w-full max-w-sm space-y-8">
        <div>
          <Link
            href="/"
            class="state-layer inline-flex items-center gap-1.5 rounded-lg px-3 py-1.5 text-sm text-[#a8a29e] transition-colors duration-200 hover:text-[#fafaf9] focus-ring"
          >
            <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M19 12H5" /><path d="M12 19l-7-7 7-7" />
            </svg>
            Back to home
          </Link>
        </div>

        <div class="space-y-2">
          <h1 class="headline-small text-[#fafaf9]" style="font-family: var(--font-display)">
            Welcome back
          </h1>
          <p class="body-medium text-[#a8a29e]">Sign in to continue</p>
        </div>

        <form class="space-y-5" @submit.prevent="handleLogin">
          <UiField
            v-model="email"
            label="Email"
            type="email"
            placeholder="you@university.ac.id"
          />
          <UiField
            v-model="password"
            label="Password"
            type="password"
            placeholder="Enter your password"
          />

          <div
            v-if="error"
            class="rounded-xl border border-red-500/20 bg-red-500/5 p-3 text-[13px] text-red-400"
          >
            {{ error }}
          </div>

          <UiButton
            title="Sign In"
            :loading="loading"
            class="w-full"
          />
        </form>

        <p class="text-center body-medium text-[#a8a29e]">
          New here?
          <Link href="/register" class="font-medium text-[#fbbf24] transition-colors duration-200 hover:text-[#f59e0b]">
            Create account
          </Link>
        </p>
      </div>
    </div>
  </div>
</template>

<style scoped>
.elevation-1 {
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.3), 0 1px 3px 1px rgba(0, 0, 0, 0.15);
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

.display-medium {
  font-size: 3rem;
  line-height: 3.5rem;
  font-weight: 600;
  letter-spacing: -0.025em;
}

.headline-small {
  font-size: 1.5rem;
  line-height: 2rem;
  font-weight: 600;
  letter-spacing: -0.025em;
}

.body-medium {
  font-size: 0.875rem;
  line-height: 1.25rem;
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
