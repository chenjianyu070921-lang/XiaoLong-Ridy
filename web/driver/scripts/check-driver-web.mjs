import { existsSync, readFileSync } from 'node:fs'
import { resolve } from 'node:path'

const root = resolve(import.meta.dirname, '..')
const repoRoot = resolve(root, '..', '..')

function file(path) {
  return resolve(repoRoot, path)
}

function assert(condition, message) {
  if (!condition) {
    throw new Error(message)
  }
}

function readIfExists(path) {
  return existsSync(file(path)) ? readFileSync(file(path), 'utf8') : ''
}

for (const path of [
  'web/driver/package.json',
  'web/driver/vite.config.js',
  'web/driver/index.html',
  'web/driver/src/main.js',
  'web/driver/src/App.vue',
  'web/driver/src/router/index.js',
  'web/driver/src/api/driver.js',
  'web/driver/src/stores/driver.js',
  'web/driver/src/views/DriverLogin.vue',
  'web/driver/src/views/DriverHome.vue',
  'web/driver/src/components/SmsCodeDialog.vue'
]) {
  assert(existsSync(file(path)), `missing independent driver frontend file: ${path}`)
}

const userRouter = readIfExists('web/user/src/router/index.js')
assert(!userRouter.includes("path: '/driver'"), 'passenger H5 must not register /driver route')
assert(!userRouter.includes("path: '/driver/login'"), 'passenger H5 must not register /driver/login route')
assert(!userRouter.includes("path: '/driver/home'"), 'passenger H5 must not register /driver/home route')
assert(!userRouter.includes('requiresDriverAuth'), 'passenger H5 must not contain driver auth guard')

for (const path of [
  'web/user/src/api/driver.js',
  'web/user/src/stores/driver.js',
  'web/user/src/views/driver/DriverLogin.vue',
  'web/user/src/views/driver/DriverHome.vue'
]) {
  assert(!existsSync(file(path)), `driver code must live under web/driver, not ${path}`)
}

const driverRouter = readIfExists('web/driver/src/router/index.js')
assert(driverRouter.includes("redirect: '/login'"), 'driver frontend root must redirect to /login')
assert(driverRouter.includes("path: '/login'"), 'driver frontend must expose /login')
assert(driverRouter.includes("path: '/home'"), 'driver frontend must expose /home')
assert(driverRouter.includes('requiresDriverAuth'), 'driver frontend must protect /home')
assert(!driverRouter.includes("'/driver/login'"), 'driver frontend must not depend on passenger /driver/login path')
assert(!driverRouter.includes("'/driver/home'"), 'driver frontend must not depend on passenger /driver/home path')

const viteConfig = readIfExists('web/driver/vite.config.js')
assert(viteConfig.includes('5175'), 'driver frontend dev server must use its own port 5175')
assert(viteConfig.includes("'/api/driver'"), 'driver frontend must proxy driver API')
assert(viteConfig.includes('8082'), 'driver API proxy must target api/driver on port 8082')

const api = readIfExists('web/driver/src/api/driver.js')
for (const endpoint of [
  '/auth/send-sms-code',
  '/auth/login-by-password',
  '/auth/login-by-sms',
  '/drivers/register',
  '/drivers/update',
  '/drivers/get',
  '/drivers/ai-score',
  '/drivers/online',
  '/drivers/offline',
  '/drivers/heartbeat',
  '/drivers/location/report',
  '/vehicles',
  '/vehicles/get',
  '/vehicles/update',
  '/vehicles/delete',
  '/drivers/certification/upload',
  '/drivers/certification',
  '/income/summary',
  '/wallet/summary',
  '/income/today',
  '/income/week',
  '/income/bills',
  '/orders/accept',
  '/orders/reject',
  '/orders/confirm-arrive',
  '/orders/start-trip',
  '/orders/finish-trip',
  '/orders/available',
  '/orders/detail',
  '/orders/trajectory',
  '/orders/list',
  '/orders/dispatches',
  '/reviews/list'
]) {
  assert(api.includes(endpoint), `missing real driver API wrapper: ${endpoint}`)
}

const home = readIfExists('web/driver/src/views/DriverHome.vue')
const login = readIfExists('web/driver/src/views/DriverLogin.vue')
const app = readIfExists('web/driver/src/App.vue')
const smsDialog = readIfExists('web/driver/src/components/SmsCodeDialog.vue')
assert(login.includes("router.replace('/home')"), 'driver login must navigate inside independent app')
assert(api.includes("router.push('/login')"), 'driver auth expiry must navigate inside independent app')
assert(home.includes("router.replace('/login')"), 'driver logout must navigate inside independent app')
assert(home.includes('van-tabbar'), 'driver H5 navigation must use fixed mobile bottom tabbar')
assert(home.includes('workStatusPayload()'), 'online/offline actions must send device and location payload')
assert(home.includes('safeApiCall'), 'driver H5 actions must keep UI usable when API calls fail')
assert(home.includes('showIncomeLoadFailure'), 'driver income API failures must open a failure dialog')
assert(home.includes('rejectOrder(orderId, reason)'), 'driver reject action must submit explicit reject reason')
assert(!home.includes('actualPriceCents'), 'driver finish trip must not ask driver to enter settlement amount')
assert(app.includes('driver-phone-shell'), 'driver app must render a centered phone shell for H5 preview')
assert(login.includes('class="login-card"'), 'driver login must use a compact H5 login card')
assert(login.includes('class="auth-mode-tabs"'), 'driver login must expose touch-friendly H5 auth mode tabs')
assert(login.includes('<SmsCodeDialog'), 'driver login must mount SMS code dialog after sending SMS')
assert(login.includes('smsCodeDialogVisible.value = true'), 'driver login must open SMS code dialog when backend returns code')
assert(login.includes("smsCodeValue.value = String(res?.code || '').trim()"), 'driver login must capture backend SMS code')
assert(home.includes('class="driver-hero"'), 'driver home must start with a driver hero')
assert(home.includes('class="quick-entry-row"'), 'driver home must expose H5 quick action controls')
assert(home.includes('class="tab-panel-scroll"'), 'driver tab content must use a mobile scroll panel')
assert(home.includes('getTodayIncome'), 'driver wallet must call backend today income endpoint')
assert(home.includes('getWeekIncome'), 'driver wallet must call backend week income endpoint')
assert(home.includes('/api/driver/v1/ws'), 'driver WebSocket must connect to registered /api/driver/v1/ws route')
assert(!home.includes('/api/driver/v1/push/ws'), 'driver WebSocket must not use stale /api/driver/v1/push/ws route')
assert(smsDialog.includes('������֤��'), 'driver SMS code dialog title is required')
assert(smsDialog.includes('��֤�����Ժ�˷��ؽ��'), 'driver SMS code dialog must explain the code source')
assert(smsDialog.includes('sms-code-dialog__code'), 'driver SMS code dialog must render the returned code')

console.log('driver frontend isolation checks passed')

