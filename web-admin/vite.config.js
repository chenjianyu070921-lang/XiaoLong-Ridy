import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// /admin 前缀的请求转发到本地 go-zero admin 网关，规避浏览器 CORS。
export default defineConfig({
  plugins: [vue()],
  server: {
    port: 5173,
    proxy: {
      '/admin': {
        target: 'http://127.0.0.1:8083',
        changeOrigin: true,
      },
    },
  },
})
