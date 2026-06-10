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

const API_URL = import.meta.env.VITE_API_URL || 'http://127.0.0.1:8080';

const role = ref<'student' | 'lecturer'>('student');
const name = ref('');
const email = ref('');
const password = ref('');

const nim = ref('');
const lecturerId = ref<number | ''>('');
const prodi = ref('');
const thesisTitle = ref('');

const nip = ref('');
const faculty = ref('');
const expertise = ref('');

const error = ref('');
const loading = ref(false);
const lecturers = ref<any[]>([]);

const visible = ref(false);
onMounted(async () => {
  setTimeout(() => { visible.value = true; }, 50);
  try {
    const res = await fetch(`${API_URL}/lecturers`);
    if (res.ok) {
      const data = await res.json();
      lecturers.value = data.data || [];
    }
  } catch (err) {
    console.error('Failed to fetch lecturers:', err);
  }
});

async function handleRegister() {
  error.value = '';
  loading.value = true;

  const payload: Record<string, any> = {
    name: name.value,
    email: email.value,
    password: password.value,
    role: role.value,
  };

  if (role.value === 'student') {
    payload.nim = nim.value;
    payload.lecturer_id = lecturerId.value ? Number(lecturerId.value) : 0;
    payload.prodi = prodi.value;
    payload.thesis_title = thesisTitle.value;
  } else {
    payload.nip = nip.value;
    payload.faculty = faculty.value;
    payload.keahlian = expertise.value;
  }

  const result = await auth.register(payload as { name: string; email: string; password: string; role: string });
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
        class="pointer-events-none absolute right-[10%] top-[25%] h-[450px] w-[450px] rounded-full opacity-[0.07]"
        style="background: radial-gradient(circle, #fbbf24, transparent 70%); animation: float 9s ease-in-out infinite;"
      />
      <div
        class="pointer-events-none absolute bottom-[20%] left-[8%] h-[300px] w-[300px] rounded-full opacity-[0.04]"
        style="background: radial-gradient(circle, #f59e0b, transparent 70%); animation: float 11s ease-in-out infinite 3s;"
      />

      <div class="relative z-10 flex flex-col items-start px-20">
        <span
          :class="[
            'inline-flex items-center gap-2 rounded-full border border-white/[0.06] bg-[#1c1c1e] px-4 py-1.5 label-medium text-[#a8a29e] transition-all duration-500',
            visible ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-4',
          ]"
          style="transition-timing-function: cubic-bezier(0.2, 0, 0, 1)"
        >
          <span class="h-1.5 w-1.5 rounded-full bg-[#fbbf24] animate-[pulse-ring_2s_ease-in-out_infinite]" />
          Join 2,000+ researchers
        </span>

        <h2
          :class="[
            'display-small mt-6 font-semibold leading-tight tracking-tight transition-all duration-500',
            visible ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-6',
          ]"
          style="font-family: var(--font-display); transition-delay: 100ms; transition-timing-function: cubic-bezier(0.2, 0, 0, 1)"
        >
          Start your<br />thesis journey
        </h2>

        <p
          :class="[
            'body-large mt-4 max-w-md font-light leading-relaxed text-[#a8a29e] transition-all duration-500',
            visible ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-4',
          ]"
          style="transition-delay: 200ms; transition-timing-function: cubic-bezier(0.2, 0, 0, 1)"
        >
          Create your profile in minutes
        </p>

        <div
          :class="[
            'mt-10 space-y-3 transition-all duration-500',
            visible ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-4',
          ]"
          style="transition-delay: 350ms; transition-timing-function: cubic-bezier(0.2, 0, 0, 1)"
        >
          <div class="flex items-center gap-3">
            <span class="flex h-6 w-6 items-center justify-center rounded-full bg-emerald-500/10">
              <svg class="h-3.5 w-3.5 text-emerald-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
                <polyline points="20 6 9 17 4 12" />
              </svg>
            </span>
            <span class="body-medium text-[#d6d3d1]">Free for students</span>
          </div>
          <div class="flex items-center gap-3">
            <span class="flex h-6 w-6 items-center justify-center rounded-full bg-emerald-500/10">
              <svg class="h-3.5 w-3.5 text-emerald-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
                <polyline points="20 6 9 17 4 12" />
              </svg>
            </span>
            <span class="body-medium text-[#d6d3d1]">AI transcription included</span>
          </div>
          <div class="flex items-center gap-3">
            <span class="flex h-6 w-6 items-center justify-center rounded-full bg-emerald-500/10">
              <svg class="h-3.5 w-3.5 text-emerald-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
                <polyline points="20 6 9 17 4 12" />
              </svg>
            </span>
            <span class="body-medium text-[#d6d3d1]">Real-time collaboration</span>
          </div>
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
            Create account
          </h1>
          <p class="body-medium text-[#a8a29e]">Complete your academic profile</p>
        </div>

        <form class="space-y-5" @submit.prevent="handleRegister">
          <div class="flex gap-2 rounded-xl border border-white/[0.06] bg-[#1c1c1e] p-1">
            <button
              type="button"
              :class="[
                'state-layer flex-1 rounded-lg px-4 py-2.5 label-medium transition-all duration-300 focus-ring',
                role === 'student'
                  ? 'bg-[#fbbf24] text-black'
                  : 'text-[#a8a29e] hover:text-[#d6d3d1]',
              ]"
              style="transition-timing-function: cubic-bezier(0.2, 0, 0, 1)"
              @click="role = 'student'"
            >
              Student
            </button>
            <button
              type="button"
              :class="[
                'state-layer flex-1 rounded-lg px-4 py-2.5 label-medium transition-all duration-300 focus-ring',
                role === 'lecturer'
                  ? 'bg-[#fbbf24] text-black'
                  : 'text-[#a8a29e] hover:text-[#d6d3d1]',
              ]"
              style="transition-timing-function: cubic-bezier(0.2, 0, 0, 1)"
              @click="role = 'lecturer'"
            >
              Advisor
            </button>
          </div>

          <div class="space-y-4">
            <UiField
              v-model="name"
              label="Name"
              placeholder="Enter your full name"
            />
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
              placeholder="Create a strong password"
            />
          </div>

          <div v-if="role === 'student'" class="space-y-4 border-t border-white/[0.06] pt-4">
            <UiField
              v-model="nim"
              label="NIM"
              placeholder="e.g. 20210001"
            />
            <div class="relative space-y-1.5">
              <label class="block text-[12px] font-medium text-stone-500 tracking-[0.5px] uppercase">
                Advisor
              </label>
              <select
                v-model="lecturerId"
                class="w-full bg-[#1c1c1e] border border-white/[0.08] rounded-xl px-4 py-3 text-[14px] text-stone-100 placeholder-stone-600 transition-all duration-200 focus:outline-none focus:border-amber-500/50 focus:ring-1 focus:ring-amber-500/20"
              >
                <option value="" disabled class="bg-[#141415] text-stone-500">Select Advisor</option>
                <option
                  v-for="lec in lecturers"
                  :key="lec.id"
                  :value="lec.id"
                  class="bg-[#141415] text-stone-100"
                >
                  {{ lec.name }}
                </option>
              </select>
            </div>
            <UiField
              v-model="prodi"
              label="Program of Study"
              placeholder="e.g. Informatics"
            />
            <UiField
              v-model="thesisTitle"
              label="Thesis Title"
              placeholder="Working thesis title"
            />
          </div>

          <div v-if="role === 'lecturer'" class="space-y-4 border-t border-white/[0.06] pt-4">
            <UiField
              v-model="nip"
              label="NIP"
              placeholder="e.g. 198501012010011001"
            />
            <UiField
              v-model="faculty"
              label="Faculty"
              placeholder="e.g. Faculty of Computer Science"
            />
            <UiField
              v-model="expertise"
              label="Expertise"
              placeholder="e.g. Software Engineering"
            />
          </div>

          <div
            v-if="error"
            class="rounded-xl border border-red-500/20 bg-red-500/5 p-3 text-[13px] text-red-400"
          >
            {{ error }}
          </div>

          <UiButton
            title="Create Account"
            :loading="loading"
            class="w-full"
          />
        </form>

        <p class="text-center body-medium text-[#a8a29e]">
          Already have an account?
          <Link href="/login" class="font-medium text-[#fbbf24] transition-colors duration-200 hover:text-[#f59e0b]">
            Sign in
          </Link>
        </p>
      </div>
    </div>
  </div>
</template>

<style scoped>
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

.display-small {
  font-size: 2.5rem;
  line-height: 3rem;
  letter-spacing: -0.025em;
}

.headline-small {
  font-size: 1.5rem;
  line-height: 2rem;
  font-weight: 600;
  letter-spacing: -0.025em;
}

.body-large {
  font-size: 1.125rem;
  line-height: 1.75rem;
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
