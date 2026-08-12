package svc

import (
	"XiaoLong-Ridy/rpc/locationsvc/internal/config"
	"XiaoLong-Ridy/rpc/locationsvc/internal/geo"
	"XiaoLong-Ridy/rpc/locationsvc/internal/model"

	"gorm.io/gorm"
)

type ServiceContext struct {
	Config   config.Config
	Db       *gorm.DB
	Geo      *geo.Client
	PoiModel *model.PoiModel
}

func NewServiceContext(c config.Config, db *gorm.DB) *ServiceContext {
	// 自动创建 poi 缓存表（表已存在时不会重复建，也不会动已有表）
	if err := db.AutoMigrate(&model.Poi{}); err != nil {
		panic(err)
	}

	return &ServiceContext{
		Config:   c,
		Db:       db,
		Geo:      geo.NewClient(c.MapService),
		PoiModel: model.NewPoiModel(db),
	}
}
