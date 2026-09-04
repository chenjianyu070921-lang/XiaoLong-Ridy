<template>
  <main class="mine-page-shell">
    <header class="mine-page-header">
      <button type="button" class="page-back" aria-label="返回" @click="goHome">
        <van-icon name="arrow-left" />
      </button>
      <div class="page-heading">
        <p>我的</p>
        <h1>订单记录</h1>
      </div>
      <button type="button" class="page-action" @click="loadRecords">
        <van-icon name="replay" />
      </button>
    </header>

    <section class="page-stack">
      <div class="page-switcher">
        <button type="button" :class="{ active: mode === 'orders' }" @click="setMode('orders')">我的订单</button>
        <button type="button" :class="{ active: mode === 'dispatches' }" @click="setMode('dispatches')">派单记录</button>
      </div>

      <div v-if="loading" class="empty-state">加载中...</div>
      <div v-else-if="records.length === 0" class="empty-state">暂无记录</div>
      <article v-for="item in records" :key="item.orderId || item.id || item.dispatchId" class="record-card">
        <div class="record-head">
          <strong>{{ item.orderNo || '订单 ' + resolveOrderId(item) }}</strong>
          <span>{{ mode === 'orders' ? formatOrderStatus(item.status) : formatDispatchStatus(item.dispatchStatus) }}</span>
        </div>
        <p class="route-line">{{ item.fromAddress || '--' }} -> {{ item.toAddress || '--' }}</p>
        <div class="record-meta">
          <span>{{ formatPrice(item.estimatedPriceCents || item.incomeCents) }}</span>
          <span>{{ formatTime(item.createdAt) }}</span>
        </div>
      </article>

      <div class="pager">
        <button type="button" :disabled="page <= 1 || loading" @click="prevPage">上一页</button>
        <span>{{ page }}</span>
        <button type="button" :disabled="page >= pageCount || loading" @click="nextPage">下一页</button>
      </div>
    </section>
  </main>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { listDriverDispatches, listDriverOrders } from '@/api/driver'
import { formatDispatchStatus, formatOrderStatus, formatPrice, formatTime } from '@/utils/driver-format'
import '@/styles/driver-home-panels.css'
import '@/styles/driver-mine-pages.css'

const router = useRouter()
const mode = ref('orders')
const page = ref(1)
const pageSize = 10
const total = ref(0)
const records = ref([])
const loading = ref(false)

const pageCount = computed(() => Math.max(1, Math.ceil((total.value || 0) / pageSize)))

function setMode(nextMode) {
  if (mode.value === nextMode) return
  mode.value = nextMode
  page.value = 1
  void loadRecords()
}

function prevPage() {
  if (page.value <= 1 || loading.value) return
  page.value -= 1
  void loadRecords()
}

function nextPage() {
  if (page.value >= pageCount.value || loading.value) return
  page.value += 1
  void loadRecords()
}

async function loadRecords() {
  loading.value = true
  try {
    const payload = { page: page.value, pageSize }
    const res = mode.value === 'dispatches'
      ? await listDriverDispatches(payload)
      : await listDriverOrders(payload)
    records.value = Array.isArray(res?.list) ? res.list : []
    total.value = Number(res?.total || records.value.length || 0)
  } catch (error) {
    records.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

function resolveOrderId(item) {
  return Number(item?.orderId || item?.orderID || item?.id || item?.dispatch?.orderId || item?.order?.orderId || 0)
}

onMounted(() => {
  void loadRecords()
})

function goHome() {
  router.back()
}
</script>
