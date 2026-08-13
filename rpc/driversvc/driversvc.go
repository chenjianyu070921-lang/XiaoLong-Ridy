// driversvc 服务入口：负责加载配置、初始化依赖、启动 gRPC 服务。
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

// configFile 指定配置文件路径，默认读取 etc/driversvc.yaml。
var configFile = flag.String("f", "etc/driversvc.yaml", "the config file")

func main() {
	flag.Parse()

	// 加载 YAML 配置到 Config 结构体
	var c config.Config
	conf.MustLoad(*configFile, &c)
	// 构建服务上下文，期间会建立 MySQL 连接
	ctx := svc.NewServiceContext(c)

	// 启动 gRPC 服务，并注册 DriversvcServer 实现
	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		proto.RegisterDriversvcServer(grpcServer, server.NewDriversvcServer(ctx))

		// 开发/测试模式下开启反射，方便 grpcurl 等工具调试
		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
