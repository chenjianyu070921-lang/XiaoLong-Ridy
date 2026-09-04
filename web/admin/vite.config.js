import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import Components from 'unplugin-vue-components/vite'
import { ElementPlusResolver } from 'unplugin-vue-components/resolvers'

// /admin 前缀的请求转发到本地管理 HTTP 网关（api/admin 默认监听 8717），规避浏览器 CORS。
export default defineConfig({
  plugins: [
    vue(),
    // 按模板实际使用的组件生成精确 import，避免将整个 Element Plus 打入首屏主包。
    Components({
      resolvers: [ElementPlusResolver({ importStyle: 'css' })],
    }),
  ],
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
