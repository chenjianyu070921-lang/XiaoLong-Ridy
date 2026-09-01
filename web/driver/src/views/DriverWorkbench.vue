<template>
  <main class="driver-workbench-page">
    <header class="workbench-nav">
      <button type="button" aria-label="返回首页" @click="goBack">
        <van-icon name="arrow-left" />
      </button>
      <strong>工作台</strong>
      <button type="button" aria-label="刷新工作台" :disabled="homeAvailableLoading" @click="loadDashboardData">
        <van-icon name="replay" />
      </button>
    </header>

    <DriverWorkbenchPanel
      :driver-store="driverStore"
      :home-available-orders="homeAvailableOrders"
      :home-available-loading="homeAvailableLoading"
      :format-price="formatPrice"
      :format-distance="formatDistance"
      :format-order-status="formatOrderStatus"
      @refresh-dashboard="loadDashboardData"
      @refresh-home-orders="loadHomeAvailableOrders({ silentError: true })"
      @order-action="handleOrderAction"
    />
  </main>
</template>

<script setup>
import { onMounted, onUnmounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { showToast } from 'vant'
import {
  acceptOrder,
  getDriverOrderDetail,
  heartbeatDriver,
  listAvailableOrders,
  reportDriverLocation
} from '@/api/driver'
import { useDriverStore } from '@/stores/driver'
import DriverWorkbenchPanel from '@/components/driver-home/DriverWorkbenchPanel.vue'
import { normalizeBrowserLocationForAmap } from '@/utils/geo'
import '@/styles/driver-home-panels.css'

const router = useRouter()
const driverStore = useDriverStore()
const homeAvailableOrders = ref([])
const homeAvailableLoading = ref(false)

let heartbeatTimer = null
let locationTimer = null
let tripTimer = null
let geoWatchId = null
let pushSocket = null
let reconnectTimer = null
let reconnectAttempts = 0
let lastLatitude = null
let lastLongitude = null

const workLocationDefault = { longitude: 116.397128, latitude: 39.916527 }

onMounted(async () => {
  await loadDashboardData()
  if (driverStore.onlineStatus > 0) startRealtimeWork()
})

onUnmounted(() => {
  stopRealtimeWork()
})

function goBack() {
  router.replace('/home')
}

async function loadDashboardData() {
  await safeApiCall(() => driverStore.refreshProfile({ silentError: true }), null, { silent: true })
  await loadCurrentOrder({ silentError: true })
  await loadHomeAvailableOrders({ silentError: true })
}

async function loadCurrentOrder(config = {}) {
  const orderId = Number(driverStore.currentOrderId || 0)
  if (!orderId) return
  const data = await safeApiCall(() => getDriverOrderDetail(orderId), null, { silent: true, ...config })
  const order = data?.order || data
  if (!order) return
  const phase = Number(order.status) === 3 ? 'trip' : Number(order.status) === 2 ? 'pickup' : driverStore.tripPhase
  driverStore.setCurrentOrder(order, phase)
}

async function loadHomeAvailableOrders(config = {}) {
  if (driverStore.onlineStatus !== 1) {
    homeAvailableOrders.value = []
    return
  }
  homeAvailableLoading.value = true
  try {
    const data = await listAvailableOrders({ page: 1, pageSize: 3, status: 1 }, config)
    homeAvailableOrders.value = Array.isArray(data?.list)
      ? data.list.filter(isWaitAcceptOrder).map((item) => ({ ...item, source: 'available' }))
      : []
  } finally {
    homeAvailableLoading.value = false
  }
}

async function handleOrderAction(action, order) {
  if (action !== 'accept') return
  const orderId = Number(order?.orderId || 0)
  if (!orderId) {
    showToast('订单ID无效')
    return
  }

  try {
    await acceptOrder(orderId)
    driverStore.setCurrentOrder(order, 'pickup')
    showToast('接单成功')
    await loadHomeAvailableOrders({ silentError: true })
  } catch (error) {
    showToast(apiErrorMessage(error, '订单操作失败'))
    await loadHomeAvailableOrders({ silentError: true })
  }
}

function isWaitAcceptOrder(order) {
  return Number(order?.status || 0) === 1
}

function startRealtimeWork() {
  startHeartbeat()
  startLocationReporting()
  startTripRealtime()
  connectPushChannel()
}

function stopRealtimeWork() {
  stopHeartbeat()
  stopLocationReporting()
  stopTripRealtime()
  if (reconnectTimer) {
    window.clearTimeout(reconnectTimer)
    reconnectTimer = null
  }
  reconnectAttempts = 0
  if (pushSocket) {
    pushSocket.close()
    pushSocket = null
  }
}

function startHeartbeat() {
  stopHeartbeat()
  heartbeatTimer = window.setInterval(() => {
    safeApiCall(() => heartbeatDriver(workStatusPayload()), null, { silent: true })
  }, 30000)
}

function stopHeartbeat() {
  if (heartbeatTimer) window.clearInterval(heartbeatTimer)
  heartbeatTimer = null
}

function startLocationReporting() {
  stopLocationReporting()
  reportCurrentLocation()
  locationTimer = window.setInterval(reportCurrentLocation, 15000)
  if (navigator.geolocation && !geoWatchId) {
    geoWatchId = navigator.geolocation.watchPosition(
      (position) => {
        rememberWorkLocation({
          longitude: position.coords.longitude,
          latitude: position.coords.latitude
        })
      },
      () => {},
      { enableHighAccuracy: true, maximumAge: 10000, timeout: 8000 }
    )
  }
}

function stopLocationReporting() {
  if (locationTimer) window.clearInterval(locationTimer)
  locationTimer = null
  if (geoWatchId && navigator.geolocation) navigator.geolocation.clearWatch(geoWatchId)
  geoWatchId = null
}

async function reportCurrentLocation() {
  await ensureWorkLocation()
  const payload = workStatusPayload()
  if (!payload.longitude || !payload.latitude) return
  safeApiCall(() => reportDriverLocation(payload), null, { silent: true })
}

function ensureWorkLocation() {
  return new Promise((resolve) => {
    if (!navigator.geolocation) {
      resolve(rememberWorkLocation(workLocationDefault))
      return
    }
    if (lastLongitude != null && lastLatitude != null) {
      resolve({ longitude: lastLongitude, latitude: lastLatitude })
      return
    }
    navigator.geolocation.getCurrentPosition(
      (position) => {
        resolve(rememberWorkLocation({
          longitude: position.coords.longitude,
          latitude: position.coords.latitude
        }))
      },
      () => resolve(rememberWorkLocation(workLocationDefault)),
      { enableHighAccuracy: true, timeout: 5000, maximumAge: 10000 }
    )
  })
}

function rememberWorkLocation(location) {
  const normalized = normalizeBrowserLocationForAmap(location)
  lastLongitude = Number(normalized.longitude)
  lastLatitude = Number(normalized.latitude)
  return { longitude: lastLongitude, latitude: lastLatitude }
}

function workStatusPayload() {
  return compact({ deviceId: deviceId(), longitude: lastLongitude, latitude: lastLatitude })
}

function startTripRealtime() {
  stopTripRealtime()
  tripTimer = window.setInterval(() => {
    void loadCurrentOrder({ silentError: true })
  }, 20000)
}

function stopTripRealtime() {
  if (tripTimer) window.clearInterval(tripTimer)
  tripTimer = null
}

function connectPushChannel() {
  if (pushSocket || !driverStore.token) return
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const url = protocol + '//' + window.location.host + '/api/driver/v1/ws?token=' + encodeURIComponent(driverStore.token)
  pushSocket = new WebSocket(url)
  pushSocket.onopen = () => { reconnectAttempts = 0 }
  pushSocket.onmessage = (event) => handlePushMessage(event.data)
  pushSocket.onclose = () => {
    pushSocket = null
    scheduleReconnect()
  }
  pushSocket.onerror = () => {
    pushSocket?.close()
  }
}

function scheduleReconnect() {
  if (reconnectTimer || !driverStore.token || !driverStore.onlineStatus) return
  const delay = Math.min(1000 * Math.pow(2, reconnectAttempts), 30000)
  reconnectAttempts += 1
  reconnectTimer = window.setTimeout(() => {
    reconnectTimer = null
    connectPushChannel()
  }, delay)
}

function handlePushMessage(raw) {
  try {
    const payload = JSON.parse(raw)
    if (payload.type === 'dispatch_order' || payload.type === 'dispatch.new') {
      showToast('收到新的派单')
      void loadHomeAvailableOrders({ silentError: true })
    }
  } catch {
    // Push messages are best-effort; malformed messages should not block the H5 page.
  }
}

async function safeApiCall(task, fallbackMessage = '请求失败', options = {}) {
  try {
    return await task()
  } catch (error) {
    if (!options.silent) showToast(apiErrorMessage(error, fallbackMessage))
    return null
  }
}

function apiErrorMessage(error, fallbackMessage = '请求失败') {
  return error?.response?.data?.message || error?.message || fallbackMessage
}

function compact(payload) {
  return Object.fromEntries(Object.entries(payload).filter(([, value]) => String(value ?? '').trim() !== ''))
}

function deviceId() {
  const key = 'driverDeviceId'
  let value = localStorage.getItem(key)
  if (!value) {
    value = 'h5-' + Date.now() + '-' + Math.random().toString(16).slice(2)
    localStorage.setItem(key, value)
  }
  return value
}

function formatPrice(cents) {
  const value = Number(cents)
  return Number.isFinite(value) ? '￥' + (value / 100).toFixed(2) : '--'
}

function formatDistance(meters) {
  const value = Number(meters)
  if (!Number.isFinite(value) || value < 0) return '--'
  return (value / 1000).toFixed(value >= 10000 ? 0 : 1)
}

function formatOrderStatus(status) {
  return { 1: '待接单', 2: '已接单', 3: '行程中', 4: '已完成', 5: '已取消', 6: '已关闭' }[Number(status || 0)] || '--'
}
</script>

<style scoped>
.driver-workbench-page {
  min-height: 100vh;
  padding: 58px 12px 18px;
  background: #f6f7fb;
  color: #172033;
}

.workbench-nav {
  position: fixed;
  top: 0;
  left: 50%;
  z-index: 20;
  display: grid;
  grid-template-columns: 40px minmax(0, 1fr) 40px;
  align-items: center;
  width: min(100vw, 390px);
  height: 52px;
  padding: 8px 10px;
  border-bottom: 1px solid #e6eaf2;
  background: rgba(246, 247, 251, .96);
  backdrop-filter: blur(12px);
  transform: translateX(-50%);
}

.workbench-nav strong {
  text-align: center;
  color: #172033;
  font-size: 17px;
  line-height: 1.2;
}

.workbench-nav button {
  display: grid;
  width: 36px;
  height: 36px;
  place-items: center;
  border: 0;
  border-radius: 50%;
  background: #fff;
  color: #344054;
  font-size: 20px;
  box-shadow: 0 4px 14px rgba(15, 23, 42, .08);
}

.workbench-nav button:disabled {
  opacity: .48;
}
</style>
