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
  },
  {
    path: '/profile/edit',
    name: 'DriverProfileEdit',
    component: () => import('@/views/DriverProfileEdit.vue'),
    meta: { requiresDriverAuth: true }
  },
  {
    path: '/mine/wallet',
    name: 'DriverMineWallet',
    component: () => import('@/views/mine/DriverWalletPage.vue'),
    meta: { requiresDriverAuth: true }
  },
  {
    path: '/mine/vehicle',
    name: 'DriverMineVehicle',
    component: () => import('@/views/mine/DriverVehiclePage.vue'),
    meta: { requiresDriverAuth: true }
  },
  {
    path: '/mine/certification',
    name: 'DriverMineCertification',
    component: () => import('@/views/mine/DriverCertificationPage.vue'),
    meta: { requiresDriverAuth: true }
  },
  {
    path: '/mine/income',
    name: 'DriverMineIncome',
    component: () => import('@/views/mine/DriverIncomePage.vue'),
    meta: { requiresDriverAuth: true }
  },
  {
    path: '/mine/orders',
    name: 'DriverMineOrders',
    component: () => import('@/views/mine/DriverOrderRecordsPage.vue'),
    meta: { requiresDriverAuth: true }
  },
  {
    path: '/mine/settings',
    name: 'DriverMineSettings',
    component: () => import('@/views/mine/DriverSettingsPage.vue'),
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
