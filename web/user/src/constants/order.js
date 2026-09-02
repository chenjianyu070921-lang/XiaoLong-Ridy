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
