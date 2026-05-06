<script setup lang="ts">
import { computed } from 'vue';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { useInitials } from '@/composables/useInitials';
import type { User } from '@/types';

type Props = {
    user: User;
    showEmail?: boolean;
};

const props = withDefaults(defineProps<Props>(), {
    showEmail: false,
});

const { getInitials } = useInitials();

// Compute whether we should show the avatar image
const showAvatar = computed(
    () => props.user.avatar && props.user.avatar !== '',
);
</script>

<template>
    <div class="relative group/avatar">
        <div class="absolute -inset-0.5 bg-linear-to-r from-indigo-500 to-cyan-500 rounded-lg blur opacity-20 group-hover/avatar:opacity-50 transition duration-500"></div>
        <Avatar class="relative h-9 w-9 overflow-hidden rounded-lg border border-white/10">
            <AvatarImage v-if="showAvatar" :src="user.avatar!" :alt="user.name" />
            <AvatarFallback class="rounded-lg bg-slate-900 text-white font-bold text-xs">
                {{ getInitials(user.name) }}
            </AvatarFallback>
        </Avatar>
    </div>

    <div class="grid flex-1 text-left text-sm leading-tight ml-3">
        <span class="truncate font-bold text-slate-200 tracking-tight">{{ user.name }}</span>
        <span v-if="showEmail" class="truncate text-[10px] font-medium text-slate-500 uppercase tracking-wider">{{
            user.email
        }}</span>
    </div>
</template>
