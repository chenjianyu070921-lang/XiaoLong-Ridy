<template>
  <div class="map-picker">
    <div id="picker-map" class="picker-map" aria-label="地图选点">
      <div v-if="mapLoading" class="map-state"><van-loading size="24px" vertical>地图加载中...</van-loading></div>
      <div v-else-if="mapError" class="map-state map-error">
        <p>{{ mapError }}</p>
        <button type="button" class="retry-btn" @click="initMap">重新加载</button>
      </div>
    </div>

    <!-- 中心图钉固定在地图可视区中心，拖动地图时由中心坐标决定选中位置。 -->
    <div class="center-pin" aria-hidden="true">
      <span class="pin-halo" :class="type"></span>
      <span class="pin-dot" :class="type"></span>
      <span class="pin-shadow"></span>
    </div>

    <div class="picker-header">
      <button type="button" class="header-btn" aria-label="返回" @click="$emit('back')">
        <van-icon name="arrow-left" size="20" />
      </button>
      <span class="header-title">{{ type === 'pickup' ? '选择上车点' : '选择目的地' }}</span>
      <span class="header-spacer"></span>
    </div>

    <div class="search-layer">
      <LocationSearch
        :type="type"
        :city="city"
        :cityCode="cityCode"
        :center="center"
        @select="handleSearchSelect"
      />
    </div>

    <CurrentLocationButton class="locate-fab" :loading="locating" @locate="locateUser" />

    <div class="picker-panel">
      <LocationPanel :type="type" :location="currentLocation" :resolving="resolving" @confirm="confirm" />
    </div>
  </div>
</template>

<script setup>
import { onBeforeUnmount, onMounted, ref } from 'vue'
import AMapLoader from '@amap/amap-jsapi-loader'
import { showToast } from 'vant'
import { getAmapConfig } from '@/config/amap'
import { reverseGeocodeLocation } from '@/api/location'
import { createLocation, hasValidCoordinate } from '@/utils/location'
import LocationSearch from './LocationSearch.vue'
import LocationPanel from './LocationPanel.vue'
import CurrentLocationButton from './CurrentLocationButton.vue'

const props = defineProps({
  type: { type: String, default: 'destination' },
  city: { type: String, default: '' },
  cityCode: { type: String, default: '' },
  initialLocation: { type: Object, default: null }
})

const emit = defineEmits(['confirm', 'back'])

const AMapSDK = ref(null)
const mapInstance = ref(null)
const geolocation = ref(null)
const citySearch = ref(null)
const mapLoading = ref(true)
const mapError = ref('')
const locating = ref(false)
const resolving = ref(false)
// 当前选中的地点，由拖动逆地理或搜索选择得到。
const currentLocation = ref(null)
// 当前中心坐标，供搜索组件按中心做距离排序。
const center = ref({ latitude: 0, longitude: 0 })
// 搜索选择的地点，命中时不再用逆地理覆盖其名称。
const selectedLocation = ref(null)
let resolveTimer = null
let resolveSequence = 0

function readLngLat(position) {
  if (!position) return null
  const longitude = typeof position.getLng === 'function' ? position.getLng() : Number(position.lng)
  const latitude = typeof position.getLat === 'function' ? position.getLat() : Number(position.lat)
  return hasValidCoordinate(latitude, longitude) ? { latitude, longitude } : null
}

async function initMap() {
  if (mapInstance.value) return
  mapLoading.value = true
  mapError.value = ''
  const { key, securityCode } = getAmapConfig()
  if (!key) {
    mapLoading.value = false
    mapError.value = '未配置高德地图 Key，无法使用地图选点'
    return
  }
  try {
    if (securityCode) window._AMapSecurityConfig = { securityJsCode: securityCode }
    const AMap = await AMapLoader.load({ key, version: '2.0', plugins: ['AMap.Geolocation', 'AMap.CitySearch'] })
    AMapSDK.value = AMap
    geolocation.value = new AMap.Geolocation({ enableHighAccuracy: true, timeout: 10000 })
    citySearch.value = new AMap.CitySearch()
    // 调用方传入的 initialLocation 是统一 Location 结构，坐标字段为 latitude/longitude。
    const initial = props.initialLocation && hasValidCoordinate(props.initialLocation.latitude, props.initialLocation.longitude)
      ? { latitude: Number(props.initialLocation.latitude), longitude: Number(props.initialLocation.longitude) }
      : null
    mapInstance.value = new AMap.Map('picker-map', {
      zoom: 16,
      viewMode: '2D',
      center: initial ? [initial.longitude, initial.latitude] : undefined
    })
    mapInstance.value.on('moveend', scheduleResolve)
    mapInstance.value.on('zoomend', scheduleResolve)
    // 点击地图等同把该点移到屏幕中心，与拖动选点复用同一条解析链路。
    mapInstance.value.on('click', event => {
      const point = readLngLat(event?.lnglat)
      if (point) mapInstance.value.setCenter([point.longitude, point.latitude])
    })
    mapLoading.value = false
    if (initial) {
      await resolveCenter()
    } else {
      await locateUser()
    }
  } catch (error) {
    console.error('地图选点页初始化失败:', error)
    mapLoading.value = false
    mapError.value = '地图加载失败，请检查网络或地图配置'
  }
}

// 三级定位降级：高德精确定位 → 浏览器 GPS → 城市中心，任一层失败都不影响地图继续使用。
async function locateUser() {
  if (locating.value || !mapInstance.value) return
  locating.value = true
  let point = null
  try {
    point = await locateByAMap()
  } catch (error) {
    console.warn('高德定位失败，尝试浏览器定位:', error)
    try { point = await locateByBrowser() } catch (e2) {
      console.warn('浏览器定位失败，尝试城市中心:', e2)
      try { point = await locateByCity() } catch (e3) {
        console.error('所有定位方式均失败:', e3)
      }
    }
  } finally {
    locating.value = false
  }
  if (!point) {
    showToast('定位失败，请手动拖动地图选点')
    return
  }
  mapInstance.value.setZoomAndCenter(16, [point.longitude, point.latitude])
}

function locateByAMap() {
  return new Promise((resolve, reject) => {
    if (!geolocation.value) return reject(new Error('高德定位服务未初始化'))
    geolocation.value.getCurrentPosition((status, result) => {
      const point = readLngLat(result?.position)
      if (status === 'complete' && point) return resolve(point)
      reject(result || new Error('高德定位失败'))
    })
  })
}

function locateByBrowser() {
  return new Promise((resolve, reject) => {
    if (!navigator.geolocation) return reject(new Error('浏览器不支持定位'))
    navigator.geolocation.getCurrentPosition(
      position => resolve({ latitude: position.coords.latitude, longitude: position.coords.longitude }),
      error => reject(error),
      { enableHighAccuracy: true, timeout: 8000, maximumAge: 15000 }
    )
  })
}

function locateByCity() {
  return new Promise((resolve, reject) => {
    if (!citySearch.value) return reject(new Error('城市定位服务未初始化'))
    citySearch.value.getCity((status, result) => {
      const point = readLngLat(result?.center)
      if (status === 'complete' && point) return resolve(point)
      reject(result || new Error('城市定位失败'))
    })
  })
}

// 地图停止移动后 300ms 触发，避免拖动过程频繁请求逆地理。
function scheduleResolve() {
  clearTimeout(resolveTimer)
  resolveTimer = setTimeout(() => { void resolveCenter() }, 300)
}

async function resolveCenter() {
  const point = readLngLat(mapInstance.value?.getCenter())
  if (!point) return
  center.value = point
  const sequence = ++resolveSequence
  resolving.value = true
  try {
    // 搜索选择后中心未变时，保留用户选择的名称与地址，避免被逆地理覆盖。
    const selected = selectedLocation.value
    const samePoint = selected && Math.abs(selected.latitude - point.latitude) < 1e-6 && Math.abs(selected.longitude - point.longitude) < 1e-6
    if (samePoint) {
      currentLocation.value = selected
      return
    }
    selectedLocation.value = null
    currentLocation.value = await reverseAddress(point)
  } finally {
    if (sequence === resolveSequence) resolving.value = false
  }
}

// 逆地理走后端代理，浏览器不持有高德 Web 服务 Key；失败时保留可确认的兜底地点。
async function reverseAddress(point) {
  try {
    const result = await reverseGeocodeLocation({ longitude: Number(point.longitude), latitude: Number(point.latitude) })
    return createLocation({
      name: result?.poiName || result?.address || '地图选点位置',
      address: result?.address || '',
      province: result?.province || '',
      city: result?.cityName || props.city,
      district: result?.district || '',
      latitude: point.latitude,
      longitude: point.longitude
    })
  } catch (error) {
    console.warn('逆地理编码失败:', error)
    showToast('网络异常，请重试')
    return createLocation({ name: '地图选点位置', address: '', city: props.city, latitude: point.latitude, longitude: point.longitude })
  }
}

function handleSearchSelect(item) {
  selectedLocation.value = item
  currentLocation.value = item
  center.value = { latitude: item.latitude, longitude: item.longitude }
  mapInstance.value?.setZoomAndCenter(17, [item.longitude, item.latitude])
}

function confirm() {
  if (!currentLocation.value) return
  emit('confirm', currentLocation.value)
}

onMounted(initMap)

onBeforeUnmount(() => {
  clearTimeout(resolveTimer)
  resolveSequence += 1
  mapInstance.value?.destroy()
  mapInstance.value = null
})
</script>

<style scoped>
.map-picker {
  --panel-height: 46vh;
  position: relative;
  min-height: 100vh;
  background: #EEF2F7;
}

.picker-map {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: #E5E7EB;
}

.map-state {
  position: absolute;
  inset: 0;
  z-index: 6;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-direction: column;
  gap: 12px;
  background: rgba(255, 255, 255, 0.9);
  color: #6B7280;
  font-size: 14px;
}

.map-error { color: #DC2626; }

.retry-btn {
  padding: 8px 18px;
  border: 0;
  border-radius: 18px;
  background: #7C3AED;
  color: #FFFFFF;
  font-size: 14px;
  cursor: pointer;
}

.center-pin {
  position: absolute;
  left: 50%;
  top: calc((100vh - var(--panel-height)) / 2);
  z-index: 7;
  display: flex;
  flex-direction: column;
  align-items: center;
  transform: translate(-50%, -50%);
  pointer-events: none;
}

.pin-halo {
  position: absolute;
  top: -16px;
  width: 46px;
  height: 46px;
  border-radius: 50%;
  opacity: 0.18;
  animation: pin-halo 2s ease-in-out infinite;
}

.pin-halo.pickup { background: #10B981; }
.pin-halo.destination { background: #F59E0B; }

.pin-dot {
  width: 18px;
  height: 18px;
  border: 3px solid #FFFFFF;
  border-radius: 50%;
  box-shadow: 0 4px 12px rgba(15, 23, 42, 0.28);
  animation: pin-bounce 1.6s ease-in-out infinite;
}

.pin-dot.pickup { background: #10B981; }
.pin-dot.destination { background: #F59E0B; }

.pin-shadow {
  width: 6px;
  height: 6px;
  margin-top: 2px;
  border-radius: 50%;
  background: rgba(15, 23, 42, 0.28);
  animation: pin-shadow 1.6s ease-in-out infinite;
}

@keyframes pin-bounce {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-5px); }
}

@keyframes pin-shadow {
  0%, 100% { transform: scale(1); opacity: 0.28; }
  50% { transform: scale(0.7); opacity: 0.16; }
}

@keyframes pin-halo {
  0%, 100% { transform: scale(0.85); }
  50% { transform: scale(1.1); }
}

.picker-header {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  z-index: 10;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: calc(env(safe-area-inset-top) + 12px) 16px 12px;
}

.header-title {
  font-size: 17px;
  font-weight: 600;
  color: #1F2937;
}

.header-btn,
.header-spacer { width: 38px; height: 38px; }

.header-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  border: 0;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.94);
  color: #1F2937;
  box-shadow: 0 2px 8px rgba(15, 23, 42, 0.12);
  cursor: pointer;
}

.search-layer {
  position: absolute;
  top: calc(env(safe-area-inset-top) + 62px);
  left: 0;
  right: 0;
  z-index: 12;
}

.locate-fab {
  position: absolute;
  right: 16px;
  bottom: calc(var(--panel-height) + 18px);
  z-index: 9;
}

.picker-panel {
  position: fixed;
  right: 0;
  bottom: 0;
  left: 0;
  z-index: 8;
  height: var(--panel-height);
  padding: 16px 16px calc(12px + env(safe-area-inset-bottom));
  border-radius: 18px 18px 0 0;
  background: #FFFFFF;
  box-shadow: 0 -6px 24px rgba(15, 23, 42, 0.14);
}
</style>
