import { fileURLToPath, URL } from 'node:url'
import { resolve } from 'path';

import { defineConfig, loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'
import vueDevTools from 'vite-plugin-vue-devtools'
import tailwindcss from '@tailwindcss/vite'

// https://vite.dev/config/
export default defineConfig(({ mode }) => {
  // 加载环境变量
  const env = loadEnv(mode, process.cwd(), '')
  
  return {
    plugins: [
      vue(),
      vueDevTools(),
      tailwindcss(),
    ],
    resolve: {
      alias: {
        '@': fileURLToPath(new URL('./src', import.meta.url))
      },
    },
    build: {
      rollupOptions: {
        input: {
          home: resolve(__dirname, 'index.html'),
          settings: resolve(__dirname, 'settings.html'),
          login: resolve(__dirname, 'login.html')
        }
      },
      outDir: "dist",
      assetsDir: 'static',
      emptyOutDir: true
    },
    server: {
      host: true,
      port: 5173,
      strictPort: true,
      // 配置代理解决开发环境跨域问题
      proxy: {
        '/token-login': {
          target: env.VITE_API_URL,
          changeOrigin: true,
          secure: false
        },
        '/auth': {
          target: env.VITE_API_URL,
          changeOrigin: true,
          secure: false
        },
        '/api': {
          target: env.VITE_API_URL,
          changeOrigin: true,
          secure: false
        }
      }
    }
  }
})