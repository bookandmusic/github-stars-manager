<template>
  <div v-if="syncing" class="px-4 md:px-6 pb-2">
    <div class="glass-card p-3">
      <div class="flex justify-between text-white text-sm mb-1">
        <span class="truncate">{{ syncMessage }}</span>
        <span class="ml-2">{{ syncProgress }}%</span>
      </div>
      <div class="progress-bar">
        <div class="progress-fill" :style="{ width: syncProgress + '%' }"></div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  syncing: boolean,
  syncMessage: string,
  syncProgress: number
}>()
</script>

<style scoped>
.progress-bar {
  height: 6px;
  background-color: rgba(255, 255, 255, 0.2);
  border-radius: 3px;
  overflow: hidden;
  margin-top: 4px;
}

.progress-fill {
  height: 100%;
  background: linear-gradient(90deg, #4F46E5, #7C3AED);
  border-radius: 3px;
  transition: width 0.3s ease;
}
</style>