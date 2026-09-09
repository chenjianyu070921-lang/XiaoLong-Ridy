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
	// RPC 服务只允许通过环境变量注入真实凭据，配置文件仅保留非敏感结构。
	// 数据库连接串与 Redis 密码均以环境变量为准，避免明文凭据进入仓库。
	if dsn := strings.TrimSpace(os.Getenv("ADMINSVC_MYSQL_DSN")); dsn != "" {
		c.MySQL.DSN = dsn
	}
	if redisPassword := strings.TrimSpace(os.Getenv("ADMINSVC_REDIS_PASSWORD")); redisPassword != "" {
		c.Cache.Password = redisPassword
	}
	// 统一补齐 MySQL 驱动参数：缺少 parseTime 时 DATETIME 列无法扫描为 time.Time，
	// 会导致优惠券、用户、订单等列表接口在本地启动场景下报扫描错误（500）。
	c.MySQL.DSN = normalizeMySQLDSN(c.MySQL.DSN)
	if strings.TrimSpace(c.MySQL.DSN) == "" {
		// 配置错误属于启动前置条件失败，使用可读日志并以非零状态退出，
		// 避免 panic 堆栈污染日志，也避免在无数据库连接时继续创建服务。
		log.Fatal("mysql dsn is empty: set ADMINSVC_MYSQL_DSN or configure MySQL.DSN")
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

// normalizeMySQLDSN 为 MySQL DSN 补齐驱动连接参数，保证时间与编码语义一致。
// 缺失 charset 时补充 utf8mb4；缺失 parseTime 时补充 parseTime=True；
// 缺失 loc 时补充 loc=Local，与仓库内既有 DSN 示例保持等价。
// 入参为原始 DSN，返回规范化后的 DSN；空串原样返回。
func normalizeMySQLDSN(dsn string) string {
	if strings.TrimSpace(dsn) == "" {
		return dsn
	}
	params := make([]string, 0, 3)
	if !strings.Contains(dsn, "charset=") {
		params = append(params, "charset=utf8mb4")
	}
	if !strings.Contains(dsn, "parseTime=") {
		params = append(params, "parseTime=True")
	}
	if !strings.Contains(dsn, "loc=") {
		params = append(params, "loc=Local")
	}
	if len(params) == 0 {
		return dsn
	}
	separator := "?"
	if strings.Contains(dsn, "?") {
		separator = "&"
	}
	return dsn + separator + strings.Join(params, "&")
}
