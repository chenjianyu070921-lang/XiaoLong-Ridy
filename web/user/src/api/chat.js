import request from './request'

// 乘客端司乘聊天（IM）接口封装，统一方案 V1.0。
// 乘客端 MVP 采用轮询（3s 拉取），底层由 rpc/chatsvc 提供能力。

// 获取/懒创建会话（司机接单后才有 opened=true）。
export function getConversation(orderId) {
  return request.get('/chat/conversation', { params: { order_id: orderId } })
}

// 游标分页拉取历史消息；cursor=0 取最新一页，返回 { messages, nextCursor, hasMore }。
export function getMessages(conversationId, cursor = 0, limit = 20) {
  return request.get('/chat/messages', {
    params: { conversation_id: conversationId, cursor, limit }
  })
}

// 发送消息。conversationId/msgType/content/clientMsgId 必填，clientMsgId 由前端生成用于幂等。
export function sendMessage(conversationId, msgType, content, clientMsgId) {
  return request.post('/chat/send', {
    conversationId,
    msgType,
    content,
    clientMsgId
  })
}

// 标记会话已读。
export function markRead(conversationId) {
  return request.post('/chat/read', { conversationId })
}

// 获取 MVP 快捷短语（5 条）。
export function getQuickPhrases() {
  return request.get('/chat/phrases')
}
