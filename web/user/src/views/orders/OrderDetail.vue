<template>
  <div class="order-detail-page">
    <!-- 顶部导航 -->
    <div class="page-header">
      <van-icon name="arrow-left" size="20" @click="goBack" />
      <span class="title">订单详情</span>
      <van-icon name="share-o" size="20" color="#6B7280" @click="shareOrder" />
    </div>

    <!-- 状态卡片 -->
    <div class="status-card" :class="getStatusClass(order.status)">
      <h2>{{ getStatusText(order.status) }}</h2>
      <p v-if="order.status === 'SEARCHING'">正在为您匹配司机，请稍候...</p>
      <p v-else-if="order.status === 'ACCEPTED'">司机已接单，正在前往您的位置</p>
      <p v-else-if="order.status === 'IN_PROGRESS'">行程进行中</p>
      <p v-else-if="order.status === 'COMPLETED'">感谢您选择花小龙打车</p>
      <p v-else-if="order.status === 'CANCELLED'">订单已取消</p>
    </div>

    <!-- 路线信息 -->
    <div class="route-card">
      <div class="route-info">
        <div class="route-item">
          <div class="dot from"></div>
          <div class="text">
            <p class="main">{{ order.fromAddress }}</p>
            <p class="sub">{{ order.createTime }} 上车</p>
          </div>
        </div>
        <div class="route-line"></div>
        <div class="route-item">
          <div class="dot to"></div>
          <div class="text">
            <p class="main">{{ order.toAddress }}</p>
            <p class="sub" v-if="order.arriveTime">{{ order.arriveTime }} 到达</p>
          </div>
        </div>
      </div>

      <!-- 行程统计 -->
      <div class="trip-stats" v-if="order.distance || order.duration">
        <div class="stat-item">
          <span class="value">{{ order.distance }}km</span>
          <span class="label">总里程</span>
        </div>
        <div class="stat-divider"></div>
        <div class="stat-item">
          <span class="value">{{ order.duration }}min</span>
          <span class="label">总时长</span>
        </div>
        <div class="stat-divider"></div>
        <div class="stat-item">
          <span class="value price">¥{{ order.totalPrice }}</span>
          <span class="label">费用</span>
        </div>
      </div>
    </div>

    <!-- 司机信息 -->
    <div class="driver-card" v-if="order.driverName">
      <h3>司机信息</h3>
      <div class="driver-info">
        <img :src="order.driverAvatar || '/default-avatar.png'" alt="" />
        <div class="info">
          <p class="name">{{ order.driverName }}</p>
          <p class="car">{{ order.plateNumber }} · {{ order.carModel }}</p>
          <div class="rating">
            <van-rate :model-value="order.driverRating || 4.9" readonly size="12" color="#F59E0B" void-icon="star" void-color="#E5E7EB" />
            <span>{{ order.driverRating || 4.9 }}分</span>
          </div>
        </div>
        <button class="call-btn" @click="callDriver">
          <van-icon name="phone-o" size="18" color="#3B82F6" />
        </button>
      </div>
    </div>

    <!-- 费用明细 -->
    <div class="fee-card" v-if="order.feeDetail">
      <h3>费用明细</h3>
      <div class="fee-list">
        <div class="fee-item" v-for="(item, index) in order.feeDetail" :key="index">
          <span>{{ item.name }}</span>
          <span :class="{ discount: item.amount < 0 }">
            {{ item.amount > 0 ? '+' : '' }}¥{{ Math.abs(item.amount) }}
          </span>
        </div>
        <div class="fee-divider"></div>
        <div class="fee-total">
          <span>实付金额</span>
          <span>¥{{ order.totalPrice }}</span>
        </div>
      </div>
    </div>

    <!-- 订单信息 -->
    <div class="info-card">
      <h3>订单信息</h3>
      <div class="info-list">
        <div class="info-item">
          <span>订单编号</span>
          <span>{{ order.orderNo }}</span>
        </div>
        <div class="info-item">
          <span>下单时间</span>
          <span>{{ order.createTime }}</span>
        </div>
        <div class="info-item">
          <span>支付方式</span>
          <span>{{ order.payMethod || '-' }}</span>
        </div>
        <div class="info-item" v-if="order.payTime">
          <span>支付时间</span>
          <span>{{ order.payTime }}</span>
        </div>
        <div class="info-item" v-if="order.couponAmount">
          <span>优惠券</span>
          <span>-¥{{ order.couponAmount }}</span>
        </div>
      </div>
    </div>

    <!-- 取消订单弹窗：预设原因直接提交，其他原因要求填写说明。 -->
    <van-dialog v-model:show="showCancelDialog" title="取消订单" show-cancel-button confirm-button-text="确定取消" cancel-button-text="再等等" :before-close="onCancelDialogClose" @confirm="confirmCancel">
      <div class="cancel-dialog-content">
        <p class="cancel-dialog-hint">请选择取消原因，帮助我们优化服务</p>
        <van-radio-group v-model="cancelReason" class="cancel-reasons">
          <van-radio name="行程有变，暂时不需要了">行程有变，暂时不需要了</van-radio>
          <van-radio name="等待时间太长">等待时间太长</van-radio>
          <van-radio name="价格太高">价格太高</van-radio>
          <van-radio name="other">其他原因</van-radio>
        </van-radio-group>
        <van-field v-if="cancelReason === 'other'" v-model="otherCancelReason" class="other-reason-field" type="textarea" rows="3" maxlength="100" show-word-limit placeholder="请填写取消原因" />
      </div>
    </van-dialog>
    <!-- 底部操作 -->
    <div class="bottom-bar safe-area-bottom" v-if="showActions || canCancel">
      <button v-if="canCancel" class="btn-cancel" :disabled="isCancelling" @click="cancelCurrentOrder">{{ isCancelling ? '正在取消...' : '取消订单' }}</button>
      <button 
        v-if="order.status === 'PENDING_PAYMENT'" 
        class="btn-primary"
        @click="goToPay"
      >
        立即支付
      </button>
      <button 
        v-if="order.status === 'COMPLETED' && !order.rated" 
        class="btn-primary"
        @click="goToRate"
      >
        去评价
      </button>
      <button 
        v-if="['COMPLETED', 'CANCELLED'].includes(order.status)" 
        class="btn-secondary"
        @click="reOrder"
      >
        再次叫车
      </button>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { showToast, showDialog, showLoadingToast, closeToast } from 'vant'
import { getOrderDetail, cancelOrder } from '@/api/order'

const router = useRouter()
const route = useRoute()

const orderId = Number(route.params.id)

// 订单详情数据由接口返回，禁止使用固定演示订单覆盖真实订单。
const order = ref({ id: orderId, status: 'SEARCHING', createTime: '--', arriveTime: '', fromAddress: '加载中...', toAddress: '加载中...', distance: '', duration: '', totalPrice: '0.00', driverName: '', driverAvatar: '', plateNumber: '', carModel: '', driverRating: 0, orderNo: '', payMethod: '', payTime: '', couponAmount: '', rated: false, feeDetail: [] })

// 将后端数字状态转换成详情页展示状态。
// 后端状态：1待接单 2已接单 3行程中 4待支付 5已完成 6已取消
const normalizeStatus = (status) => ({ 1: 'SEARCHING', 2: 'ACCEPTED', 3: 'IN_PROGRESS', 4: 'PENDING_PAYMENT', 5: 'COMPLETED', 6: 'CANCELLED' })[status] || status

// 将订单详情接口字段转换为页面展示模型。
const mapOrderDetail = (item) => {
  const distanceM = Number(item.estimatedDistanceM || 0)
  const durationS = Number(item.estimatedDurationS || 0)
  return {
    ...item,
    id: item.orderId || orderId,
    status: normalizeStatus(item.status),
    createTime: item.createdAt ? new Date(Number(item.createdAt) * 1000).toLocaleString() : '--',
    fromAddress: item.fromAddress || '未填写上车点',
    toAddress: item.toAddress || '未填写目的地',
    distance: distanceM ? (distanceM / 1000).toFixed(1) : '',
    duration: durationS ? Math.ceil(durationS / 60) : '',
    totalPrice: (Number(item.estimatedPriceCents || 0) / 100).toFixed(2),
    driverName: item.driverName || '',
    driverRating: Number(item.driverRating || 0),
    feeDetail: []
  }
}

// 是否显示操作按钮
const isCancelling = ref(false)
const showCancelDialog = ref(false)
const cancelReason = ref('行程有变，暂时不需要了')
const otherCancelReason = ref('')
const canCancel = computed(() => !isCancelling.value && ['SEARCHING', 'ACCEPTED'].includes(order.value.status))

const showActions = computed(() => {
  return ['PENDING_PAYMENT', 'COMPLETED', 'CANCELLED'].includes(order.value.status)
})

// 状态相关方法
const getStatusText = (status) => {
  const map = {
    SEARCHING: '等待接单',
    ACCEPTED: '司机已接单',
    PICKING_UP: '司机已出发',
    IN_PROGRESS: '行程中',
    PENDING_PAYMENT: '待支付',
    COMPLETED: '已完成',
    CANCELLED: '已取消'
  }
  return map[status] || status
}

const getStatusClass = (status) => {
  if (['SEARCHING', 'ACCEPTED', 'PICKING_UP', 'IN_PROGRESS'].includes(status)) return 'ongoing'
  if (status === 'COMPLETED') return 'completed'
  if (status === 'PENDING_PAYMENT') return 'pending'
  return 'cancelled'
}

// 操作方法
const goBack = () => router.back()
const shareOrder = () => showToast('分享功能开发中')
const callDriver = () => showDialog({ title: '联系司机', message: '是否拨打电话？', showCancelButton: true })
const goToPay = () => router.push(`/order/payment?orderId=${orderId}`)
const goToRate = () => router.push(`/order/success?orderId=${orderId}`)
const reOrder = () => router.push('/order/create')
// 点击"取消订单"按钮：打开取消原因选择弹窗。
const cancelCurrentOrder = () => {
  if (isCancelling.value || !canCancel.value) return
  cancelReason.value = '行程有变，暂时不需要了'
  otherCancelReason.value = ''
  showCancelDialog.value = true
}

// 弹窗关闭前校验：选择"其他原因"时必须填写说明，否则阻止关闭。
const onCancelDialogClose = (action) => {
  if (action !== 'confirm') return true
  if (cancelReason.value === 'other' && !otherCancelReason.value.trim()) {
    showToast('请填写取消原因')
    return false
  }
  return true
}

// 确定取消：预设原因直接提交对应文本，其他原因提交输入框内容。
const confirmCancel = async () => {
  if (isCancelling.value) return
  const reason = cancelReason.value === 'other' ? otherCancelReason.value.trim() : cancelReason.value
  if (!reason) {
    showToast('请填写取消原因')
    return
  }
  isCancelling.value = true
  showLoadingToast({ message: '正在取消...', forbidClick: true, duration: 0 })
  try {
    await cancelOrder(orderId, reason)
    closeToast()
    order.value.status = 'CANCELLED'
    showToast('订单已取消，优惠券已返还')
  } catch (error) {
    closeToast()
    showToast(error?.response?.data?.message || '取消订单失败')
  } finally {
    isCancelling.value = false
  }
}

onMounted(async () => {
  try {
    // 只展示当前路由对应订单的真实详情。
    const res = await getOrderDetail(orderId)
    if (!res) throw new Error('订单详情为空')
    Object.assign(order.value, mapOrderDetail(res))
  } catch (error) {
    console.error('加载订单详情失败:', error)
    showToast('订单详情加载失败，请稍后重试')
  }
})
</script>

<style scoped>
.order-detail-page {
  min-height: 100vh;
  background: #f5f5f5;
  padding-bottom: 80px;
}

.page-header {
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

.status-card {
  padding: 24px 20px;
  text-align: center;
  margin: 16px;
  border-radius: 12px;
}

.status-card.ongoing {
  background: linear-gradient(135deg, #EFF6FF 0%, #DBEAFE 100%);
}

.status-card.completed {
  background: linear-gradient(135deg, #ECFDF5 0%, #D1FAE5 100%);
}

.status-card.pending {
  background: linear-gradient(135deg, #FEF3C7 0%, #FDE68A 100%);
}

.status-card.cancelled {
  background: linear-gradient(135deg, #FEE2E2 0%, #FECACA 100%);
}

.status-card h2 {
  font-size: 20px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 8px;
}

.status-card p {
  font-size: 14px;
  color: #6B7280;
}

.route-card,
.driver-card,
.fee-card,
.info-card {
  margin: 0 16px 12px;
  background: white;
  border-radius: 12px;
  padding: 16px;
}

.route-card h3,
.driver-card h3,
.fee-card h3,
.info-card h3 {
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 14px;
}

.route-info {
  position: relative;
  padding-left: 20px;
  margin-bottom: 16px;
}

.route-item {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 10px 0;
}

.dot {
  width: 12px;
  height: 12px;
  border-radius: 50%;
  position: absolute;
  left: -17px;
  top: 14px;
}

.dot.from { background: #10B981; }
.dot.to { background: #EF4444; }

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
  left: -13px;
  top: 28px;
  bottom: 30px;
  width: 2px;
  background: #D1D5DB;
}

.trip-stats {
  display: flex;
  align-items: center;
  justify-content: space-around;
  padding-top: 16px;
  border-top: 1px solid #F3F4F6;
}

.stat-item {
  text-align: center;
}

.stat-item .value {
  display: block;
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
}

.stat-item .value.price {
  color: #EF4444;
}

.stat-item .label {
  display: block;
  font-size: 11px;
  color: #9CA3AF;
  margin-top: 4px;
}

.stat-divider {
  width: 1px;
  height: 32px;
  background: #E5E7EB;
}

.driver-info {
  display: flex;
  align-items: center;
  gap: 12px;
}

.driver-info img {
  width: 50px;
  height: 50px;
  border-radius: 50%;
  object-fit: cover;
  background: #E5E7EB;
}

.info {
  flex: 1;
}

.info .name {
  font-size: 15px;
  font-weight: 500;
  color: var(--text-primary);
  margin-bottom: 4px;
}

.info .car {
  font-size: 12px;
  color: #6B7280;
  margin-bottom: 4px;
}

.rating {
  display: flex;
  align-items: center;
  gap: 6px;
}

.rating span {
  font-size: 12px;
  color: #F59E0B;
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
}

.fee-list,
.info-list {
  display: flex;
  flex-direction: column;
}

.fee-item,
.info-item {
  display: flex;
  justify-content: space-between;
  padding: 10px 0;
  font-size: 14px;
  color: var(--text-primary);
}

.fee-item .discount {
  color: #059669;
}

.fee-divider {
  height: 1px;
  background: #F3F4F6;
  margin: 8px 0;
}

.fee-total {
  display: flex;
  justify-content: space-between;
  padding-top: 8px;
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary);
}

.info-item span:first-child {
  color: #6B7280;
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

.btn-primary {
  width: 100%;
  height: 48px;
  background: linear-gradient(135deg, #7C3AED 0%, #9333EA 100%);
  color: white;
  border: none;
  border-radius: 24px;
  font-size: 16px;
  font-weight: 500;
  cursor: pointer;
}

.btn-cancel { width: 100%; height: 44px; margin-bottom: 8px; border: 1px solid #FECACA; border-radius: 22px; color: #DC2626; background: #FEF2F2; font-size: 15px; }

.btn-secondary {
  width: 100%;
  height: 48px;
  background: white;
  border: 1px solid #D1D5DB;
  border-radius: 24px;
  font-size: 16px;
  color: #6B7280;
  cursor: pointer;
}
</style>

