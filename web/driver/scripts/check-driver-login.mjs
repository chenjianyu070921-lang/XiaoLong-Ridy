import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

const root = resolve(import.meta.dirname, '..')

function read(path) {
  return readFileSync(resolve(root, path), 'utf8')
}

function assert(condition, message) {
  if (!condition) {
    throw new Error(message)
  }
}

const login = read('src/views/DriverLogin.vue')
const viteConfig = read('vite.config.js')

assert(!/PENDING[\s\S]{0,240}driverStore\.logout\(\)/.test(login), 'pending driver login must keep the session')
assert(login.includes("router.replace('/home')"), 'successful driver login must enter home')
assert(login.includes('可补充车辆和资质信息'), 'pending driver login must explain profile completion')

assert(viteConfig.includes('VITE_DRIVER_API_TARGET'), 'driver API proxy target must be configurable for LAN/dev machines')
assert(viteConfig.includes('http://127.0.0.1:8082'), 'driver API proxy must default to local api/driver for localhost demos')
assert(!viteConfig.includes("target: 'http://localhost:8082'"), 'driver API proxy must not hard-code localhost')

console.log('driver login checks passed')
