// 乘客端聊天 WebSocket 客户端。
// 连接乘客网关 /api/passenger/v1/ws/chat，复用后端协议（snake_case 帧）：
//   上行：{ type:'chat.message', room_id, client_msg_id, message_type, content }
//   下行：{ type:'chat.message', message_id, room_id, sender_id, sender_type, message_type, content, client_msg_id, timestamp }
//         { type:'ping' } / { type:'pong' } / { type:'connected' } / { type:'auth_failed' } / { type:'error' }
//
// 断线重连按 1s/2s/5s/10s 退避；30s 心跳；重连成功后由页面侧按 cursor 补拉遗漏消息。
// 注意：vite 代理未开启 ws 转发，开发期直连乘客网关端口（默认 8091）；生产用 VITE_WS_BASE 覆盖。

const DEFAULT_WS_PORT = 8091

function buildWsUrl(path) {
  const proto = location.protocol === 'https:' ? 'wss' : 'ws'
  const base =
    import.meta.env.VITE_WS_BASE || `${proto}://${location.hostname}:${DEFAULT_WS_PORT}`
  return `${base}/api/passenger/v1${path}`
}

function genClientMsgId() {
  return 'cmsg_' + Date.now().toString(36) + '_' + Math.random().toString(36).slice(2, 8)
}

export class ChatSocket {
  constructor({ orderId, token, onMessage, onStatus }) {
    this.orderId = orderId
    this.token = token
    this.onMessage = onMessage || (() => {})
    this.onStatus = onStatus || (() => {})
    this.ws = null
    this.closedByUser = false
    this.reconnectAttempts = 0
    this.heartbeatTimer = null
    this.reconnectTimer = null
  }

  connect() {
    if (
      this.ws &&
      (this.ws.readyState === WebSocket.OPEN || this.ws.readyState === WebSocket.CONNECTING)
    ) {
      return
    }
    const url = buildWsUrl(
      `/ws/chat?orderId=${encodeURIComponent(this.orderId)}&token=${encodeURIComponent(this.token)}`
    )
    let ws
    try {
      ws = new WebSocket(url)
    } catch (e) {
      this.scheduleReconnect()
      return
    }
    this.ws = ws
    ws.onopen = () => {
      this.reconnectAttempts = 0
      this.onStatus({ type: 'open' })
      this.startHeartbeat()
    }
    ws.onmessage = (evt) => this.handleFrame(evt.data)
    ws.onclose = () => {
      this.stopHeartbeat()
      if (this.closedByUser) {
        this.onStatus({ type: 'closed' })
      } else {
        this.scheduleReconnect()
      }
    }
    ws.onerror = () => {
      this.onStatus({ type: 'error' })
    }
  }

  handleFrame(raw) {
    let frame
    try {
      frame = JSON.parse(raw)
    } catch {
      return
    }
    const t = frame.type
    if (t === 'ping') {
      this.sendRaw({ type: 'pong' })
      return
    }
    if (t === 'chat.message') {
      this.onMessage({
        messageId: frame.message_id,
        roomId: frame.room_id,
        senderId: frame.sender_id,
        senderType: frame.sender_type, // 'passenger' | 'driver'
        messageType: frame.message_type, // 'text' | 'location' | ...
        content: frame.content,
        clientMsgId: frame.client_msg_id,
        timestamp: frame.timestamp
      })
      return
    }
    // connected / auth_failed / error / push_unavailable / pong
    this.onStatus({ type: t, message: frame.message })
    if (t === 'auth_failed') this.close()
  }

  startHeartbeat() {
    this.stopHeartbeat()
    this.heartbeatTimer = setInterval(() => this.sendRaw({ type: 'ping' }), 30000)
  }

  stopHeartbeat() {
    if (this.heartbeatTimer) {
      clearInterval(this.heartbeatTimer)
      this.heartbeatTimer = null
    }
  }

  sendRaw(obj) {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(obj))
      return true
    }
    return false
  }

  // 发送一条消息，返回本地生成的 clientMsgId 供页面乐观渲染对账；ok 表示当前连接是否可发送。
  send(content, messageType = 'text') {
    const clientMsgId = genClientMsgId()
    const ok = this.sendRaw({
      type: 'chat.message',
      room_id: Number(this.orderId),
      client_msg_id: clientMsgId,
      message_type: messageType,
      content
    })
    return { clientMsgId, ok }
  }

  scheduleReconnect() {
    if (this.closedByUser) return
    this.reconnectAttempts++
    const delays = [1000, 2000, 5000, 10000]
    const delay = delays[Math.min(this.reconnectAttempts - 1, delays.length - 1)]
    this.onStatus({ type: 'reconnecting', delay })
    clearTimeout(this.reconnectTimer)
    this.reconnectTimer = setTimeout(() => this.connect(), delay)
  }

  close() {
    this.closedByUser = true
    this.stopHeartbeat()
    clearTimeout(this.reconnectTimer)
    if (this.ws) {
      try {
        this.ws.close()
      } catch (e) {
        // ignore
      }
      this.ws = null
    }
  }
}

export default ChatSocket
