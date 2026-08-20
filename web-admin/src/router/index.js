import { createRouter, createWebHashHistory } from 'vue-router'

// 占位路由：Task 3 会完整重写为带登录守卫的完整路由表。
const router = createRouter({
  history: createWebHashHistory(),
  routes: [{ path: '/', component: { template: '<div>web-admin ok</div>' } }],
})

export default router
