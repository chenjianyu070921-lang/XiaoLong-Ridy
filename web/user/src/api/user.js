import request from './request'

// 获取用户信息
export function getProfile() {
  return request.get('/profile/me')
}

// 提交实名认证
export function submitRealName(realName, idCardNo) {
  return request.post('/profile/real-name', { realName, idCardNo })
}

// 更新个人资料（昵称 / 头像）。空字段表示不修改。
export function updateProfile({ nickname, avatarUrl } = {}) {
  return request.post('/profile/update', { nickname, avatarUrl })
}

// 设置或修改当前已登录乘客的密码。
export function setPassword({ currentPassword = '', newPassword } = {}) {
  return request.post('/profile/password', { currentPassword, newPassword })
}
