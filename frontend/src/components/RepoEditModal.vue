<template>
  <div class="fixed inset-0 flex items-center justify-center z-50 p-4">
    <div class="absolute inset-0 backdrop-blur-sm bg-white/10" @click="cancelEdit"></div>
    <div class="glass-card flex flex-col h-[32rem] w-full max-w-2xl rounded-lg relative z-10" @click.stop="">
      <!-- 标签输入 -->
      <div class="p-3 md:p-4 flex-shrink-0">
        <div class="flex items-center mb-2">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-white mr-2" fill="none"
            viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
              d="M7 7h.01M7 3h5c.512 0 1.024.195 1.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A1.994 1.994 0 013 12V7a4 4 0 014-4z" />
          </svg>
          <span class="text-white font-medium">标签</span>
        </div>

        <!-- 显示已添加的标签 -->
        <div class="flex flex-wrap gap-1 mb-2 min-h-6">
          <span v-for="(tag, index) in tagArray" :key="index"
            class="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-indigo-100 text-indigo-800">
            {{ tag }}
            <button @click="removeTag(index)"
              class="ml-1 text-indigo-600 hover:text-indigo-900">
              <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3" fill="none"
                viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                  d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </span>
        </div>

        <!-- 添加新标签输入框 -->
        <div class="glass-button flex items-center rounded-lg border border-white/30">
          <input :value="newTagInput" @input="onNewTagInput" placeholder="输入新标签，支持逗号或空格分隔添加多个"
            @keyup.enter="addTag"
            class="text-white px-3 py-2 rounded-l-lg flex-grow bg-transparent placeholder-white/50 focus:outline-none w-full" />
          <button @click="addTag"
            class="text-white rounded-r-lg hover:bg-white/20 p-2 transition">
          </button>
        </div>
      </div>

      <!-- 描述编辑 -->
      <div class="px-3 md:px-4 mb-4 flex-shrink-0">
        <div class="flex items-center mb-2">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-white mr-2" fill="none"
            viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
              d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
          </svg>
          <span class="text-white font-medium">描述</span>
        </div>
        <div class="glass-button rounded-lg border border-white/30">
          <textarea :value="localRepo.description" @input="onDescriptionInput" rows="6" placeholder="请输入项目描述..."
            class="text-white px-3 py-2 rounded-lg w-full bg-transparent placeholder-white/50 focus:outline-none resize-none"></textarea>
        </div>
      </div>

      <!-- 分类选择 -->
      <div class="px-3 md:px-4 mb-4 flex-shrink-0">
        <div class="flex items-center mb-2">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-white mr-2" fill="none"
            viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
              d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
          </svg>
          <span class="text-white font-medium">分类</span>
        </div>
        <div class="glass-button rounded-lg relative border border-white/30">
          <select :value="localRepo.category" @change="onCategoryChange"
            class="text-white px-3 py-2 rounded-lg flex-grow w-full bg-transparent appearance-none focus:outline-none">
            <option v-for="category in categories" :key="category.value" :value="category.value"
              class="text-gray-800">
              {{ category.label }}
            </option>
          </select>
          <!-- 添加下拉箭头图标 -->
          <div class="absolute right-3 top-1/2 transform -translate-y-1/2 pointer-events-none">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-white" fill="none"
              viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                d="M19 9l-7 7-7-7" />
            </svg>
          </div>
        </div>
      </div>

      <!-- 操作按钮 -->
      <div class="flex justify-end gap-3 px-3 md:px-4 mt-auto pb-3 md:pb-4">
        <button @click="cancelEdit"
          class="glass-button px-4 py-2 rounded-lg text-white hover:bg-white/20">
          取消
        </button>
        <button @click="saveEdit"
          class="glass-button px-4 py-2 rounded-lg text-white hover:bg-white/20">
          保存
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, reactive } from 'vue';

const props = defineProps<{
  repo: any,
  categories: any[]
}>()

const emit = defineEmits<{
  (e: 'save', repo: any): void,
  (e: 'cancel'): void
}>()

// 创建本地副本以避免直接修改props
const localRepo = reactive({...props.repo})
const newTagInput = ref('')

const tagArray = computed(() => {
  return localRepo.tag ? localRepo.tag.split(',').map((tag: string) => tag.trim()).filter((tag: string) => tag) : [];
})

function addTag() {
  if (newTagInput.value && newTagInput.value.trim() !== '') {
    const currentTags = [...tagArray.value];

    // 支持逗号或空格分隔的多个标签
    // 先按逗号分割，再按空格分割，然后去除空字符串
    let newTags: string[] = [];
    const commaSeparated = newTagInput.value.split(',');
    commaSeparated.forEach(part => {
      if (part.includes(' ')) {
        // 如果包含空格，则按空格分割
        newTags.push(...part.split(' ').filter(tag => tag.trim() !== ''));
      } else {
        // 否则直接添加（去除首尾空格）
        const trimmed = part.trim();
        if (trimmed !== '') {
          newTags.push(trimmed);
        }
      }
    });

    // 去除所有标签的首尾空格
    newTags = newTags.map(tag => tag.trim()).filter(tag => tag !== '');

    // 添加新标签（避免重复）
    newTags.forEach(newTag => {
      if (!currentTags.includes(newTag)) {
        currentTags.push(newTag);
      }
    });

    localRepo.tag = currentTags.join(',');
    newTagInput.value = '';
  }
}

function removeTag(index: number) {
  const currentTags = [...tagArray.value];
  currentTags.splice(index, 1);
  localRepo.tag = currentTags.join(',');
}

function saveEdit() {
  emit('save', localRepo)
}

function cancelEdit() {
  emit('cancel')
}

// 监听ESC键关闭
defineExpose({
  cancelEdit
})

function onNewTagInput(event: Event) {
  newTagInput.value = (event.target as HTMLInputElement).value;
}

function onDescriptionInput(event: Event) {
  localRepo.description = (event.target as HTMLTextAreaElement).value;
}

function onCategoryChange(event: Event) {
  localRepo.category = (event.target as HTMLSelectElement).value;
}

</script>