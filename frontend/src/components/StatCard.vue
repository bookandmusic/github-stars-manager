<template>
  <div class="glass-stat px-2 py-1 flex items-center justify-center gap-1 h-10 whitespace-nowrap flex-1 md:flex-none md:min-w-24">
    <slot name="icon"></slot>
    <span class="text-white text-sm">
      <slot></slot>
    </span>
  </div>
</template>

<script setup lang="ts">
</script>
