<script setup lang="ts">
import { ref, onMounted, onUnmounted, type CSSProperties } from 'vue';

const props = defineProps<{
  id: string;
  title: string;
  x: number;
  y: number;
  w: number;
  h: number;
  zIndex: number;
  isMaximized: boolean;
}>();

const emit = defineEmits<{
  (e: 'update:position', payload: { x: number; y: number }): void;
  (e: 'update:size', payload: { w: number; h: number }): void;
  (e: 'close'): void;
  (e: 'maximize'): void;
  (e: 'focus'): void;
}>();

const windowRef = ref<HTMLDivElement | null>(null);

// ── Dragging logic ──
const isDragging = ref(false);
let startX = 0;
let startY = 0;
let startPageX = 0;
let startPageY = 0;

function onHeaderPointerDown(e: PointerEvent) {
  // Prevent dragging if clicked on window buttons
  if ((e.target as HTMLElement).closest('.window-btn')) return;
  
  emit('focus');
  if (props.isMaximized) return;

  isDragging.value = true;
  startX = props.x;
  startY = props.y;
  startPageX = e.pageX;
  startPageY = e.pageY;

  (e.target as HTMLElement).setPointerCapture(e.pointerId);
}

function onHeaderPointerMove(e: PointerEvent) {
  if (!isDragging.value) return;

  const dx = e.pageX - startPageX;
  const dy = e.pageY - startPageY;

  // Simple bounding to positive coordinates
  const newX = Math.max(0, startX + dx);
  const newY = Math.max(0, startY + dy);

  emit('update:position', { x: newX, y: newY });
}

function onHeaderPointerUp(e: PointerEvent) {
  if (isDragging.value) {
    isDragging.value = false;
    (e.target as HTMLElement).releasePointerCapture(e.pointerId);
  }
}

// ── Resizing logic ──
const isResizing = ref(false);
let startW = 0;
let startH = 0;
let startResizePageX = 0;
let startResizePageY = 0;

function onResizePointerDown(e: PointerEvent) {
  e.preventDefault();
  e.stopPropagation();
  emit('focus');
  if (props.isMaximized) return;

  isResizing.value = true;
  startW = props.w;
  startH = props.h;
  startResizePageX = e.pageX;
  startResizePageY = e.pageY;

  document.addEventListener('pointermove', onResizePointerMove);
  document.addEventListener('pointerup', onResizePointerUp);
}

function onResizePointerMove(e: PointerEvent) {
  if (!isResizing.value) return;

  const dx = e.pageX - startResizePageX;
  const dy = e.pageY - startResizePageY;

  const newW = Math.max(240, startW + dx);
  const newH = Math.max(160, startH + dy);

  emit('update:size', { w: newW, h: newH });
}

function onResizePointerUp() {
  if (isResizing.value) {
    isResizing.value = false;
    document.removeEventListener('pointermove', onResizePointerMove);
    document.removeEventListener('pointerup', onResizePointerUp);
  }
}

onUnmounted(() => {
  document.removeEventListener('pointermove', onResizePointerMove);
  document.removeEventListener('pointerup', onResizePointerUp);
});
</script>

<template>
  <div
    ref="windowRef"
    :class="[
      'absolute flex flex-col rounded-xl border border-white/[0.04] bg-[#141415]/95 shadow-2xl backdrop-blur-xl transition-all duration-75 select-none overflow-hidden',
      isDragging ? 'opacity-85 scale-[0.99] border-amber-500/20' : '',
      isMaximized ? '!inset-0 !w-full !h-full !translate-x-0 !translate-y-0 rounded-none border-none z-50' : '',
    ]"
    :style="{
      zIndex: zIndex,
      transform: isMaximized ? 'none' : `translate3d(${x}px, ${y}px, 0)`,
      width: isMaximized ? '100%' : `${w}px`,
      height: isMaximized ? '100%' : `${h}px`,
    }"
    @mousedown="emit('focus')"
  >
    <!-- Window Header -->
    <div
      class="flex h-11 shrink-0 cursor-move items-center justify-between border-b border-white/[0.04] bg-stone-900/40 px-4 active:cursor-grabbing"
      @pointerdown="onHeaderPointerDown"
      @pointermove="onHeaderPointerMove"
      @pointerup="onHeaderPointerUp"
    >
      <div class="flex items-center gap-2">
        <span class="h-2 w-2 rounded-full" :class="isMaximized ? 'bg-indigo-500' : 'bg-amber-500'" />
        <span class="text-xs font-semibold text-stone-200 font-[family-name:var(--font-display)] tracking-wide">{{ title }}</span>
      </div>
      
      <div class="flex items-center gap-1">
        <!-- Maximize Button -->
        <button
          class="window-btn flex h-6 w-6 items-center justify-center rounded-md text-stone-500 hover:bg-white/5 hover:text-stone-300 transition-colors"
          title="Maximize"
          @click.stop="emit('maximize')"
        >
          <svg v-if="isMaximized" xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
            <polyline points="4 14 10 14 10 20" />
            <polyline points="20 10 14 10 14 4" />
            <line x1="14" y1="10" x2="21" y2="3" />
            <line x1="10" y1="14" x2="3" y2="21" />
          </svg>
          <svg v-else xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
            <polyline points="15 3 21 3 21 9" />
            <polyline points="9 21 3 21 3 15" />
            <line x1="21" y1="3" x2="14" y2="10" />
            <line x1="3" y1="21" x2="10" y2="14" />
          </svg>
        </button>

        <!-- Close Button -->
        <button
          class="window-btn flex h-6 w-6 items-center justify-center rounded-md text-stone-500 hover:bg-red-500/10 hover:text-red-400 transition-colors"
          title="Close"
          @click.stop="emit('close')"
        >
          <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
            <line x1="18" y1="6" x2="6" y2="18" />
            <line x1="6" y1="6" x2="18" y2="18" />
          </svg>
        </button>
      </div>
    </div>

    <!-- Window Content Area -->
    <div class="flex-1 min-h-0 overflow-y-auto select-text scroll-area relative bg-[#0d0d0e]/60">
      <slot />
    </div>

    <!-- Resize Grip -->
    <div
      v-if="!isMaximized"
      class="absolute bottom-0 right-0 h-4 w-4 cursor-se-resize flex items-end justify-end p-0.5 pointer-events-auto"
      @pointerdown="onResizePointerDown"
    >
      <svg class="text-stone-600 opacity-40 hover:opacity-80 transition-opacity" width="8" height="8" viewBox="0 0 8 8" fill="none" stroke="currentColor" stroke-width="1.5">
        <line x1="6" y1="2" x2="2" y2="6" />
        <line x1="6" y1="4" x2="4" y2="6" />
        <line x1="6" y1="6" x2="6" y2="6" />
      </svg>
    </div>
  </div>
</template>

<style scoped>
.scroll-area {
  scrollbar-width: thin;
  scrollbar-color: rgba(255, 255, 255, 0.05) transparent;
}
.scroll-area::-webkit-scrollbar {
  width: 5px;
  height: 5px;
}
.scroll-area::-webkit-scrollbar-track {
  background: transparent;
}
.scroll-area::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.05);
  border-radius: 9999px;
}
.scroll-area::-webkit-scrollbar-thumb:hover {
  background: rgba(255, 255, 255, 0.12);
}
</style>
