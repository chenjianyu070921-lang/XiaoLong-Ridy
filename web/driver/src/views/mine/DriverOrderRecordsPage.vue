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
      <article v-for="item in records" :key="item.key" class="record-card">
        <div class="record-head">
          <strong>{{ item.orderNo || '订单 ' + item.orderId }}</strong>
          <span>{{ item.kind === 'dispatches' ? formatDispatchStatus(item.status) : formatOrderStatus(item.status) }}</span>
        </div>
        <p class="route-line">{{ item.fromAddress || '--' }} -> {{ item.toAddress || '--' }}</p>
        <div class="record-meta">
          <span>{{ formatPrice(item.estimatedPriceCents) }}</span>
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
import { showToast } from 'vant'
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
      ? await listDriverDispatches(payload, { silentError: true })
      : await listDriverOrders(payload, { silentError: true })
    records.value = (Array.isArray(res?.list) ? res.list : []).map((item) => normalizeRecord(item, mode.value))
    total.value = Number(res?.total || records.value.length || 0)
  } catch (error) {
    records.value = []
    total.value = 0
    showToast(error?.message || '订单记录加载失败')
  } finally {
    loading.value = false
  }
}

// 把后端两种列表结构规范化为同一套前端字段：
// - 我的订单：OrderBrief（扁平，字段直接在 item 上）
// - 派单记录：MyDispatchItem{ dispatch, order }（嵌套，金额/路线/时间/订单号在 order 内，状态在 dispatch 内）
// 统一后模板无需按 mode 写大量三元表达式，避免派单模式下读到 undefined。
function normalizeRecord(item, currentMode) {
  if (currentMode === 'dispatches') {
    const order = item?.order || {}
    const dispatch = item?.dispatch || {}
    const orderId = order.orderId || dispatch.orderId || 0
    return {
      kind: 'dispatches',
      key: `dispatch-${orderId}-${dispatch.status}-${order.createdAt || 0}`,
      orderId,
      orderNo: order.orderNo || '',
      fromAddress: order.fromAddress || '',
      toAddress: order.toAddress || '',
      estimatedPriceCents: order.estimatedPriceCents || 0,
      createdAt: order.createdAt || 0,
      status: dispatch.status,
    }
  }
  const orderId = item?.orderId || 0
  return {
    kind: 'orders',
    key: `order-${orderId}-${item?.status}-${item?.createdAt || 0}`,
    orderId,
    orderNo: item?.orderNo || '',
    fromAddress: item?.fromAddress || '',
    toAddress: item?.toAddress || '',
    estimatedPriceCents: item?.estimatedPriceCents || 0,
    createdAt: item?.createdAt || 0,
    status: item?.status,
  }
}

onMounted(() => {
  void loadRecords()
})

function goHome() {
  router.back()
}
</script>
