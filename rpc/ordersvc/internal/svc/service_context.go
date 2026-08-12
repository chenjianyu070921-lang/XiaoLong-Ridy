package svc

import (
	cfg "XiaoLong-Ridy/common/config"
	"XiaoLong-Ridy/common/datasource"
	"XiaoLong-Ridy/rpc/ordersvc/internal/config"

	"gorm.io/gorm"
)

type ServiceContext struct {
	Config config.Config
	DB     *gorm.DB
}

func NewServiceContext(c config.Config) *ServiceContext {
	//1.mysql连接
	client, err := datasource.NewMysqlClient(cfg.MysqlConf{
		Dsn:         c.Mysql.DSN,
		MaxOpenConn: 0,
		MaxIdleConn: 0,
		MaxLifeTime: 0,
	})
	if err != nil {
		panic(err)
	}
	return &ServiceContext{
		Config: c,
		DB:     client,
	}
}
