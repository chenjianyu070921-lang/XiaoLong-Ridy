<template>
  <van-popup
    :show="visible"
    position="bottom"
    teleport="#driver-home-popups"
    class="reviews-panel"
    :style="{ left: '0', right: '0', width: 'min(100vw, 390px)', height: '100%', margin: '0 auto', borderRadius: 0 }"
    @update:show="(v) => emit('update:visible', v)"
  >
    <div class="reviews-header">
      <button type="button" class="back" @click="close">返回</button>
      <span class="reviews-title">乘客评价</span>
    </div>

    <div class="reviews-body">
      <div v-if="summaryLoading" class="hint">加载中…</div>

      <template v-else-if="summary">
        <!-- 服务平均分概览（agent 生成） -->
        <div class="score-card">
          <div class="score-top">
            <div class="score-main">
              <span class="score-label">服务平均分</span>
              <div class="score-value-row">
                <b class="score-value">{{ displayScore }}</b>
                <span class="score-unit">分</span>
              </div>
              <van-rate :model-value="displayScore" readonly size="14" color="var(--driver-primary)" void-color="var(--driver-line)" :count="5" allow-half />
            </div>
            <div class="score-side">
              <span class="score-pill" :class="summary.canReceiveReview ? 'ok' : 'lock'">
                {{ summary.canReceiveReview ? '可接收评价' : '暂未开放' }}
              </span>
              <span class="score-meta">完成 {{ summary.completedOrderCount }} 单</span>
              <span class="score-meta">收到 {{ summary.reviewCount }} 条评价</span>
            </div>
          </div>
        </div>

        <!-- 未满 5 单：不开放评价列表 -->
        <div v-if="!summary.canReceiveReview" class="lock-tip">
          <van-icon name="lock" />
          <span>完成 <b>{{ summary.completedOrderCount }}/5</b> 单后可接收乘客评价</span>
        </div>

        <!-- 已满 5 单：展示评价列表 -->
        <div v-else class="review-list">
          <p v-if="receivedLoading" class="hint">加载中…</p>
          <p v-else-if="!receivedList.length" class="hint">暂无乘客评价</p>
          <div v-for="item in receivedList" :key="'r' + item.orderId" class="review-card">
            <van-rate :model-value="item.rating" readonly size="14" color="#F59E0B" void-color="var(--driver-line)" :count="5" />
            <p class="review-comment">{{ item.comment || '该乘客未填写文字评价' }}</p>
            <p v-if="item.tags" class="review-tags">{{ item.tags }}</p>
            <p class="review-time">{{ formatReviewTime(item.createdAt) }}</p>
          </div>
        </div>
      </template>
    </div>
  </van-popup>
</template>

<script setup>
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { showToast } from 'vant'
import { getReviewSummary, listReceivedReviews } from '@/api/driver'

const props = defineProps({
  visible: { type: Boolean, default: false }
})
const emit = defineEmits(['update:visible'])

const summary = ref(null)
const summaryLoading = ref(false)
const receivedList = ref([])
const receivedLoading = ref(false)

const displayScore = computed(() => {
  const raw = Number(summary.value?.serviceScore ?? summary.value?.avgRating ?? 0)
  return Number.isFinite(raw) ? raw : 0
})

function close() {
  emit('update:visible', false)
}

function formatReviewTime(ts) {
  if (!ts) return ''
  const d = new Date(ts * 1000)
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

async function loadSummary() {
  summaryLoading.value = true
  try {
    const res = await getReviewSummary({ silentError: true })
    summary.value = res
  } catch (e) {
    showToast(e?.response?.data?.message || '加载评价概览失败')
  } finally {
    summaryLoading.value = false
  }
}

async function loadReceived() {
  receivedLoading.value = true
  try {
    const res = await listReceivedReviews({ page: 1, pageSize: 20 }, { silentError: true })
    receivedList.value = res?.list || []
  } catch (e) {
    showToast(e?.response?.data?.message || '加载评价失败')
  } finally {
    receivedLoading.value = false
  }
}

watch(
  () => props.visible,
  (v) => {
    if (v) {
      loadSummary()
      loadReceived()
    }
  }
)

// 实时更新：司机端 WS 推送 review.new（乘客新评价导致评分统计变化）时，
// 评价面板若处于打开状态则自动刷新服务平均分与评价列表，无需手动刷新。
function onReviewUpdated() {
  if (props.visible) {
    loadSummary()
    loadReceived()
  }
}
onMounted(() => window.addEventListener('driver-review-updated', onReviewUpdated))
onBeforeUnmount(() => window.removeEventListener('driver-review-updated', onReviewUpdated))
</script>

<style scoped>
.reviews-panel {
  display: flex;
  flex-direction: column;
  background: var(--driver-soft);
}
.reviews-header {
  display: flex;
  align-items: center;
  gap: 8px;
  background: var(--driver-card);
  border-bottom: 1px solid var(--driver-line);
  padding: 12px;
}
.reviews-header .back {
  border: none;
  background: none;
  padding: 0;
  color: var(--driver-primary);
  font-size: 14px;
}
.reviews-header .reviews-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--driver-ink);
}
.reviews-body {
  flex: 1;
  overflow-y: auto;
  padding: 12px;
}
.hint {
  color: var(--driver-muted);
  text-align: center;
  padding: 24px 0;
}

/* 服务平均分概览卡 */
.score-card {
  background: var(--driver-card);
  border-radius: 12px;
  padding: 14px;
  margin-bottom: 12px;
  box-shadow: 0 2px 10px rgba(15, 23, 42, .05);
}
.score-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}
.score-main {
  display: grid;
  gap: 4px;
}
.score-label {
  color: var(--driver-muted);
  font-size: 12px;
}
.score-value-row {
  display: flex;
  align-items: baseline;
  gap: 4px;
}
.score-value {
  color: var(--driver-primary);
  font-size: 34px;
  font-weight: 800;
  line-height: 1.1;
}
.score-unit {
  color: var(--driver-muted);
  font-size: 12px;
}
.score-side {
  display: grid;
  gap: 6px;
  justify-items: end;
}
.score-pill {
  padding: 3px 10px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 700;
}
.score-pill.ok {
  background: var(--driver-soft);
  color: var(--driver-primary);
}
.score-pill.lock {
  background: var(--driver-soft);
  color: var(--driver-muted);
}
.score-meta {
  color: var(--driver-muted);
  font-size: 12px;
}

/* 未满 5 单门槛提示 */
.lock-tip {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 20px 12px;
  border-radius: 12px;
  background: var(--driver-card);
  color: var(--driver-muted);
  font-size: 13px;
}
.lock-tip b {
  color: var(--driver-primary);
}

/* 评价列表 */
.review-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.review-card {
  background: var(--driver-card);
  border-radius: 12px;
  padding: 12px;
}
.review-comment {
  margin: 8px 0 4px;
  color: var(--driver-ink);
}
.review-tags {
  color: var(--driver-primary);
  font-size: 12px;
}
.review-time {
  color: var(--driver-muted);
  font-size: 12px;
  margin-top: 4px;
}
</style>
