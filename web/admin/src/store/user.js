import { defineStore } from 'pinia'
import { me } from '../api/auth'

// 管理员登录态：token + 管理员信息 + 菜单。
// 持久化到 localStorage，刷新页面不丢登录态。
export const useUserStore = defineStore('user', {
  state: () => ({
    token: localStorage.getItem('admin_token') || '',
    admin: JSON.parse(localStorage.getItem('admin_info') || 'null'),
    menus: [],
  }),
  getters: {
    isLoggedIn: (state) => !!state.token,
  },
  actions: {
    setLogin(token, admin) {
      this.token = token
      this.admin = admin
      localStorage.setItem('admin_token', token)
      localStorage.setItem('admin_info', JSON.stringify(admin || null))
    },
    setMenus(menus) {
      this.menus = menus || []
    },
    logout() {
      this.token = ''
      this.admin = null
      this.menus = []
      localStorage.removeItem('admin_token')
      localStorage.removeItem('admin_info')
    },
    // refreshProfile 以服务端会话为准刷新管理员资料，避免仅信任可能过期的 localStorage 缓存。
    async refreshProfile() {
      if (!this.token) return null
      const data = await me()
      this.admin = data?.admin || null
      localStorage.setItem('admin_info', JSON.stringify(this.admin))
      return this.admin
    },
  },
})
