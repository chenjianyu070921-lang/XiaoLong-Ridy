<template>
  <div class="chat-page">
    <van-nav-bar title="联系司机" left-arrow @click-left="onBack" />

    <!-- 未开启：司机接单前 -->
    <div v-if="!opened" class="not-open">
      <van-icon name="chat-o" size="40" color="#cbd5e1" />
      <p>司机接单后即可在订单内与司机沟通</p>
    </div>

    <!-- 司机信息头部 -->
    <div v-else class="driver-header">
      <img class="avatar" :src="peer.avatar || '/default-avatar.png'" alt="司机头像" />
      <div class="meta">
        <div class="name">{{ peer.name || '司机' }}</div>
        <div class="plate">{{ peer.plateNo || '' }}</div>
      </div>
      <a v-if="peer.phone" class="call" :href="`tel:${peer.phone}`">
        <van-icon name="phone-o" size="18" />
        <span>电话</span>
      </a>
    </div>

    <!-- 消息列表 -->
    <div v-if="opened" class="msg-list" ref="msgListRef">
      <div v-if="loadingHistory" class="hist-tip">加载中…</div>
      <div
        v-for="m in messages"
        :key="m.clientMsgId || m.id"
        class="msg-row"
        :class="{ mine: m.mine }"
      >
        <div class="bubble" :class="m.status">
          <span>{{ m.content }}</span>
          <span v-if="m.mine && m.status === 'sending'" class="status">发送中…</span>
          <span v-else-if="m.mine && m.status === 'failed'" class="status failed" @click="resend(m)">
            发送失败·重试
          </span>
        </div>
        <div class="time">{{ formatTime(m.time) }}</div>
      </div>
    </div>

    <!-- 只读提示 -->
    <div v-if="opened && !canSend" class="readonly-tip">订单已结束，聊天室仅可查看历史</div>

    <!-- 输入栏 + 快捷短语 -->
    <div v-if="opened && canSend" class="composer safe-area-bottom">
      <div class="quick-row">
        <span
          v-for="p in phrases"
          :key="p.id"
          class="quick-chip"
          @click="sendText(p.content, 2)"
        >{{ p.content }}</span>
      </div>
      <div class="input-bar">
        <input
          class="draft"
          v-model="draft"
          placeholder="输入消息告诉司机位置…"
          @keyup.enter="sendText(draft, 1)"
        />
        <button class="send-btn" @click="sendText(draft, 1)">发送</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { showToast } from 'vant'
import { useUserStore } from '@/stores/user'
import { formatDriverDisplayName, formatPlateNumber } from '@/constants/order'
import { getConversation, getMessages, sendMessage, markRead, getQuickPhrases } from '@/api/chat'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

const orderId = Number(route.params.orderId || 0)

const opened = ref(false)
const canSend = ref(false)
const conversationId = ref(0)
const messages = ref([])
const draft = ref('')
const loadingHistory = ref(false)
const phrases = ref([])
const peer = ref({ name: '司机信息加载中', avatar: '', plateNo: '车牌信息加载中', phone: '' })
const msgListRef = ref(null)
let pollTimer = null

const genClientMsgId = () =>
  'c_' + Date.now().toString(36) + Math.random().toString(36).slice(2, 8)

const formatTime = (ts) => {
  if (!ts) return ''
  const d = new Date(ts)
  return `${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
}

const scrollToBottom = async () => {
  await nextTick()
  const el = msgListRef.value
  if (el) el.scrollTop = el.scrollHeight
}

const upsertMessage = (msg) => {
  const list = messages.value
  const idx = list.findIndex((m) => m.clientMsgId === msg.clientMsgId)
  if (idx >= 0) {
    list[idx] = { ...list[idx], ...msg }
  } else if (!list.some((m) => m.id && msg.id && m.id === msg.id)) {
    list.push(msg)
  }
}

// 拉取最新一页并合并（按 clientMsgId/id 去重）。
const pollMessages = async () => {
  if (!conversationId.value) return
  try {
    const res = await getMessages(conversationId.value, 0, 20)
    const list = Array.isArray(res?.messages) ? res.messages : []
    const atBottom = msgListRef.value
      ? msgListRef.value.scrollHeight - msgListRef.value.scrollTop - msgListRef.value.clientHeight < 60
      : true
    for (const it of list) {
      upsertMessage({
        id: it.id,
        clientMsgId: it.clientMsgId,
        mine: it.senderType === 2,
        content: it.content,
        time: (it.createAt || 0) * 1000,
        status: 'received'
      })
    }
    if (atBottom) await scrollToBottom()
  } catch (e) {
    // 轮询失败静默重试
  }
}

const loadHistory = async () => {
  loadingHistory.value = true
  try {
    await pollMessages()
    await markRead(conversationId.value)
  } catch (e) {
    console.error('加载聊天历史失败:', e)
  } finally {
    loadingHistory.value = false
  }
}

const doSend = async (content, msgType) => {
  const text = String(content || '').trim()
  if (!text || !canSend.value) return
  const clientMsgId = genClientMsgId()
  const optimistic = {
    id: null,
    clientMsgId,
    mine: true,
    content: text,
    time: Date.now(),
    status: 'sending'
  }
  messages.value.push(optimistic)
  draft.value = ''
  await scrollToBottom()
  try {
    const res = await sendMessage(conversationId.value, msgType, text, clientMsgId)
    const idx = messages.value.findIndex((m) => m.clientMsgId === clientMsgId)
    if (idx >= 0) {
      messages.value[idx].id = res.messageId
      messages.value[idx].status = 'sent'
      messages.value[idx].time = (res.createAt || Math.floor(Date.now() / 1000)) * 1000
    }
  } catch (e) {
    const code = e?.response?.data?.code
    if (code === 46000) {
      showToast('消息含敏感信息，已被拦截')
    } else {
      showToast('发送失败')
    }
    const idx = messages.value.findIndex((m) => m.clientMsgId === clientMsgId)
    if (idx >= 0) messages.value[idx].status = 'failed'
  }
}

const sendText = (text, msgType) => doSend(text, msgType)

const resend = (m) => {
  m.status = 'sending'
  doSend(m.content, 1)
}

const onBack = () => {
  if (window.history.length > 1) router.back()
  else router.replace('/home')
}

onMounted(async () => {
  if (!orderId) {
    showToast('订单参数缺失')
    return
  }
  try {
    const conv = await getConversation(orderId)
    opened.value = !!conv?.opened
    conversationId.value = conv?.conversationId || 0
    // status 1 进行中可发送；2 已归档 / 3 已关闭 只读。
    canSend.value = opened.value && conv?.status === 1
    if (conv?.peer) {
      peer.value = {
        name: formatDriverDisplayName(conv.peer.name, conv.peer.id),
        avatar: '',
        plateNo: formatPlateNumber(''),
        phone: ''
      }
    }
    if (opened.value) {
      await loadHistory()
      pollTimer = setInterval(pollMessages, 3000)
    }
  } catch (e) {
    console.error('获取会话失败:', e)
    showToast('会话加载失败')
  }
  // 快捷短语
  try {
    const p = await getQuickPhrases()
    phrases.value = Array.isArray(p?.phrases) ? p.phrases : []
  } catch (e) {
    phrases.value = []
  }
})

onUnmounted(() => {
  if (pollTimer) clearInterval(pollTimer)
})
</script>

<style scoped>
.chat-page {
  display: flex;
  flex-direction: column;
  height: 100vh;
  background: #f5f5f5;
}
.not-open {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  color: #94a3b8;
  font-size: 14px;
}
.driver-header {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  background: #fff;
  border-bottom: 1px solid #eee;
}
.driver-header .avatar {
  width: 44px;
  height: 44px;
  border-radius: 50%;
  object-fit: cover;
  background: #e5e7eb;
}
.driver-header .meta {
  flex: 1;
}
.driver-header .name {
  font-size: 16px;
  font-weight: 600;
  color: #111;
}
.driver-header .plate {
  font-size: 13px;
  color: #6b7280;
}
.driver-header .call {
  display: flex;
  flex-direction: column;
  align-items: center;
  color: #3b82f6;
  font-size: 11px;
  text-decoration: none;
}
.msg-list {
  flex: 1;
  overflow-y: auto;
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.hist-tip {
  text-align: center;
  color: #9ca3af;
  font-size: 12px;
}
.msg-row {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  max-width: 80%;
}
.msg-row.mine {
  align-self: flex-end;
  align-items: flex-end;
}
.bubble {
  padding: 9px 12px;
  border-radius: 12px;
  background: #fff;
  color: #111;
  font-size: 14px;
  line-height: 1.4;
  word-break: break-word;
  white-space: pre-wrap;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
}
.msg-row.mine .bubble {
  background: #07c160;
  color: #fff;
}
.bubble.failed {
  background: #fde2e2;
  color: #b91c1c;
}
.status {
  display: block;
  margin-top: 4px;
  font-size: 11px;
  opacity: 0.8;
}
.status.failed {
  color: #b91c1c;
  text-decoration: underline;
}
.time {
  margin-top: 3px;
  font-size: 10px;
  color: #9ca3af;
}
.readonly-tip {
  padding: 12px;
  text-align: center;
  color: #9ca3af;
  background: #fff;
  border-top: 1px solid #eee;
}
.composer {
  background: #fff;
  border-top: 1px solid #eee;
}
.quick-row {
  display: flex;
  gap: 8px;
  padding: 8px 12px 0;
  overflow-x: auto;
}
.quick-chip {
  flex-shrink: 0;
  padding: 5px 12px;
  border-radius: 16px;
  background: #eff6ff;
  color: #2563eb;
  font-size: 13px;
  white-space: nowrap;
}
.input-bar {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
}
.input-bar .draft {
  flex: 1;
  height: 38px;
  border: 1px solid #e5e7eb;
  border-radius: 19px;
  padding: 0 14px;
  font-size: 14px;
  outline: none;
}
.input-bar .send-btn {
  flex-shrink: 0;
  height: 38px;
  padding: 0 18px;
  border: none;
  border-radius: 19px;
  background: #07c160;
  color: #fff;
  font-size: 14px;
}
</style>
