<template>
  <div class="payment-page">
    <!-- 顶部状态 -->
    <div class="header">
      <van-icon name="arrow-left" size="20" @click="goBack" />
      <span class="title">行程结束</span>
      <div style="width: 20px;"></div>
    </div>

    <!-- 费用卡片 -->
    <div class="price-card">
      <p class="label">行程已结束，感谢乘坐</p>
      <h1 class="price">¥{{ orderDetail.totalPrice }}</h1>
      
      <!-- 费用明细 -->
      <div class="fee-detail">
        <div class="fee-item">
          <span>起步价({{ orderDetail.baseDistance }}公里)</span>
          <span>¥{{ orderDetail.basePrice }}</span>
        </div>
        <div class="fee-item">
          <span>里程费({{ orderDetail.distance }}公里)</span>
          <span>¥{{ orderDetail.distanceFee }}</span>
        </div>
        <div class="fee-item">
          <span>时长费({{ orderDetail.duration }}分钟)</span>
          <span>¥{{ orderDetail.timeFee }}</span>
        </div>
        <div class="fee-item discount" v-if="orderDetail.discount > 0">
          <span>优惠券抵扣</span>
          <span>-¥{{ orderDetail.discount }}</span>
        </div>
        <div class="fee-divider"></div>
        <div class="fee-item total">
          <span>实付金额</span>
          <span>¥{{ orderDetail.totalPrice }}</span>
        </div>
      </div>
    </div>

    <!-- 行程信息 -->
    <div class="trip-info-card">
      <h3 class="card-title">行程信息</h3>
      
      <div class="driver-info">
        <img :src="orderDetail.driverAvatar || '/default-avatar.png'" alt="" />
        <div class="info">
          <p class="name">{{ orderDetail.driverName }}</p>
          <p class="car">{{ orderDetail.plateNumber }}</p>
        </div>
        <button class="call-btn" @click="callDriver">
          <van-icon name="phone-o" size="16" />
        </button>
      </div>

      <div class="route-summary">
        <div class="route-item">
          <div class="dot from"></div>
          <span>{{ orderDetail.fromAddress || '出发地' }}</span>
        </div>
        <div class="route-line"></div>
        <div class="route-item">
          <div class="dot to"></div>
          <span>{{ orderDetail.toAddress || '目的地' }}</span>
        </div>
      </div>

      <div class="trip-stats">
        <div class="stat">
          <span class="value">{{ orderDetail.distance }}km</span>
          <span class="label">总里程</span>
        </div>
        <div class="stat">
          <span class="value">{{ orderDetail.duration }}min</span>
          <span class="label">总时长</span>
        </div>
      </div>
    </div>

    <!-- 支付方式 -->
    <div class="pay-method-card">
      <h3 class="card-title">支付方式</h3>
      <div class="method-list">
        <div 
          v-for="method in payMethods" 
          :key="method.id"
          class="method-item"
          :class="{ selected: selectedMethod === method.id }"
          @click="selectedMethod = method.id"
        >
          <component :is="method.icon" size="24" :color="method.color" />
          <span>{{ method.name }}</span>
          <van-icon 
            :name="selectedMethod === method.id ? 'success' : 'circle'"
            :color="selectedMethod === method.id ? '#7C3AED' : '#D1D5DB'"
            size="18"
          />
        </div>
      </div>
    </div>

    <!-- 底部支付按钮 -->
    <div class="bottom-bar safe-area-bottom">
      <button 
        class="btn-pay"
        :disabled="loading"
        @click="handlePay"
      >
        {{ loading ? '支付中...' : `确认支付 ¥${orderDetail.totalPrice}` }}
      </button>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { showToast, showLoadingToast, closeToast, showDialog } from 'vant'
import { useOrderStore } from '@/stores/order'
import { payOrder, getPaymentStatus } from '@/api/order'

const router = useRouter()
const orderStore = useOrderStore()

// 订单详情（模拟数据）
const orderDetail = ref({
  totalPrice: '20.6',
  baseDistance: 3,
  basePrice: '10.0',
  distance: 8.2,
  distanceFee: '8.2',
  duration: 25,
  timeFee: '4.0',
  discount: '-1.6',
  driverName: '张师傅',
  driverAvatar: '',
  plateNumber: '渝A·12345',
  fromAddress: '',
  toAddress: ''
})

const selectedMethod = ref('alipay')
const loading = ref(false)

// 支付方式列表
const payMethods = ref([
  { id: 'alipay', name: '支付宝', icon: 'van-icon', color: '#1677FF' },
  { id: 'wechat', name: '微信支付', icon: 'van-icon', color: '#07C160' },
  { id: 'balance', name: '余额支付', icon: 'van-icon', color: '#F59E0B' }
])

// 联系司机
const callDriver = () => {
  showDialog({
    title: '联系司机',
    message: '是否拨打司机电话？',
    showCancelButton: true
  }).then(() => {
    showToast('正在拨打电话...')
  }).catch(() => {})
}

// 确认支付
const handlePay = async () => {
  try {
    loading.value = true
    const toast = showLoadingToast({
      message: '正在支付...',
      forbidClick: true,
      duration: 0
    })

    // 调用支付接口
    await payOrder(orderStore.currentOrder?.orderId, selectedMethod.value)
    
    // 轮询支付结果
    let pollCount = 0
    const pollTimer = setInterval(async () => {
      pollCount++
      if (pollCount > 10) {
        clearInterval(pollTimer)
        closeToast()
        showToast('支付超时，请重试')
        return
      }

      const status = await getPaymentStatus(orderStore.currentOrder?.orderId)
      if (status === 'PAID') {
        clearInterval(pollTimer)
        closeToast()
        showToast('支付成功')
        
        // 跳转到支付成功页
        setTimeout(() => {
          router.replace('/order/success')
        }, 500)
      }
    }, 1000)

  } catch (error) {
    console.error('Payment error:', error)
    loading.value = false
    closeToast()
  }
}

const goBack = () => router.back()

onMounted(() => {
  // 加载订单详情
})
</script>

<style scoped>
.payment-page {
  min-height: 100vh;
  background: #f5f5f5;
  padding-bottom: 100px;
}

.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  background: white;
  position: sticky;
  top: 0;
  z-index: 10;
}

.title {
  font-size: 17px;
  font-weight: 600;
  color: var(--text-primary);
}

.price-card {
  margin: 16px;
  background: linear-gradient(135deg, #7C3AED 0%, #9333EA 100%);
  border-radius: 16px;
  padding: 30px 20px;
  text-align: center;
  color: white;
}

.price-card .label {
  font-size: 14px;
  opacity: 0.9;
  margin-bottom: 12px;
}

.price {
  font-size: 42px;
  font-weight: 700;
  margin-bottom: 24px;
}

.fee-detail {
  background: rgba(255, 255, 255, 0.15);
  border-radius: 10px;
  padding: 16px;
  text-align: left;
}

.fee-item {
  display: flex;
  justify-content: space-between;
  padding: 8px 0;
  font-size: 13px;
}

.fee-item.discount span:last-child {
  color: #FCD34D;
}

.fee-divider {
  height: 1px;
  background: rgba(255, 255, 255, 0.2);
  margin: 8px 0;
}

.fee-item.total {
  font-size: 15px;
  font-weight: 600;
}

.trip-info-card,
.pay-method-card {
  margin: 0 16px 16px;
  background: white;
  border-radius: 12px;
  padding: 16px;
}

.card-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 16px;
}

.driver-info {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 16px;
  padding-bottom: 16px;
  border-bottom: 1px solid #F3F4F6;
}

.driver-info img {
  width: 44px;
  height: 44px;
  border-radius: 50%;
  object-fit: cover;
  background: #E5E7EB;
}

.info .name {
  font-size: 15px;
  font-weight: 500;
  color: var(--text-primary);
}

.info .car {
  font-size: 12px;
  color: #6B7280;
  margin-top: 2px;
}

.call-btn {
  width: 36px;
  height: 36px;
  background: #EFF6FF;
  border: none;
  border-radius: 50%;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-left: auto;
}

.route-summary {
  position: relative;
  padding-left: 20px;
  margin-bottom: 16px;
}

.route-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 0;
  font-size: 14px;
  color: var(--text-primary);
}

.dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  position: absolute;
  left: -14px;
}

.dot.from { background: #10B981; }
.dot.to { background: #EF4444; }

.route-line {
  position: absolute;
  left: -10px;
  top: 18px;
  bottom: 22px;
  width: 2px;
  background: #D1D5DB;
}

.trip-stats {
  display: flex;
  justify-content: space-around;
  padding-top: 16px;
  border-top: 1px solid #F3F4F6;
}

.stat {
  text-align: center;
}

.stat .value {
  display: block;
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
}

.stat .label {
  display: block;
  font-size: 12px;
  color: #9CA3AF;
  margin-top: 4px;
}

.method-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.method-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px;
  border: 2px solid transparent;
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.2s;
}

.method-item.selected {
  background: #F5F3FF;
  border-color: #7C3AED;
}

.method-item span {
  flex: 1;
  font-size: 15px;
  color: var(--text-primary);
}

.bottom-bar {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  padding: 12px 20px;
  background: white;
  box-shadow: 0 -2px 10px rgba(0, 0, 0, 0.05);
}

.btn-pay {
  width: 100%;
  height: 48px;
  background: linear-gradient(135deg, #7C3AED 0%, #9333EA 100%);
  color: white;
  border: none;
  border-radius: 24px;
  font-size: 17px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  box-shadow: 0 4px 15px rgba(124, 58, 237, 0.3);
}

.btn-pay:active:not(:disabled) {
  transform: scale(0.98);
}

.btn-pay:disabled {
  opacity: 0.7;
}
</style>
