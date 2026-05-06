<script setup lang="ts">
import { computed } from 'vue';
import { Link, usePage } from '@inertiajs/vue3';
import { LayoutGrid, Mic, History } from 'lucide-vue-next';
import AppLogo from '@/components/AppLogo.vue';
import NavMain from '@/components/NavMain.vue';
import NavUser from '@/components/NavUser.vue';
import {
    Sidebar,
    SidebarContent,
    SidebarFooter,
    SidebarHeader,
    SidebarMenu,
    SidebarMenuItem,
} from '@/components/ui/sidebar';
import { dashboard } from '@/routes';
import type { NavItem } from '@/types';

const page = usePage();
const user = computed(() => page.props.auth.user);
const isLecturer = computed(() => user.value?.role === 'lecturer');

const mainNavItems = computed<NavItem[]>(() => {
    const items: NavItem[] = [
        {
            title: 'Dashboard',
            href: dashboard(),
            icon: LayoutGrid,
        }
    ];

    if (!isLecturer.value) {
        items.push(
            {
                title: 'Konsultasi AI',
                href: '/consultation',
                icon: Mic,
            },
            {
                title: 'Arsip Log',
                href: '/logs',
                icon: History,
            }
        );
    }

    return items;
});
</script>

<template>
    <Sidebar variant="inset" class="bg-[#050507] border-r border-white/5">
        <SidebarHeader class="p-6">
            <SidebarMenu>
                <SidebarMenuItem>
                    <div class="flex items-center gap-3 px-2 py-2">
                        <Link :href="dashboard()" class="flex items-center gap-3 group transition-all">
                            <div class="p-2 bg-indigo-600/20 border border-indigo-500/30 rounded-xl shadow-[0_0_15px_rgba(79,70,229,0.2)] group-hover:shadow-[0_0_20px_rgba(79,70,229,0.4)] transition-all">
                                <span class="text-xl">🧠</span>
                            </div>
                            <div class="flex flex-col">
                                <span class="text-sm font-black tracking-tighter text-white uppercase group-hover:text-indigo-400 transition-colors">TierLog</span>
                                <span class="text-[8px] font-black text-slate-500 uppercase tracking-widest">Sovereign AI</span>
                            </div>
                        </Link>
                    </div>
                </SidebarMenuItem>
            </SidebarMenu>
        </SidebarHeader>

        <SidebarContent class="px-2">
            <div class="h-px bg-linear-to-r from-transparent via-white/5 to-transparent my-4"></div>
            <NavMain :items="mainNavItems" />
        </SidebarContent>

        <SidebarFooter class="p-4 bg-zinc-950/50 backdrop-blur-md border-t border-white/5 rounded-b-4xl">
            <NavUser />
        </SidebarFooter>
    </Sidebar>
    <slot />
</template>

<style scoped>
@reference "tailwindcss";

:deep(.sidebar-menu-button) {
    @apply rounded-2xl transition-all duration-300;
}

:deep(.sidebar-menu-button:hover) {
    @apply bg-white/5 text-indigo-400;
}

:deep(.sidebar-menu-button.is-active) {
    @apply bg-indigo-600/10 text-indigo-400 border border-indigo-500/20 shadow-[0_0_20px_rgba(79,70,229,0.1)];
}
</style>
