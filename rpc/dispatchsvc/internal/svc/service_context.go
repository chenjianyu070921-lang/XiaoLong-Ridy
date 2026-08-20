package svc

import (
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
		dispatchEngine = engine.NewGeoDispatchEngineWithMock(redisClient, "default", c.EnableMockDispatch)
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
