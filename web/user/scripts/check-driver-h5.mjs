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

function read(path) {
  return readFileSync(file(path), 'utf8')
}

const requiredFiles = [
  'web/user/src/api/driver.js',
  'web/user/src/stores/driver.js',
  'web/user/src/views/driver/DriverLogin.vue',
  'web/user/src/views/driver/DriverHome.vue',
]

for (const path of requiredFiles) {
  assert(existsSync(file(path)), `missing migrated H5 file: ${path}`)
}

const router = read('web/user/src/router/index.js')
assert(router.includes("redirect: '/driver/login'"), 'default H5 entry must redirect to driver login')
for (const route of ['/driver', '/driver/login', '/driver/home']) {
  assert(router.includes(`path: '${route}'`), `missing driver H5 route: ${route}`)
}
assert(router.includes('requiresDriverAuth'), 'driver H5 routes must use a driver auth guard')

const viteConfig = read('web/user/vite.config.js')
assert(viteConfig.includes("'/api/driver'"), 'vite proxy must route driver API separately')
assert(viteConfig.includes('8082'), 'driver API proxy must target the driver API port')

const api = read('web/user/src/api/driver.js')
const endpoints = [
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
  '/reviews/list',
]
for (const endpoint of endpoints) {
  assert(api.includes(endpoint), `missing real driver API wrapper: ${endpoint}`)
}
assert(!api.includes("baseURL: '/driver'"), 'driver H5 must not call the removed web proxy')
assert(api.includes('setDriverOnline(data'), 'online API wrapper must accept payload required by backend')
assert(api.includes('setDriverOffline(data'), 'offline API wrapper must accept payload required by backend')

const home = read('web/user/src/views/driver/DriverHome.vue')
const login = read('web/user/src/views/driver/DriverLogin.vue')
const app = read('web/user/src/App.vue')
assert(app.includes('min(100vw, 375px)'), 'global H5 shell must cap width at phone size')
assert(login.includes(": '密码登录'"), 'password login submit button must visibly say 密码登录')
assert(login.includes(": '验证码登录'"), 'sms login submit button must visibly say 验证码登录')
assert(!login.includes('/splash'), 'driver login must not navigate back to passenger splash')
assert(login.includes('apiErrorMessage'), 'driver login must show request failures and clear loading state')
assert(login.includes('closeToast()'), 'driver login must close loading toast on failed requests too')
for (const marker of [
  'loadIncome',
  'loadAvailableOrders',
  'loadReviews',
  'loadTrajectory',
  'submitFinishTrip',
  'submitCertification',
  'submitVehicle',
  'setOnline',
  'setOffline',
]) {
  assert(home.includes(marker), `missing migrated mobile function: ${marker}`)
}
assert(home.includes('van-tabbar'), 'driver H5 navigation must use a fixed mobile bottom tabbar')
assert(!home.includes('<van-tabs'), 'driver H5 must not use middle/top van-tabs navigation')
assert(home.includes('trajectory-panel'), 'trajectory area must have a compact mobile panel')
assert(home.includes('class="trajectory-error"'), 'trajectory area must render request errors such as Network Error')
assert(home.includes('trajectory-action'), 'trajectory query button must be full-width touch target')
assert(home.includes('暂无轨迹点'), 'trajectory empty state must keep 暂无轨迹点 copy')
assert(home.includes('workStatusPayload()'), 'online/offline actions must send device and location payload')
assert(home.includes('safeApiCall'), 'driver H5 actions must keep UI usable when an API call fails')

for (const legacyPath of [
  'web/driver/main.go',
  'web/driver/main_test.go',
  'web/driver/templates/index.html',
  'web/driver/static/js/app.js',
  'web/driver/static/css/app.css',
]) {
  assert(!existsSync(file(legacyPath)), `legacy web driver file still exists: ${legacyPath}`)
}

console.log('driver H5 migration checks passed')
