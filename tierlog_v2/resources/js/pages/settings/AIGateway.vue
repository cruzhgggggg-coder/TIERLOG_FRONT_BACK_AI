<script setup lang="ts">
import { Form, Head, router, usePage } from '@inertiajs/vue3';
import { computed, ref, onMounted, watch } from 'vue';
import axios from 'axios';
import AIGatewayController from '@/actions/App/Http/Controllers/Settings/AIGatewayController';
import Heading from '@/components/Heading.vue';
import InputError from '@/components/InputError.vue';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Brain, Key, Zap, Shield, Sparkles } from 'lucide-vue-next';

type Props = {
    status?: string;
};

defineProps<Props>();

const page = usePage();
const user = computed(() => page.props.auth.user as any);
const isGatewayActive = computed(() => true);

const universalKey = ref('');
const detectedProvider = computed(() => {
    const key = universalKey.value.trim();
    if (key.startsWith('sk-ant-')) return 'anthropic';
    if (key.startsWith('sk-')) return 'openai';
    if (key.startsWith('AIza')) return 'gemini';
    if (key.startsWith('nvapi-')) return 'nvidia';
    return null;
});

const providerModels = ref({
    openai: ['GPT-4o', 'GPT-4o Mini', 'o1 Preview', 'o1 Mini'],
    anthropic: ['Claude 3.5 Sonnet', 'Claude 3 Opus', 'Claude 3 Haiku'],
    gemini: ['Gemini 2.0 Flash', 'Gemini 1.5 Pro', 'Gemini 1.5 Flash'],
    nvidia: ['meta/llama-3.1-70b-instruct', 'meta/llama-3.1-405b-instruct', 'nvidia/nemotron-4-340b-instruct', 'bytedance/seed-oss-36b-instruct'],
});

const isFetchingModels = ref(false);

const fetchNvidiaModels = async (apiKey: string) => {
    try {
        isFetchingModels.value = true;
        const response = await axios.get(`http://localhost:8080/api/ai/models?provider=nvidia&api_key=${apiKey}`);
        if (response.data && response.data.models) {
            providerModels.value.nvidia = response.data.models;
        }
    } catch (e) {
        console.error('Failed to fetch nvidia models:', e);
    } finally {
        isFetchingModels.value = false;
    }
};

watch(universalKey, (newKey) => {
    if (detectedProvider.value === 'nvidia' && newKey.trim() !== '') {
        fetchNvidiaModels(newKey.trim());
    }
});

onMounted(() => {
    if (user.value.nvidia_key) {
        fetchNvidiaModels(user.value.nvidia_key);
    }
});

const handleSaveUniversal = () => {
    const provider = detectedProvider.value;
    if (!provider) return;

    // Send only the relevant key update while keeping others
    const data: any = {
        preferred_model: user.value.preferred_model || 'default'
    };
    
    if (provider === 'openai') data.openai_key = universalKey.value;
    if (provider === 'anthropic') data.anthropic_key = universalKey.value;
    if (provider === 'gemini') data.gemini_key = universalKey.value;
    if (provider === 'nvidia') data.nvidia_key = universalKey.value;
    
    router.patch(AIGatewayController.update.url(), data, {
        preserveScroll: true,
        onSuccess: () => {
            universalKey.value = '';
        }
    });
};

const handleModelChange = (e: any) => {
    router.patch(AIGatewayController.update.url(), {
        preferred_model: e.target.value
    }, { preserveScroll: true });
};
</script>

<template>
    <Head title="AI Gateway Settings" />

    <div class="flex flex-col space-y-8">
        <div>
            <Heading
                variant="small"
                title="AI Gateway"
                description="Turn TierLog into your personal AI Gateway by connecting your favorite AI models."
            />
        </div>

        <!-- Information Card: Explaining the simple Plug-and-Play system -->
        <div class="bg-indigo-600/10 border border-indigo-500/20 rounded-3xl p-6 flex items-start gap-4 animate-in fade-in slide-in-from-top-4 duration-500">
            <div class="p-3 bg-indigo-600 rounded-2xl text-white shadow-lg shadow-indigo-500/20">
                <Zap class="w-6 h-6" />
            </div>
            <div>
                <h3 class="text-lg font-bold text-white">Plug-and-Play AI Gateway</h3>
                <p class="text-sm text-slate-400 mt-1">
                    Enter your API Key below. The system will automatically detect the active key and unlock new model options in your consultation workspace.
                </p>
            </div>
        </div>

        <!-- Universal API Key Input -->
        <div v-if="isGatewayActive" class="space-y-6 animate-in fade-in slide-in-from-bottom-4 duration-700">
            <div class="bg-zinc-900 border border-white/5 rounded-3xl p-8 shadow-2xl relative overflow-hidden">
                <div class="absolute top-0 right-0 p-4 opacity-10">
                    <Key class="w-24 h-24 rotate-12" />
                </div>

                <div class="relative z-10 space-y-6">
                    <div>
                        <h3 class="text-xl font-black text-white tracking-tight flex items-center gap-2">
                            <Sparkles class="w-5 h-5 text-indigo-500" />
                            Universal API Connector
                        </h3>
                        <p class="text-sm text-slate-400 mt-2">
                            Paste your API Key below. The system will detect the provider automatically.
                        </p>
                    </div>

                    <div class="relative group">
                        <div :class="[
                            'absolute -inset-1 bg-linear-to-r from-indigo-500 to-purple-600 rounded-2xl blur-sm opacity-20 transition duration-1000 group-hover:opacity-40 group-hover:duration-200',
                            detectedProvider ? 'opacity-50' : ''
                        ]"></div>
                        <input
                            v-model="universalKey"
                            type="text"
                            placeholder="Paste your API key here (sk-..., AIza...)"
                            class="relative w-full bg-black border border-white/10 rounded-xl px-6 py-5 text-lg font-mono text-white placeholder:text-slate-600 outline-none focus:border-indigo-500/50 transition-all pr-32"
                        />
                        
                        <div v-if="detectedProvider" class="absolute right-3 top-1/2 -translate-y-1/2 flex items-center gap-2">
                            <Badge variant="default" class="bg-indigo-600 animate-in zoom-in duration-300">
                                {{ detectedProvider.toUpperCase() }} Detected
                            </Badge>
                        </div>
                    </div>

                    <!-- Dynamic Model Preview -->
                    <div v-if="detectedProvider" class="bg-white/5 rounded-2xl p-6 border border-white/5 animate-in slide-in-from-top-2 duration-500">
                        <div class="flex items-center gap-4">
                            <div :class="[
                                'p-3 rounded-xl text-white shadow-lg',
                                detectedProvider === 'openai' ? 'bg-green-600 shadow-green-500/20' : 
                                detectedProvider === 'anthropic' ? 'bg-orange-600 shadow-orange-500/20' : 
                                detectedProvider === 'gemini' ? 'bg-blue-600 shadow-blue-500/20' : 'bg-emerald-600 shadow-emerald-500/20'
                            ]">
                                <Brain class="w-5 h-5" />
                            </div>
                            <div>
                                <h4 class="text-sm font-bold text-white">Unlocking Model Access:</h4>
                                <div class="flex flex-wrap gap-2 mt-2">
                                    <span v-for="m in providerModels[detectedProvider]" :key="m" class="px-2 py-1 bg-white/10 rounded-md text-[10px] font-black tracking-widest text-slate-400">
                                        {{ m }}
                                    </span>
                                    <span v-if="isFetchingModels" class="px-2 py-1 bg-white/10 rounded-md text-[10px] font-black text-indigo-400 animate-pulse">
                                        Loading Models from NVIDIA...
                                    </span>
                                </div>
                            </div>
                        </div>
                    </div>

                    <div class="flex justify-end pt-2">
                        <Button 
                            @click="handleSaveUniversal" 
                            :disabled="!detectedProvider"
                            class="bg-indigo-600 hover:bg-indigo-700 text-white px-8 h-12 rounded-xl font-bold shadow-xl shadow-indigo-500/20 disabled:opacity-20 transition-all active:scale-95"
                        >
                            <Key class="w-4 h-4 mr-2" />
                            Save & Connect
                        </Button>
                    </div>
                </div>
            </div>

            <!-- Active Keys Overview -->
            <div class="grid gap-4 sm:grid-cols-4">
                <div v-for="(val, key) in { openai: 'OpenAI', anthropic: 'Anthropic', gemini: 'Gemini', nvidia: 'NVIDIA' }" :key="key" 
                    :class="[
                        'p-4 rounded-2xl border transition-all duration-500',
                        user[key + '_key'] ? 'bg-indigo-500/10 border-indigo-500/30' : 'bg-slate-900/50 border-white/5 opacity-40'
                    ]"
                >
                    <div class="flex items-center justify-between mb-2">
                        <span class="text-[10px] font-black uppercase tracking-widest text-slate-400">{{ val }}</span>
                        <div v-if="user[key + '_key']" class="w-2 h-2 rounded-full bg-indigo-500 shadow-[0_0_8px_#6366f1] animate-pulse"></div>
                    </div>
                    <p class="text-xs text-white truncate font-mono opacity-50">
                        {{ user[key + '_key'] ? '••••••••••••' + user[key + '_key'].slice(-4) : 'Not Connected' }}
                    </p>
                </div>
            </div>

            <!-- Preferred Model Form -->
            <Form
                v-bind="AIGatewayController.update.form()"
                class="space-y-6"
                v-slot="{ errors }"
            >
                <div class="grid gap-2 sm:col-span-2">
                    <Label for="preferred_model" class="flex items-center gap-2">
                        <Brain class="w-4 h-4 text-indigo-500" />
                        Preferred Model
                    </Label>
                    <select
                        id="preferred_model"
                        name="preferred_model"
                        @change="handleModelChange"
                        class="flex h-12 w-full rounded-xl border border-white/10 bg-zinc-950 px-4 py-2 text-sm text-white focus:ring-2 focus:ring-indigo-500 transition-all appearance-none cursor-pointer outline-none"
                    >
                        <option v-if="!user.openai_key && !user.gemini_key && !user.anthropic_key && !user.nvidia_key" value="none">No AI Models Connected</option>
                        
                        <optgroup label="OpenAI" v-if="user.openai_key">
                            <option value="openai:gpt-4o">OpenAI: GPT-4o</option>
                            <option value="openai:gpt-4o-mini">OpenAI: GPT-4o Mini</option>
                            <option value="openai:o1-preview">OpenAI: o1 Preview</option>
                            <option value="openai:o1-mini">OpenAI: o1 Mini</option>
                        </optgroup>

                        <optgroup label="Google Gemini" v-if="user.gemini_key">
                            <option value="gemini:gemini-2.0-flash">Gemini: 2.0 Flash</option>
                            <option value="gemini:gemini-1.5-pro">Gemini: 1.5 Pro</option>
                            <option value="gemini:gemini-1.5-flash">Gemini: 1.5 Flash</option>
                        </optgroup>

                        <optgroup label="Anthropic Claude" v-if="user.anthropic_key">
                            <option value="anthropic:claude-3-5-sonnet-20240620">Claude: 3.5 Sonnet</option>
                            <option value="anthropic:claude-3-opus-20240229">Claude: 3 Opus</option>
                            <option value="anthropic:claude-3-haiku-20240307">Claude: 3 Haiku</option>
                        </optgroup>

                        <optgroup label="NVIDIA NIM" v-if="user.nvidia_key">
                            <option v-for="m in providerModels.nvidia" :key="m" :value="'nvidia:' + m">{{ m }}</option>
                        </optgroup>
                    </select>
                    <p class="text-xs text-slate-500">The selected model will be used as the primary assistant.</p>
                </div>
            </Form>
        </div>

        <!-- Information Card -->
        <Card class="bg-slate-50 dark:bg-slate-900/50 border-none">
            <CardHeader>
                <CardTitle class="text-sm flex items-center gap-2">
                    <Brain class="w-4 h-4 text-indigo-500" />
                    About AI Gateway
                </CardTitle>
            </CardHeader>
            <CardContent class="text-xs text-slate-600 dark:text-slate-400 space-y-2">
                <p>
                    AI Gateway allows you to bring your own AI "brain" into the TierLog infrastructure. 
                    Your data is still processed with TierLog's security standards, but using your own model capacity.
                </p>
                <ul class="list-disc pl-4 space-y-1">
                    <li>Your API keys are stored encrypted.</li>
                    <li>TierLog does not profit from your API token usage.</li>
                    <li>You are responsible for any costs incurred with your respective AI model providers.</li>
                </ul>
            </CardContent>
        </Card>
    </div>
</template>
