<template>
  <div v-if="visible" class="fixed inset-0 z-50 md:hidden">
    <div class="absolute inset-0 backdrop-blur-sm bg-white/10" @click="close"></div>
    <div
      class="absolute left-0 top-0 bottom-0 w-4/5 max-w-sm bg-gradient-to-b from-purple-600 to-indigo-700 shadow-2xl flex flex-col">
      <div class="px-4 pt-4 flex-shrink-0">
        <div class="flex justify-between items-center mb-4">
          <h2 class="text-white text-xl font-semibold">筛选</h2>
          <button @click="close" class="text-white text-2xl">&times;</button>
        </div>

        <!-- 移动端统计信息 -->
        <div class="mb-4">
          <div class="flex gap-2 h-10">
            <div
              class="glass-stat px-2 py-1 flex items-center justify-center gap-1 h-10 whitespace-nowrap flex-1 min-w-20">
              <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-white" fill="none" viewBox="0 0 24 24"
                stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                  d="M11.049 2.927c.3-.921 1.603-.921 1.902 0l1.519 4.674a1 1 0 00.95.69h4.915c.969 0 1.371 1.24.588 1.81l-3.976 2.888a1 1 0 00-.363 1.118l1.518 4.674c.3.922-.755 1.688-1.538 1.118l-3.976-2.888a1 1 0 00-1.176 0l-3.976 2.888c-.783.57-.38-1.81.588-1.81h4.914a1 1 0 00.951-.69l1.519-4.674z" />
              </svg>
              <span class="text-white text-sm">{{ stats.total_repos }}</span>
            </div>
            <div
              class="glass-stat px-2 py-1 flex items-center justify-center gap-1 h-10 whitespace-nowrap flex-1 min-w-20">
              <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-white" fill="none" viewBox="0 0 24 24"
                stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                  d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
              </svg>
              <span class="text-white text-sm">{{ stats.analyzed_repos }}</span>
            </div>
            <div
              class="glass-stat px-2 py-1 flex items-center justify-center gap-1 h-10 whitespace-nowrap flex-1 min-w-20">
              <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-white" fill="none" viewBox="0 0 24 24"
                stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                  d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              <span class="text-white text-xs">{{ formatSyncTime(stats.last_sync) }}</span>
            </div>
          </div>
        </div>

        <!-- 移动端搜索框 -->
        <div class="mb-4">
          <div class="glass-button flex items-center">
            <input type="text" placeholder="搜索仓库..." :value="searchQuery" @input="onSearchInput"
              class="flex-grow bg-transparent text-white px-3 py-2 text-sm placeholder-white/70 focus:outline-none">
            <button @click="close" class="text-white px-2">
              <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                <path fill-rule="evenodd"
                  d="M8 4a4 4 0 100 8 4 4 0 000-8zM2 8a6 6 0 1110.89 3.476l4.817 4.817a1 1 0 01-1.414 1.414l-4.816-4.816A6 6 0 012 8z"
                  clip-rule="evenodd" />
              </svg>
            </button>
          </div>
        </div>

        <!-- 移动端Tab切换 -->
        <div class="flex border-b border-white/20">
          <button @click="activeTab = 'categories'"
            :class="{ 'border-b-2 border-white text-white': activeTab === 'categories', 'text-white/70': activeTab !== 'categories' }"
            class="px-3 py-2 text-sm font-medium focus:outline-none">
            分类
          </button>
          <button @click="activeTab = 'tags'"
            :class="{ 'border-b-2 border-white text-white': activeTab === 'tags', 'text-white/70': activeTab !== 'tags' }"
            class="px-3 py-2 text-sm font-medium focus:outline-none">
            标签
          </button>
        </div>
      </div>

      <div class="flex-grow overflow-y-auto p-4">
        <div class="space-y-2">
          <!-- 分类筛选内容 -->
          <div v-show="activeTab === 'categories'" class="h-[calc(100vh-200px)] overflow-y-auto">
            <ul class="category-list space-y-0">
              <li v-for="(category, index) in allCategories" :key="index" class="tree-node">
                <button @click="setCategoryFilter(category.value)"
                  :class="{ 'bg-white/30 shadow-[0_0_10px_rgba(255,255,255,0.2)]': activeFilters.category === category.value }"
                  class="category-button w-full text-left px-3 py-1 rounded-lg text-white hover:bg-white/20 transition text-sm">
                  {{ category.label }}
                </button>
              </li>
            </ul>
          </div>

          <!-- 标签筛选内容 -->
          <div v-show="activeTab === 'tags'" class="h-[calc(100vh-200px)] overflow-y-auto">
            <div class="flex flex-wrap gap-2">
              <span v-for="tag in allTags" :key="tag" @click="toggleTagFilter(tag)"
                :class="{ 'bg-indigo-800 text-white': activeFilters.tags.includes(tag), 'bg-indigo-100 text-indigo-800': !activeFilters.tags.includes(tag) }"
                class="inline-block text-xs px-3 py-1 rounded-full cursor-pointer hover:bg-indigo-200 relative">
                {{ tag }}
                <span v-if="activeFilters.tags.includes(tag)" @click.stop="removeTagFilter(tag)"
                  class="absolute -top-1 -right-1 bg-red-500 text-white rounded-full w-4 h-4 flex items-center justify-center text-xs cursor-pointer">
                  &times;
                </span>
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  visible: boolean,
  searchQuery: string,
  activeTab: string,
  allCategories: any[],
  allTags: string[],
  activeFilters: any,
  stats: {
    total_repos: number,
    analyzed_repos: number,
    last_sync: string
  }
}>()

const emit = defineEmits<{
  (e: 'close'): void,
  (e: 'update:searchQuery', value: string): void,
  (e: 'update:activeTab', value: string): void,
  (e: 'setCategoryFilter', value: string | null): void,
  (e: 'toggleTagFilter', value: string): void,
  (e: 'removeTagFilter', value: string): void,
  (e: 'searchInput'): void
}>()

function close() {
  emit('close')
}

function onSearchInput(event: Event) {
  const target = event.target as HTMLInputElement
  emit('update:searchQuery', target.value)
  emit('searchInput')
}

function setCategoryFilter(value: string | null) {
  emit('setCategoryFilter', value)
  close()
}

function toggleTagFilter(tag: string) {
  emit('toggleTagFilter', tag)
  close()
}

function removeTagFilter(tag: string) {
  emit('removeTagFilter', tag)
}

function formatSyncTime(timeString: string) {
  if (!timeString || timeString === "暂无同步") {
    return "未同步"
  }

  // 尝试解析时间字符串
  const syncDate = new Date(timeString)
  if (isNaN(syncDate.getTime())) {
    return timeString
  }

  const now = new Date()
  const diffMs = now.getTime() - syncDate.getTime()
  const diffHours = diffMs / (1000 * 60 * 60)
  const diffDays = diffMs / (1000 * 60 * 60 * 24)

  // 一小时内显示分钟
  if (diffHours < 1) {
    const minutes = Math.floor(diffMs / (1000 * 60))
    return `${Math.max(1, minutes)}分钟前`
  }

  // 一天内显示小时
  if (diffDays < 1) {
    const hours = Math.floor(diffHours)
    return `${hours}小时前`
  }

  // 超出一天显示天数
  const days = Math.floor(diffDays)
  return `${days}天前`
}
</script>
