import request from './request'

// 管理后台业务接口统一封装。页面只依赖这些方法，避免散落 URL 和参数拼装。
const list = (url, params) => request.get(url, { params })
export const usersApi = {
  list: (params) => list('/users', params),
  detail: (id) => request.get(`/users/${id}`),
  freeze: (id, data) => request.post(`/users/${id}/freeze`, data),
  unfreeze: (id, data) => request.post(`/users/${id}/unfreeze`, data),
}
export const driversApi = {
  list: (params) => list('/driver-certifications', params),
  detail: (id) => request.get(`/driver-certifications/${id}`),
  approve: (id, data) => request.post(`/driver-certifications/${id}/approve`, data),
  reject: (id, data) => request.post(`/driver-certifications/${id}/reject`, data),
}
export const ordersApi = {
  list: (params) => list('/orders', params),
  abnormal: (params) => list('/orders/abnormal', params),
  detail: (id) => request.get(`/orders/${id}`),
  cancel: (id, data) => request.post(`/orders/${id}/cancel`, data),
}
export const marketingApi = {
  coupons: (params) => list('/coupons', params),
  createCoupon: (data) => request.post('/coupons', data),
  updateCoupon: (id, data) => request.put(`/coupons/${id}`, data),
  disableCoupon: (id) => request.post(`/coupons/${id}/disable`),
  issueCoupon: (id, data) => request.post(`/coupons/${id}/issue`, data),
  issueTasks: (params) => list('/coupon-issue-tasks', params),
  priceRules: (params) => list('/price-rules', params),
  priceRule: (id) => request.get(`/price-rules/${id}`),
  createPriceRule: (data) => request.post('/price-rules', data),
  updatePriceRule: (id, data) => request.put(`/price-rules/${id}`, data),
  enablePriceRule: (id) => request.post(`/price-rules/${id}/enable`),
  disablePriceRule: (id) => request.post(`/price-rules/${id}/disable`),
  activities: (params) => list('/promotion-activities', params),
  createActivity: (data) => request.post('/promotion-activities', data),
  updateActivity: (id, data) => request.put(`/promotion-activities/${id}`, data),
  publishActivity: (id, data) => request.post(`/promotion-activities/${id}/publish`, data),
  rollbackActivity: (id, data) => request.post(`/promotion-activities/${id}/rollback`, data),
}
export const workOrdersApi = {
  list: (params) => list('/work-orders', params),
  detail: (id) => request.get(`/work-orders/${id}`),
  create: (data) => request.post('/work-orders', data),
  action: (id, data) => request.post(`/work-orders/${id}/actions`, data),
  evidence: (id, data) => request.post(`/work-orders/${id}/evidence`, data),
}
export const statisticsApi = {
  overview: (params) => list('/statistics/overview', params),
  orders: (params) => list('/statistics/orders', params),
  coupons: (params) => list('/statistics/coupons', params),
  exports: (params) => list('/export-tasks', params),
  createExport: (data) => request.post('/export-tasks', data),
  exportDetail: (no) => request.get(`/export-tasks/${no}`),
}
export const riskApi = {
  blacklist: (params) => list('/blacklist', params),
  add: (data) => request.post('/blacklist', data),
  release: (id, data) => request.post(`/blacklist/${id}/release`, data),
  hits: (params) => list('/risk/hit-records', params),
}
export const logsApi = { list: (params) => list('/operation-logs', params) }
