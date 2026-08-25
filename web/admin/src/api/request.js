import axios from 'axios'
import { ElMessage } from 'element-plus'

// 统一的后台 API 客户端：
// - baseURL 走相对路径，由 Vite dev proxy 转发到 api/admin(8717)，规避 CORS
// - 请求拦截器注入 Bearer Token
// - 响应拦截器解包 {code, message, data}，code!==0 统一报错，401/40004 跳登录
const service = axios.create({
  baseURL: '/admin/v1',
  timeout: 15000,
})

// 会话失效跳转防重入：并发请求同时 401 时只清理/跳转一次
let redirectingToLogin = false
const handleUnauthorized = () => {
  if (redirectingToLogin) return
  redirectingToLogin = true
  localStorage.removeItem('admin_token')
  localStorage.removeItem('admin_info')
  window.location.hash = '#/login'
  setTimeout(() => {
    redirectingToLogin = false
  }, 1000)
}

service.interceptors.request.use((config) => {
  const token = localStorage.getItem('admin_token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

service.interceptors.response.use(
  (response) => {
    // CSV 下载接口返回二进制流，不遵循 {code,message,data} JSON 包装，直接交给调用方处理。
    if (response.config.responseType === 'blob') {
      return response
    }
    const res = response.data
    if (res && res.code === 0) {
      return res.data
    }
    // 防御分支：若后端以 HTTP 200 返回 40004（当前实现不会，但保留契约兼容）
    if (res && res.code === 40004) {
      handleUnauthorized()
    }
    ElMessage.error(res?.message || '请求失败')
    return Promise.reject(new Error(res?.message || '请求失败'))
  },
  (error) => {
    // 后端 token 失效返回 HTTP 401 + code 40004；axios 默认把非 2xx 路由到本分支
    const data = error.response?.data
    if (error.response?.status === 401 || data?.code === 40004) {
      handleUnauthorized()
    }
    const msg = data?.message || error.message || '网络错误'
    ElMessage.error(msg)
    return Promise.reject(error)
  },
)

export default service

// download 复用同一 Axios 实例，确保导出下载请求自动携带当前管理员 Bearer Token。
export const download = (url) => service.get(url, { responseType: 'blob' })
