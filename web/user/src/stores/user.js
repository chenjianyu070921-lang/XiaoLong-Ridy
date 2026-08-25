import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { loginBySMS, refreshToken as refreshApi } from '@/api/auth'
import { getProfile } from '@/api/user'

export const useUserStore = defineStore('user', () => {
  const token = ref(localStorage.getItem('token') || '')
  const refreshTokenVal = ref(localStorage.getItem('refreshToken') || '')
  const userInfo = ref(JSON.parse(localStorage.getItem('userInfo') || '{}'))

  // 计算属性
  const isLoggedIn = computed(() => !!token.value)
  const phone = computed(() => userInfo.value.phone || '')
  const nickname = computed(() => userInfo.value.nickname || '乘客')
  const avatarUrl = computed(() => userInfo.value.avatarUrl || '/default-avatar.png')

  // 登录
  async function login(phoneNum, code) {
    try {
      const res = await loginBySMS(phoneNum, code)
      token.value = res.token
      refreshTokenVal.value = res.refreshToken
      // 统一兼容后端字段命名，并给新账号补充可识别的默认昵称和手机号。
      userInfo.value = normalizeUserInfo(res.user, phoneNum)
      // 仅将后端确认的首次新用户标记传给首页，避免普通用户重复看到新人礼包。
      if (res.isNewUser) {
        localStorage.setItem('passenger-new-user-pending-gift', '1')
        localStorage.removeItem('passenger-welcome-coupon-claimed')
        localStorage.removeItem('passenger-login-coupon-pending')
      } else {
        // 老用户每次成功登录后只触发一次登录福利广告。
        localStorage.setItem('passenger-login-coupon-pending', '1')
        localStorage.removeItem('passenger-new-user-pending-gift')
      }
      
      localStorage.setItem('token', res.token)
      localStorage.setItem('refreshToken', res.refreshToken)
      localStorage.setItem('userInfo', JSON.stringify(userInfo.value))
      
      return res
    } catch (error) {
      // 登录接口失败必须向上抛出，不能因为浏览器残留旧 token 而错误放行。
      console.error('登录接口失败:', error)
      throw error
    }
  }


  // 规范化登录用户资料，避免新用户因昵称为空或字段大小写不同而显示成空账号。
  function normalizeUserInfo(info = {}, fallbackPhone = '') {
    return {
      ...info,
      userId: info.userId ?? info.userID ?? 0,
      phone: info.phone || fallbackPhone,
      nickname: info.nickname || '乘客',
      avatarUrl: info.avatarUrl || info.avatarURL || '/default-avatar.png',
      realNameStatus: info.realNameStatus || 'unverified'
    }
  }

  // 刷新Token
  async function refreshTokenAction() {
    try {
      if (!refreshTokenVal.value) throw new Error('No refresh token')
      
      const res = await refreshApi(refreshTokenVal.value)
      token.value = res.token
      refreshTokenVal.value = res.refreshToken
      
      localStorage.setItem('token', res.token)
      localStorage.setItem('refreshToken', res.refreshToken)
      
      return res
    } catch (error) {
      logout()
      throw error
    }
  }

  // 获取用户信息
  async function fetchUserInfo() {
    try {
      const res = await getProfile()
      userInfo.value = res.user || {}
      localStorage.setItem('userInfo', JSON.stringify(userInfo.value))
      return res
    } catch (error) {
      throw error
    }
  }

  // 退出登录
  function logout() {
    token.value = ''
    refreshTokenVal.value = ''
    userInfo.value = {}
    
    localStorage.removeItem('token')
    localStorage.removeItem('refreshToken')
    localStorage.removeItem('userInfo')
  }

  return {
    token,
    refreshToken: refreshTokenVal,
    userInfo,
    isLoggedIn,
    phone,
    nickname,
    avatarUrl,
    login,
    refreshTokenAction,
    fetchUserInfo,
    logout
  }
})
