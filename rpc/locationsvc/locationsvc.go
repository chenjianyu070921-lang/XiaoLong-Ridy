package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strconv"

	"XiaoLong-Ridy/common/datasource"
	"XiaoLong-Ridy/rpc/locationsvc/internal/config"
	"XiaoLong-Ridy/rpc/locationsvc/internal/server"
	"XiaoLong-Ridy/rpc/locationsvc/internal/svc"
	"XiaoLong-Ridy/rpc/locationsvc/locationsvc"

	"github.com/zeromicro/go-zero/core/conf"
	configcenter "github.com/zeromicro/go-zero/core/configcenter"
	"github.com/zeromicro/go-zero/core/configcenter/subscriber"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/locationsvc.yaml", "the config file")

func main() {
	flag.Parse()

	logx.DisableStat()

	var c config.Config
	// 默认从本地 YAML 加载，避免开发环境未启动 etcd 时阻断服务启动。
	// 生产环境如需配置中心与热更新，可设置 LOCATIONSVC_ETCD_ENABLED=true。
	conf.MustLoad(*configFile, &c)
	var cc configcenter.Configurator[config.Config]
	if enabled, _ := strconv.ParseBool(os.Getenv("LOCATIONSVC_ETCD_ENABLED")); enabled {
		// 配置中心要求 etcd 中预先存在 locationsvc.yaml，内容格式为 YAML。
		etcdSubscriber, err := subscriber.NewEtcdSubscriber(subscriber.EtcdConf{
			Hosts: []string{envOr("LOCATIONSVC_ETCD_ADDR", "127.0.0.1:2379")},
			Key:   "locationsvc.yaml",
		})
		if err != nil {
			logx.Errorf("连接 etcd 配置中心失败: %v，继续使用本地文件 %s", err, *configFile)
		} else if center, err := configcenter.NewConfigCenter[config.Config](configcenter.Config{Type: "yaml"}, etcdSubscriber); err != nil {
			logx.Errorf("从 etcd 配置中心加载配置失败: %v，继续使用本地文件 %s", err, *configFile)
		} else if remoteCfg, err := center.GetConfig(); err != nil {
			logx.Errorf("读取 etcd 配置内容失败: %v，继续使用本地文件 %s", err, *configFile)
		} else {
			c = remoteCfg
			cc = center
			logx.Infof("已启用 etcd 配置中心")
		}
	}
	fmt.Printf("[配置中心] 配置加载成功: Provider=%s BaseUrl=%s\n", c.MapService.Provider, c.MapService.BaseUrl)

	// 连接 MySQL
	mysqlDB, err := datasource.NewMysqlClient(c.Mysql)
	if err != nil {
		panic(err)
	}
	sqlDB, err := mysqlDB.DB()
	if err != nil {
		panic(err)
	}
	if err := sqlDB.Ping(); err != nil {
		panic(err)
	}
	fmt.Println("MySQL 连接成功")

	// 连接 Redis
	redisClient := datasource.NewRedisClient(c.RedisConf)
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		panic(err)
	}
	fmt.Println("Redis 连接成功")

	// 注入数据库、Redis 与地图客户端
	ctx := svc.NewServiceContext(c, mysqlDB, redisClient)

	// ========== 配置热更新：etcd 配置变更后自动生效，无需重启 ==========
	if cc != nil {
		cc.AddListener(func() {
			newCfg, e := cc.GetConfig()
			if e != nil {
				logx.Errorf("[配置中心] 热更新解析失败: %v", e)
				return
			}
			oldKey := ctx.GetConfig().MapService.ApiKey
			ctx.UpdateConfig(newCfg)
			fmt.Printf("[配置中心] 检测到配置变更，已热更新: 地图ApiKey %s -> %s\n",
				oldKey, newCfg.MapService.ApiKey)
		})
	}

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		locationsvc.RegisterLocationServiceServer(grpcServer, server.NewLocationServiceServer(ctx))
		if c.Mode == "dev" || c.Mode == "test" {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting locationsvc rpc server at %s...\n", c.ListenOn)
	s.Start()
}

// envOr 返回非空环境变量，否则返回默认值，用于覆盖本地开发配置。
func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
