package model

import "time"

// Conversation 对应 im_conversation 表：司乘聊天的会话聚合。
// 会话在司机接单后懒创建，按订单号唯一。
type Conversation struct {
	Id              uint64     `gorm:"primaryKey;column:id;autoIncrement"`
	OrderId         string     `gorm:"column:order_id;size:64;uniqueIndex:uk_order"`
	DriverId        int64      `gorm:"column:driver_id;default:0"`
	PassengerId     int64      `gorm:"column:passenger_id;default:0"`
	Status          int8       `gorm:"column:status;default:1"` // 1进行中 2已归档 3已关闭
	LastMsg         string     `gorm:"column:last_msg;size:500;default:''"`
	LastMsgAt       *time.Time `gorm:"column:last_msg_at"`
	UnreadDriver    int        `gorm:"column:unread_driver;default:0"`
	UnreadPassenger int        `gorm:"column:unread_passenger;default:0"`
	CreateAt        time.Time  `gorm:"column:create_at"`
}

// TableName 指定 gorm 映射的表名。
func (Conversation) TableName() string {
	return "im_conversation"
}
