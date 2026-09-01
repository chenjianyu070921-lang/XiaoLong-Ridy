<template>
  <section class="h5-panel">
    <div class="section-title">
      <h2>工作台</h2>
      <button type="button" @click="$emit('refresh-dashboard')">刷新</button>
    </div>
    <div class="current-trip-card">
      <div class="section-subtitle">
        <strong>当前行程</strong>
        <van-tag type="primary">{{ formatOrderStatus(driverStore.currentOrder?.status) }}</van-tag>
      </div>
      <p><b>上车点</b><span>{{ driverStore.currentOrder?.fromAddress || '--' }}</span></p>
      <p><b>目的地</b><span>{{ driverStore.currentOrder?.toAddress || '--' }}</span></p>
      <p><b>订单号</b><span>{{ driverStore.currentOrder?.orderNo || driverStore.currentOrderId || '--' }}</span></p>
    </div>
    <div class="map-surface live-map-surface">
      <div ref="mapContainer" class="driver-live-map" aria-label="实时接单地图"></div>
      <div v-if="mapStatusText" class="map-state" :class="{ error: mapError }">{{ mapStatusText }}</div>
      <button type="button" class="map-recenter" aria-label="回到司机当前位置" @click="centerOnDriver">
        <van-icon name="aim" />
      </button>
      <div class="map-badge">{{ locationLabel }}</div>
    </div>
    <section class="home-order-section">
      <div class="section-title">
        <h2>附近可接订单</h2>
        <button type="button" :disabled="homeAvailableLoading" @click="$emit('refresh-home-orders')">刷新</button>
      </div>
      <div v-if="homeAvailableLoading" class="home-order-loading"><van-loading size="20px" /></div>
      <div v-else-if="homeAvailableOrders.length === 0" class="home-order-empty">
        {{ driverStore.onlineStatus === 1 ? '暂无附近订单' : '开始听单后查看附近订单' }}
      </div>
      <article v-for="order in homeAvailableOrders" :key="'home-' + order.orderId" class="home-order-card">
        <div class="order-heading">
          <strong>{{ order.orderNo || '订单 ' + order.orderId }}</strong>
          <span class="order-distance">距您{{ formatDistance(order.distanceMeters) }}公里</span>
        </div>
        <p class="route-line">{{ order.fromAddress || '--' }} -> {{ order.toAddress || '--' }}</p>
        <div class="meta-row">
          <span>{{ formatPrice(order.estimatedPriceCents) }}</span>
          <button type="button" class="home-order-accept" @click="$emit('order-action', 'accept', order)">接单</button>
        </div>
      </article>
    </section>
  </section>
</template>

<script setup>
import { computed, nextTick, onMounted, onUnmounted, ref } from 'vue'
import AMapLoader from '@amap/amap-jsapi-loader'
import { getAmapConfig } from '@/config/amap'
import { normalizeBrowserLocationForAmap } from '@/utils/geo'

defineProps({
  driverStore: { type: Object, required: true },
  homeAvailableOrders: { type: Array, required: true },
  homeAvailableLoading: { type: Boolean, required: true },
  formatPrice: { type: Function, required: true },
  formatDistance: { type: Function, required: true },
  formatOrderStatus: { type: Function, required: true }
})

defineEmits(['refresh-dashboard', 'refresh-home-orders', 'order-action'])

const fallbackCenter = [116.397428, 39.90923]
const mapContainer = ref(null)
const mapReady = ref(false)
const mapError = ref('')
const locating = ref(true)
const driverPosition = ref(null)
const livePath = []

let AMapSDK = null
let mapInstance = null
let driverMarker = null
let livePolyline = null
let geoWatchId = null

const mapStatusText = computed(() => {
  if (mapError.value) return mapError.value
  if (!mapReady.value) return '地图加载中...'
  if (locating.value) return '正在获取实时位置...'
  return ''
})

const locationLabel = computed(() => driverPosition.value ? '实时位置已更新' : '等待定位')

onMounted(async () => {
  await nextTick()
  await initMap()
  if (mapReady.value) startBrowserLocation()
})

onUnmounted(() => {
  if (geoWatchId && navigator.geolocation) navigator.geolocation.clearWatch(geoWatchId)
  geoWatchId = null
  mapInstance?.destroy()
  mapInstance = null
})

async function initMap() {
  const { key, securityCode } = getAmapConfig()
  if (!key) {
    mapError.value = '未配置高德地图 Key'
    locating.value = false
    return
  }
  try {
    if (securityCode) window._AMapSecurityConfig = { securityJsCode: securityCode }
    AMapSDK = await AMapLoader.load({ key, version: '2.0' })
    window.AMap = AMapSDK
    mapInstance = new AMapSDK.Map(mapContainer.value, {
      zoom: 15,
      viewMode: '2D',
      center: fallbackCenter
    })
    mapReady.value = true
  } catch (error) {
    mapError.value = '高德地图加载失败'
    locating.value = false
    console.error('driver workbench AMap error:', error)
  }
}

function startBrowserLocation() {
  if (!navigator.geolocation) {
    mapError.value = '当前浏览器不支持定位'
    locating.value = false
    return
  }
  geoWatchId = navigator.geolocation.watchPosition(
    (position) => updateDriverPosition(position.coords.longitude, position.coords.latitude),
    () => {
      mapError.value = '定位不可用，请检查权限'
      locating.value = false
    },
    { enableHighAccuracy: true, maximumAge: 5000, timeout: 8000 }
  )
}

function updateDriverPosition(longitude, latitude) {
  const { longitude: lng, latitude: lat } = normalizeBrowserLocationForAmap({ longitude, latitude })
  if (!Number.isFinite(lng) || !Number.isFinite(lat) || (lng === 0 && lat === 0)) return
  const position = [lng, lat]
  driverPosition.value = position
  locating.value = false
  if (!mapInstance || !AMapSDK) return
  mapError.value = ''

  if (!driverMarker) {
    driverMarker = new AMapSDK.Marker({ position, title: '司机当前位置', anchor: 'center', zIndex: 120 })
    mapInstance.add(driverMarker)
  } else {
    driverMarker.setPosition(position)
  }

  livePath.push(position)
  if (livePath.length > 80) livePath.shift()
  if (livePath.length > 1) {
    if (!livePolyline) {
      livePolyline = new AMapSDK.Polyline({
        path: livePath,
        strokeColor: '#2563EB',
        strokeWeight: 6,
        strokeOpacity: 0.86,
        showDir: true
      })
      mapInstance.add(livePolyline)
    } else {
      livePolyline.setPath(livePath)
    }
  }
  mapInstance.setCenter(position)
}

function centerOnDriver() {
  if (mapInstance && driverPosition.value) mapInstance.setZoomAndCenter(16, driverPosition.value)
}
</script>
