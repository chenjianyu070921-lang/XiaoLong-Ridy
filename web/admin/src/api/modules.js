import request, { download } from './request'

// 管理后台业务接口统一封装。页面只依赖这些方法，避免散落 URL 和参数拼装。
const list = (url, params) => request.get(url, { params })
export const usersApi = {
  list: (params) => list('/users', params),
  detail: (id) => request.get(`/users/${id}`),
  // 用户历史由管理后台网关聚合，页面不直接访问订单或用户服务。
  orders: (id, params) => list(`/users/${id}/orders`, params),
  coupons: (id, params) => list(`/users/${id}/coupons`, params),
  freeze: (id, data) => request.post(`/users/${id}/freeze`, data),
  unfreeze: (id, data) => request.post(`/users/${id}/unfreeze`, data),
}
export const driversApi = {
  // 司机基础资料接口等待 driversvc 模块提供，本轮不发起跨模块请求。
  unavailable: () => Promise.resolve({ list: [], total: 0 }),
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
  evidenceList: (id, params) => list(`/work-orders/${id}/evidence`, params),
}
export const statisticsApi = {
  overview: (params) => list('/statistics/overview', params),
  orders: (params) => list('/statistics/orders', params),
  coupons: (params) => list('/statistics/coupons', params),
  exports: (params) => list('/export-tasks', params),
  createExport: (data) => request.post('/export-tasks', data),
  exportDetail: (no) => request.get(`/export-tasks/${no}`),
  downloadExport: (no) => download(`/export-tasks/${no}/download`),
}
export const riskApi = {
  blacklist: (params) => list('/blacklist', params),
  add: (data) => request.post('/blacklist', data),
  release: (id, data) => request.post(`/blacklist/${id}/release`, data),
  hits: (params) => list('/risk/hit-records', params),
}
export const logsApi = { list: (params) => list('/operation-logs', params) }

// 管理员管理接口仅供超级管理员页面使用。
export const adminsApi = {
  list: (params) => list('/admins', params),
  create: (data) => request.post('/admins', data),
  update: (id, data) => request.put(`/admins/${id}`, data),
  setStatus: (id, data) => request.post(`/admins/${id}/status`, data),
  resetPassword: (id, data) => request.post(`/admins/${id}/reset-password`, data),
}
