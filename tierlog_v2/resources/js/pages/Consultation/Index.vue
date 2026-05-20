<script setup>
import { Head, useForm } from '@inertiajs/vue3';
import { ref, onMounted, onUnmounted, computed, nextTick } from 'vue';
import { DynamicScroller, DynamicScrollerItem } from 'vue-virtual-scroller';
import 'vue-virtual-scroller/dist/vue-virtual-scroller.css';

import axios from 'axios';
import AppLayout from '@/layouts/AppLayout.vue';
import NeuralNetwork from '@/components/NeuralNetwork.vue';
import NeuralPrint from '@/components/NeuralPrint.vue';
import { toast } from 'vue-sonner';


const props = defineProps({
    auth: Object,
    consultations: Array,
});

const form = useForm({
    user_id: props.auth?.user?.id || 1, 
    paper: null,
    audio: null,
});

const isProcessing = ref(false);
const processingStep = ref(0);
const uploadProgress = ref(0);

const steps = [
    "Initializing Gemini 2.0 Flash Core...",
    "Establishing Neural Link with Audio Stream...",
    "Parsing Draft Paper Architecture...",
    "Extracting HOC (Substantial) Feedback...",
    "Extracting LOC (Technical) Feedback...",
    "Calibrating Guarded AI Assistant...",
    "Finalizing Master Mind Status..."
];

const submit = async () => {
    if (!form.paper || !form.audio) {
        toast.error('Please select a document (.docx) and an audio recording before syncing.');
        return;
    }

    isProcessing.value = true;
    processingStep.value = 0;
    
    // Simulasikan langkah-langkah progres (Visual only for now)
    const stepInterval = setInterval(() => {
        if (processingStep.value < steps.length - 1) {
            processingStep.value++;
        }
    }, 3000);

    form.post('/consultation', {
        forceFormData: true,
        onUploadProgress: (progress) => {
            uploadProgress.value = Math.round((progress.loaded / progress.total) * 100);
        },
        onSuccess: () => {
            clearInterval(stepInterval);
            isProcessing.value = false;
            form.reset();
        },
        onError: (errors) => {
            clearInterval(stepInterval);
            isProcessing.value = false;
            
            const errorMsg = Object.values(errors).flat()[0] || 'Failed to process consultation. Please try again.';
            toast.error(errorMsg);
        },
    });
};

const selectedLog = ref(null);
const chatQuery = ref('');
const chatResponse = ref('');
const isChatting = ref(false);
const chatHistory = ref([]);
const selectedAIModel = ref('');

const availableModels = computed(() => {
    const models = [];
    
    if (props.auth.user.openai_key) {
        models.push({ id: 'openai:gpt-4o', name: 'OpenAI: GPT-4o' });
        models.push({ id: 'openai:gpt-4o-mini', name: 'OpenAI: GPT-4o Mini' });
        models.push({ id: 'openai:o1-preview', name: 'OpenAI: o1 Preview' });
        models.push({ id: 'openai:o1-mini', name: 'OpenAI: o1 Mini' });
    }
    
    if (props.auth.user.gemini_key) {
        models.push({ id: 'gemini:gemini-2.0-flash', name: 'Gemini: 2.0 Flash' });
        models.push({ id: 'gemini:gemini-1.5-pro', name: 'Gemini: 1.5 Pro' });
        models.push({ id: 'gemini:gemini-1.5-flash', name: 'Gemini: 1.5 Flash' });
    }
    
    if (props.auth.user.anthropic_key) {
        models.push({ id: 'anthropic:claude-3-5-sonnet-20240620', name: 'Claude: 3.5 Sonnet' });
        models.push({ id: 'anthropic:claude-3-opus-20240229', name: 'Claude: 3 Opus' });
        models.push({ id: 'anthropic:claude-3-haiku-20240307', name: 'Claude: 3 Haiku' });
    }
    
    if (props.auth.user.nvidia_key) {
        if (dynamicNvidiaModels.value.length > 0) {
            dynamicNvidiaModels.value.forEach(m => {
                models.push({ id: `nvidia:${m}`, name: `NVIDIA: ${m}` });
            });
        } else {
            models.push({ id: 'nvidia:meta/llama-3.1-70b-instruct', name: 'NVIDIA: Llama 3.1 70B' });
            models.push({ id: 'nvidia:meta/llama-3.1-8b-instruct', name: 'NVIDIA: Llama 3.1 8B' });
            models.push({ id: 'nvidia:meta/llama-3.1-405b-instruct', name: 'NVIDIA: Llama 3.1 405B' });
            models.push({ id: 'nvidia:bytedance/seed-oss-36b-instruct', name: 'NVIDIA: Seed OSS 36B' });
        }
    }

    if (models.length === 0) {
        models.push({ id: 'none', name: 'No AI Models Connected' });
    }
    
    return models;
});

const dynamicNvidiaModels = ref([]);

onMounted(() => {
    if (props.auth.user.nvidia_key) {
        axios.get(`http://localhost:8080/api/ai/models?provider=nvidia&api_key=${props.auth.user.nvidia_key}`)
            .then(res => {
                if (res.data && res.data.models) {
                    dynamicNvidiaModels.value = res.data.models;
                }
            })
            .catch(err => console.error('Failed to fetch dynamic NVIDIA models', err));
    }

    if (availableModels.value.length > 0) {
        selectedAIModel.value = availableModels.value[0].id;
    }
});

const askAI = async () => {
    if (!selectedLog.value || !chatQuery.value) return;
    
    const userMessage = chatQuery.value;
    chatHistory.value.push({ role: 'user', content: userMessage });
    chatQuery.value = '';
    isChatting.value = true;

    try {
        const response = await axios.post('/consultation/chat', {
            log_id: selectedLog.value.id,
            query: userMessage,
            model: selectedAIModel.value,
        });
        chatHistory.value.push({ 
            role: 'ai', 
            content: response.data.ai_response,
            source: response.data.source || 'Official Feedback'
        });
    } catch (e) {
        console.error(e);
        const errorMsg = e.response?.data?.message || e.response?.data?.error || 'An error occurred while contacting Master Mind.';
        chatHistory.value.push({ role: 'system', content: errorMsg });
    } finally {
        isChatting.value = false;
    }
};

const currentChannel = ref(null);

const selectLog = (log) => {
    selectedLog.value = log;
    chatHistory.value = [
        { role: 'system', content: `Master Mind active for session: ${log.paper_filename}. Ask anything about your revision.` }
    ];

    // Join Echo channel for real-time updates
    if (window.Echo) {
        if (currentChannel.value) {
            window.Echo.leave(currentChannel.value);
        }
        currentChannel.value = `consultation.${log.id}`;
        window.Echo.channel(currentChannel.value)
            .listen('.feedback.status-updated', (e) => {
                // payload: { feedback_id, log_id, status, updated_by_role }
                const item = selectedLog.value?.feedback_items?.find(f => f.id == e.feedback_id);
                if (item) {
                    item.status = e.status;
                    if (e.status === 'Validated' && e.updated_by_role === 'lecturer') {
                        toast.success(`Dosen memvalidasi revisi Anda!`, {
                            duration: 5000,
                            icon: '✅'
                        });
                    }
                }
            });
    }
};

const updateStatus = async (feedbackId, status) => {
    try {
        const logId = selectedLog.value?.id;
        await axios.put(`/consultation/feedback/${feedbackId}/status`, { 
            status,
            log_id: logId,
        });
        if (selectedLog.value) {
            const item = selectedLog.value.feedback_items.find(f => f.id === feedbackId);
            if (item) item.status = status;
        }
        toast.info('Status feedback diperbarui ke: ' + status, { icon: '🔄' });
    } catch (e) {
        console.error('Error updating status:', e);
        toast.error(e.response?.data?.error || 'Gagal memperbarui status feedback.');
    }
};

onUnmounted(() => {
    if (window.Echo && currentChannel.value) {
        window.Echo.leave(currentChannel.value);
    }
});

const allFeedback = computed(() => {
    if (!selectedLog.value) return [];
    const items = selectedLog.value.feedback_items || [];
    return items.map(item => ({
        ...item,
        type: item.category === 'Major' ? 'hoc' : 'loc'
    }));
});

const activeTab = ref('feedback');

const isNeuralSyncing = ref(false);

const askAIContextual = async (item) => {
    isNeuralSyncing.value = true;
    const context = item.category === 'HOC' ? 'Higher Order (Struktural)' : 'Lower Order (Teknis)';
    chatQuery.value = `[CONTRADICTION DETECTED] I need help with this ${context} feedback: "${item.content}". What does Master Mind suggest to fix this?`;
    
    // Smooth transition to chat
    await nextTick();
    const chatContainer = document.querySelector('.lg\\:col-span-4');
    if (chatContainer) {
        chatContainer.scrollIntoView({ behavior: 'smooth' });
    }

    await askAI();
    isNeuralSyncing.value = false;
};

const chatContainerRef = ref(null);
const scrollToBottom = async () => {
    await nextTick();
    if (chatContainerRef.value) {
        chatContainerRef.value.scrollTo({
            top: chatContainerRef.value.scrollHeight,
            behavior: 'smooth'
        });
    }
};

import { watch } from 'vue';
watch(chatHistory, () => {
    scrollToBottom();
}, { deep: true });
</script>

<template>
    <Head title="Konsultasi AI - Master Mind" />

    <AppLayout>
        <div class="relative min-h-screen bg-[#050507] text-white overflow-hidden">
            <!-- Neural Link Overlay -->
            <transition name="fade">
                <div v-if="isProcessing" class="fixed inset-0 z-50 flex flex-col items-center justify-center bg-[#050507]/95 backdrop-blur-xl">
                    <div class="absolute inset-0 pointer-events-none">
                        <NeuralNetwork :processing="isProcessing" />
                    </div>
                    
                    <div class="relative z-10 flex flex-col items-center">
                        <div class="relative w-48 h-48 mb-12">
                            <div class="absolute inset-0 border-2 border-indigo-500/30 rounded-full animate-ping"></div>
                            <div class="absolute inset-4 border-2 border-purple-500/20 rounded-full animate-ping [animation-delay:0.5s]"></div>
                            <div class="absolute inset-0 flex items-center justify-center">
                                <div class="w-24 h-24 bg-linear-to-br from-indigo-600 to-purple-600 rounded-full shadow-[0_0_50px_rgba(79,70,229,0.5)] flex items-center justify-center">
                                    <span class="text-4xl animate-bounce">🧠</span>
                                </div>
                            </div>
                        </div>
                    </div>
                    
                    <h2 class="text-3xl font-black tracking-tighter mb-4 text-transparent bg-clip-text bg-linear-to-r from-indigo-400 to-purple-400">
                        Neural Link Active
                    </h2>
                    
                    <div class="space-y-4 text-center max-w-md px-6">
                        <p class="text-indigo-300 font-mono text-sm animate-pulse h-6">
                            {{ steps[processingStep] }}
                        </p>
                        <!-- Progress Bar -->
                        <div class="w-full bg-white/5 h-1 rounded-full overflow-hidden border border-white/5">
                            <div 
                                class="bg-linear-to-r from-indigo-500 to-purple-500 h-full transition-all duration-500"
                                :style="{ width: uploadProgress + '%' }"
                            ></div>
                        </div>
                        <p class="text-[10px] text-slate-500 uppercase tracking-widest">
                            Upload Progress: {{ uploadProgress }}%
                        </p>
                    </div>
                </div>
            </transition>

            <div class="max-w-[1600px] mx-auto p-6 lg:p-10 h-screen flex flex-col">
                <!-- Top Navigation & Header -->
                <div class="flex flex-col md:flex-row justify-between items-start md:items-center mb-8 gap-4">
                    <div>
                        <h1 class="text-3xl font-black tracking-tighter flex items-center gap-3">
                            <span class="p-2 bg-indigo-600/20 border border-indigo-500/30 rounded-xl shadow-[0_0_20px_rgba(79,70,229,0.2)]">🧠</span>
                            Consultation <span class="text-[#00f2ff] italic">Workspace</span>
                        </h1>
                        <p class="text-xs text-slate-500 mt-1 uppercase tracking-widest font-bold">Neural Interface :: Session Active</p>
                    </div>
                    
                    <div class="flex items-center gap-4">
                        <div class="flex items-center gap-2 px-4 py-2 bg-white/5 border border-white/10 rounded-full text-[10px] font-black uppercase tracking-[0.2em] text-slate-400">
                            <span class="w-1.5 h-1.5 bg-[#00f2ff] rounded-full animate-pulse shadow-[0_0_8px_#00f2ff]"></span>
                            Master Mind v2.0
                        </div>
                    </div>
                </div>

                <div class="flex-1 grid grid-cols-1 lg:grid-cols-12 gap-8 overflow-hidden">
                    <!-- Column 1: Control Center (3/12) -->
                    <div class="lg:col-span-3 flex flex-col gap-6 overflow-y-auto pr-2 custom-scrollbar">
                        <!-- Upload Card -->
                        <div class="neural-glass p-8 relative overflow-hidden group">
                            <div class="absolute top-0 right-0 p-2 opacity-10 group-hover:opacity-30 transition-opacity">
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-12 w-12" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12" /></svg>
                            </div>
                            
                            <h3 class="text-[10px] font-black uppercase tracking-[0.3em] text-slate-500 mb-8 flex items-center gap-3">
                                <span class="w-1 h-4 bg-[#00f2ff] rounded-full shadow-[0_0_10px_#00f2ff]"></span>
                                Upload Draft
                            </h3>
                            
                            <form @submit.prevent="submit" class="space-y-6">
                                <div class="space-y-4">
                                    <label class="group/upload block relative cursor-pointer">
                                        <input type="file" @input="form.paper = $event.target.files[0]" class="hidden" accept=".docx" />
                                        <div class="border border-white/5 rounded-2xl p-6 text-center hover:bg-white/5 transition-all duration-300 group-hover/upload:border-[#00f2ff]/30">
                                            <span class="text-2xl mb-2 block filter grayscale group-hover/upload:grayscale-0 transition-all">📄</span>
                                            <p class="text-[10px] font-black uppercase tracking-widest text-slate-400">{{ form.paper ? form.paper.name : 'Select Manuscript' }}</p>
                                        </div>
                                    </label>
                                    
                                    <label class="group/upload block relative cursor-pointer">
                                        <input type="file" @input="form.audio = $event.target.files[0]" class="hidden" accept="audio/*" />
                                        <div class="border border-white/5 rounded-2xl p-6 text-center hover:bg-white/5 transition-all duration-300 group-hover/upload:border-[#ff00ea]/30">
                                            <span class="text-2xl mb-2 block filter grayscale group-hover/upload:grayscale-0 transition-all">🎙️</span>
                                            <p class="text-[10px] font-black uppercase tracking-widest text-slate-400">{{ form.audio ? form.audio.name : 'Select Recording' }}</p>
                                        </div>
                                    </label>
                                </div>
                                
                                <button 
                                    type="submit" 
                                    class="w-full bg-white text-black py-4 rounded-2xl font-black text-xs uppercase tracking-[0.2em] hover:bg-[#00f2ff] transition-all active:scale-95 flex items-center justify-center gap-3 shadow-[0_20px_40px_rgba(0,0,0,0.4)]"
                                >
                                    <span>Sync Neural Link</span>
                                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M10.293 3.293a1 1 0 011.414 0l6 6a1 1 0 010 1.414l-6 6a1 1 0 01-1.414-1.414L14.586 11H3a1 1 0 110-2h11.586l-4.293-4.293a1 1 0 010-1.414z" clip-rule="evenodd" /></svg>
                                </button>
                            </form>
                        </div>

                        <!-- History List -->
                        <div class="flex-1 min-h-0 flex flex-col">
                            <h3 class="text-[10px] font-black uppercase tracking-[0.3em] text-slate-500 mb-6 flex items-center gap-3 px-2">
                                <span class="w-1 h-4 bg-[#ff00ea] rounded-full shadow-[0_0_10px_#ff00ea]"></span>
                                Archive
                            </h3>
                            <div class="space-y-3 overflow-y-auto flex-1 pr-2 custom-scrollbar">
                                <div 
                                    v-for="log in consultations" 
                                    :key="log.id"
                                    @click="selectLog(log)"
                                    :class="[
                                        'p-5 rounded-3xl border cursor-pointer transition-all duration-300 group relative overflow-hidden',
                                        selectedLog?.id === log.id ? 'bg-indigo-600/10 border-indigo-500/40 shadow-xl' : 'bg-white/5 border-white/5 hover:border-white/10 hover:bg-white/10'
                                    ]"
                                >
                                    <div class="absolute inset-0 bg-linear-to-r from-indigo-500/10 via-purple-500/10 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-1000 animate-pulse"></div>
                                    <p class="text-[11px] font-bold truncate group-hover:text-indigo-400 tracking-tight">{{ log.paper_filename }}</p>
                                    <p class="text-[9px] text-slate-500 mt-1.5 uppercase font-black tracking-widest">{{ new Date(log.created_at).toLocaleDateString(undefined, { month: 'short', day: 'numeric', year: 'numeric' }) }}</p>
                                </div>
                            </div>
                        </div>
                    </div>

                    <!-- Column 2: Feedback Repository (5/12) -->
                    <div class="lg:col-span-5 flex flex-col gap-6 overflow-hidden">
                        <div v-if="selectedLog" class="flex-1 flex flex-col neural-glass p-10 overflow-hidden relative border-white/10">
                            <!-- Integrated Processing Indicator (Internal Neural Link) -->
                            <div v-if="isProcessing" class="absolute inset-0 z-20 flex flex-col items-center justify-center bg-zinc-950/90 backdrop-blur-xl rounded-4xl">
                                <div class="absolute inset-0 opacity-40">
                                    <NeuralNetwork :processing="isProcessing || isNeuralSyncing" />
                                </div>
                                <div class="w-32 h-32 mb-10 relative z-10">
                                    <div class="absolute inset-0 border-t border-cyan-400 rounded-full animate-spin"></div>
                                    <div class="absolute inset-4 border-b border-fuchsia-400 rounded-full animate-spin animation-duration-[2s]"></div>
                                    <div class="absolute inset-0 flex items-center justify-center text-3xl filter drop-shadow-[0_0_15px_rgba(34,211,238,0.5)]">🧠</div>
                                </div>
                                <div class="text-center relative z-10">
                                    <p class="text-[10px] font-black uppercase tracking-[0.4em] text-cyan-400 mb-2 animate-pulse">{{ steps[processingStep] }}</p>
                                    <div class="w-48 h-px bg-white/5 mx-auto relative overflow-hidden">
                                        <div class="absolute inset-0 bg-linear-to-r from-transparent via-cyan-400 to-transparent animate-pulse"></div>
                                    </div>
                                </div>
                            </div>

                            <div class="flex items-center justify-between mb-10">
                                <div>
                                    <h3 class="text-2xl font-black tracking-tighter">
                                        <span v-if="activeTab === 'feedback'">Feedback <span class="text-indigo-400">Stream</span></span>
                                        <span v-else>Audio <span class="text-indigo-400">Transcript</span></span>
                                    </h3>
                                    <p class="text-[9px] text-slate-500 uppercase tracking-[0.2em] font-black mt-1">
                                        <span v-if="activeTab === 'feedback'">Official Logic Analysis</span>
                                        <span v-else>Transcription Data</span>
                                    </p>
                                </div>
                                <div class="flex gap-2">
                                    <button @click="activeTab = 'feedback'" :class="['px-4 py-2 text-[10px] font-black uppercase tracking-[0.2em] rounded-full border transition-all', activeTab === 'feedback' ? 'bg-indigo-500/10 text-cyan-400 border-cyan-400/30' : 'bg-transparent text-slate-500 border-white/5 hover:border-white/10']">
                                        Feedback
                                    </button>
                                    <button @click="activeTab = 'transcript'" :class="['px-4 py-2 text-[10px] font-black uppercase tracking-[0.2em] rounded-full border transition-all', activeTab === 'transcript' ? 'bg-indigo-500/10 text-fuchsia-400 border-fuchsia-400/30' : 'bg-transparent text-slate-500 border-white/5 hover:border-white/10']">
                                        Transcript
                                    </button>
                                </div>
                            </div>

                            <div class="flex-1 overflow-hidden">
                                <DynamicScroller
                                    v-if="activeTab === 'feedback' && allFeedback.length"
                                    :items="allFeedback"
                                    :min-item-size="100"
                                    class="h-full pr-4 custom-scrollbar"
                                    key-field="id"
                                >
                                    <template v-slot="{ item, index, active }">
                                        <DynamicScrollerItem
                                            :item="item"
                                            :active="active"
                                            :size-dependencies="[item.content]"
                                            :data-index="index"
                                            class="mb-6"
                                        >
                                            <div 
                                                :class="[
                                                    'p-6 rounded-3xl border transition-all relative overflow-hidden group bg-zinc-900/50',
                                                    item.type === 'hoc' ? 'border-white/5' : 'border-white/5'
                                                ]"
                                            >
                                                <!-- Category & Status Badges -->
                                                <div class="flex items-center justify-between mb-4">
                                                    <div class="flex items-center gap-3">
                                                        <span :class="['w-2 h-2 rounded-full shadow-lg', item.type === 'hoc' ? 'bg-fuchsia-400 shadow-fuchsia-400/50' : 'bg-cyan-400 shadow-cyan-400/50']"></span>
                                                        <span class="text-[9px] font-black uppercase tracking-[0.2em] text-slate-500">
                                                            {{ item.type === 'hoc' ? 'Structural logic :: HOC' : 'Technical precision :: LOC' }}
                                                        </span>
                                                    </div>
                                                    <span :class="[
                                                        'px-2 py-0.5 rounded-md text-[9px] font-black uppercase tracking-tighter border transition-all duration-300 backdrop-blur-xs',
                                                        item.status === 'Validated' ? 'bg-green-500/10 text-green-400 border-green-500/20 shadow-[0_0_10px_rgba(34,197,94,0.1)]' :
                                                        item.status === 'Fixed' ? 'bg-yellow-500/10 text-yellow-400 border-yellow-500/20 shadow-[0_0_10px_rgba(234,179,8,0.1)]' :
                                                        'bg-rose-500/10 text-rose-400 border-rose-500/20 shadow-[0_0_10px_rgba(244,63,94,0.1)]'
                                                    ]">
                                                        {{ item.status || 'Pending' }}
                                                    </span>
                                                </div>

                                                <div class="flex justify-between gap-6">
                                                    <p class="text-[13px] leading-relaxed text-slate-300 font-medium">{{ item.content }}</p>
                                                    <div class="flex flex-col gap-2 opacity-0 group-hover:opacity-100 transition-all translate-x-2 group-hover:translate-x-0">
                                                        <button 
                                                            @click="askAIContextual(item)"
                                                            class="p-2.5 rounded-xl transition-all hover:scale-110 active:scale-90 shadow-xl relative overflow-hidden"
                                                            :class="[
                                                                item.type === 'hoc' ? 'bg-fuchsia-400/10 text-fuchsia-400 border border-fuchsia-400/20' : 'bg-cyan-400/10 text-cyan-400 border border-cyan-400/20',
                                                                isNeuralSyncing ? 'animate-pulse' : ''
                                                            ]"
                                                            :disabled="isNeuralSyncing"
                                                            title="Analyze with Oracle"
                                                        >
                                                            <svg v-if="!isNeuralSyncing" xmlns="http://www.w3.org/2000/svg" class="h-4.5 w-4.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M13 10V3L4 14h7v7l9-11h-7z" />
                                                            </svg>
                                                            <div v-else class="w-4.5 h-4.5 border-2 border-current border-t-transparent rounded-full animate-spin"></div>
                                                        </button>

                                                        <!-- Mark as Fixed (shown when Pending) -->
                                                        <button 
                                                            v-if="item.status === 'Pending' || !item.status"
                                                            @click="updateStatus(item.id, 'Fixed')"
                                                            class="p-2.5 rounded-xl transition-all hover:scale-110 active:scale-90 shadow-xl relative overflow-hidden bg-yellow-500/10 text-yellow-400 border border-yellow-500/20"
                                                            title="Mark as Fixed"
                                                        >
                                                            <svg xmlns="http://www.w3.org/2000/svg" class="h-4.5 w-4.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M5 13l4 4L19 7" />
                                                            </svg>
                                                        </button>

                                                        <!-- Undo to Pending (shown when Fixed, before Validated) -->
                                                        <button 
                                                            v-if="item.status === 'Fixed'"
                                                            @click="updateStatus(item.id, 'Pending')"
                                                            class="p-2.5 rounded-xl transition-all hover:scale-110 active:scale-90 shadow-xl relative overflow-hidden bg-rose-500/10 text-rose-400 border border-rose-500/20"
                                                            title="Undo — Kembalikan ke Pending"
                                                        >
                                                            <svg xmlns="http://www.w3.org/2000/svg" class="h-4.5 w-4.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M3 10h10a8 8 0 018 8v2M3 10l6 6m-6-6l6-6" />
                                                            </svg>
                                                        </button>
                                                    </div>
                                                </div>
                                                
                                                <!-- Decorative Synapse Line -->
                                                <div class="absolute bottom-0 left-0 h-[2px] w-0 group-hover:w-full transition-all duration-1000 ease-out"
                                                    :class="item.type === 'hoc' ? 'bg-linear-to-r from-transparent via-fuchsia-400 to-transparent' : 'bg-linear-to-r from-transparent via-cyan-400 to-transparent'"
                                                ></div>
                                            </div>
                                        </DynamicScrollerItem>
                                    </template>
                                </DynamicScroller>
                                
                                <div v-else-if="activeTab === 'transcript'" class="h-full overflow-y-auto pr-4 custom-scrollbar text-slate-300 text-[13px] leading-relaxed p-6 rounded-3xl border border-white/5 bg-zinc-900/50">
                                    <p class="whitespace-pre-wrap">{{ selectedLog.transcript_text || 'No transcript available.' }}</p>
                                </div>
                            </div>
                        </div>
                        <div v-else class="flex-1 flex flex-col items-center justify-center bg-white/5 border border-dashed border-white/5 rounded-4xl opacity-40 backdrop-blur-sm grayscale hover:grayscale-0 transition-all duration-700">
                            <div class="w-24 h-24 bg-white/5 rounded-full flex items-center justify-center mb-8 border border-white/5 animate-pulse shadow-inner">
                                <span class="text-5xl opacity-50">📡</span>
                            </div>
                            <h4 class="text-[10px] font-black uppercase tracking-[0.5em] text-slate-500">Neural Link Offline</h4>
                            <p class="text-[9px] text-slate-600 mt-2 uppercase tracking-widest font-bold italic">Initialize session via control center</p>
                        </div>
                    </div>

                    <!-- Column 3: Guarded Oracle (4/12) -->
                    <div class="lg:col-span-4 flex flex-col overflow-hidden">
                        <div v-if="selectedLog" 
                            :class="[
                                'flex-1 flex flex-col bg-zinc-950 border border-white/5 rounded-4xl overflow-hidden shadow-2xl relative transition-all duration-500',
                                isNeuralSyncing ? 'animate-pulse' : ''
                            ]"
                        >
                            <div class="h-px bg-linear-to-r from-transparent via-indigo-500/50 to-transparent"></div>
                            
                            <!-- Chat Header -->
                            <div class="p-8 border-b border-white/5 bg-white/5 backdrop-blur-md">
                                <div class="flex items-center justify-between">
                                    <div>
                                        <h3 class="font-black flex items-center gap-3 text-lg tracking-tighter">
                                            <span class="w-2 h-2 bg-[#00f2ff] rounded-full shadow-[0_0_10px_#00f2ff] animate-pulse"></span>
                                            Oracle <span class="text-indigo-400 font-light italic">Consultant</span>
                                        </h3>
                                        <p class="text-[9px] text-slate-500 mt-1.5 uppercase tracking-[0.2em] font-black flex items-center gap-2">
                                            <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3 text-green-500" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M2.166 4.999A11.954 11.954 0 0010 1.944 11.954 11.954 0 0017.834 5c.11.65.166 1.32.166 2.001 0 5.225-3.34 9.67-8 11.317C5.34 16.67 2 12.225 2 7c0-.682.057-1.35.166-2.001zm11.541 3.708a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" /></svg>
                                            Guard: Academic Integrity
                                        </p>
                                    </div>
                                </div>
                            </div>

                            <!-- Messages Area -->
                            <div ref="chatContainerRef" class="flex-1 overflow-y-auto p-8 space-y-8 custom-scrollbar">
                                <div v-for="(msg, i) in chatHistory" :key="i" :class="['flex', msg.role === 'user' ? 'justify-end' : 'justify-start']">
                                    <div :class="[
                                        'max-w-[90%] p-5 rounded-3xl text-[13px] leading-relaxed shadow-2xl relative transition-all duration-500',
                                        msg.role === 'user' ? 'bg-indigo-600 text-white rounded-tr-none border border-white/10' : (msg.role === 'system' ? 'bg-white/2 border border-white/5 text-slate-500 text-[10px] uppercase font-black tracking-widest text-center w-full rounded-xl py-3' : 'bg-white/5 border border-white/10 text-slate-200 rounded-tl-none backdrop-blur-xl')
                                    ]">
                                        <div v-if="msg.role === 'ai'" class="absolute -top-3 -left-3 w-8 h-8 bg-indigo-600 rounded-full flex items-center justify-center text-xs shadow-lg border-2 border-[#050507]">🤖</div>
                                        <template v-if="msg.role === 'ai'">
                                            <NeuralPrint :text="msg.content" />
                                        </template>
                                        <template v-else>
                                            {{ msg.content }}
                                        </template>
                                        <div v-if="msg.source" class="mt-4 pt-3 border-t border-white/5 text-[8px] uppercase tracking-[0.3em] text-[#00f2ff] font-black flex items-center gap-2">
                                            <span class="w-1 h-1 bg-[#00f2ff] rounded-full"></span>
                                            Source :: {{ msg.source }}
                                        </div>
                                    </div>
                                </div>
                                <div v-if="isChatting" class="flex justify-start pl-8">
                                    <div class="bg-white/5 border border-white/10 px-5 py-4 rounded-3xl rounded-tl-none flex gap-1.5 items-center shadow-xl">
                                        <span class="w-1.5 h-1.5 bg-[#00f2ff] rounded-full animate-bounce animation-duration-[1s]"></span>
                                        <span class="w-1.5 h-1.5 bg-[#ff00ea] rounded-full animate-bounce animation-duration-[1s] [animation-delay:0.2s]"></span>
                                        <span class="w-1.5 h-1.5 bg-[#00f2ff] rounded-full animate-bounce animation-duration-[1s] [animation-delay:0.4s]"></span>
                                        <span class="text-[9px] font-black uppercase tracking-[0.2em] ml-2 text-slate-500">Neural Sync...</span>
                                    </div>
                                </div>
                            </div>

                            <!-- Input Area -->
                            <div class="p-8 bg-white/2 border-t border-white/5">
                                <div class="flex flex-col gap-4">
                                    <!-- Model Selector -->
                                    <div class="flex items-center gap-3">
                                        <select 
                                            v-model="selectedAIModel"
                                            class="bg-white/5 border border-white/10 rounded-xl px-3 py-1.5 text-[10px] font-black uppercase tracking-widest text-slate-400 outline-none hover:border-indigo-500/50 transition-all cursor-pointer"
                                        >
                                            <option v-for="m in availableModels" :key="m.id" :value="m.id" class="bg-zinc-900 text-slate-300">
                                                {{ m.name }}
                                            </option>
                                        </select>
                                        <div class="h-px flex-1 bg-white/5"></div>
                                    </div>

                                    <div class="relative flex items-center group">
                                        <input 
                                            v-model="chatQuery"
                                            @keyup.enter="askAI"
                                            type="text" 
                                            placeholder="Communicate with Oracle..."
                                            class="w-full bg-black/50 border border-white/5 rounded-2xl px-6 py-5 text-sm focus:ring-2 focus:ring-indigo-500/50 focus:border-indigo-500/50 outline-none transition-all pr-20 placeholder:text-slate-600 font-medium"
                                        />
                                    <button 
                                        @click="askAI"
                                        :disabled="isChatting || !chatQuery"
                                        class="absolute right-3 p-3 bg-indigo-600 rounded-xl hover:bg-indigo-500 disabled:opacity-20 disabled:grayscale transition-all shadow-lg active:scale-90 group-hover:shadow-indigo-500/20"
                                    >
                                        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor"><path d="M10.894 2.553a1 1 0 00-1.788 0l-7 14a1 1 0 001.169 1.409l5-1.429A1 1 0 009 15.571V11a1 1 0 112 0v4.571a1 1 0 00.725.962l5 1.428a1 1 0 001.17-1.408l-7-14z" /></svg>
                                    </button>
                                </div>
                                <p class="text-[8px] text-center text-slate-600 mt-4 uppercase tracking-[0.3em] font-black">Encrypted Neural Link :: TierLog Sovereign AI</p>
                            </div>
                        </div>
                        </div>
                        <div v-else class="flex-1 flex flex-col items-center justify-center bg-indigo-600/2 border border-dashed border-indigo-500/5 rounded-4xl opacity-20 transition-all hover:opacity-40">
                            <div class="relative w-20 h-20 mb-8">
                                <div class="absolute inset-0 border border-indigo-500/20 rounded-full animate-ping"></div>
                                <div class="absolute inset-0 flex items-center justify-center text-5xl grayscale">🤖</div>
                            </div>
                            <p class="text-[10px] font-black uppercase tracking-[0.4em] text-indigo-400">Oracle Standby</p>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </AppLayout>
</template>

<style scoped>
.custom-scrollbar::-webkit-scrollbar { width: 4px; }
.custom-scrollbar::-webkit-scrollbar-track { background: transparent; }
.custom-scrollbar::-webkit-scrollbar-thumb { background: rgba(255, 255, 255, 0.1); border-radius: 10px; }
.custom-scrollbar::-webkit-scrollbar-thumb:hover { background: rgba(255, 255, 255, 0.2); }

.fade-enter-active, .fade-leave-active { transition: opacity 0.5s ease; }
.fade-enter-from, .fade-leave-to { opacity: 0; }

@keyframes ping {
    75%, 100% { transform: scale(2); opacity: 0; }
}
@keyframes bounce {
    0%, 80%, 100% { transform: translateY(0); }
    40% { transform: translateY(-5px); }
}
</style>
