// driversvc 服务入口：加载配置、初始化依赖、启动 gRPC 服务。
package main

import (
	"XiaoLong-Ridy/rpc/driversvc/proto"
	"flag"
	"fmt"

	"XiaoLong-Ridy/rpc/driversvc/internal/config"
	"XiaoLong-Ridy/rpc/driversvc/internal/server"
	"XiaoLong-Ridy/rpc/driversvc/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/driversvc.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx := svc.NewServiceContext(c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		proto.RegisterDriversvcServer(grpcServer, server.NewDriversvcServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
