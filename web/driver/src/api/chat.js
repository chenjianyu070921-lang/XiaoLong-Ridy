import axios from 'axios'
import { showToast } from 'vant'
import router from '@/router'

// 聊天网关专属实例：baseURL 指向 /api/chat/v1（vite 代理到 api/chat:18090）。
// 响应信封与司机端统一：code===0 时返回 res.data；非 0 弹 toast 并 reject。
const chatRequest = axios.create({
  baseURL: '/api/chat/v1',
  timeout: 15000,
  headers: { 'Content-Type': 'application/json' }
})

chatRequest.interceptors.request.use((config) => {
  const token = localStorage.getItem('driverToken')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

chatRequest.interceptors.response.use(
  (response) => {
    const res = response.data
    if (res.code === 0 || res.code === 200) {
      return res.data !== undefined ? (res.data ?? {}) : res
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
      for (const key of [
        'driverToken', 'driverProfile', 'driverOnlineStatus', 'driverVehicle',
        'driverVehicleId', 'driverCertification', 'driverCurrentOrder',
        'driverCurrentOrderId', 'driverTripPhase'
      ]) {
        localStorage.removeItem(key)
      }
      router.push('/login')
      showToast('司机登录已过期，请重新登录')
    } else if (!error.config?.silentError) {
      showToast(message)
    }
    return Promise.reject(error)
  }
)

export function getConversation(orderId) {
  return chatRequest.get('/conversation', { params: { orderId } })
}

export function listMessages(conversationId, cursor = 0, limit = 20) {
  return chatRequest.get('/messages', { params: { conversationId, cursor, limit } })
}

export function sendMessage(conversationId, msgType, content, clientMsgId) {
  return chatRequest.post('/send', { conversationId, msgType, content, clientMsgId })
}

export function markRead(conversationId) {
  return chatRequest.post('/read', { conversationId })
}

// 生成客户端幂等 ID（与后端 uk_client_msg 配合，避免重复提交）。
export function genClientMsgId() {
  if (typeof crypto !== 'undefined' && crypto.randomUUID) {
    return crypto.randomUUID()
  }
  return `c_${Date.now()}_${Math.random().toString(36).slice(2, 10)}`
}

export default chatRequest
