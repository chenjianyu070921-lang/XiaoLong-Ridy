<template>
  <main class="driver-home-page">
    <section class="h5-status-hero">
      <div class="driver-profile-row">
        <img v-if="driverStore.driver.avatarUrl" :src="driverStore.driver.avatarUrl" alt="司机头像" />
        <span v-else>{{ driverStore.displayName.slice(0, 1) }}</span>
        <div>
          <p>{{ statusLabel }}</p>
          <h1>{{ driverStore.displayName }}</h1>
          <small>{{ driverStore.driver.phone || '--' }}</small>
        </div>
        <button type="button" class="icon-button" @click="logoutDriver">
          <van-icon name="revoke" />
        </button>
      </div>

      <div class="online-control-card">
        <div>
          <small>当前模式</small>
          <strong>{{ phaseLabel }}</strong>
        </div>
        <div class="quick-action-grid">
          <button type="button" class="go-online" :disabled="workLoading || driverStore.onlineStatus === 1" @click="setOnline">
            上线
          </button>
          <button type="button" class="go-offline" :disabled="workLoading || driverStore.onlineStatus === 0" @click="setOffline">
            下线
          </button>
        </div>
      </div>
    </section>

    <section class="mini-stat-grid">
      <div><span>服务分</span><strong>{{ serviceScore }}</strong></div>
      <div><span>订单</span><strong>{{ driverStore.currentOrderId || '--' }}</strong></div>
      <div><span>车辆</span><strong>{{ driverStore.vehicleId || '--' }}</strong></div>
    </section>

    <section class="tab-panel-scroll">
      <section v-show="activeTab === 0" class="h5-panel">
        <div class="section-title">
          <h2>当前行程</h2>
          <van-tag type="primary">{{ phaseLabel }}</van-tag>
        </div>
        <div class="map-surface">
          <span class="road road-a"></span>
          <span class="road road-b"></span>
          <span class="pin current"></span>
          <span class="pin order"></span>
          <p>接单区域</p>
        </div>
        <div class="route-card">
          <p><b>上车点</b><span>{{ driverStore.currentOrder?.fromAddress || '--' }}</span></p>
          <p><b>目的地</b><span>{{ driverStore.currentOrder?.toAddress || '--' }}</span></p>
          <p><b>状态</b><span>{{ formatOrderStatus(driverStore.currentOrder?.status) }}</span></p>
        </div>
        <button type="button" class="secondary-action" @click="loadDashboardData">刷新工作台</button>
      </section>

      <section v-show="activeTab === 1" class="h5-panel">
        <div class="filter-bar">
          <van-dropdown-menu>
            <van-dropdown-item v-model="orderMode" :options="orderModeOptions" @change="loadOrders(1)" />
            <van-dropdown-item v-model="orderStatus" :options="orderStatusOptions" @change="loadOrders(1)" />
          </van-dropdown-menu>
        </div>
        <div v-if="orders.length === 0" class="empty-state">等待后端返回订单</div>
        <article v-for="order in orders" :key="String(order.source || 'order') + '-' + String(order.orderId)" class="order-card">
          <div class="order-heading">
            <strong>{{ order.orderNo || '订单 ' + order.orderId }}</strong>
            <van-tag>{{ order.source === 'dispatch' ? formatDispatchStatus(order.dispatchStatus) : formatOrderStatus(order.status) }}</van-tag>
          </div>
          <p class="route-line">{{ order.fromAddress || '--' }} -> {{ order.toAddress || '--' }}</p>
          <div class="meta-row">
            <span>{{ formatPrice(order.estimatedPriceCents) }}</span>
            <span>{{ formatTime(order.createdAt) }}</span>
          </div>
          <div class="order-actions">
            <button type="button" @click="loadOrderDetail(order.orderId)">详情</button>
            <button type="button" @click="selectTrajectory(order.orderId)">轨迹</button>
            <button v-if="canAccept(order)" type="button" class="primary" @click="handleOrderAction('accept', order)">接单</button>
            <button v-if="canAccept(order)" type="button" @click="handleOrderAction('reject', order)">拒单</button>
            <button v-if="Number(order.status) === 2" type="button" @click="handleOrderAction('confirm-arrive', order)">到达</button>
            <button v-if="Number(order.status) === 2" type="button" @click="handleOrderAction('start-trip', order)">开始</button>
            <button v-if="Number(order.status) === 3" type="button" class="primary" @click="openFinish(order)">结束</button>
          </div>
        </article>
        <div class="pager">
          <button type="button" :disabled="orderPage <= 1" @click="loadOrders(orderPage - 1)">上一页</button>
          <span>{{ orderPage }} / {{ Math.max(1, Math.ceil(orderTotal / orderPageSize)) }}</span>
          <button type="button" :disabled="orderPage * orderPageSize >= orderTotal" @click="loadOrders(orderPage + 1)">下一页</button>
        </div>
      </section>

      <section v-show="activeTab === 2" class="h5-panel">
        <div class="section-title">
          <h2>车辆</h2>
          <button type="button" @click="loadVehicle()">查询</button>
        </div>
        <div class="info-list">
          <p><b>车辆ID</b><span>{{ driverStore.vehicle?.id || '--' }}</span></p>
          <p><b>车牌号</b><span>{{ driverStore.vehicle?.plateNo || '--' }}</span></p>
          <p><b>车型</b><span>{{ driverStore.vehicle?.brand || '--' }} {{ driverStore.vehicle?.model || '' }}</span></p>
          <p><b>状态</b><span>{{ formatVehicleStatus(driverStore.vehicle?.status) }}</span></p>
        </div>
        <van-form class="form-stack" @submit="submitVehicle">
          <van-field v-model="vehicleForm.plateNo" label="车牌号" placeholder="粤B12345" />
          <van-field v-model="vehicleForm.brand" label="品牌" placeholder="BYD" />
          <van-field v-model="vehicleForm.model" label="型号" placeholder="Han" />
          <van-field v-model="vehicleForm.color" label="颜色" placeholder="黑色" />
          <van-field v-model.number="vehicleForm.vehicleType" type="number" label="车辆类型" placeholder="1" />
          <van-field v-model="vehicleForm.registrationDate" type="date" label="注册日期" />
          <van-field v-model="vehicleForm.insuranceNo" label="保险单号" placeholder="INS-001" />
          <van-field v-model="vehicleForm.insuranceExpireAt" type="date" label="保险到期" />
          <button class="primary-action" type="submit">提交车辆</button>
        </van-form>
        <div class="two-actions">
          <button type="button" @click="submitVehicleUpdate">更新</button>
          <button type="button" class="danger" @click="removeVehicle">删除</button>
        </div>
      </section>

      <section v-show="activeTab === 3" class="h5-panel">
        <div class="section-title">
          <h2>认证</h2>
          <button type="button" @click="loadCertification()">查询</button>
        </div>
        <div class="info-list">
          <p><b>认证ID</b><span>{{ driverStore.certification?.id || '--' }}</span></p>
          <p><b>车辆ID</b><span>{{ driverStore.certification?.vehicleId || driverStore.vehicleId || '--' }}</span></p>
          <p><b>审核状态</b><span>{{ formatCertificationStatus(driverStore.certification?.auditStatus) }}</span></p>
          <p><b>审核备注</b><span>{{ driverStore.certification?.auditRemark || '--' }}</span></p>
        </div>
        <van-form class="form-stack" @submit="submitCertification">
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

      <section v-show="activeTab === 4" class="h5-panel">
        <div class="section-title">
          <h2>钱包</h2>
          <button type="button" @click="loadIncome()">刷新</button>
        </div>
        <div class="wallet-card">
          <p>累计收入</p>
          <strong>{{ formatPrice(incomeSummary.totalIncomeCents || walletSummary.balanceCents) }}</strong>
          <span>已完成订单 {{ incomeSummary.completedOrders ?? '--' }}</span>
        </div>
        <div class="income-grid">
          <div><span>今日</span><strong>{{ formatPrice(todayIncome.incomeCents || todayIncome.totalIncomeCents) }}</strong></div>
          <div><span>本周</span><strong>{{ formatPrice(weekIncome.incomeCents || weekIncome.totalIncomeCents) }}</strong></div>
        </div>
        <div v-if="incomeBills.length === 0" class="empty-state">等待后端返回收入账单</div>
        <article v-for="bill in incomeBills" :key="bill.id || bill.orderId" class="compact-card">
          <strong>{{ bill.orderNo || '订单 ' + bill.orderId }}</strong>
          <span>{{ formatPrice(bill.incomeCents) }} · {{ formatTime(bill.createdAt) }}</span>
        </article>
      </section>

      <section v-show="activeTab === 5" class="h5-panel">
        <div class="section-title">
          <h2>评价</h2>
          <button type="button" @click="loadReviews()">刷新</button>
        </div>
        <div v-if="reviews.length === 0" class="empty-state">等待后端返回评价</div>
        <article v-for="review in reviews" :key="review.id || review.reviewId || review.orderId" class="compact-card">
          <strong>订单 {{ review.orderId || '--' }} · {{ review.rating || '--' }} 分</strong>
          <span>{{ review.comment || '未填写评价' }}</span>
          <small>{{ formatTime(review.createdAt) }}</small>
        </article>
      </section>

      <section v-show="activeTab === 6" class="h5-panel">
        <div class="section-title">
          <h2>轨迹</h2>
        </div>
        <div class="trajectory-form">
          <van-field v-model.number="trajectoryOrderId" type="number" label="订单ID" placeholder="输入订单ID" />
          <p v-if="trajectoryError" class="trajectory-error">{{ trajectoryError }}</p>
          <button class="primary-action trajectory-action" type="button" @click="loadTrajectory">查询轨迹</button>
        </div>
        <div v-if="trajectoryPoints.length === 0" class="empty-state">暂无轨迹点</div>
        <article
          v-for="point in trajectoryPoints"
          :key="String(point.longitude) + '-' + String(point.latitude) + '-' + String(point.createdAt)"
          class="compact-card"
        >
          <strong>{{ point.longitude }}, {{ point.latitude }}</strong>
          <span>速度 {{ point.speedKmh || 0 }} km/h · 方向 {{ point.heading || 0 }}</span>
          <small>{{ formatTime(point.createdAt) }}</small>
        </article>
      </section>

      <section v-show="activeTab === 7" class="h5-panel">
        <div class="section-title">
          <h2>我的</h2>
          <button type="button" @click="refreshProfile">刷新</button>
        </div>
        <div class="info-list">
          <p><b>手机号</b><span>{{ driverStore.driver.phone || '--' }}</span></p>
          <p><b>驾驶证号</b><span>{{ driverStore.driver.driverLicenseNo || '--' }}</span></p>
          <p><b>账号状态</b><span>{{ formatDriverStatus(driverStore.driver.status) }}</span></p>
          <p><b>司机ID</b><span>{{ driverStore.driverId || '--' }}</span></p>
        </div>
        <van-form class="form-stack" @submit="submitProfile">
          <van-field v-model="profileForm.realName" label="姓名" placeholder="司机姓名" />
          <van-field v-model="profileForm.avatarUrl" label="头像地址" placeholder="头像 URL" />
          <van-field v-model="profileForm.driverLicenseNo" label="驾驶证号" placeholder="驾驶证号" />
          <button class="primary-action" type="submit">保存资料</button>
        </van-form>
      </section>
    </section>

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
  </main>
</template>

<script setup>
import { computed, onMounted, onUnmounted, reactive, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { closeToast, showConfirmDialog, showDialog, showLoadingToast, showToast } from 'vant'
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
  getTodayIncome,
  getVehicle,
  getWalletSummary,
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

const router = useRouter()
const driverStore = useDriverStore()

const tabItems = [
  { title: '首页', icon: 'wap-home-o' },
  { title: '订单', icon: 'orders-o' },
  { title: '车辆', icon: 'logistics' },
  { title: '认证', icon: 'certificate' },
  { title: '钱包', icon: 'balance-o' },
  { title: '评价', icon: 'comment-o' },
  { title: '轨迹', icon: 'location-o' },
  { title: '我的', icon: 'user-o' }
]

const activeTab = ref(0)
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
const todayIncome = ref({})
const weekIncome = ref({})
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
const certificationForm = reactive({ vehicleId: 0, idCardFront: '', idCardBack: '', driverLicense: '', vehicleLicense: '' })
const profileForm = reactive({ realName: '', avatarUrl: '', driverLicenseNo: '' })
const finishForm = reactive({ actualDistanceM: 0, actualDurationS: 0, actualPriceCents: 0 })

let heartbeatTimer = null
let locationTimer = null
let tripTimer = null
let geoWatchId = null
let pushSocket = null
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
  await setWorkStatus(() => setDriverOnline(workStatusPayload()), 1, '已上线，开始听单')
}

async function setOffline() {
  await setWorkStatus(() => setDriverOffline(workStatusPayload()), 0, '已下线')
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
    showToast(message)
  } catch (error) {
    closeToast()
    showToast(apiErrorMessage(error, '切换工作状态失败'))
  } finally {
    workLoading.value = false
  }
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
        lastLatitude = position.coords.latitude
        lastLongitude = position.coords.longitude
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

function reportCurrentLocation() {
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
  pushSocket.onmessage = (event) => handlePushMessage(event.data)
  pushSocket.onclose = () => {
    pushSocket = null
  }
  pushSocket.onerror = () => {
    pushSocket?.close()
  }
}

function handlePushMessage(raw) {
  try {
    const payload = JSON.parse(raw)
    if (payload.order || payload.data?.order) {
      const order = payload.order || payload.data.order
      driverStore.setCurrentOrder(order, Number(order.status) === 3 ? 'trip' : 'pickup')
      if (activeTab.value === 1) loadOrders(orderPage.value)
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
  const res = await safeApiCall(() => createVehicle(vehiclePayload()), '车辆提交失败')
  if (!res) return
  driverStore.setVehicle(res.vehicle || res)
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
    showToast('请至少上传一项资质图片')
    return
  }
  const res = await safeApiCall(() => uploadCertification(payload), '资质上传失败')
  if (!res) return
  showToast('资质已上传，请等待审核')
  await loadCertification()
}

async function readCertFile(event, field) {
  const file = event.target.files?.[0]
  if (!file) return
  certificationForm[field] = await fileToBase64(file)
}

async function loadIncome(config = {}) {
  const [summary, wallet, today, week, bills] = await Promise.allSettled([
    getIncomeSummary(config),
    getWalletSummary(config),
    getTodayIncome(config),
    getWeekIncome(config),
    listIncomeBills({ page: 1, pageSize: 20 }, config)
  ])
  incomeSummary.value = summary.status === 'fulfilled' ? summary.value : {}
  walletSummary.value = wallet.status === 'fulfilled' ? wallet.value : {}
  todayIncome.value = today.status === 'fulfilled' ? today.value : {}
  weekIncome.value = week.status === 'fulfilled' ? week.value : {}
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
.driver-home-page {
  min-height: 100vh;
  padding-bottom: calc(72px + env(safe-area-inset-bottom));
  background: #eef2f7;
  color: #111827;
}

.h5-status-hero {
  padding: 14px 12px 28px;
  border-radius: 0 0 24px 24px;
  background: linear-gradient(160deg, #6d4aff 0%, #4b2bc5 100%);
  color: #fff;
}

.driver-profile-row {
  display: flex;
  align-items: center;
  gap: 10px;
}

.driver-profile-row img,
.driver-profile-row > span {
  width: 48px;
  height: 48px;
  flex: 0 0 auto;
  display: grid;
  place-items: center;
  border: 2px solid rgba(255, 255, 255, .7);
  border-radius: 50%;
  background: #fff;
  color: #6d4aff;
  font-size: 21px;
  font-weight: 900;
  object-fit: cover;
}

.driver-profile-row div {
  min-width: 0;
  flex: 1;
}

.driver-profile-row p,
.driver-profile-row small {
  margin: 0;
  color: rgba(255, 255, 255, .76);
  font-size: 12px;
}

.driver-profile-row h1 {
  margin: 3px 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 20px;
  line-height: 1.15;
  letter-spacing: 0;
}

.icon-button {
  width: 40px;
  height: 40px;
  border: 0;
  border-radius: 10px;
  background: rgba(255, 255, 255, .16);
  color: #fff;
}

.online-control-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-top: 16px;
  padding: 12px;
  border-radius: 14px;
  background: rgba(255, 255, 255, .14);
}

.online-control-card small {
  display: block;
  color: rgba(255, 255, 255, .72);
  font-size: 12px;
}

.online-control-card strong {
  display: block;
  margin-top: 3px;
  font-size: 22px;
  line-height: 1.1;
}

.quick-action-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 8px;
  min-width: 148px;
}

.quick-action-grid button {
  min-height: 44px;
  border: 0;
  border-radius: 10px;
  font-weight: 900;
}

.quick-action-grid .go-online {
  background: #ffbe2e;
  color: #3f2b00;
}

.quick-action-grid .go-offline {
  background: #fff;
  color: #4b2bc5;
}

.mini-stat-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 8px;
  margin: -18px 12px 10px;
}

.mini-stat-grid div {
  min-width: 0;
  padding: 10px 8px;
  border-radius: 12px;
  background: #fff;
  box-shadow: 0 10px 26px rgba(77, 48, 160, .11);
  text-align: center;
}

.mini-stat-grid span {
  display: block;
  color: #667085;
  font-size: 11px;
}

.mini-stat-grid strong {
  display: block;
  margin-top: 5px;
  overflow-wrap: anywhere;
  font-size: 15px;
}

.tab-panel-scroll {
  min-height: 0;
}

.h5-panel {
  display: grid;
  gap: 10px;
  padding: 2px 12px 12px;
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
  margin: 0;
  font-size: 17px;
  line-height: 1.2;
  letter-spacing: 0;
}

.section-title button,
.secondary-action,
.pager button,
.order-actions button,
.two-actions button {
  min-height: 38px;
  border: 1px solid #d8deea;
  border-radius: 10px;
  background: #fff;
  color: #6d4aff;
  font-weight: 800;
}

.map-surface {
  position: relative;
  min-height: 156px;
  overflow: hidden;
  border: 1px solid #ded6ff;
  border-radius: 14px;
  background-color: #f1edff;
  background-image:
    linear-gradient(24deg, transparent 46%, rgba(255, 255, 255, .9) 47%, rgba(255, 255, 255, .9) 51%, transparent 52%),
    linear-gradient(112deg, transparent 42%, rgba(255, 255, 255, .72) 43%, rgba(255, 255, 255, .72) 47%, transparent 48%);
}

.map-surface p {
  position: absolute;
  top: 10px;
  right: 10px;
  margin: 0;
  padding: 5px 9px;
  border-radius: 999px;
  background: rgba(255, 255, 255, .9);
  color: #6d4aff;
  font-size: 11px;
  font-weight: 800;
}

.road {
  position: absolute;
  height: 7px;
  border-radius: 999px;
  background: rgba(109, 74, 255, .18);
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
  width: 15px;
  height: 15px;
  border: 3px solid #fff;
  border-radius: 50%;
  box-shadow: 0 4px 12px rgba(77, 48, 160, .2);
}

.pin.current {
  top: 50%;
  left: 31%;
  background: #6d4aff;
}

.pin.order {
  top: 29%;
  right: 23%;
  background: #ffbe2e;
}

.route-card,
.order-card,
.compact-card,
.info-list,
.wallet-card,
.income-grid > div,
.trajectory-form {
  padding: 12px;
  border: 1px solid #e6eaf2;
  border-radius: 12px;
  background: #fff;
}

.route-card,
.info-list {
  display: grid;
  gap: 9px;
}

.route-card p,
.info-list p {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  margin: 0;
}

.route-card span,
.info-list span {
  min-width: 0;
  overflow-wrap: anywhere;
  color: #475467;
  text-align: right;
}

.filter-bar {
  overflow: hidden;
  border-radius: 12px;
}

.driver-home-page :deep(.van-dropdown-menu) {
  box-shadow: none;
}

.driver-home-page :deep(.van-field) {
  --van-field-label-width: 74px;
  overflow: hidden;
  border-radius: 10px;
}

.order-card,
.compact-card {
  display: grid;
  gap: 8px;
}

.order-heading strong,
.compact-card strong {
  min-width: 0;
  overflow-wrap: anywhere;
}

.route-line,
.meta-row,
.compact-card span,
.compact-card small {
  margin: 0;
  color: #667085;
  font-size: 12px;
}

.order-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.order-actions button {
  min-height: 34px;
  padding: 0 10px;
  font-size: 12px;
}

.order-actions .primary,
.primary-action {
  border: 0;
  background: #6d4aff;
  color: #fff;
}

.primary-action {
  width: 100%;
  min-height: 48px;
  border-radius: 10px;
  font-size: 16px;
  font-weight: 900;
}

.secondary-action {
  width: 100%;
}

.empty-state {
  min-height: 92px;
  display: grid;
  place-items: center;
  border: 1px dashed #d9d2ee;
  border-radius: 12px;
  background: #fbfaff;
  color: #667085;
  font-size: 13px;
}

.form-stack {
  display: grid;
  gap: 8px;
}

.two-actions {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 8px;
}

.two-actions .danger {
  border-color: #fecaca;
  color: #dc2626;
}

.file-grid {
  display: grid;
  gap: 8px;
  padding: 12px;
  border-radius: 12px;
  background: #fff;
}

.file-grid label {
  display: grid;
  gap: 8px;
  color: #475467;
  font-size: 12px;
}

.wallet-card {
  background: #4b2bc5;
  color: #fff;
}

.wallet-card p {
  margin: 0;
  color: rgba(255, 255, 255, .72);
  font-size: 12px;
}

.wallet-card strong {
  display: block;
  margin-top: 6px;
  font-size: 30px;
  line-height: 1.1;
}

.wallet-card span {
  display: block;
  margin-top: 6px;
  color: rgba(255, 255, 255, .76);
  font-size: 12px;
}

.income-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 8px;
}

.income-grid span {
  color: #667085;
  font-size: 12px;
}

.income-grid strong {
  display: block;
  margin-top: 5px;
  font-size: 18px;
}

.trajectory-form {
  display: grid;
  gap: 8px;
}

.trajectory-error {
  margin: 0;
  color: #dc2626;
  font-size: 13px;
  word-break: break-word;
}

.finish-panel {
  padding: 14px 12px 18px;
}

.finish-panel h2 {
  margin: 0 0 12px;
  font-size: 17px;
  text-align: center;
}

.driver-tabbar {
  left: 50%;
  right: auto;
  width: min(100vw, 390px);
  height: calc(60px + env(safe-area-inset-bottom));
  border-top: 1px solid #e6eaf2;
  box-shadow: 0 -8px 24px rgba(15, 23, 42, .08);
  transform: translateX(-50%);
  --van-tabbar-item-active-color: #6d4aff;
}

.driver-tabbar :deep(.van-tabbar-item) {
  min-width: 0;
  padding: 5px 0 4px;
  color: #667085;
  font-size: 10px;
}

.driver-tabbar :deep(.van-tabbar-item__icon) {
  margin-bottom: 2px;
  font-size: 18px;
}

button:disabled {
  opacity: .55;
}

@media (max-width: 430px) {
  .driver-tabbar {
    width: 100vw;
  }
}
</style>
