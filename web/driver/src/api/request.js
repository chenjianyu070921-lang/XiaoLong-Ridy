import axios from 'axios'
import { showToast } from 'vant'
import router from '@/router'

const driverRequest = axios.create({
  baseURL: '/api/driver/v1',
  timeout: 15000,
  headers: {
    'Content-Type': 'application/json'
  }
})

driverRequest.interceptors.request.use((config) => {
  const token = localStorage.getItem('driverToken')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

driverRequest.interceptors.response.use(
  (response) => {
    const res = response.data
    if (res.code === 0 || res.code === 200) {
      const payload = res.data
      // 成功但 data 为 null 时降级为空对象，避免下游（persistSession / refreshProfile 等）
      // 直接访问 null.driver / null.token 触发 Cannot read properties of null。
      return payload !== undefined ? (payload ?? {}) : res
    }
    const message = res.message || '请求失败'
    if (!response.config?.silentError) {
      showToast(message)
    }
    return Promise.reject(new Error(message))
  },
  (error) => {
    const status = error.response?.status
    const message = error.response?.data?.message || error.message || '网络连接失败'
    if (status === 401) {
      clearDriverSession()
      router.push('/login')
      showToast('司机登录已过期，请重新登录')
    } else if (!error.config?.silentError) {
      showToast(message)
    }
    return Promise.reject(error)
  }
)

function clearDriverSession() {
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

export default driverRequest
