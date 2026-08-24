import request from './request'

// 发送短信验证码
export function sendSMSCode(phone) {
  return request.post('/auth/send-sms-code', { phone })
}

// 短信验证码登录
export function loginBySMS(phone, code) {
  return request.post('/auth/login-by-sms', { phone, code })
}

// 刷新Token
export function refreshToken(refreshToken) {
  return request.post('/auth/refresh-token', { refreshToken })
}

// 退出登录
export function logout() {
  return request.post('/auth/logout')
}
