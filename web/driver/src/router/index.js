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
    meta: { requiresDriverAuth: false }
  },
  {
    path: '/home',
    name: 'DriverHome',
    component: () => import('@/views/DriverHome.vue'),
    meta: { requiresDriverAuth: true }
  },
  {
    path: '/:pathMatch(.*)*',
    redirect: () => (localStorage.getItem('driverToken') ? '/home' : '/login')
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to, from, next) => {
  const driverToken = localStorage.getItem('driverToken') || ''

  if (to.meta.requiresDriverAuth && !driverToken) {
    next('/login')
  } else if (to.path === '/login' && driverToken) {
    next('/home')
  } else {
    next()
  }
})

export default router
