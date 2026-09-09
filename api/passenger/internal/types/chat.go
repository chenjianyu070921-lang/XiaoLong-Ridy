package types

// 司乘聊天（IM）统一方案 V1.0 枚举：与 rpc/chatsvc 完全一致，禁止第二套编号。
const (
	// ChatSenderDriver 发送方为司机。
	ChatSenderDriver int32 = 1
	// ChatSenderPassenger 发送方为乘客。
	ChatSenderPassenger int32 = 2
	// ChatMsgTypeText 文本消息。
	ChatMsgTypeText int32 = 1
	// ChatMsgTypeQuick 快捷短语。
	ChatMsgTypeQuick int32 = 2
)

// PeerInfo 对端（聊天对象）信息。
type PeerInfo struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// Conversation 会话视图。
type Conversation struct {
	ConversationID int64    `json:"conversationId"`
	OrderID        string   `json:"orderId"`
	Status         int32    `json:"status"` // 1进行中 2已归档 3已关闭
	Opened         bool     `json:"opened"` // 是否可聊天（司机接单后）
	LastMsg        string   `json:"lastMsg"`
	LastMsgAt      int64    `json:"lastMsgAt"`  // unix 秒
	Unread         int32    `json:"unread"`     // 调用方未读数
	SenderType     int32    `json:"senderType"` // 调用方类型：1司机 2乘客
	Peer           PeerInfo `json:"peer"`
}

// ChatMessage 单条消息。
type ChatMessage struct {
	ID             int64  `json:"id"`
	ConversationID int64  `json:"conversationId"`
	OrderID        string `json:"orderId"`
	SenderType     int32  `json:"senderType"` // 1司机 2乘客
	SenderID       int64  `json:"senderId"`
	MsgType        int32  `json:"msgType"` // 1文本 2快捷短语
	Content        string `json:"content"`
	ClientMsgID    string `json:"clientMsgId"`
	CreateAt       int64  `json:"createAt"` // unix 秒
}

// ListMessagesResponse 消息分页结果。
type ListMessagesResponse struct {
	Messages   []*ChatMessage `json:"messages"`
	NextCursor int64          `json:"nextCursor"`
	HasMore    bool           `json:"hasMore"`
}

// SendMessageResponse 发送结果。
type SendMessageResponse struct {
	MessageID int64 `json:"messageId"`
	CreateAt  int64 `json:"createAt"`
}

// MarkReadResponse 已读结果。
type MarkReadResponse struct {
	Ok bool `json:"ok"`
}

// ChatQuickPhrase 快捷短语。
type ChatQuickPhrase struct {
	ID      int32  `json:"id"`
	Content string `json:"content"`
}

// ChatQuickPhrasesResponse 快捷短语列表。
type ChatQuickPhrasesResponse struct {
	Phrases []ChatQuickPhrase `json:"phrases"`
}

// ChatSendRequest 发送消息请求体。
type ChatSendRequest struct {
	ConversationID int64  `json:"conversationId"`
	MsgType        int32  `json:"msgType"`
	Content        string `json:"content"`
	ClientMsgID    string `json:"clientMsgId"`
}

// ChatReadRequest 标记已读请求体。
type ChatReadRequest struct {
	ConversationID int64 `json:"conversationId"`
}
