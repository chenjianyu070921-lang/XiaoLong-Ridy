package svc

import (
	cfg "XiaoLong-Ridy/common/config"
	"XiaoLong-Ridy/common/datasource"
	"XiaoLong-Ridy/rpc/driversvc/internal/config"
	"XiaoLong-Ridy/rpc/driversvc/internal/repository"
	"time"

	"gorm.io/gorm"
)

type ServiceContext struct {
	Config                 config.Config
	DB                     *gorm.DB
	DriverRepository       repository.DriverRepository
	DriverVehicleRepository repository.DriverVehicleRepository
}

func NewServiceContext(c config.Config) *ServiceContext {
	client, err := datasource.NewMysqlClient(cfg.MysqlConf{
		Dsn:         c.Mysql.DSN,
		MaxOpenConn: 200,
		MaxIdleConn: 30,
		MaxLifeTime: int(time.Hour),
	})
	if err != nil {
		panic(err)
	}
	return &ServiceContext{
		Config:                  c,
		DB:                      client,
		DriverRepository:        repository.NewGormDriverRepository(client),
		DriverVehicleRepository: repository.NewGormVehicleRepository(client),
	}
}
