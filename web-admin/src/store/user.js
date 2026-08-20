import { defineStore } from 'pinia'

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
  },
})
