<template>
  <div class="hidden md:block md:w-1/3 lg:w-1/4 xl:w-1/5 glass-card p-3 md:p-4 flex-shrink-0">
    <!-- 搜索区域 -->
    <div class="glass-button flex items-center flex-shrink-0">
      <input type="text" placeholder="搜索仓库..." :value="searchQuery" @input="onSearchInput"
        class="flex-grow bg-transparent text-white px-3 py-2 text-sm placeholder-white/70 focus:outline-none">
      <button class="text-white px-2">
        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
          <path fill-rule="evenodd"
            d="M8 4a4 4 0 100 8 4 4 0 000-8zM2 8a6 6 0 1110.89 3.476l4.817 4.817a1 1 0 01-1.414 1.414l-4.816-4.816A6 6 0 012 8z"
            clip-rule="evenodd" />
        </svg>
      </button>
    </div>

    <!-- 筛选区域 -->
    <div class="flex flex-col flex-grow mt-2 h-[calc(100%-56px)]">
      <!-- Tab 切换 -->
      <div class="flex mb-4 border-b border-white/20 flex-shrink-0">
        <button @click="setActiveTab('categories')"
          :class="{'border-b-2 border-white text-white': activeTab === 'categories', 'text-white/70': activeTab !== 'categories'}"
          class="px-3 py-2 text-sm font-medium focus:outline-none">
          分类
        </button>
        <button @click="setActiveTab('tags')"
          :class="{'border-b-2 border-white text-white': activeTab === 'tags', 'text-white/70': activeTab !== 'tags'}"
          class="px-3 py-2 text-sm font-medium focus:outline-none">
          标签
        </button>
      </div>

      <!-- 分类筛选内容 -->
      <div v-show="activeTab === 'categories'" class="flex-grow overflow-y-auto">
        <ul class="category-list space-y-0 p-2">
          <li v-for="(category, index) in allCategories" :key="index" class="tree-node">
            <button @click="setCategoryFilter(category.value)"
              :class="{'bg-white/30 shadow-[0_0_10px_rgba(255,255,255,0.2)]': activeFilters.category === category.value}"
              class="category-button w-full text-left px-2 md:px-3 py-1 rounded-lg text-white hover:bg-white/20 transition text-sm">
              {{ category.label }}
            </button>
          </li>
        </ul>
      </div>

      <!-- 标签筛选内容 -->
      <div v-show="activeTab === 'tags'" class="flex-grow overflow-y-auto">
        <div class="p-2">
          <div class="flex flex-wrap gap-1 md:gap-2">
            <span v-for="tag in allTags" :key="tag" @click="toggleTagFilter(tag)"
              :class="{'bg-indigo-800 text-white': activeFilters.tags.includes(tag), 'bg-indigo-100 text-indigo-800': !activeFilters.tags.includes(tag)}"
              class="inline-block text-xs px-2 py-1 rounded-full cursor-pointer hover:bg-indigo-200 mb-1 relative">
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
</template>

<script setup lang="ts">
defineProps<{
  searchQuery: string,
  activeTab: string,
  allCategories: any[],
  allTags: string[],
  activeFilters: any
}>()

const emit = defineEmits<{
  (e: 'update:searchQuery', value: string): void,
  (e: 'update:activeTab', value: string): void,
  (e: 'setCategoryFilter', value: string | null): void,
  (e: 'toggleTagFilter', value: string): void,
  (e: 'removeTagFilter', value: string): void,
  (e: 'searchInput'): void
}>()

function onSearchInput(event: Event) {
  const target = event.target as HTMLInputElement
  emit('update:searchQuery', target.value)
  emit('searchInput')
}

function setActiveTab(tab: string) {
  emit('update:activeTab', tab)
}

function setCategoryFilter(value: string | null) {
  emit('setCategoryFilter', value)
}

function toggleTagFilter(tag: string) {
  emit('toggleTagFilter', tag)
}

function removeTagFilter(tag: string) {
  emit('removeTagFilter', tag)
}
</script>
