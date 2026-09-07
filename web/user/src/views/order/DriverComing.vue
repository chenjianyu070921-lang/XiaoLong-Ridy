<template>
  <div class="driver-coming-page">
    <!-- 地图 -->
    <div class="map-container" id="driver-map">
      <div v-if="mapError" class="map-error">{{ mapError }}</div>
    </div>

    <!-- 司机信息卡片 -->
    <div class="driver-card animate-slideUp">
      <!-- 司机信息 -->
      <div class="driver-info">
        <div class="avatar">
          <img :src="driverInfo.avatar || '/default-avatar.png'" alt="司机头像" />
          <span class="online-status"></span>
        </div>
        <div class="info">
          <h3>{{ driverInfo.name }}</h3>
          <div class="rating">
            <van-rate v-model="driverInfo.rating" readonly size="12" color="#F59E0B" void-icon="star" void-color="#E5E7EB" />
            <span class="score">{{ driverInfo.rating }}分</span>
          </div>
          <p class="car-info">{{ driverInfo.carModel }}</p>
          <p class="plate-number">{{ driverInfo.plateNumber }}</p>
        </div>
        <div class="actions">
          <button class="action-btn call" @click="callDriver">
            <van-icon name="phone-o" size="20" />
            <span>联系司机</span>
          </button>
        </div>
      </div>

      <!-- 到达时间 -->
      <div class="arrival-time">
        <div class="time-item">
          <van-icon name="clock-o" size="18" color="#7C3AED" />
          <span>预计 {{ arrivalMinutes }} 分钟后到达</span>
        </div>
        <div class="time-item">
          <van-icon name="guide-o" size="18" color="#10B981" />
          <span>距离您 {{ distance }} 米</span>
        </div>
      </div>

      <!-- 行程详情 -->
      <div class="trip-detail">
        <div class="route-line"></div>
        <div class="address from">
          <div class="dot"></div>
          <span>{{ orderStore.orderParams.fromAddress || '我的位置' }}</span>
        </div>
        <div class="address to">
          <div class="dot"></div>
          <span>{{ orderStore.orderParams.toAddress }}</span>
        </div>
      </div>
    </div>

    <!-- 底部操作栏 -->
    <div class="bottom-bar safe-area-bottom">
      <button class="btn-secondary" @click="shareTrip">分享行程</button>
      <button class="btn-danger" @click="showCancelDialog = true">取消订单</button>
    </div>

    <!-- 取消确认弹窗 -->
    <van-dialog
      v-model:show="showCancelDialog"
      title="取消订单"
      show-cancel-button
      @confirm="handleCancel"
    >
      <p style="padding: 20px; text-align: center;">确定要取消当前订单吗？</p>
    </van-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { showToast, showDialog, showLoadingToast, closeToast } from 'vant'
import AMapLoader from '@amap/amap-jsapi-loader'
import { getAmapConfig } from '@/config/amap'
import { useOrderStore } from '@/stores/order'
import { formatDriverDisplayName, formatPlateNumber } from '@/constants/order'
import { cancelOrder, pollOrderStatus, getOrderTracking } from '@/api/order'

const router = useRouter()
const orderStore = useOrderStore()

// 乘客接口当前只提供司机 ID；姓名通过 driversvc 实时聚合后按"姓氏+师傅"展示，
// 真实姓名或车牌未返回时降级使用占位文案，禁止凭空伪造信息。
const driverInfo = ref({
  name: '司机信息加载中',
  avatar: '',
  rating: 0,
  carColor: '',
  carModel: '',
  plateNumber: '车牌信息加载中',
  phone: ''
})

// 司机姓名与车牌号统一使用 constants/order 的脱敏规则：姓名只展示"姓氏+师傅"，
// 车牌为后端返回值；两者缺失时由共享函数降级为占位文案。

const arrivalMinutes = ref(0)
const distance = ref(0)
const showCancelDialog = ref(false)
const mapError = ref('')

let pollTimer = null
let trackingTimer = null
let mapInstance = null
let driverMarker = null

// 把订单快照里的司机信息同步到展示字段。真实姓名按"姓氏+师傅"展示，
// 真实姓名或车牌号缺失时用占位文案，绝不把"司机 #ID"当真实姓名展示。
const syncDriverInfo = (order = {}) => {
  const driverID = Number(order.driverId || order.driverID || 0)
  driverInfo.value.name = formatDriverDisplayName(order.driverName, driverID)
  driverInfo.value.plateNumber = formatPlateNumber(order.plateNumber || order.plate || '')
  driverInfo.value.carModel = (order.carModel || '').trim() || '车型信息加载中'
  driverInfo.value.phone = order.driverPhone || order.phone || ''
  driverInfo.value.avatar = order.driverAvatar || order.avatar || ''
}

// 使用真实追踪快照更新司机位置、距离和预计到达时间。
const refreshTracking = async () => {
  const orderId = orderStore.currentOrder?.orderId
  if (!orderId) return
  try {
    const snapshot = await getOrderTracking(orderId)
    // 司机接单信息来自订单快照（轮询携带真实姓名/车牌），实时追踪接口只负责位置，二者合并展示。
    syncDriverInfo(orderStore.currentOrder || {})
    distance.value = Number(snapshot?.remainingDistanceM || 0)
    arrivalMinutes.value = Math.max(0, Math.ceil(Number(snapshot?.remainingDurationS || 0) / 60))
    const position = [Number(snapshot?.driverLongitude), Number(snapshot?.driverLatitude)]
    if (mapInstance && position.every(Number.isFinite)) {
      if (!driverMarker) {
        driverMarker = new window.AMap.Marker({
          position,
          title: '司机当前位置',
          anchor: 'center',
          // 使用汽车图标突出司机实时位置，并保留 heading 方向。
          content: '<div style="width:32px;height:32px;display:flex;align-items:center;justify-content:center;border-radius:50%;background:#fff;box-shadow:0 2px 8px #0003;font-size:21px">🚕</div>',
          angle: Number(snapshot?.heading || 0)
        })
        mapInstance.add(driverMarker)
      } else {
        driverMarker.setPosition(position)
        if (typeof driverMarker.setAngle === 'function') driverMarker.setAngle(Number(snapshot?.heading || 0))
      }
      mapInstance.setCenter(position)
    }
  } catch (error) {
    mapError.value = '司机位置暂不可用，正在重试...'
    console.error('获取司机追踪失败:', error)
  }
}

// 联系司机
const callDriver = () => {
  if (!driverInfo.value.phone) {
    showToast('司机暂未提供联系电话')
    return
  }
  showDialog({
    title: '联系司机',
    message: `是否拨打司机电话：${driverInfo.value.phone}`,
    confirmButtonText: '拨打',
    cancelButtonText: '取消'
  }).then(() => {
    // 实际项目中这里会调用系统拨号功能
    window.location.href = `tel:${driverInfo.value.phone.replace(/\*/g, '')}`
  }).catch(() => {})
}

// 分享行程
const shareTrip = () => {
  if (navigator.share) {
    navigator.share({
      title: '花小龙打车 - 分享行程',
      text: '我正在乘坐花小龙打车，预计5分钟后到达目的地',
      url: window.location.href
    })
  } else {
    showToast('分享功能暂不可用')
  }
}

// 取消订单
const handleCancel = async () => {
  try {
    const toast = showLoadingToast({
      message: '正在取消...',
      forbidClick: true,
      duration: 0
    })

    await cancelOrder(orderStore.currentOrder?.orderId)
    
    closeToast()
    showToast('订单已取消')
    router.replace('/home')
  } catch (error) {
    console.error(error)
  }
}

// 轮询状态
const pollStatus = async () => {
  try {
    const result = await pollOrderStatus(orderStore.currentOrder?.orderId)
    const status = Number(result?.status)
    // 轮询接口会带回司机姓名/车牌号；只覆盖当前已存在的字段，避免把 driverId 等重置。
    const merged = { ...orderStore.currentOrder, ...result }
    orderStore.setCurrentOrder(merged)
    // 同步司机姓名、车牌号到展示字段。
    syncDriverInfo(merged)
    if (status === 3) {
      router.replace('/order/trip')
    } else if (status === 6 || status === 7) {
      showToast('订单已取消')
      router.replace('/home')
    }
  } catch (error) {
    console.error(error)
  }
}

onMounted(async () => {
  // 进入页面时把 store 中已有的司机姓名/车牌立刻同步展示，避免在轮询返回前一直显示占位。
  syncDriverInfo(orderStore.currentOrder || {})
  // 初始化地图并立即拉取一次真实司机位置。
  await initDriverMap()
  await refreshTracking()
  
  // 开始轮询状态
  pollTimer = setInterval(pollStatus, 3000)
  trackingTimer = setInterval(refreshTracking, 5000)
})

onUnmounted(() => {
  if (pollTimer) clearInterval(pollTimer)
  if (trackingTimer) clearInterval(trackingTimer)
  mapInstance?.destroy()
  mapInstance = null
})

// 初始化高德地图；地图中心会在追踪接口返回后移动到司机坐标。
const initDriverMap = async () => {
  const { key, securityCode } = getAmapConfig()
  if (!key) {
    mapError.value = '未配置高德地图 Key，无法显示司机位置'
    return
  }
  try {
    if (securityCode) window._AMapSecurityConfig = { securityJsCode: securityCode }
    const AMap = await AMapLoader.load({ key, version: '2.0' })
    window.AMap = AMap
    mapInstance = new AMap.Map('driver-map', { zoom: 15, viewMode: '2D' })
  } catch (error) {
    mapError.value = '高德地图加载失败，请检查配置'
    console.error('AMap map error:', error)
  }
}
</script>

<style scoped>
.driver-coming-page {
  min-height: 100vh;
  background: #f5f5f5;
}

.map-container {
  /* 适当降低地图高度，避免在矮屏设备上司机卡片被裁切或与底部按钮重叠。 */
  height: 46vh;
  min-height: 280px;
  background: #E5E7EB;
  position: relative;
}

.map-error {
  position: absolute;
  top: 16px;
  left: 50%;
  transform: translateX(-50%);
  padding: 8px 12px;
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.94);
  color: #92400E;
  font-size: 13px;
  white-space: nowrap;
}

.driver-card {
  /* 相对地图仅有 18px 叠放即可保持视觉层次，同时保证头部信息不被地图遮挡。 */
  margin: -18px 16px 96px;
  position: relative;
  z-index: 2;
  background: white;
  border-radius: 16px;
  padding: 20px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);
}

.driver-info {
  display: flex;
  gap: 14px;
  margin-bottom: 20px;
  padding-bottom: 20px;
  border-bottom: 1px solid #F3F4F6;
}

.avatar {
  position: relative;
  width: 60px;
  height: 60px;
  flex-shrink: 0;
}

.avatar img {
  width: 100%;
  height: 100%;
  border-radius: 50%;
  object-fit: cover;
  background: #E5E7EB;
}

.online-status {
  position: absolute;
  bottom: 2px;
  right: 2px;
  width: 12px;
  height: 12px;
  background: #10B981;
  border-radius: 50%;
  border: 2px solid white;
}

.info {
  flex: 1;
}

.info h3 {
  font-size: 17px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 6px;
}

.rating {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-bottom: 4px;
}

.score {
  font-size: 13px;
  color: #F59E0B;
  font-weight: 500;
}

.car-info {
  font-size: 13px;
  color: #6B7280;
  margin-bottom: 2px;
}

.plate-number {
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary);
}

.actions {
  display: flex;
  flex-direction: column;
  justify-content: center;
}

.action-btn {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  padding: 8px 12px;
  border-radius: 8px;
  cursor: pointer;
  font-size: 11px;
  transition: all 0.2s;
}

.action-btn.call {
  background: #EFF6FF;
  color: #3B82F6;
}

.action-btn span {
  white-space: nowrap;
}

.arrival-time {
  display: flex;
  justify-content: space-around;
  padding: 16px 0;
  margin-bottom: 16px;
  background: #F9FAFB;
  border-radius: 10px;
}

.time-item {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  color: var(--text-primary);
}

.trip-detail {
  position: relative;
  padding-left: 20px;
}

.route-line {
  position: absolute;
  left: 7px;
  top: 8px;
  bottom: 32px;
  width: 2px;
  background: #D1D5DB;
}

.address {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 10px 0;
  position: relative;
}

.address .dot {
  width: 14px;
  height: 14px;
  border-radius: 50%;
  margin-top: 2px;
  position: absolute;
  left: -17px;
}

.address.from .dot {
  background: #10B981;
}

.address.to .dot {
  background: #EF4444;
}

.address span {
  font-size: 14px;
  color: var(--text-primary);
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

.btn-secondary,
.btn-danger {
  flex: 1;
  height: 44px;
  border-radius: 22px;
  font-size: 15px;
  cursor: pointer;
  border: none;
  transition: all 0.2s;
}

.btn-secondary {
  background: #F3F4F6;
  color: #374151;
}

.btn-secondary:active {
  background: #E5E7EB;
}

.btn-danger {
  background: #FEF2F2;
  color: #DC2626;
}

.btn-danger:active {
  background: #FEE2E2;
}
</style>
