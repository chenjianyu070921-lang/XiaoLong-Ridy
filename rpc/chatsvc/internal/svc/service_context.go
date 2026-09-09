package svc

import (
	"time"

	cfg "XiaoLong-Ridy/common/config"
	"XiaoLong-Ridy/common/datasource"
	"XiaoLong-Ridy/rpc/chatsvc/internal/config"
	"XiaoLong-Ridy/rpc/chatsvc/internal/model"
	orderproto "XiaoLong-Ridy/rpc/ordersvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gorm.io/gorm"
)

// ServiceContext 持有 chatsvc 运行时依赖。
type ServiceContext struct {
	Config      config.Config
	DB          *gorm.DB
	OrderClient orderproto.OrderClient
}

// NewServiceContext 初始化 MySQL、自动建表并连接 ordersvc。
func NewServiceContext(c config.Config) *ServiceContext {
	client, err := datasource.NewMysqlClient(cfg.MysqlConf{
		Dsn:         c.Mysql.Dsn,
		MaxOpenConn: 50,
		MaxIdleConn: 10,
		MaxLifeTime: int((time.Minute * 30).Seconds()),
	})
	if err != nil {
		panic(err)
	}
	// 自动建表：开发环境免去手动执行 scripts/chat.sql；生产以 SQL 脚本为准。
	if err := client.AutoMigrate(&model.Conversation{}, &model.Message{}); err != nil {
		logx.Errorf("chatsvc automigrate failed: %v", err)
	}

	orderAddr := c.OrderRPCAddr
	if orderAddr == "" {
		orderAddr = "127.0.0.1:50051"
	}
	conn, err := grpc.NewClient(orderAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logx.Errorf("chatsvc connect ordersvc failed: %v", err)
	}

	svc := &ServiceContext{Config: c, DB: client}
	if conn != nil {
		svc.OrderClient = orderproto.NewOrderClient(conn)
	}
	return svc
}
