package model

import "time"

// IMConversation 司乘聊天会话（与 scripts/sql/migrate/18_im_chat.sql 对齐）。
// 说明：order_id 用 BIGINT 对齐 ordersvc.order_id；order_no 存业务单号供显示。
type IMConversation struct {
	Id              int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderId         int64     `gorm:"column:order_id;not null;index:uk_order,unique" json:"orderId"`
	OrderNo         string    `gorm:"column:order_no;not null;default:''" json:"orderNo"`
	DriverId        int64     `gorm:"column:driver_id;not null;index" json:"driverId"`
	PassengerId     int64     `gorm:"column:passenger_id;not null;index" json:"passengerId"`
	Status          int8      `gorm:"column:status;not null;default:1" json:"status"` // 1进行中 2已归档 3已关闭
	LastMsg         string    `gorm:"column:last_msg;not null;default:''" json:"lastMsg"`
	LastMsgAt       *time.Time `gorm:"column:last_msg_at" json:"lastMsgAt"`
	UnreadDriver    int       `gorm:"column:unread_driver;not null;default:0" json:"unreadDriver"`
	UnreadPassenger int       `gorm:"column:unread_passenger;not null;default:0" json:"unreadPassenger"`
	CreatedAt       time.Time `gorm:"column:create_at;autoCreateTime" json:"createdAt"`
	UpdatedAt       time.Time `gorm:"column:update_at;autoUpdateTime" json:"updatedAt"`
}

func (IMConversation) TableName() string { return "im_conversation" }

// IMMessage 司乘聊天消息。client_msg_id 唯一，保证发送幂等。
type IMMessage struct {
	Id             int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	ConversationId int64     `gorm:"column:conversation_id;not null;index:idx_conversation" json:"conversationId"`
	OrderId        int64     `gorm:"column:order_id;not null" json:"orderId"`
	SenderType     int8      `gorm:"column:sender_type;not null" json:"senderType"` // 1司机 2乘客
	SenderId       int64     `gorm:"column:sender_id;not null" json:"senderId"`
	MsgType        int8      `gorm:"column:msg_type;not null;default:1" json:"msgType"` // 1文本 2快捷短语
	Content        string    `gorm:"column:content;not null" json:"content"`
	ClientMsgId    string    `gorm:"column:client_msg_id;not null;uniqueIndex:uk_client_msg" json:"clientMsgId"`
	CreatedAt      time.Time `gorm:"column:create_at;autoCreateTime" json:"createdAt"`
}

func (IMMessage) TableName() string { return "im_message" }
