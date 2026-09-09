package main

import (
	"context"
	"flag"
	"time"

	"XiaoLong-Ridy/rpc/ordersvc/internal/config"
	"XiaoLong-Ridy/rpc/ordersvc/internal/logic"
	"XiaoLong-Ridy/rpc/ordersvc/internal/server"
	"XiaoLong-Ridy/rpc/ordersvc/internal/svc"
	"XiaoLong-Ridy/rpc/ordersvc/proto"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/ordersvc.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx := svc.NewServiceContext(c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		proto.RegisterOrderServer(grpcServer, server.NewOrderServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	logx.Infof("Starting ordersvc rpc server at %s...", c.ListenOn)

	// 司机接单后迟迟未开始行程：定时回收，取消订单并释放司机运力，避免订单悬挂。
	// 缺陷#1：已接单状态超时回收（仅取消，重派属改派域不在此处理）。
	go func() {
		logic.NewTimeoutAcceptLogic(context.Background(), ctx).Run()
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			logic.NewTimeoutAcceptLogic(context.Background(), ctx).Run()
		}
	}()

	s.Start()
}
