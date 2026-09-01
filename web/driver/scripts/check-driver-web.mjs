import { existsSync, readFileSync, readdirSync } from 'node:fs'
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

function collectDriverFiles(dir, exts, out = []) {
  const abs = file(dir)
  if (!existsSync(abs)) return out
  for (const entry of readdirSync(abs, { withFileTypes: true })) {
    const rel = dir + '/' + entry.name
    if (entry.isDirectory()) {
      if (entry.name === 'dist' || entry.name === 'node_modules') continue
      collectDriverFiles(rel, exts, out)
      continue
    }
    if (exts.some((ext) => entry.name.endsWith(ext))) out.push(rel)
  }
  return out
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
  'web/driver/src/views/DriverHome.vue'
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
assert(driverRouter.includes('redirect:'), 'driver frontend root must redirect based on login state')
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
const minePanel = readIfExists('web/driver/src/components/driver-home/DriverMinePanel.vue')
const workbenchPanel = readIfExists('web/driver/src/components/driver-home/DriverWorkbenchPanel.vue')
const trajectoryPanel = readIfExists('web/driver/src/components/driver-home/DriverTrajectoryPanel.vue')
const driverAmapConfig = readIfExists('web/driver/src/config/amap.js')
const login = readIfExists('web/driver/src/views/DriverLogin.vue')
const app = readIfExists('web/driver/src/App.vue')
const orderLogic = readIfExists('api/driver/internal/logic/order_logic.go')
const driverTextFiles = [
  'web/driver/index.html',
  ...collectDriverFiles('web/driver/src', ['.js', '.vue']),
  ...collectDriverFiles('api/driver', ['.go', '.api', '.yaml']),
  ...collectDriverFiles('rpc/driversvc', ['.go', '.proto', '.yaml'])
]
const mojibakePattern = /�|���|鏄|鏉|鏈|鎵|鍦|绛|瀵|嗛|渶|绔|繚|鎸|鑷|楠|鐮|瘉|銆|锛|€|閹|劗|骞囬|惄|绋|垮|璇|榫|濮|韬|椹|澶|鎻|娉|绠|鍛/
const unicodeEscapePattern = /\\u[0-9a-fA-F]{4}/
for (const path of driverTextFiles) {
  const text = readIfExists(path)
  assert(!text.startsWith('\uFEFF'), 'driver text file must not start with BOM: ' + path)
  assert(!mojibakePattern.test(text), 'driver text file contains mojibake text: ' + path)
  assert(!unicodeEscapePattern.test(text), 'driver text file contains escaped unicode text: ' + path)
}
const driverLogicFiles = [
  ...collectDriverFiles('api/driver/internal/logic', ['.go']),
  ...collectDriverFiles('rpc/driversvc/internal/logic', ['.go'])
].filter((path) => !path.endsWith('_test.go'))
const directSQLPattern = /\.(Raw|Exec|Table)\(|\b(SELECT|INSERT|UPDATE|DELETE)\s+|\bFROM\s+|\bJOIN\s+|\bWHERE\s+|\bORDER BY\b|\bGROUP BY\b/
for (const path of driverLogicFiles) {
  assert(!directSQLPattern.test(readIfExists(path)), 'driver logic must not contain direct SQL: ' + path)
}
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
assert(!login.includes('SmsCodeDialog'), 'driver login must not mount SMS code dialog')
assert(login.includes('验证码已发送（联调验证码见服务端日志）'), 'driver login must not reveal plaintext code')
assert(home.includes('class="driver-hero"'), 'driver home must start with a driver hero')
assert(home.includes('class="quick-entry-row"'), 'driver home must expose H5 quick action controls')
assert(home.includes('class="tab-panel-scroll"'), 'driver tab content must use a mobile scroll panel')
assert(home.includes('openHeatmapVisible'), 'driver home must manage heatmap popup visibility state')
assert(home.includes('@click="openHeatmap"'), 'driver heatmap button must open a popup')
assert(home.includes('v-model:show="heatmapVisible"'), 'driver heatmap must render in a popup, not inline')
assert(home.includes(':teleport="false"'), 'driver heatmap popup must stay inside the driver H5 shell')
assert(home.includes('driver-heatmap-popup'), 'driver heatmap popup must have its own container')
assert(home.includes('getOrderHeatmap'), 'driver heatmap popup must fetch backend heatmap data')
assert(home.includes('const heatmapRadiusMeters = 5000'), 'driver heatmap radius must be fixed to 5km')
assert(!home.includes('const heatmapRadiusMeters = 10000'), 'driver heatmap radius must not remain 10km')
assert(home.includes('zoom: 12'), 'driver heatmap initial zoom must start from city-level H5 view')
assert(!home.includes('zoom: 4'), 'driver heatmap must not flash from nationwide zoom')
assert(home.includes("height: 'min(88vh, 760px)'"), 'driver heatmap popup must use a tall mobile sheet')
assert(home.includes("'AMap.Geolocation'"), 'driver heatmap must use AMap geolocation like passenger H5')
assert(home.includes('locateHeatmapByAMap'), 'driver heatmap must locate through AMap before falling back')
assert(home.includes('readRememberedWorkLocation'), 'driver heatmap must prefer existing watched driver location before relocalizing')
assert(home.includes('heatmap-h5-sheet'), 'driver heatmap popup must use an H5 bottom sheet layout')
assert(home.includes('heatmapPhoneSheetStyle'), 'driver heatmap popup must use a phone-width sheet style')
assert(home.includes("width: 'min(100vw, 390px)'"), 'driver heatmap popup must not stretch to desktop viewport width')
assert(home.includes('radius: 24'), 'driver heatmap radius must stay tight enough to avoid lake spill')
assert(home.includes("0.2: 'rgba(37, 99, 235, 0)'"), 'driver heatmap must hide low-intensity interpolation tails')
assert(home.includes('heatmap-sheet-grabber'), 'driver heatmap H5 sheet must expose a mobile drag handle')
assert(home.includes('heatmap-floating-actions'), 'driver heatmap H5 sheet must use floating map actions')
assert(home.includes('heatmap-chip-strip'), 'driver heatmap H5 sheet must use compact horizontal heat point chips')
assert(!home.includes('热力图</b></button>'), 'driver heatmap must not remain a dead static button')
assert(driverAmapConfig.includes('getAmapConfig'), 'driver app must own its AMap config helper')
assert(driverAmapConfig.includes('VITE_AMAP_KEY'), 'driver AMap config must support env override')
assert(!workbenchPanel.includes('@/../user') && !trajectoryPanel.includes('@/../user'), 'driver maps must not import user app internals')
assert(workbenchPanel.includes('@amap/amap-jsapi-loader'), 'driver workbench must use the same AMap loader as user app')
assert(workbenchPanel.includes("from '@/config/amap'"), 'driver workbench must use driver-owned AMap config')
assert(workbenchPanel.includes('watchPosition'), 'driver workbench map must follow live browser location')
assert(workbenchPanel.includes('.Polyline'), 'driver workbench map must render live movement path')
assert(trajectoryPanel.includes('@amap/amap-jsapi-loader'), 'driver trajectory panel must use the same AMap loader as user app')
assert(trajectoryPanel.includes("from '@/config/amap'"), 'driver trajectory panel must use driver-owned AMap config')
assert(trajectoryPanel.includes('setInterval') && trajectoryPanel.includes('refreshIntervalMs'), 'driver trajectory map must refresh while visible')
assert(trajectoryPanel.includes('.Polyline'), 'driver trajectory panel must render order path on the map')
assert(home.includes("mineSection.value = 'trajectory'"), 'trajectory action must request the mine trajectory section')
assert(home.includes('activeTab.value = 3'), 'trajectory action must open the mine tab')
assert(!home.includes("driverStore.setCurrentOrder(order, Number(order.status) === 3 ? 'trip' : 'pickup')"), 'dispatch pushes must not mark pending dispatch as current order')
assert(home.includes('dispatchStatus: item.dispatch?.status'), 'dispatch rows must expose dispatch status')
assert(home.includes('orderId: item.dispatch?.orderId || item.order?.orderId'), 'dispatch rows must expose order id from dispatch payload')
assert(minePanel.includes('defaultSection'), 'mine panel must accept an externally selected section')
assert(orderLogic.includes('if len(orderIDStrs) == 0 {'), 'available orders must not fall back to global wait-accept list')
assert(home.includes('getTodayIncome'), 'driver wallet must call backend today income endpoint')
assert(home.includes('getWeekIncome'), 'driver wallet must call backend week income endpoint')
assert(home.includes('/api/driver/v1/ws'), 'driver WebSocket must connect to registered /api/driver/v1/ws route')
assert(!home.includes('/api/driver/v1/push/ws'), 'driver WebSocket must not use stale /api/driver/v1/push/ws route')

console.log('driver frontend isolation checks passed')
