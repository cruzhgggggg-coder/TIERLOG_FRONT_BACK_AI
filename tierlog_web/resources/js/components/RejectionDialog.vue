<script setup lang="ts">
import { ref, watch } from 'vue';

const props = defineProps<{
  open: boolean;
  title: string;
  message: string;
  confirmLabel?: string;
  cancelLabel?: string;
}>();

const emit = defineEmits<{
  confirm: [comment: string];
  cancel: [];
}>();

const comment = ref('');
const errorMsg = ref('');

watch(() => props.open, (newVal) => {
  if (newVal) {
    comment.value = '';
    errorMsg.value = '';
  }
});

function handleConfirm() {
  const trimmed = comment.value.trim();
  if (!trimmed) {
    errorMsg.value = 'Please provide an explanation for the rejection.';
    return;
  }
  emit('confirm', trimmed);
}
</script>

<template>
  <Teleport to="body">
    <Transition
      enter-active-class="transition duration-200 ease-out"
      enter-from-class="opacity-0"
      enter-to-class="opacity-100"
      leave-active-class="transition duration-150 ease-in"
      leave-from-class="opacity-100"
      leave-to-class="opacity-0"
    >
      <div
        v-if="open"
        class="fixed inset-0 z-[200] flex items-center justify-center p-4"
      >
        <!-- Backdrop -->
        <div
          class="absolute inset-0 bg-black/60 backdrop-blur-sm"
          @click="emit('cancel')"
        />

        <!-- Dialog -->
        <Transition
          enter-active-class="transition duration-200 ease-out"
          enter-from-class="scale-95 opacity-0"
          enter-to-class="scale-100 opacity-100"
          leave-active-class="transition duration-150 ease-in"
          leave-from-class="scale-100 opacity-100"
          leave-to-class="scale-95 opacity-0"
        >
          <div
            v-if="open"
            class="relative w-full max-w-md rounded-2xl border border-white/[0.06] bg-[#1c1c1e] p-6 shadow-2xl"
          >
            <h3 class="text-base font-semibold text-stone-100 font-[family-name:var(--font-display)]">
              {{ title }}
            </h3>
            <p class="mt-2 text-sm leading-relaxed text-stone-400">
              {{ message }}
            </p>

            <div class="mt-4">
              <label class="block text-xs font-semibold uppercase tracking-wider text-stone-400 mb-1.5">
                Explanation / Reason
              </label>
              <textarea
                v-model="comment"
                rows="4"
                class="w-full rounded-xl border border-white/[0.08] bg-white/[0.02] p-3 text-sm text-stone-200 placeholder-stone-600 outline-none transition duration-200 focus:border-red-500/50 focus:bg-white/[0.04] focus:ring-1 focus:ring-red-500/50"
                placeholder="Type why you are rejecting this revision..."
                @input="errorMsg = ''"
              ></textarea>
              <p v-if="errorMsg" class="mt-1.5 text-xs text-red-400">
                {{ errorMsg }}
              </p>
            </div>

            <div class="mt-5 flex items-center justify-end gap-2.5">
              <button
                class="rounded-lg px-4 py-2 text-sm font-medium text-stone-400 transition-colors hover:bg-white/[0.04] hover:text-stone-200"
                @click="emit('cancel')"
              >
                {{ cancelLabel ?? 'Cancel' }}
              </button>
              <button
                class="rounded-lg px-4 py-2 text-sm font-medium transition-colors bg-red-500/15 text-red-400 hover:bg-red-500/25"
                @click="handleConfirm"
              >
                {{ confirmLabel ?? 'Reject Revision' }}
              </button>
            </div>
          </div>
        </Transition>
      </div>
    </Transition>
  </Teleport>
</template>
