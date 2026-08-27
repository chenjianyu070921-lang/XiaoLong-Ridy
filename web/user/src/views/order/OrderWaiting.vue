<template>
  <div class="waiting-page">
    <!-- 返回首页但不取消当前订单，订单轮询会继续运行。 -->
    <button type="button" class="back-home-btn" aria-label="返回首页" @click="backToHome">
      <van-icon name="arrow-left" size="20" />
    </button>
    <!-- 地图背景 -->
    <div class="map-bg">
      <div class="searching-animation">
        <div class="pulse-ring"></div>
        <div class="pulse-ring delay"></div>
        <van-icon name="location-o" size="40" color="#7C3AED" />
      </div>
    </div>

    <!-- 状态卡片 -->
    <div class="status-card animate-slideUp">
      <h2 class="status-title">正在寻找司机</h2>
      <p class="status-desc">正在为您匹配附近司机，请稍候...</p>
      
      <!-- 搜索动画 -->
      <div class="searching-dots">
        <span></span><span></span><span></span>
      </div>

      <!-- 行程信息 -->
      <div class="trip-info">
        <div class="trip-item">
          <span class="label">出发地</span>
          <span class="value">{{ orderStore.orderParams.fromAddress || '我的位置' }}</span>
        </div>
        <div class="trip-item">
          <span class="label">目的地</span>
          <span class="value">{{ orderStore.orderParams.toAddress }}</span>
        </div>
        <div class="trip-item">
          <span class="label">车型</span>
          <span class="value">{{ currentCarType }}</span>
        </div>
      </div>

      <!-- 计时器 -->
      <div class="timer">
        <van-icon name="clock-o" size="16" color="#6B7280" />
        <span>已等待 {{ formatTime(waitTime) }}</span>
      </div>
    </div>

    <!-- 底部操作 -->
    <div class="bottom-actions safe-area-bottom">
      <button class="btn-cancel" :disabled="isCancelling" @click="showCancelDialog = true">
        {{ isCancelling ? '正在取消...' : '取消订单' }}
      </button>
      <p class="tip">预计匹配时间：{{ estimatedWaitTime }}分钟</p>
    </div>

    <!-- 取消确认弹窗 -->
    <van-dialog
      v-model:show="showCancelDialog"
      title="取消订单"
      show-cancel-button
      confirm-button-text="确定取消"
      cancel-button-text="再等等"
      @confirm="handleCancel"
    >
      <div class="cancel-content">
        <p>确定要取消当前订单吗？</p>
        <div class="cancel-reasons">
          <van-radio-group v-model="cancelReason">
            <van-radio name="change_plan">行程有变，暂时不需要了</van-radio>
            <van-radio name="wait_too_long">等待时间太长</van-radio>
            <van-radio name="price_high">价格太高</van-radio>
            <van-radio name="other">其他原因</van-radio>
          </van-radio-group>
        </div>
        <van-field
          v-if="cancelReason === 'other'"
          v-model="otherCancelReason"
          class="other-reason-field"
          type="textarea"
          rows="3"
          maxlength="100"
          show-word-limit
          placeholder="请告诉我们取消的原因"
        />
      </div>
    </van-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { showToast, showLoadingToast, closeToast } from 'vant'
import { useOrderStore } from '@/stores/order'
import { cancelOrder, pollOrderStatus } from '@/api/order'

const router = useRouter()
const orderStore = useOrderStore()

// 返回首页继续等待，保留当前订单和轮询状态。
const backToHome = () => router.replace('/home')

// 状态
const waitTime = ref(0)
const showCancelDialog = ref(false)
const cancelReason = ref('change_plan')
const otherCancelReason = ref('')
const isCancelling = ref(false)

// 将预设原因或用户填写的其他原因转换为提交给订单服务的文本。
const submittedCancelReason = computed(() => cancelReason.value === 'other' ? otherCancelReason.value.trim() : cancelReason.value)
const orderStatus = ref(1)
let timer = null
let pollTimer = null

// 当前车型
const currentCarType = computed(() => {
  const selected = orderStore.carTypes.find(c => c.selected)
  return selected?.name || ''
})

// 预估等待时间
const estimatedWaitTime = computed(() => {
  const selected = orderStore.carTypes.find(c => c.selected)
  return selected?.time?.replace('~', '').replace('分钟', '') || '5'
})

// 格式化时间
const formatTime = (seconds) => {
  const mins = Math.floor(seconds / 60)
  const secs = seconds % 60
  return `${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`
}

// 轮询订单状态
const pollStatus = async () => {
  try {
    const orderId = orderStore.currentOrder?.orderId
    if (!orderId) return
    const result = await pollOrderStatus(orderId, orderStatus.value)
    orderStatus.value = Number(result?.status || orderStatus.value)
    orderStore.setCurrentOrder({ ...orderStore.currentOrder, ...result })
    if (orderStatus.value === 2) router.replace('/order/driver-coming')
    if (orderStatus.value === 6) {
      showToast('订单已取消')
      orderStore.setCurrentOrder(null)
      router.replace('/home')
    }
  } catch (error) {
    console.error('Poll status error:', error)
  }
}

// 取消订单
const handleCancel = async () => {
  // 防止确认弹窗重复触发时并发提交多个取消请求。
  if (isCancelling.value) return
  isCancelling.value = true
  if (!submittedCancelReason.value) {
    isCancelling.value = false
    showToast('请先填写取消原因')
    return
  }
  try {
    showLoadingToast({ message: '正在取消...', forbidClick: true, duration: 0 })
    if (orderStore.currentOrder?.orderId) {
      await cancelOrder(orderStore.currentOrder.orderId, submittedCancelReason.value)
    }
    closeToast()
    showToast('订单已取消')
    orderStore.resetOrderParams()
    orderStore.setCurrentOrder(null)
    router.replace('/home')
  } catch (error) {
    console.error('Cancel error:', error)
    closeToast()
    showToast(error?.response?.data?.message || '取消订单失败，请稍后重试')
  } finally {
    isCancelling.value = false
  }
}

onMounted(() => {
  // 开始计时
  timer = setInterval(() => {
    waitTime.value++
  }, 1000)

  // 先立即刷新一次，再按 3 秒间隔轮询，减少用户下单后的空白等待。
  void pollStatus()
  pollTimer = setInterval(pollStatus, 3000)
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
  if (pollTimer) clearInterval(pollTimer)
})
</script>

<style scoped>
.back-home-btn { position: fixed; top: calc(env(safe-area-inset-top) + 12px); left: 16px; z-index: 20; width: 40px; height: 40px; display: inline-flex; align-items: center; justify-content: center; border: 0; border-radius: 50%; color: #374151; background: rgba(255, 255, 255, 0.92); box-shadow: 0 4px 14px rgba(15, 23, 42, 0.14); cursor: pointer; }
.waiting-page {
  min-height: 100vh;
  background: #f5f5f5;
}

.map-bg {
  height: 50vh;
  background: linear-gradient(180deg, #EDE9FE 0%, #F5F5F5 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
}

.searching-animation {
  position: relative;
  width: 120px;
  height: 120px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.pulse-ring {
  position: absolute;
  width: 80px;
  height: 80px;
  border-radius: 50%;
  border: 2px solid rgba(124, 58, 237, 0.4);
  animation: pulse-ring 2s ease-out infinite;
}

.pulse-ring.delay {
  animation-delay: 1s;
}

@keyframes pulse-ring {
  0% {
    transform: scale(0.8);
    opacity: 1;
  }
  100% {
    transform: scale(2);
    opacity: 0;
  }
}

.status-card {
  /* 通过层级覆盖地图背景，避免负外边距区域被地图遮挡。 */
  position: relative;
  z-index: 2;
  margin: -40px 16px 20px;
  background: white;
  border-radius: 16px;
  padding: 24px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);
  text-align: center;
}

.status-title {
  font-size: 22px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 8px;
}

.status-desc {
  font-size: 14px;
  color: #6B7280;
  margin-bottom: 24px;
}

.searching-dots {
  display: flex;
  justify-content: center;
  gap: 8px;
  margin-bottom: 24px;
}

.searching-dots span {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background: #7C3AED;
  animation: bounce 1.4s ease-in-out infinite both;
}

.searching-dots span:nth-child(1) { animation-delay: -0.32s; }
.searching-dots span:nth-child(2) { animation-delay: -0.16s; }

@keyframes bounce {
  0%, 80%, 100% {
    transform: scale(0);
  }
  40% {
    transform: scale(1);
  }
}

.trip-info {
  background: #F9FAFB;
  border-radius: 10px;
  padding: 16px;
  margin-bottom: 16px;
  text-align: left;
}

.trip-item {
  display: flex;
  justify-content: space-between;
  padding: 8px 0;
  font-size: 14px;
}

.trip-item .label {
  color: #6B7280;
}

.trip-item .value {
  color: var(--text-primary);
  font-weight: 500;
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.timer {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  font-size: 14px;
  color: #6B7280;
}

.bottom-actions {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  /* 底部操作区预留安全区，避免全面屏设备遮挡按钮和提示。 */
  padding: 16px 20px calc(16px + env(safe-area-inset-bottom));
  background: white;
  text-align: center;
  box-shadow: 0 -2px 10px rgba(0, 0, 0, 0.05);
}

.btn-cancel {
  width: 100%;
  height: 44px;
  background: white;
  border: 1px solid #D1D5DB;
  border-radius: 22px;
  color: #6B7280;
  font-size: 15px;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-cancel:active {
  background: #F9FAFB;
}

.tip {
  margin-top: 12px;
  font-size: 12px;
  color: #9CA3AF;
}

.cancel-content {
  padding: 8px 0 4px;
}

.cancel-content > p {
  padding: 0 4px;
  font-size: 15px;
  color: var(--text-primary);
  margin-bottom: 20px;
}

.cancel-reasons {
  padding: 8px 12px;
  border-radius: 12px;
  background: #f8fafc;
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.other-reason-field {
  margin-top: 12px;
  border: 1px solid #e5e7eb;
  border-radius: 10px;
  background: #fff;
}
</style>







