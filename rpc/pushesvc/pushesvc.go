package main

import (
	"context"
	"flag"
	"fmt"

	"XiaoLong-Ridy/common/datasource"
	"XiaoLong-Ridy/rpc/pushesvc/internal/config"
	"XiaoLong-Ridy/rpc/pushesvc/internal/model"
	"XiaoLong-Ridy/rpc/pushesvc/internal/server"
	"XiaoLong-Ridy/rpc/pushesvc/internal/svc"
	"XiaoLong-Ridy/rpc/pushesvc/pushesvc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/pushesvc.yaml", "the config file")

func main() {
	flag.Parse()

	logx.DisableStat()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	// 连接 MySQL
	mysqlDB, err := datasource.NewMysqlClient(c.Mysql)
	if err != nil {
		panic(err)
	}
	sqlDB, err := mysqlDB.DB()
	if err != nil {
		panic(err)
	}
	if err := sqlDB.Ping(); err != nil {
		panic(err)
	}
	// 自动迁移：确保 notices / push_log 表结构与 model 定义一致（含 biz_type、extras 等新增列）
	if err := mysqlDB.AutoMigrate(&model.Notice{}, &model.PushLog{}); err != nil {
		panic(fmt.Errorf("自动迁移表结构失败: %w", err))
	}
	fmt.Println("MySQL 连接成功")

	// 注入数据库
	ctx := svc.NewServiceContext(c, mysqlDB)

	// 连接 Redis
	if err := datasource.NewRedisClient(c.RedisConf).Ping(context.Background()).Err(); err != nil {
		panic(err)
	}
	fmt.Println("Redis 连接成功")

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		pushesvc.RegisterPushServiceServer(grpcServer, server.NewPushServiceServer(ctx))
		if c.Mode == "dev" || c.Mode == "test" {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting pushesvc rpc server at %s...\n", c.ListenOn)
	s.Start()
}
