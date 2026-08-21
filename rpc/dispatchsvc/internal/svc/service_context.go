package svc

import (
	"context"
	"time"

	cfg "XiaoLong-Ridy/common/config"
	"XiaoLong-Ridy/common/constants"
	"XiaoLong-Ridy/common/datasource"
	"XiaoLong-Ridy/common/events"
	"XiaoLong-Ridy/rpc/dispatchsvc/internal/config"
	"XiaoLong-Ridy/rpc/dispatchsvc/internal/engine"
	"XiaoLong-Ridy/rpc/dispatchsvc/internal/repository"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config             config.Config
	DB                 *gorm.DB
	Redis              *redis.Client
	EventBus           events.Bus
	DispatchRepository repository.DispatchRepository
	DispatchEngine     engine.DispatchEngine
}

func NewServiceContext(c config.Config) *ServiceContext {
	client, err := datasource.NewMysqlClient(cfg.MysqlConf{
		Dsn:         c.Mysql.DSN,
		MaxOpenConn: 200,
		MaxIdleConn: 30,
		MaxLifeTime: int((time.Hour * 1).Seconds()),
	})
	if err != nil {
		panic(err)
	}

	redisClient := datasource.NewRedisClient(c.Redis)

	var dispatchEngine engine.DispatchEngine
	if c.Redis.Host != "" {
		// 注入真实司机评分提供器：从 driver_score 读取服务分与完单率，替换写死权重。
		repo := repository.NewGormDispatchRepository(client)
		scoreProvider := func(ctx context.Context, driverID uint64) (float64, float64) {
			s, err := repo.GetDriverScore(ctx, driverID)
			if err != nil || s == nil {
				// 评分数据缺失时降级为 0（加权失效），告警以便排查 driver_score 是否未初始化（P2-M4-6）。
				logx.Errorf("driver score not found for driverId=%d, fallback to zero weight: err=%v", driverID, err)
				return 0, 0
			}
			rating := s.Score / 20 // 服务分(0~100) 归一化到 rating(0~5)
			if rating > 5 {
				rating = 5
			}
			completion := 1 - s.MonthCancelRate/100 // 取消率(%) 反推完单率(0~1)
			if completion < 0 {
				completion = 0
			}
			return rating, completion
		}
		// 默认城市键从配置读取，消除硬编码 "default"，与 locationsvc 写入的 GEO key 保持对齐（P1-M4-5）。
		defaultCity := c.DefaultCityCode
		if defaultCity == "" {
			defaultCity = "default"
		}
		dispatchEngine = engine.NewGeoDispatchEngineWithScore(redisClient, defaultCity, c.EnableMockDispatch, scoreProvider)
	} else {
		dispatchEngine = engine.NewMockDispatchEngine()
	}

	var eventBus events.Bus
	if c.Redis.Host != "" {
		eventBus = events.NewRedisStreamBus(redisClient, constants.OrderEventStream)
	}

	return &ServiceContext{
		Config:             c,
		DB:                 client,
		Redis:              redisClient,
		EventBus:           eventBus,
		DispatchRepository: repository.NewGormDispatchRepository(client),
		DispatchEngine:     dispatchEngine,
	}
}
