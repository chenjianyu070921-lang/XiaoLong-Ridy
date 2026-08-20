import { createRouter, createWebHashHistory } from 'vue-router'
import { useUserStore } from '../store/user'

// 路由 path 与后端 /admin/v1/menus 返回的 path 保持一致。
// hash 路由：build 后任意静态服务器可直接部署。
const routes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('../views/login/index.vue'),
    meta: { title: '登录' },
  },
  {
    path: '/',
    component: () => import('../layout/index.vue'),
    redirect: '/dashboard',
    children: [
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('../views/dashboard/index.vue'),
        meta: { title: '工作台', requiresAuth: true },
      },
      {
        path: 'users',
        name: 'UserList',
        component: () => import('../views/user/list.vue'),
        meta: { title: '用户管理', requiresAuth: true },
      },
      {
        path: 'users/:id',
        name: 'UserDetail',
        component: () => import('../views/user/detail.vue'),
        meta: { title: '用户详情', requiresAuth: true },
      },
      {
        path: 'driver-certifications',
        name: 'DriverCertList',
        component: () => import('../views/driver/list.vue'),
        meta: { title: '司机审核', requiresAuth: true },
      },
      {
        path: 'driver-certifications/:id',
        name: 'DriverCertDetail',
        component: () => import('../views/driver/detail.vue'),
        meta: { title: '审核详情', requiresAuth: true },
      },
      {
        path: 'orders',
        name: 'OrderList',
        component: () => import('../views/order/list.vue'),
        meta: { title: '订单管理', requiresAuth: true },
      },
      {
        path: 'orders/:id',
        name: 'OrderDetail',
        component: () => import('../views/order/detail.vue'),
        meta: { title: '订单详情', requiresAuth: true },
      },
      {
        path: 'operation-logs',
        name: 'OperationLogs',
        component: () => import('../views/log/list.vue'),
        meta: { title: '操作日志', requiresAuth: true },
      },
    ],
  },
  { path: '/:pathMatch(.*)*', redirect: '/dashboard' },
]

const router = createRouter({
  history: createWebHashHistory(),
  routes,
})

// 登录守卫：未登录访问受保护页面 -> /login；已登录访问 /login -> /dashboard
router.beforeEach((to) => {
  const store = useUserStore()
  if (to.meta.requiresAuth && !store.isLoggedIn) {
    return { path: '/login', query: { redirect: to.fullPath } }
  }
  if (to.path === '/login' && store.isLoggedIn) {
    return { path: '/dashboard' }
  }
  document.title = to.meta.title ? `${to.meta.title} - 小隆出行运营后台` : '小隆出行运营后台'
  return true
})

export default router
