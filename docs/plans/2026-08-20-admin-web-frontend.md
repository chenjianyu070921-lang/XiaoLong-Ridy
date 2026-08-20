# 管理后台前端（web-admin）实现计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 为已实现的管理后台接口（`api/admin`，127.0.0.1:8083）构建 Vue 3 + Element Plus 前端，跑通 P0 核心闭环（登录、工作台、用户、司机审核、订单、操作日志）。

**Architecture:** 新建 `web-admin/` 独立前端工程，Vite dev proxy 将 `/admin` 转发到后端 8083 规避 CORS。axios 统一封装拦截 `{code,message,data}` 响应与 Bearer Token；Pinia 管理登录态；hash 路由 + 守卫；布局为经典三段式。所有页面只对接真实后端，无 mock。

**Tech Stack:** Vue 3（`<script setup>`）、Element Plus、Vue Router（hash 模式）、Pinia、axios、dayjs、Vite。

**设计文档:** `docs/superpowers/specs/2026-08-20-admin-web-design.md`

---

## 已确认的后端契约（实现时必须严格遵守）

1. **统一响应**：`{code, message, data}`；`code=0` 成功；`40004` 未登录/token 失效；字段一律**蛇形**（如 `driver_phone`、`audit_status`）。
2. **请求体严格校验**：后端 `decodeJSON` 使用 `DisallowUnknownFields()`，POST body **不得传多余字段**：
   - 登录：`{username, password}`
   - 司机审核通过/驳回：`{remark}`
   - 订单取消：`{reason}`
   - 用户冻结/解封：`{reason, remark}`
3. **枚举值**（来自 SQL 迁移脚本与 proto）：
   - 订单状态：1待接单 2已接单 3行程中 4待支付 5已完成 6已取消
   - 司机状态：1待审核 2正常 3冻结 4注销；车辆状态：1待审核 2正常 3禁用
   - 审核状态：1待审核 2通过 3驳回
   - 用户状态：1正常 2冻结；性别：0未知 1男 2女
   - 管理员角色：1超管 2运营 3客服；车型：1特惠快车 2快车 3拼车
4. **菜单接口** `GET /admin/v1/menus` 返回 `{items:[{name,path,icon,perm,children}]}`，path 形如 `/users`、`/driver-certifications`、`/orders`、`/operation-logs`。
5. **分页**：query 参数 `page`、`page_size`（最大 100）；列表响应 `{list, total, page, page_size}`。
6. **金额字段为字符串**，前端原样展示。

## 关键技术决策

- **hash 路由**（`createWebHashHistory`）：`npm run build` 后任意静态服务器可直接部署，无需 fallback 配置。
- **request.js 跳转登录用 `window.location.hash = '#/login'`**：避免 request -> router -> views -> api -> request 的循环依赖。
- **菜单策略**：登录后调 `/menus`，取 path 能匹配前端路由的项渲染；接口失败或为空时用静态兜底菜单（工作台恒显）。
- **无单元测试**：纯 UI 对接工程，验证手段为 `npm run build` + dev 启动手动验收（验收清单见设计文档第 8 节）。

## 验证环境

- Node v24.11.1、npm 11.16.0（本机已确认）。
- 后端前置：`rpc/adminsvc`（8084）与 `api/admin`（8083）已启动；未启动时前端仍可构建，仅联调功能不可用。
- 测试账号：`admin/123456`（见验收清单）。

---

### Task 1: 工程脚手架

**Files:**
- Create: `web-admin/package.json`
- Create: `web-admin/vite.config.js`
- Create: `web-admin/index.html`
- Create: `web-admin/src/main.js`
- Create: `web-admin/src/App.vue`
- Create: `web-admin/.gitignore`

**Step 1: 创建工程文件**

`web-admin/package.json`：

```json
{
  "name": "xiaolong-admin-web",
  "private": true,
  "version": "0.1.0",
  "type": "module",
  "scripts": {
    "dev": "vite",
    "build": "vite build",
    "preview": "vite preview"
  },
  "dependencies": {
    "@element-plus/icons-vue": "^2.3.1",
    "axios": "^1.7.9",
    "dayjs": "^1.11.13",
    "element-plus": "^2.9.1",
    "pinia": "^2.3.0",
    "vue": "^3.5.13",
    "vue-router": "^4.5.0"
  },
  "devDependencies": {
    "@vitejs/plugin-vue": "^5.2.1",
    "vite": "^6.0.5"
  }
}
```

`web-admin/vite.config.js`：

```js
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// /admin 前缀的请求转发到本地 go-zero admin 网关，规避浏览器 CORS。
export default defineConfig({
  plugins: [vue()],
  server: {
    port: 5173,
    proxy: {
      '/admin': {
        target: 'http://127.0.0.1:8083',
        changeOrigin: true,
      },
    },
  },
})
```

`web-admin/index.html`：

```html
<!doctype html>
<html lang="zh-CN">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>小隆出行运营管理后台</title>
  </head>
  <body>
    <div id="app"></div>
    <script type="module" src="/src/main.js"></script>
  </body>
</html>
```

`web-admin/src/main.js`：

```js
import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import zhCn from 'element-plus/es/locale/lang/zh-cn'
import 'element-plus/dist/index.css'
import App from './App.vue'
import router from './router'

const app = createApp(App)
app.use(createPinia())
app.use(router)
app.use(ElementPlus, { locale: zhCn })
app.mount('#app')
```

`web-admin/src/App.vue`：

```vue
<template>
  <router-view />
</template>
```

`web-admin/.gitignore`：

```
node_modules
dist
*.local
```

注意：`src/router/index.js` 在 Task 3 才创建，本任务先创建一个最小占位以保证 `npm run dev` 可启动：

`web-admin/src/router/index.js`（占位，Task 3 会完整重写）：

```js
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
```

**Step 2: 安装依赖并验证 dev 启动**

```powershell
cd web-admin
npm install
npm run dev
```

预期：Vite 启动在 http://127.0.0.1:5173，浏览器访问显示 `web-admin ok`。Ctrl+C 停止。

**Step 3: 验证构建**

```powershell
npm run build
```

预期：构建成功，产出 `dist/`。

**Step 4: Commit**

```powershell
git add web-admin
git commit -m "feat(admin-web): 初始化 Vite + Vue3 + Element Plus 工程脚手架"
```

---

### Task 2: axios 统一封装 + 枚举工具

**Files:**
- Create: `web-admin/src/api/request.js`
- Create: `web-admin/src/utils/enums.js`

**Step 1: 写 request.js**

> 注意：后端 token 失效返回 **HTTP 401 + code 40004**（axios 默认把非 2xx 路由到错误拦截器），
> 所以会话失效处理必须放在错误分支；成功分支的 40004 判断仅作契约防御。

```js
import axios from 'axios'
import { ElMessage } from 'element-plus'

// 统一的后台 API 客户端：
// - baseURL 走相对路径，由 Vite dev proxy 转发到 api/admin(8083)，规避 CORS
// - 请求拦截器注入 Bearer Token
// - 响应拦截器解包 {code, message, data}，code!==0 统一报错，401/40004 跳登录
const service = axios.create({
  baseURL: '/admin/v1',
  timeout: 15000,
})

// 会话失效跳转防重入：并发请求同时 401 时只清理/跳转一次
let redirectingToLogin = false
const handleUnauthorized = () => {
  if (redirectingToLogin) return
  redirectingToLogin = true
  localStorage.removeItem('admin_token')
  localStorage.removeItem('admin_info')
  window.location.hash = '#/login'
  setTimeout(() => {
    redirectingToLogin = false
  }, 1000)
}

service.interceptors.request.use((config) => {
  const token = localStorage.getItem('admin_token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

service.interceptors.response.use(
  (response) => {
    const res = response.data
    if (res && res.code === 0) {
      return res.data
    }
    // 防御分支：若后端以 HTTP 200 返回 40004（当前实现不会，但保留契约兼容）
    if (res && res.code === 40004) {
      handleUnauthorized()
    }
    ElMessage.error(res?.message || '请求失败')
    return Promise.reject(new Error(res?.message || '请求失败'))
  },
  (error) => {
    // 后端 token 失效返回 HTTP 401 + code 40004；axios 默认把非 2xx 路由到本分支
    const data = error.response?.data
    if (error.response?.status === 401 || data?.code === 40004) {
      handleUnauthorized()
    }
    const msg = data?.message || error.message || '网络错误'
    ElMessage.error(msg)
    return Promise.reject(error)
  },
)

export default service
```

**Step 2: 写 enums.js**

```js
// 全局枚举 -> 中文文案映射。
// 数值来源：scripts/sql/migrate/*.sql 注释与 rpc/*/proto 定义。

export const orderStatusText = (s) => ({ 1: '待接单', 2: '已接单', 3: '行程中', 4: '待支付', 5: '已完成', 6: '已取消' }[s] || `未知(${s})`)
export const orderStatusTag = (s) => ({ 1: 'info', 2: 'primary', 3: 'warning', 4: 'warning', 5: 'success', 6: 'danger' }[s] || 'info')

export const userStatusText = (s) => ({ 1: '正常', 2: '冻结' }[s] || `未知(${s})`)
export const userStatusTag = (s) => ({ 1: 'success', 2: 'danger' }[s] || 'info')

export const driverStatusText = (s) => ({ 1: '待审核', 2: '正常', 3: '冻结', 4: '注销' }[s] || `未知(${s})`)
export const vehicleStatusText = (s) => ({ 1: '待审核', 2: '正常', 3: '禁用' }[s] || `未知(${s})`)
export const auditStatusText = (s) => ({ 1: '待审核', 2: '通过', 3: '驳回' }[s] || `未知(${s})`)
export const auditStatusTag = (s) => ({ 1: 'warning', 2: 'success', 3: 'danger' }[s] || 'info')

export const genderText = (g) => ({ 0: '未知', 1: '男', 2: '女' }[g] || '未知')
export const carTypeText = (t) => ({ 1: '特惠快车', 2: '快车', 3: '拼车' }[t] || `未知(${t})`)
export const roleText = (r) => ({ 1: '超级管理员', 2: '运营', 3: '客服' }[r] || `未知(${r})`)

// 通用：空值展示占位
export const orDash = (v) => (v === null || v === undefined || v === '' ? '-' : v)
```

**Step 3: 验证构建**

```powershell
cd web-admin
npm run build
```

预期：构建成功。

**Step 4: Commit**

```powershell
git add web-admin/src
git commit -m "feat(admin-web): axios 统一封装与全局枚举映射"
```

---

### Task 3: Pinia store + 路由与守卫

**Files:**
- Create: `web-admin/src/store/user.js`
- Modify: `web-admin/src/router/index.js`（完整重写占位文件）

**Step 1: 写 store/user.js**

```js
import { defineStore } from 'pinia'

// 管理员登录态：token + 管理员信息 + 菜单。
// 持久化到 localStorage，刷新页面不丢登录态。
export const useUserStore = defineStore('user', {
  state: () => ({
    token: localStorage.getItem('admin_token') || '',
    admin: JSON.parse(localStorage.getItem('admin_info') || 'null'),
    menus: [],
  }),
  getters: {
    isLoggedIn: (state) => !!state.token,
  },
  actions: {
    setLogin(token, admin) {
      this.token = token
      this.admin = admin
      localStorage.setItem('admin_token', token)
      localStorage.setItem('admin_info', JSON.stringify(admin || null))
    },
    setMenus(menus) {
      this.menus = menus || []
    },
    logout() {
      this.token = ''
      this.admin = null
      this.menus = []
      localStorage.removeItem('admin_token')
      localStorage.removeItem('admin_info')
    },
  },
})
```

**Step 2: 完整重写 router/index.js**

```js
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
```

**Step 3: 创建各页面占位文件**（保证路由可用，后续任务逐个实现）

每个文件内容相同（替换注释中的页面名）：

```vue
<template>
  <div>页面开发中：XXX</div>
</template>
```

需要创建：`src/views/login/index.vue`、`src/views/dashboard/index.vue`、`src/views/user/list.vue`、`src/views/user/detail.vue`、`src/views/driver/list.vue`、`src/views/driver/detail.vue`、`src/views/order/list.vue`、`src/views/order/detail.vue`、`src/views/log/list.vue`。

**Step 4: 验证**

```powershell
cd web-admin
npm run build
npm run dev
```

预期：构建成功；dev 启动后访问 http://127.0.0.1:5173 自动跳到 `#/login` 显示占位。

**Step 5: Commit**

```powershell
git add web-admin/src
git commit -m "feat(admin-web): Pinia 登录态、hash 路由与登录守卫"
```

---

### Task 4: 后台整体布局

**Files:**
- Create: `web-admin/src/layout/index.vue`
- Create: `web-admin/src/api/auth.js`

**Step 1: 写 api/auth.js**

```js
import request from './request'

// 登录：返回 {token, expires_in, admin}
export const login = (data) => request.post('/auth/login', data)
// 退出登录
export const logout = () => request.post('/auth/logout', {})
// 当前管理员信息：返回 {admin}
export const me = () => request.get('/auth/me')
// 角色菜单：返回 {items: [{name, path, icon, perm, children}]}
export const getMenus = () => request.get('/menus')
```

**Step 2: 写 layout/index.vue**

```vue
<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessageBox } from 'element-plus'
import { Odometer, User, Avatar, ClipboardList, Memo, Fold, Expand } from '@element-plus/icons-vue'
import { logout, getMenus } from '../api/auth'
import { useUserStore } from '../store/user'
import { roleText, orDash } from '../utils/enums'

const route = useRoute()
const router = useRouter()
const store = useUserStore()
const collapsed = ref(false)

// 后端菜单接口 path -> 前端路由映射
const pathToRoute = {
  '/users': '/users',
  '/driver-certifications': '/driver-certifications',
  '/orders': '/orders',
  '/operation-logs': '/operation-logs',
  '/admins': '/dashboard', // 后端管理员菜单暂无对应页面，映射到工作台
}

// 静态兜底菜单（菜单接口失败或为空时使用）
const fallbackMenus = [
  { name: '工作台', path: '/dashboard', icon: 'Odometer' },
  { name: '用户管理', path: '/users', icon: 'User' },
  { name: '司机审核', path: '/driver-certifications', icon: 'Avatar' },
  { name: '订单管理', path: '/orders', icon: 'ClipboardList' },
  { name: '操作日志', path: '/operation-logs', icon: 'Memo' },
]

const menuItems = ref(fallbackMenus)
const iconMap = { Odometer, User, Avatar, ClipboardList, Memo }

// 登录后尝试用后端 /menus 渲染（工作台固定在最前）
onMounted(async () => {
  try {
    const data = await getMenus()
    const items = (data?.items || [])
      .map((i) => ({ name: i.name, path: pathToRoute[i.path] || i.path, icon: '' }))
      .filter((i) => router.resolve(i.path).matched.length > 0)
    if (items.length > 0) {
      menuItems.value = [{ name: '工作台', path: '/dashboard', icon: 'Odometer' }, ...items]
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
</script>

<template>
  <el-container class="layout">
    <el-aside :width="collapsed ? '64px' : '220px'" class="aside">
      <div class="logo">
        <span v-if="!collapsed">小隆出行 · 运营后台</span>
        <span v-else>小隆</span>
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
  background: #001529;
  transition: width 0.2s;
  overflow: hidden;
}
.logo {
  height: 56px;
  line-height: 56px;
  color: #fff;
  font-size: 16px;
  font-weight: 600;
  text-align: center;
  white-space: nowrap;
}
.aside :deep(.el-menu) {
  border-right: none;
}
.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid #e4e7ed;
  background: #fff;
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
  background: #f0f2f5;
  padding: 16px;
  overflow: auto;
}
</style>
```

**Step 3: 验证**

```powershell
cd web-admin
npm run build
npm run dev
```

预期：构建成功。因未登录，访问根路径仍跳登录页（守卫生效）。布局需登录后才能看到（下一任务）。

**Step 4: Commit**

```powershell
git add web-admin/src
git commit -m "feat(admin-web): 三段式后台布局（侧边菜单/顶栏/面包屑）"
```

---

### Task 5: 登录页

**Files:**
- Modify: `web-admin/src/views/login/index.vue`（替换占位）

**Step 1: 实现登录页**

```vue
<script setup>
import { reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { User, Lock } from '@element-plus/icons-vue'
import { login } from '../../api/auth'
import { useUserStore } from '../../store/user'

const router = useRouter()
const route = useRoute()
const store = useUserStore()
const formRef = ref()
const loading = ref(false)

const form = reactive({
  username: '',
  password: '',
})

const rules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
}

const handleLogin = async () => {
  await formRef.value.validate()
  loading.value = true
  try {
    // 后端契约：{token, expires_in, admin{id,username,real_name,role,status}}
    const data = await login({ username: form.username, password: form.password })
    store.setLogin(data.token, data.admin)
    ElMessage.success('登录成功')
    router.push(route.query.redirect || '/dashboard')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="login-page">
    <el-card class="login-card">
      <h2 class="title">小隆出行运营管理后台</h2>
      <el-form ref="formRef" :model="form" :rules="rules" size="large" @keyup.enter="handleLogin">
        <el-form-item prop="username">
          <el-input v-model="form.username" placeholder="用户名" :prefix-icon="User" />
        </el-form-item>
        <el-form-item prop="password">
          <el-input v-model="form.password" type="password" placeholder="密码" :prefix-icon="Lock" show-password />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" style="width: 100%" :loading="loading" @click="handleLogin">登 录</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<style scoped>
.login-page {
  height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #1f3a5f 0%, #0f2027 100%);
}
.login-card {
  width: 380px;
}
.title {
  text-align: center;
  margin: 0 0 24px;
  color: #303133;
}
</style>
```

**Step 2: 联调验证（后端已启动时）**

```powershell
cd web-admin
npm run dev
```

预期：`admin/123456` 登录成功进入工作台（占位页），侧边栏菜单、顶栏管理员信息正常显示；错误密码有报错提示。刷新页面登录态保持。

**Step 3: Commit**

```powershell
git add web-admin/src
git commit -m "feat(admin-web): 管理员登录页与登录态流转"
```

---

### Task 6: 工作台（数据总览）

**Files:**
- Create: `web-admin/src/api/statistics.js`
- Modify: `web-admin/src/views/dashboard/index.vue`（替换占位）

**Step 1: 写 api/statistics.js**

```js
import request from './request'

// 运营总览：{user_count, driver_count, order_count, completed_order_count, abnormal_order_count, gmv, coupon_issue_count, blacklist_count}
export const getOverview = (params) => request.get('/statistics/overview', { params })
// 订单统计：{order_count, completed_order_count, canceled_order_count, timeout_order_count, payment_abnormal_count, completion_rate, cancel_rate}
export const getOrderStatistics = (params) => request.get('/statistics/orders', { params })
```

**Step 2: 实现工作台**

```vue
<script setup>
import { onMounted, ref } from 'vue'
import dayjs from 'dayjs'
import { getOverview, getOrderStatistics } from '../../api/statistics'
import { orDash } from '../../utils/enums'

const loading = ref(false)
const range = ref([])
const overview = ref(null)
const orderStats = ref(null)

const overviewCards = [
  { key: 'user_count', label: '注册用户' },
  { key: 'driver_count', label: '注册司机' },
  { key: 'order_count', label: '订单总量' },
  { key: 'completed_order_count', label: '完成订单' },
  { key: 'abnormal_order_count', label: '异常订单' },
  { key: 'gmv', label: 'GMV（元）' },
  { key: 'coupon_issue_count', label: '发券量' },
  { key: 'blacklist_count', label: '黑名单' },
]

const orderCards = [
  { key: 'order_count', label: '订单量' },
  { key: 'completed_order_count', label: '完成订单' },
  { key: 'canceled_order_count', label: '取消订单' },
  { key: 'timeout_order_count', label: '超时订单' },
  { key: 'payment_abnormal_count', label: '支付异常' },
  { key: 'completion_rate', label: '完成率', suffix: '%' },
  { key: 'cancel_rate', label: '取消率', suffix: '%' },
]

const buildParams = () => {
  const params = {}
  if (range.value && range.value.length === 2) {
    params.start_time = dayjs(range.value[0]).format('YYYY-MM-DD 00:00:00')
    params.end_time = dayjs(range.value[1]).format('YYYY-MM-DD 23:59:59')
  }
  return params
}

const fetchData = async () => {
  loading.value = true
  try {
    const params = buildParams()
    const [o, s] = await Promise.all([getOverview(params), getOrderStatistics(params)])
    overview.value = o
    orderStats.value = s
  } finally {
    loading.value = false
  }
}

onMounted(fetchData)
</script>

<template>
  <div v-loading="loading">
    <el-card shadow="never" class="filter-card">
      <el-date-picker
        v-model="range"
        type="daterange"
        range-separator="至"
        start-placeholder="开始日期"
        end-placeholder="结束日期"
        value-format="YYYY-MM-DD"
        clearable
      />
      <el-button type="primary" style="margin-left: 12px" @click="fetchData">查询</el-button>
    </el-card>

    <el-card shadow="never" header="运营总览">
      <el-empty v-if="!overview" description="暂无数据" />
      <el-row v-else :gutter="16">
        <el-col v-for="c in overviewCards" :key="c.key" :span="6" class="stat-col">
          <div class="stat-item">
            <div class="stat-label">{{ c.label }}</div>
            <div class="stat-value">{{ orDash(overview[c.key]) }}</div>
          </div>
        </el-col>
      </el-row>
    </el-card>

    <el-card shadow="never" header="订单统计" style="margin-top: 16px">
      <el-empty v-if="!orderStats" description="暂无数据" />
      <el-row v-else :gutter="16">
        <el-col v-for="c in orderCards" :key="c.key" :span="6" class="stat-col">
          <div class="stat-item">
            <div class="stat-label">{{ c.label }}</div>
            <div class="stat-value">{{ orDash(orderStats[c.key]) }}{{ orderStats[c.key] != null ? (c.suffix || '') : '' }}</div>
          </div>
        </el-col>
      </el-row>
    </el-card>
  </div>
</template>

<style scoped>
.filter-card {
  margin-bottom: 16px;
}
.stat-col {
  margin-bottom: 16px;
}
.stat-item {
  background: #f7f8fa;
  border-radius: 8px;
  padding: 16px;
}
.stat-label {
  color: #909399;
  font-size: 13px;
}
.stat-value {
  color: #303133;
  font-size: 24px;
  font-weight: 600;
  margin-top: 8px;
}
</style>
```

**Step 3: 验证**

```powershell
cd web-admin
npm run build && npm run dev
```

预期：构建成功；登录后工作台展示统计卡片（后端有数据时显示数值，无数据时显示 0 或空态）。

**Step 4: Commit**

```powershell
git add web-admin/src
git commit -m "feat(admin-web): 工作台运营总览与订单统计卡片"
```

---

### Task 7: 用户管理（列表 / 详情 / 冻结 / 解封）

**Files:**
- Create: `web-admin/src/api/user.js`
- Modify: `web-admin/src/views/user/list.vue`（替换占位）
- Modify: `web-admin/src/views/user/detail.vue`（替换占位）

**Step 1: 写 api/user.js**

```js
import request from './request'

// 用户列表：{list, total, page, page_size}
export const listUsers = (params) => request.get('/users', { params })
// 用户详情
export const getUser = (id) => request.get(`/users/${id}`)
// 冻结/解封：body 只允许 reason 和 remark 字段（后端 DisallowUnknownFields）
export const freezeUser = (id, data) => request.post(`/users/${id}/freeze`, data)
export const unfreezeUser = (id, data) => request.post(`/users/${id}/unfreeze`, data)
```

**Step 2: 实现用户列表页** `src/views/user/list.vue`

```vue
<script setup>
import { onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import dayjs from 'dayjs'
import { listUsers, freezeUser, unfreezeUser } from '../../api/user'
import { userStatusText, userStatusTag, genderText, orDash } from '../../utils/enums'

const router = useRouter()
const loading = ref(false)
const list = ref([])
const total = ref(0)

const query = reactive({
  page: 1,
  page_size: 20,
  keyword: '',
  status: 0,
  range: [],
})

const statusOptions = [
  { label: '全部状态', value: 0 },
  { label: '正常', value: 1 },
  { label: '冻结', value: 2 },
]

const buildParams = () => {
  const params = { page: query.page, page_size: query.page_size, keyword: query.keyword, status: query.status }
  if (query.range && query.range.length === 2) {
    params.start_time = dayjs(query.range[0]).format('YYYY-MM-DD 00:00:00')
    params.end_time = dayjs(query.range[1]).format('YYYY-MM-DD 23:59:59')
  }
  return params
}

const fetchData = async () => {
  loading.value = true
  try {
    const data = await listUsers(buildParams())
    list.value = data?.list || []
    total.value = Number(data?.total || 0)
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  query.page = 1
  fetchData()
}

const handleReset = () => {
  query.keyword = ''
  query.status = 0
  query.range = []
  handleSearch()
}

// 冻结/解封：二次确认 + 必填原因
const handleChangeStatus = async (row, action) => {
  const isFreeze = action === 'freeze'
  const label = isFreeze ? '冻结' : '解封'
  const { value } = await ElMessageBox.prompt(`请输入${label}原因`, `${label}用户「${row.phone}」`, {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    inputPlaceholder: '原因（必填）',
    inputValidator: (v) => (v && v.trim() ? true : '原因不能为空'),
    type: 'warning',
  })
  const api = isFreeze ? freezeUser : unfreezeUser
  await api(row.id, { reason: value.trim(), remark: '' })
  ElMessage.success(`${label}成功`)
  fetchData()
}

onMounted(fetchData)
</script>

<template>
  <el-card shadow="never">
    <el-form inline @submit.prevent>
      <el-form-item label="关键词">
        <el-input v-model="query.keyword" placeholder="手机号/昵称/实名" clearable style="width: 200px" @keyup.enter="handleSearch" />
      </el-form-item>
      <el-form-item label="状态">
        <el-select v-model="query.status" style="width: 120px">
          <el-option v-for="o in statusOptions" :key="o.value" :label="o.label" :value="o.value" />
        </el-select>
      </el-form-item>
      <el-form-item label="注册时间">
        <el-date-picker v-model="query.range" type="daterange" range-separator="至" start-placeholder="开始日期" end-placeholder="结束日期" value-format="YYYY-MM-DD" clearable />
      </el-form-item>
      <el-form-item>
        <el-button type="primary" @click="handleSearch">查询</el-button>
        <el-button @click="handleReset">重置</el-button>
      </el-form-item>
    </el-form>

    <el-table v-loading="loading" :data="list">
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="phone" label="手机号" width="130" />
      <el-table-column label="昵称" min-width="120">
        <template #default="{ row }">{{ orDash(row.nickname) }}</template>
      </el-table-column>
      <el-table-column label="实名" width="100">
        <template #default="{ row }">{{ orDash(row.real_name) }}</template>
      </el-table-column>
      <el-table-column label="性别" width="70">
        <template #default="{ row }">{{ genderText(row.gender) }}</template>
      </el-table-column>
      <el-table-column label="状态" width="90">
        <template #default="{ row }">
          <el-tag :type="userStatusTag(row.status)" size="small">{{ userStatusText(row.status) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="created_at" label="注册时间" width="170" />
      <el-table-column label="操作" width="200" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="router.push(`/users/${row.id}`)">详情</el-button>
          <el-button v-if="row.status === 1" link type="danger" @click="handleChangeStatus(row, 'freeze')">冻结</el-button>
          <el-button v-if="row.status === 2" link type="success" @click="handleChangeStatus(row, 'unfreeze')">解封</el-button>
        </template>
      </el-table-column>
      <template #empty>
        <el-empty description="暂无用户数据" />
      </template>
    </el-table>

    <el-pagination
      v-model:current-page="query.page"
      v-model:page-size="query.page_size"
      :total="total"
      :page-sizes="[10, 20, 50, 100]"
      layout="total, sizes, prev, pager, next, jumper"
      style="margin-top: 16px; justify-content: flex-end"
      @change="fetchData"
    />
  </el-card>
</template>
```

**Step 3: 实现用户详情页** `src/views/user/detail.vue`

```vue
<script setup>
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getUser } from '../../api/user'
import { userStatusText, userStatusTag, genderText, orDash } from '../../utils/enums'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const user = ref(null)

onMounted(async () => {
  loading.value = true
  try {
    user.value = await getUser(route.params.id)
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <el-card shadow="never" v-loading="loading">
    <template #header>
      <div class="card-header">
        <span>用户详情</span>
        <el-button @click="router.back()">返回</el-button>
      </div>
    </template>
    <el-empty v-if="!user && !loading" description="用户不存在" />
    <el-descriptions v-else-if="user" :column="2" border>
      <el-descriptions-item label="用户ID">{{ user.id }}</el-descriptions-item>
      <el-descriptions-item label="手机号">{{ orDash(user.phone) }}</el-descriptions-item>
      <el-descriptions-item label="昵称">{{ orDash(user.nickname) }}</el-descriptions-item>
      <el-descriptions-item label="性别">{{ genderText(user.gender) }}</el-descriptions-item>
      <el-descriptions-item label="实名">{{ orDash(user.real_name) }}</el-descriptions-item>
      <el-descriptions-item label="身份证号">{{ orDash(user.id_card_no) }}</el-descriptions-item>
      <el-descriptions-item label="注册来源">{{ orDash(user.register_source) }}</el-descriptions-item>
      <el-descriptions-item label="状态">
        <el-tag :type="userStatusTag(user.status)" size="small">{{ userStatusText(user.status) }}</el-tag>
      </el-descriptions-item>
      <el-descriptions-item label="注册时间">{{ orDash(user.created_at) }}</el-descriptions-item>
      <el-descriptions-item label="更新时间">{{ orDash(user.updated_at) }}</el-descriptions-item>
    </el-descriptions>
  </el-card>
</template>

<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
```

**Step 4: 验证**

```powershell
cd web-admin
npm run build && npm run dev
```

预期：用户列表可查、筛选、分页；点详情进入详情页；冻结弹二次确认+必填原因；后端无数据时显示空态。

**Step 5: Commit**

```powershell
git add web-admin/src
git commit -m "feat(admin-web): 用户列表/详情/冻结/解封"
```

---

### Task 8: 司机审核（列表 / 详情 / 通过 / 驳回）

**Files:**
- Create: `web-admin/src/api/driver.js`
- Modify: `web-admin/src/views/driver/list.vue`（替换占位）
- Modify: `web-admin/src/views/driver/detail.vue`（替换占位）

**Step 1: 写 api/driver.js**

```js
import request from './request'

// 审核列表：{list, total, page, page_size}
export const listCertifications = (params) => request.get('/driver-certifications', { params })
// 审核详情
export const getCertification = (id) => request.get(`/driver-certifications/${id}`)
// 通过/驳回：body 只允许 remark 字段
export const approveCertification = (id, data) => request.post(`/driver-certifications/${id}/approve`, data)
export const rejectCertification = (id, data) => request.post(`/driver-certifications/${id}/reject`, data)
```

**Step 2: 实现审核列表页** `src/views/driver/list.vue`

```vue
<script setup>
import { onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import dayjs from 'dayjs'
import { listCertifications, approveCertification, rejectCertification } from '../../api/driver'
import { auditStatusText, auditStatusTag, orDash } from '../../utils/enums'

const router = useRouter()
const loading = ref(false)
const list = ref([])
const total = ref(0)

const query = reactive({
  page: 1,
  page_size: 20,
  keyword: '',
  audit_status: 0,
  range: [],
})

const statusOptions = [
  { label: '全部状态', value: 0 },
  { label: '待审核', value: 1 },
  { label: '通过', value: 2 },
  { label: '驳回', value: 3 },
]

const buildParams = () => {
  const params = { page: query.page, page_size: query.page_size, keyword: query.keyword, audit_status: query.audit_status }
  if (query.range && query.range.length === 2) {
    params.start_time = dayjs(query.range[0]).format('YYYY-MM-DD 00:00:00')
    params.end_time = dayjs(query.range[1]).format('YYYY-MM-DD 23:59:59')
  }
  return params
}

const fetchData = async () => {
  loading.value = true
  try {
    const data = await listCertifications(buildParams())
    list.value = data?.list || []
    total.value = Number(data?.total || 0)
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  query.page = 1
  fetchData()
}

const handleReset = () => {
  query.keyword = ''
  query.audit_status = 0
  query.range = []
  handleSearch()
}

// 审核通过：二次确认，备注选填
const handleApprove = async (row) => {
  const { value } = await ElMessageBox.prompt('可填写审核备注（选填）', `通过「${row.driver_name || row.driver_phone}」的认证`, {
    confirmButtonText: '确认通过',
    cancelButtonText: '取消',
    inputPlaceholder: '审核备注（选填）',
    inputValue: '',
    type: 'success',
  })
  await approveCertification(row.id, { remark: (value || '').trim() })
  ElMessage.success('已通过')
  fetchData()
}

// 审核驳回：二次确认，备注必填（作为驳回原因）
const handleReject = async (row) => {
  const { value } = await ElMessageBox.prompt('请填写驳回原因', `驳回「${row.driver_name || row.driver_phone}」的认证`, {
    confirmButtonText: '确认驳回',
    cancelButtonText: '取消',
    inputPlaceholder: '驳回原因（必填）',
    inputValidator: (v) => (v && v.trim() ? true : '驳回原因不能为空'),
    type: 'warning',
  })
  await rejectCertification(row.id, { remark: value.trim() })
  ElMessage.success('已驳回')
  fetchData()
}

onMounted(fetchData)
</script>

<template>
  <el-card shadow="never">
    <el-form inline @submit.prevent>
      <el-form-item label="关键词">
        <el-input v-model="query.keyword" placeholder="司机姓名/手机号/车牌" clearable style="width: 220px" @keyup.enter="handleSearch" />
      </el-form-item>
      <el-form-item label="审核状态">
        <el-select v-model="query.audit_status" style="width: 120px">
          <el-option v-for="o in statusOptions" :key="o.value" :label="o.label" :value="o.value" />
        </el-select>
      </el-form-item>
      <el-form-item label="提交时间">
        <el-date-picker v-model="query.range" type="daterange" range-separator="至" start-placeholder="开始日期" end-placeholder="结束日期" value-format="YYYY-MM-DD" clearable />
      </el-form-item>
      <el-form-item>
        <el-button type="primary" @click="handleSearch">查询</el-button>
        <el-button @click="handleReset">重置</el-button>
      </el-form-item>
    </el-form>

    <el-table v-loading="loading" :data="list">
      <el-table-column prop="id" label="审核ID" width="80" />
      <el-table-column prop="driver_name" label="司机姓名" width="100">
        <template #default="{ row }">{{ orDash(row.driver_name) }}</template>
      </el-table-column>
      <el-table-column prop="driver_phone" label="手机号" width="130" />
      <el-table-column prop="plate_no" label="车牌号" width="110">
        <template #default="{ row }">{{ orDash(row.plate_no) }}</template>
      </el-table-column>
      <el-table-column label="审核状态" width="100">
        <template #default="{ row }">
          <el-tag :type="auditStatusTag(row.audit_status)" size="small">{{ auditStatusText(row.audit_status) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="audit_remark" label="审核备注" min-width="140">
        <template #default="{ row }">{{ orDash(row.audit_remark) }}</template>
      </el-table-column>
      <el-table-column prop="created_at" label="提交时间" width="170" />
      <el-table-column label="操作" width="200" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="router.push(`/driver-certifications/${row.id}`)">详情</el-button>
          <template v-if="row.audit_status === 1">
            <el-button link type="success" @click="handleApprove(row)">通过</el-button>
            <el-button link type="danger" @click="handleReject(row)">驳回</el-button>
          </template>
        </template>
      </el-table-column>
      <template #empty>
        <el-empty description="暂无审核数据" />
      </template>
    </el-table>

    <el-pagination
      v-model:current-page="query.page"
      v-model:page-size="query.page_size"
      :total="total"
      :page-sizes="[10, 20, 50, 100]"
      layout="total, sizes, prev, pager, next, jumper"
      style="margin-top: 16px; justify-content: flex-end"
      @change="fetchData"
    />
  </el-card>
</template>
```

**Step 3: 实现审核详情页** `src/views/driver/detail.vue`

```vue
<script setup>
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getCertification, approveCertification, rejectCertification } from '../../api/driver'
import { auditStatusText, auditStatusTag, driverStatusText, vehicleStatusText, orDash } from '../../utils/enums'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const cert = ref(null)

const fetchData = async () => {
  loading.value = true
  try {
    cert.value = await getCertification(route.params.id)
  } finally {
    loading.value = false
  }
}

const handleApprove = async () => {
  const { value } = await ElMessageBox.prompt('可填写审核备注（选填）', '确认通过该司机认证', {
    confirmButtonText: '确认通过',
    cancelButtonText: '取消',
    inputPlaceholder: '审核备注（选填）',
    inputValue: '',
    type: 'success',
  })
  await approveCertification(cert.value.id, { remark: (value || '').trim() })
  ElMessage.success('已通过')
  fetchData()
}

const handleReject = async () => {
  const { value } = await ElMessageBox.prompt('请填写驳回原因', '确认驳回该司机认证', {
    confirmButtonText: '确认驳回',
    cancelButtonText: '取消',
    inputPlaceholder: '驳回原因（必填）',
    inputValidator: (v) => (v && v.trim() ? true : '驳回原因不能为空'),
    type: 'warning',
  })
  await rejectCertification(cert.value.id, { remark: value.trim() })
  ElMessage.success('已驳回')
  fetchData()
}

onMounted(fetchData)
</script>

<template>
  <el-card shadow="never" v-loading="loading">
    <template #header>
      <div class="card-header">
        <span>司机认证详情</span>
        <div>
          <el-button v-if="cert?.audit_status === 1" type="success" @click="handleApprove">通过</el-button>
          <el-button v-if="cert?.audit_status === 1" type="danger" @click="handleReject">驳回</el-button>
          <el-button @click="router.back()">返回</el-button>
        </div>
      </div>
    </template>

    <el-empty v-if="!cert && !loading" description="认证记录不存在" />

    <template v-else-if="cert">
      <el-descriptions title="基础信息" :column="2" border style="margin-bottom: 16px">
        <el-descriptions-item label="审核ID">{{ cert.id }}</el-descriptions-item>
        <el-descriptions-item label="审核状态">
          <el-tag :type="auditStatusTag(cert.audit_status)" size="small">{{ auditStatusText(cert.audit_status) }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="司机ID">{{ cert.driver_id }}</el-descriptions-item>
        <el-descriptions-item label="车辆ID">{{ orDash(cert.vehicle_id) }}</el-descriptions-item>
        <el-descriptions-item label="司机姓名">{{ orDash(cert.driver_name) }}</el-descriptions-item>
        <el-descriptions-item label="手机号">{{ orDash(cert.driver_phone) }}</el-descriptions-item>
        <el-descriptions-item label="司机状态">{{ driverStatusText(cert.driver_status) }}</el-descriptions-item>
        <el-descriptions-item label="车牌号">{{ orDash(cert.plate_no) }}</el-descriptions-item>
        <el-descriptions-item label="车辆状态">{{ vehicleStatusText(cert.vehicle_status) }}</el-descriptions-item>
        <el-descriptions-item label="提交时间">{{ orDash(cert.created_at) }}</el-descriptions-item>
      </el-descriptions>

      <el-descriptions title="证照材料" :column="4" border style="margin-bottom: 16px">
        <el-descriptions-item label="身份证人像面">
          <el-image v-if="cert.id_card_front_url" :src="cert.id_card_front_url" :preview-src-list="[cert.id_card_front_url]" fit="cover" style="width: 100px; height: 70px" />
          <span v-else>无</span>
        </el-descriptions-item>
        <el-descriptions-item label="身份证国徽面">
          <el-image v-if="cert.id_card_back_url" :src="cert.id_card_back_url" :preview-src-list="[cert.id_card_back_url]" fit="cover" style="width: 100px; height: 70px" />
          <span v-else>无</span>
        </el-descriptions-item>
        <el-descriptions-item label="驾驶证">
          <el-image v-if="cert.driver_license_url" :src="cert.driver_license_url" :preview-src-list="[cert.driver_license_url]" fit="cover" style="width: 100px; height: 70px" />
          <span v-else>无</span>
        </el-descriptions-item>
        <el-descriptions-item label="行驶证">
          <el-image v-if="cert.vehicle_license_url" :src="cert.vehicle_license_url" :preview-src-list="[cert.vehicle_license_url]" fit="cover" style="width: 100px; height: 70px" />
          <span v-else>无</span>
        </el-descriptions-item>
      </el-descriptions>

      <el-descriptions title="审核记录" :column="2" border>
        <el-descriptions-item label="审核备注">{{ orDash(cert.audit_remark) }}</el-descriptions-item>
        <el-descriptions-item label="审核人ID">{{ cert.audited_by || '-' }}</el-descriptions-item>
        <el-descriptions-item label="审核时间">{{ orDash(cert.audited_at) }}</el-descriptions-item>
        <el-descriptions-item label="更新时间">{{ orDash(cert.updated_at) }}</el-descriptions-item>
      </el-descriptions>
    </template>
  </el-card>
</template>

<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
```

**Step 4: 验证**

```powershell
cd web-admin
npm run build && npm run dev
```

预期：审核列表可按状态筛选（待审核=1）；详情页展示证照图片与审核记录；通过/驳回二次确认，驳回原因必填。

**Step 5: Commit**

```powershell
git add web-admin/src
git commit -m "feat(admin-web): 司机审核列表/详情/通过/驳回"
```

---

### Task 9: 订单管理（列表 / 异常 Tab / 详情 / 取消）

**Files:**
- Create: `web-admin/src/api/order.js`
- Modify: `web-admin/src/views/order/list.vue`（替换占位）
- Modify: `web-admin/src/views/order/detail.vue`（替换占位）

**Step 1: 写 api/order.js**

```js
import request from './request'

// 订单列表：{list, total, page, page_size}
export const listOrders = (params) => request.get('/orders', { params })
// 异常订单列表：abnormal_type = cancel / payment / dispatch
export const listAbnormalOrders = (params) => request.get('/orders/abnormal', { params })
// 订单详情：{order, status_logs, dispatch_records, price, payment, settlement}
export const getOrder = (id) => request.get(`/orders/${id}`)
// 后台取消订单：body 只允许 reason 字段
export const cancelOrder = (id, data) => request.post(`/orders/${id}/cancel`, data)
```

**Step 2: 实现订单列表页（含异常订单 Tab）** `src/views/order/list.vue`

```vue
<script setup>
import { onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import dayjs from 'dayjs'
import { listOrders, listAbnormalOrders, cancelOrder } from '../../api/order'
import { orderStatusText, orderStatusTag, carTypeText, orDash } from '../../utils/enums'

const router = useRouter()
const loading = ref(false)
const list = ref([])
const total = ref(0)
const activeTab = ref('all')

const query = reactive({
  page: 1,
  page_size: 20,
  keyword: '',
  status: 0,
  user_id: '',
  driver_id: '',
  range: [],
  abnormal_type: '',
})

const statusOptions = [
  { label: '全部状态', value: 0 },
  { label: '待接单', value: 1 },
  { label: '已接单', value: 2 },
  { label: '行程中', value: 3 },
  { label: '待支付', value: 4 },
  { label: '已完成', value: 5 },
  { label: '已取消', value: 6 },
]

const abnormalOptions = [
  { label: '全部异常', value: '' },
  { label: '取消异常', value: 'cancel' },
  { label: '支付异常', value: 'payment' },
  { label: '派单异常', value: 'dispatch' },
]

const buildParams = () => {
  const params = { page: query.page, page_size: query.page_size, keyword: query.keyword }
  if (activeTab.value === 'all') {
    params.status = query.status
    if (query.user_id) params.user_id = query.user_id
    if (query.driver_id) params.driver_id = query.driver_id
  } else {
    params.abnormal_type = query.abnormal_type
    if (query.user_id) params.user_id = query.user_id
    if (query.driver_id) params.driver_id = query.driver_id
  }
  if (query.range && query.range.length === 2) {
    params.start_time = dayjs(query.range[0]).format('YYYY-MM-DD 00:00:00')
    params.end_time = dayjs(query.range[1]).format('YYYY-MM-DD 23:59:59')
  }
  return params
}

const fetchData = async () => {
  loading.value = true
  try {
    const api = activeTab.value === 'all' ? listOrders : listAbnormalOrders
    const data = await api(buildParams())
    list.value = data?.list || []
    total.value = Number(data?.total || 0)
  } finally {
    loading.value = false
  }
}

const handleTabChange = () => {
  query.page = 1
  fetchData()
}

const handleSearch = () => {
  query.page = 1
  fetchData()
}

const handleReset = () => {
  query.keyword = ''
  query.status = 0
  query.user_id = ''
  query.driver_id = ''
  query.range = []
  query.abnormal_type = ''
  handleSearch()
}

// 后台取消订单：二次确认 + 必填原因；仅未完成订单可取消
const handleCancel = async (row) => {
  const { value } = await ElMessageBox.prompt('请输入取消原因', `取消订单「${row.order_no}」`, {
    confirmButtonText: '确认取消订单',
    cancelButtonText: '返回',
    inputPlaceholder: '取消原因（必填）',
    inputValidator: (v) => (v && v.trim() ? true : '取消原因不能为空'),
    type: 'warning',
  })
  await cancelOrder(row.id, { reason: value.trim() })
  ElMessage.success('订单已取消')
  fetchData()
}

onMounted(fetchData)
</script>

<template>
  <el-card shadow="never">
    <el-tabs v-model="activeTab" @tab-change="handleTabChange">
      <el-tab-pane label="全部订单" name="all" />
      <el-tab-pane label="异常订单" name="abnormal" />
    </el-tabs>

    <el-form inline @submit.prevent>
      <el-form-item label="订单号">
        <el-input v-model="query.keyword" placeholder="订单号" clearable style="width: 200px" @keyup.enter="handleSearch" />
      </el-form-item>
      <el-form-item v-if="activeTab === 'all'" label="状态">
        <el-select v-model="query.status" style="width: 120px">
          <el-option v-for="o in statusOptions" :key="o.value" :label="o.label" :value="o.value" />
        </el-select>
      </el-form-item>
      <el-form-item v-if="activeTab === 'abnormal'" label="异常类型">
        <el-select v-model="query.abnormal_type" style="width: 120px">
          <el-option v-for="o in abnormalOptions" :key="o.value" :label="o.label" :value="o.value" />
        </el-select>
      </el-form-item>
      <el-form-item label="用户ID">
        <el-input v-model="query.user_id" placeholder="用户ID" clearable style="width: 110px" />
      </el-form-item>
      <el-form-item label="司机ID">
        <el-input v-model="query.driver_id" placeholder="司机ID" clearable style="width: 110px" />
      </el-form-item>
      <el-form-item label="下单时间">
        <el-date-picker v-model="query.range" type="daterange" range-separator="至" start-placeholder="开始日期" end-placeholder="结束日期" value-format="YYYY-MM-DD" clearable />
      </el-form-item>
      <el-form-item>
        <el-button type="primary" @click="handleSearch">查询</el-button>
        <el-button @click="handleReset">重置</el-button>
      </el-form-item>
    </el-form>

    <el-table v-loading="loading" :data="list">
      <el-table-column prop="order_no" label="订单号" width="200" />
      <el-table-column prop="user_id" label="用户ID" width="90" />
      <el-table-column label="司机ID" width="90">
        <template #default="{ row }">{{ row.driver_id || '-' }}</template>
      </el-table-column>
      <el-table-column label="车型" width="100">
        <template #default="{ row }">{{ carTypeText(row.car_type) }}</template>
      </el-table-column>
      <el-table-column label="起点" min-width="140" show-overflow-tooltip>
        <template #default="{ row }">{{ orDash(row.from_address) }}</template>
      </el-table-column>
      <el-table-column label="终点" min-width="140" show-overflow-tooltip>
        <template #default="{ row }">{{ orDash(row.to_address) }}</template>
      </el-table-column>
      <el-table-column label="预估价" width="90">
        <template #default="{ row }">{{ orDash(row.estimated_price) }}</template>
      </el-table-column>
      <el-table-column label="状态" width="90">
        <template #default="{ row }">
          <el-tag :type="orderStatusTag(row.status)" size="small">{{ orderStatusText(row.status) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column v-if="activeTab === 'abnormal'" label="异常原因" min-width="140" show-overflow-tooltip>
        <template #default="{ row }">{{ orDash(row.abnormal_reason) }}</template>
      </el-table-column>
      <el-table-column prop="created_at" label="下单时间" width="170" />
      <el-table-column label="操作" width="140" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="router.push(`/orders/${row.id}`)">详情</el-button>
          <el-button v-if="row.status !== 5 && row.status !== 6" link type="danger" @click="handleCancel(row)">取消</el-button>
        </template>
      </el-table-column>
      <template #empty>
        <el-empty description="暂无订单数据" />
      </template>
    </el-table>

    <el-pagination
      v-model:current-page="query.page"
      v-model:page-size="query.page_size"
      :total="total"
      :page-sizes="[10, 20, 50, 100]"
      layout="total, sizes, prev, pager, next, jumper"
      style="margin-top: 16px; justify-content: flex-end"
      @change="fetchData"
    />
  </el-card>
</template>
```

**Step 3: 实现订单详情页** `src/views/order/detail.vue`

```vue
<script setup>
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getOrder, cancelOrder } from '../../api/order'
import { orderStatusText, orderStatusTag, carTypeText, orDash } from '../../utils/enums'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const detail = ref(null)

const fetchData = async () => {
  loading.value = true
  try {
    detail.value = await getOrder(route.params.id)
  } finally {
    loading.value = false
  }
}

// 后台取消：仅未完成/未取消订单显示按钮
const handleCancel = async () => {
  const { value } = await ElMessageBox.prompt('请输入取消原因', `取消订单「${detail.value.order.order_no}」`, {
    confirmButtonText: '确认取消订单',
    cancelButtonText: '返回',
    inputPlaceholder: '取消原因（必填）',
    inputValidator: (v) => (v && v.trim() ? true : '取消原因不能为空'),
    type: 'warning',
  })
  await cancelOrder(detail.value.order.id, { reason: value.trim() })
  ElMessage.success('订单已取消')
  fetchData()
}

onMounted(fetchData)
</script>

<template>
  <el-card shadow="never" v-loading="loading">
    <template #header>
      <div class="card-header">
        <span>订单详情</span>
        <div>
          <el-button
            v-if="detail?.order && detail.order.status !== 5 && detail.order.status !== 6"
            type="danger"
            @click="handleCancel"
          >取消订单</el-button>
          <el-button @click="router.back()">返回</el-button>
        </div>
      </div>
    </template>

    <el-empty v-if="!detail && !loading" description="订单不存在" />

    <template v-else-if="detail">
      <el-descriptions title="订单主信息" :column="3" border style="margin-bottom: 16px">
        <el-descriptions-item label="订单号">{{ detail.order.order_no }}</el-descriptions-item>
        <el-descriptions-item label="订单ID">{{ detail.order.id }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="orderStatusTag(detail.order.status)" size="small">{{ orderStatusText(detail.order.status) }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="用户ID">{{ detail.order.user_id }}</el-descriptions-item>
        <el-descriptions-item label="司机ID">{{ detail.order.driver_id || '-' }}</el-descriptions-item>
        <el-descriptions-item label="车型">{{ carTypeText(detail.order.car_type) }}</el-descriptions-item>
        <el-descriptions-item label="起点" :span="2">{{ orDash(detail.order.from_address) }}</el-descriptions-item>
        <el-descriptions-item label="终点">{{ orDash(detail.order.to_address) }}</el-descriptions-item>
        <el-descriptions-item label="预估距离">{{ detail.order.estimated_distance_m ? `${(detail.order.estimated_distance_m / 1000).toFixed(2)} km` : '-' }}</el-descriptions-item>
        <el-descriptions-item label="预估时长">{{ detail.order.estimated_duration_s ? `${Math.round(detail.order.estimated_duration_s / 60)} 分钟` : '-' }}</el-descriptions-item>
        <el-descriptions-item label="预估价格">¥{{ orDash(detail.order.estimated_price) }}</el-descriptions-item>
        <el-descriptions-item label="取消原因" :span="2">{{ orDash(detail.order.cancel_reason) }}</el-descriptions-item>
        <el-descriptions-item label="取消方">{{ orDash(detail.order.cancel_by) }}</el-descriptions-item>
        <el-descriptions-item label="下单时间">{{ orDash(detail.order.created_at) }}</el-descriptions-item>
        <el-descriptions-item label="更新时间" :span="2">{{ orDash(detail.order.updated_at) }}</el-descriptions-item>
      </el-descriptions>

      <el-descriptions v-if="detail.price" title="价格明细" :column="4" border style="margin-bottom: 16px">
        <el-descriptions-item label="预估价">¥{{ orDash(detail.price.estimated_price) }}</el-descriptions-item>
        <el-descriptions-item label="实际价">¥{{ orDash(detail.price.actual_price) }}</el-descriptions-item>
        <el-descriptions-item label="起步价">¥{{ orDash(detail.price.base_fee) }}</el-descriptions-item>
        <el-descriptions-item label="里程费">¥{{ orDash(detail.price.distance_fee) }}</el-descriptions-item>
        <el-descriptions-item label="时长费">¥{{ orDash(detail.price.time_fee) }}</el-descriptions-item>
        <el-descriptions-item label="夜间费">¥{{ orDash(detail.price.night_fee) }}</el-descriptions-item>
        <el-descriptions-item label="动态加价">¥{{ orDash(detail.price.dynamic_fee) }}</el-descriptions-item>
        <el-descriptions-item label="应付金额">¥{{ orDash(detail.price.payable_amount) }}</el-descriptions-item>
        <el-descriptions-item label="优惠金额">¥{{ orDash(detail.price.discount_amount) }}</el-descriptions-item>
        <el-descriptions-item label="平台补贴">¥{{ orDash(detail.price.platform_subsidy) }}</el-descriptions-item>
      </el-descriptions>

      <el-descriptions v-if="detail.payment" title="支付信息" :column="4" border style="margin-bottom: 16px">
        <el-descriptions-item label="支付单号">{{ orDash(detail.payment.payment_no) }}</el-descriptions-item>
        <el-descriptions-item label="金额">¥{{ orDash(detail.payment.amount) }}</el-descriptions-item>
        <el-descriptions-item label="渠道">{{ orDash(detail.payment.channel) }}</el-descriptions-item>
        <el-descriptions-item label="支付状态">{{ orDash(detail.payment.status) }}</el-descriptions-item>
        <el-descriptions-item label="第三方流水号">{{ orDash(detail.payment.transaction_id) }}</el-descriptions-item>
        <el-descriptions-item label="退款金额">¥{{ orDash(detail.payment.refund_amount) }}</el-descriptions-item>
        <el-descriptions-item label="支付时间" :span="2">{{ orDash(detail.payment.paid_at) }}</el-descriptions-item>
      </el-descriptions>

      <el-descriptions v-if="detail.settlement" title="结算信息" :column="4" border style="margin-bottom: 16px">
        <el-descriptions-item label="结算单号">{{ orDash(detail.settlement.settlement_no) }}</el-descriptions-item>
        <el-descriptions-item label="总金额">¥{{ orDash(detail.settlement.total_amount) }}</el-descriptions-item>
        <el-descriptions-item label="平台抽成">¥{{ orDash(detail.settlement.platform_commission) }}</el-descriptions-item>
        <el-descriptions-item label="司机收入">¥{{ orDash(detail.settlement.driver_income) }}</el-descriptions-item>
        <el-descriptions-item label="结算时间" :span="2">{{ orDash(detail.settlement.settled_at) }}</el-descriptions-item>
        <el-descriptions-item label="状态">{{ orDash(detail.settlement.status) }}</el-descriptions-item>
      </el-descriptions>

      <el-card shadow="never" header="状态流转日志" style="margin-bottom: 16px">
        <el-empty v-if="!detail.status_logs?.length" description="暂无状态日志" />
        <el-table v-else :data="detail.status_logs" size="small">
          <el-table-column label="流转" width="180">
            <template #default="{ row }">{{ orderStatusText(row.from_status) }} -> {{ orderStatusText(row.to_status) }}</template>
          </el-table-column>
          <el-table-column prop="operator_type" label="操作方" width="110" />
          <el-table-column prop="operator_id" label="操作人ID" width="100" />
          <el-table-column prop="remark" label="备注" min-width="140">
            <template #default="{ row }">{{ orDash(row.remark) }}</template>
          </el-table-column>
          <el-table-column prop="created_at" label="时间" width="170" />
        </el-table>
      </el-card>

      <el-card shadow="never" header="派单记录">
        <el-empty v-if="!detail.dispatch_records?.length" description="暂无派单记录" />
        <el-table v-else :data="detail.dispatch_records" size="small">
          <el-table-column prop="id" label="ID" width="70" />
          <el-table-column prop="driver_id" label="司机ID" width="90" />
          <el-table-column prop="dispatch_type" label="派单类型" width="100" />
          <el-table-column prop="status" label="状态" width="80" />
          <el-table-column prop="match_score" label="匹配分" width="90" />
          <el-table-column prop="remark" label="备注" min-width="140">
            <template #default="{ row }">{{ orDash(row.remark) }}</template>
          </el-table-column>
          <el-table-column prop="created_at" label="时间" width="170" />
        </el-table>
      </el-card>
    </template>
  </el-card>
</template>

<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
```

**Step 4: 验证**

```powershell
cd web-admin
npm run build && npm run dev
```

预期：订单列表 Tab 切换正常；筛选分页正常；详情页分区展示聚合数据；未完结订单可取消（二次确认+必填原因）。

**Step 5: Commit**

```powershell
git add web-admin/src
git commit -m "feat(admin-web): 订单列表/异常订单/详情/后台取消"
```

---

### Task 10: 操作日志

**Files:**
- Create: `web-admin/src/api/log.js`
- Modify: `web-admin/src/views/log/list.vue`（替换占位）

**Step 1: 写 api/log.js**

```js
import request from './request'

// 操作日志列表：{list, total, page, page_size}
export const listOperationLogs = (params) => request.get('/operation-logs', { params })
```

**Step 2: 实现日志列表页** `src/views/log/list.vue`

```vue
<script setup>
import { onMounted, reactive, ref } from 'vue'
import dayjs from 'dayjs'
import { listOperationLogs } from '../../api/log'
import { orDash } from '../../utils/enums'

const loading = ref(false)
const list = ref([])
const total = ref(0)

const query = reactive({
  page: 1,
  page_size: 20,
  admin_id: '',
  module: '',
  action: '',
  range: [],
})

const moduleOptions = ['auth', 'user', 'driver', 'order', 'coupon', 'promotion', 'price_rule', 'risk', 'export'].map((m) => ({ label: m, value: m }))

const buildParams = () => {
  const params = { page: query.page, page_size: query.page_size, module: query.module, action: query.action }
  if (query.admin_id) params.admin_id = query.admin_id
  if (query.range && query.range.length === 2) {
    params.start_time = dayjs(query.range[0]).format('YYYY-MM-DD 00:00:00')
    params.end_time = dayjs(query.range[1]).format('YYYY-MM-DD 23:59:59')
  }
  return params
}

const fetchData = async () => {
  loading.value = true
  try {
    const data = await listOperationLogs(buildParams())
    list.value = data?.list || []
    total.value = Number(data?.total || 0)
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  query.page = 1
  fetchData()
}

const handleReset = () => {
  query.admin_id = ''
  query.module = ''
  query.action = ''
  query.range = []
  handleSearch()
}

onMounted(fetchData)
</script>

<template>
  <el-card shadow="never">
    <el-form inline @submit.prevent>
      <el-form-item label="管理员ID">
        <el-input v-model="query.admin_id" placeholder="管理员ID" clearable style="width: 120px" @keyup.enter="handleSearch" />
      </el-form-item>
      <el-form-item label="模块">
        <el-select v-model="query.module" clearable placeholder="全部" style="width: 140px">
          <el-option v-for="o in moduleOptions" :key="o.value" :label="o.label" :value="o.value" />
        </el-select>
      </el-form-item>
      <el-form-item label="动作">
        <el-input v-model="query.action" placeholder="如 login/approve" clearable style="width: 140px" @keyup.enter="handleSearch" />
      </el-form-item>
      <el-form-item label="时间">
        <el-date-picker v-model="query.range" type="daterange" range-separator="至" start-placeholder="开始日期" end-placeholder="结束日期" value-format="YYYY-MM-DD" clearable />
      </el-form-item>
      <el-form-item>
        <el-button type="primary" @click="handleSearch">查询</el-button>
        <el-button @click="handleReset">重置</el-button>
      </el-form-item>
    </el-form>

    <el-table v-loading="loading" :data="list">
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="admin_id" label="管理员ID" width="100" />
      <el-table-column prop="module" label="模块" width="110" />
      <el-table-column prop="action" label="动作" width="110" />
      <el-table-column prop="target_type" label="目标类型" width="100">
        <template #default="{ row }">{{ orDash(row.target_type) }}</template>
      </el-table-column>
      <el-table-column prop="target_id" label="目标ID" width="90" />
      <el-table-column prop="detail" label="详情" min-width="200" show-overflow-tooltip>
        <template #default="{ row }">{{ orDash(row.detail) }}</template>
      </el-table-column>
      <el-table-column prop="ip" label="IP" width="130" />
      <el-table-column prop="created_at" label="操作时间" width="170" />
      <template #empty>
        <el-empty description="暂无日志数据" />
      </template>
    </el-table>

    <el-pagination
      v-model:current-page="query.page"
      v-model:page-size="query.page_size"
      :total="total"
      :page-sizes="[10, 20, 50, 100]"
      layout="total, sizes, prev, pager, next, jumper"
      style="margin-top: 16px; justify-content: flex-end"
      @change="fetchData"
    />
  </el-card>
</template>
```

**Step 3: 验证**

```powershell
cd web-admin
npm run build && npm run dev
```

预期：日志列表可查、筛选、分页；空态正常。

**Step 4: Commit**

```powershell
git add web-admin/src
git commit -m "feat(admin-web): 操作日志列表"
```

---

### Task 11: README 与最终验收

**Files:**
- Create: `web-admin/README.md`

**Step 1: 写 README.md**

````markdown
# 小隆出行运营管理后台（前端）

Vue 3 + Element Plus + Vite 实现的管理后台 P0 前端，对接 `api/admin`（go-zero）。

## 功能

- 管理员登录/退出（Bearer Token，会话失效自动跳登录）
- 工作台：运营总览 + 订单统计
- 用户管理：列表/详情/冻结/解封
- 司机审核：列表/详情/通过/驳回
- 订单管理：列表/异常订单/详情/后台取消
- 操作日志查询

## 快速开始

前置：后端 `rpc/adminsvc`（8084）与 `api/admin`（8083）已启动。

```bash
npm install
npm run dev     # http://127.0.0.1:5173，代理 /admin -> 127.0.0.1:8083
```

构建：

```bash
npm run build   # 产出 dist/，hash 路由，任意静态服务器可直接部署
```

测试账号：`admin / 123456`（见 docs/api/admin接口验收清单.md）

## 目录结构

- `src/api/`：axios 封装与各模块接口
- `src/router/`：hash 路由 + 登录守卫
- `src/store/`：Pinia 登录态
- `src/layout/`：三段式后台布局
- `src/views/`：页面
- `src/utils/enums.js`：枚举中文映射
````

**Step 2: 全量验收**

```powershell
cd web-admin
npm run build
npm run dev
```

按设计文档第 8 节验收清单逐项手动验证：
- 登录/退出/守卫正常
- 六个页面列表可查、可筛选、可分页
- 写操作均弹二次确认并调用后端成功
- 详情页正确渲染聚合数据
- 空数据场景显示空态
- build 成功

**Step 3: Commit**

```powershell
git add web-admin
git commit -m "docs(admin-web): 前端 README 与最终验收"
```
