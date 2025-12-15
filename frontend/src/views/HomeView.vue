<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import axios from 'axios'
import ToastInfo from '@/components/ToastInfo.vue'
import RepositoryCard from '@/components/RepositoryCard.vue'
import RepoEditModal from '@/components/RepoEditModal.vue'
import MobileFilterDrawer from '@/components/MobileFilterDrawer.vue'
import SidebarFilter from '@/components/SidebarFilter.vue'
import Pagination from '@/components/Pagination.vue'
import SyncProgress from '@/components/SyncProgress.vue'
import IconButton from '@/components/IconButton.vue'
import StatCard from '@/components/StatCard.vue'
import FullScreenLoader from '@/components/FullScreenLoader.vue'
import FilterIcon from '@/components/icons/FilterIcon.vue'
import SettingsIcon from '@/components/icons/SettingsIcon.vue'
import SyncIcon from '@/components/icons/SyncIcon.vue'
import LogoutIcon from '@/components/icons/LogoutIcon.vue'
import StarIcon from '@/components/icons/StarIcon.vue'
import ChartIcon from '@/components/icons/ChartIcon.vue'
import ClockIcon from '@/components/icons/ClockIcon.vue'

// Toast 引用
const toastRef = ref()

// 添加全屏加载状态
const loading = ref(true)

// 数据状态
const user = ref({
  login: '',
  avatar_url: ''
})
const repos = ref<any[]>([])
const categories = ref([
  { value: '', label: '未分类' }
])
const allCategories = ref([
  { value: null, label: '全部' },
  { value: '', label: '未分类' }
])
const stats = ref({
  total_repos: 0,
  analyzed_repos: 0,
  last_sync: '暂无同步'
})

// 同步状态
const syncing = ref(false)
const syncMessage = ref('')
const syncProgress = ref(0)
const ws = ref<WebSocket | null>(null)

// 筛选和分页状态
const showMobileFilter = ref(false)
const activeFilters = ref<{
  category: string | null;
  tags: string[];
}>({
  category: null,
  tags: []
})
const searchQuery = ref('')
const activeTab = ref('categories')
const currentPage = ref(1)
const perPage = ref(9)

// 编辑状态
const repoEditing = ref(false)
const editingRepo = ref<any>({})

// 计算属性
const totalPages = computed(() => {
  return Math.ceil(filteredRepos.value.length / perPage.value)
})

const allTags = computed(() => {
  const tags = new Set<string>()
  repos.value.forEach(repo => {
    // 只添加用户自定义标签，最多3个
    if (repo.tag) {
      const userTags = repo.tag.split(',').map((t: string) => t.trim()).slice(0, 3)
      userTags.forEach((tag: string) => {
        if (tag) {
          tags.add(tag)
        }
      })
    }
  })
  return Array.from(tags).sort()
})

const filteredRepos = computed(() => {
  return repos.value.filter(repo => {
    // 分类筛选
    if (activeFilters.value.category !== null) {
      // 如果选择了特定分类（包括"未分类"），则筛选对应分类的仓库
      if (activeFilters.value.category === '' && repo.category !== '') {
        return false
      }
      if (activeFilters.value.category !== '' && repo.category !== activeFilters.value.category) {
        return false
      }
    }
    // 如果activeFilters.category为null，表示选择"全部"，不需要筛选分类

    // 标签筛选
    if (activeFilters.value.tags.length > 0) {
      const repoTags: string[] = []
      if (repo.tag) {
        repoTags.push(...repo.tag.split(',').map((t: string) => t.trim()))
      }
      if (repo.topics) {
        repoTags.push(...repo.topics.map((t: string) => t.trim()))
      }

      const hasMatchingTag = activeFilters.value.tags.some(tag =>
        repoTags.includes(tag)
      )

      if (!hasMatchingTag) {
        return false
      }
    }

    // 搜索关键词筛选
    if (searchQuery.value) {
      const query = searchQuery.value.toLowerCase()
      const matchesName = repo.name && repo.name.toLowerCase().includes(query)
      const matchesDescription = repo.description && repo.description.toLowerCase().includes(query)
      const matchesLanguage = repo.language && repo.language.toLowerCase().includes(query)

      let matchesTags = false
      if (repo.tag) {
        matchesTags = repo.tag.toLowerCase().includes(query)
      }

      let matchesTopics = false
      if (repo.topics) {
        matchesTopics = repo.topics.some((topic: string) =>
          topic && topic.toLowerCase().includes(query)
        )
      }

      if (!matchesName && !matchesDescription && !matchesLanguage && !matchesTags && !matchesTopics) {
        return false
      }
    }

    return true
  })
})

const filteredAndPaginatedRepos = computed(() => {
  // 在返回分页结果前，确保当前页码有效
  const maxPage = Math.max(1, Math.ceil(filteredRepos.value.length / perPage.value))
  if (currentPage.value > maxPage) {
    currentPage.value = maxPage
  }
  if (currentPage.value < 1) {
    currentPage.value = 1
  }

  const start = (currentPage.value - 1) * perPage.value
  const end = start + perPage.value
  return filteredRepos.value.slice(start, end)
})

// 方法
async function fetchUser() {
  try {
    const res = await axios.get('/api/user')
    user.value = res.data
  } catch (error) {
    console.error('获取用户信息失败:', error)
    toastRef.value.showToast("获取用户信息失败", "error")
  }
}

async function fetchRepos() {
  try {
    const res = await axios.get('/api/repos')
    repos.value = res.data
  } catch (error) {
    console.error('获取仓库列表失败:', error)
    toastRef.value.showToast("获取仓库列表失败", "error")
    // 发生错误时确保repos是空数组
    repos.value = []
  }
}

async function fetchCategories() {
  try {
    const res = await axios.get('/api/categories')
    // 保存所有分类到allCategories，包括"全部"和"未分类"
    allCategories.value = [
      { value: null, label: '全部' },
      { value: '', label: '未分类' },
      ...res.data.map((category: any) => ({
        value: category.value,
        label: category.label,
      })),
    ]
    // 保持categories变量用于其他用途
    categories.value = [
      { value: '', label: '未分类' },
      ...res.data.map((category: any) => ({
        value: category.value || category.id,
        label: category.label || category.name
      }))
    ]
  } catch (error) {
    console.error("获取分类列表失败:", error)
    // 即使获取失败，也要确保基础分类存在
    allCategories.value = [
      { value: null, label: '全部' },
      { value: '', label: '未分类' }
    ]
    categories.value = [
      { value: '', label: '未分类' }
    ]
    toastRef.value.showToast("获取分类列表失败: " + (error as Error).message, "error")
  }
}

async function fetchStats() {
  try {
    const res = await axios.get('/api/stats')
    stats.value = res.data
  } catch (error) {
    console.error("获取统计信息失败:", error)
    toastRef.value.showToast("获取统计信息失败: " + (error as Error).message, "error")
  }
}

async function loadData() {
  // 显示加载动画
  loading.value = true
  
  try {
    // 并行执行所有请求以提高性能
    await Promise.all([
      fetchUser(),
      fetchRepos(),
      fetchStats(),
      fetchCategories()
    ])
  } catch (error) {
    console.error('加载数据时出错:', error)
    toastRef.value.showToast("加载数据时出错", "error")
  } finally {
    // 隐藏加载动画
    loading.value = false
  }
}

function setCategoryFilter(value: string | null) {
  activeFilters.value.category = value
  resetPage()
}

function toggleTagFilter(tag: string) {
  const index = activeFilters.value.tags.indexOf(tag)
  if (index > -1) {
    // 如果标签已经在激活状态，则移除它
    activeFilters.value.tags.splice(index, 1)
  } else {
    // 否则添加标签到激活列表
    activeFilters.value.tags.push(tag)
  }
  resetPage()
}

function removeTagFilter(tag: string) {
  const index = activeFilters.value.tags.indexOf(tag)
  if (index > -1) {
    activeFilters.value.tags.splice(index, 1)
    resetPage()
  }
}

function resetPage() {
  // 重置到第一页
  currentPage.value = 1
}

function syncStars() {
  if (syncing.value) return

  syncing.value = true
  syncProgress.value = 0
  syncMessage.value = '正在连接...'
  
  // 连接到WebSocket
  const wsUrl = (location.protocol === 'https:' ? 'wss://' : 'ws://') +
    location.host + '/api/sync-progress'
  ws.value = new WebSocket(wsUrl)

  ws.value.onopen = () => {
    syncMessage.value = '已连接，准备同步...'
  }

  ws.value.onmessage = (event) => {
    const data = JSON.parse(event.data)
    switch (data.type) {
      case 'start':
        syncMessage.value = data.message
        syncProgress.value = data.progress
        break
      case 'info':
        syncMessage.value = data.message
        syncProgress.value = data.progress
        break
      case 'progress':
        syncMessage.value = data.message
        syncProgress.value = data.progress
        break
      case 'complete':
        syncMessage.value = data.message
        syncProgress.value = data.progress

        // 关闭WebSocket连接
        if (ws.value) {
          ws.value.close()
        }

        // 短暂延迟后重置状态并刷新数据
        setTimeout(async () => {
          syncing.value = false
          await fetchRepos()
          await fetchStats()
          toastRef.value.showToast("同步完成", "success")
        }, 500)
        break
      case 'error':
        syncMessage.value = '错误: ' + data.message
        syncing.value = false
        if (ws.value) {
          ws.value.close()
        }
        toastRef.value.showToast("同步失败: " + data.message, "error")
        break
    }
  }

  ws.value.onerror = (error) => {
    syncing.value = false
    syncMessage.value = '连接错误'
    console.error('WebSocket error:', error)
    toastRef.value.showToast("同步连接失败", "error")
  }

  ws.value.onclose = () => {
    // 如果不是因为完成而关闭，且仍在同步状态，则标记为错误
    if (syncing.value) {
      syncing.value = false
      syncMessage.value = '连接已关闭'
    }
  }
}

function openRepoEdit(repo: any) {
  // 创建当前编辑仓库的副本
  editingRepo.value = {
    ...repo,
    newTags: repo.tag || "",
    newTagInput: "",
    description: repo.description || ""
  }
  repoEditing.value = true
}

async function saveRepoEdit(repo: any) {
  // 处理标签数组
  if (repo.newTags !== undefined) {
    repo.tag = repo.newTags
  }

  // 调用后端接口更新标签和分类
  try {
    await axios.post(`/api/repos/${repo.id}/tag`, { tag: repo.tag })
    await axios.post(`/api/repos/${repo.id}/category`, { category: repo.category })
    await axios.post(`/api/repos/${repo.id}/description`, { description: repo.description })
    repoEditing.value = false

    // 更新仓库列表中的数据
    const index = repos.value.findIndex(r => r.id === repo.id)
    if (index !== -1) {
      repos.value[index] = {
        ...repos.value[index],
        tag: repo.tag,
        category: repo.category,
        description: repo.description
      }
    }

    toastRef.value.showToast("更新成功", "success")
  } catch (error) {
    toastRef.value.showToast("更新失败: " + (error as Error).message, "error")
  }
}

function cancelRepoEdit() {
  repoEditing.value = false
}

// AI分析仓库
async function analyzeRepo(repo: any) {
  // 防止重复点击
  if (repo.analyzing) {
    return
  }

  try {
    // 设置分析状态
    repo.analyzing = true

    // 显示分析中提示
    toastRef.value.showToast("正在使用AI分析仓库...", "info")

    // 调用后端AI分析接口
    const response = await axios.post(`/api/repos/${repo.id}/analyze`)

    // 更新仓库信息
    const index = repos.value.findIndex((r: any) => r.id === repo.id)
    if (index !== -1) {
      repos.value[index].category = response.data.category
      repos.value[index].tag = response.data.tags.join(',')
      // 更新描述为AI分析的结果
      repos.value[index].description = response.data.description
    }

    // 显示成功提示
    toastRef.value.showToast("AI分析完成，结果已保存", "success")

    // 刷新统计信息
    await fetchStats()
  } catch (error: any) {
    console.error("AI分析失败:", error)
    const errorMsg = error.response?.data?.error || error.message || "未知错误"
    toastRef.value.showToast("AI分析失败: " + errorMsg, "error")
  } finally {
    // 重置分析状态
    repo.analyzing = false
  }
}

function logout() {
  window.location.href = "/logout"
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

onMounted(() => {
  loadData()
})
</script>

<template>
  <div class="flex flex-col h-screen bg-gradient-to-br from-indigo-500 via-purple-500 to-pink-500">
    <!-- 全屏加载动画 -->
    <FullScreenLoader v-if="loading" />
    
    <!-- Toast容器 -->
    <ToastInfo ref="toastRef" />

    <!-- 移动端筛选抽屉 -->
    <MobileFilterDrawer
      :visible="showMobileFilter"
      :search-query="searchQuery"
      :active-tab="activeTab"
      :all-categories="allCategories"
      :all-tags="allTags"
      :active-filters="activeFilters"
      :stats="stats"
      @close="showMobileFilter = false"
      @update:search-query="searchQuery = $event"
      @update:active-tab="activeTab = $event"
      @set-category-filter="setCategoryFilter"
      @toggle-tag-filter="toggleTagFilter"
      @remove-tag-filter="removeTagFilter"
      @search-input="resetPage"
    />

    <!-- 顶部固定区域 -->
    <div class="sticky top-0 z-20 backdrop-blur-md bg-white/20">
      <!-- Header -->
      <header class="p-3 md:p-4">
        <div
          class="flex flex-col md:flex-row justify-between items-center gap-3 md:gap-4">
          <!-- 左侧用户信息 -->
          <div class="flex items-center gap-3 w-full md:w-auto justify-between md:justify-start">
            <div class="flex items-center gap-3">
              <img :src="user.avatar_url"
                class="w-10 h-10 md:w-12 md:h-12 rounded-full border-2 border-white" />
              <span class="font-semibold text-white text-base md:text-lg hidden md:inline">{{ user.login }}</span>
            </div>

            <!-- 移动端按钮 -->
            <div class="flex md:hidden gap-2">
              <IconButton @click="showMobileFilter = true">
                <FilterIcon />
              </IconButton>
              <a href="/settings">
                <IconButton>
                  <SettingsIcon />
                </IconButton>
              </a>
              <IconButton @click="syncStars" :disabled="syncing">
                <div v-if="syncing" class="spinner"></div>
                <SyncIcon v-else />
              </IconButton>
              <IconButton @click="logout">
                <LogoutIcon />
              </IconButton>
            </div>
          </div>

          <!-- 右侧统计信息和按钮 -->
          <div class="flex flex-col md:flex-row items-center gap-3 w-full md:w-auto mt-3 md:mt-0">
            <!-- 统计信息 -->
            <div class="hidden md:flex gap-2 md:gap-3 h-10 w-full md:w-auto justify-between">
              <StatCard>
                <template #icon>
                  <StarIcon />
                </template>
                {{ stats.total_repos }}
              </StatCard>
              <StatCard>
                <template #icon>
                  <ChartIcon />
                </template>
                {{ stats.analyzed_repos }}
              </StatCard>
              <StatCard>
                <template #icon>
                  <ClockIcon />
                </template>
                <span class="text-xs">{{ formatSyncTime(stats.last_sync) }}</span>
              </StatCard>
            </div>

            <!-- 桌面端按钮 -->
            <div class="hidden md:flex gap-2">
              <a href="/settings">
                <IconButton>
                  <SettingsIcon />
                </IconButton>
              </a>
              <IconButton @click="syncStars" :disabled="syncing">
                <div v-if="syncing" class="spinner"></div>
                <SyncIcon v-else />
              </IconButton>
              <IconButton @click="logout">
                <LogoutIcon />
              </IconButton>
            </div>
          </div>
        </div>
      </header>

      <!-- 同步进度条 -->
      <SyncProgress 
        :syncing="syncing" 
        :sync-message="syncMessage" 
        :sync-progress="syncProgress" 
      />

      <!-- 分页控件 -->
      <Pagination
        v-model:current-page="currentPage"
        v-model:per-page="perPage"
        :total-pages="totalPages"
      />
    </div>

    <!-- 主体内容区域 -->
    <main class="flex-grow flex flex-col md:flex-row p-3 md:p-4 gap-3 md:gap-4 overflow-hidden">
      <!-- 左侧筛选区域 -->
      <SidebarFilter
        :search-query="searchQuery"
        :active-tab="activeTab"
        :all-categories="allCategories"
        :all-tags="allTags"
        :active-filters="activeFilters"
        @update:search-query="searchQuery = $event"
        @update:active-tab="activeTab = $event"
        @set-category-filter="setCategoryFilter"
        @toggle-tag-filter="toggleTagFilter"
        @remove-tag-filter="removeTagFilter"
        @search-input="resetPage"
      />

      <!-- 右侧仓库卡片展示 -->
      <div class="flex-grow flex flex-col overflow-hidden">
        <!-- 仓库卡片网格 -->
        <div class="flex-grow overflow-y-auto">
          <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4 md:gap-6 p-4">
            <RepositoryCard
              v-for="repo in filteredAndPaginatedRepos"
              :key="repo.id"
              :repo="repo"
              @edit="openRepoEdit"
              @analyze="analyzeRepo"
            />
          </div>
        </div>
      </div>

      <!-- 编辑区域 - 弹窗形式 -->
      <RepoEditModal
        v-if="repoEditing"
        :repo="editingRepo"
        :categories="categories"
        @save="saveRepoEdit"
        @cancel="cancelRepoEdit"
      />
    </main>
  </div>
</template>