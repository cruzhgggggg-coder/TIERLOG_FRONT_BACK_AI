<script setup>
import { ref, onMounted, watch, computed } from 'vue';

const props = defineProps({
    text: {
        type: String,
        required: true
    },
    speed: {
        type: Number,
        default: 15
    }
});

const displayedText = ref('');
let currentIndex = 0;
let timeoutId = null;

const streamText = () => {
    if (currentIndex < props.text.length) {
        // Handle rapid multi-char bursts for a more "organic" feel
        const burstSize = Math.random() > 0.85 ? 3 : 1;
        displayedText.value += props.text.substring(currentIndex, currentIndex + burstSize);
        currentIndex += burstSize;
        timeoutId = setTimeout(streamText, props.speed);
    }
};

const copyPatch = (content) => {
    navigator.clipboard.writeText(content);
    // You could emit an event here to show a toast in the parent
};

onMounted(() => {
    streamText();
});

watch(() => props.text, (newVal) => {
    clearTimeout(timeoutId);
    displayedText.value = '';
    currentIndex = 0;
    streamText();
});

const segments = computed(() => {
    // Regex to find [PATCH]...[/PATCH] blocks
    const patchRegex = /\[PATCH\]([\s\S]*?)\[\/PATCH\]/g;
    const parts = [];
    let lastIndex = 0;
    let match;

    while ((match = patchRegex.exec(displayedText.value)) !== null) {
        // Text before the patch
        if (match.index > lastIndex) {
            parts.push({
                type: 'text',
                content: displayedText.value.substring(lastIndex, match.index)
            });
        }
        // The patch itself
        parts.push({
            type: 'patch',
            content: match[1].trim()
        });
        lastIndex = patchRegex.lastIndex;
    }

    // Remaining text
    if (lastIndex < displayedText.value.length) {
        parts.push({
            type: 'text',
            content: displayedText.value.substring(lastIndex)
        });
    }

    return parts;
});

const formatText = (text) => {
    return text
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/\*\*(.*?)\*\*/g, '<span class="synaptic-highlight">$1</span>');
};
</script>

<template>
    <div class="neural-print leading-relaxed">
        <template v-for="(segment, i) in segments" :key="i">
            <span v-if="segment.type === 'text'" v-html="formatText(segment.content)"></span>
            
            <div v-else-if="segment.type === 'patch'" class="patch-block group/patch">
                <pre class="text-[11px] leading-relaxed text-[#00f2ff]/90 overflow-x-auto custom-scrollbar">{{ segment.content }}</pre>
                
                <button 
                    @click="copyPatch(segment.content)"
                    class="absolute top-2 right-2 p-1.5 bg-[#00f2ff]/10 hover:bg-[#00f2ff]/20 text-[#00f2ff] rounded-lg border border-[#00f2ff]/20 opacity-0 group-hover/patch:opacity-100 transition-all duration-300"
                    title="Copy Patch to Clipboard"
                >
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5" viewBox="0 0 20 20" fill="currentColor">
                        <path d="M8 3a1 1 0 011-1h2a1 1 0 110 2H9a1 1 0 01-1-1z" />
                        <path d="M6 3a2 2 0 00-2 2v11a2 2 0 002 2h8a2 2 0 002-2V5a2 2 0 00-2-2 3 3 0 01-3 3H9a3 3 0 01-3-3z" />
                    </svg>
                </button>
                
                <div class="h-px bg-linear-to-r from-transparent via-indigo-500/50 to-transparent"></div>
            </div>
        </template>
        
        <span v-if="currentIndex < text.length" class="animate-pulse inline-block w-1.5 h-4 bg-cyan-400 ml-1 align-middle shadow-[0_0_10px_#22d3ee]"></span>
    </div>
</template>

<style scoped>
.neural-print {
    text-shadow: 0 0 1px rgba(0, 242, 255, 0.05);
}

.custom-scrollbar::-webkit-scrollbar { height: 3px; }
.custom-scrollbar::-webkit-scrollbar-track { background: transparent; }
.custom-scrollbar::-webkit-scrollbar-thumb { background: rgba(0, 242, 255, 0.1); border-radius: 10px; }

:deep(.synaptic-highlight) {
    color: #00f2ff;
    text-shadow: 0 0 8px rgba(0, 242, 255, 0.4);
    font-weight: 600;
}
</style>
