<template>
  <main class="driver-home-page">
    <section class="driver-hero">
      <div class="hero-topline">
        <div class="driver-profile-row">
          <img v-if="driverStore.driver.avatarUrl" :src="driverStore.driver.avatarUrl" alt="司机头像" />
          <span v-else class="avatar-fallback">{{ driverStore.displayName.slice(0, 1) || '--' }}</span>
          <div class="driver-copy">
            <p>{{ statusLabel }}</p>
            <h1>{{ driverStore.displayName || '--' }}</h1>
            <small>{{ driverStore.driver.phone ? String(driverStore.driver.phone).slice(0, 3) + '****' + String(driverStore.driver.phone).slice(-4) : '--' }}</small>
          </div>
        </div>
        <button type="button" class="hero-icon-button" aria-label="退出登录" @click="logoutDriver">
          <van-icon name="revoke" />
        </button>
      </div>

      <div class="hero-income">
        <span>今日收入</span>
        <strong>{{ formatPrice(todayIncome.totalIncomeCents) }}</strong>
        <small>{{ phaseLabel || '--' }}</small>
      </div>
    </section>

    <section class="hero-metric-row">
      <div class="hero-metric-card">
        <span>今日完单</span>
        <strong>{{ todayIncome.completedOrders ?? incomeSummary.todayCompletedOrders ?? '--' }}</strong>
      </div>
      <div class="hero-metric-card">
        <span>还可拒单</span>
        <strong>--</strong>
      </div>
      <div class="hero-metric-card">
        <span>服务指标</span>
        <strong>{{ serviceScore || '--' }}</strong>
      </div>
    </section>

    <section class="quick-entry-row">
      <button type="button"><span><van-icon name="fire-o" /></span><b>热力图</b></button>
      <button type="button"><span><van-icon name="service-o" /></span><b>听单检测</b></button>
      <button type="button"><span><van-icon name="gift-o" /></span><b>司机福利社</b></button>
      <button type="button"><span><van-icon name="friends-o" /></span><b>长期伙伴</b></button>
    </section>

    <section v-if="activeTab === 0" class="work-status-card">
      <div class="work-status-heading">
        <div>
          <span>接单状态</span>
          <strong>{{ workStatusText }}</strong>
        </div>
        <van-tag :type="workStatusTagType">{{ workStatusTagText }}</van-tag>
      </div>
      <p class="work-status-hint">{{ workStatusHint }}</p>
      <button type="button" class="work-primary-action" :class="workActionClass" :disabled="workActionDisabled" @click="toggleListening">
        <van-icon :name="workActionIcon" />
        <span>{{ workActionText }}</span>
      </button>
    </section>

    <section class="mini-stat-grid">
      <div><span>服务分</span><strong>{{ serviceScore || '--' }}</strong></div>
      <div><span>当前订单</span><strong>{{ driverStore.currentOrderId || '--' }}</strong></div>
      <div><span>车辆</span><strong>{{ driverStore.vehicle?.plateNo || driverStore.vehicleId || '--' }}</strong></div>
    </section>

    <section class="tab-panel-scroll">
      <component
        :is="activePanelComponent"
        v-bind="activePanelProps"
        @refresh-dashboard="loadDashboardData"
        @refresh-home-orders="loadHomeAvailableOrders({ silentError: true })"
        @load-orders="loadOrders"
        @update:order-mode="orderMode = $event"
        @update:order-status="orderStatus = $event"
        @order-detail="loadOrderDetail"
        @select-trajectory="selectTrajectory"
        @order-action="handleOrderAction"
        @open-finish="openFinish"
        @load-vehicle="loadVehicle"
        @submit-vehicle="submitVehicle"
        @submit-vehicle-update="submitVehicleUpdate"
        @remove-vehicle="removeVehicle"
        @load-certification="loadCertification"
        @read-cert-file="readCertFile"
        @remove-cert-image="removeCertImage"
        @submit-certification="submitCertification"
        @load-income="loadIncome"
        @open-withdraw="openWithdraw"
        @load-reviews="loadReviews"
        @update:trajectory-order-id="trajectoryOrderId = $event"
        @load-trajectory="loadTrajectory"
        @submit-profile="submitProfile"
      />
    </section>

    <van-tabbar v-model="activeTab" class="driver-tabbar" fixed safe-area-inset-bottom>
      <van-tabbar-item v-for="item in tabItems" :key="item.title" :icon="item.icon">{{ item.title }}</van-tabbar-item>
    </van-tabbar>

    <van-popup v-model:show="finishVisible" round position="bottom">
      <section class="finish-panel">
        <h2>结束行程</h2>
        <van-form @submit="submitFinishTrip">
          <van-field v-model.number="finishForm.actualDistanceM" type="number" label="实际里程(米)" />
          <van-field v-model.number="finishForm.actualDurationS" type="number" label="实际时长(秒)" />
          <button class="primary-action" type="submit">确认结束</button>
        </van-form>
      </section>
    </van-popup>

    <van-popup v-model:show="withdrawVisible" round position="bottom">
      <section class="withdraw-panel">
        <h2>申请提现</h2>
        <van-form @submit="submitWithdraw">
          <van-field v-model="withdrawForm.amount" type="number" label="提现金额" placeholder="请输入金额" />
          <van-field v-model="withdrawForm.payeeName" label="收款人" placeholder="请输入收款人姓名" />
          <van-field v-model="withdrawForm.payAccount" label="收款账号" placeholder="请输入收款账号" />
          <button class="primary-action" type="submit" :disabled="withdrawLoading">
            {{ withdrawLoading ? '提交中...' : '确认提现' }}
          </button>
        </van-form>
      </section>
    </van-popup>
  </main>
</template>

<script setup>
import { computed, onMounted, onUnmounted, reactive, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { closeToast, showConfirmDialog, showDialog, showLoadingToast, showToast } from 'vant'
import {
  acceptOrder,
  confirmArrive,
  createWithdraw,
  createVehicle,
  deleteVehicle,
  finishTrip,
  getCertification,
  getDriverAiScore,
  getDriverOrderDetail,
  getIncomeSummary,
  getOrderTrajectory,
  getTodayIncome,
  getVehicle,
  getWeekIncome,
  heartbeatDriver,
  listAvailableOrders,
  listDriverDispatches,
  listDriverOrders,
  listIncomeBills,
  listPassengerReviews,
  rejectOrder,
  reportDriverLocation,
  setDriverOffline,
  setDriverOnline,
  startTrip,
  updateVehicle,
  uploadCertification
} from '@/api/driver'
import { useDriverStore } from '@/stores/driver'
import { resolveCreatedVehicle } from '@/utils/vehicle'
import DriverWorkbenchPanel from '@/components/driver-home/DriverWorkbenchPanel.vue'
import DriverOrdersPanel from '@/components/driver-home/DriverOrdersPanel.vue'
import DriverAssetsPanel from '@/components/driver-home/DriverAssetsPanel.vue'
import DriverMinePanel from '@/components/driver-home/DriverMinePanel.vue'
import '@/styles/driver-home-panels.css'

const router = useRouter()
const driverStore = useDriverStore()

const tabItems = [
  { title: '工作台', icon: 'wap-home-o' },
  { title: '订单', icon: 'orders-o' },
  { title: '资产', icon: 'balance-o' },
  { title: '我的', icon: 'user-o' }
]

const activeTab = ref(0)
const tabPanelComponents = [
  DriverWorkbenchPanel,
  DriverOrdersPanel,
  DriverAssetsPanel,
  DriverMinePanel
]
const activePanelComponent = computed(() => tabPanelComponents[activeTab.value] || DriverWorkbenchPanel)
const activePanelProps = computed(() => {
  const commonFormatters = { formatPrice, formatDistance, formatTime, formatOrderStatus, formatDispatchStatus }
  if (activeTab.value === 0) {
    return {
      driverStore,
      homeAvailableOrders: homeAvailableOrders.value,
      homeAvailableLoading: homeAvailableLoading.value,
      formatPrice,
      formatDistance,
      formatOrderStatus
    }
  }
  if (activeTab.value === 1) {
    return {
      ...commonFormatters,
      orders: orders.value,
      orderMode: orderMode.value,
      orderStatus: orderStatus.value,
      orderModeOptions,
      orderStatusOptions,
      orderPage: orderPage.value,
      orderPageSize: orderPageSize.value,
      orderTotal: orderTotal.value,
      canAccept
    }
  }
  if (activeTab.value === 2) {
    return {
      driverStore,
      vehicleForm,
      certificationForm,
      certUploading,
      certItems,
      certSubmitting: certSubmitting.value,
      certStatusIcon: certStatusIcon.value,
      incomeSummary: incomeSummary.value,
      todayIncome: todayIncome.value,
      weekIncome: weekIncome.value,
      incomeBills: incomeBills.value,
      formatPrice,
      formatTime,
      formatVehicleStatus,
      formatCertificationStatus
    }
  }
  return {
    driverStore,
    profileForm,
    reviews: reviews.value,
    trajectoryOrderId: trajectoryOrderId.value,
    trajectoryError: trajectoryError.value,
    trajectoryPoints: trajectoryPoints.value,
    formatDriverStatus,
    formatTime
  }
})
const workLoading = ref(false)
const serviceScore = ref('--')
const orders = ref([])
const homeAvailableOrders = ref([])
const homeAvailableLoading = ref(false)
const orderMode = ref('orders')
const orderStatus = ref(0)
const orderPage = ref(1)
const orderPageSize = ref(8)
const orderTotal = ref(0)
const incomeSummary = ref({})
const todayIncome = ref({})
const weekIncome = ref({})
const incomeBills = ref([])
const reviews = ref([])
const trajectoryOrderId = ref('')
const trajectoryPoints = ref([])
const trajectoryError = ref('')
const finishVisible = ref(false)
const finishOrder = ref(null)
const withdrawVisible = ref(false)
const withdrawLoading = ref(false)
const withdrawForm = reactive({ amount: '', payeeName: '', payAccount: '' })

const orderModeOptions = [
  { text: '我的订单', value: 'orders' },
  { text: '派单记录', value: 'dispatches' },
  { text: '可接订单', value: 'available' }
]

const orderStatusOptions = [
  { text: '全部', value: 0 },
  { text: '待接单', value: 1 },
  { text: '已接单', value: 2 },
  { text: '行程中', value: 3 },
  { text: '已完成', value: 4 },
  { text: '已取消', value: 5 },
  { text: '已关闭', value: 6 }
]

const vehicleForm = reactive({
  plateNo: '',
  brand: '',
  model: '',
  color: '',
  vehicleType: 1,
  registrationDate: '',
  insuranceNo: '',
  insuranceExpireAt: ''
})
const certificationForm = reactive({ idCardFront: '', idCardBack: '', driverLicense: '', vehicleLicense: '' })
const certUploading = reactive({ idCardFront: false, idCardBack: false, driverLicense: false, vehicleLicense: false })
const certFileRefs = reactive({})
const certSubmitting = ref(false)
const certItems = [
  { key: 'idCardFront', title: '身份证正面', tip: '人像面，信息清晰' },
  { key: 'idCardBack', title: '身份证反面', tip: '国徽面，有效期清晰' },
  { key: 'driverLicense', title: '驾驶证', tip: '正副页，准驾车型清晰' },
  { key: 'vehicleLicense', title: '行驶证', tip: '正副页，车牌号清晰' }
]
const profileForm = reactive({ realName: '', avatarUrl: '', driverLicenseNo: '' })
const finishForm = reactive({ actualDistanceM: 0, actualDurationS: 0 })

let heartbeatTimer = null
let locationTimer = null
let tripTimer = null
let geoWatchId = null
let pushSocket = null
let reconnectTimer = null
let reconnectAttempts = 0
let lastLatitude = null
let lastLongitude = null

const statusLabel = computed(() => {
  if (driverStore.tripPhase === 'pickup') return '前往上车点'
  if (driverStore.tripPhase === 'trip' || driverStore.onlineStatus === 2) return '行程服务中'
  if (driverStore.onlineStatus === 1) return '在线听单'
  return '离线休息'
})

const phaseLabel = computed(() => {
  if (driverStore.tripPhase === 'pickup') return '接驾中'
  if (driverStore.tripPhase === 'trip') return '行程中'
  return driverStore.onlineStatus === 1 ? '听单中' : '未上线'
})

const workStatusText = computed(() => {
  if (driverStore.onlineStatus === 2 || driverStore.tripPhase === 'trip') return '行程中服务'
  return driverStore.onlineStatus === 1 ? '在线听单中' : '已下线'
})

const workStatusTagText = computed(() => {
  if (driverStore.onlineStatus === 2 || driverStore.tripPhase === 'trip') return '行程中'
  return driverStore.onlineStatus === 1 ? '已开启' : '未开启'
})

const workStatusTagType = computed(() => {
  if (driverStore.onlineStatus === 2 || driverStore.tripPhase === 'trip') return 'warning'
  return driverStore.onlineStatus === 1 ? 'success' : 'default'
})

const workStatusHint = computed(() => {
  if (driverStore.onlineStatus === 2 || driverStore.tripPhase === 'trip') return '当前有进行中的订单，请完成行程后再调整听单状态'
  return driverStore.onlineStatus === 1 ? '正在接收附近订单，可随时停止听单' : '上线后将接收附近订单推送'
})

const workActionText = computed(() => {
  if (driverStore.onlineStatus === 2 || driverStore.tripPhase === 'trip') return '行程中服务'
  return driverStore.onlineStatus === 1 ? '停止听单' : '开始听单'
})

const workActionIcon = computed(() => driverStore.onlineStatus === 1 ? 'pause-circle-o' : 'play-circle-o')
const workActionClass = computed(() => driverStore.onlineStatus === 1 ? 'go-offline' : 'go-online')
const workActionDisabled = computed(() => workLoading.value || driverStore.onlineStatus === 2 || driverStore.tripPhase === 'trip')

onMounted(async () => {
  await loadDashboardData()
  if (driverStore.onlineStatus > 0) startRealtimeWork()
})

onUnmounted(() => stopRealtimeWork())

watch(activeTab, () => loadCurrentTabData())

async function loadDashboardData() {
  const [profile, score] = await Promise.allSettled([
    safeApiCall(() => driverStore.refreshProfile({ silentError: true }), null, { silent: true }),
    safeApiCall(() => getDriverAiScore({ silentError: true }), null, { silent: true })
  ])

  if (score.status === 'fulfilled' && score.value) {
    serviceScore.value = Number(score.value.aiScore || score.value.score || 0).toFixed(1)
  }
  if (profile.status === 'fulfilled' && profile.value) syncForms()
  await loadCurrentTabData()
}

async function loadCurrentTabData() {
  const config = { silentError: true }
  if (activeTab.value === 0) return safeApiCall(() => loadHomeAvailableOrders(config), null, { silent: true })
  if (activeTab.value === 1) return safeApiCall(() => loadOrders(1, config), null, { silent: true })
  if (activeTab.value === 2) {
    return Promise.allSettled([
      safeApiCall(() => loadIncome(config), null, { silent: true }),
      safeApiCall(() => loadVehicle(config), null, { silent: true }),
      safeApiCall(() => loadCertification(config), null, { silent: true })
    ])
  }
  if (activeTab.value === 3) return safeApiCall(() => loadReviews(config), null, { silent: true })
  return null
}

function syncForms() {
  profileForm.realName = driverStore.driver.realName || ''
  profileForm.avatarUrl = driverStore.driver.avatarUrl || ''
  profileForm.driverLicenseNo = driverStore.driver.driverLicenseNo || ''
  if (driverStore.vehicle) {
    Object.assign(vehicleForm, {
      plateNo: driverStore.vehicle.plateNo || '',
      brand: driverStore.vehicle.brand || '',
      model: driverStore.vehicle.model || '',
      color: driverStore.vehicle.color || '',
      vehicleType: Number(driverStore.vehicle.vehicleType || 1),
      insuranceNo: driverStore.vehicle.insuranceNo || ''
    })
  }
}

async function setOnline() {
  await ensureWorkLocation()
  await setWorkStatus(() => setDriverOnline(workStatusPayload()), 1, '已上线，开始听单')
}

async function setOffline() {
  await ensureWorkLocation()
  await setWorkStatus(() => setDriverOffline(workStatusPayload()), 0, '已下线')
}

function toggleListening() {
  if (driverStore.onlineStatus === 1) return setOffline()
  if (driverStore.onlineStatus === 0) return setOnline()
}

async function setWorkStatus(request, status, message) {
  try {
    workLoading.value = true
    showLoadingToast({ message: '处理中...', forbidClick: true, duration: 0 })
    await request()
    closeToast()
    driverStore.setWorkState(status)
    if (status > 0) startRealtimeWork()
    else stopRealtimeWork()
    await loadHomeAvailableOrders({ silentError: true })
    showToast(message)
  } catch (error) {
    closeToast()
    showToast(apiErrorMessage(error, '切换工作状态失败'))
  } finally {
    workLoading.value = false
  }
}


// workLocationDefault 仅在浏览器无法获取定位时作为兜底，保证司机能上线听单（演示/同步环境可重新配置）。
const workLocationDefault = { longitude: 116.397128, latitude: 39.916527 }
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
  lastLongitude = Number(location.longitude)
  lastLatitude = Number(location.latitude)
  return { longitude: lastLongitude, latitude: lastLatitude }
}

function workStatusPayload() {
  return compact({ deviceId: deviceId(), longitude: lastLongitude, latitude: lastLatitude })
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

function startTripRealtime() {
  stopTripRealtime()
  tripTimer = window.setInterval(() => {
    if (driverStore.currentOrderId) {
      safeApiCall(() => getDriverOrderDetail(Number(driverStore.currentOrderId)), null, { silent: true })
    }
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
    if (payload.type === 'dispatch_order') {
      const order = {
        orderId: payload.orderId,
        orderNo: payload.orderNo,
        fromAddress: payload.fromAddress,
        toAddress: payload.toAddress,
        status: payload.status,
        estimatedPriceCents: payload.estimatedPriceCents,
        createdAt: payload.createdAt
      }
      driverStore.setCurrentOrder(order, Number(order.status) === 3 ? 'trip' : 'pickup')
      if (activeTab.value === 1) loadOrders(orderPage.value)
    } else if (payload.type === 'dispatch.new') {
      showToast('收到新的派单')
      void loadHomeAvailableOrders({ silentError: true })
    }
  } catch {
    // Push messages are best-effort; malformed messages should not block the H5 page.
  }
}

async function loadOrders(page = orderPage.value, config = {}) {
  orderPage.value = Number(page || 1)
  const payload = compact({
    page: orderPage.value,
    pageSize: orderPageSize.value,
    status: orderStatus.value
  })
  let data

  if (orderMode.value === 'available') data = await loadAvailableOrders(payload, config)
  else if (orderMode.value === 'dispatches') data = await loadDispatches(payload, config)
  else data = await listDriverOrders(payload, config)

  orders.value = Array.isArray(data?.list) ? data.list : []
  orderTotal.value = Number(data?.total || orders.value.length || 0)
  syncCurrentTripFromOrders(orders.value)
}

async function loadDispatches(payload, config = {}) {
  const data = await listDriverDispatches(payload, config)
  return {
    ...data,
    list: Array.isArray(data.list)
      ? data.list.map((item) => ({
        ...item,
        source: 'dispatch',
        orderId: item.orderId || item.order?.orderId,
        fromAddress: item.fromAddress || item.order?.fromAddress,
        toAddress: item.toAddress || item.order?.toAddress,
        estimatedPriceCents: item.estimatedPriceCents || item.order?.estimatedPriceCents,
        createdAt: item.createdAt || item.order?.createdAt
      }))
      : []
  }
}

async function loadAvailableOrders(payload, config = {}) {
  const data = await listAvailableOrders(payload, config)
  return {
    ...data,
    list: Array.isArray(data.list) ? data.list.map((item) => ({ ...item, source: 'available' })) : []
  }
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
      ? data.list.map((item) => ({ ...item, source: 'available' }))
      : []
  } finally {
    homeAvailableLoading.value = false
  }
}

function syncCurrentTripFromOrders(list) {
  const current = list.find((item) => String(item.orderId) === String(driverStore.currentOrderId)) || list.find((item) => [2, 3].includes(Number(item.status)))
  if (!current) return
  const phase = Number(current.status) === 3 ? 'trip' : Number(current.status) === 2 ? 'pickup' : 'idle'
  if (phase !== 'idle') driverStore.setCurrentOrder(current, phase)
}

function canAccept(order) {
  if (order.source === 'dispatch') return Number(order.dispatchStatus) === 1
  if (order.source === 'available') return Number(order.status || 1) === 1
  return Number(order.status) === 1
}


async function handleOrderAction(action, order) {
  const orderId = Number(order?.orderId || 0)
  if (!orderId) {
    showToast('订单ID无效')
    return
  }

  const config = {
    accept: { request: () => acceptOrder(orderId), phase: 'pickup', message: '接单成功' },
    reject: { request: async () => { const reason = await askRejectReason(); return rejectOrder(orderId, reason) }, phase: 'idle', message: '已拒单' },
    'confirm-arrive': { request: () => confirmArrive(orderId), phase: 'pickup', message: '已确认到达' },
    'start-trip': { request: () => startTrip(orderId), phase: 'trip', message: '行程已开始' }
  }[action]

  if (!config) return

  try {
    await config.request()
  } catch (error) {
    if (error?.message === 'cancelled') return
    showToast(apiErrorMessage(error, '订单操作失败'))
    return
  }

  driverStore.setCurrentOrder(config.phase === 'idle' ? null : order, config.phase)
  if (config.phase === 'trip') driverStore.setWorkState(2)
  if (config.phase === 'idle' && driverStore.onlineStatus === 2) driverStore.setWorkState(1)
  showToast(config.message)
  await loadOrders(orderPage.value)
  await loadHomeAvailableOrders({ silentError: true })
}

async function loadOrderDetail(orderId) {
  const res = await safeApiCall(() => getDriverOrderDetail(Number(orderId)), '订单详情加载失败')
  if (!res) return
  const order = res.order || res
  driverStore.setCurrentOrder(order, Number(order.status) === 3 ? 'trip' : Number(order.status) === 2 ? 'pickup' : driverStore.tripPhase)
  showDialog({
    title: order.orderNo || '订单 ' + order.orderId,
    message: (order.fromAddress || '--') + '\n到\n' + (order.toAddress || '--') + '\n' + formatPrice(order.estimatedPriceCents)
  })
}

async function askRejectReason() {
  const reason = window.prompt('请输入拒单原因', '司机当前不方便接单')
  if (reason === null) throw new Error('cancelled')
  return reason.trim() || '司机当前不方便接单'
}

function openFinish(order) {
  finishOrder.value = order
  finishVisible.value = true
}

async function submitFinishTrip() {
  const orderId = Number(finishOrder.value?.orderId || driverStore.currentOrderId || 0)
  if (!orderId) {
    showToast('订单ID无效')
    return
  }

  try {
    const res = await finishTrip({
      orderId,
      actualDistanceM: Number(finishForm.actualDistanceM || 0),
      actualDurationS: Number(finishForm.actualDurationS || 0)
    })
    finishVisible.value = false
    driverStore.setCurrentOrder(null, 'idle')
    if (driverStore.onlineStatus === 2) driverStore.setWorkState(1)
    showToast('行程已结束，应收' + formatPrice(res?.payableAmountCents))
    await loadOrders(orderPage.value)
  } catch (error) {
    showDialog({
      title: '结束行程失败',
      message: apiErrorMessage(error, '结束行程失败')
    }).catch(() => {})
  }
}

async function loadVehicle(config = {}) {
  if (!driverStore.vehicleId) return
  const res = await getVehicle(driverStore.vehicleId, config)
  driverStore.setVehicle(res.vehicle || res)
  syncForms()
}

function vehiclePayload() {
  return compact({
    plateNo: vehicleForm.plateNo,
    brand: vehicleForm.brand,
    model: vehicleForm.model,
    color: vehicleForm.color,
    vehicleType: Number(vehicleForm.vehicleType || 0),
    registrationDate: dateToUnixSeconds(vehicleForm.registrationDate),
    insuranceNo: vehicleForm.insuranceNo,
    insuranceExpireAt: dateToUnixSeconds(vehicleForm.insuranceExpireAt)
  })
}

async function submitVehicle() {
  const payload = vehiclePayload()
  const res = await safeApiCall(() => createVehicle(payload), '车辆提交失败')
  if (!res) return
  const vehicle = await resolveCreatedVehicle(payload, res, (id) => getVehicle(id, { silentError: true }))
  driverStore.setVehicle(vehicle)
  syncForms()
  showToast('车辆提交成功')
}

async function submitVehicleUpdate() {
  if (!driverStore.vehicleId) {
    showToast('请先查询或提交车辆')
    return
  }
  const res = await safeApiCall(() => updateVehicle({ ...vehiclePayload(), id: driverStore.vehicleId }), '车辆更新失败')
  if (!res) return
  showToast('车辆已更新')
  await loadVehicle()
}

async function removeVehicle() {
  if (!driverStore.vehicleId) {
    showToast('请先查询或提交车辆')
    return
  }
  try {
    await showConfirmDialog({ title: '删除车辆', message: '确认删除当前车辆？' })
  } catch {
    return
  }
  const res = await safeApiCall(() => deleteVehicle(driverStore.vehicleId), '车辆删除失败')
  if (!res) return
  driverStore.setVehicle(null)
  showToast('车辆已删除')
}

async function loadCertification(config = {}) {
  const res = await getCertification(config)
  driverStore.setCertification(res.found === false ? null : (res.certification || res))
  syncForms()
}

async function submitCertification() {
  const vehicleId = driverStore.vehicleId || driverStore.certification?.vehicleId
  if (!vehicleId) {
    showToast('请先在「车辆」页面绑定车辆')
    return
  }
  const hasImage = certItems.some(item => certificationForm[item.key])
  if (!hasImage) {
    showToast('请至少上传一项资质图片')
    return
  }
  certSubmitting.value = true
  const payload = compact({
    vehicleId: Number(vehicleId),
    idCardFront: certificationForm.idCardFront,
    idCardBack: certificationForm.idCardBack,
    driverLicense: certificationForm.driverLicense,
    vehicleLicense: certificationForm.vehicleLicense
  })
  try {
    const res = await safeApiCall(() => uploadCertification(payload), '资质上传失败')
    if (!res) return
    showToast('资质已提交，请等待审核')
    await loadCertification()
  } finally {
    certSubmitting.value = false
  }
}

const certStatusIcon = computed(() => {
  const status = driverStore.certification?.auditStatus
  if (status === 2) return 'checked'
  if (status === 3) return 'warning-o'
  return 'clock-o'
})

async function readCertFile(event, field) {
  const file = event.target.files?.[0]
  if (!file) return
  if (file.size > 5 * 1024 * 1024) {
    showToast('图片不能超过5MB')
    event.target.value = ''
    return
  }
  certUploading[field] = true
  try {
    certificationForm[field] = await fileToBase64(file)
  } finally {
    certUploading[field] = false
    event.target.value = ''
  }
}

function triggerCertUpload(field) {
  certFileRefs[field]?.click()
}

function removeCertImage(field) {
  certificationForm[field] = ''
}

async function loadIncome(config = {}) {
  const incomeRequests = [
    { label: '收入汇总', task: () => getIncomeSummary(config) },
    { label: '今日收入', task: () => getTodayIncome(config) },
    { label: '本周收入', task: () => getWeekIncome(config) },
    { label: '收入明细', task: () => listIncomeBills({ page: 1, pageSize: 20 }, config) }
  ]
  const [summary, today, week, bills] = await Promise.allSettled(incomeRequests.map((item) => item.task()))
  incomeSummary.value = summary.status === 'fulfilled' ? summary.value : {}
  todayIncome.value = today.status === 'fulfilled' ? today.value : {}
  weekIncome.value = week.status === 'fulfilled' ? week.value : {}
  incomeBills.value = bills.status === 'fulfilled' && Array.isArray(bills.value.list) ? bills.value.list : []

  const failures = [summary, today, week, bills]
    .map((result, index) => result.status === 'rejected' ? incomeRequests[index].label + ': ' + apiErrorMessage(result.reason, '请求失败') : '')
    .filter(Boolean)
  if (failures.length) showIncomeLoadFailure(failures)
}

function openWithdraw() {
  withdrawVisible.value = true
}

async function submitWithdraw() {
  const amount = Number(withdrawForm.amount)
  if (!Number.isFinite(amount) || amount <= 0 || !withdrawForm.payeeName.trim() || !withdrawForm.payAccount.trim()) {
    showToast('请填写完整的提现信息')
    return
  }
  withdrawLoading.value = true
  try {
    const res = await safeApiCall(() => createWithdraw({
      amount,
      payeeName: withdrawForm.payeeName.trim(),
      payAccount: withdrawForm.payAccount.trim()
    }), '提现申请失败')
    if (!res) return
    withdrawVisible.value = false
    Object.assign(withdrawForm, { amount: '', payeeName: '', payAccount: '' })
    showToast('提现申请已提交')
    await loadIncome({ silentError: true })
  } finally {
    withdrawLoading.value = false
  }
}

function showIncomeLoadFailure(failures) {
  showDialog({
    title: '收入数据加载失败',
    message: failures.join('\n')
  }).catch(() => {})
}

async function loadReviews(config = {}) {
  const res = await listPassengerReviews({ page: 1, pageSize: 20 }, config)
  reviews.value = Array.isArray(res.list) ? res.list : []
}

function selectTrajectory(orderId) {
  trajectoryOrderId.value = Number(orderId || 0)
  activeTab.value = 6
  loadTrajectory()
}

async function loadTrajectory() {
  const orderId = Number(trajectoryOrderId.value || driverStore.currentOrderId || 0)
  if (!orderId) {
    trajectoryError.value = '请输入订单ID'
    showToast('请输入订单ID')
    return
  }

  try {
    trajectoryError.value = ''
    const res = await getOrderTrajectory(orderId, { silentError: true })
    trajectoryPoints.value = Array.isArray(res.points) ? res.points : []
  } catch (error) {
    trajectoryPoints.value = []
    trajectoryError.value = apiErrorMessage(error, '轨迹查询失败')
  }
}

async function submitProfile() {
  const res = await safeApiCall(() => driverStore.saveProfile(compact(profileForm)), '资料保存失败')
  if (!res) return
  showToast('资料已保存')
}

async function refreshProfile() {
  const res = await safeApiCall(() => driverStore.refreshProfile(), '资料刷新失败')
  if (res) syncForms()
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

function logoutDriver() {
  stopRealtimeWork()
  driverStore.logout()
  router.replace('/login')
}

function compact(payload) {
  return Object.fromEntries(Object.entries(payload).filter(([, value]) => String(value ?? '').trim() !== ''))
}

function fileToBase64(file) {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => resolve(String(reader.result || ''))
    reader.onerror = () => reject(new Error('文件读取失败'))
    reader.readAsDataURL(file)
  })
}

function dateToUnixSeconds(value) {
  if (!value) return 0
  const time = new Date(value + 'T00:00:00').getTime()
  return Number.isFinite(time) ? Math.floor(time / 1000) : 0
}

function formatPrice(cents) {
  const value = Number(cents)
  return Number.isFinite(value) ? '¥' + (value / 100).toFixed(2) : '--'
}

function formatDistance(meters) {
  const value = Number(meters)
  if (!Number.isFinite(value) || value < 0) return '--'
  return (value / 1000).toFixed(value >= 10000 ? 0 : 1)
}

function formatTime(timestamp) {
  const value = Number(timestamp)
  if (!value) return '--'
  return new Date(value * 1000).toLocaleString('zh-CN', { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })
}

function formatOrderStatus(status) {
  return { 1: '待接单', 2: '已接单', 3: '行程中', 4: '已完成', 5: '已取消', 6: '已关闭' }[Number(status || 0)] || '--'
}

function formatDispatchStatus(status) {
  return { 1: '等待响应', 2: '司机接受', 3: '司机拒绝', 4: '派单超时', 5: '派单取消' }[Number(status || 0)] || '--'
}

function formatDriverStatus(status) {
  return {
    DRIVER_STATUS_PENDING: '待审核',
    DRIVER_STATUS_NORMAL: '正常',
    DRIVER_STATUS_FROZEN: '冻结',
    DRIVER_STATUS_CANCELLED: '注销'
  }[status] || status || '--'
}

function formatVehicleStatus(status) {
  return {
    VEHICLE_STATUS_PENDING: '待审核',
    VEHICLE_STATUS_NORMAL: '正常',
    VEHICLE_STATUS_DISABLED: '停用'
  }[status] || status || '--'
}

function formatCertificationStatus(status) {
  return { 1: '待审核', 2: '审核通过', 3: '审核驳回' }[Number(status || 0)] || '--'
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
</script>

<style scoped>
.driver-home-page { min-height: 100vh; padding: 0 12px 86px; background: #f6f7fb; color: #172033; }
.driver-hero { margin: 0 -12px; padding: 18px 18px 72px; border-radius: 0 0 28px 28px; background: linear-gradient(135deg, #7B61FF 0%, #5B5CFF 100%); color: #fff; }
.driver-profile-row { justify-content: flex-start; min-width: 0; }
.driver-profile-row img, .avatar-fallback { width: 52px; height: 52px; flex: 0 0 52px; border: 2px solid rgba(255,255,255,.72); border-radius: 50%; background: rgba(255,255,255,.22); object-fit: cover; }
.avatar-fallback { display: grid; place-items: center; color: #fff; font-size: 22px; font-weight: 800; }
.driver-copy { min-width: 0; }
.driver-copy p, .driver-copy h1, .driver-copy small, .hero-income span, .hero-income small { margin: 0; color: rgba(255,255,255,.78); }
.driver-copy h1 { margin: 4px 0; overflow: hidden; font-size: 20px; line-height: 1.2; text-overflow: ellipsis; white-space: nowrap; color: #fff; }
.hero-icon-button { width: 38px; height: 38px; flex: 0 0 38px; border: 0; border-radius: 50%; background: rgba(255,255,255,.18); color: #fff; font-size: 20px; }
.hero-income { margin-top: 26px; }
.hero-income strong { display: block; margin: 6px 0; font-size: 38px; line-height: 1; letter-spacing: 0; }
.hero-metric-row { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 10px; margin-top: -46px; }
.hero-metric-card { min-height: 84px; padding: 14px 10px; text-align: center; }
.quick-entry-row { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); gap: 8px; margin-top: 12px; padding: 16px 10px; }
.quick-entry-row button { min-width: 0; border: 0; background: transparent; color: #344054; font-size: 12px; font-weight: 700; }
.quick-entry-row span { display: grid; width: 46px; height: 46px; margin: 0 auto 8px; place-items: center; border-radius: 50%; background: #f0efff; color: #5B5CFF; font-size: 23px; }
.quick-entry-row b { display: block; overflow-wrap: anywhere; font-weight: 700; line-height: 1.2; }
.work-status-card { display: grid; gap: 12px; margin-top: 12px; padding: 18px; }
.work-status-heading { display: flex; align-items: flex-start; justify-content: space-between; gap: 12px; }
.work-status-heading > div { display: grid; gap: 4px; min-width: 0; }
.work-status-heading span, .work-status-hint { color: #7a8496; font-size: 12px; }
.work-status-heading strong { color: #172033; font-size: 26px; line-height: 1.15; }
.work-status-hint { margin: 0; line-height: 1.5; }
.work-primary-action { display: flex; align-items: center; justify-content: center; gap: 8px; width: 100%; min-height: 56px; border: 0; border-radius: 999px; font-size: 17px; font-weight: 800; }
.work-primary-action .van-icon { font-size: 21px; }
.go-online { border: 0; background: #ffb72c; color: #fff; box-shadow: 0 8px 16px rgba(255,183,44,.28); }
.mini-stat-grid { margin-top: 12px; }
.finish-panel, .withdraw-panel { padding: 18px 14px 22px; background: #f6f7fb; }
.finish-panel h2, .withdraw-panel h2 { margin: 0 0 14px; text-align: center; }
.driver-tabbar { left: 50%; right: auto; width: min(100vw, 430px); height: calc(60px + env(safe-area-inset-bottom)); border-top: 1px solid #e6eaf2; box-shadow: 0 -8px 24px rgba(15,23,42,.08); transform: translateX(-50%); --van-tabbar-item-active-color: #5B5CFF; --van-tabbar-item-text-color: #98a2b3; }
button:disabled { opacity: .48; }
@media (max-width: 360px) { .driver-home-page { padding-inline: 10px; } .driver-hero { margin-inline: -10px; } .hero-income strong { font-size: 32px; } .quick-entry-row { gap: 4px; } .work-status-card { display: grid; } .work-status-heading strong, .income-today-card strong { font-size: 30px; } .income-today-card { align-items: flex-start; } .withdraw-entry { min-width: 68px; padding-inline: 10px; } }
</style>
