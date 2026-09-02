<template>
  <div class="success-page">
    <!-- 成功图标 -->
    <div class="success-icon animate-fadeIn">
      <van-icon name="passed" size="80" color="#10B981" />
    </div>

    <h1 class="title">支付成功</h1>
    <p class="amount">¥{{ orderDetail.totalPrice }}</p>

    <!-- 订单信息 -->
    <div class="order-info">
      <div class="info-item">
        <span class="label">支付方式</span>
        <span class="value">{{ payMethodName }}</span>
      </div>
      <div class="info-item">
        <span class="label">订单号</span>
        <span class="value">{{ orderDetail.orderNo }}</span>
      </div>
      <div class="info-item">
        <span class="label">支付时间</span>
        <span class="value">{{ orderDetail.payTime }}</span>
      </div>
    </div>

    <!-- 司机评分 -->
    <div class="rating-card">
      <h3>为司机评分</h3>
      <p class="driver-name">{{ orderDetail.driverName }}</p>
      
      <van-rate 
        v-model="rating" 
        size="32" 
        color="#F59E0B" 
        void-icon="star" 
        void-color="#E5E7EB"
        :count="5"
      />
      
      <textarea
        v-model="comment"
        placeholder="您的评价对我们很重要（选填）"
        maxlength="100"
        rows="3"
        class="comment-input"
      ></textarea>
    </div>

    <!-- 操作按钮 -->
    <div class="actions">
      <button class="btn-primary submit-btn" :disabled="submitting" @click="submitRating">
        {{ submitting ? '提交中...' : '提交评价' }}
      </button>
      <button class="btn-secondary" @click="goHome">
        返回首页
      </button>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { showToast, showLoadingToast, closeToast } from 'vant'
import { getOrderDetail, submitReview } from '@/api/order'

const router = useRouter()
const route = useRoute()

// 订单详情
const orderDetail = ref({
  totalPrice: '0.00',
  orderNo: '--',
  payTime: '--',
  driverName: '司机信息加载中',
  driverId: ''
})

const rating = ref(0)
const comment = ref('')
const payMethodName = ref('--')
const submitting = ref(false)
const orderId = Number(route.query.orderId)

// payMethodText 将支付页路由参数转换为乘客可读的支付方式名称。
const payMethodText = (payMethod) => ({ wechat: '微信支付', alipay: '支付宝', balance: '余额支付' })[payMethod] || '--'

// mapOrderDetail 将后端金额和司机标识转换为页面展示模型，禁止使用演示数据替代真实订单。
const mapOrderDetail = (data) => ({
  totalPrice: (Number(data?.paidCents || data?.payableCents || data?.estimatedPriceCents || 0) / 100).toFixed(2),
  orderNo: data?.orderNo || '--',
  // ordersvc 暂未提供支付完成时间，使用订单最近状态更新时间作为可追溯的完成时间。
  payTime: data?.updatedAt ? new Date(Number(data.updatedAt) * 1000).toLocaleString() : '--',
  driverName: data?.driverId ? `司机 #${data.driverId}` : '司机信息暂不可用',
  driverId: data?.driverId || ''
})

onMounted(async () => {
  payMethodName.value = payMethodText(route.query.payMethod)
  if (!Number.isInteger(orderId) || orderId <= 0) {
    showToast('未找到支付订单')
    return
  }
  try {
    const detail = await getOrderDetail(orderId)
    orderDetail.value = mapOrderDetail(detail)
  } catch (error) {
    console.error('加载支付成功订单失败:', error)
    showToast('订单详情加载失败')
  }
})

// 提交评价
const submitRating = async () => {
  if (rating.value === 0) {
    showToast('请选择评分')
    return
  }
  if (!Number.isInteger(orderId) || orderId <= 0 || submitting.value) return

  submitting.value = true
  showLoadingToast({ message: '正在提交评价...', forbidClick: true, duration: 0 })
  try {
    await submitReview({ orderId, rating: rating.value, comment: comment.value.trim(), tags: '' })
    closeToast()
    showToast('感谢您的评价')
    router.replace('/orders')
  } catch (error) {
    closeToast()
    console.error('提交评价失败:', error)
  } finally {
    submitting.value = false
  }
}

// 返回首页
const goHome = () => {
  router.replace('/home')
}
</script>

<style scoped>
.success-page {
  min-height: 100vh;
  background: #f5f5f5;
  padding: 60px 24px 40px;
  text-align: center;
}

.success-icon {
  margin-bottom: 20px;
}

.title {
  font-size: 24px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 12px;
}

.amount {
  font-size: 36px;
  font-weight: 700;
  color: #EF4444;
  margin-bottom: 30px;
}

.order-info {
  background: white;
  border-radius: 12px;
  padding: 20px;
  margin-bottom: 20px;
  text-align: left;
}

.info-item {
  display: flex;
  justify-content: space-between;
  padding: 12px 0;
  border-bottom: 1px solid #F3F4F6;
}

.info-item:last-child {
  border-bottom: none;
}

.label {
  font-size: 14px;
  color: #6B7280;
}

.value {
  font-size: 14px;
  color: var(--text-primary);
  font-weight: 500;
}

.rating-card {
  background: white;
  border-radius: 12px;
  padding: 24px 20px;
  margin-bottom: 30px;
}

.rating-card h3 {
  font-size: 17px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 8px;
}

.driver-name {
  font-size: 14px;
  color: #6B7280;
  margin-bottom: 16px;
}

.comment-input {
  width: 100%;
  margin-top: 16px;
  padding: 12px;
  border: 1px solid #E5E7EB;
  border-radius: 8px;
  font-size: 14px;
  resize: none;
  outline: none;
  transition: all 0.2s;
}

.comment-input:focus {
  border-color: #7C3AED;
  box-shadow: 0 0 0 2px rgba(124, 58, 237, 0.1);
}

.actions {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.submit-btn,
.btn-secondary {
  width: 100%;
  height: 48px;
  border-radius: 24px;
  font-size: 16px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.submit-btn {
  background: linear-gradient(135deg, #7C3AED 0%, #9333EA 100%);
  color: white;
  border: none;
  box-shadow: 0 4px 15px rgba(124, 58, 237, 0.3);
}

.submit-btn:active {
  transform: scale(0.98);
}

.btn-secondary {
  background: white;
  color: #6B7280;
  border: 1px solid #D1D5DB;
}

.btn-secondary:active {
  background: #F9FAFB;
}
</style>
