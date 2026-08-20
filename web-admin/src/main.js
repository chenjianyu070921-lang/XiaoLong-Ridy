import { createApp } from 'vue'
import { createPinia } from 'pinia'
// 全量引入 Element Plus 是 P0 决策（内部工具，接受首屏 1MB chunk）；P1 可切 unplugin-vue-components 按需加载。
import ElementPlus from 'element-plus'
import zhCn from 'element-plus/es/locale/lang/zh-cn'
import 'element-plus/dist/index.css'
import App from './App.vue'
import router from './router'

const app = createApp(App)
app.use(createPinia())
app.use(router)
app.use(ElementPlus, { locale: zhCn })
app.mount('#app')
