<template>
  <main class="mine-page-shell">
    <header class="mine-page-header">
      <button type="button" class="page-back" aria-label="返回" @click="goHome">
        <van-icon name="arrow-left" />
      </button>
      <div class="page-heading">
        <p>我的</p>
        <h1>收益明细</h1>
      </div>
      <button type="button" class="page-action" @click="loadIncome">
        <van-icon name="replay" />
      </button>
    </header>

    <section class="page-stack">
      <div class="wallet-card">
        <span>累计收入</span>
        <strong>{{ formatPrice(incomeSummary.totalIncomeCents) }}</strong>
        <p>已完成订单 {{ incomeSummary.completedOrders ?? '--' }}</p>
      </div>
      <div class="income-grid">
        <div><span>今日</span><strong>{{ formatPrice(todayIncome.totalIncomeCents) }}</strong></div>
        <div><span>本周</span><strong>{{ formatPrice(weekIncome.totalIncomeCents) }}</strong></div>
      </div>
      <div class="detail-section-head">
        <h2>收益明细</h2>
        <span>{{ incomeBills.length }} 条</span>
      </div>
      <div v-if="incomeBills.length === 0" class="empty-state">暂无明细</div>
      <article v-for="bill in incomeBills" :key="bill.orderId" class="compact-card">
        <strong>{{ bill.orderNo || '订单 ' + bill.orderId }}</strong>
        <span>{{ formatPrice(bill.incomeCents) }} · {{ formatTime(bill.createdAt) }}</span>
      </article>
    </section>
  </main>
</template>

<script setup>
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useDriverAssets } from '@/composables/useDriverAssets'
import { formatPrice, formatTime } from '@/utils/driver-format'
import '@/styles/driver-home-panels.css'
import '@/styles/driver-mine-pages.css'

const router = useRouter()
const {
  incomeSummary,
  todayIncome,
  weekIncome,
  incomeBills,
  loadIncome
} = useDriverAssets()

onMounted(() => {
  void loadIncome({ silentError: true })
})

function goHome() {
  if (window.history.length > 1) router.back()
  else router.replace('/home')
}
</script>
