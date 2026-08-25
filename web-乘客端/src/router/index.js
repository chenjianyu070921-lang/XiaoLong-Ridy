import { createRouter, createWebHistory } from 'vue-router'
import { useUserStore } from '@/stores/user'

const routes = [
  {
    path: '/',
    redirect: '/splash'
  },
  {
    path: '/splash',
    name: 'Splash',
    component: () => import('@/views/Splash.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/login/LoginPhone.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/login/phone',
    name: 'LoginPhone',
    component: () => import('@/views/login/LoginPhone.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/login/code',
    name: 'LoginCode',
    component: () => import('@/views/login/LoginCode.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/login/register',
    name: 'Register',
    component: () => import('@/views/login/Register.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/home',
    name: 'Home',
    component: () => import('@/views/home/Home.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/safety',
    name: 'Safety',
    component: () => import('@/views/Safety.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/order/create',
    name: 'OrderCreate',
    component: () => import('@/views/order/OrderCreate.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/order/waiting',
    name: 'OrderWaiting',
    component: () => import('@/views/order/OrderWaiting.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/order/driver-coming',
    name: 'DriverComing',
    component: () => import('@/views/order/DriverComing.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/order/trip',
    name: 'Trip',
    component: () => import('@/views/order/Trip.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/order/payment',
    name: 'Payment',
    component: () => import('@/views/order/Payment.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/order/success',
    name: 'PaymentSuccess',
    component: () => import('@/views/order/PaymentSuccess.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/orders',
    name: 'OrderList',
    component: () => import('@/views/orders/OrderList.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/orders/:id',
    name: 'OrderDetail',
    component: () => import('@/views/orders/OrderDetail.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/profile',
    name: 'Profile',
    component: () => import('@/views/profile/Profile.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/coupons',
    name: 'CouponList',
    component: () => import('@/views/profile/CouponList.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/wallet',
    name: 'Wallet',
    component: () => import('@/views/profile/Wallet.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/settings',
    name: 'Settings',
    component: () => import('@/views/profile/Settings.vue'),
    meta: { requiresAuth: true }
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// 路由守卫
router.beforeEach((to, from, next) => {
  const userStore = useUserStore()
  
  if (to.meta.requiresAuth && !userStore.isLoggedIn) {
    next('/login')
  } else if ((to.path === '/login' || to.path === '/splash') && userStore.isLoggedIn) {
    next('/home')
  } else {
    next()
  }
})

export default router

