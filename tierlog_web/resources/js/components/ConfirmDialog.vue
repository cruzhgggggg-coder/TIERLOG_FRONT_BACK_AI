<script setup lang="ts">
defineProps<{
  open: boolean;
  title: string;
  message: string;
  confirmLabel?: string;
  cancelLabel?: string;
  danger?: boolean;
}>();

const emit = defineEmits<{
  confirm: [];
  cancel: [];
}>();
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
            <div class="mt-5 flex items-center justify-end gap-2.5">
              <button
                class="rounded-lg px-4 py-2 text-sm font-medium text-stone-400 transition-colors hover:bg-white/[0.04] hover:text-stone-200"
                @click="emit('cancel')"
              >
                {{ cancelLabel ?? 'Cancel' }}
              </button>
              <button
                :class="[
                  'rounded-lg px-4 py-2 text-sm font-medium transition-colors',
                  danger
                    ? 'bg-red-500/15 text-red-400 hover:bg-red-500/25'
                    : 'bg-amber-500/15 text-amber-400 hover:bg-amber-500/25',
                ]"
                @click="emit('confirm')"
              >
                {{ confirmLabel ?? 'Confirm' }}
              </button>
            </div>
          </div>
        </Transition>
      </div>
    </Transition>
  </Teleport>
</template>
