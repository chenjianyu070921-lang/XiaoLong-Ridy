<template>
  <main class="driver-home-page">
    <div id="driver-home-popups"></div>

    <section v-show="activeTab === 0" class="home-workbench">
      <header class="driver-status-bar">
        <button type="button" class="driver-avatar-button" aria-label="编辑司机资料" @click="openProfileEdit">
          <img v-if="driverStore.driver.avatarUrl" :src="driverStore.driver.avatarUrl" alt="司机头像" />
          <span v-else class="avatar-fallback">{{ driverStore.displayName.slice(0, 1) || '--' }}</span>
        </button>
        <div class="status-copy">
          <strong>{{ workStatusText }}</strong>
          <span>{{ driverStore.displayName || '司机' }}</span>
        </div>
        <button
          type="button"
          class="status-toggle"
          :class="{ online: driverStore.onlineStatus === 1, driving: driverStore.onlineStatus === 2 || driverStore.tripPhase === 'trip' }"
          :disabled="workActionDisabled"
          @click="toggleStatusBarOnline"
        >
          {{ driverStore.onlineStatus === 1 ? '在线' : driverStore.onlineStatus === 2 || driverStore.tripPhase === 'trip' ? '行驶中' : '离线' }}
        </button>
        <div class="status-metric">
          <span>服务分</span>
          <b>{{ serviceScore || '--' }}</b>
        </div>
        <div class="status-metric">
          <span>今日预估</span>
          <b>{{ formatPrice(todayIncome.totalIncomeCents) }}</b>
        </div>
        <button type="button" class="message-button" aria-label="消息通知" @click="showNotifications">
          <van-icon name="bell" />
        </button>
      </header>

      <section class="home-map-stage">
        <div ref="homeMapContainer" class="home-amap" aria-label="司机首页实时地图"></div>
        <div v-if="homeMapStatusText" class="map-state" :class="{ error: homeMapError }">{{ homeMapStatusText }}</div>
        <div class="heatmap-legend">
          <span>低</span><i class="heat-low"></i><i class="heat-mid"></i><i class="heat-high"></i><b>高</b>
        </div>
        <div v-if="selectedHomeOrder" class="route-direction-chip">
          <van-icon name="guide-o" />
          <span>{{ selectedHomeOrder.fromAddress || '--' }} -> {{ selectedHomeOrder.toAddress || '--' }}</span>
        </div>
        <div class="map-floating-actions">
          <button type="button" aria-label="刷新热力图" :disabled="heatmapLoading" @click="refreshHomeWorkbench">
            <van-icon name="replay" />
          </button>
          <button type="button" aria-label="回到当前位置" @click="centerHomeMapOnDriver">
            <van-icon name="aim" />
          </button>
          <button type="button" aria-label="打开热力图详情" @click="openHeatmap">
            <van-icon name="fire-o" />
          </button>
        </div>
      </section>

      <section class="home-floating-panel">
        <template v-if="homePanelMode === 'driving'">
          <div class="panel-heading">
            <span>行驶中</span>
            <strong>{{ selectedHomeOrder?.toAddress || '导航中' }}</strong>
          </div>
          <p class="route-line">{{ selectedHomeOrder?.fromAddress || '--' }} -> {{ selectedHomeOrder?.toAddress || '--' }}</p>
          <div class="panel-actions three-actions">
            <button type="button" class="primary" @click="navigateToPickup(selectedHomeOrder)">导航</button>
            <button type="button" @click="contactPassenger(selectedHomeOrder)">联系乘客</button>
            <button type="button" @click="reportAbnormal">异常上报</button>
            <button type="button" @click="openFinish(selectedHomeOrder)">结束行程</button>
          </div>
        </template>
        <template v-else-if="selectedHomeOrder">
          <div class="panel-heading">
            <span>新订单 {{ passengerTail(selectedHomeOrder) }}</span>
            <strong>{{ formatPrice(selectedHomeOrder.estimatedPriceCents) }}</strong>
          </div>
          <p class="route-line">{{ selectedHomeOrder.fromAddress || '--' }} -> {{ selectedHomeOrder.toAddress || '--' }}</p>
          <div class="order-brief-grid">
            <span>{{ formatDistance(selectedHomeOrder.distanceMeters || selectedHomeOrder.estimatedDistanceM) }} km</span>
            <span>{{ formatDuration(selectedHomeOrder.estimatedDurationS) }}</span>
            <span>{{ formatOrderStatus(selectedHomeOrder.status || 1) }}</span>
          </div>
          <div class="panel-actions three-actions">
            <button type="button" class="primary" @click="handleOrderAction('accept', selectedHomeOrder)">接单</button>
            <button type="button" @click="handleOrderAction('reject', selectedHomeOrder)">拒单</button>
            <button type="button" @click="previewHomeOrder(selectedHomeOrder)">先看后接</button>
          </div>
        </template>
        <template v-else>
          <div class="panel-heading">
            <span>{{ workStatusHint }}</span>
            <strong>{{ phaseLabel || '--' }}</strong>
          </div>
          <div class="idle-controls">
            <button type="button" class="work-primary-action go-online" :disabled="workActionDisabled" @click="startAcceptingOrders">
              <van-icon name="play-circle-o" />
              <span>开始接单</span>
            </button>
            <label class="listen-switch">
              <span>听单模式</span>
              <van-switch v-model="listenModeEnabled" size="22px" />
            </label>
          </div>
          <button type="button" class="home-route-button" @click="navigateHomeRoute">
            <van-icon name="wap-home-o" />
            <span>{{ homeRouteLabel }}</span>
          </button>
        </template>
      </section>
    </section>
    <section v-if="activePanelComponent" class="tab-panel-scroll">
      <component
        :is="activePanelComponent"
        v-bind="activePanelProps"
        @edit-profile="openProfileEdit"
        @update:trajectory-order-id="trajectoryOrderId = $event"
        @load-trajectory="loadTrajectory"
      />
    </section>

    <van-tabbar v-model="activeTab" class="driver-tabbar" fixed safe-area-inset-bottom>
      <van-tabbar-item v-for="item in tabItems" :key="item.title" :icon="item.icon">{{ item.title }}</van-tabbar-item>
    </van-tabbar>

    <van-popup v-model:show="finishVisible" round position="bottom" teleport="#driver-home-popups">
      <section class="finish-panel">
        <h2>结束行程</h2>
        <van-form @submit="submitFinishTrip">
          <van-field v-model.number="finishForm.actualDistanceM" type="number" label="实际里程(米)" />
          <van-field v-model.number="finishForm.actualDurationS" type="number" label="实际时长(秒)" />
          <button class="primary-action" type="submit">确认结束</button>
        </van-form>
      </section>
    </van-popup>

    <van-popup v-model:show="heatmapVisible" teleport="#driver-home-popups" class="driver-heatmap-popup" round position="bottom" :style="heatmapPhoneSheetStyle">
      <section class="heatmap-panel heatmap-h5-sheet">
        <div class="heatmap-sheet-grabber" aria-hidden="true"></div>
        <div class="heatmap-heading heatmap-sheet-header">
          <div>
            <h2>附近热力</h2>
            <p>{{ heatmapSummary }}</p>
          </div>
          <van-tag type="warning">{{ heatmapTotalOrders }}单</van-tag>
        </div>
        <div class="heatmap-map-shell">
          <div ref="heatmapMapContainer" class="driver-heatmap-map" aria-label="附近订单热力图"></div>
          <div v-if="heatmapStatusText" class="map-state" :class="{ error: heatmapMapError }">{{ heatmapStatusText }}</div>
          <div class="heatmap-floating-actions">
            <button type="button" class="heatmap-refresh" aria-label="刷新热力图" :disabled="heatmapLoading" @click="refreshHeatmap">
              <van-icon name="replay" />
            </button>
          </div>
          <div class="heatmap-badge">
            <span>{{ formatDistance(heatmapRadiusMeters) }}km</span>
            <strong>{{ heatmapTotalOrders }}单</strong>
          </div>
        </div>
        <div v-if="heatmapPoints.length" class="heatmap-chip-strip">
          <div v-for="point in heatmapPoints" :key="point.longitude + ':' + point.latitude" class="heatmap-chip">
            <span><van-icon name="fire-o" /></span>
            <div>
              <strong>{{ point.weight }} 单</strong>
              <small>{{ Number(point.longitude).toFixed(6) }}, {{ Number(point.latitude).toFixed(6) }}</small>
            </div>
          </div>
        </div>
      </section>
    </van-popup>

    <van-popup v-model:show="withdrawVisible" round position="bottom" teleport="#driver-home-popups">
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
import { computed, nextTick, onMounted, onUnmounted, reactive, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { closeToast, showConfirmDialog, showDialog, showLoadingToast, showToast } from 'vant'
import AMapLoader from '@amap/amap-jsapi-loader'
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
  getOrderHeatmap,
  getOrderTrajectory,
  getTodayIncome,
  getVehicle,
  getWeekIncome,
  heartbeatDriver,
  listAvailableOrders,
  listDriverDispatches,
  listDriverOrders,
  listIncomeBills,
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
import DriverOrdersPanel from '@/components/driver-home/DriverOrdersPanel.vue'
import DriverMinePanel from '@/components/driver-home/DriverMinePanel.vue'
import { getAmapConfig } from '@/config/amap'
import { normalizeBrowserLocationForAmap } from '@/utils/geo'
import '@/styles/driver-home-panels.css'

const router = useRouter()
const driverStore = useDriverStore()

const tabItems = [
  { title: '首页', icon: 'wap-home-o' },
  { title: '订单', icon: 'orders-o' },
  { title: '我的', icon: 'user-o' }
]

const activeTab = ref(0)
const tabPanelComponents = [
  null,
  DriverOrdersPanel,
  DriverMinePanel
]
const activePanelComponent = computed(() => tabPanelComponents[activeTab.value] || null)
const activePanelProps = computed(() => {
  const commonFormatters = { formatPrice, formatDistance, formatTime, formatOrderStatus, formatDispatchStatus }
  if (activeTab.value === 0) {
    return {}
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
      nearbyOrders: nearbyOrders.value,
      nearbyOrderLoading: nearbyOrderLoading.value,
      nearbyOrderPage: nearbyOrderPage.value,
      nearbyOrderPageSize,
      nearbyOrderTotal: nearbyOrderTotal.value,
      nearbyOrderPopupVisible: nearbyOrderPopupVisible.value,
      nearbyOrderExpandedOrders: nearbyOrderExpandedOrders.value,
      nearbyOrderExpandedLoading: nearbyOrderExpandedLoading.value,
      nearbyOrderExpandedPage: nearbyOrderExpandedPage.value,
      nearbyOrderExpandedPageSize,
      nearbyOrderExpandedTotal: nearbyOrderExpandedTotal.value,
      canAccept
    }
  }
  if (activeTab.value === 2) {
    return {
      driverStore,
      serviceScore: serviceScore.value,
      todayIncome: todayIncome.value,
      incomeSummary: incomeSummary.value,
      formatDriverStatus,
      formatPrice,
      formatVehicleStatus,
      formatCertificationStatus,
      trajectoryOrderId: trajectoryOrderId.value,
      trajectoryError: trajectoryError.value,
      trajectoryPoints: trajectoryPoints.value,
      formatTime
    }
  }
  return {}
})
const workLoading = ref(false)
const serviceScore = ref('--')
const orders = ref([])
const orderMode = ref('orders')
const orderStatus = ref(0)
const orderPage = ref(1)
const orderPageSize = ref(8)
const orderTotal = ref(0)
const nearbyOrders = ref([])
const nearbyOrderLoading = ref(false)
const nearbyOrderPage = ref(1)
const nearbyOrderPageSize = 5
const nearbyOrderTotal = ref(0)
const nearbyOrderPopupVisible = ref(false)
const nearbyOrderExpandedOrders = ref([])
const nearbyOrderExpandedLoading = ref(false)
const nearbyOrderExpandedPage = ref(1)
const nearbyOrderExpandedPageSize = 10
const nearbyOrderExpandedTotal = ref(0)
const dashboardSessionReady = ref(false)
const incomeSummary = ref({})
const todayIncome = ref({})
const weekIncome = ref({})
const incomeBills = ref([])
const trajectoryOrderId = ref('')
const trajectoryPoints = ref([])
const trajectoryError = ref('')
const finishVisible = ref(false)
const finishOrder = ref(null)
const heatmapVisible = ref(false)
const heatmapLoading = ref(false)
const heatmapRadiusMeters = 5000
const heatmapPoints = ref([])
const heatmapCenter = ref(null)
const homeMapContainer = ref(null)
const homeMapReady = ref(false)
const homeMapError = ref('')
const heatmapMapContainer = ref(null)
const heatmapMapReady = ref(false)
const heatmapMapError = ref('')
const withdrawVisible = ref(false)
const withdrawLoading = ref(false)
const withdrawForm = reactive({ amount: '', payeeName: '', payAccount: '' })
const listenModeEnabled = ref(localStorage.getItem('driverListenModeEnabled') !== '0')
const previewOrder = ref(null)
const homeRouteAddress = ref(localStorage.getItem('driverHomeRouteAddress') || '')

const orderModeOptions = [
  { text: '我的订单', value: 'orders' },
  { text: '派单记录', value: 'dispatches' }
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
let homeAMap = null
let homeMapInstance = null
let homeDriverMarker = null
let homeHeatmapLayer = null
let homeOrderMarkers = []
let homeRouteLine = null
let heatmapAMap = null
let heatmapMapInstance = null
let heatmapLayer = null
let heatmapCenterMarker = null
let heatmapGeolocation = null
let heatmapRefreshTimer = null

const phaseLabel = computed(() => {
  if (driverStore.tripPhase === 'pickup') return '接驾中'
  if (driverStore.tripPhase === 'trip') return '行程中'
  return driverStore.onlineStatus === 1 ? '听单中' : '未上线'
})

const workStatusText = computed(() => {
  if (driverStore.onlineStatus === 2 || driverStore.tripPhase === 'trip') return '行程中服务'
  return driverStore.onlineStatus === 1 ? '在线听单中' : '已下线'
})

const workStatusHint = computed(() => {
  if (driverStore.onlineStatus === 2 || driverStore.tripPhase === 'trip') return '当前有进行中的订单，请完成行程后再调整听单状态'
  return driverStore.onlineStatus === 1 ? '正在接收附近订单，可随时停止听单' : '上线后将接收附近订单推送'
})

const workActionDisabled = computed(() => workLoading.value || driverStore.onlineStatus === 2 || driverStore.tripPhase === 'trip')
const heatmapSummary = computed(() => {
  if (!heatmapCenter.value) return '按司机当前位置实时刷新'
  return `${Number(heatmapCenter.value.longitude).toFixed(5)}, ${Number(heatmapCenter.value.latitude).toFixed(5)}`
})
const heatmapTotalOrders = computed(() => heatmapPoints.value.reduce((total, point) => total + Number(point.weight || 0), 0))
const selectedHomeOrder = computed(() => {
  return driverStore.currentOrder || previewOrder.value || null
})
const homePanelMode = computed(() => {
  if (driverStore.tripPhase === 'pickup' || driverStore.tripPhase === 'trip' || driverStore.onlineStatus === 2 || Number(selectedHomeOrder.value?.status) === 3) return 'driving'
  if (selectedHomeOrder.value) return 'order'
  return 'idle'
})
const homeMapStatusText = computed(() => {
  if (homeMapError.value) return homeMapError.value
  if (!homeMapReady.value) return '地图加载中...'
  if (heatmapLoading.value) return '正在刷新订单热力...'
  if (!heatmapPoints.value.length && driverStore.onlineStatus === 1) return '附近暂无线索，继续听单'
  return ''
})
const homeRouteLabel = computed(() => homeRouteAddress.value ? '回家顺路: ' + homeRouteAddress.value : '设置回家顺路模式')
const heatmapStatusText = computed(() => {
  if (heatmapMapError.value) return heatmapMapError.value
  if (!heatmapMapReady.value) return '地图加载中...'
  if (heatmapLoading.value) return '正在刷新热力...'
  if (!heatmapPoints.value.length) return '附近暂无待接单订单'
  return ''
})
const heatmapPhoneSheetStyle = {
  height: 'min(88vh, 760px)',
  width: 'min(100vw, 390px)'
}

onMounted(async () => {
  await loadDashboardData()
  await ensureHomeMap()
  await refreshHomeWorkbench()
  if (driverStore.onlineStatus > 0) startRealtimeWork()
})

onUnmounted(() => {
  stopRealtimeWork()
  stopHeatmapRealtime()
  destroyHomeMap()
  destroyHeatmapMap()
})

watch(activeTab, (tab) => {
  if (tab === 0) {
    nextTick(() => refreshHomeWorkbench())
  }
  loadCurrentTabData()
})

watch(listenModeEnabled, (enabled) => {
  localStorage.setItem('driverListenModeEnabled', enabled ? '1' : '0')
})

watch(selectedHomeOrder, () => {
  renderHomeMapData()
})

watch(heatmapVisible, async (visible) => {
  if (!visible) {
    stopHeatmapRealtime()
    destroyHeatmapMap()
    return
  }
  await nextTick()
  await ensureHeatmapMap()
  await refreshHeatmap()
  if (!heatmapVisible.value) return
  startHeatmapRealtime()
})

async function loadDashboardData() {
  dashboardSessionReady.value = false
  const profile = await safeApiCall(() => driverStore.refreshProfile({ silentError: true }), null, { silent: true })
  if (!profile) {
    serviceScore.value = '--'
    return
  }
  dashboardSessionReady.value = true
  syncForms()

  const score = await safeApiCall(() => getDriverAiScore({ silentError: true }), null, { silent: true })
  if (score) {
    serviceScore.value = Number(score.aiScore || score.score || 0).toFixed(1)
  }
  await loadIncome({ silentError: true })
  await loadCurrentTabData()
  if (activeTab.value === 0) {
    await refreshHomeWorkbench()
  }
}

async function loadCurrentTabData() {
  if (!dashboardSessionReady.value) return null
  const config = { silentError: true }
  if (activeTab.value === 0) return null
  if (activeTab.value === 1) {
    return Promise.allSettled([
      safeApiCall(() => loadOrders(1, config), null, { silent: true }),
      safeApiCall(() => loadNearbyOrders(1, config), null, { silent: true })
    ])
  }
  if (activeTab.value === 2) {
    return Promise.allSettled([
      safeApiCall(() => loadIncome(config), null, { silent: true }),
      safeApiCall(() => loadVehicle(config), null, { silent: true }),
      safeApiCall(() => loadCertification(config), null, { silent: true })
    ])
  }
  return null
}

function syncForms() {
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
  await setWorkStatus(() => setDriverOnline(workStatusPayload()), 1, '已上线，开始接单')
}

async function setOffline() {
  await ensureWorkLocation()
  await setWorkStatus(() => setDriverOffline(workStatusPayload()), 0, '已下线')
}

function toggleListening() {
  if (driverStore.onlineStatus === 1) return setOffline()
  if (driverStore.onlineStatus === 0) return setOnline()
}

function toggleStatusBarOnline() {
  return toggleListening()
}

async function startAcceptingOrders() {
  if (!listenModeEnabled.value) {
    listenModeEnabled.value = true
  }
  if (driverStore.onlineStatus === 0) {
    await setOnline()
    return
  }
  await refreshHomeWorkbench()
  showToast('正在接单中')
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
    await loadNearbyOrders(nearbyOrderPage.value, { silentError: true })
    showToast(message)
  } catch (error) {
    closeToast()
    showToast(apiErrorMessage(error, '切换工作状态失败'))
  } finally {
    workLoading.value = false
  }
}


// workLocationDefault 仅在浏览器无法获取定位时作为兜底，保证司机能上线听单（演示/联调环境可重新配置）。
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
  const normalized = normalizeBrowserLocationForAmap(location)
  lastLongitude = Number(normalized.longitude)
  lastLatitude = Number(normalized.latitude)
  syncHomeMapCenterToDriver({ longitude: lastLongitude, latitude: lastLatitude })
  renderHomeMapData()
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

function openProfileEdit() {
  router.push('/profile/edit')
}

function showNotifications() {
  showToast('暂无新消息')
}

async function refreshHomeWorkbench() {
  if (activeTab.value !== 0 || !dashboardSessionReady.value) return
  await ensureHomeMap()
  await Promise.allSettled([
    refreshHomeHeatmap()
  ])
  renderHomeMapData()
}

async function ensureHomeMap() {
  if (homeMapInstance) return
  await nextTick()
  const container = getConnectedHomeMapContainer()
  if (!container) return
  const { key, securityCode } = getAmapConfig()
  if (!key) {
    homeMapError.value = '未配置高德地图Key'
    return
  }
  try {
    const currentLocation = await ensureWorkLocation()
    if (securityCode) window._AMapSecurityConfig = { securityJsCode: securityCode }
    homeAMap = await AMapLoader.load({ key, version: '2.0', plugins: ['AMap.HeatMap'] })
    if (homeMapInstance) return
    const activeContainer = getConnectedHomeMapContainer()
    if (!activeContainer) return
    window.AMap = homeAMap
    homeMapInstance = new homeAMap.Map(activeContainer, {
      zoom: 15,
      viewMode: '2D',
      center: [currentLocation.longitude, currentLocation.latitude]
    })
    homeHeatmapLayer = new homeAMap.HeatMap(homeMapInstance, {
      radius: 24,
      opacity: [0, 0.82],
      gradient: {
        0.2: 'rgba(37, 99, 235, 0)',
        0.45: 'rgba(37, 99, 235, .48)',
        0.65: 'rgba(34, 197, 94, .64)',
        0.82: 'rgba(245, 158, 11, .78)',
        1: 'rgba(239, 68, 68, .9)'
      }
    })
    homeMapReady.value = true
    homeMapError.value = ''
    syncHomeMapCenterToDriver(currentLocation)
    renderHomeMapData()
  } catch (error) {
    homeMapReady.value = false
    homeMapError.value = '高德地图加载失败'
    console.error('driver home AMap error:', error)
  }
}

async function refreshHomeHeatmap() {
  if (!dashboardSessionReady.value) return
  heatmapLoading.value = true
  try {
    const location = await ensureWorkLocation()
    heatmapCenter.value = location
    const data = await getOrderHeatmap({
      longitude: location.longitude,
      latitude: location.latitude,
      radiusMeters: heatmapRadiusMeters
    }, { silentError: true })
    heatmapPoints.value = Array.isArray(data?.points) ? data.points : []
  } catch (error) {
    showToast(apiErrorMessage(error, '热力图加载失败'))
  } finally {
    heatmapLoading.value = false
  }
}

function renderHomeMapData() {
  if (!homeMapInstance || !homeAMap) return
  renderHomeDriverMarker()
  renderHomeHeatmapLayer()
  renderHomeOrderMarkers()
  renderHomeRouteLine()
}

function renderHomeDriverMarker() {
  const location = readRememberedWorkLocation() || workLocationDefault
  const position = [Number(location.longitude), Number(location.latitude)]
  if (!homeDriverMarker) {
    homeDriverMarker = new homeAMap.Marker({ position, title: '司机当前位置', anchor: 'center', zIndex: 120 })
    homeMapInstance.add(homeDriverMarker)
  } else {
    homeDriverMarker.setPosition(position)
  }
}

function getConnectedHomeMapContainer() {
  const container = homeMapContainer.value
  if (!container || !container.isConnected) return null
  return container
}

function syncHomeMapCenterToDriver(location = readRememberedWorkLocation(), zoom = 15) {
  if (!homeMapInstance || !location) return
  const longitude = Number(location.longitude)
  const latitude = Number(location.latitude)
  if (!Number.isFinite(longitude) || !Number.isFinite(latitude) || (longitude === 0 && latitude === 0)) return
  homeMapInstance.setZoomAndCenter(zoom, [longitude, latitude])
}

function renderHomeHeatmapLayer() {
  const data = heatmapPoints.value
    .map((point) => ({
      lng: Number(point.longitude),
      lat: Number(point.latitude),
      count: Number(point.weight || 0)
    }))
    .filter((point) => Number.isFinite(point.lng) && Number.isFinite(point.lat) && point.count > 0)
  homeHeatmapLayer?.setDataSet({ data, max: Math.max(1, ...data.map((point) => point.count)) })
}

function renderHomeOrderMarkers() {
  homeOrderMarkers.forEach((marker) => homeMapInstance.remove(marker))
  homeOrderMarkers = []
  const ordersWithPoint = nearbyOrders.value
    .map((order) => ({ order, position: orderPickupPosition(order) }))
    .filter((item) => item.position)
  homeOrderMarkers = ordersWithPoint.map(({ order, position }) => new homeAMap.Marker({
    position,
    title: order.fromAddress || '乘客上车点',
    anchor: 'bottom-center',
    zIndex: 110
  }))
  if (homeOrderMarkers.length) homeMapInstance.add(homeOrderMarkers)
}

function renderHomeRouteLine() {
  if (homeRouteLine) {
    homeMapInstance.remove(homeRouteLine)
    homeRouteLine = null
  }
  const order = selectedHomeOrder.value
  const from = orderPickupPosition(order)
  const to = orderDropoffPosition(order)
  if (!from || !to) return
  homeRouteLine = new homeAMap.Polyline({
    path: [from, to],
    strokeColor: '#2563EB',
    strokeWeight: 6,
    strokeOpacity: 0.86,
    showDir: true
  })
  homeMapInstance.add(homeRouteLine)
}

function centerHomeMapOnDriver() {
  const location = readRememberedWorkLocation() || workLocationDefault
  syncHomeMapCenterToDriver(location, 15)
}

function destroyHomeMap() {
  homeHeatmapLayer?.setDataSet?.({ data: [], max: 1 })
  homeMapInstance?.destroy()
  homeAMap = null
  homeMapInstance = null
  homeDriverMarker = null
  homeHeatmapLayer = null
  homeOrderMarkers = []
  homeRouteLine = null
  homeMapReady.value = false
  homeMapError.value = ''
}

function openHeatmap() {
  heatmapPoints.value = []
  heatmapMapError.value = ''
  heatmapVisible.value = true
}

async function ensureHeatmapMap() {
  if (heatmapMapInstance || !heatmapMapContainer.value) return
  const { key, securityCode } = getAmapConfig()
  if (!key) {
    heatmapMapError.value = '未配置高德地图Key'
    return
  }
  try {
    if (securityCode) window._AMapSecurityConfig = { securityJsCode: securityCode }
    heatmapAMap = await AMapLoader.load({ key, version: '2.0', plugins: ['AMap.HeatMap', 'AMap.Geolocation'] })
    window.AMap = heatmapAMap
    heatmapMapInstance = new heatmapAMap.Map(heatmapMapContainer.value, {
      zoom: 12,
      viewMode: '2D',
      center: readRememberedWorkLocation()
        ? [lastLongitude, lastLatitude]
        : [workLocationDefault.longitude, workLocationDefault.latitude]
    })
    heatmapGeolocation = new heatmapAMap.Geolocation({
      enableHighAccuracy: true,
      timeout: 10000,
      zoomToAccuracy: true,
      position: 'RB',
      offset: [16, 108]
    })
    heatmapMapInstance.addControl(heatmapGeolocation)
    heatmapLayer = new heatmapAMap.HeatMap(heatmapMapInstance, {
      radius: 24,
      opacity: [0, 0.86],
      gradient: {
        0.2: 'rgba(37, 99, 235, 0)',
        0.45: 'rgba(37, 99, 235, .48)',
        0.65: 'rgba(34, 197, 94, .64)',
        0.82: 'rgba(245, 158, 11, .78)',
        1: 'rgba(239, 68, 68, .9)'
      }
    })
    heatmapMapReady.value = true
    heatmapMapError.value = ''
  } catch (error) {
    heatmapMapInstance?.destroy()
    heatmapAMap = null
    heatmapMapInstance = null
    heatmapLayer = null
    heatmapCenterMarker = null
    heatmapGeolocation = null
    heatmapMapReady.value = false
    heatmapMapError.value = '高德地图加载失败'
    console.error('driver heatmap AMap error:', error)
  }
}

async function refreshHeatmap() {
  heatmapLoading.value = true
  try {
    await ensureHeatmapMap()
    const location = await ensureHeatmapLocation()
    heatmapCenter.value = location
    const data = await getOrderHeatmap({
      longitude: location.longitude,
      latitude: location.latitude,
      radiusMeters: heatmapRadiusMeters
    }, { silentError: true })
    heatmapPoints.value = Array.isArray(data?.points) ? data.points : []
    renderHeatmapPoints()
  } catch (error) {
    showToast(apiErrorMessage(error, '热力图加载失败'))
  } finally {
    heatmapLoading.value = false
  }
}

async function ensureHeatmapLocation() {
  const watchedLocation = readRememberedWorkLocation()
  if (watchedLocation) return watchedLocation
  try {
    return rememberWorkLocation(await locateHeatmapByAMap())
  } catch {
    return ensureWorkLocation()
  }
}

function readRememberedWorkLocation() {
  if (Number.isFinite(lastLongitude) && Number.isFinite(lastLatitude) && (lastLongitude !== 0 || lastLatitude !== 0)) {
    return { longitude: lastLongitude, latitude: lastLatitude }
  }
  return null
}

function locateHeatmapByAMap() {
  return new Promise((resolve, reject) => {
    if (!heatmapGeolocation) {
      reject(new Error('高德定位未初始化'))
      return
    }
    heatmapGeolocation.getCurrentPosition((status, result) => {
      if (status === 'complete' && result?.position) {
        resolve({
          longitude: result.position.lng,
          latitude: result.position.lat
        })
        return
      }
      reject(result || new Error('高德定位失败'))
    })
  })
}

function renderHeatmapPoints() {
  if (!heatmapMapInstance || !heatmapAMap) return
  const center = heatmapCenter.value || workLocationDefault
  const centerPosition = [Number(center.longitude), Number(center.latitude)]
  const data = heatmapPoints.value
    .map((point) => ({
      lng: Number(point.longitude),
      lat: Number(point.latitude),
      count: Number(point.weight || 0)
    }))
    .filter((point) => Number.isFinite(point.lng) && Number.isFinite(point.lat) && point.count > 0)
  const max = Math.max(1, ...data.map((point) => point.count))

  heatmapLayer?.setDataSet({ data, max })
  if (!heatmapCenterMarker) {
    heatmapCenterMarker = new heatmapAMap.Marker({ position: centerPosition, title: '司机当前位置', anchor: 'center', zIndex: 120 })
    heatmapMapInstance.add(heatmapCenterMarker)
  } else {
    heatmapCenterMarker.setPosition(centerPosition)
  }
  heatmapMapInstance.setZoomAndCenter(data.length ? 14 : 13, centerPosition)
}

function startHeatmapRealtime() {
  stopHeatmapRealtime()
  heatmapRefreshTimer = window.setInterval(refreshHeatmap, 15000)
}

function stopHeatmapRealtime() {
  if (heatmapRefreshTimer) window.clearInterval(heatmapRefreshTimer)
  heatmapRefreshTimer = null
}

function destroyHeatmapMap() {
  heatmapLayer?.setDataSet?.({ data: [], max: 1 })
  heatmapMapInstance?.destroy()
  heatmapAMap = null
  heatmapMapInstance = null
  heatmapLayer = null
  heatmapCenterMarker = null
  heatmapGeolocation = null
  heatmapMapReady.value = false
  heatmapMapError.value = ''
}

function startTripRealtime() {
  stopTripRealtime()
  tripTimer = window.setInterval(() => {
    if (!driverStore.currentOrderId) return
    safeApiCall(async () => {
      const res = await getDriverOrderDetail(Number(driverStore.currentOrderId))
      const order = res?.order || res
      if (order && Number(order.orderId || 0) === Number(driverStore.currentOrderId)) {
        syncCurrentOrderFromPoll(order)
      }
      return order
    }, null, { silent: true })
  }, 20000)
}

// 行程中轮询：把最新订单状态落回 currentOrder，感知乘客取消/退款等状态变更
function syncCurrentOrderFromPoll(order) {
  const status = Number(order.status || 0)
  if (status === 6 || status === 7) {
    driverStore.setCurrentOrder(null, 'idle')
    if (driverStore.onlineStatus === 2) driverStore.setWorkState(1)
    showToast(status === 6 ? '订单已取消' : '订单已退款')
    renderHomeMapData()
    return
  }
  const phase = status === 3 ? 'trip' : status === 2 ? 'pickup' : driverStore.tripPhase
  if (phase !== 'idle') {
    driverStore.setCurrentOrder({ ...(driverStore.currentOrder || {}), ...order, orderId: order.orderId }, phase)
    renderHomeMapData()
  }
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
      // 首页空闲状态下直接把新派单显示到浮动面板，避免司机切 tab 才能看到新单
      if (activeTab.value === 0 && !driverStore.currentOrder && driverStore.onlineStatus === 1) {
        previewOrder.value = normalizePushOrder(payload)
      }
      void loadOrders(orderPage.value)
      void loadNearbyOrders(nearbyOrderPage.value, { silentError: true })
      if (nearbyOrderPopupVisible.value) void loadNearbyExpandedOrders(nearbyOrderExpandedPage.value, { silentError: true })
    } else if (payload.type === 'dispatch.new') {
      showToast('收到新的派单')
      void loadOrders(orderPage.value)
      void loadNearbyOrders(nearbyOrderPage.value, { silentError: true })
      if (nearbyOrderPopupVisible.value) void loadNearbyExpandedOrders(nearbyOrderExpandedPage.value, { silentError: true })
    }
  } catch {
    // Push messages are best-effort; malformed messages should not block the H5 page.
  }
}

// 将 WS 推送的 dispatch_order 消息规范化为订单对象（字段与 /orders/available 返回的 OrderBrief 对齐）
function normalizePushOrder(payload) {
  return {
    source: 'available',
    orderId: Number(payload.orderId || payload.orderID || 0),
    orderNo: payload.orderNo || '',
    fromAddress: payload.fromAddress || '',
    toAddress: payload.toAddress || '',
    status: Number(payload.status || 1),
    estimatedPriceCents: Number(payload.estimatedPriceCents || 0),
    createdAt: Number(payload.createdAt || 0)
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
  // 派单记录按派单状态维度查询，不能把订单状态当派单状态传（枚举含义不同），只传分页。
  else if (orderMode.value === 'dispatches') data = await loadDispatches({ page: orderPage.value, pageSize: orderPageSize.value }, config)
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
        orderId: item.dispatch?.orderId || item.order?.orderId,
        dispatchId: item.dispatch?.id,
        dispatchStatus: item.dispatch?.status,
        status: item.order?.status ?? item.dispatch?.status,
        fromAddress: item.order?.fromAddress || item.dispatch?.fromAddress,
        toAddress: item.order?.toAddress || item.dispatch?.toAddress,
        estimatedPriceCents: item.order?.estimatedPriceCents || item.dispatch?.estimatedPriceCents,
        createdAt: item.order?.createdAt || item.dispatch?.createdAt
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

async function loadNearbyOrders(page = nearbyOrderPage.value, config = {}) {
  if (driverStore.onlineStatus !== 1) {
    nearbyOrders.value = []
    nearbyOrderTotal.value = 0
    previewOrder.value = null
    renderHomeMapData()
    return
  }
  nearbyOrderPage.value = Number(page || 1)
  nearbyOrderLoading.value = true
  try {
    const data = await listAvailableOrders({ page: nearbyOrderPage.value, pageSize: nearbyOrderPageSize, status: 1 }, config)
    nearbyOrders.value = Array.isArray(data?.list)
      ? data.list.filter(isWaitAcceptOrder).map((item) => ({ ...item, source: 'available' }))
      : []
    nearbyOrderTotal.value = Number(data?.total || nearbyOrders.value.length || 0)
    renderHomeMapData()
  } finally {
    nearbyOrderLoading.value = false
  }
}

async function loadNearbyExpandedOrders(page = nearbyOrderExpandedPage.value, config = {}) {
  if (driverStore.onlineStatus !== 1) {
    nearbyOrderExpandedOrders.value = []
    nearbyOrderExpandedTotal.value = 0
    return
  }
  nearbyOrderExpandedPage.value = Number(page || 1)
  nearbyOrderExpandedLoading.value = true
  try {
    const data = await listAvailableOrders({ page: nearbyOrderExpandedPage.value, pageSize: nearbyOrderExpandedPageSize, status: 1 }, config)
    nearbyOrderExpandedOrders.value = Array.isArray(data?.list)
      ? data.list.filter(isWaitAcceptOrder).map((item) => ({ ...item, source: 'available' }))
      : []
    nearbyOrderExpandedTotal.value = Number(data?.total || nearbyOrderExpandedOrders.value.length || 0)
  } finally {
    nearbyOrderExpandedLoading.value = false
  }
}

async function openNearbyOrderPopup() {
  nearbyOrderPopupVisible.value = true
  await loadNearbyExpandedOrders(1, { silentError: true })
}

function syncCurrentTripFromOrders(list) {
  const current = list.find((item) => String(resolveOrderId(item)) === String(driverStore.currentOrderId)) || list.find((item) => [2, 3].includes(Number(item.status)))
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
  const orderId = resolveOrderId(order)
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
    const actionResult = await config.request()
    const nextOrder = mergeOrderActionResult(order, actionResult, config.phase)
    driverStore.setCurrentOrder(config.phase === 'idle' ? null : nextOrder, config.phase)
  } catch (error) {
    if (error?.message === 'cancelled') return
    showToast(apiErrorMessage(error, '订单操作失败'))
    await loadNearbyOrders(nearbyOrderPage.value, { silentError: true })
    if (nearbyOrderPopupVisible.value) await loadNearbyExpandedOrders(nearbyOrderExpandedPage.value, { silentError: true })
    return
  }

  previewOrder.value = null
  if (config.phase === 'trip') driverStore.setWorkState(2)
  if (config.phase === 'idle' && driverStore.onlineStatus === 2) driverStore.setWorkState(1)
  showToast(config.message)
  await loadOrders(orderPage.value)
  await loadNearbyOrders(nearbyOrderPage.value, { silentError: true })
  if (nearbyOrderPopupVisible.value) await loadNearbyExpandedOrders(nearbyOrderExpandedPage.value, { silentError: true })
  renderHomeMapData()
}

function resolveOrderId(order) {
  return Number(order?.orderId || order?.orderID || order?.id || order?.dispatch?.orderId || order?.order?.orderId || 0)
}

function mergeOrderActionResult(order, result, phase) {
  if (phase === 'idle') return null
  const orderId = Number(result?.orderId || result?.orderID || result?.id || resolveOrderId(order))
  const status = Number(result?.status || (phase === 'trip' ? 3 : phase === 'pickup' ? 2 : order?.status || 0))
  return {
    ...(order || {}),
    orderId,
    status,
    source: 'order',
    dispatchStatus: undefined
  }
}

function isWaitAcceptOrder(order) {
  return Number(order?.status || 0) === 1
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

function previewHomeOrder(order) {
  previewOrder.value = order || null
  void loadOrderDetail(resolveOrderId(order))
}

function navigateToPickup(order) {
  const position = orderPickupPosition(order)
  if (!position) {
    showToast('暂无上车点坐标')
    return
  }
  const name = encodeURIComponent(order?.fromAddress || '上车点')
  const url = `https://uri.amap.com/navigation?to=${position[0]},${position[1]},${name}&mode=car&policy=1&coordinate=gaode`
  window.open(url, '_blank')
}

function navigateHomeRoute() {
  if (!homeRouteAddress.value) {
    const value = window.prompt('请输入回家顺路目的地', '')
    if (!value) return
    homeRouteAddress.value = value.trim()
    localStorage.setItem('driverHomeRouteAddress', homeRouteAddress.value)
  }
  showToast('已开启回家顺路 ' + homeRouteAddress.value)
}

function contactPassenger(order) {
  const phone = order?.passengerPhone || order?.userPhone || order?.phone || ''
  if (!phone) {
    showToast('暂无乘客联系方式')
    return
  }
  window.location.href = 'tel:' + phone
}

function reportAbnormal() {
  showDialog({
    title: '异常上报',
    message: '已记录异常入口，请联系平台客服处理当前行程。'
  }).catch(() => {})
}

function passengerTail(order) {
  const phone = String(order?.passengerPhone || order?.userPhone || order?.phone || '')
  return phone ? '尾号' + phone.slice(-4) : '乘客'
}

function orderPickupPosition(order) {
  const lng = Number(order?.fromLongitude || order?.pickupLongitude || order?.longitude)
  const lat = Number(order?.fromLatitude || order?.pickupLatitude || order?.latitude)
  return Number.isFinite(lng) && Number.isFinite(lat) && (lng !== 0 || lat !== 0) ? [lng, lat] : null
}

function orderDropoffPosition(order) {
  const lng = Number(order?.toLongitude || order?.dropoffLongitude)
  const lat = Number(order?.toLatitude || order?.dropoffLatitude)
  return Number.isFinite(lng) && Number.isFinite(lat) && (lng !== 0 || lat !== 0) ? [lng, lat] : null
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

async function loadTrajectory() {
  const orderId = Number(trajectoryOrderId.value || driverStore.currentOrderId || 0)
  if (!Number.isFinite(orderId) || orderId <= 0) {
    trajectoryError.value = '请输入订单ID'
    showToast('请输入订单ID')
    return
  }

  trajectoryOrderId.value = orderId
  try {
    trajectoryError.value = ''
    const res = await getOrderTrajectory({ orderId }, { silentError: true })
    trajectoryPoints.value = Array.isArray(res.points) ? res.points : []
  } catch (error) {
    trajectoryPoints.value = []
    trajectoryError.value = apiErrorMessage(error, '轨迹查询失败')
  }
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

function formatDuration(seconds) {
  const value = Number(seconds)
  if (!Number.isFinite(value) || value <= 0) return '--'
  const minutes = Math.max(1, Math.round(value / 60))
  return minutes + '分钟'
}

function formatTime(timestamp) {
  const value = Number(timestamp)
  if (!value) return '--'
  return new Date(value * 1000).toLocaleString('zh-CN', { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })
}

function formatOrderStatus(status) {
  return { 1: '待接单', 2: '已接单', 3: '行程中', 4: '待支付', 5: '已完成', 6: '已取消', 7: '已退款' }[Number(status || 0)] || '--'
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
.home-workbench { min-height: calc(100vh - 86px); margin: 0 -12px; padding: 10px 12px 172px; background: #f6f7fb; }
.driver-status-bar { display: grid; grid-template-columns: 44px minmax(0, 1fr) auto auto auto 38px; gap: 8px; align-items: center; min-height: 58px; padding: 8px 10px; border-radius: 8px; background: #fff; box-shadow: 0 8px 20px rgba(15,23,42,.08); }
.driver-status-bar .driver-avatar-button { width: 42px; height: 42px; flex-basis: 42px; }
.driver-status-bar .driver-avatar-button img, .driver-status-bar .avatar-fallback { width: 42px; height: 42px; flex-basis: 42px; border-color: #e6eaf2; color: #5B5CFF; background: #eef2ff; font-size: 18px; }
.status-copy { display: grid; gap: 2px; min-width: 0; }
.status-copy strong { overflow: hidden; color: #172033; font-size: 15px; line-height: 1.2; text-overflow: ellipsis; white-space: nowrap; }
.status-copy span, .status-metric span { color: #7a8496; font-size: 11px; line-height: 1.2; }
.status-metric { display: grid; gap: 2px; min-width: 52px; text-align: center; }
.status-metric b { color: #172033; font-size: 13px; line-height: 1.2; white-space: nowrap; }
.status-toggle { min-width: 48px; min-height: 30px; padding: 0 10px; border: 0; border-radius: 999px; background: #eef2ff; color: #5B5CFF; font-size: 12px; font-weight: 800; }
.status-toggle.online { background: #5B5CFF; color: #fff; box-shadow: 0 6px 14px rgba(91,92,255,.24); }
.status-toggle.driving { background: #ffb72c; color: #fff; box-shadow: 0 6px 14px rgba(255,183,44,.26); }
.message-button { display: grid; width: 34px; height: 34px; place-items: center; border: 0; border-radius: 50%; background: #fff7e6; color: #f59e0b; font-size: 18px; }
.home-map-stage { position: relative; height: calc(100vh - 180px); min-height: 460px; margin: 10px -12px 0; overflow: hidden; background: #dfe8f3; }
.home-amap { position: absolute; inset: 0; width: 100%; height: 100%; }
.heatmap-legend { position: absolute; left: 12px; top: 12px; z-index: 5; display: inline-flex; align-items: center; gap: 5px; padding: 7px 9px; border-radius: 999px; background: rgba(255,255,255,.94); color: #667085; font-size: 11px; font-weight: 800; box-shadow: 0 6px 16px rgba(15,23,42,.14); }
.heatmap-legend i { width: 18px; height: 8px; border-radius: 999px; }
.heat-low { background: #3b82f6; }
.heat-mid { background: #22c55e; }
.heat-high { background: #ef4444; }
.route-direction-chip { position: absolute; left: 12px; right: 12px; bottom: 160px; z-index: 5; display: flex; align-items: center; gap: 8px; min-height: 40px; padding: 9px 11px; border-radius: 8px; background: rgba(255,255,255,.96); color: #172033; font-size: 12px; font-weight: 800; box-shadow: 0 8px 20px rgba(15,23,42,.14); }
.route-direction-chip span { min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.map-floating-actions { position: absolute; right: 12px; top: 58px; z-index: 5; display: grid; gap: 8px; }
.map-floating-actions button { display: grid; width: 38px; height: 38px; place-items: center; border: 0; border-radius: 50%; background: rgba(255,255,255,.96); color: #2563eb; font-size: 18px; box-shadow: 0 6px 16px rgba(15,23,42,.16); }
.home-floating-panel { position: fixed; left: 50%; bottom: calc(70px + env(safe-area-inset-bottom)); z-index: 30; display: grid; gap: 10px; width: min(calc(100vw - 24px), 406px); padding: 14px; border-radius: 8px; background: rgba(255,255,255,.98); box-shadow: 0 -8px 28px rgba(15,23,42,.16); transform: translateX(-50%); }
.panel-heading { display: flex; align-items: flex-start; justify-content: space-between; gap: 10px; }
.panel-heading span { min-width: 0; color: #7a8496; font-size: 12px; line-height: 1.35; overflow-wrap: anywhere; }
.panel-heading strong { color: #172033; font-size: 20px; line-height: 1.2; text-align: right; overflow-wrap: anywhere; }
.order-brief-grid { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 8px; }
.order-brief-grid span { min-height: 30px; padding: 7px 8px; border-radius: 8px; background: #f2f5fa; color: #344054; text-align: center; font-size: 12px; font-weight: 800; }
.panel-actions, .idle-controls { display: grid; gap: 8px; }
.three-actions { grid-template-columns: 1.1fr .9fr .9fr; }
.panel-actions button, .home-route-button { min-height: 42px; border: 1px solid #d7dce5; border-radius: 999px; background: #fff; color: #344054; font-size: 14px; font-weight: 800; }
.panel-actions button.primary { border-color: #5B5CFF; background: #5B5CFF; color: #fff; }
.idle-controls { grid-template-columns: minmax(0, 1fr) 92px; align-items: center; }
.listen-switch { display: grid; justify-items: center; gap: 6px; color: #667085; font-size: 12px; font-weight: 800; }
.home-route-button { display: flex; align-items: center; justify-content: center; gap: 6px; width: 100%; }
.driver-avatar-button { width: 52px; height: 52px; flex: 0 0 52px; padding: 0; border: 0; border-radius: 50%; background: transparent; }
.driver-avatar-button img, .avatar-fallback { width: 52px; height: 52px; flex: 0 0 52px; border: 2px solid rgba(255,255,255,.72); border-radius: 50%; background: rgba(255,255,255,.22); object-fit: cover; }
.avatar-fallback { display: grid; place-items: center; color: #fff; font-size: 22px; font-weight: 800; }
.work-status-hint { margin: 0; line-height: 1.5; }
.work-primary-action { display: flex; align-items: center; justify-content: center; gap: 8px; width: 100%; min-height: 56px; border: 0; border-radius: 999px; font-size: 17px; font-weight: 800; }
.work-primary-action .van-icon { font-size: 21px; }
.go-online { border: 0; background: #5B5CFF; color: #fff; box-shadow: 0 8px 18px rgba(91,92,255,.28); }
.finish-panel, .withdraw-panel, .heatmap-panel { padding: 18px 14px 22px; background: #f6f7fb; }
.finish-panel h2, .withdraw-panel h2 { margin: 0 0 14px; text-align: center; }
.driver-heatmap-popup { left: 0; right: 0; width: min(100vw, 390px); margin: 0 auto; overflow: hidden; }
.heatmap-panel { display: grid; grid-template-rows: auto auto minmax(0, 1fr) auto; height: 100%; padding: 8px 12px calc(12px + env(safe-area-inset-bottom)); background: #f6f7fb; }
.heatmap-sheet-grabber { width: 42px; height: 4px; margin: 0 auto 10px; border-radius: 999px; background: #cbd5e1; }
.heatmap-heading { display: flex; align-items: flex-start; justify-content: space-between; gap: 12px; margin-bottom: 12px; }
.heatmap-sheet-header { min-height: 44px; align-items: center; }
.heatmap-heading h2 { margin: 0 0 4px; color: #172033; font-size: 18px; }
.heatmap-heading p { margin: 0; color: #7a8496; font-size: 12px; }
.heatmap-floating-actions { position: absolute; top: 12px; right: 12px; display: grid; gap: 8px; z-index: 2; }
.heatmap-refresh { display: grid; width: 40px; height: 40px; flex: 0 0 40px; place-items: center; border: 0; border-radius: 50%; background: #fff; color: #5B5CFF; font-size: 18px; box-shadow: 0 4px 14px rgba(15,23,42,.18); }
.heatmap-map-shell { position: relative; min-height: 0; overflow: hidden; border-radius: 8px; background: #e8edf5; }
.driver-heatmap-map { width: 100%; height: 100%; min-height: 0; }
.heatmap-badge { position: absolute; right: 12px; bottom: 12px; display: grid; gap: 2px; min-width: 78px; padding: 8px 10px; border-radius: 8px; background: rgba(255,255,255,.94); color: #172033; box-shadow: 0 6px 18px rgba(15,23,42,.14); }
.heatmap-badge span { color: #7a8496; font-size: 11px; }
.heatmap-badge strong { font-size: 16px; line-height: 1.1; }
.heatmap-chip-strip { display: flex; gap: 8px; margin: 10px -12px 0; padding: 0 12px 2px; overflow-x: auto; scrollbar-width: none; }
.heatmap-chip-strip::-webkit-scrollbar { display: none; }
.heatmap-chip { display: inline-flex; min-width: 132px; align-items: center; gap: 8px; padding: 9px 10px; border-radius: 8px; background: #fff; box-shadow: 0 4px 14px rgba(15,23,42,.06); }
.heatmap-chip > span { display: grid; width: 32px; height: 32px; flex: 0 0 32px; place-items: center; border-radius: 50%; background: #fff4e5; color: #f59e0b; font-size: 18px; }
.heatmap-chip div { display: grid; gap: 2px; min-width: 0; }
.heatmap-chip strong { color: #172033; font-size: 15px; line-height: 1.1; white-space: nowrap; }
.heatmap-chip small { max-width: 76px; overflow: hidden; color: #7a8496; font-size: 10px; text-overflow: ellipsis; white-space: nowrap; }
.driver-tabbar { left: 50%; right: auto; width: min(100vw, 430px); height: calc(60px + env(safe-area-inset-bottom)); border-top: 1px solid #e6eaf2; box-shadow: 0 -8px 24px rgba(15,23,42,.08); transform: translateX(-50%); --van-tabbar-item-active-color: #5B5CFF; --van-tabbar-item-text-color: #98a2b3; }
button:disabled { opacity: .48; }
@media (max-width: 360px) { .driver-home-page { padding-inline: 10px; } .income-today-card strong { font-size: 30px; } .income-today-card { align-items: flex-start; } .withdraw-entry { min-width: 68px; padding-inline: 10px; } }
</style>
