import axios from 'axios'
import { ElMessage } from 'element-plus'

// 统一的后台 API 客户端：
// - baseURL 走相对路径，由 Vite dev proxy 转发到 api/admin(8083)，规避 CORS
// - 请求拦截器注入 Bearer Token
// - 响应拦截器解包 {code, message, data}，code!==0 统一报错，40004 跳登录
const service = axios.create({
  baseURL: '/admin/v1',
  timeout: 15000,
})

service.interceptors.request.use((config) => {
  const token = localStorage.getItem('admin_token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

service.interceptors.response.use(
  (response) => {
    const res = response.data
    if (res && res.code === 0) {
      return res.data
    }
    // 未登录或 token 失效：清理本地态并回登录页
    if (res && res.code === 40004) {
      localStorage.removeItem('admin_token')
      localStorage.removeItem('admin_info')
      window.location.hash = '#/login'
    }
    ElMessage.error(res?.message || '请求失败')
    return Promise.reject(new Error(res?.message || '请求失败'))
  },
  (error) => {
    const msg = error.response?.data?.message || error.message || '网络错误'
    ElMessage.error(msg)
    return Promise.reject(error)
  },
)

export default service
