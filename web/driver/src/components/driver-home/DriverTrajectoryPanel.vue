<template>
  <section class="h5-panel">
    <div class="section-title"><h2>轨迹</h2></div>
    <div class="trajectory-form">
      <van-field v-model.number="trajectoryOrderIdModel" type="number" label="订单ID" placeholder="输入订单ID" />
      <p v-if="trajectoryError" class="trajectory-error">{{ trajectoryError }}</p>
      <button class="primary-action" type="button" @click="$emit('load-trajectory')">查询轨迹</button>
    </div>
    <div v-if="trajectoryPoints.length === 0" class="empty-state">--</div>
    <article v-for="point in trajectoryPoints" :key="point.id || point.reportTime || point.createdAt" class="compact-card">
      <strong>{{ point.latitude || '--' }}, {{ point.longitude || '--' }}</strong>
      <span>速度 {{ point.speedKmh ?? '--' }} km/h · {{ formatTime(point.reportTime || point.createdAt) }}</span>
    </article>
  </section>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  trajectoryOrderId: { type: [String, Number], required: true },
  trajectoryError: { type: String, required: true },
  trajectoryPoints: { type: Array, required: true },
  formatTime: { type: Function, required: true }
})

const emit = defineEmits(['update:trajectoryOrderId', 'load-trajectory'])

const trajectoryOrderIdModel = computed({
  get: () => props.trajectoryOrderId,
  set: (value) => emit('update:trajectoryOrderId', value)
})
</script>
