<template>
  <div class="glass-card rounded-xl overflow-hidden flex flex-col h-full">
    <!-- 卡片头部 -->
    <div class="p-3 md:p-4 border-b border-white/20 flex-shrink-0 relative">
      <div class="absolute top-2 right-2 flex gap-1">
        <button @click="openRepoEdit"
          class="glass-button p-1.5 rounded-full hover:bg-white/20 transition-colors"
          title="编辑">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-white" fill="none"
            viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
              d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
          </svg>
        </button>
        <button @click="analyzeRepo" :disabled="repo.analyzing"
          class="glass-button p-1.5 rounded-full hover:bg-white/20 transition-colors disabled:opacity-50"
          :title="repo.analyzing ? '分析中...' : 'AI分析'">
          <div v-if="repo.analyzing" class="w-4 h-4 flex items-center justify-center">
            <div
              class="w-3 h-3 border-2 border-white border-t-transparent rounded-full animate-spin">
            </div>
          </div>
          <svg v-else xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-white"
            fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
              d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" />
          </svg>
        </button>
      </div>
      <div class="flex items-start gap-3">
        <div class="flex-grow min-w-0">
          <h3 class="font-bold text-white truncate text-base md:text-lg">
            <a :href="repo.html_url" target="_blank"
              class="text-lg md:text-xl font-semibold text-white hover:underline truncate">
              {{ repo.name }}
            </a>
          </h3>
          <div class="flex flex-wrap gap-1 mt-1 items-center">
            <span v-if="repo.language"
              :style="{ backgroundColor: getLanguageColor(repo.language) }"
              class="inline-block px-2 py-0.5 text-xs rounded-full bg-white/20 text-white/90 leading-tight">
              {{ repo.language }}
            </span>
            <!-- 分类标签 -->
            <div v-if="repo.category" class="inline-flex items-center">
              <span
                class="bg-blue-500 text-white text-xs px-2 py-0.5 rounded-full leading-tight">
                {{ repo.category }}
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 卡片主体 -->
    <div class="p-3 md:p-4 flex-grow relative group">
      <p class="text-white/90 text-sm md:text-base line-clamp-3" ref="descRefs"
        @mouseenter="showTooltip" @mouseleave="hideTooltip"
        @touchstart.prevent="touchStart" @touchend.prevent="touchEnd">
        {{ repo.description || '暂无描述' }}
      </p>

      <!-- 标签 -->
      <div class="mt-3 flex flex-wrap gap-1.5">
        <span v-for="tag in getRepoTags" :key="tag"
          class="inline-block px-2 py-0.5 text-xs rounded-full bg-purple-500/80 text-white truncate max-w-[120px]">
          {{ tag }}
        </span>
        <span v-if="getRepoTags.length === 0"
          class="inline-block px-2 py-0.5 text-xs rounded-full bg-gray-500/80 text-white">
          未标记
        </span>
      </div>
    </div>
    
    <!-- Tooltip 放到 body -->
    <teleport to="body">
      <div v-if="tooltip.visible"
        :style="{ position: 'fixed', top: tooltip.top + 'px', left: tooltip.left + 'px', transform: 'translateY(-100%)' }"
        class="glass-card bg-gradient-to-br from-indigo-500 via-purple-500 to-pink-500  p-3 text-white text-sm rounded-lg w-64 md:w-80 break-words z-[9999]">
        {{ tooltip.text }}
      </div>
    </teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';

const props = defineProps<{
  repo: any
}>()

const emit = defineEmits<{
  (e: 'edit', repo: any): void,
  (e: 'analyze', repo: any): void
}>()

const tooltip = ref({ visible: false, text: '', top: 0, left: 0 })
const touchTimer = ref<ReturnType<typeof setTimeout> | null>(null)

const getRepoTags = computed(() => {
  if (props.repo.tag) {
    return props.repo.tag.split(',').map((tag: string) => tag.trim()).filter((tag: string) => tag);
  }
  return [];
})

function showTooltip(event: MouseEvent) {
  const el = event.target as HTMLElement;
  if (el.scrollHeight > el.clientHeight) {
    const rect = el.getBoundingClientRect();
    tooltip.value = {
      visible: true,
      text: props.repo.description || '暂无描述',
      top: rect.top,
      left: rect.left,
    };
  }
}

function hideTooltip() {
  tooltip.value.visible = false;
  if (touchTimer.value) {
    clearTimeout(touchTimer.value);
    touchTimer.value = null;
  }
}

function touchStart(event: TouchEvent) {
  const el = event.target as HTMLElement;
  if (el.scrollHeight > el.clientHeight) {
    // 500ms 长按显示 tooltip
    touchTimer.value = setTimeout(() => {
      const rect = el.getBoundingClientRect();
      tooltip.value = {
        visible: true,
        text: props.repo.description || '暂无描述',
        top: rect.top,
        left: rect.left,
      };
    }, 500);
  }
}

function touchEnd() {
  if (touchTimer.value) {
    clearTimeout(touchTimer.value);
    touchTimer.value = null;
  }
  // 手指移开立即隐藏 tooltip
  hideTooltip();
}

// 获取语言颜色
function getLanguageColor(language: string) {
  const languageColors: Record<string, string> = {
    'JavaScript': '#f1e05a',
    'Python': '#3572A5',
    'Java': '#b07219',
    'TypeScript': '#2b7489',
    'C++': '#f34b7d',
    'C#': '#178600',
    'Go': '#00ADD8',
    'Ruby': '#701516',
    'PHP': '#4F5D95',
    'Swift': '#ffac45',
    'Kotlin': '#F18E33',
    'Rust': '#dea584',
    'Scala': '#c22d40',
    'Shell': '#89e051',
    'Vue': '#41b883',
    'React': '#61dafb',
    'CSS': '#563d7c',
    'HTML': '#e34c26',
    'C': '#555555',
    'Dart': '#00B4AB',
    'Elixir': '#6e4a7e',
    'Erlang': '#B83998',
    'Haskell': '#5e5086',
    'Lua': '#000080',
    'Objective-C': '#438eff',
    'Perl': '#0298c3',
    'R': '#198ce7',
    'SQL': '#e38c00',
    'Dockerfile': '#384d54'
  };
  return languageColors[language] || '#ccc';
}

function openRepoEdit() {
  emit('edit', props.repo)
}

function analyzeRepo() {
  emit('analyze', props.repo)
}
</script>