<template>
  <section class="h5-panel">
    <section class="nearby-order-section">
      <div class="section-title">
        <h2>附近订单</h2>
        <div class="section-actions">
          <button type="button" :disabled="nearbyOrderLoading" @click="emit('load-nearby-orders', nearbyOrderPage)">刷新</button>
          <button type="button" @click="emit('open-nearby-popup')">展开</button>
        </div>
      </div>

      <div v-if="nearbyOrderLoading" class="home-order-loading"><van-loading size="20px" /></div>
      <div v-else-if="nearbyOrders.length === 0" class="home-order-empty">暂无附近可接订单</div>
      <article v-for="order in nearbyOrders" :key="'nearby-' + resolveOrderId(order)" class="order-card nearby-order-card">
        <div class="order-heading">
          <strong>{{ order.orderNo || '订单 ' + resolveOrderId(order) }}</strong>
          <span class="order-distance">距您 {{ formatDistance(order.distanceMeters) }} 公里</span>
        </div>
        <p class="route-line">{{ order.fromAddress || '--' }} -> {{ order.toAddress || '--' }}</p>
        <div class="meta-row">
          <span>{{ formatPrice(order.estimatedPriceCents) }}</span>
          <span>{{ formatTime(order.createdAt) }}</span>
        </div>
        <div class="order-actions">
          <button type="button" @click="emit('order-detail', resolveOrderId(order))">详情</button>
          <button v-if="canAccept(order)" type="button" class="primary" @click="emit('order-action', 'accept', order)">接单</button>
          <button v-if="canAccept(order)" type="button" @click="emit('order-action', 'reject', order)">拒单</button>
        </div>
      </article>
      <div class="pager">
        <button type="button" :disabled="nearbyOrderPage <= 1 || nearbyOrderLoading" @click="emit('load-nearby-orders', nearbyOrderPage - 1)">上一页</button>
        <span>{{ nearbyOrderPage }} / {{ nearbyOrderPageCount }}</span>
        <button type="button" :disabled="nearbyOrderPage >= nearbyOrderPageCount || nearbyOrderLoading" @click="emit('load-nearby-orders', nearbyOrderPage + 1)">下一页</button>
      </div>
    </section>

    <div class="section-title">
      <h2>订单</h2>
      <button type="button" @click="emit('load-orders', orderPage)">刷新</button>
    </div>
    <div class="filter-bar">
      <van-dropdown-menu>
        <van-dropdown-item v-model="orderModeModel" :options="orderModeOptions" @change="emit('load-orders', 1)" />
        <van-dropdown-item v-model="orderStatusModel" :options="orderStatusOptions" @change="emit('load-orders', 1)" />
      </van-dropdown-menu>
    </div>
    <div v-if="orders.length === 0" class="empty-state">--</div>
    <article v-for="order in orders" :key="String(order.source || 'order') + '-' + String(resolveOrderId(order))" class="order-card">
      <div class="order-heading">
        <strong>{{ order.orderNo || '订单 ' + resolveOrderId(order) }}</strong>
        <van-tag>{{ order.source === 'dispatch' ? formatDispatchStatus(order.dispatchStatus) : formatOrderStatus(order.status) }}</van-tag>
      </div>
      <p class="route-line">{{ order.fromAddress || '--' }} -> {{ order.toAddress || '--' }}</p>
      <div class="meta-row">
        <span>{{ formatPrice(order.estimatedPriceCents) }}</span>
        <span v-if="order.source === 'available'">距您 {{ formatDistance(order.distanceMeters) }} 公里</span>
        <span>{{ formatTime(order.createdAt) }}</span>
      </div>
      <div class="order-actions">
        <button type="button" @click="emit('order-detail', resolveOrderId(order))">详情</button>
        <button type="button" @click="emit('select-trajectory', resolveOrderId(order))">轨迹</button>
        <button v-if="canAccept(order)" type="button" class="primary" @click="emit('order-action', 'accept', order)">接单</button>
        <button v-if="canAccept(order)" type="button" @click="emit('order-action', 'reject', order)">拒单</button>
        <button v-if="Number(order.status) === 2" type="button" @click="emit('order-action', 'confirm-arrive', order)">到达</button>
        <button v-if="Number(order.status) === 2" type="button" @click="emit('order-action', 'start-trip', order)">开始</button>
        <button v-if="Number(order.status) === 3" type="button" class="primary" @click="emit('open-finish', order)">结束</button>
      </div>
    </article>
    <div class="pager">
      <button type="button" :disabled="orderPage <= 1" @click="emit('load-orders', orderPage - 1)">上一页</button>
      <span>{{ orderPage }} / {{ Math.max(1, Math.ceil(orderTotal / orderPageSize)) }}</span>
      <button type="button" :disabled="orderPage * orderPageSize >= orderTotal" @click="emit('load-orders', orderPage + 1)">下一页</button>
    </div>

    <van-popup v-model:show="popupVisibleModel" teleport="#driver-home-popups" round position="bottom" class="nearby-order-popup" :style="nearbyOrderPopupStyle">
      <section class="nearby-order-sheet">
        <div class="heatmap-sheet-grabber" aria-hidden="true"></div>
        <div class="section-title">
          <h2>附近可接订单</h2>
          <button type="button" :disabled="nearbyOrderExpandedLoading" @click="emit('load-nearby-expanded-orders', nearbyOrderExpandedPage)">刷新</button>
        </div>
        <div v-if="nearbyOrderExpandedLoading" class="home-order-loading"><van-loading size="20px" /></div>
        <div v-else-if="nearbyOrderExpandedOrders.length === 0" class="home-order-empty">暂无附近可接订单</div>
        <div v-else class="nearby-order-popup-list">
          <article v-for="order in nearbyOrderExpandedOrders" :key="'nearby-expanded-' + resolveOrderId(order)" class="order-card nearby-order-card">
            <div class="order-heading">
              <strong>{{ order.orderNo || '订单 ' + resolveOrderId(order) }}</strong>
              <span class="order-distance">距您 {{ formatDistance(order.distanceMeters) }} 公里</span>
            </div>
            <p class="route-line">{{ order.fromAddress || '--' }} -> {{ order.toAddress || '--' }}</p>
            <div class="meta-row">
              <span>{{ formatPrice(order.estimatedPriceCents) }}</span>
              <span>{{ formatTime(order.createdAt) }}</span>
            </div>
            <div class="order-actions">
              <button type="button" @click="emit('order-detail', resolveOrderId(order))">详情</button>
              <button v-if="canAccept(order)" type="button" class="primary" @click="emit('order-action', 'accept', order)">接单</button>
              <button v-if="canAccept(order)" type="button" @click="emit('order-action', 'reject', order)">拒单</button>
            </div>
          </article>
        </div>
        <div class="pager">
          <button type="button" :disabled="nearbyOrderExpandedPage <= 1 || nearbyOrderExpandedLoading" @click="emit('load-nearby-expanded-orders', nearbyOrderExpandedPage - 1)">上一页</button>
          <span>{{ nearbyOrderExpandedPage }} / {{ nearbyOrderExpandedPageCount }}</span>
          <button type="button" :disabled="nearbyOrderExpandedPage >= nearbyOrderExpandedPageCount || nearbyOrderExpandedLoading" @click="emit('load-nearby-expanded-orders', nearbyOrderExpandedPage + 1)">下一页</button>
        </div>
      </section>
    </van-popup>
  </section>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  orders: { type: Array, required: true },
  orderMode: { type: String, required: true },
  orderStatus: { type: Number, required: true },
  orderModeOptions: { type: Array, required: true },
  orderStatusOptions: { type: Array, required: true },
  orderPage: { type: Number, required: true },
  orderPageSize: { type: Number, required: true },
  orderTotal: { type: Number, required: true },
  nearbyOrders: { type: Array, required: true },
  nearbyOrderLoading: { type: Boolean, required: true },
  nearbyOrderPage: { type: Number, required: true },
  nearbyOrderPageSize: { type: Number, required: true },
  nearbyOrderTotal: { type: Number, required: true },
  nearbyOrderPopupVisible: { type: Boolean, required: true },
  nearbyOrderExpandedOrders: { type: Array, required: true },
  nearbyOrderExpandedLoading: { type: Boolean, required: true },
  nearbyOrderExpandedPage: { type: Number, required: true },
  nearbyOrderExpandedPageSize: { type: Number, required: true },
  nearbyOrderExpandedTotal: { type: Number, required: true },
  canAccept: { type: Function, required: true },
  formatPrice: { type: Function, required: true },
  formatDistance: { type: Function, required: true },
  formatTime: { type: Function, required: true },
  formatOrderStatus: { type: Function, required: true },
  formatDispatchStatus: { type: Function, required: true }
})

const emit = defineEmits([
  'update:orderMode',
  'update:orderStatus',
  'update:nearby-order-popup-visible',
  'load-orders',
  'load-nearby-orders',
  'load-nearby-expanded-orders',
  'open-nearby-popup',
  'order-detail',
  'select-trajectory',
  'order-action',
  'open-finish'
])

const orderModeModel = computed({
  get: () => props.orderMode,
  set: (value) => emit('update:orderMode', value)
})

const orderStatusModel = computed({
  get: () => props.orderStatus,
  set: (value) => emit('update:orderStatus', Number(value))
})

const popupVisibleModel = computed({
  get: () => props.nearbyOrderPopupVisible,
  set: (value) => emit('update:nearby-order-popup-visible', value)
})

const nearbyOrderPageCount = computed(() => Math.max(1, Math.ceil(props.nearbyOrderTotal / props.nearbyOrderPageSize)))
const nearbyOrderExpandedPageCount = computed(() => Math.max(1, Math.ceil(props.nearbyOrderExpandedTotal / props.nearbyOrderExpandedPageSize)))
const nearbyOrderPopupStyle = {
  height: 'min(86vh, 720px)',
  width: 'min(100vw, 390px)'
}

function resolveOrderId(order) {
  return Number(order?.orderId || order?.orderID || order?.id || order?.dispatch?.orderId || order?.order?.orderId || 0)
}
</script>

<style scoped>
.nearby-order-section {
  display: grid;
  gap: 10px;
  margin-bottom: 12px;
}

.section-actions {
  display: inline-flex;
  gap: 8px;
}

.nearby-order-card {
  border-left: 3px solid #5B5CFF;
}

.order-distance {
  flex: 0 0 auto;
  color: #5B5CFF;
  font-size: 12px;
  font-weight: 800;
}

.nearby-order-popup {
  left: 0;
  right: 0;
  width: min(100vw, 390px);
  margin: 0 auto;
  overflow: hidden;
}

.nearby-order-sheet {
  display: grid;
  grid-template-rows: auto auto minmax(0, 1fr) auto;
  height: 100%;
  padding: 8px 12px calc(12px + env(safe-area-inset-bottom));
  background: #f6f7fb;
}

.nearby-order-popup-list {
  min-height: 0;
  overflow-y: auto;
}

.nearby-order-popup-list .order-card {
  margin-bottom: 10px;
}
</style>
