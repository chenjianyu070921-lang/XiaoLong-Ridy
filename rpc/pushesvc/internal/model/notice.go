package model

import (
	"time"

	"gorm.io/gorm"
)

// Notice 站内信表（对应开发文档 notices 表）
type Notice struct {
	ID        uint64    `gorm:"column:id;primaryKey;autoIncrement"`
	UserID    uint64    `gorm:"column:user_id;not null;index:idx_user_id"`
	Title     string    `gorm:"column:title;size:200;not null"`
	Content   string    `gorm:"column:content;type:text;not null"`
	BizType   int8      `gorm:"column:biz_type;type:tinyint;default:1"`
	IsRead    int8      `gorm:"column:is_read;type:tinyint;default:0"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (Notice) TableName() string {
	return "notices"
}

// NoticeModel 站内信数据访问层
type NoticeModel struct {
	db *gorm.DB
}

func NewNoticeModel(db *gorm.DB) *NoticeModel {
	return &NoticeModel{db: db}
}

// Insert 写入一条站内信
func (m *NoticeModel) Insert(n *Notice) error {
	return m.db.Create(n).Error
}
