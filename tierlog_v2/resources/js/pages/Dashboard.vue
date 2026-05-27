<script setup>
import { Head, Link, usePage } from '@inertiajs/vue3';
import AppLayout from '@/layouts/AppLayout.vue';
import { ref, onMounted, computed } from 'vue';
import axios from 'axios';
import { User } from 'lucide-vue-next';
import LecturerDashboard from '@/components/LecturerDashboard.vue';

const page = usePage();
const user = computed(() => page.props.auth.user);
const isLecturer = computed(() => user.value?.role === 'lecturer');

defineOptions({
    layout: {
        breadcrumbs: [
            {
                title: 'Dashboard',
                href: '/dashboard',
            },
        ],
    },
});

const stats = ref({
    total_consultations: 0,
    total_feedback: 0,
    pending_feedback: 0,
    major_feedback: 0,
    completion_rate: 0,
    draft_count: 0,
    upcoming_quests: []
});
const fetchStats = async () => {
    try {
        const response = await axios.get('/dashboard/stats');
        stats.value = response.data;
    } catch (error) {
        console.error('Error fetching stats:', error);
    }
};

onMounted(() => {
    fetchStats();
});
</script>

<template>
    <Head title="Dashboard" />

    <AppLayout>
        <div class="py-12 px-4 sm:px-6 lg:px-8 bg-[#030305] min-h-screen text-white animate-in relative">
            <!-- Ambient Background Glows -->
            <div class="absolute top-0 left-1/4 w-96 h-96 bg-indigo-600/10 blur-[120px] rounded-full pointer-events-none"></div>
            <div class="absolute bottom-0 right-1/4 w-96 h-96 bg-purple-600/10 blur-[120px] rounded-full pointer-events-none"></div>

            <div class="max-w-7xl mx-auto space-y-8 relative z-10">
                <!-- Lecturer View -->
                <div v-if="isLecturer">
                    <LecturerDashboard />
                </div>

                <!-- Student View (Original Content) -->
                <div v-else class="space-y-8">
                    <!-- Neural Sync Status Bar -->
                    <div class="flex items-center justify-between bg-white/5 border border-white/10 rounded-2xl px-6 py-3 backdrop-blur-md">
                        <div class="flex items-center gap-4">
                            <div class="flex items-center gap-2">
                                <span class="relative flex h-3 w-3">
                                    <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-indigo-400 opacity-75"></span>
                                    <span class="relative inline-flex rounded-full h-3 w-3 bg-indigo-500"></span>
                                </span>
                                <span class="text-[10px] font-black uppercase tracking-[0.2em] text-indigo-400">Neural Sync Active</span>
                            </div>
                            <div class="h-4 w-px bg-white/10"></div>
                            <p class="text-[10px] text-slate-500 uppercase tracking-widest font-bold">Latency: 24ms</p>
                        </div>
                        <div class="text-[10px] text-slate-500 font-mono">Last Calibrated: Just Now</div>
                    </div>

                    <!-- Hero Stats -->
                    <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
                        <!-- Sovereign AI Gateway Node -->
                        <div class="bg-indigo-600 rounded-4xl p-6 text-white shadow-2xl shadow-indigo-500/40 relative overflow-hidden group border border-white/20">
                            <div class="absolute -right-4 -bottom-4 text-white/10 text-6xl font-black group-hover:scale-110 transition-transform duration-500">AI</div>
                            <div class="relative z-10">
                                <p class="text-xs font-bold uppercase tracking-widest opacity-70">AI Gateway Status</p>
                                <h3 class="text-3xl font-black mt-2 tracking-tighter">
                                    {{ stats.completion_rate > 80 ? 'OPTIMAL' : (stats.completion_rate > 40 ? 'SYNERGY' : 'CONNECTED') }}
                                </h3>
                                <div class="mt-4 pt-4 border-t border-white/10 flex items-center gap-3">
                                    <div class="w-8 h-8 rounded-xl bg-white/10 flex items-center justify-center">
                                        <User class="w-4 h-4 text-indigo-200" />
                                    </div>
                                    <div class="flex flex-col">
                                        <p class="text-[8px] uppercase tracking-[0.2em] opacity-50 font-black">Lecturer Paired</p>
                                        <p class="text-[10px] font-black truncate max-w-[120px]">{{ stats.lecturer_name || 'Not Assigned' }}</p>
                                    </div>
                                </div>
                            </div>
                        </div>

                        <!-- Completion Node -->
                        <div class="bg-white/5 border border-white/10 rounded-4xl p-6 text-white backdrop-blur-xl group hover:border-indigo-500/50 transition-all duration-500">
                            <p class="text-xs font-bold uppercase tracking-widest text-slate-500 group-hover:text-indigo-400 transition-colors">Completion</p>
                            <h3 class="text-4xl font-black mt-2 tabular-nums">{{ stats.completion_rate }}<span class="text-lg opacity-50 ml-1">%</span></h3>
                            <div class="w-full bg-white/5 h-1.5 mt-6 rounded-full overflow-hidden border border-white/5">
                                <div class="bg-linear-to-r from-indigo-500 to-purple-500 h-full transition-all duration-1000 shadow-[0_0_10px_rgba(99,102,241,0.5)]" :style="{ width: stats.completion_rate + '%' }"></div>
                            </div>
                        </div>

                        <!-- Feedback Node -->
                        <div class="bg-white/5 border border-white/10 rounded-4xl p-6 text-white backdrop-blur-xl group hover:border-purple-500/50 transition-all duration-500">
                            <p class="text-xs font-bold uppercase tracking-widest text-slate-500 group-hover:text-purple-400 transition-colors">Feedback Pending</p>
                            <h3 class="text-4xl font-black mt-2 tabular-nums">{{ stats.pending_feedback }}</h3>
                            <div class="flex items-center gap-4 mt-4">
                                <div class="flex items-center gap-1.5">
                                    <span class="w-1.5 h-1.5 bg-red-500 rounded-full"></span>
                                    <span class="text-[10px] text-slate-400 font-bold uppercase tracking-tighter">{{ stats.major_feedback }} Major</span>
                                </div>
                                <div class="flex items-center gap-1.5">
                                    <span class="w-1.5 h-1.5 bg-blue-400 rounded-full"></span>
                                    <span class="text-[10px] text-slate-400 font-bold uppercase tracking-tighter">{{ stats.total_feedback - stats.major_feedback }} Minor</span>
                                </div>
                            </div>
                        </div>

                        <!-- Session Node -->
                        <div class="bg-linear-to-br from-purple-600 to-pink-600 rounded-4xl p-6 text-white shadow-2xl shadow-purple-500/30 border border-white/10 relative overflow-hidden group">
                            <div class="absolute top-0 right-0 p-4 opacity-20 group-hover:rotate-12 transition-transform duration-500">
                                <svg class="w-12 h-12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 2v20M2 12h20" /></svg>
                            </div>
                            <p class="text-xs font-bold uppercase tracking-widest opacity-70">Total Sessions</p>
                            <h3 class="text-4xl font-black mt-2 tabular-nums">{{ stats.total_consultations }}</h3>
                            <p class="text-[9px] mt-4 bg-white/20 inline-block px-2 py-1 rounded-full uppercase tracking-tighter font-black">Sync: June 2026</p>
                        </div>
                    </div>

                    <div class="grid grid-cols-1 lg:grid-cols-3 gap-8">
                        <!-- Main Action Card -->
                        <div class="lg:col-span-1 bg-white/5 border border-white/10 rounded-4xl p-8 backdrop-blur-xl group hover:border-indigo-500/30 transition-all relative overflow-hidden">
                            <div class="absolute -right-20 -top-20 w-80 h-80 bg-indigo-600/10 blur-[100px] rounded-full group-hover:bg-indigo-600/20 transition-all duration-700"></div>
                            <div class="absolute -left-20 -bottom-20 w-80 h-80 bg-purple-600/5 blur-[100px] rounded-full group-hover:bg-purple-600/10 transition-all duration-700"></div>
                            
                            <div class="relative z-10">
                                <span class="inline-block px-3 py-1 bg-indigo-500/20 border border-indigo-500/30 rounded-full text-[10px] font-black tracking-[0.2em] text-indigo-400 uppercase mb-4">Neural Interface Ready</span>
                                <h2 class="text-4xl font-black mb-4 tracking-tighter">Start Exploring <span class="italic text-indigo-500">AI Oracle</span></h2>
                                <p class="text-slate-400 max-w-md text-lg leading-relaxed mb-10">
                                    Real-time audio transcription and paper draft analysis with Gemini 2.0 Flash Core accuracy.
                                </p>
                                
                                <Link 
                                    href="/consultation" 
                                    class="inline-flex items-center gap-4 bg-white text-black px-10 py-5 rounded-2xl font-black hover:bg-indigo-50 transition-all active:scale-95 group/btn"
                                >
                                    <span class="tracking-tight text-lg">Start AI Consultation</span>
                                    <div class="bg-black text-white p-1 rounded-full group-hover/btn:translate-x-1 transition-transform">
                                        <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" viewBox="0 0 20 20" fill="currentColor">
                                            <path fill-rule="evenodd" d="M10.293 3.293a1 1 0 011.414 0l6 6a1 1 0 010 1.414l-6 6a1 1 0 01-1.414-1.414L14.586 11H3a1 1 0 110-2h11.586l-4.293-4.293a1 1 0 010-1.414z" clip-rule="evenodd" />
                                        </svg>
                                    </div>
                                </Link>
                            </div>
                        </div>

                        <!-- Sidebar: Upcoming Tasks -->
                        <div class="bg-slate-900/50 border border-white/5 rounded-[2.5rem] p-8 flex flex-col backdrop-blur-xl relative overflow-hidden group/sidebar">
                            <div class="absolute -right-10 -top-10 w-40 h-40 bg-indigo-500/5 blur-[50px] rounded-full"></div>
                            
                            <div class="flex items-center justify-between mb-8 relative z-10">
                                <h3 class="text-xl font-black tracking-tight">Upcoming Quests</h3>
                                <div class="flex items-center gap-2 px-3 py-1 bg-white/5 border border-white/10 rounded-full">
                                    <span class="w-1.5 h-1.5 bg-indigo-400 rounded-full animate-pulse"></span>
                                    <span class="text-[9px] font-black uppercase tracking-widest text-slate-400">{{ stats.upcoming_quests.length }} Active</span>
                                </div>
                            </div>
                            
                            <div class="space-y-4 flex-1 overflow-y-auto max-h-[420px] pr-2 custom-scrollbar relative z-10">
                                <div v-if="stats.upcoming_quests.length === 0" class="flex flex-col items-center justify-center py-20 text-center">
                                    <div class="w-16 h-16 bg-white/5 rounded-2xl flex items-center justify-center mb-4 border border-white/5">
                                        <svg class="w-8 h-8 text-slate-600" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
                                    </div>
                                    <p class="text-slate-500 text-xs font-black uppercase tracking-[0.2em]">All Quests Completed</p>
                                    <p class="text-[10px] text-slate-600 mt-2 font-bold uppercase tracking-widest">Oracle is Satisfied</p>
                                </div>
                                
                                <div 
                                    v-for="quest in stats.upcoming_quests" 
                                    :key="quest.id"
                                    class="p-5 bg-white/5 border border-white/5 rounded-3xl hover:border-indigo-500/40 hover:bg-white/5 transition-all duration-300 cursor-pointer group/quest relative overflow-hidden"
                                >
                                    <div class="absolute top-0 left-0 w-1 h-full" :class="quest.category === 'Major' ? 'bg-red-500' : 'bg-indigo-500'"></div>
                                    <div class="flex flex-col gap-3">
                                        <div class="flex items-center justify-between">
                                            <span :class="[
                                                'text-[9px] font-black uppercase tracking-widest px-2 py-0.5 rounded-md border',
                                                quest.category === 'Major' ? 'bg-red-500/10 border-red-500/20 text-red-400' : 'bg-indigo-500/10 border-indigo-500/20 text-indigo-400'
                                            ]">{{ quest.category }}</span>
                                            <span class="text-[9px] font-mono text-slate-600">#NODE-{{ quest.id.toString().padStart(3, '0') }}</span>
                                        </div>
                                        <p class="text-sm font-bold text-slate-200 leading-snug group-hover/quest:text-white transition-colors">{{ quest.content }}</p>
                                        
                                        <Link 
                                            href="/consultation"
                                            class="flex items-center gap-2 text-[10px] font-black uppercase tracking-widest text-indigo-400 mt-2 opacity-0 group-hover/quest:opacity-100 transition-all transform translate-y-1 group-hover/quest:translate-y-0"
                                        >
                                            Resolve Node
                                            <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="3" d="M14 5l7 7m0 0l-7 7m7-7H3" /></svg>
                                        </Link>
                                    </div>
                                </div>
                            </div>
                        </div>

                        <!-- Progress Heatmap -->
                        <div class="lg:col-span-3 bg-white/5 border border-white/10 rounded-[2.5rem] p-8">
                            <div class="flex items-center justify-between mb-6">
                                <h3 class="text-sm font-black uppercase tracking-[0.2em] text-slate-400">Activity Heatmap</h3>
                                <span class="text-[10px] text-indigo-400 font-bold uppercase tracking-widest">AI Gateway Active: 12 Days</span>
                            </div>
                            <div class="flex flex-wrap gap-2">
                                <div v-for="i in 56" :key="i" 
                                    :class="[
                                        'w-4 h-4 rounded-sm transition-all duration-500 hover:scale-125 cursor-help',
                                        i % 7 === 0 ? 'bg-indigo-600' : (i % 5 === 0 ? 'bg-indigo-400/50' : (i % 3 === 0 ? 'bg-indigo-900/30' : 'bg-white/5'))
                                    ]"
                                    :title="`Day ${i}: ${i % 7 === 0 ? 'High Activity' : 'Baseline'}`"
                                ></div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </AppLayout>
</template>

<style scoped>
@keyframes fade-in { from { opacity: 0; transform: translateY(10px); } to { opacity: 1; transform: translateY(0); } }
.animate-in { animation: fade-in 0.7s cubic-bezier(0.2, 0.8, 0.2, 1) forwards; }

.tilt-card {
    transition: transform 0.3s ease-out, box-shadow 0.3s ease-out;
    transform-style: preserve-3d;
    perspective: 1000px;
}

.tilt-card:hover {
    transform: translateY(-5px) rotateX(2deg) rotateY(1deg);
    box-shadow: 0 20px 40px rgba(0, 0, 0, 0.4), 0 0 20px rgba(99, 102, 241, 0.1);
}

/* Custom Scrollbar for sidebar */
.custom-scrollbar::-webkit-scrollbar { width: 4px; }
.custom-scrollbar::-webkit-scrollbar-track { background: transparent; }
.custom-scrollbar::-webkit-scrollbar-thumb { background: rgba(255, 255, 255, 0.1); border-radius: 10px; }
</style>
