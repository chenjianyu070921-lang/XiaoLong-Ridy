import driverRequest from '@/api/request'

export function sendDriverSMSCode(phone, config = {}) {
  return driverRequest.post('/auth/send-sms-code', { phone }, config)
}

export function loginDriverByPassword(phone, password, config = {}) {
  return driverRequest.post('/auth/login-by-password', { phone, password }, config)
}

export function loginDriverBySMS(phone, code, config = {}) {
  return driverRequest.post('/auth/login-by-sms', { phone, code }, config)
}

export function registerDriver(data, config = {}) {
  return driverRequest.post('/drivers/register', data, config)
}

export function updateDriver(data) {
  return driverRequest.post('/drivers/update', data)
}

export function uploadDriverAvatar(avatar) {
  return driverRequest.post('/drivers/avatar/upload', { avatar })
}

export function getDriver(config = {}) {
  return driverRequest.get('/drivers/get', config)
}

export function listNearbyDrivers(data = {}, config = {}) {
  return driverRequest.post('/drivers/nearby', data, config)
}

export function getDriverAiScore(config = {}) {
  return driverRequest.get('/drivers/ai-score', config)
}

export function setDriverOnline(data = {}) {
  return driverRequest.post('/drivers/online', data)
}

export function setDriverOffline(data = {}) {
  return driverRequest.post('/drivers/offline', data)
}

export function heartbeatDriver(data) {
  return driverRequest.post('/drivers/heartbeat', data)
}

export function reportDriverLocation(data) {
  return driverRequest.post('/drivers/location/report', data)
}

export function createVehicle(data) {
  return driverRequest.post('/vehicles', data)
}

export function getVehicle(id, config = {}) {
  return driverRequest.get('/vehicles/get', { ...config, params: { id, ...(config.params || {}) } })
}

export function updateVehicle(data) {
  return driverRequest.post('/vehicles/update', data)
}

export function deleteVehicle(id) {
  return driverRequest.post('/vehicles/delete', { id })
}

export function createWithdraw(data) {
  return driverRequest.post('/withdraws', data)
}

export function uploadCertification(data) {
  return driverRequest.post('/drivers/certification/upload', data)
}

export function getCertification(config = {}) {
  return driverRequest.get('/drivers/certification', config)
}

export function getIncomeSummary(config = {}) {
  return driverRequest.get('/income/summary', config)
}

export function getTodayIncome(config = {}) {
  return driverRequest.get('/income/today', config)
}

export function getWeekIncome(config = {}) {
  return driverRequest.get('/income/week', config)
}

export function listIncomeBills(data = {}, config = {}) {
  return driverRequest.post('/income/bills', data, config)
}

export function acceptOrder(orderId, config = {}) {
  return driverRequest.post('/orders/accept', { orderId }, config)
}

export function rejectOrder(orderId, reason = '司机主动拒单', config = {}) {
  return driverRequest.post('/orders/reject', { orderId, reason }, config)
}

export function confirmArrive(orderId) {
  return driverRequest.post('/orders/confirm-arrive', { orderId })
}

export function startTrip(orderId) {
  return driverRequest.post('/orders/start-trip', { orderId })
}

export function finishTrip(data) {
  return driverRequest.post('/orders/finish-trip', data)
}

export function listAvailableOrders(data = {}, config = {}) {
  return driverRequest.post('/orders/available', data, config)
}

export function getOrderHeatmap(data = {}, config = {}) {
  return driverRequest.post('/orders/heatmap', data, config)
}

export function getDriverOrderDetail(orderId) {
  return driverRequest.post('/orders/detail', { orderId })
}

export function getOrderTrajectory(orderId, config = {}) {
  return driverRequest.post('/orders/trajectory', { orderId }, config)
}

export function listDriverOrders(data = {}, config = {}) {
  return driverRequest.post('/orders/list', data, config)
}

export function listDriverDispatches(data = {}, config = {}) {
  return driverRequest.post('/orders/dispatches', data, config)
}

export function listPassengerReviews(data = {}, config = {}) {
  return driverRequest.post('/reviews/list', data, config)
}

export default driverRequest
