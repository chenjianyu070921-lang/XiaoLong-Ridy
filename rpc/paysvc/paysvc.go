package main

import (
	"flag"
	"fmt"
	"os"

	"XiaoLong-Ridy/rpc/paysvc/internal/config"
	"XiaoLong-Ridy/rpc/paysvc/internal/server"
	"XiaoLong-Ridy/rpc/paysvc/internal/svc"
	"XiaoLong-Ridy/rpc/paysvc/proto"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/paysvc.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	// 启动期所有依赖校验在这里收口：失败返回 error，让进程退出码非 0 并打印日志（M5-7 优雅退出）。
	ctx, err := svc.NewServiceContext(c)
	if err != nil {
		fmt.Fprintf(os.Stderr, "paysvc bootstrap failed: %v\n", err)
		os.Exit(1)
	}

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		proto.RegisterPayServer(grpcServer, server.NewPayServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	// defer 是 LIFO：先注册的后执行。下方顺序保证：
	//   1. server.Stop 先调用（停止接收新请求，等待 in-flight 处理完成）
	//   2. ctx.Close 后调用（释放 DB 连接池 / Kafka Producer）
	// 严禁把两条 defer 调换位置，否则可能在飞行中的请求仍持有连接时关闭 DB。
	defer func() {
		if cerr := ctx.Close(); cerr != nil {
			fmt.Fprintf(os.Stderr, "paysvc resource close: %v\n", cerr)
		}
	}()
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
