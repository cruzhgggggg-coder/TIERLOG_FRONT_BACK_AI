<script setup lang="ts">
import AppLogoIcon from '@/components/AppLogoIcon.vue';
</script>

<template>
    <div
        class="flex aspect-square size-10 items-center justify-center rounded-xl bg-linear-to-br from-indigo-600 to-cyan-600 shadow-[0_0_15px_rgba(79,70,229,0.4)]"
    >
        <AppLogoIcon class="size-6 fill-current text-white" />
    </div>
    <div class="ml-3 grid flex-1 text-left">
        <span class="text-base font-black tracking-tighter text-white uppercase leading-none">
            Tier<span class="text-indigo-400">Log</span>
        </span>
        <span class="text-[10px] font-bold text-slate-500 uppercase tracking-[0.2em] mt-0.5">
            Neural Nexus
        </span>
    </div>
</template>
