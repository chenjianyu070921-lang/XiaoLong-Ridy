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

// estimatePrice 是价格预估的统一入口：上车点与目的地齐备后触发。
// 本次仅做 Location 到后端字段的映射与透传，不改动 /orders/estimate 的价格计算逻辑。
export function estimatePrice({ pickup, destination, carType, cityCode, estimatedDistanceM, estimatedDurationS, userCouponId }) {
  return estimateOrder({
    carType: Number(carType),
    fromAddress: pickup?.name || '',
    fromLongitude: Number(pickup?.longitude || 0),
    fromLatitude: Number(pickup?.latitude || 0),
    toAddress: destination?.name || '',
    toLongitude: Number(destination?.longitude || 0),
    toLatitude: Number(destination?.latitude || 0),
    cityCode: cityCode || '',
    estimatedDistanceM: Number(estimatedDistanceM || 0),
    estimatedDurationS: Number(estimatedDurationS || 0),
    userCouponId: Number(userCouponId || 0)
  })
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
