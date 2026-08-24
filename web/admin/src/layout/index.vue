<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessageBox } from 'element-plus'
import { Odometer, User, Avatar, Tickets, Memo, Fold, Expand, Discount, List, Setting, Warning, DataAnalysis, Document, Management, Sunny, Moon } from '@element-plus/icons-vue'
import { logout, getMenus } from '../api/auth'
import { useUserStore } from '../store/user'
import { roleText, orDash } from '../utils/enums'

const route = useRoute()
const router = useRouter()
const store = useUserStore()
const collapsed = ref(false)
const isLightTheme = ref(document.documentElement.classList.contains('light-theme'))

// 后端菜单接口 path -> 前端路由映射
const pathToRoute = {
  '/users': '/users',
  '/driver-certifications': '/driver-certifications',
  '/drivers': '/drivers',
  '/orders': '/orders',
  '/orders/abnormal': '/orders/abnormal',
  '/operation-logs': '/operation-logs',
  '/admins': '/admins',
}

// 静态兜底菜单（菜单接口失败或为空时使用）
const fallbackMenus = [
  { name: '工作台', path: '/dashboard', icon: 'Odometer' },
  { name: '管理员管理', path: '/admins', icon: 'Management' },
  { name: '用户管理', path: '/users', icon: 'User' },
  { name: '司机审核', path: '/driver-certifications', icon: 'Avatar' },
  { name: '司机列表', path: '/drivers', icon: 'Avatar' },
  { name: '订单管理', path: '/orders', icon: 'Tickets' },
  { name: '异常订单', path: '/orders/abnormal', icon: 'Warning' },
  { name: '操作日志', path: '/operation-logs', icon: 'Memo' },
  { name: '优惠券模板', path: '/coupons', icon: 'Discount' },
  { name: '发券任务', path: '/coupon-issue-tasks', icon: 'List' },
  { name: '计价规则', path: '/price-rules', icon: 'Setting' },
  { name: '活动配置', path: '/promotion-activities', icon: 'Discount' },
  { name: '投诉与申诉工单', path: '/work-orders', icon: 'Document' },
  { name: '数据统计', path: '/statistics', icon: 'DataAnalysis' },
  { name: '导出任务', path: '/export-tasks', icon: 'Document' },
  { name: '黑名单', path: '/blacklist', icon: 'Warning' },
  { name: '风控命中记录', path: '/risk-hits', icon: 'Management' },
]

const menuItems = ref(fallbackMenus)
const iconMap = { Odometer, User, Avatar, Tickets, Memo, Discount, List, Setting, Warning, DataAnalysis, Document, Management }

// 登录后尝试用后端 /menus 渲染（工作台固定在最前）
onMounted(async () => {
  try {
    const data = await getMenus()
    const items = (data?.items || [])
      .map((i) => ({ name: i.name, path: pathToRoute[i.path] || i.path, icon: '' }))
      .filter((i) => router.resolve(i.path).matched.length > 0)
      if (items.length > 0) {
        menuItems.value = [{ name: '工作台', path: '/dashboard', icon: 'Odometer' }, ...items, ...fallbackMenus.slice(5)]
      }
  } catch {
    // 菜单接口失败时保留静态兜底菜单
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

// 切换并保存主题，供登录页、注册页和后台页面下次启动时恢复。
const toggleTheme = () => {
  isLightTheme.value = !isLightTheme.value
  const theme = isLightTheme.value ? 'light' : 'dark'
  document.documentElement.classList.remove('dark-theme', 'light-theme')
  document.documentElement.classList.add(`${theme}-theme`)
  localStorage.setItem('admin-theme', theme)
}
</script>

<template>
  <el-container class="layout">
    <el-aside :width="collapsed ? '64px' : '220px'" class="aside">
      <div class="logo">
        <img class="logo-mark" src="/huaxiaolong-logo.png" alt="花小龙出行" /><span v-if="!collapsed" class="logo-text">花小龙出行<br><small>运营管理中心</small></span>
      </div>
      <el-menu
        :default-active="activeMenu"
        :collapse="collapsed"
        :collapse-transition="false"
        background-color="#001529"
        text-color="#a6adb4"
        active-text-color="#ffffff"
        @select="handleSelect"
      >
      <el-menu-item v-for="item in menuItems" :key="item.path" :index="item.path">
          <el-icon v-if="iconMap[item.icon]"><component :is="iconMap[item.icon]" /></el-icon>
          <template #title>{{ item.name }}</template>
        </el-menu-item>
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
          <span class="admin-name">{{ orDash(store.admin?.real_name || store.admin?.username) }}</span>
          <el-tag size="small" type="info">{{ roleText(store.admin?.role) }}</el-tag>
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
.aside {
  background: var(--aside-bg, #08111d);
  transition: width 0.2s;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}
.logo {
  height: 56px;
  line-height: 56px;
  color: #f3f7fb;
  font-size: 16px;
  font-weight: 600;
  display:flex; align-items:center; gap:10px; padding:0 20px; text-align:left;
  white-space: nowrap;
}
.aside :deep(.el-menu) {
  border-right: none;
  background:var(--aside-bg, #08111d);
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  overflow-x: hidden;
}
.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid var(--border-color,#e4e7ed);
  background: var(--header-bg,#0c1724);
  color:var(--text-color,#dce7f3);
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
.admin-name {
  font-size: 14px;
}
.main {
  background: var(--page-bg, #070e17);
  padding: 16px;
  overflow: auto;
}
.logo-mark{width:34px;height:34px;border-radius:10px;object-fit:cover;background:#ff7625;display:block}.logo-text{font-weight:700;line-height:1.05}.logo-text small{color:#7d91a5;font-weight:400;font-size:11px}
.theme-toggle{width:calc(100% - 24px);height:42px;flex:0 0 42px;margin:10px 12px 14px;color:var(--muted-color,#a6adb4);justify-content:flex-start}.theme-toggle:hover{color:#ff8538;background:var(--active-bg,#1d2d3d)}
:global(:root.light-theme){--aside-bg:#ffffff}.theme-toggle :deep(.el-icon){font-size:17px}
</style>
