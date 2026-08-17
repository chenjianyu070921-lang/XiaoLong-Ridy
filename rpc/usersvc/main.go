package main

import (
	"flag"
	"fmt"
	"log"

	"XiaoLong-Ridy/rpc/usersvc/internal/config"
	"XiaoLong-Ridy/rpc/usersvc/internal/logic"
	"XiaoLong-Ridy/rpc/usersvc/internal/repository"
	"XiaoLong-Ridy/rpc/usersvc/internal/server"
	"XiaoLong-Ridy/rpc/usersvc/internal/svc"
	userproto "XiaoLong-Ridy/rpc/usersvc/proto"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/usersvc.yaml", "the config file")

// main 是 usersvc 的 go-zero RPC 启动入口，负责加载配置、组装依赖并注册 gRPC 服务。
func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx := newServiceContext(c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		userproto.RegisterUserServer(grpcServer, server.NewUserServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}

// newServiceContext 创建 usersvc 运行时依赖；当前阶段使用内存用户仓储和本地短信验证码服务。
func newServiceContext(c config.Config) *svc.ServiceContext {
	signingKey := c.TokenAuth.SigningKey
	if signingKey == "" {
		signingKey = "local-development-signing-key"
	}

	users := repository.NewMemoryUserRepository()
	smsService := logic.NewMemorySMSCodeService(func(phone, code string) {
		log.Printf("本地短信验证码：phone=%s code=%s", phone, code)
	})
	tokens := logic.NewTokenManager(signingKey)
	return svc.NewServiceContext(c, users, smsService, smsService, tokens)
}
