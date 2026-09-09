<template>
  <van-popup
    :show="visible"
    position="bottom"
    teleport="#driver-home-popups"
    class="chat-panel"
    :style="{ left: '0', right: '0', width: 'min(100vw, 390px)', height: '100%', margin: '0 auto', borderRadius: 0 }"
    @update:show="(v) => emit('update:visible', v)"
  >
    <header class="cp-header">
      <button type="button" class="cp-close" aria-label="关闭" @click="close">
        <van-icon name="cross" />
      </button>
      <div class="cp-title">
        <strong>与乘客沟通</strong>
        <small v-if="peer">{{ peer.name }} · {{ orderNo }}</small>
        <small v-else>{{ orderNo }}</small>
      </div>
    </header>

    <div v-if="!opened" class="cp-locked">
      <van-icon name="comment-o" />
      <p>接单成功后可在此与乘客沟通</p>
    </div>

    <template v-else>
      <div ref="listEl" class="cp-messages">
        <p v-if="loading" class="cp-hint">加载中…</p>
        <div
          v-for="m in messages"
          :key="m.clientMsgId || m.id"
          class="cp-row"
          :class="m.senderType === 1 ? 'mine' : 'peer'"
        >
          <div class="cp-bubble">
            <span class="cp-text">{{ m.content }}</span>
            <span class="cp-time">{{ formatTime(m.createAt) }}</span>
          </div>
        </div>
      </div>

      <div class="cp-quick">
        <button
          v-for="q in quickReplies"
          :key="q"
          type="button"
          class="cp-quick-item"
          @click="onQuick(q)"
        >{{ q }}</button>
      </div>

      <footer class="cp-input-bar">
        <input
          v-model="inputText"
          class="cp-input"
          type="text"
          maxlength="200"
          placeholder="输入消息，禁止线下交易/交换联系方式"
          @keyup.enter="onSend"
        />
        <button type="button" class="cp-send" :disabled="sending" @click="onSend">发送</button>
      </footer>
    </template>
  </van-popup>
</template>

<script setup>
import { computed, nextTick, ref, watch } from 'vue'
import { showToast } from 'vant'
import { getConversation, listMessages, sendMessage, markRead, genClientMsgId } from '@/api/chat'

const props = defineProps({
  visible: { type: Boolean, default: false },
  order: { type: Object, default: null }
})
const emit = defineEmits(['update:visible'])

const loading = ref(false)
const opened = ref(false)
const convStatus = ref(0)
const conversationId = ref(0)
const peer = ref(null)
const messages = ref([])
const inputText = ref('')
const sending = ref(false)
const listEl = ref(null)

// 快捷语：MVP 前端硬编码（方案 B），后续可下沉为后端配置。
const quickReplies = ['我马上到', '我在附近，请稍等', '请到上车点等待', '我已到达上车点', '请打开车门']

const orderNo = computed(() => props.order?.orderNo || '--')

function close() {
  emit('update:visible', false)
}

function scrollToBottom() {
  nextTick(() => {
    const el = listEl.value
    if (el) el.scrollTop = el.scrollHeight
  })
}

function formatTime(ts) {
  if (!ts) return ''
  const d = new Date(ts * 1000)
  const hh = String(d.getHours()).padStart(2, '0')
  const mm = String(d.getMinutes()).padStart(2, '0')
  return `${hh}:${mm}`
}

async function loadMessages() {
  const res = await listMessages(conversationId.value, 0, 20)
  messages.value = res.messages || []
  try {
    await markRead(conversationId.value)
  } catch (e) { /* 标记已读失败不阻断 */ }
  scrollToBottom()
}

async function loadConversation() {
  if (!props.order?.orderId) return
  loading.value = true
  opened.value = false
  conversationId.value = 0
  messages.value = []
  try {
    const conv = await getConversation(props.order.orderId)
    conversationId.value = conv.conversationId || 0
    opened.value = !!conv.opened
    convStatus.value = conv.status
    peer.value = conv.peer || null
    if (opened.value && conversationId.value) {
      await loadMessages()
    }
  } catch (e) {
    showToast('会话加载失败')
  } finally {
    loading.value = false
  }
}

// DriverHome 经 WebSocket push 调用：追加对端（乘客）新消息。
function appendMessage(msg) {
  if (!conversationId.value || msg.conversationId !== conversationId.value) return
  if (messages.value.some((m) => (msg.id && m.id === msg.id) || (msg.clientMsgId && m.clientMsgId === msg.clientMsgId))) {
    return
  }
  messages.value.push(msg)
  scrollToBottom()
}

async function doSend(content, msgType = 1) {
  const text = (content ?? inputText.value).trim()
  if (!text) return
  if (!conversationId.value) {
    showToast('会话未就绪')
    return
  }
  const cid = genClientMsgId()
  messages.value.push({
    id: 0,
    conversationId: conversationId.value,
    orderId: props.order?.orderId,
    senderType: 1,
    senderId: 0,
    msgType,
    content: text,
    clientMsgId: cid,
    createAt: Math.floor(Date.now() / 1000),
    _tmp: true
  })
  inputText.value = ''
  scrollToBottom()
  sending.value = true
  try {
    const res = await sendMessage(conversationId.value, msgType, text, cid)
    const idx = messages.value.findIndex((m) => m.clientMsgId === cid)
    if (idx >= 0) {
      messages.value[idx].id = res.messageId
      messages.value[idx].createAt = res.createAt
      messages.value[idx]._tmp = false
    }
  } catch (e) {
    messages.value = messages.value.filter((m) => m.clientMsgId !== cid)
    showToast('发送失败，请重试')
  } finally {
    sending.value = false
  }
}

function onSend() {
  doSend(inputText.value, 1)
}

function onQuick(q) {
  doSend(q, 2)
}

watch(
  () => props.visible,
  (v) => {
    if (v && props.order) {
      loadConversation()
    }
  }
)

defineExpose({ appendMessage })
</script>

<style scoped>
.chat-panel {
  display: flex;
  flex-direction: column;
  background: var(--driver-soft);
}

.cp-header {
  display: grid;
  grid-template-columns: 44px minmax(0, 1fr);
  align-items: center;
  gap: 8px;
  padding: 16px 16px 18px;
  color: var(--driver-on-primary);
  background: linear-gradient(135deg, var(--driver-primary) 0%, var(--driver-primary) 100%);
}
.cp-close {
  width: 32px;
  height: 32px;
  display: grid;
  place-items: center;
  border: 0;
  border-radius: 50%;
  background: rgba(255, 255, 255, .18);
  color: var(--driver-on-primary);
  font-size: 18px;
}
.cp-title { display: flex; flex-direction: column; gap: 4px; min-width: 0; }
.cp-title strong { font-size: 17px; font-weight: 800; }
.cp-title small { font-size: 12px; opacity: .85; }

.cp-locked {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 10px;
  color: var(--driver-muted);
}
.cp-locked .van-icon { font-size: 40px; }
.cp-locked p { font-size: 13px; }

.cp-messages {
  flex: 1;
  overflow-y: auto;
  padding: 14px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.cp-hint { text-align: center; color: var(--driver-muted); font-size: 12px; }

.cp-row { display: flex; }
.cp-row.mine { justify-content: flex-end; }
.cp-row.peer { justify-content: flex-start; }

.cp-bubble {
  max-width: 76%;
  padding: 9px 12px;
  border-radius: 14px;
  display: flex;
  flex-direction: column;
  gap: 4px;
  background: var(--driver-card);
  box-shadow: 0 4px 12px rgba(15, 23, 42, .04);
}
.cp-row.mine .cp-bubble {
  background: var(--driver-primary);
  color: var(--driver-on-primary);
}
.cp-text { font-size: 14px; line-height: 1.4; word-break: break-all; white-space: pre-wrap; }
.cp-time { font-size: 10px; opacity: .6; align-self: flex-end; }

.cp-quick {
  display: flex;
  gap: 8px;
  padding: 8px 12px;
  overflow-x: auto;
  border-top: 1px solid var(--driver-line);
  background: var(--driver-soft);
}
.cp-quick-item {
  flex: 0 0 auto;
  padding: 6px 12px;
  border: 1px solid var(--driver-line);
  border-radius: 16px;
  background: var(--driver-card);
  color: var(--driver-ink);
  font-size: 12px;
  white-space: nowrap;
}

.cp-input-bar {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: 8px;
  padding: 10px 12px calc(10px + env(safe-area-inset-bottom));
  background: var(--driver-card);
  border-top: 1px solid var(--driver-line);
}
.cp-input {
  min-width: 0;
  height: 40px;
  padding: 0 14px;
  border: 1px solid var(--driver-line);
  border-radius: 20px;
  background: var(--driver-soft);
  color: var(--driver-ink);
  font-size: 14px;
}
.cp-send {
  height: 40px;
  padding: 0 20px;
  border: 0;
  border-radius: 20px;
  background: var(--driver-primary);
  color: var(--driver-on-primary);
  font-size: 14px;
  font-weight: 700;
}
.cp-send:disabled { opacity: .6; }
</style>
