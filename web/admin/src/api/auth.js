import request from './request'

// 登录：返回 {token, expires_in, admin}
export const login = (data) => request.post('/auth/login', data)
// 退出登录
export const logout = () => request.post('/auth/logout', {})
// 当前管理员信息：返回 {admin}
export const me = () => request.get('/auth/me')
// 角色菜单：返回 {items: [{name, path, icon, perm, children}]}
export const getMenus = () => request.get('/menus')
