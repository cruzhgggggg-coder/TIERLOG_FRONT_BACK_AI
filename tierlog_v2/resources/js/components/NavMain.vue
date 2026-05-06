<script setup lang="ts">
import { Link } from '@inertiajs/vue3';
import {
    SidebarGroup,
    SidebarGroupLabel,
    SidebarMenu,
    SidebarMenuButton,
    SidebarMenuItem,
} from '@/components/ui/sidebar';
import { useCurrentUrl } from '@/composables/useCurrentUrl';
import type { NavItem } from '@/types';

defineProps<{
    items: NavItem[];
}>();

const { isCurrentUrl } = useCurrentUrl();
</script>

<template>
    <SidebarGroup class="px-2 py-0">
        <SidebarGroupLabel class="text-[10px] font-black uppercase tracking-[0.3em] text-slate-600 mb-2">Neural Nodes</SidebarGroupLabel>
        <SidebarMenu class="gap-1">
            <SidebarMenuItem v-for="item in items" :key="item.title">
                <SidebarMenuButton
                    as-child
                    :class="[
                        'group/nav relative',
                        isCurrentUrl(item.href) ? 'is-active' : ''
                    ]"
                    :tooltip="item.title"
                >
                    <Link :href="item.href" class="flex items-center gap-3 px-4 py-3">
                        <component 
                            :is="item.icon" 
                            :class="[
                                'w-4 h-4 transition-all duration-500',
                                isCurrentUrl(item.href) ? 'text-indigo-400' : 'text-slate-500 group-hover/nav:text-white'
                            ]"
                        />
                        <span :class="[
                            'text-xs font-bold tracking-tight transition-all duration-300',
                            isCurrentUrl(item.href) ? 'text-white' : 'text-slate-500 group-hover/nav:text-slate-300'
                        ]">{{ item.title }}</span>
                        
                        <!-- Active Indicator -->
                        <div v-if="isCurrentUrl(item.href)" class="absolute right-2 w-1.5 h-1.5 bg-indigo-500 rounded-full shadow-[0_0_10px_#6366f1]"></div>
                    </Link>
                </SidebarMenuButton>
            </SidebarMenuItem>
        </SidebarMenu>
    </SidebarGroup>
</template>
