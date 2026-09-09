<template>
  <van-popup
    :show="visible"
    position="bottom"
    teleport="#driver-home-popups"
    class="income-detail"
    :style="{ left: '0', right: '0', width: 'min(100vw, 390px)', height: '100%', margin: '0 auto', borderRadius: 0 }"
    @update:show="(v) => emit('update:visible', v)"
  >
    <div class="inc-header">
      <button type="button" class="back" @click="close">返回</button>
      <span class="inc-title">收益明细</span>
    </div>

    <div class="inc-body">
      <!-- 汇总 -->
      <div class="wallet-card">
        <span>累计收入（元）</span>
        <strong>{{ formatPrice(incomeSummary.totalIncomeCents) }}</strong>
        <p>已完成订单 {{ incomeSummary.completedOrders ?? '--' }} · 可提现 {{ formatPrice(incomeSummary.withdrawableCents) }}</p>
      </div>

      <!-- 今日 / 本周 -->
      <div class="income-grid">
        <div><span>今日</span><strong>{{ formatPrice(todayIncome.totalIncomeCents) }}</strong></div>
        <div><span>本周</span><strong>{{ formatPrice(weekIncome.totalIncomeCents) }}</strong></div>
      </div>

      <!-- 明细记录 -->
      <div class="detail-head">
        <h2>明细记录</h2>
        <span>{{ incomeBills.length }} 条</span>
        <button type="button" class="refresh" :disabled="loading" @click="reload">
          <van-icon name="replay" />
        </button>
      </div>

      <div v-if="loading" class="empty-state">加载中…</div>
      <div v-else-if="incomeBills.length === 0" class="empty-state">暂无明细</div>
      <article v-for="bill in incomeBills" :key="bill.orderId" class="bill-row">
        <div class="bill-main">
          <strong>{{ bill.orderNo || ('订单 ' + bill.orderId) }}</strong>
          <span class="bill-time">{{ formatTime(bill.createdAt) }}</span>
        </div>
        <div class="bill-side">
          <span class="bill-amt">+{{ formatPrice(bill.incomeCents) }}</span>
          <span class="bill-status" :class="bill.status >= 1 ? 'settled' : 'pending'">
            {{ bill.status >= 1 ? '已入账' : '结算中' }}
          </span>
        </div>
      </article>
    </div>
  </van-popup>
</template>

<script setup>
import { ref, watch } from 'vue'
import { useDriverAssets } from '@/composables/useDriverAssets'
import { formatPrice, formatTime } from '@/utils/driver-format'

const props = defineProps({
  visible: { type: Boolean, default: false }
})
const emit = defineEmits(['update:visible'])

const { incomeSummary, todayIncome, weekIncome, incomeBills, loadIncome } = useDriverAssets()
const loading = ref(false)

function close() {
  emit('update:visible', false)
}

async function reload() {
  loading.value = true
  try {
    await loadIncome({ silentError: true })
  } finally {
    loading.value = false
  }
}

// 打开时自动加载一次收益数据，无需手动触发。
watch(
  () => props.visible,
  (v) => {
    if (v) reload()
  }
)
</script>

<style scoped>
.income-detail {
  display: flex;
  flex-direction: column;
  background: var(--driver-soft);
}
.inc-header {
  display: flex;
  align-items: center;
  gap: 8px;
  background: var(--driver-card);
  border-bottom: 1px solid var(--driver-line);
  padding: 12px;
}
.inc-header .back {
  border: none;
  background: none;
  color: var(--driver-primary);
  font-size: 14px;
  padding: 4px 8px;
}
.inc-title {
  font-size: 16px;
  font-weight: 700;
  color: var(--driver-ink);
}
.inc-body {
  flex: 1;
  overflow-y: auto;
  padding: 12px;
  display: grid;
  gap: 12px;
  align-content: start;
}
.wallet-card {
  background: #172033;
  color: #fff;
  border-radius: 14px;
  padding: 18px 16px;
  display: grid;
  gap: 4px;
}
.wallet-card span { font-size: 12px; color: #b9c2d2; }
.wallet-card strong { font-size: 30px; line-height: 1.2; }
.wallet-card p { margin: 4px 0 0; font-size: 12px; color: #b9c2d2; }
.income-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
}
.income-grid div {
  background: var(--driver-card);
  border: 1px solid var(--driver-line);
  border-radius: 12px;
  padding: 14px;
  display: grid;
  gap: 4px;
}
.income-grid span { font-size: 12px; color: var(--driver-muted); }
.income-grid strong { font-size: 20px; color: var(--driver-ink); }
.detail-head {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 4px;
}
.detail-head h2 { margin: 0; font-size: 15px; color: var(--driver-ink); flex: 1; }
.detail-head span { font-size: 12px; color: var(--driver-muted); }
.detail-head .refresh {
  border: none;
  background: none;
  color: var(--driver-primary);
  font-size: 18px;
  padding: 4px;
}
.detail-head .refresh:disabled { opacity: .5; }
.empty-state {
  text-align: center;
  color: var(--driver-muted);
  font-size: 13px;
  padding: 24px 0;
}
.bill-row {
  background: var(--driver-card);
  border: 1px solid var(--driver-line);
  border-radius: 12px;
  padding: 12px 14px;
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.bill-main { display: flex; flex-direction: column; gap: 3px; min-width: 0; }
.bill-main strong { font-size: 14px; color: var(--driver-ink); }
.bill-time { font-size: 12px; color: var(--driver-muted); }
.bill-side { display: flex; flex-direction: column; align-items: flex-end; gap: 3px; }
.bill-amt { font-size: 15px; font-weight: 700; color: #059669; }
.bill-status { font-size: 11px; padding: 1px 6px; border-radius: 999px; }
.bill-status.settled { background: rgba(16,163,74,.14); color: #10b981; }
.bill-status.pending { background: rgba(217,119,6,.14); color: #D97706; }
</style>
