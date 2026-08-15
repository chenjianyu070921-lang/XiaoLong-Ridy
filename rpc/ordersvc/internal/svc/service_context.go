package svc

import (
	cfg "XiaoLong-Ridy/common/config"
	"XiaoLong-Ridy/common/datasource"
	dispatch "XiaoLong-Ridy/rpc/dispatchsvc/dispatch"
	"XiaoLong-Ridy/rpc/ordersvc/internal/config"
	"XiaoLong-Ridy/rpc/ordersvc/internal/repository"
	"time"

	"github.com/zeromicro/go-zero/zrpc"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config          config.Config
	DB              *gorm.DB
	OrderRepository repository.OrderRepository
	DispatchClient  dispatch.Dispatch
}

// NewServiceContext 初始化 MySQL 连接与订单仓储。
func NewServiceContext(c config.Config) *ServiceContext {
	//1.mysql连接
	client, err := datasource.NewMysqlClient(cfg.MysqlConf{
		Dsn:         c.Mysql.DSN,
		MaxOpenConn: 200,
		MaxIdleConn: 30,
		MaxLifeTime: int((time.Hour * 1).Seconds()),
	})
	if err != nil {
		panic(err)
	}

	dispatchRPC := c.DispatchRPC
	if dispatchRPC.Target == "" && len(dispatchRPC.Endpoints) == 0 {
		dispatchRPC.Target = "127.0.0.1:8083"
	}
	dispatchClient, err := zrpc.NewClient(dispatchRPC)
	if err != nil {
		panic(err)
	}

	return &ServiceContext{
		Config:          c,
		DB:              client,
		OrderRepository: repository.NewGormOrderRepository(client),
		DispatchClient:  dispatch.NewDispatch(dispatchClient),
	}
}
