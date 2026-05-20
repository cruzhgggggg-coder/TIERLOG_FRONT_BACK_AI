<script setup>
import { Head, Link } from '@inertiajs/vue3';
import { login, register, dashboard } from '@/routes';
import { onMounted, ref } from 'vue';

defineProps({
    canLogin: Boolean,
    canRegister: Boolean,
});

const isLoaded = ref(false);
onMounted(() => {
    isLoaded.value = true;
});
</script>

<template>
    <Head title="TierLog - AI Powered Thesis Supervision" />

    <div class="relative min-h-screen bg-[#030305] text-white selection:bg-indigo-500/30 selection:text-indigo-200 overflow-hidden font-sans">
        <!-- Futuristic Grid Background -->
        <div class="absolute inset-0 pointer-events-none">
            <div class="absolute inset-0 bg-[linear-gradient(to_right,#80808012_1px,transparent_1px),linear-gradient(to_bottom,#80808012_1px,transparent_1px)] bg-size-[40px_40px] mask-[radial-gradient(ellipse_60%_50%_at_50%_0%,#000_70%,transparent_100%)]"></div>
            <div class="absolute top-0 left-1/2 -translate-x-1/2 w-[1000px] h-[600px] bg-indigo-600/20 blur-[120px] rounded-full opacity-50"></div>
            <div class="absolute bottom-0 right-0 w-[500px] h-[500px] bg-purple-600/10 blur-[100px] rounded-full opacity-30"></div>
        </div>

        <!-- Navigation -->
        <nav class="relative z-10 flex items-center justify-between px-10 py-8 max-w-7xl mx-auto">
            <div class="flex items-center gap-3 group cursor-pointer">
                <div class="w-12 h-12 bg-linear-to-br from-indigo-500 to-purple-600 rounded-2xl flex items-center justify-center shadow-2xl shadow-indigo-500/20 group-hover:scale-110 transition-transform duration-500">
                    <span class="font-black text-2xl">T</span>
                </div>
                <span class="text-2xl font-black tracking-tighter">TierLog<span class="text-indigo-500">.</span></span>
            </div>
            
            <div class="hidden md:flex items-center gap-8">
                <a href="#features" class="text-sm font-medium text-slate-400 hover:text-white transition-colors">Features</a>
                <a href="#" class="text-sm font-medium text-slate-400 hover:text-white transition-colors">Methodology</a>
                <div class="h-4 w-px bg-white/10"></div>
                <Link v-if="$page.props.auth.user" :href="dashboard()" class="px-6 py-2 rounded-xl bg-white/5 hover:bg-white/10 text-white text-sm font-bold transition-all border border-white/10">Dashboard</Link>
                <template v-else>
                    <Link :href="login()" class="text-sm font-bold text-slate-400 hover:text-white transition-colors">Sign In</Link>
                    <Link :href="register()" class="px-6 py-3 rounded-xl bg-indigo-600 hover:bg-indigo-500 text-white text-sm font-black transition-all shadow-xl shadow-indigo-500/20 active:scale-95">Get Started</Link>
                </template>
            </div>
        </nav>

        <!-- Hero Section -->
        <main class="relative z-10 max-w-7xl mx-auto px-8 pt-24 pb-40 text-center flex flex-col items-center">
            <div 
                class="inline-flex items-center gap-3 bg-white/5 border border-white/10 px-5 py-2 rounded-full mb-10 transition-all duration-1000 transform"
                :class="isLoaded ? 'translate-y-0 opacity-100' : 'translate-y-4 opacity-0'"
            >
                <span class="relative flex h-2 w-2">
                  <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-indigo-400 opacity-75"></span>
                  <span class="relative inline-flex rounded-full h-2 w-2 bg-indigo-500"></span>
                </span>
                <span class="text-[10px] font-black uppercase tracking-[0.2em] text-indigo-300">Guarded AI Integration Active</span>
            </div>

            <h1 
                class="text-7xl md:text-[120px] font-black tracking-tighter mb-10 leading-[0.85] transition-all duration-1000 delay-100 transform"
                :class="isLoaded ? 'translate-y-0 opacity-100' : 'translate-y-8 opacity-0'"
            >
                Master <span class="text-transparent bg-clip-text bg-linear-to-r from-indigo-400 via-purple-400 to-pink-400">Mind</span><br />
                Supervision.
            </h1>

            <p 
                class="text-xl md:text-2xl text-slate-400 max-w-3xl mb-14 leading-relaxed transition-all duration-1000 delay-200 transform"
                :class="isLoaded ? 'translate-y-0 opacity-100' : 'translate-y-8 opacity-0'"
            >
                An intelligent ecosystem that turns consultation transcripts into structured action plans. 
                Powered by <span class="text-white font-bold">Gemini 2.0 Flash</span> for native audio and draft paper analysis.
            </p>

            <div 
                class="flex flex-col sm:flex-row gap-6 transition-all duration-1000 delay-300 transform"
                :class="isLoaded ? 'translate-y-0 opacity-100' : 'translate-y-8 opacity-0'"
            >
                <Link :href="register()" class="group bg-white text-black px-12 py-5 rounded-4xl font-black text-xl shadow-2xl shadow-white/10 hover:bg-indigo-50 transition-all active:scale-95 flex items-center gap-3">
                    Unlock AI Access
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 group-hover:translate-x-1 transition-transform" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="3" d="M13 7l5 5m0 0l-5 5m5-5H6" /></svg>
                </Link>
                <a href="#features" class="bg-white/5 hover:bg-white/10 border border-white/10 text-white px-12 py-5 rounded-4xl font-black text-xl transition-all active:scale-95 backdrop-blur-sm">
                    Visual Demo
                </a>
            </div>

            <!-- Floating Status Cards -->
            <div class="mt-32 w-full max-w-6xl grid grid-cols-1 md:grid-cols-3 gap-6 relative animate-in">
                <div class="bg-white/5 border border-white/10 p-8 rounded-[40px] backdrop-blur-xl group hover:border-indigo-500/30 transition-all">
                    <div class="text-4xl mb-6">🎙️</div>
                    <h3 class="text-xl font-bold mb-3">Native Audio</h3>
                    <p class="text-sm text-slate-500 leading-relaxed">Direct analysis from your consultation MP3 files. No intermediaries, maximum accuracy.</p>
                </div>
                <div class="bg-linear-to-br from-indigo-600/20 to-purple-600/20 border border-indigo-500/30 p-8 rounded-[40px] backdrop-blur-xl relative overflow-hidden group">
                    <div class="absolute -right-4 -top-4 text-white/5 text-8xl font-black">AI</div>
                    <div class="text-4xl mb-6">🧠</div>
                    <h3 class="text-xl font-bold mb-3">Guarded Oracle</h3>
                    <p class="text-sm text-slate-300 leading-relaxed font-medium">AI Assistant securely locked to your lecturer's feedback. Anti-hallucination.</p>
                </div>
                <div class="bg-white/5 border border-white/10 p-8 rounded-[40px] backdrop-blur-xl group hover:border-purple-500/30 transition-all">
                    <div class="text-4xl mb-6">📄</div>
                    <h3 class="text-xl font-bold mb-3">Structure Sync</h3>
                    <p class="text-sm text-slate-500 leading-relaxed">Automatic synchronization between spoken revisions and written paper drafts.</p>
                </div>
            </div>
        </main>

        <!-- Footer -->
        <footer class="relative z-10 border-t border-white/5 py-12">
            <div class="max-w-7xl mx-auto px-8 flex justify-between items-center text-slate-500 text-xs font-bold uppercase tracking-widest">
                <p>© 2026 TierLog Master Mind</p>
                <div class="flex gap-8">
                    <a href="#" class="hover:text-white transition-colors">Privacy</a>
                    <a href="#" class="hover:text-white transition-colors">Terms</a>
                    <a href="#" class="hover:text-white transition-colors">GitHub</a>
                </div>
            </div>
        </footer>
    </div>
</template>

<style scoped>
.animate-in {
    animation: fade-up 1s ease-out forwards;
    animation-delay: 0.5s;
    opacity: 0;
}

@keyframes fade-up {
    from { transform: translateY(2rem); opacity: 0; }
    to { transform: translateY(0); opacity: 1; }
}

@font-face {
    font-family: 'Inter';
    src: url('https://fonts.googleapis.com/css2?family=Inter:wght@400;700;900&display=swap');
}
</style>
