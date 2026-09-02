<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessageBox } from 'element-plus'
import { Odometer, User, Avatar, Tickets, Memo, Fold, Expand, Discount, List, Setting, Warning, DataAnalysis, Document, Management, Sunny, Moon, Money, Location } from '@element-plus/icons-vue'
import { logout, getMenus } from '../api/auth'
import { useUserStore } from '../store/user'
import { roleText, orDash } from '../utils/enums'

const route = useRoute()
const router = useRouter()
const store = useUserStore()
const collapsed = ref(false)
const isLightTheme = ref(!document.documentElement.classList.contains('dark-theme'))

// 后端菜单接口 path -> 前端路由映射
const pathToRoute = {
  '/users': '/users',
  '/driver-certifications': '/driver-certifications',
  '/drivers': '/drivers',
  '/orders': '/orders',
  '/orders/abnormal': '/orders/abnormal',
  '/operation-logs': '/operation-logs',
  '/admins': '/admins',
  '/capacity': '/capacity',
}

// 静态菜单全集仅用于超级管理员的故障兜底。
const fallbackMenus = [
  { name: '工作台', path: '/dashboard', icon: 'Odometer' },
  { name: '管理员管理', path: '/admins', icon: 'Management' },
  { name: '用户管理', path: '/users', icon: 'User' },
  { name: '司机审核', path: '/driver-certifications', icon: 'Avatar' },
  { name: '司机列表', path: '/drivers', icon: 'Avatar' },
  { name: '司机提现', path: '/driver-withdrawals', icon: 'Money' },
  { name: '订单管理', path: '/orders', icon: 'Tickets' },
  { name: '异常订单', path: '/orders/abnormal', icon: 'Warning' },
  { name: '操作日志', path: '/operation-logs', icon: 'Memo' },
  { name: '优惠券模板', path: '/coupons', icon: 'Discount' },
  { name: '发券任务', path: '/coupon-issue-tasks', icon: 'List' },
  { name: '计价规则', path: '/price-rules', icon: 'Setting' },
  { name: '活动配置', path: '/promotion-activities', icon: 'Discount' },
  { name: '投诉与申诉工单', path: '/work-orders', icon: 'Document' },
  { name: '数据统计', path: '/statistics', icon: 'DataAnalysis' },
  { name: '实时运力', path: '/capacity', icon: 'Location' },
  { name: '导出任务', path: '/export-tasks', icon: 'Document' },
  { name: '黑名单', path: '/blacklist', icon: 'Warning' },
  { name: '风控命中记录', path: '/risk-hits', icon: 'Management' },
]

// 菜单接口失败时，普通角色只展示与其日常查询职责匹配的基础入口。
// 这是安全兜底，不依赖前端传入的 role，也不将受限配置类模块暴露到导航。
const restrictedRoleFallbackPaths = {
  2: ['/users', '/driver-certifications', '/drivers', '/driver-withdrawals', '/orders', '/orders/abnormal', '/operation-logs', '/coupons', '/coupon-issue-tasks', '/price-rules', '/promotion-activities', '/work-orders', '/statistics', '/export-tasks', '/blacklist', '/risk-hits'],
  3: ['/users', '/driver-certifications', '/drivers', '/orders', '/orders/abnormal', '/operation-logs', '/work-orders'],
}

// 工作台是所有已登录管理员都可访问的固定入口。
const dashboardMenu = { name: '工作台', path: '/dashboard', icon: 'Odometer' }
const menuItems = ref([dashboardMenu])
const iconMap = { Odometer, User, Avatar, Tickets, Memo, Discount, List, Setting, Warning, DataAnalysis, Document, Management, Money, Location }

// 后端菜单不返回图标，前端按路径补齐，保证侧边栏视觉一致。
const pathIcon = {
  '/dashboard': 'Odometer',
  '/admins': 'Management',
  '/users': 'User',
  '/driver-certifications': 'Avatar',
  '/drivers': 'Avatar',
  '/driver-withdrawals': 'Money',
  '/orders': 'Tickets',
  '/orders/abnormal': 'Warning',
  '/operation-logs': 'Memo',
  '/coupons': 'Discount',
  '/coupon-issue-tasks': 'List',
  '/price-rules': 'Setting',
  '/promotion-activities': 'Discount',
  '/work-orders': 'Document',
  '/statistics': 'DataAnalysis',
  '/capacity': 'Location',
  '/export-tasks': 'Document',
  '/blacklist': 'Warning',
  '/risk-hits': 'Management',
}
const iconOf = (item) => iconMap[item.icon] || iconMap[pathIcon[item.path]] || null

// 侧边栏分组：仅影响视觉分区，不改变菜单来源和权限逻辑。
const groupOf = (path) => ({
  '/dashboard': '概览',
  '/users': '用户与司机',
  '/driver-certifications': '用户与司机',
  '/drivers': '用户与司机',
  '/driver-withdrawals': '用户与司机',
  '/orders': '订单中心',
  '/orders/abnormal': '订单中心',
  '/coupons': '营销中心',
  '/coupon-issue-tasks': '营销中心',
  '/price-rules': '营销中心',
  '/promotion-activities': '营销中心',
  '/work-orders': '工单与风控',
  '/blacklist': '工单与风控',
  '/risk-hits': '工单与风控',
  '/statistics': '数据与系统',
  '/capacity': '订单中心',
  '/export-tasks': '数据与系统',
  '/operation-logs': '数据与系统',
  '/admins': '数据与系统',
}[path] || '其他')

// menuGroups 按分组顺序聚合当前可见菜单，组内保持后端返回顺序。
const menuGroups = computed(() => {
  const order = ['概览', '用户与司机', '订单中心', '营销中心', '工单与风控', '数据与系统', '其他']
  const groups = new Map()
  menuItems.value.forEach((item) => {
    const name = groupOf(item.path)
    if (!groups.has(name)) groups.set(name, [])
    groups.get(name).push(item)
  })
  return order.filter((name) => groups.has(name)).map((name) => ({ name, items: groups.get(name) }))
})

// 返回按当前角色收敛后的静态兜底菜单。
// 超级管理员可以看到完整菜单；运营和客服只保留基础查询入口。
const getFallbackMenus = (role) => {
  if (role === 1) return fallbackMenus.filter((item) => item.path !== dashboardMenu.path)
  const allowedPaths = restrictedRoleFallbackPaths[role] || []
  return fallbackMenus.filter((item) => allowedPaths.includes(item.path))
}

// 登录后使用后端 /menus 渲染，后端返回结果是菜单权限的唯一来源。
onMounted(async () => {
  try {
    const data = await getMenus()
    const items = (data?.items || [])
      .map((i) => ({ name: i.name, path: pathToRoute[i.path] || i.path, icon: '' }))
      .filter((i) => router.resolve(i.path).matched.length > 0)
    menuItems.value = [dashboardMenu, ...items]
  } catch {
    // 菜单接口失败时只使用按当前角色收敛后的静态菜单，避免故障状态扩大可见权限。
    menuItems.value = [dashboardMenu, ...getFallbackMenus(store.admin?.role)]
  }
})

const activeMenu = computed(() => route.path)
const breadcrumbs = computed(() => {
  const crumbs = route.matched
    .filter((r) => r.meta?.title)
    .map((r) => ({ title: r.meta.title, path: r.path }))
  return crumbs.length > 0 ? crumbs : [{ title: '工作台', path: '/dashboard' }]
})

const handleSelect = (path) => {
  router.push(path)
}

const handleLogout = async () => {
  await ElMessageBox.confirm('确定退出登录吗？', '提示', { type: 'warning' })
  try {
    await logout()
  } finally {
    store.logout()
    router.push('/login')
  }
}

// 切换主题仅对当前会话生效，不写入本地存储；每次打开页面都恢复浅色。
const toggleTheme = () => {
  isLightTheme.value = !isLightTheme.value
  const theme = isLightTheme.value ? 'light' : 'dark'
  document.documentElement.classList.remove('dark-theme', 'light-theme')
  document.documentElement.classList.add(`${theme}-theme`)
}
</script>

<template>
  <el-container class="layout">
    <el-aside :width="collapsed ? '64px' : '232px'" class="aside">
      <div class="logo">
        <img class="logo-mark" src="/huaxiaolong-logo.png" alt="花小龙出行" /><span v-if="!collapsed" class="logo-text">花小龙出行<br><small>运营管理中心</small></span>
      </div>
      <el-menu
        :default-active="activeMenu"
        :collapse="collapsed"
        :collapse-transition="false"
        background-color="transparent"
        text-color="rgba(255,255,255,.72)"
        active-text-color="#584ac9"
        @select="handleSelect"
      >
        <template v-for="group in menuGroups" :key="group.name">
          <div v-if="!collapsed" class="menu-group">{{ group.name }}</div>
          <el-menu-item v-for="item in group.items" :key="item.path" :index="item.path">
            <el-icon v-if="iconOf(item)"><component :is="iconOf(item)" /></el-icon>
            <template #title>{{ item.name }}</template>
          </el-menu-item>
        </template>
      </el-menu>
      <el-button class="theme-toggle" text @click="toggleTheme">
        <el-icon><component :is="isLightTheme ? Moon : Sunny" /></el-icon>
        <span v-if="!collapsed">{{ isLightTheme ? '深色模式' : '浅色模式' }}</span>
      </el-button>
    </el-aside>

    <el-container>
      <el-header class="header">
        <div class="header-left">
          <el-icon class="fold-btn" @click="collapsed = !collapsed">
            <component :is="collapsed ? Expand : Fold" />
          </el-icon>
          <el-breadcrumb separator="/">
            <el-breadcrumb-item v-for="c in breadcrumbs" :key="c.path">{{ c.title }}</el-breadcrumb-item>
          </el-breadcrumb>
        </div>
        <div class="header-right">
          <span class="admin-avatar">{{ (store.admin?.real_name || store.admin?.username || '管').slice(0, 1) }}</span>
          <span class="admin-name">{{ orDash(store.admin?.real_name || store.admin?.username) }}</span>
          <el-tag size="small" effect="light">{{ roleText(store.admin?.role) }}</el-tag>
          <el-button link type="danger" @click="handleLogout">退出登录</el-button>
        </div>
      </el-header>
      <el-main class="main">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<style scoped>
.layout {
  height: 100vh;
}
/* 紫色侧边栏：渐变随主题切换，浅色鲜亮、深色沉稳。 */
.aside {
  background: var(--aside-grad, linear-gradient(180deg, #6a5ae2 0%, #5847c9 60%, #4f41bc 100%));
  transition: width 0.2s;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}
.logo {
  height: 60px;
  line-height: 60px;
  color: #ffffff;
  font-size: 16px;
  font-weight: 600;
  display:flex; align-items:center; gap:10px; padding:0 18px; text-align:left;
  white-space: nowrap;
}
.aside :deep(.el-menu) {
  border-right: none;
  background: transparent;
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  overflow-x: hidden;
  padding: 4px 0;
}
.aside :deep(.el-menu-item) {
  margin: 2px 10px;
  border-radius: 10px;
  color: rgba(255,255,255,.72);
}
.aside :deep(.el-menu-item:hover) {
  background: rgba(255,255,255,.12);
  color: #ffffff;
}
/* 选中项：白色圆角胶囊 + 紫色文字，贴合原型图。 */
.aside :deep(.el-menu-item.is-active) {
  background: #ffffff !important;
  color: #584ac9 !important;
  font-weight: 600;
  box-shadow: 0 4px 12px rgba(40,28,110,.25);
}
.aside :deep(.el-menu--collapse .el-menu-item) {
  margin: 2px 8px;
}
.menu-group {
  padding: 14px 22px 6px;
  font-size: 11px;
  letter-spacing: .12em;
  color: rgba(255,255,255,.45);
  white-space: nowrap;
}
.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid var(--border-color,#e5e4f0);
  background: var(--header-bg,#ffffff);
  color: var(--text-color,#2e2c4e);
  box-shadow: 0 1px 4px rgba(46,44,78,.05);
  z-index: 1;
}
.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}
.fold-btn {
  font-size: 20px;
  cursor: pointer;
}
.header-right {
  display: flex;
  align-items: center;
  gap: 10px;
}
.admin-avatar {
  width: 30px;
  height: 30px;
  border-radius: 50%;
  background: linear-gradient(135deg, #6c5ce7, #8b7ff0);
  color: #ffffff;
  font-size: 13px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}
.admin-name {
  font-size: 14px;
}
.main {
  background: var(--page-bg, #f2f3fa);
  padding: 20px;
  overflow: auto;
}
.logo-mark{width:36px;height:36px;border-radius:10px;object-fit:cover;background:#ffffff;display:block;box-shadow:0 2px 8px rgba(40,28,110,.3)}
.logo-text{font-weight:700;line-height:1.1}
.logo-text small{color:rgba(255,255,255,.6);font-weight:400;font-size:11px}
.theme-toggle{width:calc(100% - 24px);height:42px;flex:0 0 42px;margin:10px 12px 14px;color:rgba(255,255,255,.72);justify-content:flex-start;border-radius:10px}
.theme-toggle:hover{color:#ffffff;background:rgba(255,255,255,.14)}
.theme-toggle :deep(.el-icon){font-size:17px}
</style>
