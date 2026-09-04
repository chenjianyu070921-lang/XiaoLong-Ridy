import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import {
  getDriver,
  loginDriverByPassword,
  loginDriverBySMS,
  registerDriver,
  updateDriver
} from '@/api/driver'
import { resolveDriverVehicleId } from '@/utils/vehicle'

function readJSON(key, fallback = null) {
  try {
    return JSON.parse(localStorage.getItem(key) || 'null') ?? fallback
  } catch {
    return fallback
  }
}

export const useDriverStore = defineStore('driver', () => {
  const token = ref(localStorage.getItem('driverToken') || '')
  const driver = ref(readJSON('driverProfile', {}))
  const vehicle = ref(readJSON('driverVehicle', null))
  const vehicleId = ref(Number(localStorage.getItem('driverVehicleId') || vehicle.value?.id || 0))
  const certification = ref(readJSON('driverCertification', null))
  const onlineStatus = ref(Number(driver.value?.onlineStatus ?? localStorage.getItem('driverOnlineStatus') ?? 0))
  const currentOrder = ref(readJSON('driverCurrentOrder', null))
  const currentOrderId = ref(localStorage.getItem('driverCurrentOrderId') || '')
  const tripPhase = ref(localStorage.getItem('driverTripPhase') || 'idle')

  const isLoggedIn = computed(() => !!token.value)
  const driverId = computed(() => Number(driver.value?.id || driver.value?.driverId || 0))
  const displayName = computed(() => driver.value?.realName || driver.value?.nickname || '司机')

  // 夜间模式：darkMode 持久化到 localStorage，并同步到 <html class="dark">，
  // 由 App.vue 的 .dark CSS 变量与全局兜底规则生效。
  const darkMode = ref(localStorage.getItem('driverDarkMode') === '1')
  function applyDarkMode() {
    const root = document.documentElement
    if (darkMode.value) root.classList.add('dark')
    else root.classList.remove('dark')
  }
  function setDarkMode(value) {
    darkMode.value = !!value
    localStorage.setItem('driverDarkMode', darkMode.value ? '1' : '0')
    applyDarkMode()
  }
  function toggleDarkMode() {
    setDarkMode(!darkMode.value)
  }
  applyDarkMode()

  function persistSession(data = {}) {
    token.value = data.token || token.value
    driver.value = data.driver || driver.value || {}
    onlineStatus.value = Number(driver.value?.onlineStatus ?? onlineStatus.value ?? 0)
    vehicleId.value = resolveDriverVehicleId(driver.value, vehicle.value)
    if (token.value) localStorage.setItem('driverToken', token.value)
    localStorage.setItem('driverProfile', JSON.stringify(driver.value))
    localStorage.setItem('driverOnlineStatus', String(onlineStatus.value))
    if (vehicleId.value > 0) localStorage.setItem('driverVehicleId', String(vehicleId.value))
  }

  async function loginPassword(phone, password, config = {}) {
    const res = await loginDriverByPassword(phone, password, config)
    persistSession(res)
    return res
  }

  async function loginSMS(phone, code, config = {}) {
    const res = await loginDriverBySMS(phone, code, config)
    persistSession(res)
    return res
  }

  async function register(data, config = {}) {
    return registerDriver(data, config)
  }

  async function refreshProfile(config = {}) {
    const res = await getDriver(config)
    driver.value = res.driver || res
    onlineStatus.value = Number(driver.value?.onlineStatus ?? onlineStatus.value ?? 0)
    vehicleId.value = resolveDriverVehicleId(driver.value, vehicle.value)
    localStorage.setItem('driverProfile', JSON.stringify(driver.value))
    localStorage.setItem('driverOnlineStatus', String(onlineStatus.value))
    if (vehicleId.value > 0) localStorage.setItem('driverVehicleId', String(vehicleId.value))
    return res
  }

  async function saveProfile(payload, config = {}) {
    if (!driverId.value) {
      throw new Error('请先重新登录后再保存资料')
    }
    const body = { ...payload, id: driverId.value }
    const res = await updateDriver(body, config)
    driver.value = { ...(driver.value || {}), ...payload }
    localStorage.setItem('driverProfile', JSON.stringify(driver.value))
    try {
      // 静默：刷新资料只是保存后的补偿动作，失败不打扰用户（提示由调用方负责）
      await refreshProfile({ silentError: true })
    } catch {
      // keep locally updated profile when refresh is unavailable
    }
    return res
  }

  function setVehicle(nextVehicle) {
    vehicle.value = nextVehicle
    vehicleId.value = Number(nextVehicle?.id || 0)
    if (nextVehicle) {
      localStorage.setItem('driverVehicle', JSON.stringify(nextVehicle))
      localStorage.setItem('driverVehicleId', String(vehicleId.value))
    } else {
      localStorage.removeItem('driverVehicle')
      localStorage.removeItem('driverVehicleId')
    }
  }

  function setCertification(nextCertification) {
    certification.value = nextCertification
    if (nextCertification) {
      localStorage.setItem('driverCertification', JSON.stringify(nextCertification))
    } else {
      localStorage.removeItem('driverCertification')
    }
  }

  function setWorkState(status) {
    onlineStatus.value = Number(status || 0)
    driver.value = { ...(driver.value || {}), onlineStatus: onlineStatus.value }
    localStorage.setItem('driverOnlineStatus', String(onlineStatus.value))
    localStorage.setItem('driverProfile', JSON.stringify(driver.value))
  }

  function setCurrentOrder(order, phase = tripPhase.value) {
    currentOrder.value = order || null
    currentOrderId.value = order?.orderId ? String(order.orderId) : ''
    tripPhase.value = phase || 'idle'
    localStorage.setItem('driverTripPhase', tripPhase.value)
    if (currentOrder.value && currentOrderId.value) {
      localStorage.setItem('driverCurrentOrderId', currentOrderId.value)
      localStorage.setItem('driverCurrentOrder', JSON.stringify(currentOrder.value))
    } else {
      localStorage.removeItem('driverCurrentOrderId')
      localStorage.removeItem('driverCurrentOrder')
    }
  }

  function logout() {
    token.value = ''
    driver.value = {}
    vehicle.value = null
    vehicleId.value = 0
    certification.value = null
    onlineStatus.value = 0
    currentOrder.value = null
    currentOrderId.value = ''
    tripPhase.value = 'idle'
    for (const key of [
      'driverToken',
      'driverProfile',
      'driverOnlineStatus',
      'driverVehicle',
      'driverVehicleId',
      'driverCertification',
      'driverCurrentOrder',
      'driverCurrentOrderId',
      'driverTripPhase'
    ]) {
      localStorage.removeItem(key)
    }
  }

  return {
    token,
    driver,
    vehicle,
    vehicleId,
    certification,
    onlineStatus,
    currentOrder,
    currentOrderId,
    tripPhase,
    isLoggedIn,
    driverId,
    displayName,
    loginPassword,
    loginSMS,
    register,
    refreshProfile,
    saveProfile,
    setVehicle,
    setCertification,
    setWorkState,
    setCurrentOrder,
    darkMode,
    setDarkMode,
    toggleDarkMode,
    logout
  }
})
