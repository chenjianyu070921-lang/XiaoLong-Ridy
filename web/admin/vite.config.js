import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// /admin 前缀的请求转发到本地管理 HTTP 网关（api/admin 默认监听 8717），规避浏览器 CORS。
export default defineConfig({
  plugins: [vue()],
  server: {
    port: 5173,
    proxy: {
      '/admin': {
        target: 'http://127.0.0.1:8717',
        changeOrigin: true,
      },
    },
  },
})
