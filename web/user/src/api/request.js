import axios from 'axios'
import { useUserStore } from '@/stores/user'
import { showToast } from 'vant'
import router from '@/router'

const request = axios.create({
  baseURL: '/api/passenger/v1',
  timeout: 15000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// 请求拦截器
request.interceptors.request.use(
  (config) => {
    const userStore = useUserStore()
    if (userStore.token) {
      config.headers.Authorization = `Bearer ${userStore.token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 响应拦截器
request.interceptors.response.use(
  (response) => {
    const res = response.data
    if (res.code === 0 || res.code === 200) {
      return res.data !== undefined ? res.data : res
    }
    showToast(res.message || '请求失败')
    return Promise.reject(new Error(res.message || '请求失败'))
  },
  (error) => {
    if (error.response) {
      const { status, data } = error.response
      // 允许低优先级后台查询按请求关闭全局错误提示，避免非关键接口影响当前页面操作。
      const silentError = error.config?.silentError === true
      // 登录和发送验证码接口返回 401 时表示本次认证失败，不能触发全局登出跳转。
      const requestUrl = error.config?.url || ''
      const isAuthRequest = requestUrl.includes('/auth/login-by-sms') || requestUrl.includes('/auth/send-sms-code')
      
      switch (status) {
        case 401:
          if (!isAuthRequest) {
            const userStore = useUserStore()
            userStore.logout()
            router.push('/login')
            if (!silentError) showToast('登录已过期，请重新登录')
          }
          break
        case 403:
          if (!silentError) showToast('没有权限')
          break
        case 429:
          if (!silentError) showToast('操作太频繁，请稍后再试')
          break
        default:
          if (!silentError) showToast(data?.message || '服务器错误')
      }
    } else {
      if (error.config?.silentError !== true) showToast('网络连接失败，请检查网络')
    }
    return Promise.reject(error)
  }
)

export default request
