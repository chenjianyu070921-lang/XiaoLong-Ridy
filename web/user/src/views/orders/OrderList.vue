<template>
  <div class="order-list-page">
    <!-- 顶部导航 -->
    <div class="page-header">
      <van-icon name="arrow-left" size="20" @click="goBack" />
      <span class="title">我的订单</span>
      <div style="width: 20px;"></div>
    </div>

    <!-- 筛选标签 -->
    <div class="filter-tabs">
      <button 
        v-for="tab in tabs" 
        :key="tab.value"
        class="tab-btn"
        :class="{ active: activeTab === tab.value }"
        @click="activeTab = tab.value"
      >
        {{ tab.label }}
      </button>
    </div>

    <!-- 订单列表 -->
    <div class="order-list" v-if="orders.length > 0">
      <div 
        v-for="order in filteredOrders" 
        :key="order.id"
        class="order-card"
        @click="goToDetail(order.id)"
      >
        <div class="card-header">
          <div class="status-tag" :class="getStatusClass(order.status)">
            {{ getStatusText(order.status) }}
          </div>
          <span class="time">{{ order.createTime }}</span>
        </div>

        <div class="route-info">
          <div class="route-item">
            <div class="dot from"></div>
            <span>{{ order.fromAddress }}</span>
          </div>
          <div class="route-line"></div>
          <div class="route-item">
            <div class="dot to"></div>
            <span>{{ order.toAddress }}</span>
          </div>
        </div>

        <div class="card-footer">
          <div class="car-type">{{ order.carTypeName }}</div>
          <div class="price">¥{{ order.totalPrice }}</div>
        </div>

        <!-- 操作按钮 -->
        <div class="actions" v-if="showActions(order.status)">
          <button 
            v-if="order.status === 'PENDING_PAYMENT'" 
            class="btn-pay"
            @click.stop="goToPay(order.id)"
          >
            去支付
          </button>
          <button 
            v-if="order.status === 'COMPLETED' && !order.rated" 
            class="btn-rate"
            @click.stop="goToRate(order.id)"
          >
            去评价
          </button>
          <button 
            v-if="['COMPLETED', 'CANCELLED', 'REFUNDED'].includes(order.status)"
            class="btn-reorder"
            @click.stop="reOrder(order)"
          >
            再次叫车
          </button>
        </div>
      </div>
    </div>

    <!-- 空状态 -->
    <div class="empty-state" v-else>
      <van-icon name="orders-o" size="80" color="#D1D5DB" />
      <p>暂无订单记录</p>
      <button class="btn-primary" @click="goHome">去叫车</button>
    </div>

    <!-- 底部导航栏 -->
    <van-tabbar v-model="activeNav" active-color="#7C3AED" inactive-color="#6B7280">
      <van-tabbar-item icon="home-o" name="home" @click="goHome">首页</van-tabbar-item>
      <van-tabbar-item icon="orders-o" name="orders">订单</van-tabbar-item>
      <van-tabbar-item icon="user-o" name="profile" @click="goProfile">我的</van-tabbar-item>
    </van-tabbar>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { showToast, showLoadingToast, closeToast } from 'vant'
import { getOrders, getOrderDetail } from '@/api/order'
import { useOrderStore } from '@/stores/order'
import { ORDER_STATUS, getOrderStatusText, normalizeOrderStatus } from '@/constants/order'

const router = useRouter()
const orderStore = useOrderStore()

// 标签页
const tabs = [
  { label: '全部', value: 'all' },
  { label: '进行中', value: 'ongoing' },
  { label: '已完成', value: 'completed' },
  { label: '已取消', value: 'cancelled' }
]

const activeTab = ref('all')
const activeNav = ref('orders')
const orders = ref([])

// 过滤后的订单
const filteredOrders = computed(() => {
  if (activeTab.value === 'all') return orders.value
  
  const statusMap = {
    ongoing: [ORDER_STATUS.SEARCHING, ORDER_STATUS.ACCEPTED, ORDER_STATUS.IN_PROGRESS, ORDER_STATUS.PENDING_PAYMENT],
    completed: [ORDER_STATUS.COMPLETED],
    cancelled: [ORDER_STATUS.CANCELLED, ORDER_STATUS.REFUNDED]
  }
  
  return orders.value.filter(o => statusMap[activeTab.value]?.includes(o.status))
})

// 获取状态文本
const getStatusText = (status) => {
  return getOrderStatusText(status)
}

// 获取状态样式类
const getStatusClass = (status) => {
  if ([ORDER_STATUS.SEARCHING, ORDER_STATUS.ACCEPTED, ORDER_STATUS.IN_PROGRESS].includes(status)) {
    return 'ongoing'
  }
  if (status === 'COMPLETED') return 'completed'
  if (status === 'PENDING_PAYMENT') return 'pending'
  return 'cancelled'
}

// 是否显示操作按钮
const showActions = (status) => {
  return [ORDER_STATUS.PENDING_PAYMENT, ORDER_STATUS.COMPLETED, ORDER_STATUS.CANCELLED, ORDER_STATUS.REFUNDED].includes(status)
}

// 加载真实订单列表，包含已取消订单，避免使用固定演示数据覆盖接口结果。
const loadOrders = async () => {
  try {
    showLoadingToast({ message: '加载中...', forbidClick: true, duration: 0 })
    const res = await getOrders({ page: 1, pageSize: 20, status: 0 })
    const list = Array.isArray(res?.list) ? res.list : []
    orders.value = list.map(item => ({
      id: item.orderId,
      status: normalizeOrderStatus(item.status),
      createTime: item.createdAt ? new Date(Number(item.createdAt) * 1000).toLocaleString() : '--',
      fromAddress: item.fromAddress || '未填写上车点',
      toAddress: item.toAddress || '未填写目的地',
      carTypeName: ({ 1: '特惠快车', 2: '快车', 3: '专车' })[item.carType] || '快车',
      totalPrice: (Number(item.estimatedPriceCents || 0) / 100).toFixed(2),
      rated: Boolean(item.rated)
    }))
  } catch (error) {
    console.error('加载订单列表失败:', error)
    orders.value = []
  } finally {
    closeToast()
  }
}

// 导航方法
const goBack = () => router.back()
const goHome = () => router.replace('/home')
const goProfile = () => router.replace('/profile')
const goToDetail = async (id) => {
  // 进行中的订单必须恢复 currentOrder，等待页和行程页依赖该上下文继续轮询。
  const order = orders.value.find(item => item.id === id)
  if (order && [ORDER_STATUS.SEARCHING, ORDER_STATUS.ACCEPTED, ORDER_STATUS.IN_PROGRESS].includes(order.status)) {
    try {
      const detail = await getOrderDetail(id)
      orderStore.setCurrentOrder({ ...detail, orderId: detail?.orderId || id })
    } catch (error) {
      console.warn('恢复进行中订单失败:', error)
      return
    }
    const target = { [ORDER_STATUS.SEARCHING]: '/order/waiting', [ORDER_STATUS.ACCEPTED]: '/order/driver-coming', [ORDER_STATUS.IN_PROGRESS]: '/order/trip' }[order.status]
    router.push(target)
    return
  }
  router.push(`/orders/${id}`)
}
const goToPay = (id) => router.push(`/order/payment?orderId=${id}`)
const goToRate = (id) => router.push(`/order/success?orderId=${id}`)

// 再次叫车
const reOrder = (order) => {
  showToast('正在为您创建新订单...')
  // 可以预填目的地等信息
  setTimeout(() => {
    router.push('/order/create')
  }, 500)
}

onMounted(() => {
  loadOrders()
})
</script>

<style scoped>
.order-list-page {
  min-height: 100vh;
  background: #f5f5f5;
  padding-bottom: 60px;
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

.filter-tabs {
  display: flex;
  gap: 8px;
  padding: 12px 16px;
  background: white;
  overflow-x: auto;
}

.tab-btn {
  padding: 6px 16px;
  border-radius: 20px;
  font-size: 13px;
  color: #6B7280;
  background: #F3F4F6;
  border: none;
  cursor: pointer;
  white-space: nowrap;
  transition: all 0.2s;
}

.tab-btn.active {
  background: linear-gradient(135deg, #7C3AED 0%, #9333EA 100%);
  color: white;
}

.order-list {
  padding: 0 16px;
}

.order-card {
  background: white;
  border-radius: 12px;
  padding: 16px;
  margin-bottom: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
  cursor: pointer;
  transition: all 0.2s;
}

.order-card:active {
  transform: scale(0.98);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 14px;
}

.status-tag {
  padding: 4px 10px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 500;
}

.status-tag.ongoing {
  background: #EFF6FF;
  color: #3B82F6;
}

.status-tag.completed {
  background: #ECFDF5;
  color: #059669;
}

.status-tag.pending {
  background: #FEF3C7;
  color: #D97706;
}

.status-tag.cancelled {
  background: #FEE2E2;
  color: #DC2626;
}

.time {
  font-size: 12px;
  color: #9CA3AF;
}

.route-info {
  position: relative;
  padding-left: 18px;
  margin-bottom: 14px;
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
  left: -12px;
}

.dot.from { background: #10B981; }
.dot.to { background: #EF4444; }

.route-line {
  position: absolute;
  left: -8px;
  top: 22px;
  bottom: 26px;
  width: 2px;
  background: #D1D5DB;
}

.card-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-top: 12px;
  border-top: 1px solid #F3F4F6;
}

.car-type {
  font-size: 13px;
  color: #6B7280;
}

.price {
  font-size: 18px;
  font-weight: 700;
  color: #EF4444;
}

.actions {
  display: flex;
  gap: 10px;
  margin-top: 14px;
  padding-top: 14px;
  border-top: 1px solid #F3F4F6;
}

.btn-pay,
.btn-rate,
.btn-reorder {
  flex: 1;
  height: 36px;
  border-radius: 18px;
  font-size: 13px;
  cursor: pointer;
  border: none;
  transition: all 0.2s;
}

.btn-pay {
  background: linear-gradient(135deg, #7C3AED 0%, #9333EA 100%);
  color: white;
}

.btn-rate {
  background: #FEF3C7;
  color: #D97706;
}

.btn-reorder {
  background: #EFF6FF;
  color: #3B82F6;
}

.empty-state {
  text-align: center;
  padding: 80px 24px;
}

.empty-state p {
  font-size: 15px;
  color: #6B7280;
  margin: 20px 0 30px;
}

.empty-state button {
  min-width: 160px;
}
</style>
