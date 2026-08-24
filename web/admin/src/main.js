import { createApp } from 'vue'
import { createPinia } from 'pinia'
// 模板组件由 Vite 的 ElementPlusResolver 按需导入；以下样式服务于命令式消息 API。
import 'element-plus/es/components/message/style/css'
import 'element-plus/es/components/message-box/style/css'
import App from './App.vue'
import router from './router'
import { useUserStore } from './store/user'

const app = createApp(App)
const pinia = createPinia()
app.use(pinia)
app.use(router)
// startApp 在挂载前校验本地 token，确保顶栏显示的管理员资料来自当前服务端会话。
// 使用异步函数替代顶层 await，以兼容项目当前 Vite 浏览器构建目标。
const startApp = async () => {
  const userStore = useUserStore(pinia)
  if (userStore.token) {
    try {
      await userStore.refreshProfile()
    } catch {
      // 请求拦截器会处理会话失效跳转；刷新失败不阻塞登录页或失效会话页面渲染。
    }
  }
  app.mount('#app')
}

void startApp()
