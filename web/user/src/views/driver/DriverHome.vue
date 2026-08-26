<template>
  <div class="driver-home-page">
    <header class="driver-header">
      <div class="identity">
        <img v-if="driverStore.driver.avatarUrl" :src="driverStore.driver.avatarUrl" alt="司机头像" />
        <span v-else>{{ driverStore.displayName.slice(0, 1) }}</span>
        <div>
          <p>{{ statusLabel }}</p>
          <h1>{{ driverStore.displayName }}</h1>
          <small>{{ driverStore.driver.phone || '--' }}</small>
        </div>
      </div>
      <button type="button" class="logout-btn" @click="logoutDriver">
        <van-icon name="revoke" />
      </button>
    </header>

    <section class="status-card">
      <div>
        <span>服务分</span>
        <strong>{{ serviceScore }}</strong>
      </div>
      <div>
        <span>当前订单</span>
        <strong>{{ driverStore.currentOrderId || '--' }}</strong>
      </div>
      <div>
        <span>车辆</span>
        <strong>{{ driverStore.vehicleId || '--' }}</strong>
      </div>
    </section>

    <section class="work-buttons">
      <button type="button" class="online" :disabled="workLoading || driverStore.onlineStatus === 1" @click="setOnline">
        上线接单
      </button>
      <button type="button" class="offline" :disabled="workLoading || driverStore.onlineStatus === 0" @click="setOffline">
        下线休息
      </button>
    </section>

    <main class="driver-content">
      <section v-show="activeTab === 0" class="panel workbench-panel">
          <div class="section-title">
            <h2>当前行程</h2>
            <van-tag type="primary">{{ phaseLabel }}</van-tag>
          </div>
          <div class="map-surface">
            <span class="road road-a"></span>
            <span class="road road-b"></span>
            <span class="pin current"></span>
            <span class="pin order"></span>
            <p>实时接单区域</p>
          </div>
          <div class="route-card">
            <p><b>上车点</b><span>{{ driverStore.currentOrder?.fromAddress || '--' }}</span></p>
            <p><b>目的地</b><span>{{ driverStore.currentOrder?.toAddress || '--' }}</span></p>
            <p><b>订单状态</b><span>{{ formatOrderStatus(driverStore.currentOrder?.status) }}</span></p>
          </div>
          <button type="button" class="ghost-action" @click="loadDashboardData">刷新工作台</button>
      </section>

      <section v-show="activeTab === 1" class="panel orders-panel">
          <div class="toolbar">
            <van-dropdown-menu>
              <van-dropdown-item v-model="orderMode" :options="orderModeOptions" @change="loadOrders(1)" />
              <van-dropdown-item v-model="orderStatus" :options="orderStatusOptions" @change="loadOrders(1)" />
            </van-dropdown-menu>
          </div>

          <div v-if="orders.length === 0" class="empty">暂无订单</div>
          <article v-for="order in orders" :key="`${order.source || 'order'}-${order.orderId}`" class="order-item">
            <div class="order-heading">
              <strong>{{ order.orderNo || `订单 ${order.orderId}` }}</strong>
              <van-tag>{{ order.source === 'dispatch' ? formatDispatchStatus(order.dispatchStatus) : formatOrderStatus(order.status) }}</van-tag>
            </div>
            <p class="route-line">{{ order.fromAddress || '--' }} → {{ order.toAddress || '--' }}</p>
            <div class="meta-row">
              <span>{{ formatPrice(order.estimatedPriceCents) }}</span>
              <span>{{ formatTime(order.createdAt) }}</span>
            </div>
            <div class="order-actions">
              <button type="button" @click="loadOrderDetail(order.orderId)">详情</button>
              <button type="button" @click="selectTrajectory(order.orderId)">轨迹</button>
              <button v-if="canAccept(order)" type="button" class="primary" @click="handleOrderAction('accept', order)">接单</button>
              <button v-if="canAccept(order)" type="button" @click="handleOrderAction('reject', order)">拒单</button>
              <button v-if="Number(order.status) === 2" type="button" @click="handleOrderAction('confirm-arrive', order)">确认到达</button>
              <button v-if="Number(order.status) === 2" type="button" @click="handleOrderAction('start-trip', order)">开始行程</button>
              <button v-if="Number(order.status) === 3" type="button" class="primary" @click="openFinish(order)">结束行程</button>
            </div>
          </article>

          <div class="pager">
            <button type="button" :disabled="orderPage <= 1" @click="loadOrders(orderPage - 1)">上一页</button>
            <span>{{ orderPage }} / {{ Math.max(1, Math.ceil(orderTotal / orderPageSize)) }}</span>
            <button type="button" :disabled="orderPage * orderPageSize >= orderTotal" @click="loadOrders(orderPage + 1)">下一页</button>
          </div>
      </section>

      <section v-show="activeTab === 2" class="panel vehicle-panel">
          <div class="section-title">
            <h2>车辆信息</h2>
            <button type="button" @click="loadVehicle">查询</button>
          </div>
          <div class="info-grid">
            <p><b>车辆ID</b><span>{{ driverStore.vehicle?.id || '--' }}</span></p>
            <p><b>车牌号</b><span>{{ driverStore.vehicle?.plateNo || '--' }}</span></p>
            <p><b>车型</b><span>{{ driverStore.vehicle?.brand || '--' }} {{ driverStore.vehicle?.model || '' }}</span></p>
            <p><b>状态</b><span>{{ formatVehicleStatus(driverStore.vehicle?.status) }}</span></p>
          </div>
          <van-form @submit="submitVehicle">
            <van-field v-model="vehicleForm.plateNo" label="车牌号" placeholder="粤B12345" />
            <van-field v-model="vehicleForm.brand" label="品牌" placeholder="BYD" />
            <van-field v-model="vehicleForm.model" label="型号" placeholder="Han" />
            <van-field v-model="vehicleForm.color" label="颜色" placeholder="黑色" />
            <van-field v-model.number="vehicleForm.vehicleType" type="number" label="车辆类型" placeholder="1" />
            <van-field v-model="vehicleForm.registrationDate" type="date" label="注册日期" />
            <van-field v-model="vehicleForm.insuranceNo" label="保险单号" placeholder="INS-001" />
            <van-field v-model="vehicleForm.insuranceExpireAt" type="date" label="保险到期" />
            <button class="primary-action" type="submit">提交车辆信息</button>
          </van-form>
          <div class="two-actions">
            <button type="button" @click="submitVehicleUpdate">更新车辆</button>
            <button type="button" class="danger" @click="removeVehicle">删除车辆</button>
          </div>
      </section>

      <section v-show="activeTab === 3" class="panel certification-panel">
          <div class="section-title">
            <h2>资质认证</h2>
            <button type="button" @click="loadCertification">查询</button>
          </div>
          <div class="info-grid">
            <p><b>认证ID</b><span>{{ driverStore.certification?.id || '--' }}</span></p>
            <p><b>车辆ID</b><span>{{ driverStore.certification?.vehicleId || driverStore.vehicleId || '--' }}</span></p>
            <p><b>审核状态</b><span>{{ formatCertificationStatus(driverStore.certification?.auditStatus) }}</span></p>
            <p><b>审核备注</b><span>{{ driverStore.certification?.auditRemark || '--' }}</span></p>
          </div>
          <van-form @submit="submitCertification">
            <van-field v-model.number="certificationForm.vehicleId" type="number" label="车辆ID" placeholder="请输入车辆ID" />
            <div class="file-grid">
              <label>身份证正面<input type="file" accept="image/*" @change="readCertFile($event, 'idCardFront')" /></label>
              <label>身份证反面<input type="file" accept="image/*" @change="readCertFile($event, 'idCardBack')" /></label>
              <label>驾驶证<input type="file" accept="image/*" @change="readCertFile($event, 'driverLicense')" /></label>
              <label>行驶证<input type="file" accept="image/*" @change="readCertFile($event, 'vehicleLicense')" /></label>
            </div>
            <button class="primary-action" type="submit">上传资质</button>
          </van-form>
      </section>

      <section v-show="activeTab === 4" class="panel wallet-panel">
          <div class="section-title">
            <h2>钱包收入</h2>
            <button type="button" @click="loadIncome">刷新</button>
          </div>
          <div class="income-card">
            <p>累计收入</p>
            <strong>{{ formatPrice(incomeSummary.totalIncomeCents) }}</strong>
            <span>已完成订单 {{ incomeSummary.completedOrders ?? '--' }}</span>
            <small>数据源：{{ incomeSummary.source || walletSummary.source || '--' }}</small>
          </div>
          <div v-if="incomeBills.length === 0" class="empty">暂无收入账单</div>
          <article v-for="bill in incomeBills" :key="bill.id || bill.orderId" class="compact-item">
            <strong>{{ bill.orderNo || `订单 ${bill.orderId}` }}</strong>
            <span>{{ formatPrice(bill.incomeCents) }} · {{ formatTime(bill.createdAt) }}</span>
          </article>
      </section>

      <section v-show="activeTab === 5" class="panel review-panel">
          <div class="section-title">
            <h2>乘客评价</h2>
            <button type="button" @click="loadReviews">刷新</button>
          </div>
          <div v-if="reviews.length === 0" class="empty">暂无乘客评价</div>
          <article v-for="review in reviews" :key="review.id || review.reviewId || review.orderId" class="compact-item">
            <strong>订单 {{ review.orderId || '--' }} · {{ review.rating || '--' }} 分</strong>
            <span>{{ review.comment || '未填写评价' }}</span>
            <small>{{ formatTime(review.createdAt) }}</small>
          </article>
      </section>

      <section v-show="activeTab === 6" class="panel trajectory-panel">
          <div class="section-title trajectory-title">
            <h2>轨迹查询</h2>
          </div>
          <div class="trajectory-form">
            <van-field v-model.number="trajectoryOrderId" type="number" label="订单ID" placeholder="输入订单ID" />
            <p v-if="trajectoryError" class="trajectory-error">{{ trajectoryError }}</p>
            <button class="primary-action trajectory-action" type="button" @click="loadTrajectory">查询轨迹</button>
          </div>
          <div v-if="trajectoryPoints.length === 0" class="empty trajectory-empty">暂无轨迹点</div>
          <article v-for="point in trajectoryPoints" :key="`${point.longitude}-${point.latitude}-${point.createdAt}`" class="compact-item">
            <strong>{{ point.longitude }}, {{ point.latitude }}</strong>
            <span>速度 {{ point.speedKmh || 0 }} km/h · 方向 {{ point.heading || 0 }}</span>
            <small>{{ formatTime(point.createdAt) }}</small>
          </article>
      </section>

      <section v-show="activeTab === 7" class="panel profile-panel">
          <div class="section-title">
            <h2>司机资料</h2>
            <button type="button" @click="driverStore.refreshProfile">刷新</button>
          </div>
          <div class="info-grid">
            <p><b>手机号</b><span>{{ driverStore.driver.phone || '--' }}</span></p>
            <p><b>驾驶证号</b><span>{{ driverStore.driver.driverLicenseNo || '--' }}</span></p>
            <p><b>账号状态</b><span>{{ formatDriverStatus(driverStore.driver.status) }}</span></p>
            <p><b>司机ID</b><span>{{ driverStore.driverId || '--' }}</span></p>
          </div>
          <van-form @submit="submitProfile">
            <van-field v-model="profileForm.realName" label="姓名" placeholder="司机姓名" />
            <van-field v-model="profileForm.avatarUrl" label="头像地址" placeholder="头像 URL" />
            <van-field v-model="profileForm.driverLicenseNo" label="驾驶证号" placeholder="驾驶证号" />
            <button class="primary-action" type="submit">保存资料</button>
          </van-form>
      </section>
    </main>

    <van-tabbar v-model="activeTab" class="driver-tabbar" fixed safe-area-inset-bottom>
      <van-tabbar-item v-for="item in tabItems" :key="item.title" :icon="item.icon">
        {{ item.title }}
      </van-tabbar-item>
    </van-tabbar>

    <van-popup v-model:show="finishVisible" round position="bottom">
      <section class="finish-panel">
        <h2>结束行程</h2>
        <van-form @submit="submitFinishTrip">
          <van-field v-model.number="finishForm.actualDistanceM" type="number" label="实际里程(米)" />
          <van-field v-model.number="finishForm.actualDurationS" type="number" label="实际时长(秒)" />
          <van-field v-model.number="finishForm.actualPriceCents" type="number" label="实际金额(分)" />
          <button class="primary-action" type="submit">确认结束</button>
        </van-form>
      </section>
    </van-popup>
  </div>
</template>

<script setup>
import { computed, onMounted, onUnmounted, reactive, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { closeToast, showDialog, showLoadingToast, showToast } from 'vant'
import {
  acceptOrder,
  confirmArrive,
  createVehicle,
  deleteVehicle,
  finishTrip,
  getCertification,
  getDriverAiScore,
  getDriverOrderDetail,
  getIncomeSummary,
  getOrderTrajectory,
  getVehicle,
  getWalletSummary,
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

const router = useRouter()
const driverStore = useDriverStore()

const tabItems = [
  { title: '工作台', icon: 'wap-home-o' },
  { title: '订单', icon: 'orders-o' },
  { title: '车辆', icon: 'logistics' },
  { title: '认证', icon: 'certificate' },
  { title: '钱包', icon: 'balance-o' },
  { title: '评价', icon: 'comment-o' },
  { title: '轨迹', icon: 'location-o' },
  { title: '我的', icon: 'user-o' }
]

const activeTab = ref(6)
const workLoading = ref(false)
const serviceScore = ref('--')
const orders = ref([])
const orderMode = ref('orders')
const orderStatus = ref(0)
const orderPage = ref(1)
const orderPageSize = ref(8)
const orderTotal = ref(0)
const incomeSummary = ref({})
const walletSummary = ref({})
const incomeBills = ref([])
const reviews = ref([])
const trajectoryOrderId = ref('')
const trajectoryPoints = ref([])
const trajectoryError = ref('')
const finishVisible = ref(false)
const finishOrder = ref(null)

const orderModeOptions = [
  { text: '我的订单', value: 'orders' },
  { text: '派单记录', value: 'dispatches' },
  { text: '抢单大厅', value: 'available' }
]

const orderStatusOptions = [
  { text: '全部', value: 0 },
  { text: '待接单', value: 1 },
  { text: '已接单', value: 2 },
  { text: '行程中', value: 3 },
  { text: '待支付', value: 4 },
  { text: '已完成', value: 5 },
  { text: '已取消', value: 6 }
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

const certificationForm = reactive({
  vehicleId: 0,
  idCardFront: '',
  idCardBack: '',
  driverLicense: '',
  vehicleLicense: ''
})

const profileForm = reactive({
  realName: '',
  avatarUrl: '',
  driverLicenseNo: ''
})

const finishForm = reactive({
  actualDistanceM: 0,
  actualDurationS: 0,
  actualPriceCents: 0
})

let heartbeatTimer = null
let locationTimer = null
let tripTimer = null
let geoWatchId = null
let pushSocket = null
let lastLatitude = null
let lastLongitude = null

const statusLabel = computed(() => {
  if (driverStore.tripPhase === 'pickup') return '正在接驾'
  if (driverStore.tripPhase === 'trip' || driverStore.onlineStatus === 2) return '行程进行中'
  if (driverStore.onlineStatus === 1) return '在线接单'
  return '下线休息'
})

const phaseLabel = computed(() => {
  if (driverStore.tripPhase === 'pickup') return '接驾中'
  if (driverStore.tripPhase === 'trip') return '行程中'
  return '空闲'
})

onMounted(async () => {
  await loadDashboardData()
  if (driverStore.onlineStatus > 0) {
    startHeartbeat()
    startLocationReporting()
    startTripRealtime()
    connectPushChannel()
  }
})

onUnmounted(() => {
  stopHeartbeat()
  stopLocationReporting()
  stopTripRealtime()
  if (pushSocket) pushSocket.close()
})

async function loadDashboardData() {
  try {
    const [profile, score] = await Promise.allSettled([
      safeApiCall(() => driverStore.refreshProfile({ silentError: true }), null, { silent: true }),
      safeApiCall(() => getDriverAiScore({ silentError: true }), null, { silent: true })
    ])
    if (score.status === 'fulfilled') {
      serviceScore.value = score.value ? Number(score.value.aiScore || 0).toFixed(1) : '--'
    }
    if (profile.status === 'fulfilled' && profile.value) {
      syncForms()
    }
    await loadCurrentTabData()
  } catch (error) {
    showToast(error.message || '工作台加载失败')
  }
}

watch(activeTab, () => {
  loadCurrentTabData()
})

async function loadCurrentTabData() {
  const config = { silentError: true }
  if (activeTab.value === 1) return safeApiCall(() => loadOrders(1, config), null, { silent: true })
  if (activeTab.value === 2) return safeApiCall(() => loadVehicle(config), null, { silent: true })
  if (activeTab.value === 3) return safeApiCall(() => loadCertification(config), null, { silent: true })
  if (activeTab.value === 4) return safeApiCall(() => loadIncome(config), null, { silent: true })
  if (activeTab.value === 5) return safeApiCall(() => loadReviews(config), null, { silent: true })
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
  certificationForm.vehicleId = driverStore.certification?.vehicleId || driverStore.vehicleId || 0
}

async function setOnline() {
  await setWorkStatus(() => setDriverOnline(workStatusPayload()), 1, '已上线接单')
}

async function setOffline() {
  await setWorkStatus(() => setDriverOffline(workStatusPayload()), 0, '已下线休息')
}

async function setWorkStatus(action, fallbackStatus, successText) {
  try {
    workLoading.value = true
    const res = await action()
    driverStore.setWorkState(res.onlineStatus ?? fallbackStatus)
    if (driverStore.onlineStatus === 0) {
      driverStore.setCurrentOrder(null, 'idle')
      stopHeartbeat()
      stopLocationReporting()
      stopTripRealtime()
    } else {
      startHeartbeat()
      startLocationReporting()
      startTripRealtime()
      connectPushChannel()
    }
    showToast(successText)
  } catch (error) {
    showToast(apiErrorMessage(error, '状态切换失败'))
  } finally {
    workLoading.value = false
  }
}

function workStatusPayload() {
  return {
    deviceId: getDeviceId(),
    longitude: Number(lastLongitude ?? 0),
    latitude: Number(lastLatitude ?? 0)
  }
}

function getDeviceId() {
  let deviceId = localStorage.getItem('driverDeviceId')
  if (!deviceId) {
    deviceId = `h5-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`
    localStorage.setItem('driverDeviceId', deviceId)
  }
  return deviceId
}

function startHeartbeat() {
  if (heartbeatTimer) return
  sendHeartbeat()
  heartbeatTimer = window.setInterval(sendHeartbeat, 15000)
}

function stopHeartbeat() {
  if (heartbeatTimer) {
    window.clearInterval(heartbeatTimer)
    heartbeatTimer = null
  }
}

async function sendHeartbeat() {
  if (!driverStore.isLoggedIn) {
    stopHeartbeat()
    return
  }
  const res = await safeApiCall(
    () => heartbeatDriver({ deviceId: getDeviceId(), longitude: lastLongitude || 0, latitude: lastLatitude || 0 }),
    null,
    { silent: true }
  )
  if (!res) return
  if (res?.kicked) {
    logoutDriver()
    showToast('账号已在其他设备登录')
  }
}

function startLocationReporting() {
  if (locationTimer) return
  if (navigator.geolocation && geoWatchId === null) {
    geoWatchId = navigator.geolocation.watchPosition(
      (position) => {
        lastLatitude = position.coords.latitude
        lastLongitude = position.coords.longitude
        sendLocation()
      },
      () => {},
      { enableHighAccuracy: true, maximumAge: 5000 }
    )
  }
  sendLocation()
  locationTimer = window.setInterval(sendLocation, 10000)
}

function stopLocationReporting() {
  if (locationTimer) {
    window.clearInterval(locationTimer)
    locationTimer = null
  }
  if (navigator.geolocation && geoWatchId !== null) {
    navigator.geolocation.clearWatch(geoWatchId)
    geoWatchId = null
  }
}

async function sendLocation() {
  if (!driverStore.isLoggedIn || lastLatitude === null || lastLongitude === null) return
  await safeApiCall(
    () => reportDriverLocation({
      deviceId: getDeviceId(),
      longitude: lastLongitude,
      latitude: lastLatitude,
      orderId: Number(driverStore.currentOrderId || 0)
    }),
    null,
    { silent: true }
  )
}

function startTripRealtime() {
  if (tripTimer) return
  refreshRealtimeTrip()
  tripTimer = window.setInterval(refreshRealtimeTrip, 5000)
}

function stopTripRealtime() {
  if (tripTimer) {
    window.clearInterval(tripTimer)
    tripTimer = null
  }
}

async function refreshRealtimeTrip() {
  if (!driverStore.isLoggedIn) return
  await safeApiCall(() => loadOrders(orderPage.value), null, { silent: true })
}

function connectPushChannel() {
  if (!driverStore.isLoggedIn || (pushSocket && pushSocket.readyState <= 1)) return
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const url = `${protocol}//${window.location.host}/api/driver/v1/ws`
  try {
    pushSocket = new WebSocket(url)
  } catch {
    return
  }
  pushSocket.onopen = () => pushSocket.send(JSON.stringify({ type: 'auth', token: driverStore.token }))
  pushSocket.onmessage = (event) => {
    try {
      const payload = JSON.parse(event.data || '{}')
      if (payload.type === 'dispatch' || payload.type === 'order') refreshRealtimeTrip()
    } catch {}
  }
  pushSocket.onclose = () => {
    if (driverStore.isLoggedIn) window.setTimeout(connectPushChannel, 5000)
  }
}

async function loadOrders(page = orderPage.value, config = {}) {
  const payload = { page, pageSize: orderPageSize.value, status: orderStatus.value }
  let data
  if (orderMode.value === 'dispatches') {
    data = normalizeDispatchResult(await listDriverDispatches(payload, config))
  } else if (orderMode.value === 'available') {
    data = await loadAvailableOrders(payload, config)
  } else {
    data = await listDriverOrders(payload, config)
  }
  orders.value = Array.isArray(data.list) ? data.list : []
  orderPage.value = Number(data.page || page || 1)
  orderPageSize.value = Number(data.pageSize || orderPageSize.value)
  orderTotal.value = Number(data.total || orders.value.length)
  syncCurrentTripFromOrders(orders.value)
}

function normalizeDispatchResult(data = {}) {
  return {
    ...data,
    list: Array.isArray(data.list)
      ? data.list.map((item) => ({
          ...(item.order || {}),
          source: 'dispatch',
          dispatchStatus: item.dispatch?.status || 0,
          dispatchId: item.dispatch?.id || 0,
          matchScore: item.dispatch?.matchScore || 0
        }))
      : []
  }
}

async function loadAvailableOrders(payload = { page: orderPage.value, pageSize: orderPageSize.value, status: orderStatus.value }, config = {}) {
  const data = await listAvailableOrders(payload, config)
  return {
    ...data,
    list: Array.isArray(data.list) ? data.list.map((item) => ({ ...item, source: 'available' })) : []
  }
}

function syncCurrentTripFromOrders(list) {
  const current = list.find((item) => String(item.orderId) === String(driverStore.currentOrderId)) ||
    list.find((item) => [2, 3].includes(Number(item.status)))
  if (!current) return
  const phase = Number(current.status) === 3 ? 'trip' : Number(current.status) === 2 ? 'pickup' : 'idle'
  if (phase !== 'idle') driverStore.setCurrentOrder(current, phase)
}

function canAccept(order) {
  return order.source === 'dispatch' || order.source === 'available' || Number(order.status) === 1
}

async function handleOrderAction(action, order) {
  const orderId = Number(order?.orderId || 0)
  if (!orderId) {
    showToast('订单ID无效')
    return
  }
  const config = {
    accept: { request: () => acceptOrder(orderId), phase: 'pickup', message: '接单成功' },
    reject: { request: () => rejectOrder(orderId), phase: 'idle', message: '已拒单' },
    'confirm-arrive': { request: () => confirmArrive(orderId), phase: 'pickup', message: '已确认到达' },
    'start-trip': { request: () => startTrip(orderId), phase: 'trip', message: '行程已开始' }
  }[action]
  if (!config) return
  await config.request()
  driverStore.setCurrentOrder(config.phase === 'idle' ? null : order, config.phase)
  if (config.phase === 'trip') driverStore.setWorkState(2)
  if (config.phase === 'idle' && driverStore.onlineStatus === 2) driverStore.setWorkState(1)
  showToast(config.message)
  await loadOrders(orderPage.value)
}

async function loadOrderDetail(orderId) {
  const res = await getDriverOrderDetail(Number(orderId))
  const order = res.order || res
  driverStore.setCurrentOrder(order, Number(order.status) === 3 ? 'trip' : Number(order.status) === 2 ? 'pickup' : driverStore.tripPhase)
  showDialog({
    title: order.orderNo || `订单 ${order.orderId}`,
    message: `${order.fromAddress || '--'}\n到\n${order.toAddress || '--'}\n${formatPrice(order.estimatedPriceCents)}`
  })
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
  await finishTrip({
    orderId,
    actualDistanceM: Number(finishForm.actualDistanceM || 0),
    actualDurationS: Number(finishForm.actualDurationS || 0),
    actualPriceCents: Number(finishForm.actualPriceCents || 0)
  })
  finishVisible.value = false
  driverStore.setCurrentOrder(null, 'idle')
  if (driverStore.onlineStatus === 2) driverStore.setWorkState(1)
  showToast('行程已结束')
  await loadOrders(orderPage.value)
}

async function loadVehicle(config = {}) {
  if (!driverStore.vehicleId) return
  const res = await getVehicle(driverStore.vehicleId, config)
  driverStore.setVehicle(res.vehicle || res)
  syncForms()
}

function vehiclePayload() {
  return {
    plateNo: vehicleForm.plateNo,
    brand: vehicleForm.brand,
    model: vehicleForm.model,
    color: vehicleForm.color,
    vehicleType: Number(vehicleForm.vehicleType || 0),
    registrationDate: dateToUnixSeconds(vehicleForm.registrationDate),
    insuranceNo: vehicleForm.insuranceNo,
    insuranceExpireAt: dateToUnixSeconds(vehicleForm.insuranceExpireAt)
  }
}

async function submitVehicle() {
  const res = await safeApiCall(() => createVehicle(vehiclePayload()), '车辆信息提交失败')
  if (!res) return
  driverStore.setVehicle(res.vehicle || res)
  showToast('车辆信息已提交')
}

async function submitVehicleUpdate() {
  if (!driverStore.vehicleId) {
    showToast('请先提交或查询车辆')
    return
  }
  const res = await safeApiCall(() => updateVehicle({ ...vehiclePayload(), id: driverStore.vehicleId }), '车辆信息更新失败')
  if (!res) return
  showToast('车辆信息已更新')
  await loadVehicle()
}

async function removeVehicle() {
  if (!driverStore.vehicleId) {
    showToast('请先提交或查询车辆')
    return
  }
  try {
    await showDialog({ title: '删除车辆', message: '确定删除当前车辆？', showCancelButton: true })
  } catch {
    return
  }
  const res = await safeApiCall(() => deleteVehicle(driverStore.vehicleId), '车辆删除失败')
  if (!res) return
  driverStore.setVehicle(null)
  showToast('车辆信息已删除')
}

async function loadCertification(config = {}) {
  const res = await getCertification(config)
  driverStore.setCertification(res.found === false ? null : (res.certification || res))
  syncForms()
}

async function submitCertification() {
  if (!certificationForm.vehicleId) {
    showToast('请输入车辆ID')
    return
  }
  const payload = compact({
    vehicleId: Number(certificationForm.vehicleId),
    idCardFront: certificationForm.idCardFront,
    idCardBack: certificationForm.idCardBack,
    driverLicense: certificationForm.driverLicense,
    vehicleLicense: certificationForm.vehicleLicense
  })
  if (Object.keys(payload).length <= 1) {
    showToast('请至少上传一张资质图片')
    return
  }
  const res = await safeApiCall(() => uploadCertification(payload), '资质上传失败')
  if (!res) return
  showToast('资质已上传，等待审核')
  await loadCertification()
}

async function readCertFile(event, field) {
  const file = event.target.files?.[0]
  if (!file) return
  certificationForm[field] = await fileToBase64(file)
}

async function loadIncome(config = {}) {
  const [summary, wallet, bills] = await Promise.allSettled([
    getIncomeSummary(config),
    getWalletSummary(config),
    listIncomeBills({ page: 1, pageSize: 20 }, config)
  ])
  incomeSummary.value = summary.status === 'fulfilled' ? summary.value : {}
  walletSummary.value = wallet.status === 'fulfilled' ? wallet.value : {}
  incomeBills.value = bills.status === 'fulfilled' && Array.isArray(bills.value.list) ? bills.value.list : []
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
    trajectoryError.value = error.response?.data?.message || error.message || 'Network Error'
  }
}

async function submitProfile() {
  const res = await safeApiCall(() => driverStore.saveProfile(compact(profileForm)), '司机资料保存失败')
  if (!res) return
  showToast('司机资料已保存')
}

async function safeApiCall(task, fallbackMessage = '请求失败', options = {}) {
  try {
    return await task()
  } catch (error) {
    if (!options.silent) {
      showToast(apiErrorMessage(error, fallbackMessage))
    }
    return null
  }
}

function apiErrorMessage(error, fallbackMessage = '请求失败') {
  return error?.response?.data?.message || error?.message || fallbackMessage
}

function logoutDriver() {
  stopHeartbeat()
  stopLocationReporting()
  stopTripRealtime()
  if (pushSocket) pushSocket.close()
  driverStore.logout()
  router.replace('/driver/login')
}

function compact(payload) {
  return Object.fromEntries(Object.entries(payload).filter(([, value]) => String(value ?? '').trim() !== ''))
}

function fileToBase64(file) {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => resolve(String(reader.result || ''))
    reader.onerror = () => reject(new Error('图片读取失败'))
    reader.readAsDataURL(file)
  })
}

function dateToUnixSeconds(value) {
  if (!value) return 0
  const time = new Date(`${value}T00:00:00`).getTime()
  return Number.isFinite(time) ? Math.floor(time / 1000) : 0
}

function formatPrice(cents) {
  const value = Number(cents)
  return Number.isFinite(value) ? `¥${(value / 100).toFixed(2)}` : '--'
}

function formatTime(timestamp) {
  const value = Number(timestamp)
  if (!value) return '--'
  return new Date(value * 1000).toLocaleString('zh-CN', { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })
}

function formatOrderStatus(status) {
  return {
    1: '待接单',
    2: '已接单',
    3: '行程中',
    4: '待支付',
    5: '已完成',
    6: '已取消'
  }[Number(status || 0)] || '--'
}

function formatDispatchStatus(status) {
  return {
    1: '待处理派单',
    2: '已接受派单',
    3: '已拒绝派单',
    4: '派单超时',
    5: '派单取消'
  }[Number(status || 0)] || '--'
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
    VEHICLE_STATUS_DISABLED: '禁用'
  }[status] || status || '--'
}

function formatCertificationStatus(status) {
  return {
    1: '待审核',
    2: '已通过',
    3: '已驳回'
  }[Number(status || 0)] || '--'
}
</script>

<style scoped>
.driver-home-page {
  min-height: 100vh;
  background: #f4f6fa;
  padding-bottom: calc(74px + env(safe-area-inset-bottom));
  color: #172033;
}

.driver-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 14px 20px;
  background: #6d4aff;
  color: #fff;
}

.identity {
  display: flex;
  min-width: 0;
  align-items: center;
  gap: 10px;
}

.identity img,
.identity > span {
  width: 48px;
  height: 48px;
  border: 2px solid rgba(255, 255, 255, 0.72);
  border-radius: 50%;
  flex: 0 0 auto;
  background: #fff;
  color: #6d4aff;
  display: grid;
  place-items: center;
  font-size: 22px;
  font-weight: 800;
  object-fit: cover;
}

.identity > div {
  min-width: 0;
}

.identity p,
.identity small {
  color: rgba(255, 255, 255, 0.74);
  display: block;
  font-size: 12px;
  line-height: 1.2;
}

.identity h1 {
  margin: 3px 0;
  font-size: 18px;
  line-height: 1.15;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.logout-btn {
  width: 40px;
  height: 40px;
  border: 0;
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.18);
  color: #fff;
}

.status-card {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 6px;
  margin: -10px 10px 8px;
  padding: 10px 8px;
  border-radius: 8px;
  background: #fff;
  box-shadow: 0 8px 22px rgba(77, 48, 160, 0.12);
}

.status-card div {
  display: grid;
  gap: 4px;
  min-width: 0;
  text-align: center;
}

.status-card span {
  color: #74798a;
  font-size: 11px;
}

.status-card strong {
  color: #1c2437;
  font-size: 15px;
  line-height: 1.2;
  overflow-wrap: anywhere;
}

.work-buttons {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 8px;
  padding: 0 10px 8px;
}

.work-buttons button,
.primary-action,
.ghost-action,
.two-actions button,
.section-title button,
.order-actions button,
.pager button {
  min-height: 44px;
  border: 0;
  border-radius: 8px;
  font-weight: 700;
  touch-action: manipulation;
}

.work-buttons .online,
.primary-action,
.order-actions .primary {
  background: #ffc83d;
  color: #3f2b00;
}

.work-buttons .offline,
.ghost-action,
.two-actions button,
.section-title button,
.order-actions button,
.pager button {
  background: #fff;
  color: #4b2bc5;
  border: 1px solid #d9d2ee;
}

.work-buttons button {
  min-height: 52px;
  font-size: 16px;
}

button:disabled {
  opacity: 0.58;
}

.driver-content {
  display: block;
}

.driver-home-page :deep(.van-dropdown-menu) {
  box-shadow: none;
}

.driver-home-page :deep(.van-field) {
  --van-field-label-width: 68px;
  border-radius: 8px;
  overflow: hidden;
}

.driver-home-page :deep(.van-cell) {
  padding-left: 12px;
  padding-right: 12px;
  line-height: 24px;
}

.driver-home-page :deep(.van-field__label) {
  font-size: 12px;
  color: #5f6678;
}

.driver-home-page :deep(.van-field__control) {
  font-size: 14px;
}

.panel {
  display: grid;
  gap: 10px;
  padding: 8px 10px 10px;
}

.section-title,
.order-heading,
.meta-row,
.pager {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
}

.section-title h2 {
  font-size: 16px;
  line-height: 1.2;
}

.trajectory-title {
  min-height: 28px;
}

.map-surface {
  position: relative;
  min-height: 140px;
  overflow: hidden;
  border: 1px solid #ded6ff;
  border-radius: 8px;
  background-color: #f1edff;
  background-image:
    linear-gradient(24deg, transparent 46%, rgba(255, 255, 255, 0.92) 47%, rgba(255, 255, 255, 0.92) 51%, transparent 52%),
    linear-gradient(112deg, transparent 42%, rgba(255, 255, 255, 0.72) 43%, rgba(255, 255, 255, 0.72) 47%, transparent 48%);
}

.map-surface p {
  position: absolute;
  top: 10px;
  right: 10px;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.9);
  color: #4b2bc5;
  padding: 5px 9px;
  font-size: 11px;
  font-weight: 700;
}

.road {
  position: absolute;
  height: 7px;
  border-radius: 999px;
  background: rgba(109, 74, 255, 0.18);
}

.road-a {
  top: 42%;
  left: -8%;
  width: 62%;
  transform: rotate(-18deg);
}

.road-b {
  top: 66%;
  right: -8%;
  width: 70%;
  transform: rotate(19deg);
}

.pin {
  position: absolute;
  width: 14px;
  height: 14px;
  border: 3px solid #fff;
  border-radius: 50%;
  box-shadow: 0 4px 12px rgba(77, 48, 160, 0.2);
}

.pin.current {
  top: 50%;
  left: 31%;
  background: #6d4aff;
}

.pin.order {
  top: 29%;
  right: 23%;
  background: #ffc83d;
}

.route-card,
.order-item,
.compact-item,
.info-grid,
.income-card {
  border: 1px solid #ebe7f5;
  border-radius: 8px;
  background: #fff;
  padding: 10px;
}

.route-card,
.info-grid {
  display: grid;
  gap: 8px;
}

.route-card p,
.info-grid p {
  display: flex;
  justify-content: space-between;
  gap: 12px;
}

.route-card span,
.info-grid span {
  text-align: right;
  overflow-wrap: anywhere;
  min-width: 0;
}

.toolbar {
  overflow: hidden;
  border-radius: 8px;
}

.order-item,
.compact-item {
  display: grid;
  gap: 8px;
}

.route-line,
.meta-row,
.compact-item span,
.compact-item small,
.income-card span,
.income-card small {
  color: #74798a;
  font-size: 12px;
}

.order-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.order-actions button {
  min-height: 36px;
  padding: 0 10px;
  font-size: 12px;
}

.pager {
  color: #74798a;
}

.pager button {
  padding: 0 12px;
}

.empty {
  min-height: 96px;
  display: grid;
  place-items: center;
  border: 1px dashed #d9d2ee;
  border-radius: 8px;
  color: #74798a;
  background: #fbfaff;
}

.primary-action,
.ghost-action {
  width: 100%;
}

.primary-action {
  min-height: 50px;
  font-size: 16px;
}

.trajectory-form {
  display: grid;
  gap: 8px;
  padding: 10px;
  border: 1px solid #ebe7f5;
  border-radius: 8px;
  background: #fff;
}

.trajectory-error {
  min-height: 20px;
  color: #dc2626;
  font-size: 13px;
  line-height: 20px;
  word-break: break-word;
}

.trajectory-action {
  margin-top: 2px;
}

.trajectory-empty {
  margin-top: 0;
}

.two-actions {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 8px;
}

.two-actions .danger {
  color: #dc2626;
  border-color: #fee2e2;
}

.file-grid {
  display: grid;
  gap: 8px;
  padding: 10px;
  border-radius: 8px;
  background: #fff;
}

.file-grid label {
  display: grid;
  gap: 8px;
  color: #4b5563;
  font-size: 12px;
}

.income-card {
  background: #6d4aff;
  color: #fff;
}

.income-card p,
.income-card small {
  color: rgba(255, 255, 255, 0.74);
}

.income-card strong {
  display: block;
  margin: 6px 0;
  font-size: 26px;
}

.finish-panel {
  padding: 12px 10px 16px;
}

.finish-panel h2 {
  margin: 0 0 12px;
  font-size: 16px;
  text-align: center;
}

.driver-tabbar {
  left: 50%;
  right: auto;
  width: min(100vw, 375px);
  height: calc(58px + env(safe-area-inset-bottom));
  transform: translateX(-50%);
  border-top: 1px solid #e6e8ef;
  box-shadow: 0 -8px 24px rgba(17, 24, 39, 0.08);
  --van-tabbar-item-active-color: #6d4aff;
}

.driver-tabbar :deep(.van-tabbar-item) {
  min-width: 0;
  padding: 5px 0 4px;
  color: #687083;
  font-size: 10px;
}

.driver-tabbar :deep(.van-tabbar-item__icon) {
  margin-bottom: 2px;
  font-size: 18px;
}
</style>
