<template>
  <main class="driver-order-chat-page">
    <header class="oc-header">
      <button type="button" class="oc-back" aria-label="返回" @click="goBack">
        <van-icon name="arrow-left" />
      </button>
      <div class="oc-title">
        <strong>与乘客沟通</strong>
        <small v-if="peer">{{ peer.name }} · {{ orderNo }}</small>
        <small v-else-if="orderNo">订单号 {{ orderNo }}</small>
        <small v-else>仅本订单乘客可见</small>
      </div>
    </header>

    <section v-if="loading" class="oc-state">
      <p>加载中…</p>
    </section>

    <section v-else-if="loadError" class="oc-state oc-state-error">
      <van-icon name="warning-o" />
      <p>{{ loadError }}</p>
      <button type="button" class="oc-back-btn" @click="goBack">返回</button>
    </section>

    <div v-else-if="!opened" class="oc-locked">
      <van-icon name="comment-o" />
      <p v-if="convStatus === 5">行程已结束，聊天已归档不可发送</p>
      <p v-else-if="convStatus === 6 || convStatus === 7">订单已关闭，无法继续沟通</p>
      <p v-else>接单成功后可在此与乘客沟通</p>
    </div>

    <template v-else>
      <div ref="listEl" class="oc-messages">
        <div
          v-for="m in messages"
          :key="m.clientMsgId || m.id"
          class="oc-row"
          :class="m.senderType === SENDER_TYPE_DRIVER ? 'mine' : 'peer'"
        >
          <div class="oc-avatar-col">
            <img v-if="avatarFor(m)" :src="avatarFor(m)" class="oc-avatar" alt="" />
            <div v-else class="oc-avatar oc-avatar-fallback">{{ avatarText(m) }}</div>
          </div>
          <div class="oc-bubble-col">
            <div class="oc-name">{{ nameFor(m) }}</div>
            <div class="oc-bubble">
              <span class="oc-text">{{ m.content }}</span>
              <span class="oc-time">{{ formatTime(m.createAt) }}</span>
            </div>
          </div>
        </div>
      </div>

      <div class="oc-quick">
        <button
          v-for="q in quickReplies"
          :key="q"
          type="button"
          class="oc-quick-item"
          @click="onQuick(q)"
        >{{ q }}</button>
      </div>

      <footer class="oc-input-bar">
        <input
          v-model="inputText"
          class="oc-input"
          type="text"
          maxlength="200"
          placeholder="输入消息，禁止线下交易/交换联系方式"
          @keyup.enter="onSend"
        />
        <button type="button" class="oc-send" :disabled="sending" @click="onSend">发送</button>
      </footer>
    </template>
  </main>
</template>

<script setup>
// 独立私信页：从 /chat/:orderId 进入，仅本订单司机与乘客可访问（后端 GetOrCreateConversation
// 用 ordersvc 校验订单参与方，非参与方返回 40301）。复用 chatsvc（经 @/api/chat 网关）。
// 实时：复用司机全局 WS（DriverHome 持有，经 App.vue keep-alive 跨路由存活），消息统一收口到
// driverChat store，本页直接订阅 byOrder[orderId]，不再单独开 WS。仅当深链接直达本站
// （DriverHome 未挂载、realtimeActive=false）时，才兜底开一个紧凑 WS 写入同一 store。
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { showToast } from 'vant'
import { getConversation, listMessages, sendMessage, markRead, genClientMsgId, MSG_TYPE_TEXT, MSG_TYPE_QUICK, SENDER_TYPE_DRIVER } from '@/api/chat'
import { useDriverChatStore } from '@/stores/driverChat'
import { useDriverStore } from '@/stores/driver'

const route = useRoute()
const router = useRouter()
const driverChat = useDriverChatStore()
const driverStore = useDriverStore()
const orderId = computed(() => Number(route.params.orderId) || 0)

const loading = ref(true)
const loadError = ref('')
const opened = ref(false)
const convStatus = ref(0)
const conversationId = ref(0)
const peer = ref(null)
const orderNo = ref('')
const inputText = ref('')
const sending = ref(false)
const listEl = ref(null)

// 统一数据源：直接订阅共享 store（DriverHome 的 WS 推送会写入这里）
const messages = computed(() => driverChat.byOrder[orderId.value] || [])

// 双方身份：司机取自身 store，乘客取会话 peer（头像后端未透传时回退首字占位）
const myName = computed(() => driverStore.displayName || '我')
const myAvatar = computed(() => driverStore.driver?.avatarUrl || '')
const peerName = computed(() => peer.value?.name || '乘客')
const peerAvatar = computed(() => peer.value?.avatar || '')

function isMine(m) { return m && m.senderType === SENDER_TYPE_DRIVER }
function avatarFor(m) { return isMine(m) ? myAvatar.value : peerAvatar.value }
function nameFor(m) { return isMine(m) ? myName.value : peerName.value }
function avatarText(m) {
  const n = nameFor(m) || ''
  return n ? Array.from(n)[0] : '?'
}

const quickReplies = ['我马上到', '我在附近，请稍等', '请到上车点等待', '我已到达上车点', '请打开车门']

function goBack() {
  if (window.history.length > 1) router.back()
  else router.replace('/home')
}

function scrollToBottom() {
  nextTick(() => {
    const el = listEl.value
    if (el) el.scrollTop = el.scrollHeight
  })
}
watch(messages, scrollToBottom)

function formatTime(ts) {
  if (!ts) return ''
  const d = new Date(ts * 1000)
  const hh = String(d.getHours()).padStart(2, '0')
  const mm = String(d.getMinutes()).padStart(2, '0')
  return `${hh}:${mm}`
}

function isNearBottom() {
  const el = listEl.value
  if (!el) return true
  return el.scrollHeight - el.scrollTop - el.clientHeight < 80
}

async function loadMessages() {
  const res = await listMessages(conversationId.value, 0, 20)
  const server = res.messages || []
  // 保留本地乐观消息（_tmp 尚未被服务端确认），避免轮询覆盖未确认发送
  const list = driverChat.byOrder[orderId.value] || []
  const localTmp = list.filter((m) => m._tmp)
  const seen = new Set(server.map((m) => (m.id ? 'id:' + m.id : 'c:' + m.clientMsgId)))
  const merged = server.slice()
  for (const t of localTmp) {
    const k = t.id ? 'id:' + t.id : 'c:' + t.clientMsgId
    if (!seen.has(k)) merged.push(t)
  }
  driverChat.seedMessages(orderId.value, merged)
  try { await markRead(conversationId.value) } catch (e) { /* 标记已读失败不阻断 */ }
  if (isNearBottom()) scrollToBottom()
}

async function loadConversation() {
  if (!orderId.value) {
    loadError.value = '订单参数无效'
    loading.value = false
    return
  }
  loading.value = true
  loadError.value = ''
  opened.value = false
  conversationId.value = 0
  try {
    const conv = await getConversation(orderId.value)
    conversationId.value = conv.conversationId || 0
    opened.value = !!conv.opened
    convStatus.value = conv.status
    peer.value = conv.peer || null
    orderNo.value = conv.orderNo || ''
    if (opened.value && conversationId.value) {
      await loadMessages()
      startPolling()
    }
  } catch (e) {
    // 后端对非订单参与方返回 40301，由 chat.js 拦截器统一 toast，这里再覆盖为页面级提示
    loadError.value = (e && e.message) || '会话加载失败'
  } finally {
    loading.value = false
  }
}

async function doSend(content, msgType = MSG_TYPE_TEXT) {
  const text = (content ?? inputText.value).trim()
  if (!text) return
  if (!conversationId.value) {
    showToast('会话未就绪')
    return
  }
  const cid = genClientMsgId()
  const temp = {
    id: 0,
    conversationId: conversationId.value,
    orderId: orderId.value,
    senderType: SENDER_TYPE_DRIVER,
    senderId: 0,
    msgType,
    content: text,
    clientMsgId: cid,
    createAt: Math.floor(Date.now() / 1000),
    _tmp: true
  }
  driverChat.appendMessage(temp)
  inputText.value = ''
  scrollToBottom()
  sending.value = true
  try {
    const res = await sendMessage(conversationId.value, msgType, text, cid)
    const list = driverChat.byOrder[orderId.value] || []
    const idx = list.findIndex((m) => m.clientMsgId === cid)
    if (idx >= 0) {
      list[idx].id = res.messageId
      list[idx].createAt = res.createAt
      list[idx]._tmp = false
    }
  } catch (e) {
    const list = driverChat.byOrder[orderId.value]
    if (list) {
      const filtered = list.filter((m) => m.clientMsgId !== cid)
      driverChat.byOrder[orderId.value] = filtered
    }
    showToast('发送失败，请重试')
  } finally {
    sending.value = false
  }
}

function onSend() { doSend(inputText.value, MSG_TYPE_TEXT) }
function onQuick(q) { doSend(q, MSG_TYPE_QUICK) }

// ─── 深链接兜底实时（仅当 DriverHome 未挂载、全局 WS 未在线时）───
let pushSocket = null
let reconnectTimer = null
let reconnectAttempts = 0
let intentionalClose = false

function buildPushUrl() {
  const token = localStorage.getItem('driverToken') || ''
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  return protocol + '//' + window.location.host + '/api/driver/v1/ws?token=' + encodeURIComponent(token)
}

function connectFallbackPush() {
  if (intentionalClose || pushSocket || driverChat.realtimeActive) return
  const token = localStorage.getItem('driverToken')
  if (!token) return
  let socket
  try { socket = new WebSocket(buildPushUrl()) } catch { return }
  pushSocket = socket
  socket.onmessage = (ev) => {
    try {
      const payload = JSON.parse(ev.data)
      if (payload && payload.type === 'chat.message' && Number(payload.orderId) === orderId.value && payload.message) {
        driverChat.appendMessage(payload.message)
      }
    } catch { /* ignore */ }
  }
  socket.onclose = () => {
    if (pushSocket === socket) pushSocket = null
    if (!intentionalClose) scheduleReconnect()
  }
  socket.onerror = () => { try { socket.close() } catch { /* ignore */ } }
}

function scheduleReconnect() {
  if (intentionalClose || reconnectTimer || driverChat.realtimeActive) return
  const delay = Math.min(1000 * Math.pow(2, reconnectAttempts), 15000)
  reconnectAttempts += 1
  reconnectTimer = window.setTimeout(() => {
    reconnectTimer = null
    connectFallbackPush()
  }, delay)
}

function disconnectFallbackPush() {
  intentionalClose = true
  if (reconnectTimer) { window.clearTimeout(reconnectTimer); reconnectTimer = null }
  if (pushSocket) {
    const s = pushSocket
    pushSocket = null
    try { s.close() } catch { /* ignore */ }
  }
}

let pollTimer = null
function startPolling() {
  stopPolling()
  if (!opened.value || !conversationId.value) return
  pollTimer = window.setInterval(() => {
    if (opened.value && conversationId.value) loadMessages().catch(() => {})
  }, 3000)
}
function stopPolling() {
  if (pollTimer) { window.clearInterval(pollTimer); pollTimer = null }
}

onMounted(() => {
  driverChat.setActive(orderId.value)
  loadConversation()
  // 正常流程（自首页进入）DriverHome 已被 keep-alive 保活，全局 WS 在线；
  // 仅深链接直达本站时 realtimeActive=false，此处兜底开一个紧凑 WS。
  intentionalClose = false
  reconnectAttempts = 0
  if (!driverChat.realtimeActive) connectFallbackPush()
})

onBeforeUnmount(() => {
  driverChat.clearActive(orderId.value)
  stopPolling()
  disconnectFallbackPush()
})
</script>

<style scoped>
.driver-order-chat-page {
  display: flex;
  flex-direction: column;
  height: 100vh;
  background: var(--driver-soft);
  width: min(100vw, 390px);
  margin: 0 auto;
}

.oc-header {
  display: grid;
  grid-template-columns: 44px minmax(0, 1fr);
  align-items: center;
  gap: 8px;
  padding: 16px 16px 18px;
  color: var(--driver-on-primary);
  background: linear-gradient(135deg, var(--driver-primary) 0%, var(--driver-primary) 100%);
}
.oc-back {
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
.oc-title { display: flex; flex-direction: column; gap: 4px; min-width: 0; }
.oc-title strong { font-size: 17px; font-weight: 800; }
.oc-title small { font-size: 12px; opacity: .85; }

.oc-state {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  color: var(--driver-muted);
  padding: 24px;
  text-align: center;
}
.oc-state-error { color: var(--driver-ink); }
.oc-state-error .van-icon { font-size: 40px; color: #f56c6c; }
.oc-back-btn {
  margin-top: 8px;
  padding: 8px 24px;
  border: 0;
  border-radius: 20px;
  background: var(--driver-primary);
  color: var(--driver-on-primary);
  font-size: 14px;
}

.oc-locked {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 10px;
  color: var(--driver-muted);
  padding: 24px;
  text-align: center;
}
.oc-locked .van-icon { font-size: 40px; }
.oc-locked p { font-size: 13px; }

.oc-messages {
  flex: 1;
  overflow-y: auto;
  padding: 14px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.oc-row { display: flex; gap: 8px; align-items: flex-start; }
.oc-row.mine { flex-direction: row-reverse; }
.oc-row.peer { flex-direction: row; }

.oc-avatar-col { flex: 0 0 auto; }
.oc-avatar {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  object-fit: cover;
  display: block;
  background: var(--driver-line);
}
.oc-avatar-fallback {
  display: grid;
  place-items: center;
  background: var(--driver-primary);
  color: var(--driver-on-primary);
  font-size: 14px;
  font-weight: 700;
}

.oc-bubble-col {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
  max-width: calc(100% - 52px);
}
.oc-row.mine .oc-bubble-col { align-items: flex-end; }
.oc-name { font-size: 11px; color: var(--driver-muted); padding: 0 4px; }
.oc-row.mine .oc-name { text-align: right; }

.oc-bubble {
  max-width: 100%;
  padding: 9px 12px;
  border-radius: 14px;
  display: flex;
  flex-direction: column;
  gap: 4px;
  background: var(--driver-card);
  box-shadow: 0 4px 12px rgba(15, 23, 42, .04);
}
.oc-row.mine .oc-bubble {
  background: var(--driver-primary);
  color: var(--driver-on-primary);
}
.oc-text { font-size: 14px; line-height: 1.4; word-break: break-all; white-space: pre-wrap; }
.oc-time { font-size: 10px; opacity: .6; align-self: flex-end; }

.oc-quick {
  display: flex;
  gap: 8px;
  padding: 8px 12px;
  overflow-x: auto;
  border-top: 1px solid var(--driver-line);
  background: var(--driver-soft);
}
.oc-quick-item {
  flex: 0 0 auto;
  padding: 6px 12px;
  border: 1px solid var(--driver-line);
  border-radius: 16px;
  background: var(--driver-card);
  color: var(--driver-ink);
  font-size: 12px;
  white-space: nowrap;
}

.oc-input-bar {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: 8px;
  padding: 10px 12px calc(10px + env(safe-area-inset-bottom));
  background: var(--driver-card);
  border-top: 1px solid var(--driver-line);
}
.oc-input {
  min-width: 0;
  height: 40px;
  padding: 0 14px;
  border: 1px solid var(--driver-line);
  border-radius: 20px;
  background: var(--driver-soft);
  color: var(--driver-ink);
  font-size: 14px;
}
.oc-send {
  height: 40px;
  padding: 0 20px;
  border: 0;
  border-radius: 20px;
  background: var(--driver-primary);
  color: var(--driver-on-primary);
  font-size: 14px;
  font-weight: 700;
}
.oc-send:disabled { opacity: .6; }
</style>