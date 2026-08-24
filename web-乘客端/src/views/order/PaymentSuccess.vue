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
      <button class="btn-primary submit-btn" @click="submitRating">
        提交评价
      </button>
      <button class="btn-secondary" @click="goHome">
        返回首页
      </button>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { showToast } from 'vant'
import { useOrderStore } from '@/stores/order'

const router = useRouter()
const orderStore = useOrderStore()

// 订单详情
const orderDetail = ref({
  totalPrice: '20.6',
  orderNo: 'XL202401201234567890',
  payTime: new Date().toLocaleString(),
  driverName: '张师傅',
  driverId: ''
})

const rating = ref(0)
const comment = ref('')
const payMethodName = ref('支付宝')

onMounted(() => {
  // 加载订单数据
})

// 提交评价
const submitRating = () => {
  if (rating.value === 0) {
    showToast('请选择评分')
    return
  }
  
  showToast('感谢您的评价')
  
  // 跳转到订单列表或首页
  setTimeout(() => {
    router.replace('/orders')
  }, 1000)
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
