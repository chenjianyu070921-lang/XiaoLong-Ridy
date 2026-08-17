package svc

import (
	"time"

	cfg "XiaoLong-Ridy/common/config"
	"XiaoLong-Ridy/common/datasource"
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
		dispatchEngine = engine.NewGeoDispatchEngine(redisClient, "default")
	} else {
		dispatchEngine = engine.NewMockDispatchEngine()
	}

	return &ServiceContext{
		Config:             c,
		DB:                 client,
		Redis:              redisClient,
		DispatchRepository: repository.NewGormDispatchRepository(client),
		DispatchEngine:     dispatchEngine,
	}
}
