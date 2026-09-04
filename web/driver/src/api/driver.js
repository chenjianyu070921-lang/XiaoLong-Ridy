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

export function updateDriver(data, config = {}) {
  return driverRequest.post('/drivers/update', data, config)
}

export function getDriver(config = {}) {
  return driverRequest.get('/drivers/get', config)
}

export function getDriverAiScore(config = {}) {
  return driverRequest.get('/drivers/ai-score', config)
}

export function setDriverOnline(data = {}, config = {}) {
  return driverRequest.post('/drivers/online', data, config)
}

export function setDriverOffline(data = {}, config = {}) {
  return driverRequest.post('/drivers/offline', data, config)
}

export function heartbeatDriver(data, config = {}) {
  return driverRequest.post('/drivers/heartbeat', data, config)
}

export function reportDriverLocation(data, config = {}) {
  return driverRequest.post('/drivers/location/report', data, config)
}

export function createVehicle(data, config = {}) {
  return driverRequest.post('/vehicles', data, config)
}

export function getVehicle(id, config = {}) {
  return driverRequest.get('/vehicles/get', { ...config, params: { id, ...(config.params || {}) } })
}

export function updateVehicle(data, config = {}) {
  return driverRequest.post('/vehicles/update', data, config)
}

export function deleteVehicle(id, config = {}) {
  return driverRequest.post('/vehicles/delete', { id }, config)
}

export function createWithdraw(data, config = {}) {
  return driverRequest.post('/withdraws', data, config)
}

export function listWithdraws(data = {}, config = {}) {
  return driverRequest.post('/withdraws/list', data, config)
}

export function uploadCertification(data, config = {}) {
  return driverRequest.post('/drivers/certification/upload', data, config)
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

export function confirmArrive(orderId, config = {}) {
  return driverRequest.post('/orders/confirm-arrive', { orderId }, config)
}

export function startTrip(orderId, config = {}) {
  return driverRequest.post('/orders/start-trip', { orderId }, config)
}

export function finishTrip(data, config = {}) {
  return driverRequest.post('/orders/finish-trip', data, config)
}

export function getRealtimeFare(data = {}, config = {}) {
  return driverRequest.post('/orders/realtime-fare', data, config)
}

export function getOrderTrajectory(orderId, config = {}) {
  return driverRequest.post('/orders/trajectory', { orderId }, config)
}

export function listAvailableOrders(data = {}, config = {}) {
  return listGrabOrders(data, config)
}

export function listGrabOrders(data = {}, config = {}) {
  return driverRequest.post('/orders/grab-list', data, config)
}

export function getOrderHeatmap(data = {}, config = {}) {
  return driverRequest.post('/orders/heatmap', data, config)
}

export function getDriverOrderDetail(orderId, config = {}) {
  return driverRequest.post('/orders/detail', { orderId }, config)
}

export function listDriverOrders(data = {}, config = {}) {
  return driverRequest.post('/orders/list', data, config)
}

export function listDriverDispatches(data = {}, config = {}) {
  return driverRequest.post('/orders/dispatches', data, config)
}

export function listReceivedReviews(params = {}, config = {}) {
  return driverRequest.get('/reviews/received', {
    ...config,
    params: { page: params.page || 1, pageSize: params.pageSize || 20, ...(config.params || {}) }
  })
}

export function listGivenReviews(data = {}, config = {}) {
  return driverRequest.post('/reviews/given', data, config)
}

export function submitDriverReview(data, config = {}) {
  return driverRequest.post('/reviews/submit', data, config)
}

export default driverRequest
