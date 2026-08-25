<template>
  <div class="driver-coming-page">
    <!-- 地图 -->
    <div class="map-container" id="driver-map"></div>

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
          <p class="car-info">{{ driverInfo.carColor }} · {{ driverInfo.carModel }}</p>
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
import { useOrderStore } from '@/stores/order'
import { cancelOrder, pollOrderStatus } from '@/api/order'

const router = useRouter()
const orderStore = useOrderStore()

// 司机信息（模拟数据）
const driverInfo = ref({
  name: '张师傅',
  avatar: '',
  rating: 4.9,
  carColor: '白色',
  carModel: '大众·大众朗逸',
  plateNumber: '渝A·12345',
  phone: '138****5678'
})

const arrivalMinutes = ref(5)
const distance = ref(1200)
const showCancelDialog = ref(false)

let pollTimer = null

// 联系司机
const callDriver = () => {
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
    if (Number(result?.status) === 3) {
      router.replace('/order/trip')
    }
  } catch (error) {
    console.error(error)
  }
}

onMounted(() => {
  // 初始化地图显示司机位置
  initDriverMap()
  
  // 开始轮询状态
  pollTimer = setInterval(pollStatus, 3000)
  
  // 模拟倒计时
  setInterval(() => {
    if (arrivalMinutes.value > 0) {
      arrivalMinutes.value--
      distance.value = Math.max(0, distance.value - 200)
    }
  }, 60000)
})

onUnmounted(() => {
  if (pollTimer) clearInterval(pollTimer)
})

const initDriverMap = () => {
  // 初始化地图，显示司机位置和路线
}
</script>

<style scoped>
.driver-coming-page {
  min-height: 100vh;
  background: #f5f5f5;
}

.map-container {
  height: 55vh;
  background: #E5E7EB;
}

.driver-card {
  margin: -30px 16px 80px;
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
