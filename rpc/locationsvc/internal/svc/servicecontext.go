package svc

import (
	"fmt"
	"strings"
	"sync"

	"XiaoLong-Ridy/common/constants"
	"XiaoLong-Ridy/rpc/locationsvc/internal/config"
	"XiaoLong-Ridy/rpc/locationsvc/internal/geo"
	"XiaoLong-Ridy/rpc/locationsvc/internal/model"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

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
	// 自动建 poi 缓存表 / 司机位置表（表已存在时不会重复建，也不会动已有表）。
	// 本地库已存在 poi 表与命名索引 idx_poi_city 时，GORM 在某些驱动版本会重复尝试创建命名索引，
	// 触发 MySQL "Duplicate key name" 错误；该错误属幂等场景（表与索引实际已就绪），忽略以允许服务启动，
	// 其他迁移错误仍上抛 panic。
	if err := db.AutoMigrate(&model.Poi{}, &model.DriverLocation{}); err != nil {
		if !strings.Contains(err.Error(), "Duplicate key name") && !strings.Contains(err.Error(), "idx_poi_city") {
			panic(err)
		}
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
