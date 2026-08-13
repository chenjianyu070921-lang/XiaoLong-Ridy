package svc

import (
	"sync"

	"XiaoLong-Ridy/rpc/locationsvc/internal/config"
	"XiaoLong-Ridy/rpc/locationsvc/internal/geo"
	"XiaoLong-Ridy/rpc/locationsvc/internal/model"

	"gorm.io/gorm"
)

type ServiceContext struct {
	mu       sync.RWMutex
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

// GetConfig 加锁读取当前配置，配合配置热更新使用
func (s *ServiceContext) GetConfig() config.Config {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Config
}

// GetGeo 加锁读取地图客户端，配置热更新后会重建新的客户端
func (s *ServiceContext) GetGeo() *geo.Client {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Geo
}

// UpdateConfig 配置热更新：替换配置并重建地图客户端
func (s *ServiceContext) UpdateConfig(c config.Config) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Config = c
	s.Geo = geo.NewClient(c.MapService)
}
