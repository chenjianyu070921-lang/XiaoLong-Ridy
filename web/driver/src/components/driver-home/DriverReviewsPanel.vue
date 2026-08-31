<template>
  <section class="h5-panel">
    <div class="section-title">
      <h2>评价</h2>
      <button type="button" @click="$emit('load-reviews')">刷新</button>
    </div>
    <div v-if="reviews.length === 0" class="empty-state">--</div>
    <article v-for="review in reviews" :key="review.id || review.reviewId || review.orderId" class="compact-card">
      <strong>订单 {{ review.orderId || '--' }} · {{ review.rating || '--' }} 分</strong>
      <span>{{ review.comment || '--' }}</span>
      <small>{{ formatTime(review.createdAt) }}</small>
    </article>
  </section>
</template>

<script setup>
defineProps({
  reviews: { type: Array, required: true },
  formatTime: { type: Function, required: true }
})

defineEmits(['load-reviews'])
</script>
