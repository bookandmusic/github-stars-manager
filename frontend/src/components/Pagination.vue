<template>
  <div class="py-2 md:py-3 px-4 md:px-6">
    <div class="flex flex-col sm:flex-row justify-between items-center gap-2">
      <div class="mb-2 sm:mb-0 hidden md:flex items-center">
        <label class="mr-2 text-white">每页显示:</label>
        <select v-model="localPerPage" @change="onUpdatePage"
          class="glass-button text-white px-3 py-1 rounded-lg">
          <option value="9" class="bg-gray-800 text-white">9</option>
          <option value="18" class="bg-gray-800 text-white">18</option>
          <option value="36" class="bg-gray-800 text-white">36</option>
        </select>
      </div>

      <div class="flex flex-wrap justify-center gap-1 w-full sm:w-auto items-center">
        <button @click="goToFirstPage" :disabled="currentPage === 1"
          class="glass-button text-white px-2 py-1 rounded-lg text-sm disabled:opacity-50">
          首页
        </button>
        <button @click="goToPrevPage" :disabled="currentPage === 1"
          class="glass-button text-white px-2 py-1 rounded-lg text-sm disabled:opacity-50">
          上一页
        </button>
        <span class="glass-button text-white px-2 py-1 text-sm self-center">
          第 {{ currentPage }} 页，共 {{ totalPages }} 页
        </span>
        <button @click="goToNextPage" :disabled="currentPage === totalPages"
          class="glass-button text-white px-2 py-1 rounded-lg text-sm disabled:opacity-50">
          下一页
        </button>
        <button @click="goToLastPage" :disabled="currentPage === totalPages"
          class="glass-button text-white px-2 py-1 rounded-lg text-sm disabled:opacity-50">
          末页
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';

const props = defineProps<{
  currentPage: number,
  perPage: number,
  totalPages: number
}>()

const emit = defineEmits<{
  (e: 'update:currentPage', value: number): void,
  (e: 'update:perPage', value: number): void
}>()

const localPerPage = computed({
  get: () => props.perPage,
  set: (value) => emit('update:perPage', Number(value))
})

function onUpdatePage() {
  // 重置到第一页
  emit('update:currentPage', 1)
}

function goToFirstPage() {
  emit('update:currentPage', 1)
}

function goToPrevPage() {
  if (props.currentPage > 1) {
    emit('update:currentPage', props.currentPage - 1)
  }
}

function goToNextPage() {
  if (props.currentPage < props.totalPages) {
    emit('update:currentPage', props.currentPage + 1)
  }
}

function goToLastPage() {
  emit('update:currentPage', props.totalPages)
}
</script>