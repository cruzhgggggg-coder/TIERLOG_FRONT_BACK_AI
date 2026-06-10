<script setup lang="ts">
</script>

<template>
  <div class="relative min-h-screen bg-[#09090b] overflow-hidden">
    <div class="pointer-events-none absolute inset-0">
      <div class="absolute -top-48 -left-48 h-[600px] w-[600px] rounded-full bg-amber-500/[0.04] blur-[140px]" />
      <div class="absolute top-1/3 -right-32 h-[450px] w-[450px] rounded-full bg-orange-600/[0.03] blur-[120px]" />
      <div class="absolute -bottom-32 left-1/4 h-[350px] w-[350px] rounded-full bg-amber-600/[0.025] blur-[100px]" />
    </div>
    <div class="relative mx-auto max-w-[1600px] px-4 py-8 sm:px-6 lg:px-8 xl:px-12">
      <slot />
    </div>
  </div>
</template>
