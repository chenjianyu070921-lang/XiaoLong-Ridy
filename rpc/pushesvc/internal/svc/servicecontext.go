package svc

import (
	"XiaoLong-Ridy/rpc/pushesvc/internal/config"
	"XiaoLong-Ridy/rpc/pushesvc/internal/model"

	"gorm.io/gorm"
)

type ServiceContext struct {
	Config       config.Config
	Db           *gorm.DB
	NoticeModel  *model.NoticeModel
	PushLogModel *model.PushLogModel
}

func NewServiceContext(c config.Config, db *gorm.DB) *ServiceContext {
	// 自动建站内信表 / 推送日志表（表已存在时不会重复建）
	if err := db.AutoMigrate(&model.Notice{}, &model.PushLog{}); err != nil {
		panic(err)
	}

	return &ServiceContext{
		Config:       c,
		Db:           db,
		NoticeModel:  model.NewNoticeModel(db),
		PushLogModel: model.NewPushLogModel(db),
	}
}
