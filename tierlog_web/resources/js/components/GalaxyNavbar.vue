<script setup lang="ts">
import { ref, reactive, onMounted, onUnmounted } from 'vue';
import { Link } from '@inertiajs/vue3';

const hoveredNavItem = ref<string | null>(null);
const isMobileMenuOpen = ref(false);
const mobileDropdowns = reactive({
  features: false,
  enterprise: false,
  resources: false,
});

function handleMouseEnterNavItem(item: string) { hoveredNavItem.value = item; }
function handleMouseLeaveNavItem() { hoveredNavItem.value = null; }

function toggleMobileMenu() {
  isMobileMenuOpen.value = !isMobileMenuOpen.value;
  if (!isMobileMenuOpen.value) {
    mobileDropdowns.features = false;
    mobileDropdowns.enterprise = false;
    mobileDropdowns.resources = false;
  }
}

function toggleMobileDropdown(key: keyof typeof mobileDropdowns) {
  mobileDropdowns[key] = !mobileDropdowns[key];
}

function navLinkClass(itemName: string, extraClasses = '') {
  const isHovered = hoveredNavItem.value === itemName;
  const isAnotherHovered = hoveredNavItem.value !== null && !isHovered;
  const colorClass = isHovered ? 'text-white' : isAnotherHovered ? 'text-stone-600' : 'text-stone-400';
  return `text-sm transition duration-150 ${colorClass} ${extraClasses}`;
}

function handleResize() {
  if (window.innerWidth >= 1024 && isMobileMenuOpen.value) {
    isMobileMenuOpen.value = false;
    mobileDropdowns.features = false;
    mobileDropdowns.enterprise = false;
    mobileDropdowns.resources = false;
  }
}

onMounted(() => window.addEventListener('resize', handleResize));
onUnmounted(() => window.removeEventListener('resize', handleResize));
</script>

<template>
  <nav
    class="fixed top-0 left-0 right-0 z-20"
    :style="{
      backgroundColor: 'rgba(10, 10, 11, 0.3)',
      backdropFilter: 'blur(12px)',
      WebkitBackdropFilter: 'blur(12px)',
      borderRadius: '0 0 15px 15px',
    }"
  >
    <div class="container mx-auto px-4 py-4 md:px-6 lg:px-8 flex items-center justify-between">
      <div class="flex items-center space-x-6 lg:space-x-8">
        <Link href="/" class="flex items-center gap-2 text-white font-black tracking-tight" style="font-family: var(--font-display)">
          <span class="text-amber-400 text-xs">&#x25CF;</span>
          TierLog
        </Link>

        <div class="hidden lg:flex items-center space-x-6">
          <div class="relative group" @mouseenter="handleMouseEnterNavItem('features')" @mouseleave="handleMouseLeaveNavItem">
            <button :class="navLinkClass('features', 'flex items-center cursor-pointer')">
              Features
              <svg class="ml-1 w-3 h-3 group-hover:rotate-180 transition-transform duration-200" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" /></svg>
            </button>
            <div class="absolute left-0 mt-2 w-48 bg-[#141415]/90 rounded-xl shadow-lg py-2 border border-white/[0.06] opacity-0 invisible group-hover:opacity-100 group-hover:visible transition-all duration-200 z-30 backdrop-blur-xl">
              <Link href="/dashboard" class="block px-4 py-2 text-sm text-stone-400 hover:text-stone-100 hover:bg-white/[0.03] transition duration-150">Upload &amp; Transcribe</Link>
              <Link href="/dashboard" class="block px-4 py-2 text-sm text-stone-400 hover:text-stone-100 hover:bg-white/[0.03] transition duration-150">AI Analysis</Link>
              <Link href="/dashboard" class="block px-4 py-2 text-sm text-stone-400 hover:text-stone-100 hover:bg-white/[0.03] transition duration-150">Track &amp; Validate</Link>
            </div>
          </div>

          <div class="relative group" @mouseenter="handleMouseEnterNavItem('enterprise')" @mouseleave="handleMouseLeaveNavItem">
            <button :class="navLinkClass('enterprise', 'flex items-center cursor-pointer')">
              Enterprise
              <svg class="ml-1 w-3 h-3 group-hover:rotate-180 transition-transform duration-200" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" /></svg>
            </button>
            <div class="absolute left-0 mt-2 w-48 bg-[#141415]/90 rounded-xl shadow-lg py-2 border border-white/[0.06] opacity-0 invisible group-hover:opacity-100 group-hover:visible transition-all duration-200 z-30 backdrop-blur-xl">
              <Link href="/lecturer-dashboard" class="block px-4 py-2 text-sm text-stone-400 hover:text-stone-100 hover:bg-white/[0.03] transition duration-150">For Advisors</Link>
              <Link href="/register" class="block px-4 py-2 text-sm text-stone-400 hover:text-stone-100 hover:bg-white/[0.03] transition duration-150">For Students</Link>
            </div>
          </div>

          <div class="relative group" @mouseenter="handleMouseEnterNavItem('resources')" @mouseleave="handleMouseLeaveNavItem">
            <button :class="navLinkClass('resources', 'flex items-center cursor-pointer')">
              Resources
              <svg class="ml-1 w-3 h-3 group-hover:rotate-180 transition-transform duration-200" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" /></svg>
            </button>
            <div class="absolute left-0 mt-2 w-48 bg-[#141415]/90 rounded-xl shadow-lg py-2 border border-white/[0.06] opacity-0 invisible group-hover:opacity-100 group-hover:visible transition-all duration-200 z-30 backdrop-blur-xl">
              <Link href="/consultations" class="block px-4 py-2 text-sm text-stone-400 hover:text-stone-100 hover:bg-white/[0.03] transition duration-150">Consultations</Link>
              <Link href="/archive" class="block px-4 py-2 text-sm text-stone-400 hover:text-stone-100 hover:bg-white/[0.03] transition duration-150">Archive</Link>
              <Link href="/settings/ai-gateway" class="block px-4 py-2 text-sm text-stone-400 hover:text-stone-100 hover:bg-white/[0.03] transition duration-150">AI Gateway</Link>
            </div>
          </div>

          <Link href="#features" :class="navLinkClass('pricing')" @mouseenter="handleMouseEnterNavItem('pricing')" @mouseleave="handleMouseLeaveNavItem">
            Pricing
          </Link>
        </div>
      </div>

      <div class="flex items-center space-x-4 md:space-x-6">
        <Link href="/login" class="hidden sm:block text-stone-400 hover:text-white text-sm transition-colors">Sign In</Link>
        <Link
          href="/register"
          class="bg-amber-500/10 hover:bg-amber-500/20 text-amber-400 font-semibold py-2 px-5 rounded-full text-sm md:text-base border border-amber-500/20 transition-all duration-200"
        >
          Start Free
        </Link>
        <button class="lg:hidden text-white p-2 cursor-pointer" @click="toggleMobileMenu" aria-label="Toggle mobile menu">
          <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" :d="isMobileMenuOpen ? 'M6 18L18 6M6 6l12 12' : 'M4 6h16M4 12h16M4 18h16'" />
          </svg>
        </button>
      </div>
    </div>

    <div
      class="lg:hidden bg-[#0a0a0b]/90 border-t border-white/[0.04] absolute top-full left-0 right-0 z-30 overflow-hidden transition-all duration-300 ease-in-out backdrop-blur-xl"
      :class="isMobileMenuOpen ? 'max-h-screen opacity-100 pointer-events-auto' : 'max-h-0 opacity-0 pointer-events-none'"
    >
      <div class="px-4 py-6 flex flex-col space-y-4">
        <div class="relative">
          <button class="text-stone-400 hover:text-stone-100 flex items-center justify-between w-full text-left text-sm py-2 cursor-pointer" @click="toggleMobileDropdown('features')">
            Features
            <svg class="ml-2 w-3 h-3 transition-transform duration-200" :class="mobileDropdowns.features ? 'rotate-180' : ''" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" /></svg>
          </button>
          <div class="pl-4 space-y-2 mt-2 overflow-hidden transition-all duration-300 ease-in-out" :class="mobileDropdowns.features ? 'max-h-[200px] opacity-100 pointer-events-auto' : 'max-h-0 opacity-0 pointer-events-none'">
            <Link href="/dashboard" class="block text-stone-400 hover:text-stone-100 text-sm py-1 transition duration-150" @click="toggleMobileMenu">Upload &amp; Transcribe</Link>
            <Link href="/dashboard" class="block text-stone-400 hover:text-stone-100 text-sm py-1 transition duration-150" @click="toggleMobileMenu">AI Analysis</Link>
            <Link href="/dashboard" class="block text-stone-400 hover:text-stone-100 text-sm py-1 transition duration-150" @click="toggleMobileMenu">Track &amp; Validate</Link>
          </div>
        </div>
        <div class="relative">
          <button class="text-stone-400 hover:text-stone-100 flex items-center justify-between w-full text-left text-sm py-2 cursor-pointer" @click="toggleMobileDropdown('enterprise')">
            Enterprise
            <svg class="ml-2 w-3 h-3 transition-transform duration-200" :class="mobileDropdowns.enterprise ? 'rotate-180' : ''" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" /></svg>
          </button>
          <div class="pl-4 space-y-2 mt-2 overflow-hidden transition-all duration-300 ease-in-out" :class="mobileDropdowns.enterprise ? 'max-h-[200px] opacity-100 pointer-events-auto' : 'max-h-0 opacity-0 pointer-events-none'">
            <Link href="/lecturer-dashboard" class="block text-stone-400 hover:text-stone-100 text-sm py-1 transition duration-150" @click="toggleMobileMenu">For Advisors</Link>
            <Link href="/register" class="block text-stone-400 hover:text-stone-100 text-sm py-1 transition duration-150" @click="toggleMobileMenu">For Students</Link>
          </div>
        </div>
        <div class="relative">
          <button class="text-stone-400 hover:text-stone-100 flex items-center justify-between w-full text-left text-sm py-2 cursor-pointer" @click="toggleMobileDropdown('resources')">
            Resources
            <svg class="ml-2 w-3 h-3 transition-transform duration-200" :class="mobileDropdowns.resources ? 'rotate-180' : ''" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" /></svg>
          </button>
          <div class="pl-4 space-y-2 mt-2 overflow-hidden transition-all duration-300 ease-in-out" :class="mobileDropdowns.resources ? 'max-h-[250px] opacity-100 pointer-events-auto' : 'max-h-0 opacity-0 pointer-events-none'">
            <Link href="/consultations" class="block text-stone-400 hover:text-stone-100 text-sm py-1 transition duration-150" @click="toggleMobileMenu">Consultations</Link>
            <Link href="/archive" class="block text-stone-400 hover:text-stone-100 text-sm py-1 transition duration-150" @click="toggleMobileMenu">Archive</Link>
            <Link href="/settings/ai-gateway" class="block text-stone-400 hover:text-stone-100 text-sm py-1 transition duration-150" @click="toggleMobileMenu">AI Gateway</Link>
          </div>
        </div>
        <Link href="#features" class="text-stone-400 hover:text-stone-100 text-sm py-2 transition duration-150" @click="toggleMobileMenu">Pricing</Link>
        <Link href="/login" class="text-stone-400 hover:text-stone-100 text-sm py-2 transition duration-150" @click="toggleMobileMenu">Sign In</Link>
      </div>
    </div>
  </nav>
</template>
