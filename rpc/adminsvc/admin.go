package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"XiaoLong-Ridy/rpc/adminsvc/adminsvc"
	"XiaoLong-Ridy/rpc/adminsvc/internal/config"
	adminserviceServer "XiaoLong-Ridy/rpc/adminsvc/internal/server/adminservice"
	"XiaoLong-Ridy/rpc/adminsvc/internal/svc"

	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gopkg.in/yaml.v3"
)

var configFile = flag.String("f", "etc/admin.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	log.Printf("adminsvc loading config: %s", *configFile)
	// 使用 yaml.v3 读取配置，避免 go-zero 配置反射器在当前 Go 版本下解析嵌入 RpcServerConf 时异常占用内存。
	// 配置字段及默认值仍由既有 Config 结构体和服务初始化逻辑统一约束。
	configData, err := os.ReadFile(*configFile)
	if err != nil {
		log.Fatalf("read config: %v", err)
	}
	if err := yaml.Unmarshal(configData, &c); err != nil {
		log.Fatalf("parse config: %v", err)
	}
	// RpcServerConf 来自第三方嵌入结构，yaml.v3 不会自动展开其字段；单独读取基础监听配置后回填。
	var rpcBase struct {
		Name     string `yaml:"Name"`
		ListenOn string `yaml:"ListenOn"`
	}
	if err := yaml.Unmarshal(configData, &rpcBase); err != nil {
		log.Fatalf("parse rpc base config: %v", err)
	}
	c.RpcServerConf.Name = rpcBase.Name
	c.RpcServerConf.ListenOn = rpcBase.ListenOn
	if c.RpcServerConf.ListenOn == "" {
		log.Fatal("rpc ListenOn is empty")
	}
	log.Printf("adminsvc config loaded")
	// RPC 服务只允许通过环境变量注入真实数据库凭据，配置文件仅保留非敏感结构。
	if dsn := strings.TrimSpace(os.Getenv("ADMINSVC_MYSQL_DSN")); dsn != "" {
		c.MySQL.DSN = dsn
	}
	if strings.TrimSpace(c.MySQL.DSN) == "" {
		panic("mysql dsn is empty: set ADMINSVC_MYSQL_DSN")
	}
	log.Printf("adminsvc creating service context")
	ctx := svc.NewServiceContext(c)
	log.Printf("adminsvc service context created")

	log.Printf("adminsvc creating grpc server")
	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		adminsvc.RegisterAdminServiceServer(grpcServer, adminserviceServer.NewAdminServiceServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	log.Printf("adminsvc grpc server created")
	// 所有后台业务 RPC 在服务端统一验证管理员会话和角色权限，防止绕过 HTTP 网关直接调用。
	s.AddUnaryInterceptors(adminserviceServer.NewAuthorizationInterceptor(ctx))
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
