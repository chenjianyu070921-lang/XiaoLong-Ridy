import path from 'path'
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import Components from 'unplugin-vue-components/vite'
import { VantResolver } from '@vant/auto-import-resolver'

// 默认 18082：8082 在部分 Windows 机器上会被 QQ 抢注，改用 18082 避免联调 404。
const driverApiTarget = process.env.VITE_DRIVER_API_TARGET || 'http://127.0.0.1:18082'

export default defineConfig({
  plugins: [
    vue(),
    Components({
      resolvers: [VantResolver()]
    })
  ],
  server: {
    port: 5175,
    fs: {
      strict: false,
      allow: [
        path.resolve(__dirname),
        path.resolve(__dirname, 'src')
      ]
    },
    proxy: {
      '/api/driver': {
        target: driverApiTarget,
        changeOrigin: true,
        ws: true
      }
    }
  },
  resolve: {
    alias: {
      '@': path.resolve(__dirname, 'src')
    }
  }
})
