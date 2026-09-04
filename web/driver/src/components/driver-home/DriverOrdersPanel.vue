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
        <div class="card-header">
          <span class="status-tag ongoing">{{ formatOrderStatus(order.status || 1) }}</span>
          <span class="time">{{ formatTime(order.createdAt) }}</span>
        </div>
        <div class="route-info">
          <div class="route-item">
            <div class="dot from"></div>
            <span>{{ order.fromAddress || '--' }}</span>
          </div>
          <div class="route-line"></div>
          <div class="route-item">
            <div class="dot to"></div>
            <span>{{ order.toAddress || '--' }}</span>
          </div>
        </div>
        <div class="card-footer">
          <div class="car-type">{{ order.orderNo || '订单 ' + resolveOrderId(order) }}</div>
          <div class="price">{{ formatPrice(order.estimatedPriceCents) }}</div>
        </div>
        <div class="trip-stats">
          <div class="stat-item"><span class="value">{{ formatDistance(order.distanceMeters) }}</span><span class="label">距您</span></div>
          <div class="stat-divider"></div>
          <div class="stat-item"><span class="value">{{ formatTime(order.createdAt) }}</span><span class="label">发单时间</span></div>
        </div>
        <div class="actions">
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
      <div class="card-header">
        <span class="status-tag" :class="orderStatusClass(order)">{{ order.source === 'dispatch' ? formatDispatchStatus(order.dispatchStatus) : formatOrderStatus(order.status) }}</span>
        <span class="time">{{ formatTime(order.createdAt) }}</span>
      </div>
      <div class="route-info">
        <div class="route-item">
          <div class="dot from"></div>
          <span>{{ order.fromAddress || '--' }}</span>
        </div>
        <div class="route-line"></div>
        <div class="route-item">
          <div class="dot to"></div>
          <span>{{ order.toAddress || '--' }}</span>
        </div>
      </div>
      <div class="card-footer">
        <div class="car-type">{{ order.orderNo || '订单 ' + resolveOrderId(order) }}</div>
        <div class="price">{{ formatPrice(order.estimatedPriceCents) }}</div>
      </div>
      <div class="trip-stats">
        <div class="stat-item"><span class="value">{{ order.source === 'available' ? formatDistance(order.distanceMeters) : '--' }}</span><span class="label">距离</span></div>
        <div class="stat-divider"></div>
        <div class="stat-item"><span class="value">{{ formatTime(order.createdAt) }}</span><span class="label">时间</span></div>
      </div>
      <div class="actions">
        <button type="button" @click="emit('order-detail', resolveOrderId(order))">详情</button>
        <button v-if="canAccept(order)" type="button" class="primary" @click="emit('order-action', 'accept', order)">接单</button>
        <button v-if="canAccept(order)" type="button" @click="emit('order-action', 'reject', order)">拒单</button>
        <button v-if="canViewTrajectory(order)" type="button" @click="emit('open-trajectory', order)">轨迹</button>
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
            <div class="card-header">
              <span class="status-tag ongoing">{{ formatOrderStatus(order.status || 1) }}</span>
              <span class="time">{{ formatTime(order.createdAt) }}</span>
            </div>
            <div class="route-info">
              <div class="route-item">
                <div class="dot from"></div>
                <span>{{ order.fromAddress || '--' }}</span>
              </div>
              <div class="route-line"></div>
              <div class="route-item">
                <div class="dot to"></div>
                <span>{{ order.toAddress || '--' }}</span>
              </div>
            </div>
            <div class="card-footer">
              <div class="car-type">{{ order.orderNo || '订单 ' + resolveOrderId(order) }}</div>
              <div class="price">{{ formatPrice(order.estimatedPriceCents) }}</div>
            </div>
            <div class="trip-stats">
              <div class="stat-item"><span class="value">{{ formatDistance(order.distanceMeters) }}</span><span class="label">距您</span></div>
              <div class="stat-divider"></div>
              <div class="stat-item"><span class="value">{{ formatTime(order.createdAt) }}</span><span class="label">发单时间</span></div>
            </div>
            <div class="actions">
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
  'order-action',
  'open-finish',
  'open-trajectory'
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

function orderStatusClass(order) {
  const status = Number(order?.status || order?.dispatchStatus || 0)
  if ([1, 2, 3].includes(status)) return 'ongoing'
  if (status === 5) return 'completed'
  if (status === 6 || status === 7) return 'cancelled'
  if (Number(order?.dispatchStatus || 0) === 1) return 'pending'
  return 'pending'
}

function canViewTrajectory(order) {
  return [2, 3, 4, 5].includes(Number(order?.status || 0))
}
</script>

<style scoped>
.h5-panel { padding: 10px 0 24px; color: #172033; }
.section-title { display:flex; align-items:center; justify-content:space-between; gap:12px; margin: 4px 0 10px; }
.section-title h2 { margin:0; font-size:20px; letter-spacing:0; }
.section-title button, .section-actions button { min-height:34px; padding:0 12px; border:1px solid #e1e6ef; border-radius:8px; background:#fff; color:#344054; font-size:12px; font-weight:700; }
.section-actions { gap:6px; }
.nearby-order-section, .h5-panel > .section-title, .filter-bar, .h5-panel > .order-card, .h5-panel > .pager { margin-left: 2px; margin-right: 2px; }
.nearby-order-section { padding:14px; border-radius:14px; background:#fff; box-shadow:0 6px 18px rgba(15,23,42,.06); }
.home-order-empty, .empty-state { display:grid; min-height:106px; place-items:center; color:#98a2b3; font-size:13px; background:#f8f9fc; border-radius:10px; }
.filter-bar { margin-bottom:10px; overflow:hidden; border-radius:10px; background:#fff; box-shadow:0 3px 12px rgba(15,23,42,.05); }
.filter-bar :deep(.van-dropdown-menu__bar) { height:44px; box-shadow:none; }
.order-card { display:grid; gap:10px; margin-bottom:10px; padding:14px; border:1px solid #e9edf4; border-radius:12px; background:#fff; box-shadow:0 5px 16px rgba(15,23,42,.05); }
.order-heading { display:flex; align-items:center; justify-content:space-between; gap:10px; }.order-heading strong { overflow:hidden; font-size:15px; text-overflow:ellipsis; white-space:nowrap; }.order-heading :deep(.van-tag) { border-radius:6px; }
.route-line { margin:0; color:#344054; font-size:14px; font-weight:600; line-height:1.5; overflow-wrap:anywhere; }
.meta-row { display:flex; justify-content:space-between; gap:8px; color:#98a2b3; font-size:12px; }.meta-row span:first-child { color:#5b5cff; font-size:17px; font-weight:800; }
.order-actions { display:flex; justify-content:flex-end; gap:7px; padding-top:2px; }.order-actions button { min-height:34px; padding:0 13px; border:1px solid #dfe4ec; border-radius:8px; background:#fff; color:#475467; font-size:12px; font-weight:700; }.order-actions button.primary { border-color:#5b5cff; background:#5b5cff; color:#fff; }
.pager { display:flex; align-items:center; justify-content:center; gap:12px; margin:14px 0 18px; color:#667085; font-size:13px; }.pager button { min-height:34px; padding:0 12px; border:1px solid #e1e6ef; border-radius:8px; background:#fff; color:#344054; font-size:12px; }.pager button:disabled { opacity:.45; }
.nearby-order-card { border-left:3px solid #5b5cff; }
.nearby-order-section {
  display: grid;
  gap: 10px;
  margin-bottom: 12px;
}

.section-actions {
  display: inline-flex;
  gap: 8px;
}

.order-card {
  gap: 0;
  padding: 16px;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, .04);
}

.order-card:active {
  transform: scale(.98);
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 14px;
}

.status-tag {
  flex: 0 0 auto;
  padding: 4px 10px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 600;
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
  min-width: 0;
  overflow: hidden;
  color: #9CA3AF;
  font-size: 12px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.route-info {
  position: relative;
  display: grid;
  gap: 0;
  margin-bottom: 14px;
  padding-left: 18px;
}

.route-item {
  position: relative;
  display: flex;
  align-items: center;
  gap: 10px;
  min-height: 30px;
  padding: 6px 0;
  color: #172033;
  font-size: 14px;
  font-weight: 600;
}

.route-item span {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.dot {
  position: absolute;
  left: -12px;
  width: 10px;
  height: 10px;
  border-radius: 50%;
}

.dot.from {
  background: #10B981;
}

.dot.to {
  background: #EF4444;
}

.route-info .route-line {
  position: absolute;
  left: -8px;
  top: 24px;
  bottom: 24px;
  width: 2px;
  background: #D1D5DB;
}

.card-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding-top: 12px;
  border-top: 1px solid #F3F4F6;
}

.car-type {
  min-width: 0;
  overflow: hidden;
  color: #6B7280;
  font-size: 13px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.price {
  flex: 0 0 auto;
  color: #EF4444;
  font-size: 18px;
  font-weight: 800;
}

.trip-stats {
  display: flex;
  align-items: center;
  justify-content: space-around;
  gap: 10px;
  margin-top: 14px;
  padding: 12px 10px;
  border-radius: 10px;
  background: #F9FAFB;
}

.stat-item {
  display: grid;
  gap: 4px;
  min-width: 0;
  flex: 1;
  text-align: center;
}

.stat-item .value {
  overflow: hidden;
  color: #172033;
  font-size: 13px;
  font-weight: 800;
  line-height: 1.15;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.stat-item .label {
  color: #6B7280;
  font-size: 11px;
}

.stat-divider {
  width: 1px;
  height: 30px;
  background: #E5E7EB;
}

.actions {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  margin-top: 14px;
  padding-top: 14px;
  border-top: 1px solid #F3F4F6;
}

.actions button {
  min-width: 74px;
  min-height: 36px;
  padding: 0 14px;
  border: 0;
  border-radius: 18px;
  background: #EFF6FF;
  color: #3B82F6;
  font-size: 13px;
  font-weight: 800;
}

.actions .primary {
  background: linear-gradient(135deg, #7C3AED 0%, #9333EA 100%);
  color: #fff;
}

.actions button:not(.primary):nth-child(3) {
  background: #FEF3C7;
  color: #D97706;
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
