import { defineStore } from 'pinia'
import { reactive, ref } from 'vue'

// 司机端聊天的统一数据源 + 实时状态。
// DriverHome 拥有真实 WS 连接（随司机会话存活，经 App.vue 的 keep-alive 跨路由保持），
// 本 store 作为「首页聊天面板」与「独立私信页 /chat/:orderId」的共享数据层：
//  - DriverHome 收到 chat.message 推送给本 store.appendMessage
//  - 独立私信页直接订阅 byOrder[orderId]，不再各自开 WS（避免重复连接与重复 toast）
export const useDriverChatStore = defineStore('driverChat', () => {
  // 当前打开独立私信页的订单（同一时刻仅一个），用于抑制首页对该订单的重复提示
  const activeOrderId = ref(null)
  // 各订单聊天消息缓存，作为统一数据源 { [orderId]: Message[] }
  const byOrder = reactive({})
  // 全局司机 WS 是否在线（由 DriverHome 同步）
  const realtimeActive = ref(false)

  function ensureOrder(orderId) {
    if (!byOrder[orderId]) byOrder[orderId] = []
    return byOrder[orderId]
  }

  function appendMessage(msg) {
    if (!msg || !msg.orderId) return
    const list = ensureOrder(msg.orderId)
    const dup = list.some(
      (m) => (msg.id && m.id === msg.id) || (msg.clientMsgId && m.clientMsgId === msg.clientMsgId)
    )
    if (dup) return
    list.push(msg)
  }

  // 进入独立页 / 重新拉取时，用 REST 快照覆盖该订单缓存（保持单一数据源，避免与推送重复）
  function seedMessages(orderId, list) {
    byOrder[orderId] = (list || []).slice()
  }

  function setActive(orderId) {
    activeOrderId.value = orderId
  }
  function clearActive(orderId) {
    if (activeOrderId.value === orderId) activeOrderId.value = null
  }

  return { activeOrderId, byOrder, realtimeActive, ensureOrder, appendMessage, seedMessages, setActive, clearActive }
})