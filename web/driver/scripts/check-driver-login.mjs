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

const login = read('src/views/DriverLogin.vue')
const viteConfig = read('vite.config.js')

assert(!/PENDING[\s\S]{0,240}driverStore\.logout\(\)/.test(login), 'pending driver login must keep the session')
assert(login.includes("router.replace('/home')"), 'successful driver login must enter home')
assert(login.includes('可补充车辆和资质信息'), 'pending driver login must explain profile completion')
assert(functionSource(login, 'handleRegister').includes('validatePassword(registerForm.password)'), 'driver register must validate password length before submitting')
assert(functionSource(login, 'validatePassword').includes('value.length >= 8'), 'driver register password validation must require at least 8 characters')

assert(viteConfig.includes('VITE_DRIVER_API_TARGET'), 'driver API proxy target must be configurable for LAN/dev machines')
assert(viteConfig.includes('http://127.0.0.1:18082'), 'driver API proxy must default to local api/driver for localhost demos')
assert(!viteConfig.includes("target: 'http://localhost:8082'"), 'driver API proxy must not hard-code localhost')

console.log('driver login checks passed')
