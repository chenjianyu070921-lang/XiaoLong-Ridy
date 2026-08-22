package main

import (
	"XiaoLong-Ridy/rpc/ordersvc/proto"
	"flag"
	"fmt"

	"XiaoLong-Ridy/rpc/ordersvc/internal/config"
	"XiaoLong-Ridy/rpc/ordersvc/internal/server"
	"XiaoLong-Ridy/rpc/ordersvc/internal/svc"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
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

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
