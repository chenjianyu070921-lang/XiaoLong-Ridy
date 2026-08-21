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
		dispatchEngine = engine.NewGeoDispatchEngineWithScore(redisClient, "default", c.EnableMockDispatch, scoreProvider)
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
