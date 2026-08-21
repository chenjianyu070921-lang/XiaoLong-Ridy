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
      { path: 'dashboard', name: 'Dashboard', component: () => import('../views/console/index.vue'), meta: { title: '工作台', requiresAuth: true } },
      {
        path: 'users',
        name: 'UserList',
        component: () => import('../views/console/index.vue'),
        meta: { title: '用户管理', requiresAuth: true },
      },
      {
        path: 'users/:id',
        name: 'UserDetail',
        component: () => import('../views/console/index.vue'),
        meta: { title: '用户详情', requiresAuth: true },
      },
      {
        path: 'driver-certifications',
        name: 'DriverCertList',
        component: () => import('../views/console/index.vue'),
        meta: { title: '司机审核', requiresAuth: true },
      },
      {
        path: 'driver-certifications/:id',
        name: 'DriverCertDetail',
        component: () => import('../views/console/index.vue'),
        meta: { title: '审核详情', requiresAuth: true },
      },
      {
        path: 'orders',
        name: 'OrderList',
        component: () => import('../views/console/index.vue'),
        meta: { title: '订单管理', requiresAuth: true },
      },
      {
        path: 'orders/:id',
        name: 'OrderDetail',
        component: () => import('../views/console/index.vue'),
        meta: { title: '订单详情', requiresAuth: true },
      },
      {
        path: 'operation-logs',
        name: 'OperationLogs',
        component: () => import('../views/console/index.vue'),
        meta: { title: '操作日志', requiresAuth: true },
      },
      { path: 'orders/abnormal', name: 'AbnormalOrders', component: () => import('../views/console/index.vue'), meta: { title: '异常订单', requiresAuth: true } },
      { path: 'coupons', name: 'Coupons', component: () => import('../views/console/index.vue'), meta: { title: '优惠券模板', requiresAuth: true } },
      { path: 'coupon-issue-tasks', name: 'CouponTasks', component: () => import('../views/console/index.vue'), meta: { title: '发券任务', requiresAuth: true } },
      { path: 'price-rules', name: 'PriceRules', component: () => import('../views/console/index.vue'), meta: { title: '计价规则', requiresAuth: true } },
      { path: 'promotion-activities', name: 'Activities', component: () => import('../views/console/index.vue'), meta: { title: '活动配置', requiresAuth: true } },
      { path: 'work-orders', name: 'WorkOrders', component: () => import('../views/console/index.vue'), meta: { title: '投诉与申诉工单', requiresAuth: true } },
      { path: 'statistics', name: 'Statistics', component: () => import('../views/console/index.vue'), meta: { title: '数据统计', requiresAuth: true } },
      { path: 'export-tasks', name: 'Exports', component: () => import('../views/console/index.vue'), meta: { title: '导出任务', requiresAuth: true } },
      { path: 'export-tasks/:taskNo', name: 'ExportDetail', component: () => import('../views/console/index.vue'), meta: { title: '导出任务详情', requiresAuth: true } },
      { path: 'blacklist', name: 'Blacklist', component: () => import('../views/console/index.vue'), meta: { title: '黑名单', requiresAuth: true } },
      { path: 'risk-hits', name: 'RiskHits', component: () => import('../views/console/index.vue'), meta: { title: '风控命中记录', requiresAuth: true } },
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
