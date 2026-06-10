<script setup lang="ts">
withDefaults(
  defineProps<{
    label: string;
    modelValue?: string;
    type?: string;
    placeholder?: string;
    error?: string;
  }>(),
  {
    modelValue: '',
    type: 'text',
    placeholder: '',
  },
);

defineEmits<{
  'update:modelValue': [value: string];
}>();
</script>

<template>
  <div class="relative space-y-1.5">
    <label class="block text-[12px] font-medium text-stone-500 tracking-[0.5px] uppercase">
      {{ label }}
    </label>
    <input
      :type="type"
      :value="modelValue"
      :placeholder="placeholder"
      :class="[
        'w-full bg-transparent border rounded-xl px-4 py-3 text-[14px] text-stone-100 placeholder-stone-600 transition-all duration-200 focus:outline-none',
        error
          ? 'border-red-500/50 focus:border-red-500/70 focus:ring-1 focus:ring-red-500/20'
          : 'border-white/[0.08] focus:border-amber-500/50 focus:ring-1 focus:ring-amber-500/20',
      ]"
      @input="$emit('update:modelValue', ($event.target as HTMLInputElement).value)"
    />
    <p v-if="error" class="text-[12px] text-red-400">{{ error }}</p>
  </div>
</template>
