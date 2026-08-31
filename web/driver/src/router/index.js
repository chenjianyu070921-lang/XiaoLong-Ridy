import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  {
    path: '/',
    redirect: () => (localStorage.getItem('driverToken') ? '/home' : '/login')
  },
  {
    path: '/login',
    name: 'DriverLogin',
    component: () => import('@/views/DriverLogin.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/home',
    name: 'DriverHome',
    component: () => import('@/views/DriverHome.vue'),
    meta: { requiresDriverAuth: true }
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to, from, next) => {
  const driverToken = localStorage.getItem('driverToken') || ''
  // 受保护页面无 token → 登录页；登录页不再因"存在 token"强制跳回 /home，
  // 避免过期/无效 token 在 /login 与 /home 之间来回踢（登录一直循环）。
  if (to.meta.requiresDriverAuth && !driverToken) {
    next('/login')
  } else {
    next()
  }
})

export default router
