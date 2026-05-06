<script setup>
import { Head, Link } from '@inertiajs/vue3';
import { ref, computed } from 'vue';
import AppLayout from '@/layouts/AppLayout.vue';
import { format } from 'date-fns';

const props = defineProps({
    consultations: Array,
});

const searchQuery = ref('');
const filteredConsultations = computed(() => {
    if (!searchQuery.value) return props.consultations;
    return props.consultations.filter(c => 
        c.paper_filename.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
        c.student?.name?.toLowerCase().includes(searchQuery.value.toLowerCase())
    );
});

defineOptions({
    layout: {
        breadcrumbs: [
            { title: 'Dashboard', href: '/dashboard' },
            { title: 'Arsip Log', href: '/logs' },
        ],
    },
});
</script>

<template>
    <Head title="Arsip Log Konsultasi" />

    <AppLayout>
        <div class="py-12 px-4 sm:px-6 lg:px-8 bg-[#030305] min-h-screen text-white relative">
            <!-- Background Glows -->
            <div class="absolute top-0 right-1/4 w-96 h-96 bg-indigo-600/5 blur-[120px] rounded-full pointer-events-none"></div>
            
            <div class="max-w-7xl mx-auto space-y-8 relative z-10">
                <div class="flex flex-col md:flex-row md:items-center justify-between gap-6">
                    <div>
                        <h1 class="text-4xl font-black tracking-tighter uppercase">Arsip <span class="text-indigo-500">Log</span></h1>
                        <p class="text-slate-500 font-bold uppercase tracking-widest text-[10px] mt-2">Historical Neural Synchronization Data</p>
                    </div>
                    
                    <div class="relative w-full md:w-96">
                        <input 
                            v-model="searchQuery"
                            type="text" 
                            placeholder="Cari berdasarkan nama file..." 
                            class="w-full bg-white/5 border border-white/10 rounded-2xl px-6 py-4 text-sm focus:outline-none focus:border-indigo-500/50 transition-all placeholder:text-slate-600 font-bold"
                        />
                        <div class="absolute right-4 top-1/2 -translate-y-1/2 text-slate-600">
                            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" /></svg>
                        </div>
                    </div>
                </div>

                <div class="bg-white/2 border border-white/5 rounded-[2.5rem] overflow-hidden backdrop-blur-xl">
                    <div class="overflow-x-auto">
                        <table class="w-full text-left border-collapse">
                            <thead>
                                <tr class="border-b border-white/5">
                                    <th class="px-8 py-6 text-[10px] font-black uppercase tracking-[0.2em] text-slate-500">Neural Log</th>
                                    <th class="px-8 py-6 text-[10px] font-black uppercase tracking-[0.2em] text-slate-500">Timestamp</th>
                                    <th class="px-8 py-6 text-[10px] font-black uppercase tracking-[0.2em] text-slate-500">Artifacts</th>
                                    <th class="px-8 py-6 text-[10px] font-black uppercase tracking-[0.2em] text-slate-500">Feedback Node</th>
                                    <th class="px-8 py-6 text-[10px] font-black uppercase tracking-[0.2em] text-slate-500 text-right">Action</th>
                                </tr>
                            </thead>
                            <tbody class="divide-y divide-white/5">
                                <tr v-for="log in filteredConsultations" :key="log.id" class="group hover:bg-white/2 transition-all">
                                    <td class="px-8 py-6">
                                        <div class="flex items-center gap-4">
                                            <div class="w-10 h-10 bg-indigo-500/10 border border-indigo-500/20 rounded-xl flex items-center justify-center text-indigo-400 group-hover:scale-110 transition-all duration-500">
                                                <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" /></svg>
                                            </div>
                                            <div>
                                                <p class="text-sm font-black tracking-tight text-slate-200">{{ log.paper_filename.split('_').slice(1).join('_') || log.paper_filename }}</p>
                                                <p class="text-[10px] font-bold text-slate-500 uppercase tracking-widest mt-0.5">ID: #SYNC-{{ log.id.toString().padStart(4, '0') }}</p>
                                            </div>
                                        </div>
                                    </td>
                                    <td class="px-8 py-6">
                                        <p class="text-xs font-bold text-slate-300">{{ format(new Date(log.created_at), 'dd MMM yyyy') }}</p>
                                        <p class="text-[10px] font-mono text-slate-500 mt-1">{{ format(new Date(log.created_at), 'HH:mm:ss') }}</p>
                                    </td>
                                    <td class="px-8 py-6">
                                        <div class="flex gap-2">
                                            <span class="px-2 py-1 bg-blue-500/10 border border-blue-500/20 rounded-md text-[9px] font-black uppercase text-blue-400">DOCX</span>
                                            <span class="px-2 py-1 bg-purple-500/10 border border-purple-500/20 rounded-md text-[9px] font-black uppercase text-purple-400">WAV</span>
                                        </div>
                                    </td>
                                    <td class="px-8 py-6">
                                        <div class="flex items-center gap-3">
                                            <div class="flex -space-x-1">
                                                <div v-for="i in Math.min(log.feedback_items?.length || 0, 3)" :key="i" class="w-2 h-2 rounded-full bg-indigo-500 border border-[#030305]"></div>
                                            </div>
                                            <span class="text-xs font-bold text-indigo-400">{{ log.feedback_items?.length || 0 }} Items</span>
                                        </div>
                                    </td>
                                    <td class="px-8 py-6 text-right">
                                        <Link 
                                            href="/consultation" 
                                            class="inline-flex items-center gap-2 px-4 py-2 bg-white/5 border border-white/10 rounded-xl text-[10px] font-black uppercase tracking-widest hover:bg-white hover:text-black transition-all"
                                        >
                                            Open Matrix
                                            <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="3" d="M14 5l7 7m0 0l-7 7m7-7H3" /></svg>
                                        </Link>
                                    </td>
                                </tr>
                                <tr v-if="filteredConsultations.length === 0">
                                    <td colspan="5" class="px-8 py-20 text-center">
                                        <div class="flex flex-col items-center">
                                            <div class="w-20 h-20 bg-white/5 rounded-3xl flex items-center justify-center mb-6 border border-white/5">
                                                <svg class="w-10 h-10 text-slate-700" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4" /></svg>
                                            </div>
                                            <p class="text-slate-500 font-black uppercase tracking-[0.2em] text-xs">Arsip Kosong</p>
                                            <p class="text-[10px] text-slate-600 mt-2 font-bold uppercase tracking-widest">Mulai sinkronisasi pertama Anda di menu Konsultasi</p>
                                        </div>
                                    </td>
                                </tr>
                            </tbody>
                        </table>
                    </div>
                </div>
            </div>
        </div>
    </AppLayout>
</template>
