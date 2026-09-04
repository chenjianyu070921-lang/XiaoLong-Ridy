import request from './request'

// 发送短信验证码
export function sendSMSCode(phone) {
  return request.post('/auth/send-sms-code', { phone })
}

// 短信验证码登录
export function loginBySMS(phone, code) {
  return request.post('/auth/login-by-sms', { phone, code })
}

// 手机号密码登录；服务端校验密码后返回与短信登录一致的令牌结构。
export function loginByPassword(phone, password) {
  return request.post('/auth/login-by-password', { phone, password })
}

// 刷新Token
export function refreshToken(refreshToken) {
  return request.post('/auth/refresh-token', { refreshToken })
}

// 退出登录
export function logout() {
  return request.post('/auth/logout')
}
