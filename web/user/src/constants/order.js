// 订单状态常量与展示规则统一维护在此处，必须与 ordersvc.proto 的 OrderStatus 枚举保持一致。
export const ORDER_STATUS = Object.freeze({
  SEARCHING: 'SEARCHING',
  ACCEPTED: 'ACCEPTED',
  IN_PROGRESS: 'IN_PROGRESS',
  PENDING_PAYMENT: 'PENDING_PAYMENT',
  COMPLETED: 'COMPLETED',
  CANCELLED: 'CANCELLED',
  REFUNDED: 'REFUNDED'
})

// normalizeOrderStatus 将后端数字状态转换为前端稳定使用的字符串状态。
export function normalizeOrderStatus(status) {
  const statusMap = {
    1: ORDER_STATUS.SEARCHING,
    2: ORDER_STATUS.ACCEPTED,
    3: ORDER_STATUS.IN_PROGRESS,
    4: ORDER_STATUS.PENDING_PAYMENT,
    5: ORDER_STATUS.COMPLETED,
    6: ORDER_STATUS.CANCELLED,
    7: ORDER_STATUS.REFUNDED
  }
  return typeof status === 'number' ? (statusMap[status] || String(status)) : status
}

// getOrderStatusText 返回订单状态面向乘客的展示文案。
export function getOrderStatusText(status) {
  const textMap = {
    [ORDER_STATUS.SEARCHING]: '等待接单',
    [ORDER_STATUS.ACCEPTED]: '司机已接单',
    [ORDER_STATUS.IN_PROGRESS]: '行程中',
    [ORDER_STATUS.PENDING_PAYMENT]: '待支付',
    [ORDER_STATUS.COMPLETED]: '已完成',
    [ORDER_STATUS.CANCELLED]: '已取消',
    [ORDER_STATUS.REFUNDED]: '已退款'
  }
  return textMap[status] || status
}

// driverFallbackNamePattern 匹配后端在 driversvc 不可用时生成的兜底称呼（如"司机37师傅"）。
// 这类文案本身已是乘客可读的称呼，前端不应再按姓氏截取，否则会把数字当成姓氏。
const driverFallbackNamePattern = /^司机\d+师傅$/

// formatDriverDisplayName 把司机真实姓名转换为乘客侧称呼：姓氏 + 师傅。
// 这是乘客端统一的脱敏规则——任何页面都不得直接展示司机完整实名。
// 真实姓名缺失时降级为"司机 #ID"；连司机 ID 都没有时返回加载占位。
export function formatDriverDisplayName(realName, driverID) {
  const raw = String(realName || '').trim()
  if (!raw) {
    const id = Number(driverID || 0)
    return id > 0 ? `司机 #${id}` : '司机信息加载中'
  }
  if (driverFallbackNamePattern.test(raw)) return raw
  const surname = raw.slice(0, 1)
  return surname ? `${surname}师傅` : '司机师傅'
}

// formatPlateNumber 规范化车牌号展示；后端未返回时统一显示占位文案，禁止留空或伪造。
export function formatPlateNumber(plateNumber) {
  const value = String(plateNumber || '').trim()
  return value || '车牌信息加载中'
}
