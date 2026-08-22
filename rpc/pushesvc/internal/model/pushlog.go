package model

import (
	"time"

	"gorm.io/gorm"
)

// PushLog 推送日志表（对应开发文档 push_log 表）
type PushLog struct {
	ID        uint64    `gorm:"column:id;primaryKey;autoIncrement"`
	UserID    uint64    `gorm:"column:user_id;default:0;index:idx_user_id"`
	PushType  int8      `gorm:"column:push_type;type:tinyint;not null"`
	Title     string    `gorm:"column:title;size:200"`
	Content   string    `gorm:"column:content;type:text;not null"`
	Target    string    `gorm:"column:target;size:100"`
	BizType   int8      `gorm:"column:biz_type;type:tinyint;default:0"`
	Result    int8      `gorm:"column:result;type:tinyint;default:0"`
	ErrorMsg  string    `gorm:"column:error_msg;size:500"`
	Extras    string    `gorm:"column:extras;type:text"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (PushLog) TableName() string {
	return "push_log"
}

// PushLogModel 推送日志数据访问层
type PushLogModel struct {
	db *gorm.DB
}

func NewPushLogModel(db *gorm.DB) *PushLogModel {
	return &PushLogModel{db: db}
}

// Insert 写入一条推送/短信日志
func (m *PushLogModel) Insert(l *PushLog) error {
	return m.db.Create(l).Error
}
