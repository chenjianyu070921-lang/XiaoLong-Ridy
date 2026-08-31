<template>
  <div class="trip-page">
    <!-- 地图 -->
    <div class="map-container" id="trip-map">
      <div v-if="trackingLoading" class="map-state">正在获取实时位置...</div>
      <div v-else-if="trackingError" class="map-state error">{{ trackingError }}</div>
    </div>

    <!-- 行程信息卡片 -->
    <div class="trip-card animate-slideUp">
      <!-- 状态标签 -->
      <div class="status-badge">
        <van-icon name="guide-o" size="16" />
        <span>行程进行中</span>
      </div>

      <!-- 司机信息（简化版） -->
      <div class="driver-mini">
        <img :src="driverInfo.avatar || '/default-avatar.png'" alt="" />
        <div class="info">
          <span class="name">{{ driverInfo.name }}</span>
          <span class="car">{{ driverInfo.plateNumber }} · {{ driverInfo.carModel }}</span>
        </div>
        <button class="call-btn" @click="callDriver">
          <van-icon name="phone-o" size="18" color="#3B82F6" />
        </button>
      </div>

      <!-- 行程统计 -->
      <div class="trip-stats">
        <div class="stat-item">
          <p class="value">{{ tripStats.distance }}</p>
          <p class="label">已行驶</p>
        </div>
        <div class="stat-divider"></div>
        <div class="stat-item">
          <p class="value">{{ tripStats.duration }}</p>
          <p class="label">已用时</p>
        </div>
        <div class="stat-divider"></div>
        <div class="stat-item">
          <p class="value price">¥{{ tripStats.estimatedPrice }}</p>
          <p class="label">预估费用</p>
        </div>
      </div>

      <!-- 路线信息 -->
      <div class="route-info">
        <div class="route-item">
          <div class="dot from"></div>
          <div class="text">
            <p class="main">{{ orderStore.orderParams.fromAddress || '出发地' }}</p>
            <p class="sub">上车点</p>
          </div>
        </div>
        <div class="route-line"></div>
        <div class="route-item">
          <div class="dot to"></div>
          <div class="text">
            <p class="main">{{ orderStore.orderParams.toAddress || '目的地' }}</p>
            <p class="sub">预计{{ tripStats.estimatedArrival }}到达</p>
          </div>
        </div>
      </div>

      <!-- 安全提示 -->
      <div class="safety-tips">
        <van-icon name="shield-o" size="16" color="#10B981" />
        <span>行程中如遇紧急情况，请点击下方紧急求助</span>
      </div>
    </div>

    <!-- 底部操作栏 -->
    <div class="bottom-bar safe-area-bottom">
      <button class="btn-sos" @click="handleSOS">
        <van-icon name="warning-o" size="18" />
        紧急求助
      </button>
      <button class="btn-share" @click="shareTrip">
        <van-icon name="share-o" size="18" />
        分享行程
      </button>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import AMapLoader from '@amap/amap-jsapi-loader'
import { useRouter } from 'vue-router'
import { showToast, showDialog } from 'vant'
import { useOrderStore } from '@/stores/order'
import { getOrderTracking, pollOrderStatus } from '@/api/order'
import { getAmapConfig } from '@/config/amap'

const router = useRouter()
const orderStore = useOrderStore()

// 司机信息
const driverInfo = ref({
  name: '张师傅',
  avatar: '',
  plateNumber: '渝A·12345',
  carModel: '大众朗逸',
  phone: '138****5678'
})

// 行程统计数据
const tripStats = ref({
  distance: '3.2公里',
  duration: '8分钟',
  estimatedPrice: '20.6',
  estimatedArrival: '12:30'
})

let pollTimer = null
let trackingTimer = null
let mapInstance = null
let driverMarker = null
let routePolyline = null
const trackingLoading = ref(true)
const trackingError = ref('')

// 联系司机
const callDriver = () => {
  showDialog({
    title: '联系司机',
    message: `是否拨打司机电话：${driverInfo.value.phone}`,
    showCancelButton: true
  }).then(() => {
    window.location.href = `tel:${driverInfo.value.phone.replace(/\*/g, '')}`
  }).catch(() => {})
}

// 分享行程
const shareTrip = () => {
  if (navigator.share) {
    navigator.share({
      title: '花小龙打车 - 实时分享',
      text: '我正在乘坐花小龙打车，可实时查看我的位置',
      url: window.location.href
    })
  } else {
    showToast('分享功能暂不可用')
  }
}

// 紧急求助：弹窗确认后进入客服接入流程。
const handleSOS = () => {
  showDialog({
    title: '紧急求助',
    message: '是否立即联系客服或报警？',
    confirmButtonText: '联系客服',
    cancelButtonText: '取消',
    showCancelButton: true
  }).then(() => {
    // 联系客服或报警
    showToast('正在为您接通客服...')
  }).catch(() => {})
}

// 轮询订单状态
const pollStatus = async () => {
  try {
    const result = await pollOrderStatus(orderStore.currentOrder?.orderId)
    const status = Number(result?.status)
    orderStore.setCurrentOrder({ ...orderStore.currentOrder, ...result })

    if (status === 4) {
      router.replace('/order/payment')
    } else if (status === 6 || status === 7) {
      showToast('订单已被取消')
      router.replace('/home')
    }
  } catch (error) {
    console.error(error)
  }
}

// 将秒数格式化为乘客容易阅读的行程时长。
const formatDuration = (seconds) => `${Math.max(1, Math.ceil(Number(seconds || 0) / 60))}分钟`

// 将米数格式化为公里，短距离仍保留米制展示。
const formatDistance = (meters) => {
  const value = Number(meters || 0)
  return value < 1000 ? `${Math.max(0, Math.round(value))}米` : `${(value / 1000).toFixed(1)}公里`
}

// 把后端返回的高德点串转换为地图坐标数组。
const parsePolyline = (polyline) => String(polyline || '').split(';').map(point => {
  const [lng, lat] = point.split(',').map(Number)
  return Number.isFinite(lng) && Number.isFinite(lat) ? [lng, lat] : null
}).filter(Boolean)

// 更新司机标记、剩余路线和行程统计，地图对象只创建一次。
const renderTracking = (snapshot) => {
  // 追踪接口可能先于地图完成初始化返回，必须等待地图实例就绪。
  if (!mapInstance || typeof mapInstance.add !== 'function' || !snapshot) return
  const position = [Number(snapshot.driverLongitude), Number(snapshot.driverLatitude)]
  if (!position.every(Number.isFinite)) return

  if (!driverMarker) {
    driverMarker = new window.AMap.Marker({ position, title: '司机当前位置', anchor: 'center' })
    mapInstance.add(driverMarker)
  } else {
    driverMarker.setPosition(position)
  }

  const path = parsePolyline(snapshot.polyline)
  if (path.length > 1) {
    if (!routePolyline) {
      routePolyline = new window.AMap.Polyline({ path, strokeColor: '#7C3AED', strokeWeight: 7, showDir: true })
      mapInstance.add(routePolyline)
    } else {
      routePolyline.setPath(path)
    }
    mapInstance.setFitView([driverMarker, routePolyline], false, [70, 40, 70, 40])
  } else {
    mapInstance.setCenter(position)
  }

  tripStats.value.distance = formatDistance(snapshot.travelledDistanceM)
  tripStats.value.duration = formatDuration(snapshot.elapsedDurationS ?? snapshot.durationS ?? 0)
  tripStats.value.estimatedPrice = (Number(snapshot.estimatedPriceCents || 0) / 100).toFixed(2)
  if (snapshot.remainingDurationS) {
    tripStats.value.estimatedArrival = formatDuration(snapshot.remainingDurationS)
  }
}

// 拉取单次追踪快照；短暂网络失败不会清空地图上的最后有效位置。
const refreshTracking = async () => {
  const orderId = orderStore.currentOrder?.orderId
  if (!orderId) {
    trackingError.value = '未找到进行中的订单'
    trackingLoading.value = false
    return
  }
  try {
    const snapshot = await getOrderTracking(orderId)
    trackingError.value = snapshot?.stale ? '司机位置更新时间较早，正在重新连接...' : ''
    renderTracking(snapshot)
  } catch (error) {
    console.error('获取行程追踪失败', error)
    trackingError.value = '实时位置暂不可用，正在重试...'
  } finally {
    trackingLoading.value = false
  }
}

onMounted(() => {
  initMap().then(refreshTracking)
  pollTimer = setInterval(pollStatus, 5000)
  trackingTimer = setInterval(refreshTracking, 3000)
})

onUnmounted(() => {
  if (pollTimer) clearInterval(pollTimer)
  if (trackingTimer) clearInterval(trackingTimer)
  mapInstance?.destroy()
  mapInstance = null
})

// 初始化高德地图；密钥配置与首页地图保持一致。
const initMap = async () => {
  const { key, securityCode } = getAmapConfig()
  window._AMapSecurityConfig = { securityJsCode: securityCode }
  const AMap = await AMapLoader.load({ key, version: '2.0' })
  window.AMap = AMap
  mapInstance = new AMap.Map('trip-map', {
    zoom: 15,
    center: [Number(orderStore.orderParams.fromLng) || 116.397428, Number(orderStore.orderParams.fromLat) || 39.90923]
  })
}
</script>

<style scoped>
.trip-page {
  min-height: 100vh;
  background: #f5f5f5;
}

.map-container {
  position: relative;
  height: 50vh;
  background: #E5E7EB;
}

.map-state {
  position: absolute;
  z-index: 10;
  top: 16px;
  left: 50%;
  max-width: calc(100% - 32px);
  padding: 8px 12px;
  transform: translateX(-50%);
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.94);
  color: #374151;
  font-size: 13px;
  white-space: nowrap;
  box-shadow: 0 2px 8px rgba(15, 23, 42, 0.14);
}

.map-state.error {
  color: #B45309;
}

.trip-card {
  margin: -40px 16px 80px;
  background: white;
  border-radius: 16px;
  padding: 20px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);
}

.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background: linear-gradient(135deg, #7C3AED 0%, #9333EA 100%);
  color: white;
  padding: 6px 14px;
  border-radius: 20px;
  font-size: 13px;
  font-weight: 500;
  margin-bottom: 16px;
}

.driver-mini {
  display: flex;
  align-items: center;
  gap: 10px;
  padding-bottom: 16px;
  margin-bottom: 16px;
  border-bottom: 1px solid #F3F4F6;
}

.driver-mini img {
  width: 44px;
  height: 44px;
  border-radius: 50%;
  object-fit: cover;
  background: #E5E7EB;
}

.info {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.name {
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary);
}

.car {
  font-size: 12px;
  color: #6B7280;
  margin-top: 2px;
}

.call-btn {
  width: 40px;
  height: 40px;
  background: #EFF6FF;
  border: none;
  border-radius: 50%;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}

.call-btn:active {
  transform: scale(0.95);
}

.trip-stats {
  display: flex;
  align-items: center;
  justify-content: space-around;
  padding: 20px 0;
  margin-bottom: 20px;
  background: #F9FAFB;
  border-radius: 10px;
}

.stat-item {
  text-align: center;
}

.stat-item .value {
  font-size: 20px;
  font-weight: 700;
  color: var(--text-primary);
  margin-bottom: 4px;
}

.stat-item .value.price {
  color: #EF4444;
}

.stat-item .label {
  font-size: 12px;
  color: #6B7280;
}

.stat-divider {
  width: 1px;
  height: 36px;
  background: #E5E7EB;
}

.route-info {
  position: relative;
  padding-left: 24px;
}

.route-item {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 12px 0;
  position: relative;
}

.dot {
  width: 12px;
  height: 12px;
  border-radius: 50%;
  position: absolute;
  left: -19px;
  top: 16px;
}

.dot.from {
  background: #10B981;
}

.dot.to {
  background: #EF4444;
}

.text .main {
  font-size: 14px;
  font-weight: 500;
  color: var(--text-primary);
  margin-bottom: 2px;
}

.text .sub {
  font-size: 12px;
  color: #9CA3AF;
}

.route-line {
  position: absolute;
  left: -14px;
  top: 28px;
  bottom: 28px;
  width: 2px;
  background: #D1D5DB;
}

.safety-tips {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 16px;
  padding: 12px;
  background: #ECFDF5;
  border-radius: 8px;
  font-size: 12px;
  color: #059669;
}

.bottom-bar {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  padding: 12px 20px;
  background: white;
  display: flex;
  gap: 12px;
  box-shadow: 0 -2px 10px rgba(0, 0, 0, 0.05);
}

.btn-sos,
.btn-share {
  flex: 1;
  height: 44px;
  border-radius: 22px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  border: none;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  transition: all 0.2s;
}

.btn-sos {
  background: linear-gradient(135deg, #FEF2F2 0%, #FEE2E2 100%);
  color: #DC2626;
}

.btn-sos:active {
  transform: scale(0.98);
}

.btn-share {
  background: linear-gradient(135deg, #EFF6FF 0%, #DBEAFE 100%);
  color: #2563EB;
}

.btn-share:active {
  transform: scale(0.98);
}
</style>
