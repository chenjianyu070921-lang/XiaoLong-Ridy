import { createRouter, createWebHashHistory } from 'vue-router'
import { h } from 'vue'

// 占位路由：Task 3 会完整重写为带登录守卫的完整路由表。
// 注意：Vite 默认使用 runtime-only 版 Vue（无模板编译器），
// 组件选项必须用 h() 渲染函数，不能写 template 字符串，否则浏览器白屏。
const router = createRouter({
  history: createWebHashHistory(),
  routes: [{ path: '/', component: { render: () => h('div', 'web-admin ok') } }],
})

export default router
