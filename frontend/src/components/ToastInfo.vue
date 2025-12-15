<template>
  <div class="fixed top-5 right-5 z-1000 w-75">
    <div
      v-for="(toast, index) in toasts"
      :key="index"
      :class="[
        'flex items-center p-3 mb-2.5 rounded-lg shadow-md transition-all duration-300 transform',
        toast.show ? 'translate-x-0 opacity-100' : 'translate-x-25 opacity-0',
        getToastClass(toast.type)
      ]"
    >
      <span class="mr-2.5 text-lg">{{ getToastIcon(toast.type) }}</span>
      <span class="flex-1 text-sm">{{ toast.message }}</span>
      <button class="bg-transparent border-none text-white text-lg cursor-pointer p-0 ml-2.5" @click="removeToast(index)">
        &times;
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';

interface Toast {
  message: string;
  type: 'success' | 'error' | 'info';
  show: boolean;
}

const toasts = ref<Toast[]>([]);

function showToast(message: string, type: 'success' | 'error' | 'info' = 'info', duration = 3000) {
  const toast: Toast = {
    message,
    type,
    show: true
  };

  toasts.value.push(toast);

  setTimeout(() => {
    const index = toasts.value.indexOf(toast);
    if (index !== -1) {
      removeToast(index);
    }
  }, duration);
}

function removeToast(index: number) {
  if (toasts.value[index]) {
    toasts.value[index].show = false;
    setTimeout(() => {
      toasts.value.splice(index, 1);
    }, 300);
  }
}

function getToastIcon(type: string) {
  switch (type) {
    case 'success': return '✓';
    case 'error': return '✕';
    case 'info': return 'ℹ';
    default: return '';
  }
}

function getToastClass(type: string) {
  switch (type) {
    case 'success': 
      return 'bg-green-500/90 color-white';
    case 'error': 
      return 'bg-red-500/90 color-white';
    case 'info': 
      return 'bg-blue-500/90 color-white';
    default:
      return 'bg-gray-500/90 color-white';
  }
}

defineExpose({
  showToast
});
</script>