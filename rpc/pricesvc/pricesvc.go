package main

import (
	"flag"
	"fmt"
	"os"

	"XiaoLong-Ridy/rpc/pricesvc/internal/config"
	"XiaoLong-Ridy/rpc/pricesvc/internal/server"
	"XiaoLong-Ridy/rpc/pricesvc/internal/svc"
	"XiaoLong-Ridy/rpc/pricesvc/proto"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/pricesvc.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	// 启动期所有依赖校验在这里收口：失败返回 error，让进程退出码非 0 并打印日志（M5-7 优雅退出）。
	ctx, err := svc.NewServiceContext(c)
	if err != nil {
		fmt.Fprintf(os.Stderr, "pricesvc bootstrap failed: %v\n", err)
		os.Exit(1)
	}

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		proto.RegisterPriceServer(grpcServer, server.NewPriceServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	// defer 是 LIFO：先注册的后执行。下方顺序保证 Stop 在 Close 之前调用。
	defer func() {
		if cerr := ctx.Close(); cerr != nil {
			fmt.Fprintf(os.Stderr, "pricesvc resource close: %v\n", cerr)
		}
	}()
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
