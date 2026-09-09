package model

import "time"

// Message 对应 im_message 表：单条聊天消息。
// client_msg_id 唯一，用于客户端幂等去重。
type Message struct {
	Id             uint64    `gorm:"primaryKey;column:id;autoIncrement"`
	ConversationId int64     `gorm:"column:conversation_id;index:idx_conversation,priority:2"`
	OrderId        string    `gorm:"column:order_id;size:64"`
	SenderType     int8      `gorm:"column:sender_type"` // 1司机 2乘客
	SenderId       int64     `gorm:"column:sender_id"`
	MsgType        int8      `gorm:"column:msg_type;default:1"` // 1文本 2快捷短语
	Content        string    `gorm:"column:content;size:500"`
	ClientMsgId    string    `gorm:"column:client_msg_id;size:64;uniqueIndex:uk_client_msg"`
	CreateAt       time.Time `gorm:"column:create_at"`
}

// TableName 指定 gorm 映射的表名。
func (Message) TableName() string {
	return "im_message"
}
