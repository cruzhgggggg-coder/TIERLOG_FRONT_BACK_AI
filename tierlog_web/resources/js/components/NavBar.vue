<script setup lang="ts">
import { computed } from 'vue';
import { Link, usePage } from '@inertiajs/vue3';
import { useAuthStore } from '@/stores/auth';
import { LogoutIcon } from '@/components/icons';

const auth = useAuthStore();
const page = usePage();

const user = computed(() => auth.user);
const currentUrl = computed(() => page.url);

interface NavItem {
  label: string;
  href: string;
}

const studentLinks: NavItem[] = [
  { label: 'Dashboard', href: '/dashboard' },
  { label: 'Consultation', href: '/consultations' },
  { label: 'Archive', href: '/archive' },
  { label: 'AI Gateway', href: '/settings/ai-gateway' },
  { label: 'Settings', href: '/settings/profile' },
];

const lecturerLinks: NavItem[] = [
  { label: 'Students', href: '/lecturer-dashboard' },
  { label: 'Archive', href: '/archive' },
  { label: 'Settings', href: '/settings/profile' },
];

const navLinks = computed(() =>
  user.value?.role === 'lecturer' ? lecturerLinks : studentLinks,
);

function isActive(href: string): boolean {
  return currentUrl.value === href || currentUrl.value.startsWith(href + '/');
}

async function handleLogout() {
  await auth.logout();
  window.location.href = '/login';
}
</script>

<template>
  <nav class="sticky top-0 z-50 bg-[#0a0a0b]/90 backdrop-blur-xl border-b border-white/[0.04]">
    <div class="mx-auto flex h-14 max-w-[1600px] items-center justify-between px-4 sm:px-6 lg:px-8 xl:px-12">
      <div class="flex items-center gap-8">
        <Link
          href="/dashboard"
          class="flex items-center gap-2 font-[family-name:var(--font-display)] font-black tracking-tight text-stone-100"
        >
          <span class="text-amber-400 text-xs">&#x25CF;</span>
          TierLog
        </Link>

        <div class="hidden items-center gap-0.5 md:flex">
          <Link
            v-for="item in navLinks"
            :key="item.href"
            :href="item.href"
            :class="[
              'label-large rounded-full px-3 py-1 transition-colors duration-200',
              isActive(item.href)
                ? 'bg-amber-500/10 text-amber-400'
                : 'text-stone-500 hover:text-stone-300',
            ]"
          >
            {{ item.label }}
          </Link>
        </div>
      </div>

      <div class="flex items-center gap-3">
        <div v-if="user" class="hidden items-center gap-2 sm:flex">
          <span class="text-[13px] font-medium text-stone-300">{{ user.name }}</span>
          <span class="rounded-full bg-amber-500/10 px-2.5 py-0.5 text-[10px] font-medium capitalize text-amber-400">
            {{ user.role }}
          </span>
        </div>
        <button
          class="flex items-center gap-1.5 rounded-md px-2.5 py-1.5 text-[13px] text-stone-600 transition-colors duration-200 hover:text-red-400 state-layer"
          @click="handleLogout"
        >
          <LogoutIcon :size="14" />
          <span class="hidden sm:inline">Logout</span>
        </button>
      </div>
    </div>
  </nav>
</template>
