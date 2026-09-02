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

function functionSource(text, name) {
  const start = text.indexOf(`function ${name}(`)
  assert(start >= 0, `missing function ${name}`)
  const openBrace = text.indexOf('{', start)
  assert(openBrace >= 0, `missing function body for ${name}`)
  let depth = 0
  for (let index = openBrace; index < text.length; index += 1) {
    const char = text[index]
    if (char === '{') depth += 1
    if (char === '}') {
      depth -= 1
      if (depth === 0) return text.slice(start, index + 1)
    }
  }
  throw new Error(`unterminated function body for ${name}`)
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
  'web/driver/src/api/request.js',
  'web/driver/src/api/driver.js',
  'web/driver/src/api/upload.js',
  'web/driver/src/stores/driver.js',
  'web/driver/src/views/DriverLogin.vue',
  'web/driver/src/views/DriverHome.vue',
  'web/driver/src/views/DriverProfileEdit.vue'
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
assert(driverRouter.includes("path: '/profile/edit'"), 'driver frontend must expose dedicated profile edit page')
assert(driverRouter.includes("name: 'DriverProfileEdit'"), 'driver profile edit route must be named')
assert(driverRouter.includes('requiresDriverAuth'), 'driver frontend must protect /home')
assert(driverRouter.includes('DriverProfileEdit.vue'), 'driver profile edit route must load its dedicated view')
assert(!driverRouter.includes("path: '/workbench'"), 'driver frontend must not keep a standalone workbench route when home is the workbench')
assert(!driverRouter.includes('DriverWorkbench.vue'), 'driver frontend must not load a standalone workbench view')
assert(!driverRouter.includes("'/driver/login'"), 'driver frontend must not depend on passenger /driver/login path')
assert(!driverRouter.includes("'/driver/home'"), 'driver frontend must not depend on passenger /driver/home path')

const viteConfig = readIfExists('web/driver/vite.config.js')
assert(viteConfig.includes('5175'), 'driver frontend dev server must use its own port 5175')
assert(viteConfig.includes("'/api/driver'"), 'driver frontend must proxy driver API')
assert(viteConfig.includes('18082'), 'driver API proxy must target api/driver on port 18082')

const api = readIfExists('web/driver/src/api/driver.js')
const uploadApi = readIfExists('web/driver/src/api/upload.js')
const apiRequest = readIfExists('web/driver/src/api/request.js')
const driverIndex = readIfExists('web/driver/index.html')
assert(api.includes("import driverRequest from '@/api/request'"), 'driver API module must reuse the shared driver request wrapper')
assert(!api.includes('axios.create'), 'driver API module must not create axios instances inline')
assert(!api.includes('interceptors.request'), 'driver API module must not own request interceptors')
assert(!api.includes('interceptors.response'), 'driver API module must not own response interceptors')
assert(!api.includes('clearDriverSession'), 'driver API module must not own session clearing logic')
assert(apiRequest.includes('axios.create'), 'driver request wrapper must create the axios instance')
assert(apiRequest.includes('interceptors.request'), 'driver request wrapper must own request interceptors')
assert(apiRequest.includes('interceptors.response'), 'driver request wrapper must own response interceptors')
assert(apiRequest.includes('clearDriverSession'), 'driver request wrapper must own driver session clearing')
for (const endpoint of [
  '/auth/send-sms-code',
  '/auth/login-by-password',
  '/auth/login-by-sms',
  '/drivers/register',
  '/drivers/update',
  '/upload/avatar-token',
  '/drivers/get',
  '/drivers/nearby',
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
  '/orders/list',
  '/orders/dispatches'
]) {
  assert(api.includes(endpoint) || uploadApi.includes(endpoint), `missing real driver API wrapper: ${endpoint}`)
}
assert(uploadApi.includes("import axios from 'axios'"), 'driver avatar upload helper must direct-upload with axios like passenger H5')
assert(uploadApi.includes('getAvatarUploadToken'), 'driver avatar upload helper must request a qiniu avatar token')
assert(uploadApi.includes('uploadToQiniu'), 'driver avatar upload helper must expose qiniu direct upload')
assert(uploadApi.includes("driverRequest.post('/upload/avatar-token'"), 'driver avatar token helper must call the driver-owned qiniu token endpoint')
assert(uploadApi.includes('new FormData()'), 'driver avatar qiniu upload must send multipart form data')

const home = readIfExists('web/driver/src/views/DriverHome.vue')
const ordersPanel = readIfExists('web/driver/src/components/driver-home/DriverOrdersPanel.vue')
const minePanel = readIfExists('web/driver/src/components/driver-home/DriverMinePanel.vue')
const driverMineData = readIfExists('web/driver/src/components/driver-home/driver-mine-data.js')
const profilePanel = readIfExists('web/driver/src/components/driver-home/DriverProfilePanel.vue')
const profileEdit = readIfExists('web/driver/src/views/DriverProfileEdit.vue')
const driverApp = readIfExists('web/driver/src/App.vue')
const walletPage = readIfExists('web/driver/src/views/mine/DriverWalletPage.vue')
const vehiclePage = readIfExists('web/driver/src/views/mine/DriverVehiclePage.vue')
const certificationPage = readIfExists('web/driver/src/views/mine/DriverCertificationPage.vue')
const incomePage = readIfExists('web/driver/src/views/mine/DriverIncomePage.vue')
const orderRecordsPage = readIfExists('web/driver/src/views/mine/DriverOrderRecordsPage.vue')
const driverAssets = readIfExists('web/driver/src/composables/useDriverAssets.js')
const driverAmapConfig = readIfExists('web/driver/src/config/amap.js')
const login = readIfExists('web/driver/src/views/DriverLogin.vue')
const app = readIfExists('web/driver/src/App.vue')
const orderLogic = readIfExists('api/driver/internal/logic/order_logic.go')
const refreshHomeWorkbenchSource = functionSource(home, 'refreshHomeWorkbench')
const handlePushMessageSource = functionSource(home, 'handlePushMessage')
const driverTextFiles = [
  'web/driver/index.html',
  ...collectDriverFiles('web/driver/src', ['.js', '.vue']),
  ...collectDriverFiles('api/driver', ['.go', '.api', '.yaml']),
]
const mojibakeCodePoints = [
  0xfffd, 0x93c4, 0x93c9, 0x93c8, 0x93b5, 0x93a6, 0x5726, 0x7edb, 0x7025,
  0x55db, 0x6e36, 0x7ad4, 0x7e5a, 0x93b8, 0x9477, 0x6960, 0x7629, 0x9286,
  0x951b, 0x95f9, 0x5297, 0x60c4, 0x7ecb, 0x57ae, 0x69ab, 0x6fee, 0x97ec,
  0x6939, 0x6fb6, 0x93bb, 0x5a09, 0x7ba0, 0x935b
]
const mojibakePattern = new RegExp(mojibakeCodePoints.map((codePoint) => String.fromCodePoint(codePoint)).join('|'))
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
assert(apiRequest.includes("router.push('/login')"), 'driver auth expiry must navigate inside independent app')
assert(driverIndex.includes('<link rel="icon" type="image/png" href="/logo.png" />'), 'driver frontend must set the website favicon to the app logo')
assert(api.includes('export function listNearbyDrivers'), 'driver API wrapper must expose the registered nearby drivers endpoint')
assert(api.includes('acceptOrder(orderId, config = {})'), 'driver accept order API must accept axios config for consistent request handling')
assert(api.includes("driverRequest.post('/orders/accept', { orderId }, config)"), 'driver accept order API must pass config to axios')
assert(api.includes('rejectOrder(orderId, reason = \'司机主动拒单\', config = {})'), 'driver reject order API must accept axios config for consistent request handling')
assert(api.includes("driverRequest.post('/orders/reject', { orderId, reason }, config)"), 'driver reject order API must pass config to axios')
assert(ordersPanel.includes('class="card-header"'), 'driver order cards must copy passenger order card header structure')
assert(ordersPanel.includes('class="status-tag"'), 'driver order cards must copy passenger status tag styling')
assert(ordersPanel.includes('class="route-info"'), 'driver order cards must copy passenger route timeline structure')
assert(ordersPanel.includes('class="dot from"'), 'driver order cards must copy passenger pickup dot')
assert(ordersPanel.includes('class="dot to"'), 'driver order cards must copy passenger destination dot')
assert(ordersPanel.includes('class="card-footer"'), 'driver order cards must copy passenger card footer structure')
assert(ordersPanel.includes('class="price"'), 'driver order cards must emphasize price like passenger orders')
assert(ordersPanel.includes('class="actions"'), 'driver order cards must use passenger-style action row')
assert(home.includes("router.replace('/login')"), 'driver logout must navigate inside independent app')
assert(home.includes('van-tabbar'), 'driver H5 navigation must use fixed mobile bottom tabbar')
assert(!home.includes('DriverAssetsPanel'), 'driver home must not keep the old assets tab shell')
assert(home.includes('driver-status-bar'), 'driver home must show a top status bar')
assert(home.includes('<section v-show="activeTab === 0" class="home-workbench">'), 'driver home map container must stay mounted while switching tabs')
assert(!home.includes('<section v-if="activeTab === 0" class="home-workbench">'), 'driver home map container must not be destroyed by tab switching')
assert(home.includes('class="status-toggle"'), 'driver home top status bar must expose online/offline toggle button')
assert(home.includes('home-map-stage'), 'driver home must use a central map stage')
assert(home.includes('home-floating-panel'), 'driver home must use a floating bottom operation panel')
assert(home.includes('listenModeEnabled'), 'driver home must expose a listening mode switch')
assert(home.includes('selectedHomeOrder'), 'driver home must select one nearby/current order for the bottom panel')
assert(home.includes('previewHomeOrder'), 'driver home must support preview before accepting an order')
assert(home.includes('navigateToPickup'), 'driver home must support one-tap navigation to pickup')
assert(home.includes('navigateHomeRoute'), 'driver home must support manual home route mode')
assert(home.includes('reportAbnormal'), 'driver home must provide an abnormal report entry while driving')
assert(home.includes('AMapLoader'), 'driver home must own live map loading after workbench merge')
assert(home.includes('startAcceptingOrders'), 'driver home idle primary action must explicitly start accepting orders')
assert(home.includes('<span>开始接单</span>'), 'driver home idle primary action text must remain 开始接单')
assert(!home.includes('workActionText'), 'driver home idle primary action must not revert to dynamic stop-listening text')
assert(home.includes('.go-online { border: 0; background: #5B5CFF;'), 'driver home start accepting button must keep purple primary color')
assert(home.includes('workStatusPayload()'), 'online/offline actions must send device and location payload')
assert(home.includes('safeApiCall'), 'driver H5 actions must keep UI usable when API calls fail')
assert(home.includes('showIncomeLoadFailure'), 'driver income API failures must open a failure dialog')
assert(home.includes('formatDriverStatus') && home.includes('activeTab.value === 2'), 'driver asset tab must pass driver status formatter to mine panel')
assert(home.includes('rejectOrder(orderId, reason)'), 'driver reject action must submit explicit reject reason')
assert(!home.includes('actualPriceCents'), 'driver finish trip must not ask driver to enter settlement amount')
assert(app.includes('driver-phone-shell'), 'driver app must render a centered phone shell for H5 preview')
assert(login.includes('class="login-card"'), 'driver login must use a compact H5 login card')
assert(login.includes('class="auth-mode-tabs"'), 'driver login must expose touch-friendly H5 auth mode tabs')
assert(!login.includes('SmsCodeDialog'), 'driver login must not mount SMS code dialog')
assert(login.includes('验证码已发送（联调验证码见服务端日志）'), 'driver login must not reveal plaintext code')
assert(!login.includes('registerForm.avatarUrl'), 'driver register form must not expose a dead avatar URL field')
assert(!login.includes('name="avatarUrl"'), 'driver register form must not include an avatar field without a public upload flow')
assert(home.includes('@click="openProfileEdit"'), 'driver home avatar must navigate to profile edit page')
assert(home.includes("router.push('/profile/edit')"), 'driver home avatar click must open dedicated profile edit route')
assert(home.includes('class="home-workbench"'), 'driver home must start with the merged workbench')
assert(home.includes('class="map-floating-actions"'), 'driver home must expose map quick action controls')
assert(home.includes('class="tab-panel-scroll"'), 'driver tab content must use a mobile scroll panel')
assert(home.includes('@click="openHeatmap"'), 'driver heatmap button must open a popup')
assert(home.includes('v-model:show="heatmapVisible"'), 'driver heatmap must render in a popup, not inline')
assert(!home.includes(':teleport="false"'), 'driver popups must not pass boolean false to Vant teleport')
assert(home.includes('teleport="#driver-home-popups"'), 'driver heatmap popup must stay inside the driver H5 shell via selector teleport')
assert(driverApp.includes('id="driver-home-popups"'), 'driver popup teleport target must be mounted outside DriverHome')
assert(!home.includes('id="driver-home-popups"'), 'DriverHome must not create its own popup teleport target')
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
assert(!home.includes('@/../user'), 'driver maps must not import user app internals')
assert(home.includes('@amap/amap-jsapi-loader'), 'driver home map must use the same AMap loader as user app')
assert(home.includes("from '@/config/amap'"), 'driver home map must use driver-owned AMap config')
assert(home.includes('watchPosition'), 'driver home map must follow live browser location')
assert(home.includes('const currentLocation = await ensureWorkLocation()'), 'driver home map must locate before creating the initial map')
assert(home.includes('getConnectedHomeMapContainer'), 'driver home map must verify the container is still mounted before AMap init')
assert(home.includes('container.isConnected'), 'driver home map must not initialize AMap on a detached container')
assert(home.includes('syncHomeMapCenterToDriver'), 'driver home map must recenter on the driver when browser location updates')
assert(home.includes('.Polyline'), 'driver home map must render route direction or live movement path')
assert(!home.includes('homeDetectionCircle'), 'driver home map must delete the old AMap.Circle detection ring state')
assert(!home.includes('renderHomeDetectionCircle'), 'driver home map must delete the old AMap.Circle detection ring renderer')
assert(!home.includes('new homeAMap.Circle'), 'driver home map must not draw the pickup detection radius with AMap.Circle')
assert(home.includes('homeDriverPulseContent'), 'driver home marker must render the copied passenger pulse content')
assert(home.includes('class="driver-location-pulse"'), 'driver home marker must include the driver pulse container')
assert(home.includes('class="pulse-ring"'), 'driver home marker must include the first copied passenger pulse ring')
assert(home.includes('class="pulse-ring delay"'), 'driver home marker must include the delayed copied passenger pulse ring')
assert(home.includes('@keyframes pulse-ring'), 'driver home marker must include the copied passenger pulse animation')
assert(refreshHomeWorkbenchSource.includes('loadNearbyOrders(nearbyOrderPage.value, { silentError: true })'), 'driver home refresh must fetch nearby available orders for homepage detection')
assert(refreshHomeWorkbenchSource.includes('refreshHomeHeatmap()'), 'driver home refresh must still fetch heatmap data')
assert(!handlePushMessageSource.includes('if (activeTab.value === 1) void loadNearbyOrders'), 'driver dispatch push must not refresh nearby orders only on the order tab')
assert(handlePushMessageSource.includes('void loadNearbyOrders(nearbyOrderPage.value, { silentError: true })'), 'driver dispatch push must refresh nearby orders for homepage detection')
assert(!home.includes("driverStore.setCurrentOrder(order, Number(order.status) === 3 ? 'trip' : 'pickup')"), 'dispatch pushes must not mark pending dispatch as current order')
assert(home.includes('dispatchStatus: item.dispatch?.status'), 'dispatch rows must expose dispatch status')
assert(home.includes('orderId: item.dispatch?.orderId || item.order?.orderId'), 'dispatch rows must expose order id from dispatch payload')
assert(orderLogic.includes('if len(orderIDStrs) == 0 {'), 'available orders must not fall back to global wait-accept list')
assert(driverAssets.includes('getTodayIncome'), 'driver wallet must call backend today income endpoint')
assert(driverAssets.includes('getWeekIncome'), 'driver wallet must call backend week income endpoint')
assert(home.includes('/api/driver/v1/ws'), 'driver WebSocket must connect to registered /api/driver/v1/ws route')
assert(!home.includes('/api/driver/v1/push/ws'), 'driver WebSocket must not use stale /api/driver/v1/push/ws route')
assert(!home.includes('openWorkbench'), 'driver home must not navigate to a separate workbench page')
assert(home.includes('nearbyOrders'), 'driver home must own nearby order state for the order tab')
assert(home.includes('const nearbyOrderPageSize = 5'), 'driver order tab nearby orders must show 5 items per inline page')
assert(home.includes('const nearbyOrderExpandedPageSize = 10'), 'driver expanded nearby order popup must show 10 items per page')
assert(home.includes('nearbyOrderPopupVisible'), 'driver order tab must expose expanded nearby order popup state')
assert(home.includes('openNearbyOrderPopup'), 'driver order tab must open a dedicated nearby order popup')
assert(ordersPanel.includes('class="nearby-order-section"'), 'driver orders page must render nearby orders on the order tab')
assert(ordersPanel.includes('class="nearby-order-popup"'), 'driver orders page must render expanded nearby orders in an H5 popup')
assert(ordersPanel.includes("'update:nearby-order-popup-visible'"), 'driver nearby order popup must emit the kebab-case update event listened by DriverHome')
assert(ordersPanel.includes("@click=\"emit('order-action', 'accept', order)\""), 'nearby order accept button must emit order-action with the order object')
assert(ordersPanel.includes("@click=\"emit('order-action', 'reject', order)\""), 'nearby order reject button must emit order-action with the order object')
assert(ordersPanel.includes("@click=\"emit('open-nearby-popup')\""), 'driver order tab must provide an expand button for nearby orders')
assert(home.includes("driverStore.tripPhase === 'pickup'"), 'driver home accepted pickup phase must switch the bottom panel to driving/navigation mode')
assert(home.includes('const actionResult = await config.request()'), 'driver order actions must read the backend action result')
assert(home.includes('mergeOrderActionResult'), 'driver order actions must merge backend status into the current order')
assert(home.includes("driverStore.setCurrentOrder(config.phase === 'idle' ? null : nextOrder, config.phase)"), 'driver order actions must store the updated accepted order after accept')
assert(!profilePanel.includes('v-model="profileForm.avatarUrl"'), 'mine profile panel must not expose inline avatar URL editing')
assert(!minePanel.includes('@submit-profile'), 'mine panel must not submit profile edits inline')
assert(profileEdit.includes('accept="image/*"'), 'profile edit page must choose avatar images from local gallery')
assert(profileEdit.includes('class="avatar-gallery-button"'), 'profile edit page must show an explicit local gallery avatar button')
assert(profileEdit.includes('@click="chooseAvatar"'), 'profile edit avatar button must trigger local gallery input')
assert(profileEdit.includes('getAvatarUploadToken'), 'profile edit page must request qiniu token before saving avatar')
assert(profileEdit.includes('uploadToQiniu'), 'profile edit page must direct-upload selected avatar to qiniu')
assert(profileEdit.includes('buildAvatarUrl'), 'profile edit page must save the qiniu CDN avatar URL')
assert(!profileEdit.includes('uploadDriverAvatar'), 'profile edit page must not upload avatar base64 to local driver API')
assert(!profileEdit.includes('fileToBase64'), 'profile edit page must not convert selected avatar to base64 for local upload')
assert(profileEdit.includes('driverStore.saveProfile'), 'profile edit page must save all editable driver profile fields')
assert(profileEdit.includes('profileForm.phone'), 'profile edit page must allow phone updates')
assert(profileEdit.includes('profileForm.idCardNo'), 'profile edit page must allow ID card updates')
assert(profileEdit.includes('profileForm.driverLicenseNo'), 'profile edit page must allow driver license updates')
assert(walletPage.includes("from '@/components/driver-home/DriverWalletPanel.vue'"), 'wallet page must reuse the wallet panel')
assert(vehiclePage.includes("from '@/components/driver-home/DriverVehiclePanel.vue'"), 'vehicle page must reuse the vehicle panel')
assert(certificationPage.includes("from '@/components/driver-home/DriverCertificationPanel.vue'"), 'certification page must reuse the certification panel')
assert(incomePage.includes('incomeBills'), 'income page must render income details independently')
assert(orderRecordsPage.includes('listDriverOrders'), 'order records page must load driver order history')
assert(driverRouter.includes("path: '/mine/wallet'"), 'driver router must expose wallet page')
assert(driverRouter.includes("path: '/mine/vehicle'"), 'driver router must expose vehicle page')
assert(driverRouter.includes("path: '/mine/certification'"), 'driver router must expose certification page')
assert(driverRouter.includes("path: '/mine/income'"), 'driver router must expose income page')
assert(driverRouter.includes("path: '/mine/orders'"), 'driver router must expose order records page')
assert(driverMineData.includes("route: '/mine/wallet'"), 'mine menu must link wallet to a standalone page')
assert(driverMineData.includes("route: '/mine/vehicle'"), 'mine menu must link vehicle to a standalone page')
assert(driverMineData.includes("route: '/mine/certification'"), 'mine menu must link certification to a standalone page')
assert(driverMineData.includes("route: '/mine/income'"), 'mine menu must link income to a standalone page')
assert(driverMineData.includes("route: '/mine/orders'"), 'mine menu must link order records to a standalone page')
assert(!existsSync(file('web/driver/src/components/driver-home/DriverAssetsPanel.vue')), 'old assets panel must be removed')

console.log('driver frontend isolation checks passed')
