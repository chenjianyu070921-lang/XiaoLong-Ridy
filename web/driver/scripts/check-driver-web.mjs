import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const root = path.dirname(path.dirname(fileURLToPath(import.meta.url)))

function read(relativePath) {
  return fs.readFileSync(path.join(root, relativePath), 'utf8')
}

function assert(condition, message) {
  if (!condition) {
    throw new Error(message)
  }
}

const requiredFiles = [
  'package.json',
  'index.html',
  'vite.config.js',
  'src/main.js',
  'src/App.vue',
  'src/router/index.js',
  'src/api/driver.js',
  'src/stores/driver.js',
  'src/views/DriverLogin.vue',
  'src/views/DriverHome.vue'
]

for (const file of requiredFiles) {
  assert(fs.existsSync(path.join(root, file)), `missing driver web file: ${file}`)
}

const packageJson = JSON.parse(read('package.json'))
assert(packageJson.name === 'xiaolong-ridy-driver', 'driver package must be independently named')
assert(packageJson.scripts?.dev?.includes('5175'), 'driver dev server must use its own port 5175')

const viteConfig = read('vite.config.js')
assert(viteConfig.includes("target: 'http://localhost:8082'"), 'driver proxy must target driver API 8082')
assert(!viteConfig.includes('/api/passenger'), 'driver app must not proxy passenger API')

const router = read('src/router/index.js')
assert(router.includes("path: '/login'"), 'driver app must expose /login')
assert(router.includes("path: '/home'"), 'driver app must expose /home')
assert(router.includes('requiresDriverAuth'), 'driver home route must require driver auth')
assert(!router.includes('/driver/login'), 'driver app must not keep old /driver/login route')
assert(!router.includes('/driver/home'), 'driver app must not keep old /driver/home route')
assert(!router.includes('useUserStore'), 'driver router must not depend on passenger user store')

const api = read('src/api/driver.js')
assert(api.includes("baseURL: '/api/driver/v1'"), 'driver API must use driver API base URL')
assert(api.includes("router.push('/login')"), 'driver API auth expiry must route to /login')
assert(api.includes('listAvailableOrders'), 'driver API must keep available orders API')
assert(api.includes('listIncomeBills'), 'driver API must keep income bills API')

const store = read('src/stores/driver.js')
assert(store.includes("from '@/api/driver'"), 'driver store must use local driver API wrapper')

const login = read('src/views/DriverLogin.vue')
assert(login.includes("router.replace('/home')"), 'driver login success must route to /home')

const home = read('src/views/DriverHome.vue')
assert(home.includes("router.replace('/login')"), 'driver logout must route to /login')
assert(!home.includes('realtime-price'), 'driver app must not include realtime price feature')

console.log('driver web split checks passed')
