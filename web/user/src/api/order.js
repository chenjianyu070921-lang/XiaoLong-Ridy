import request from './request'

// 查询行程预估价格，不创建订单。
export function getMyCoupons(status = 1) {
  return request.post('/coupons/my', { status })
}

// 领取新用户首次登录时展示的新人优惠券礼包。
export function claimWelcomeGift() {
  return request.post('/coupons/welcome-gift')
}

export function estimateOrder(data) {
  return request.post('/orders/estimate', data)
}

// 创建订单
export function createOrder(data) {
  return request.post('/orders/create', data)
}

// 获取订单列表
export function getOrders(params) {
  return request.post('/orders/list', params)
}

// 获取订单详情
export function getOrderDetail(orderId) {
  return request.post('/orders/detail', { orderId })
}

// 轮询订单状态
export function pollOrderStatus(orderId, knownStatus = 0) {
  return request.post('/orders/status', { orderId, knownStatus })
}

// 查询乘客当前订单的实时追踪快照，包含司机位置、行程进度和路线信息。
export function getOrderTracking(orderId) {
  return request.post('/orders/tracking', { orderId })
}

// 取消订单
export function cancelOrder(orderId, reason = '') {
  return request.post('/orders/cancel', { orderId, reason })
}

// 发起支付；后端使用 paysvc 的数字枚举：微信 1、支付宝 2、余额 3。
export function payOrder(orderId, payMethod = 'alipay') {
  const channelMap = { wechat: 1, alipay: 2, balance: 3 }
  return request.post('/orders/pay', { orderId, channel: channelMap[payMethod] || 2 })
}

// 查询支付状态
export function getPaymentStatus(orderId) {
  return request.post('/orders/payment-status', { orderId })
}

// 提交已完成订单的乘客评价，后端会校验订单归属、完成状态和重复提交。
export function submitReview(data) {
  return request.post('/reviews/submit', data)
}

// 查询派单状态
export function getDispatchStatus(orderId) {
  return request.post('/orders/dispatch-status', { orderId })
}
