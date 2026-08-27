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
      return res.data !== undefined ? res.data : res
    }
    if (!response.config?.silentError) {
      showToast(res.message || '请求失败')
    }
    return Promise.reject(new Error(res.message || '请求失败'))
  },
  (error) => {
    const status = error.response?.status
    const message = error.response?.data?.message || error.message || '网络连接失败'
    if (status === 401) {
      localStorage.removeItem('driverToken')
      localStorage.removeItem('driverProfile')
      localStorage.removeItem('driverVehicle')
      localStorage.removeItem('driverVehicleId')
      localStorage.removeItem('driverCertification')
      localStorage.removeItem('driverCurrentOrder')
      localStorage.removeItem('driverCurrentOrderId')
      localStorage.removeItem('driverTripPhase')
      router.push('/login')
      showToast('司机登录已过期，请重新登录')
    } else {
      if (!error.config?.silentError) {
        showToast(message)
      }
    }
    return Promise.reject(error)
  }
)

export function sendDriverSMSCode(phone, config = {}) {
  return driverRequest.post('/auth/send-sms-code', { phone }, config)
}

export function loginDriverByPassword(phone, password, config = {}) {
  return driverRequest.post('/auth/login-by-password', { phone, password }, config)
}

export function loginDriverBySMS(phone, code, config = {}) {
  return driverRequest.post('/auth/login-by-sms', { phone, code }, config)
}

export function registerDriver(data, config = {}) {
  return driverRequest.post('/drivers/register', data, config)
}

export function updateDriver(data) {
  return driverRequest.post('/drivers/update', data)
}

export function getDriver(config = {}) {
  return driverRequest.get('/drivers/get', config)
}

export function getDriverAiScore(config = {}) {
  return driverRequest.get('/drivers/ai-score', config)
}

export function setDriverOnline(data = {}) {
  return driverRequest.post('/drivers/online', data)
}

export function setDriverOffline(data = {}) {
  return driverRequest.post('/drivers/offline', data)
}

export function heartbeatDriver(data) {
  return driverRequest.post('/drivers/heartbeat', data)
}

export function reportDriverLocation(data) {
  return driverRequest.post('/drivers/location/report', data)
}

export function createVehicle(data) {
  return driverRequest.post('/vehicles', data)
}

export function getVehicle(id, config = {}) {
  return driverRequest.get('/vehicles/get', { ...config, params: { id, ...(config.params || {}) } })
}

export function updateVehicle(data) {
  return driverRequest.post('/vehicles/update', data)
}

export function deleteVehicle(id) {
  return driverRequest.post('/vehicles/delete', { id })
}

export function uploadCertification(data) {
  return driverRequest.post('/drivers/certification/upload', data)
}

export function getCertification(config = {}) {
  return driverRequest.get('/drivers/certification', config)
}

export function getIncomeSummary(config = {}) {
  return driverRequest.get('/income/summary', config)
}

export function getWalletSummary(config = {}) {
  return driverRequest.get('/wallet/summary', config)
}

export function listIncomeBills(data = {}, config = {}) {
  return driverRequest.post('/income/bills', data, config)
}

export function acceptOrder(orderId) {
  return driverRequest.post('/orders/accept', { orderId })
}

export function rejectOrder(orderId, reason = '司机主动拒单') {
  return driverRequest.post('/orders/reject', { orderId, reason })
}

export function confirmArrive(orderId) {
  return driverRequest.post('/orders/confirm-arrive', { orderId })
}

export function startTrip(orderId) {
  return driverRequest.post('/orders/start-trip', { orderId })
}

export function finishTrip(data) {
  return driverRequest.post('/orders/finish-trip', data)
}

export function listAvailableOrders(data = {}, config = {}) {
  return driverRequest.post('/orders/available', data, config)
}

export function getDriverOrderDetail(orderId) {
  return driverRequest.post('/orders/detail', { orderId })
}

export function getOrderTrajectory(orderId, config = {}) {
  return driverRequest.post('/orders/trajectory', { orderId }, config)
}

export function listDriverOrders(data = {}, config = {}) {
  return driverRequest.post('/orders/list', data, config)
}

export function listDriverDispatches(data = {}, config = {}) {
  return driverRequest.post('/orders/dispatches', data, config)
}

export function listPassengerReviews(data = {}, config = {}) {
  return driverRequest.post('/reviews/list', data, config)
}

export default driverRequest
