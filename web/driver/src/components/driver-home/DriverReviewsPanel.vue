<template>
  <van-popup
    :show="visible"
    position="right"
    class="reviews-panel"
    :style="{ width: '100%', height: '100%' }"
    @update:show="(v) => emit('update:visible', v)"
  >
    <div class="reviews-header">
      <button type="button" class="back" @click="close">返回</button>
      <van-tabs v-model="activeTab" class="reviews-tabs">
        <van-tab title="收到的评价" name="received" />
        <van-tab title="我给出的" name="given" />
      </van-tabs>
    </div>

    <div class="reviews-body">
      <div v-if="activeTab === 'received'" class="review-list">
        <p v-if="receivedLoading" class="hint">加载中…</p>
        <p v-else-if="!receivedList.length" class="hint">暂无乘客评价</p>
        <div v-for="item in receivedList" :key="'r' + item.orderId" class="review-card">
          <van-rate :model-value="item.rating" readonly size="14" color="#F59E0B" void-color="#E5E7EB" :count="5" />
          <p class="review-comment">{{ item.comment || '该乘客未填写文字评价' }}</p>
          <p v-if="item.tags" class="review-tags">{{ item.tags }}</p>
          <p class="review-time">{{ formatReviewTime(item.createdAt) }}</p>
        </div>
      </div>

      <div v-else class="review-list">
        <div class="given-actions">
          <button type="button" class="primary" @click="showSubmit = !showSubmit">{{ showSubmit ? '收起' : '评价乘客' }}</button>
        </div>
        <div v-if="showSubmit" class="submit-form">
          <input v-model="form.orderId" type="number" placeholder="订单ID（已完成且归属于你的订单）" />
          <van-rate v-model="form.rating" size="28" color="#F59E0B" void-color="#E5E7EB" :count="5" />
          <textarea v-model="form.comment" placeholder="评价内容（选填）" maxlength="200" rows="3"></textarea>
          <input v-model="form.tags" placeholder="标签，逗号分隔（选填）" />
          <button type="button" class="primary" :disabled="submitting" @click="submitReview">{{ submitting ? '提交中…' : '提交评价' }}</button>
        </div>
        <p v-if="givenLoading" class="hint">加载中…</p>
        <p v-else-if="!givenList.length" class="hint">暂无评价记录</p>
        <div v-for="item in givenList" :key="'g' + item.orderId" class="review-card">
          <van-rate :model-value="item.rating" readonly size="14" color="#F59E0B" void-color="#E5E7EB" :count="5" />
          <p class="review-comment">{{ item.comment || '未填写文字评价' }}</p>
          <p v-if="item.tags" class="review-tags">{{ item.tags }}</p>
          <p class="review-time">{{ formatReviewTime(item.createdAt) }}</p>
        </div>
      </div>
    </div>
  </van-popup>
</template>

<script setup>
import { ref, watch } from 'vue'
import { showToast } from 'vant'
import { listReceivedReviews, listGivenReviews, submitDriverReview } from '@/api/driver'

const props = defineProps({
  visible: { type: Boolean, default: false },
  mode: { type: String, default: 'received' }
})
const emit = defineEmits(['update:visible'])

const activeTab = ref(props.mode)
const receivedList = ref([])
const givenList = ref([])
const receivedLoading = ref(false)
const givenLoading = ref(false)
const showSubmit = ref(false)
const submitting = ref(false)
const form = ref({ orderId: '', rating: 0, comment: '', tags: '' })

function close() {
  emit('update:visible', false)
}

function formatReviewTime(ts) {
  if (!ts) return ''
  const d = new Date(ts * 1000)
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
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

async function loadGiven() {
  givenLoading.value = true
  try {
    const res = await listGivenReviews({ page: 1, pageSize: 20 }, { silentError: true })
    givenList.value = res?.list || []
  } catch (e) {
    showToast(e?.response?.data?.message || '加载评价失败')
  } finally {
    givenLoading.value = false
  }
}

async function submitReview() {
  const orderId = Number(form.value.orderId)
  if (!orderId || form.value.rating < 1) {
    showToast('请填写订单ID并打分')
    return
  }
  submitting.value = true
  try {
    await submitDriverReview({
      orderId,
      rating: form.value.rating,
      comment: form.value.comment,
      tags: form.value.tags
    }, { silentError: true })
    showToast('评价已提交')
    form.value = { orderId: '', rating: 0, comment: '', tags: '' }
    showSubmit.value = false
    await loadGiven()
  } catch (e) {
    showToast(e?.response?.data?.message || '提交失败')
  } finally {
    submitting.value = false
  }
}

watch(
  () => props.visible,
  (v) => {
    if (!v) return
    activeTab.value = props.mode
    if (activeTab.value === 'received') loadReceived()
    else loadGiven()
  }
)
watch(activeTab, (tab) => {
  if (!props.visible) return
  if (tab === 'received' && !receivedList.value.length) loadReceived()
  if (tab === 'given' && !givenList.value.length) loadGiven()
})
</script>

<style scoped>
.reviews-panel {
  display: flex;
  flex-direction: column;
  background: #f7f8fa;
}
.reviews-header {
  background: #fff;
  border-bottom: 1px solid #ebedf0;
}
.reviews-header .back {
  border: none;
  background: none;
  padding: 12px;
  color: #1989fa;
  font-size: 14px;
}
.reviews-body {
  flex: 1;
  overflow-y: auto;
  padding: 12px;
}
.review-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.review-card {
  background: #fff;
  border-radius: 8px;
  padding: 12px;
}
.review-comment {
  margin: 8px 0 4px;
  color: #333;
}
.review-tags {
  color: #1989fa;
  font-size: 12px;
}
.review-time {
  color: #999;
  font-size: 12px;
  margin-top: 4px;
}
.hint {
  color: #999;
  text-align: center;
  padding: 24px 0;
}
.submit-form {
  background: #fff;
  border-radius: 8px;
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-bottom: 12px;
}
.submit-form input,
.submit-form textarea {
  border: 1px solid #ebedf0;
  border-radius: 6px;
  padding: 8px;
  font-size: 14px;
}
.submit-form .primary,
.given-actions .primary {
  background: #1989fa;
  color: #fff;
  border: none;
  border-radius: 6px;
  padding: 10px;
}
.given-actions {
  margin-bottom: 12px;
}
</style>
