package svc

import (
	"fmt"
	"sync"

	"XiaoLong-Ridy/common/constants"
	"XiaoLong-Ridy/rpc/locationsvc/internal/config"
	"XiaoLong-Ridy/rpc/locationsvc/internal/geo"
	"XiaoLong-Ridy/rpc/locationsvc/internal/model"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// DriverGeoKey Redis GEO 集合 key：司机实时位置（默认城市，与派单引擎/位置消费端保持一致）
const DriverGeoKey = "driver:geo:default"

// GeoKey 按城市构造 Redis GEO key，city 为空时使用默认城市。
func GeoKey(city string) string {
	if city == "" {
		city = "default"
	}
	return fmt.Sprintf(constants.RedisDriverGeo, city)
}

type ServiceContext struct {
	mu                  sync.RWMutex
	Config              config.Config
	Db                  *gorm.DB
	Redis               *redis.Client
	Geo                 *geo.Client
	PoiModel            *model.PoiModel
	DriverLocationModel *model.DriverLocationModel
	RideTrackPointModel *model.RideTrackPointModel
}

func NewServiceContext(c config.Config, db *gorm.DB, redisClient *redis.Client) *ServiceContext {
	// 自动建 poi 缓存表 / 司机位置表（表已存在时不会重复建，也不会动已有表）
	if err := db.AutoMigrate(&model.Poi{}, &model.DriverLocation{}); err != nil {
		panic(err)
	}

	return &ServiceContext{
		Config:              c,
		Db:                  db,
		Redis:               redisClient,
		Geo:                 geo.NewClient(c.MapService),
		PoiModel:            model.NewPoiModel(db),
		DriverLocationModel: model.NewDriverLocationModel(db),
		RideTrackPointModel: model.NewRideTrackPointModel(db),
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
