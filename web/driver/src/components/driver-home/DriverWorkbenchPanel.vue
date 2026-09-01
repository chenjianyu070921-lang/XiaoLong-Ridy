<template>
  <section class="h5-panel">
    <div class="section-title">
      <h2>工作台</h2>
      <button type="button" @click="$emit('refresh-dashboard')">刷新</button>
    </div>
    <div class="current-trip-card">
      <div class="section-subtitle">
        <strong>当前行程</strong>
        <van-tag type="primary">{{ formatOrderStatus(driverStore.currentOrder?.status) }}</van-tag>
      </div>
      <p><b>上车点</b><span>{{ driverStore.currentOrder?.fromAddress || '--' }}</span></p>
      <p><b>目的地</b><span>{{ driverStore.currentOrder?.toAddress || '--' }}</span></p>
      <p><b>订单号</b><span>{{ driverStore.currentOrder?.orderNo || driverStore.currentOrderId || '--' }}</span></p>
    </div>
    <div class="map-surface">
      <span class="road road-a"></span>
      <span class="road road-b"></span>
      <span class="pin current"></span>
      <span class="pin order"></span>
      <p>接单区域</p>
    </div>
    <section class="home-order-section">
      <div class="section-title">
        <h2>附近可接订单</h2>
        <button type="button" :disabled="homeAvailableLoading" @click="$emit('refresh-home-orders')">刷新</button>
      </div>
      <div v-if="homeAvailableLoading" class="home-order-loading"><van-loading size="20px" /></div>
      <div v-else-if="homeAvailableOrders.length === 0" class="home-order-empty">
        {{ driverStore.onlineStatus === 1 ? '暂无附近订单' : '开始听单后查看附近订单' }}
      </div>
      <article v-for="order in homeAvailableOrders" :key="'home-' + order.orderId" class="home-order-card">
        <div class="order-heading">
          <strong>{{ order.orderNo || '订单 ' + order.orderId }}</strong>
          <span class="order-distance">距您{{ formatDistance(order.distanceMeters) }}公里</span>
        </div>
        <p class="route-line">{{ order.fromAddress || '--' }} -> {{ order.toAddress || '--' }}</p>
        <div class="meta-row">
          <span>{{ formatPrice(order.estimatedPriceCents) }}</span>
          <button type="button" class="home-order-accept" @click="$emit('order-action', 'accept', order)">接单</button>
        </div>
      </article>
    </section>
  </section>
</template>

<script setup>
defineProps({
  driverStore: { type: Object, required: true },
  homeAvailableOrders: { type: Array, required: true },
  homeAvailableLoading: { type: Boolean, required: true },
  formatPrice: { type: Function, required: true },
  formatDistance: { type: Function, required: true },
  formatOrderStatus: { type: Function, required: true }
})

defineEmits(['refresh-dashboard', 'refresh-home-orders', 'order-action'])
</script>
