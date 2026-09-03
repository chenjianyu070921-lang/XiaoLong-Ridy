export function compact(payload) {
  return Object.fromEntries(Object.entries(payload).filter(([, value]) => String(value ?? '').trim() !== ''))
}

export function dateToUnixSeconds(value) {
  if (!value) return 0
  if (typeof value === 'number') return value > 0 ? value : 0
  const time = new Date(value + 'T00:00:00').getTime()
  return Number.isFinite(time) ? Math.floor(time / 1000) : 0
}

export function unixSecondsToDateInput(value) {
  if (!value) return ''
  if (typeof value === 'string' && /^\d{4}-\d{2}-\d{2}$/.test(value)) return value
  const seconds = Number(value)
  if (!Number.isFinite(seconds) || seconds <= 0) return ''
  const date = new Date(seconds * 1000)
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

export function formatPrice(cents) {
  const value = Number(cents)
  return Number.isFinite(value) ? '¥' + (value / 100).toFixed(2) : '--'
}

export function formatDistance(meters) {
  const value = Number(meters)
  if (!Number.isFinite(value) || value < 0) return '--'
  return (value / 1000).toFixed(value >= 10000 ? 0 : 1)
}

export function formatDuration(seconds) {
  const value = Number(seconds)
  if (!Number.isFinite(value) || value <= 0) return '--'
  const minutes = Math.max(1, Math.round(value / 60))
  return minutes + '分钟'
}

export function formatTime(timestamp) {
  const value = Number(timestamp)
  if (!value) return '--'
  return new Date(value * 1000).toLocaleString('zh-CN', { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })
}

export function formatOrderStatus(status) {
  return { 1: '待接单', 2: '已接单', 3: '行程中', 4: '待支付', 5: '已完成', 6: '已取消', 7: '已退款' }[Number(status || 0)] || '--'
}

export function formatDispatchStatus(status) {
  return { 1: '等待响应', 2: '司机接受', 3: '司机拒绝', 4: '派单超时', 5: '派单取消' }[Number(status || 0)] || '--'
}

export function formatDriverStatus(status) {
  return {
    DRIVER_STATUS_PENDING: '待审核',
    DRIVER_STATUS_NORMAL: '正常',
    DRIVER_STATUS_FROZEN: '冻结',
    DRIVER_STATUS_CANCELLED: '注销'
  }[status] || status || '--'
}

export function formatVehicleStatus(status) {
  return {
    VEHICLE_STATUS_PENDING: '待审核',
    VEHICLE_STATUS_NORMAL: '正常',
    VEHICLE_STATUS_DISABLED: '停用'
  }[status] || status || '--'
}

export function formatCertificationStatus(status) {
  return { 1: '待审核', 2: '审核通过', 3: '审核驳回' }[Number(status || 0)] || '--'
}
