package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"XiaoLong-Ridy/rpc/adminsvc/adminsvc"
	"XiaoLong-Ridy/rpc/adminsvc/internal/config"
	adminserviceServer "XiaoLong-Ridy/rpc/adminsvc/internal/server/adminservice"
	"XiaoLong-Ridy/rpc/adminsvc/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/admin.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	// RPC 服务只允许通过环境变量注入真实数据库凭据，配置文件仅保留非敏感结构。
	if dsn := strings.TrimSpace(os.Getenv("ADMINSVC_MYSQL_DSN")); dsn != "" {
		c.MySQL.DSN = dsn
	}
	if strings.TrimSpace(c.MySQL.DSN) == "" {
		panic("mysql dsn is empty: set ADMINSVC_MYSQL_DSN")
	}
	ctx := svc.NewServiceContext(c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		adminsvc.RegisterAdminServiceServer(grpcServer, adminserviceServer.NewAdminServiceServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	// 所有后台业务 RPC 在服务端统一验证管理员会话和角色权限，防止绕过 HTTP 网关直接调用。
	s.AddUnaryInterceptors(adminserviceServer.NewAuthorizationInterceptor(ctx))
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
