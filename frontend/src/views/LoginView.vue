<script setup lang="ts">
import { ref } from 'vue';
import axios from 'axios';
import ToastInfo from '@/components/ToastInfo.vue';

const token = ref('');
const loading = ref(false);
const toastRef = ref();

async function tokenLogin() {
  if (!token.value.trim()) {
    toastRef.value.showToast('请输入 GitHub Token', 'error');
    return;
  }

  loading.value = true;

  try {
    await axios.post('/auth/token-login', {
      token: token.value
    });

    window.location.href = '/';
  } catch (error: unknown) {
    let errorMessage = '未知错误';
    if (error && typeof error === 'object') {
      if ('response' in error && error.response && typeof error.response === 'object') {
        if ('data' in error.response && error.response.data && typeof error.response.data === 'object') {
          if ('msg' in error.response.data) {
            errorMessage = error.response.data.msg as string;
          }
        }
      } else if ('message' in error) {
        errorMessage = error.message as string;
      }
    } else if (typeof error === 'string') {
      errorMessage = error;
    }
    toastRef.value.showToast('登录失败: ' + errorMessage, 'error');
  } finally {
    loading.value = false;
  }
}
</script>

<template>
  <div class="min-h-screen flex items-center justify-center bg-gradient-to-br from-indigo-500 via-purple-500 to-pink-500">
    <ToastInfo ref="toastRef" />
    
    <div class="w-full max-w-md p-8 rounded-lg backdrop-blur-sm bg-white/15 border border-white/20 shadow-2xl">
      <h1 class="text-2xl font-bold text-center mb-6 text-white">GitHub Stars 管理器</h1>
      
      <!-- Token 登录 -->
      <form class="flex flex-col gap-4">
        <input 
          v-model="token" 
          type="password" 
          placeholder="输入 GitHub Token"
          class="bg-white/10 border border-white/20 color-white rounded-lg p-2 focus:outline-none focus:border-indigo-400/80 focus:shadow-[0_0_0_3px_rgba(99,102,241,0.3)]"
        />
        <button 
          type="button" 
          @click="tokenLogin" 
          :disabled="loading"
          class="backdrop-blur-sm bg-white/20 border border-white/30 rounded-lg transition-all py-2 hover:bg-white/30 hover:shadow-lg hover:-translate-y-0.5 disabled:opacity-60 disabled:cursor-not-allowed flex items-center justify-center"
        >
          <span v-if="!loading">使用 Token 登录</span>
          <span v-else class="flex items-center">
            <span class="border-2 border-white/30 rounded-full border-t-white w-4 h-4 animate-spin inline-block mr-2"></span>
            登录中...
          </span>
        </button>
      </form>
      
      <div class="my-4 text-center text-white/70">或</div>
      
      <!-- GitHub OAuth 登录 -->
      <a 
        href="/auth/github"
        class="block text-center backdrop-blur-sm bg-white/20 border border-white/30 rounded-lg transition-all py-2 hover:bg-gray-900/80 hover:shadow-lg"
      >
        使用 GitHub 登录
      </a>
      
      <p class="text-sm text-white/80 mt-6 text-center">
        登录后可管理你的 GitHub Star 仓库
      </p>
    </div>
  </div>
</template>
