<template>
  <section class="h5-panel">
    <div class="section-title">
      <h2>订单</h2>
      <button type="button" @click="$emit('load-orders', orderPage)">刷新</button>
    </div>
    <div class="filter-bar">
      <van-dropdown-menu>
        <van-dropdown-item v-model="orderModeModel" :options="orderModeOptions" @change="$emit('load-orders', 1)" />
        <van-dropdown-item v-model="orderStatusModel" :options="orderStatusOptions" @change="$emit('load-orders', 1)" />
      </van-dropdown-menu>
    </div>
    <div v-if="orders.length === 0" class="empty-state">--</div>
    <article v-for="order in orders" :key="String(order.source || 'order') + '-' + String(order.orderId)" class="order-card">
      <div class="order-heading">
        <strong>{{ order.orderNo || '订单 ' + order.orderId }}</strong>
        <van-tag>{{ order.source === 'dispatch' ? formatDispatchStatus(order.dispatchStatus) : formatOrderStatus(order.status) }}</van-tag>
      </div>
      <p class="route-line">{{ order.fromAddress || '--' }} -> {{ order.toAddress || '--' }}</p>
      <div class="meta-row">
        <span>{{ formatPrice(order.estimatedPriceCents) }}</span>
        <span v-if="order.source === 'available'">距您{{ formatDistance(order.distanceMeters) }}公里</span>
        <span>{{ formatTime(order.createdAt) }}</span>
      </div>
      <div class="order-actions">
        <button type="button" @click="$emit('order-detail', order.orderId)">详情</button>
        <button type="button" @click="$emit('select-trajectory', order.orderId)">轨迹</button>
        <button v-if="canAccept(order)" type="button" class="primary" @click="$emit('order-action', 'accept', order)">接单</button>
        <button v-if="canAccept(order)" type="button" @click="$emit('order-action', 'reject', order)">拒单</button>
        <button v-if="Number(order.status) === 2" type="button" @click="$emit('order-action', 'confirm-arrive', order)">到达</button>
        <button v-if="Number(order.status) === 2" type="button" @click="$emit('order-action', 'start-trip', order)">开始</button>
        <button v-if="Number(order.status) === 3" type="button" class="primary" @click="$emit('open-finish', order)">结束</button>
      </div>
    </article>
    <div class="pager">
      <button type="button" :disabled="orderPage <= 1" @click="$emit('load-orders', orderPage - 1)">上一页</button>
      <span>{{ orderPage }} / {{ Math.max(1, Math.ceil(orderTotal / orderPageSize)) }}</span>
      <button type="button" :disabled="orderPage * orderPageSize >= orderTotal" @click="$emit('load-orders', orderPage + 1)">下一页</button>
    </div>
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
  'load-orders',
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
</script>
