<script setup lang="ts">
import { ref, onMounted } from 'vue';
import axios from 'axios';
import ToastInfo from '@/components/ToastInfo.vue';

// Toast 引用
const toastRef = ref();

// 表单数据
const settings = ref({
  openai: {
    key: '',
    endpoint: '',
    model: 'gpt-3.5-turbo',
    headers: [] as { key: string; value: string }[],
    body: [] as { key: string; value: string }[]
  },
  webdav: {
    url: '',
    username: '',
    password: ''
  }
});

// 状态标志
const testingOpenAI = ref(false);
const testingWebDAV = ref(false);
const saving = ref(false);

// 添加自定义请求头
function addHeader() {
  if (!settings.value.openai.headers) {
    settings.value.openai.headers = [];
  }
  settings.value.openai.headers.push({ key: '', value: '' });
}

// 删除自定义请求头
function removeHeader(index: number) {
  if (settings.value.openai.headers) {
    settings.value.openai.headers.splice(index, 1);
  }
}

// 添加自定义请求体字段
function addBodyField() {
  if (!settings.value.openai.body) {
    settings.value.openai.body = [];
  }
  settings.value.openai.body.push({ key: '', value: '' });
}

// 删除自定义请求体字段
function removeBodyField(index: number) {
  if (settings.value.openai.body) {
    settings.value.openai.body.splice(index, 1);
  }
}

// 返回上一页
function goBack() {
  window.location.href = '/';
}

// 加载设置
async function loadSettings() {
  try {
    const response = await axios.get('/api/settings');
    settings.value = response.data;
    // 确保headers和body字段存在且为数组
    if (!settings.value.openai) {
      settings.value.openai = {
        key: '',
        endpoint: '',
        model: 'gpt-3.5-turbo',
        headers: [],
        body: []
      };
    } else {
      if (!settings.value.openai.headers) {
        settings.value.openai.headers = [];
      }
      if (!settings.value.openai.body) {
        settings.value.openai.body = [];
      }
    }
    
    if (!settings.value.webdav) {
      settings.value.webdav = {
        url: '',
        username: '',
        password: ''
      };
    }
  } catch (error: any) {
    console.error('加载设置失败:', error);
    toastRef.value.showToast('加载设置失败', 'error');
  }
}

// 验证OpenAI表单
function validateOpenAIForm() {
  if (!settings.value.openai) {
    toastRef.value.showToast('OpenAI配置不存在', 'error');
    return false;
  }
  
  if (!settings.value.openai.key || !settings.value.openai.key.trim()) {
    toastRef.value.showToast('请填写OpenAI API Key', 'error');
    return false;
  }
  
  if (!settings.value.openai.model || !settings.value.openai.model.trim()) {
    toastRef.value.showToast('请填写OpenAI模型名称', 'error');
    return false;
  }
  
  if (!settings.value.openai.endpoint || !settings.value.openai.endpoint.trim()) {
    toastRef.value.showToast('请填写OpenAI API地址', 'error');
    return false;
  }
  
  return true;
}

// 验证WebDAV表单
function validateWebDAVForm() {
  if (!settings.value.webdav) {
    toastRef.value.showToast('WebDAV配置不存在', 'error');
    return false;
  }
  
  if (!settings.value.webdav.url || !settings.value.webdav.url.trim()) {
    toastRef.value.showToast('请填写WebDAV服务器地址', 'error');
    return false;
  }
  
  if (!settings.value.webdav.username || !settings.value.webdav.username.trim()) {
    toastRef.value.showToast('请填写WebDAV用户名', 'error');
    return false;
  }
  
  if (!settings.value.webdav.password || !settings.value.webdav.password.trim()) {
    toastRef.value.showToast('请填写WebDAV密码', 'error');
    return false;
  }
  
  return true;
}

// 测试 OpenAI 连接
async function testOpenAI() {
  // 先验证表单
  if (!validateOpenAIForm()) {
    return;
  }
  
  testingOpenAI.value = true;
  try {
    // 确保openai对象存在
    if (!settings.value.openai) {
      settings.value.openai = {
        key: '',
        endpoint: '',
        model: 'gpt-3.5-turbo',
        headers: [],
        body: []
      };
    }
    
    const response = await axios.post('/api/test-openai', settings.value.openai);
    toastRef.value.showToast(response.data.message, response.data.success ? 'success' : 'error');
  } catch (error: any) {
    toastRef.value.showToast('测试 OpenAI 连接失败: ' + (error.response?.data?.error || error.message), 'error');
  } finally {
    testingOpenAI.value = false;
  }
}

// 测试 WebDAV 连接
async function testWebDAV() {
  // 先验证表单
  if (!validateWebDAVForm()) {
    return;
  }
  
  testingWebDAV.value = true;
  try {
    // 确保webdav对象存在
    if (!settings.value.webdav) {
      settings.value.webdav = {
        url: '',
        username: '',
        password: ''
      };
    }
    
    const response = await axios.post('/api/test-webdav', settings.value.webdav);
    toastRef.value.showToast(response.data.message, response.data.success ? 'success' : 'error');
  } catch (error: any) {
    toastRef.value.showToast('测试 WebDAV 连接失败: ' + (error.response?.data?.error || error.message), 'error');
  } finally {
    testingWebDAV.value = false;
  }
}

// 保存设置
async function saveSettings() {
  saving.value = true;
  try {
    await axios.post('/api/settings', settings.value);
    toastRef.value.showToast('设置已保存', 'success');
  } catch (error: any) {
    toastRef.value.showToast('保存设置失败: ' + (error.response?.data?.error || error.message), 'error');
  } finally {
    saving.value = false;
  }
}

onMounted(() => {
  loadSettings();
});
</script>

<template>
  <div class="flex flex-col h-screen bg-gradient-to-br from-indigo-500 via-purple-500 to-pink-500">
    <!-- Toast容器 -->
    <ToastInfo ref="toastRef" />

    <!-- 顶部固定区域 -->
    <div class="sticky top-0 z-20 backdrop-blur-md bg-white/20">
      <!-- Header -->
      <header class="p-3 md:p-4">
        <div class="backdrop-filter backdrop-blur-lg bg-white/15 border border-white/20 p-3 md:p-4 rounded-xl flex flex-col md:flex-row justify-between items-center gap-3 md:gap-4">
          <!-- 左侧标题和返回按钮 -->
          <div class="flex items-center gap-3 w-full md:w-auto">
            <button @click="goBack"
              class="backdrop-filter backdrop-blur-lg bg-white/20 border border-white/30 rounded-full w-10 h-10 flex items-center justify-center hover:bg-white/30 transition-all">
              <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-white" fill="none"
                viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                  d="M10 19l-7-7m0 0l7-7m-7 7h18" />
              </svg>
            </button>
            <h1 class="text-white text-xl font-bold">设置</h1>
          </div>
        </div>
      </header>
    </div>

    <!-- 主要内容区域 -->
    <main class="flex-grow overflow-y-auto p-4">
      <div class="max-w-4xl mx-auto">
        <!-- OpenAI 配置 -->
        <div class="backdrop-filter backdrop-blur-lg bg-white/15 border border-white/20 p-6 rounded-xl mb-6">
          <h2 class="text-white text-xl font-bold mb-4">OpenAI 配置</h2>
          
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4 mb-4">
            <div>
              <label class="block text-white text-sm font-medium mb-1">API Key</label>
              <input type="password" v-model="settings.openai.key" placeholder="输入 OpenAI API Key"
                class="bg-white/10 border border-white/20 color-white rounded-lg p-2 focus:outline-none focus:border-indigo-400/80 focus:shadow-[0_0_0_3px_rgba(99,102,241,0.3)] w-full">
            </div>
            
            <div>
              <label class="block text-white text-sm font-medium mb-1">API Endpoint</label>
              <input type="text" v-model="settings.openai.endpoint" placeholder="OpenAI API 地址"
                class="bg-white/10 border border-white/20 color-white rounded-lg p-2 focus:outline-none focus:border-indigo-400/80 focus:shadow-[0_0_0_3px_rgba(99,102,241,0.3)] w-full">
            </div>

            <div>
              <label class="block text-white text-sm font-medium mb-1">模型</label>
              <input type="text" v-model="settings.openai.model" placeholder="例如: gpt-3.5-turbo"
                class="bg-white/10 border border-white/20 color-white rounded-lg p-2 focus:outline-none focus:border-indigo-400/80 focus:shadow-[0_0_0_3px_rgba(99,102,241,0.3)] w-full">
            </div>
          </div>

          <!-- 自定义请求头 -->
          <div class="mb-6">
            <div class="flex items-center justify-between mb-3">
              <h3 class="text-white text-lg font-semibold">自定义请求头</h3>
              <button @click="addHeader" class="backdrop-filter backdrop-blur-lg bg-white/20 border border-white/30 rounded-lg px-4 py-2 text-sm flex items-center hover:bg-white/30 transition-all">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
                </svg>
                添加
              </button>
            </div>
            
            <div v-if="!settings.openai.headers || settings.openai.headers.length === 0" class="text-white/60 text-sm mb-2">
              暂无自定义请求头
            </div>
            
            <div v-for="(header, index) in settings.openai.headers" :key="index" class="grid grid-cols-12 gap-2 mb-2">
              <div class="col-span-5">
                <input type="text" v-model="header.key" placeholder="请求头键"
                  class="bg-white/10 border border-white/20 color-white rounded-lg p-2 focus:outline-none focus:border-indigo-400/80 focus:shadow-[0_0_0_3px_rgba(99,102,241,0.3)] w-full text-sm h-10">
              </div>
              <div class="col-span-5">
                <input type="text" v-model="header.value" placeholder="请求头值"
                  class="bg-white/10 border border-white/20 color-white rounded-lg p-2 focus:outline-none focus:border-indigo-400/80 focus:shadow-[0_0_0_3px_rgba(99,102,241,0.3)] w-full text-sm h-10">
              </div>
              <div class="col-span-2">
                <button @click="removeHeader(index)" 
                  class="backdrop-filter backdrop-blur-lg bg-white/20 border border-white/30 rounded-lg w-full h-10 px-2 text-sm flex items-center justify-center text-red-400 hover:text-red-300 hover:bg-white/30 transition-all">
                  <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                  </svg>
                </button>
              </div>
            </div>
          </div>

          <!-- 自定义请求体 -->
          <div class="mb-6">
            <div class="flex items-center justify-between mb-3">
              <h3 class="text-white text-lg font-semibold">自定义请求体</h3>
              <button @click="addBodyField" class="backdrop-filter backdrop-blur-lg bg-white/20 border border-white/30 rounded-lg px-4 py-2 text-sm flex items-center hover:bg-white/30 transition-all">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
                </svg>
                添加
              </button>
            </div>
            
            <div v-if="!settings.openai.body || settings.openai.body.length === 0" class="text-white/60 text-sm mb-2">
              暂无自定义请求体参数
            </div>
            
            <div v-for="(field, index) in settings.openai.body" :key="index" class="grid grid-cols-12 gap-2 mb-2">
              <div class="col-span-5">
                <input type="text" v-model="field.key" placeholder="字段键 (支持嵌套，如: parameters.temperature)"
                  class="bg-white/10 border border-white/20 color-white rounded-lg p-2 focus:outline-none focus:border-indigo-400/80 focus:shadow-[0_0_0_3px_rgba(99,102,241,0.3)] w-full text-sm h-10">
              </div>
              <div class="col-span-5">
                <input type="text" v-model="field.value" placeholder="字段值"
                  class="bg-white/10 border border-white/20 color-white rounded-lg p-2 focus:outline-none focus:border-indigo-400/80 focus:shadow-[0_0_0_3px_rgba(99,102,241,0.3)] w-full text-sm h-10">
              </div>
              <div class="col-span-2">
                <button @click="removeBodyField(index)" 
                  class="backdrop-filter backdrop-blur-lg bg-white/20 border border-white/30 rounded-lg w-full h-10 px-2 text-sm flex items-center justify-center text-red-400 hover:text-red-300 hover:bg-white/30 transition-all">
                  <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                  </svg>
                </button>
              </div>
            </div>
          </div>
          
          <div class="flex flex-wrap gap-2">
            <button @click="testOpenAI" :disabled="testingOpenAI"
              class="backdrop-filter backdrop-blur-lg bg-white/20 border border-white/30 rounded-lg px-4 py-2 flex items-center gap-2 hover:bg-white/30 transition-all disabled:opacity-60 disabled:cursor-not-allowed">
              <span v-if="testingOpenAI" class="border-2 border-white/30 rounded-full border-t-white w-4 h-4 animate-spin inline-block"></span>
              <span>{{ testingOpenAI ? '测试中...' : '测试连接' }}</span>
            </button>
          </div>
        </div>

        <!-- WebDAV 配置 -->
        <div class="backdrop-filter backdrop-blur-lg bg-white/15 border border-white/20 p-6 rounded-xl mb-6">
          <h2 class="text-white text-xl font-bold mb-4">WebDAV 配置</h2>
          
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4 mb-4">
            <div>
              <label class="block text-white text-sm font-medium mb-1">服务器地址</label>
              <input type="text" v-model="settings.webdav.url" placeholder="WebDAV 服务器地址"
                class="bg-white/10 border border-white/20 color-white rounded-lg p-2 focus:outline-none focus:border-indigo-400/80 focus:shadow-[0_0_0_3px_rgba(99,102,241,0.3)] w-full">
            </div>
            
            <div>
              <label class="block text-white text-sm font-medium mb-1">用户名</label>
              <input type="text" v-model="settings.webdav.username" placeholder="WebDAV 用户名"
                class="bg-white/10 border border-white/20 color-white rounded-lg p-2 focus:outline-none focus:border-indigo-400/80 focus:shadow-[0_0_0_3px_rgba(99,102,241,0.3)] w-full">
            </div>
            
            <div>
              <label class="block text-white text-sm font-medium mb-1">密码</label>
              <input type="password" v-model="settings.webdav.password" placeholder="WebDAV 密码"
                class="bg-white/10 border border-white/20 color-white rounded-lg p-2 focus:outline-none focus:border-indigo-400/80 focus:shadow-[0_0_0_3px_rgba(99,102,241,0.3)] w-full">
            </div>
          </div>
          
          <div class="flex flex-wrap gap-2">
            <button @click="testWebDAV" :disabled="testingWebDAV"
              class="backdrop-filter backdrop-blur-lg bg-white/20 border border-white/30 rounded-lg px-4 py-2 flex items-center gap-2 hover:bg-white/30 transition-all disabled:opacity-60 disabled:cursor-not-allowed">
              <span v-if="testingWebDAV" class="border-2 border-white/30 rounded-full border-t-white w-4 h-4 animate-spin inline-block"></span>
              <span>{{ testingWebDAV ? '测试中...' : '测试连接' }}</span>
            </button>
          </div>
        </div>

        <!-- 保存按钮 -->
        <div class="backdrop-filter backdrop-blur-lg bg-white/15 border border-white/20 p-6 rounded-xl">
          <div class="flex justify-end">
            <button @click="saveSettings" :disabled="saving"
              class="backdrop-filter backdrop-blur-lg bg-white/20 border border-white/30 rounded-lg px-6 py-2 font-medium text-white flex items-center gap-2 hover:bg-white/30 transition-all disabled:opacity-60 disabled:cursor-not-allowed">
              <span v-if="saving" class="border-2 border-white/30 rounded-full border-t-white w-4 h-4 animate-spin inline-block"></span>
              <span>{{ saving ? '保存中...' : '保存设置' }}</span>
            </button>
          </div>
        </div>
      </div>
    </main>
  </div>
</template>