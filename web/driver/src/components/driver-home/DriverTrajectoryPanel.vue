<template>
  <section class="h5-panel">
    <div class="section-title"><h2>轨迹</h2></div>
    <div class="trajectory-map-surface">
      <div ref="mapContainer" class="driver-live-map" aria-label="订单实时轨迹地图"></div>
      <div v-if="mapStatusText" class="map-state" :class="{ error: mapError || trajectoryError }">{{ mapStatusText }}</div>
    </div>
    <div class="trajectory-form">
      <van-field v-model.number="trajectoryOrderIdModel" type="number" label="订单ID" placeholder="输入订单ID" />
      <p v-if="trajectoryError" class="trajectory-error">{{ trajectoryError }}</p>
      <button class="primary-action" type="button" @click="$emit('load-trajectory')">查询轨迹</button>
    </div>
    <div v-if="trajectoryPoints.length === 0" class="empty-state">--</div>
    <div v-else class="trajectory-list">
      <article v-for="point in trajectoryPoints" :key="point.id || point.reportTime || point.createdAt" class="compact-card">
        <strong>{{ point.latitude || '--' }}, {{ point.longitude || '--' }}</strong>
        <span>速度 {{ point.speedKmh ?? '--' }} km/h · {{ formatTime(point.reportTime || point.createdAt) }}</span>
      </article>
    </div>
  </section>
</template>

<script setup>
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import AMapLoader from '@amap/amap-jsapi-loader'
import { getAmapConfig } from '@/config/amap'

const props = defineProps({
  trajectoryOrderId: { type: [String, Number], required: true },
  trajectoryError: { type: String, required: true },
  trajectoryPoints: { type: Array, required: true },
  formatTime: { type: Function, required: true },
  refreshIntervalMs: { type: Number, default: 5000 }
})

const emit = defineEmits(['update:trajectoryOrderId', 'load-trajectory'])

const fallbackCenter = [116.397428, 39.90923]
const mapContainer = ref(null)
const mapReady = ref(false)
const mapError = ref('')

let AMapSDK = null
let mapInstance = null
let trajectoryPolyline = null
let latestMarker = null
let refreshTimer = null

const trajectoryOrderIdModel = computed({
  get: () => props.trajectoryOrderId,
  set: (value) => emit('update:trajectoryOrderId', value)
})

const trajectoryPath = computed(() => props.trajectoryPoints.map(readPointPosition).filter(Boolean))

const mapStatusText = computed(() => {
  if (mapError.value) return mapError.value
  if (props.trajectoryError) return props.trajectoryError
  if (!mapReady.value) return '地图加载中...'
  if (trajectoryPath.value.length === 0) return '暂无轨迹点'
  return ''
})

onMounted(async () => {
  await nextTick()
  await initMap()
  renderTrajectory()
  startAutoRefresh()
})

onUnmounted(() => {
  stopAutoRefresh()
  mapInstance?.destroy()
  mapInstance = null
})

watch(() => props.trajectoryPoints, () => renderTrajectory(), { deep: true })
watch(() => props.trajectoryOrderId, () => startAutoRefresh())

async function initMap() {
  const { key, securityCode } = getAmapConfig()
  if (!key) {
    mapError.value = '未配置高德地图 Key'
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
    console.error('driver trajectory AMap error:', error)
  }
}

function startAutoRefresh() {
  stopAutoRefresh()
  if (!Number(props.trajectoryOrderId || 0)) return
  refreshTimer = window.setInterval(() => emit('load-trajectory'), props.refreshIntervalMs)
}

function stopAutoRefresh() {
  if (refreshTimer) window.clearInterval(refreshTimer)
  refreshTimer = null
}

function renderTrajectory() {
  if (!mapInstance || !AMapSDK) return
  const path = trajectoryPath.value
  if (path.length === 0) {
    if (trajectoryPolyline) trajectoryPolyline.setPath([])
    if (latestMarker) {
      mapInstance.remove(latestMarker)
      latestMarker = null
    }
    return
  }

  if (path.length > 1) {
    if (!trajectoryPolyline) {
      trajectoryPolyline = new AMapSDK.Polyline({
        path,
        strokeColor: '#7C3AED',
        strokeWeight: 7,
        strokeOpacity: 0.88,
        showDir: true
      })
      mapInstance.add(trajectoryPolyline)
    } else {
      trajectoryPolyline.setPath(path)
    }
  } else if (trajectoryPolyline) {
    trajectoryPolyline.setPath([])
  }

  const latestPosition = path[path.length - 1]
  if (!latestMarker) {
    latestMarker = new AMapSDK.Marker({ position: latestPosition, title: '最新位置', anchor: 'center', zIndex: 120 })
    mapInstance.add(latestMarker)
  } else {
    latestMarker.setPosition(latestPosition)
  }

  if (path.length > 1) {
    mapInstance.setFitView([trajectoryPolyline, latestMarker], false, [48, 24, 48, 24])
  } else {
    mapInstance.setZoomAndCenter(16, latestPosition)
  }
}

function readPointPosition(point) {
  const lng = Number(point?.longitude)
  const lat = Number(point?.latitude)
  if (!Number.isFinite(lng) || !Number.isFinite(lat) || (lng === 0 && lat === 0)) return null
  return [lng, lat]
}
</script>
