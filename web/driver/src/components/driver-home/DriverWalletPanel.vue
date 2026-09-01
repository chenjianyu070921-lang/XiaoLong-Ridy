<template>
  <section class="h5-panel">
    <div class="section-title">
      <h2>钱包</h2>
      <button type="button" @click="$emit('load-income')">刷新</button>
    </div>
    <div class="income-today-card">
      <div>
        <span>今日收入</span>
        <strong>{{ formatPrice(todayIncome.totalIncomeCents) }}</strong>
        <small>已完成订单 {{ todayIncome.completedOrders ?? '--' }}</small>
      </div>
      <button type="button" class="withdraw-entry" @click="$emit('open-withdraw')">
        <van-icon name="cash-back-record" />
        <span>提现</span>
      </button>
    </div>
    <div class="wallet-card">
      <span>累计收入</span>
      <strong>{{ formatPrice(incomeSummary.totalIncomeCents) }}</strong>
      <p>已完成订单 {{ incomeSummary.completedOrders ?? '--' }}</p>
    </div>
    <div class="income-grid">
      <div><span>今日</span><strong>{{ formatPrice(todayIncome.totalIncomeCents) }}</strong></div>
      <div><span>本周</span><strong>{{ formatPrice(weekIncome.totalIncomeCents) }}</strong></div>
    </div>
    <div v-if="incomeBills.length === 0" class="empty-state">--</div>
    <article v-for="bill in incomeBills" :key="bill.id || bill.orderId" class="compact-card">
      <strong>{{ bill.orderNo || '订单 ' + bill.orderId }}</strong>
      <span>{{ formatPrice(bill.incomeCents) }} · {{ formatTime(bill.createdAt) }}</span>
    </article>
  </section>
</template>

<script setup>
defineProps({
  incomeSummary: { type: Object, required: true },
  todayIncome: { type: Object, required: true },
  weekIncome: { type: Object, required: true },
  incomeBills: { type: Array, required: true },
  formatPrice: { type: Function, required: true },
  formatTime: { type: Function, required: true }
})

defineEmits(['load-income', 'open-withdraw'])
</script>
