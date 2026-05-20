<script setup>
import { ref, onMounted, onUnmounted, computed } from 'vue';
import { usePage } from '@inertiajs/vue3';
import axios from 'axios';
import { CheckCircle, Clock, AlertCircle, User, FileText, ChevronRight, Activity } from 'lucide-vue-next';
import { toast } from 'vue-sonner';

const page = usePage();
const user = computed(() => page.props.auth.user);
const lecturer = computed(() => user.value?.lecturer);

const consultations = ref([]);
const students = ref([]);
const loading = ref(true);
const selectedLog = ref(null);

const currentChannel = ref(null);
const liveChatMessages = ref([]);

const fetchConsultations = async () => {
    if (!lecturer.value) return;
    try {
        loading.value = true;
        const [consRes, studRes] = await Promise.all([
            axios.get(`/lecturer/consultations`),
            axios.get(`/lecturer/students`)
        ]);
        consultations.value = consRes.data;
        students.value = studRes.data;
    } catch (error) {
        console.error('Error fetching dashboard data:', error);
    } finally {
        loading.value = false;
    }
};

const selectLog = (log) => {
    selectedLog.value = log;
    liveChatMessages.value = []; // Clear chat monitor for new selection

    if (window.Echo) {
        if (currentChannel.value) {
            window.Echo.leave(currentChannel.value);
        }
        currentChannel.value = `consultation.${log.id}`;
        window.Echo.channel(currentChannel.value)
            .listen('.feedback.status-updated', (e) => {
                const feedbackId = e.feedback_id;
                const newStatus = e.status;
                const updatedByRole = e.updated_by_role;

                // Update in selectedLog
                const item = selectedLog.value?.feedback_items?.find(f => f.id == feedbackId);
                if (item) {
                    item.status = newStatus;
                }
                // Update in consultations list
                consultations.value.forEach(clog => {
                    const citem = clog.feedback_items?.find(f => f.id == feedbackId);
                    if (citem) citem.status = newStatus;
                });

                // Show dynamic premium Sonner toast if it was modified by student
                if (updatedByRole === 'student') {
                    const content = item ? item.content : 'Revisi bimbingan';
                    toast.success(`Mahasiswa menandai revisi sebagai 'Fixed': "${content.substring(0, 35)}..."`, {
                        duration: 5000,
                        icon: '🔧'
                    });
                }
            })
            .listen('.chat.message', (e) => {
                liveChatMessages.value.push({ role: e.role, content: e.content });
            });
    }
};

const updateStatus = async (feedbackId, status) => {
    try {
        let logId = selectedLog.value?.id;
        if (selectedLog.value && selectedLog.value.feedback_items) {
            const item = selectedLog.value.feedback_items.find(f => f.id === feedbackId);
            if (item && item.log_id) {
                logId = item.log_id;
            }
        }

        await axios.put(`/consultation/feedback/${feedbackId}/status`, { 
            status,
            log_id: logId
        });
        // Update local state
        if (selectedLog.value) {
            const item = selectedLog.value.feedback_items.find(f => f.id === feedbackId);
            if (item) item.status = status;
        }
        // Also update in the main list
        consultations.value.forEach(log => {
            const item = log.feedback_items.find(f => f.id === feedbackId);
            if (item) item.status = status;
        });
    } catch (error) {
        console.error('Error updating status:', error);
    }
};

const getStatusColor = (status) => {
    switch (status) {
        case 'Validated': return 'text-green-400 bg-green-400/10 border-green-400/20';
        case 'Fixed': return 'text-blue-400 bg-blue-400/10 border-blue-400/20';
        default: return 'text-amber-400 bg-amber-400/10 border-amber-400/20';
    }
};

onMounted(() => {
    fetchConsultations();
});

onUnmounted(() => {
    if (window.Echo && currentChannel.value) {
        window.Echo.leave(currentChannel.value);
    }
});

const stats = computed(() => {
    const totalStudents = students.value.length;
    const pending = consultations.value.reduce((acc, log) => {
        return acc + log.feedback_items.filter(f => f.status === 'Pending').length;
    }, 0);
    return { students: totalStudents, pending, totalLogs: consultations.value.length };
});

</script>

<template>
    <div class="space-y-8 animate-in">
        <!-- Command Header -->
        <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
            <div class="bg-indigo-600 rounded-3xl p-6 shadow-xl shadow-indigo-500/20 border border-white/10 relative overflow-hidden group">
                <div class="absolute -right-4 -bottom-4 text-white/10 text-6xl font-black group-hover:scale-110 transition-transform">CMD</div>
                <p class="text-[10px] font-black uppercase tracking-[0.2em] opacity-70">Total Mentored Students</p>
                <h3 class="text-4xl font-black mt-2">{{ stats.students }}</h3>
                <div class="mt-4 flex items-center gap-2">
                    <User class="w-4 h-4" />
                    <span class="text-xs font-bold">Active Students</span>
                </div>
            </div>
            
            <div class="bg-white/5 border border-white/10 rounded-3xl p-6 backdrop-blur-xl group hover:border-indigo-500/30 transition-all">
                <p class="text-[10px] font-black uppercase tracking-[0.2em] text-slate-500">Validation Queue</p>
                <h3 class="text-4xl font-black mt-2 text-white">{{ stats.pending }}</h3>
                <div class="mt-4 flex items-center gap-2 text-amber-400">
                    <Activity class="w-4 h-4" />
                    <span class="text-xs font-bold uppercase tracking-widest">Pending Review</span>
                </div>
            </div>

            <div class="bg-white/5 border border-white/10 rounded-3xl p-6 backdrop-blur-xl group hover:border-purple-500/30 transition-all">
                <p class="text-[10px] font-black uppercase tracking-[0.2em] text-slate-500">Total Consultation Sessions</p>
                <h3 class="text-4xl font-black mt-2 text-white">{{ stats.totalLogs }}</h3>
                <div class="mt-4 flex items-center gap-2 text-purple-400">
                    <FileText class="w-4 h-4" />
                    <span class="text-xs font-bold uppercase tracking-widest">Logs Recorded</span>
                </div>
            </div>
        </div>

        <div class="grid grid-cols-1 lg:grid-cols-12 gap-8">
            <!-- Student Activity Stream -->
            <div class="lg:col-span-5 space-y-4">
                <div class="flex items-center justify-between mb-2">
                    <h3 class="text-sm font-black uppercase tracking-[0.2em] text-slate-400">Activity Stream</h3>
                    <button @click="fetchConsultations" class="text-[10px] font-black text-indigo-400 hover:text-indigo-300 uppercase tracking-widest">Refresh Sync</button>
                </div>
                
                <div v-if="loading" class="space-y-4">
                    <div v-for="i in 3" :key="i" class="h-24 bg-white/5 rounded-2xl animate-pulse"></div>
                </div>

                <div v-else-if="consultations.length === 0" class="p-12 text-center bg-white/5 border border-dashed border-white/10 rounded-3xl mb-8">
                    <p class="text-slate-500 font-bold uppercase tracking-widest text-[10px]">No mentoring activity yet</p>
                </div>

                <div v-else class="space-y-3 max-h-[400px] overflow-y-auto pr-2 custom-scrollbar mb-8">
                    <div 
                        v-for="log in consultations" 
                        :key="log.id"
                        @click="selectLog(log)"
                        :class="[
                            'p-4 rounded-2xl border transition-all cursor-pointer group',
                            selectedLog?.id === log.id 
                                ? 'bg-indigo-600/20 border-indigo-500/50 shadow-lg shadow-indigo-500/10' 
                                : 'bg-white/5 border-white/10 hover:border-white/20'
                        ]"
                    >
                        <!-- ... rest of consultation item ... -->
                        <div class="flex justify-between items-start mb-2">
                            <div class="flex items-center gap-3">
                                <div class="w-10 h-10 rounded-full bg-linear-to-br from-indigo-500 to-purple-500 flex items-center justify-center font-black text-xs text-white">
                                    {{ log.student.name.charAt(0) }}
                                </div>
                                <div>
                                    <p class="text-sm font-black group-hover:text-indigo-400 transition-colors">{{ log.student.name }}</p>
                                    <p class="text-[10px] font-bold text-slate-500 uppercase tracking-tighter">{{ log.student.nim }} • {{ new Date(log.created_at).toLocaleDateString() }}</p>
                                </div>
                            </div>
                            <ChevronRight class="w-4 h-4 text-slate-600 group-hover:text-white transition-all" />
                        </div>
                        <div class="flex gap-2">
                            <span v-if="log.feedback_items && log.feedback_items.length > 0 && log.feedback_items.some(f => f.status === 'Pending')" class="px-2 py-0.5 bg-amber-500/10 text-amber-500 border border-amber-500/20 rounded-md text-[9px] font-black uppercase tracking-tighter">
                                {{ log.feedback_items.filter(f => f.status === 'Pending').length }} Pending
                            </span>
                            <span v-else-if="log.feedback_items && log.feedback_items.length > 0" class="px-2 py-0.5 bg-green-500/10 text-green-500 border border-green-500/20 rounded-md text-[9px] font-black uppercase tracking-tighter">
                                All Validated
                            </span>
                            <span v-else class="px-2 py-0.5 bg-slate-500/10 text-slate-400 border border-slate-500/20 rounded-md text-[9px] font-black uppercase tracking-tighter">
                                No Feedback
                            </span>
                        </div>
                    </div>
                </div>

                <!-- All Students List -->
                <div class="pt-4 border-t border-white/10">
                    <h3 class="text-sm font-black uppercase tracking-[0.2em] text-slate-400 mb-4">Mentored Students List</h3>
                    <div v-if="students.length === 0" class="text-xs text-slate-500 italic">No students registered yet.</div>
                    <div class="grid grid-cols-1 gap-2">
                        <div v-for="student in students" :key="student.id" class="flex items-center justify-between p-3 bg-white/3 border border-white/5 rounded-xl">
                            <div class="flex items-center gap-3">
                                <User class="w-4 h-4 text-slate-500" />
                                <div>
                                    <p class="text-xs font-bold">{{ student.name }}</p>
                                    <p class="text-[9px] text-slate-500 uppercase tracking-tighter">{{ student.nim }} • {{ student.prodi }}</p>
                                </div>
                            </div>
                            <div class="w-2 h-2 rounded-full" :class="consultations.some(c => c.student_id === student.id) ? 'bg-indigo-500' : 'bg-slate-700'"></div>
                        </div>
                    </div>
                </div>
            </div>

            <!-- Detail View -->
            <div class="lg:col-span-7">
                <div v-if="selectedLog" class="bg-white/5 border border-white/10 rounded-3xl p-8 backdrop-blur-xl sticky top-8">
                    <div class="flex justify-between items-start mb-8">
                        <div>
                            <span class="inline-block px-3 py-1 bg-indigo-500/20 border border-indigo-500/30 rounded-full text-[10px] font-black tracking-[0.2em] text-indigo-400 uppercase mb-4">Consultation Detail</span>
                            <h2 class="text-2xl font-black tracking-tight">{{ selectedLog.student.name }}</h2>
                            <p class="text-slate-500 text-sm font-medium mt-1">{{ selectedLog.student.thesis_title }}</p>
                        </div>
                        <div class="text-right">
                            <p class="text-[10px] font-black text-slate-500 uppercase tracking-widest">Recorded At</p>
                            <p class="text-xs font-bold text-white">{{ new Date(selectedLog.created_at).toLocaleString() }}</p>
                        </div>
                    </div>

                    <div class="space-y-6">
                        <!-- Paper Reference -->
                        <div v-if="selectedLog.paper_filename" class="p-4 bg-white/5 border border-white/10 rounded-2xl flex items-center justify-between">
                            <div class="flex items-center gap-3">
                                <FileText class="w-5 h-5 text-indigo-400" />
                                <div>
                                    <p class="text-xs font-black">Draft Paper Analyzed</p>
                                    <p class="text-[10px] text-slate-500">{{ selectedLog.paper_filename }}</p>
                                </div>
                            </div>
                            <a :href="'http://localhost:8080/storage/paper/' + selectedLog.paper_filename" target="_blank" class="text-[10px] font-black text-indigo-400 hover:underline uppercase tracking-widest">View Draft</a>
                        </div>

                        <!-- Live AI Chat Monitor -->
                        <div class="p-6 bg-zinc-950/60 border border-white/5 rounded-2xl relative overflow-hidden group">
                            <!-- Glowing background synapse effect -->
                            <div class="absolute -right-20 -top-20 w-40 h-40 bg-indigo-500/5 rounded-full blur-3xl group-hover:bg-indigo-500/10 transition-all duration-700"></div>
                            
                            <div class="flex items-center justify-between mb-4">
                                <div class="flex items-center gap-3">
                                    <span class="relative flex h-2 w-2">
                                        <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-green-400 opacity-75"></span>
                                        <span class="relative inline-flex rounded-full h-2 w-2 bg-green-500"></span>
                                    </span>
                                    <h4 class="text-[10px] font-black uppercase tracking-[0.2em] text-slate-400">Live AI Chat Oversight</h4>
                                </div>
                                <span class="text-[8px] font-mono text-slate-500 uppercase tracking-widest">WebSocket Active</span>
                            </div>

                            <div v-if="liveChatMessages.length === 0" class="py-6 text-center text-slate-500 text-xs italic border border-dashed border-white/5 rounded-xl bg-black/20">
                                Standing by. Student messages to the AI Oracle will stream here in real time...
                            </div>
                            
                            <div v-else class="space-y-4 max-h-[220px] overflow-y-auto pr-2 custom-scrollbar">
                                <div v-for="(msg, index) in liveChatMessages" :key="index" class="flex flex-col gap-1 animate-slide-in">
                                    <div class="flex items-center gap-2">
                                        <span :class="[
                                            'text-[8px] font-black uppercase tracking-widest px-1.5 py-0.5 rounded',
                                            msg.role === 'user' ? 'bg-fuchsia-500/10 text-fuchsia-400 border border-fuchsia-500/20' : 'bg-cyan-500/10 text-cyan-400 border border-cyan-400/20'
                                        ]">
                                            {{ msg.role === 'user' ? 'Student' : 'Oracle AI' }}
                                        </span>
                                    </div>
                                    <p class="text-[11px] leading-relaxed text-slate-300 font-medium pl-1">{{ msg.content }}</p>
                                </div>
                            </div>
                        </div>

                        <!-- Feedback Items -->
                        <div>
                            <h4 class="text-[10px] font-black uppercase tracking-[0.2em] text-slate-500 mb-4">Neural Feedback & Validation</h4>
                            <div class="space-y-4">
                                <div v-for="feedback in selectedLog.feedback_items" :key="feedback.id" class="p-5 bg-white/5 border border-white/10 rounded-2xl relative group/item">
                                    <div class="flex justify-between items-start gap-4">
                                        <div class="flex-1">
                                            <div class="flex items-center gap-2 mb-2">
                                                <span :class="[
                                                    'px-2 py-0.5 rounded-md text-[9px] font-black uppercase tracking-tighter border',
                                                    feedback.category === 'Major' ? 'bg-red-500/10 text-red-500 border-red-500/20' : 'bg-blue-500/10 text-blue-500 border-blue-500/20'
                                                ]">
                                                    {{ feedback.category }}
                                                </span>
                                                <span :class="['px-2 py-0.5 rounded-md text-[9px] font-black uppercase tracking-tighter border transition-all', getStatusColor(feedback.status)]">
                                                    {{ feedback.status }}
                                                </span>
                                            </div>
                                            <p class="text-sm leading-relaxed text-slate-200">{{ feedback.content }}</p>
                                        </div>
                                        
                                        <!-- Actions -->
                                        <div class="flex flex-col gap-2 opacity-0 group-hover/item:opacity-100 transition-opacity">
                                            <button 
                                                v-if="feedback.status !== 'Validated'"
                                                @click="updateStatus(feedback.id, 'Validated')"
                                                class="p-2 bg-green-500/20 text-green-400 hover:bg-green-500/30 rounded-xl border border-green-500/30 transition-all title='Approve Revision'"
                                            >
                                                <CheckCircle class="w-4 h-4" />
                                            </button>
                                            <button 
                                                v-if="feedback.status === 'Pending'"
                                                @click="updateStatus(feedback.id, 'Fixed')"
                                                class="p-2 bg-blue-500/20 text-blue-400 hover:bg-blue-500/30 rounded-xl border border-blue-500/30 transition-all title='Mark as Done'"
                                            >
                                                <Activity class="w-4 h-4" />
                                            </button>
                                            <button 
                                                v-if="feedback.status !== 'Pending'"
                                                @click="updateStatus(feedback.id, 'Pending')"
                                                class="p-2 bg-amber-500/20 text-amber-400 hover:bg-amber-500/30 rounded-xl border border-amber-500/30 transition-all title='Needs Improvement'"
                                            >
                                                <Clock class="w-4 h-4" />
                                            </button>
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div>

                        <!-- Transcript Preview -->
                        <div class="mt-8">
                            <h4 class="text-[10px] font-black uppercase tracking-[0.2em] text-slate-500 mb-4">Transcription Log</h4>
                            <div class="p-6 bg-black/40 border border-white/5 rounded-2xl max-h-[200px] overflow-y-auto custom-scrollbar italic text-xs text-slate-400 leading-relaxed">
                                "{{ selectedLog.transcript_text }}"
                            </div>
                        </div>
                    </div>
                </div>

                <div v-else class="h-full flex flex-col items-center justify-center p-12 bg-white/5 border border-dashed border-white/10 rounded-3xl text-slate-600">
                    <Activity class="w-12 h-12 mb-4 opacity-20" />
                    <p class="font-black uppercase tracking-widest text-xs">Select a log to view details</p>
                </div>
            </div>
        </div>
    </div>
</template>

<style scoped>
.custom-scrollbar::-webkit-scrollbar { width: 4px; }
.custom-scrollbar::-webkit-scrollbar-track { background: transparent; }
.custom-scrollbar::-webkit-scrollbar-thumb { background: rgba(255, 255, 255, 0.1); border-radius: 10px; }

@keyframes fade-in { from { opacity: 0; transform: translateY(10px); } to { opacity: 1; transform: translateY(0); } }
.animate-in { animation: fade-in 0.7s cubic-bezier(0.2, 0.8, 0.2, 1) forwards; }

@keyframes slide-in {
    from { opacity: 0; transform: translateY(4px); }
    to { opacity: 1; transform: translateY(0); }
}
.animate-slide-in {
    animation: slide-in 0.3s cubic-bezier(0.16, 1, 0.3, 1) forwards;
}
</style>
