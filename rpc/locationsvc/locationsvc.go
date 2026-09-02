package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"time"

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
	"gorm.io/gorm"
)

var configFile = flag.String("f", "etc/locationsvc.yaml", "the config file")

func main() {
	flag.Parse()

	logx.DisableStat()

	// ========== 配置中心：从 etcd 拉取配置，失败降级本地 yaml ==========
	// 注意：启动前需要先把配置写入 etcd，key 为 locationsvc.yaml
	// docker exec etcd etcdctl put locationsvc.yaml < locationsvc.yaml
	// etcd 不可用时（连接失败/拉取失败）降级加载本地文件，保证服务可启动
	var c config.Config
	var cc configcenter.Configurator[config.Config]

	sub, err := subscriber.NewEtcdSubscriber(subscriber.EtcdConf{
		Hosts: []string{"127.0.0.1:2379"},
		Key:   "locationsvc.yaml",
	})
	if err == nil {
		cc, err = configcenter.NewConfigCenter[config.Config](
			configcenter.Config{Type: "yaml"}, sub)
	}
	if err != nil {
		cc = nil
		logx.Errorf("配置中心(etcd)不可用: %v，降级加载本地文件 %s", err, *configFile)
		conf.MustLoad(*configFile, &c)
	} else if c, err = cc.GetConfig(); err != nil {
		cc = nil
		logx.Errorf("从配置中心(etcd)加载配置失败: %v，降级加载本地文件 %s", err, *configFile)
		conf.MustLoad(*configFile, &c)
	} else {
		fmt.Printf("[配置中心] 配置加载成功: Provider=%s BaseUrl=%s\n", c.MapService.Provider, c.MapService.BaseUrl)
	}

	// 连接 MySQL
	mysqlDB, _ := connectMysqlWithRetry(c)
	fmt.Println("MySQL 连接成功")

	// 连接 Redis
	redisClient := datasource.NewRedisClient(c.RedisConf)
	for {
		if err := redisClient.Ping(context.Background()).Err(); err == nil {
			break
		} else {
			logx.Errorf("Redis 连接失败: %v，5 秒后重试", err)
			time.Sleep(5 * time.Second)
		}
	}
	fmt.Println("Redis 连接成功")

	// 注入数据库、Redis 与地图客户端
	ctx := svc.NewServiceContext(c, mysqlDB, redisClient)

	// ========== 配置热更新：etcd 配置变更后自动生效，无需重启 ==========
	// etcd 不可用（已降级本地 yaml）时不注册热更新监听
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

// connectMysqlWithRetry 创建数据库连接并持续探测，避免依赖服务短暂不可用时 locationsvc 直接退出。
// 连接成功后返回 ORM 客户端及底层 SQL 客户端，供服务上下文使用。
func connectMysqlWithRetry(c config.Config) (*gorm.DB, *sql.DB) {
	for {
		mysqlDB, err := datasource.NewMysqlClient(c.Mysql)
		if err == nil {
			sqlDB, dbErr := mysqlDB.DB()
			if dbErr == nil {
				if pingErr := sqlDB.Ping(); pingErr == nil {
					return mysqlDB, sqlDB
				} else {
					logx.Errorf("MySQL 连接失败: %v，5 秒后重试", pingErr)
				}
			} else {
				logx.Errorf("获取 MySQL 连接池失败: %v，5 秒后重试", dbErr)
			}
		} else {
			logx.Errorf("创建 MySQL 客户端失败: %v，5 秒后重试", err)
		}
		time.Sleep(5 * time.Second)
	}
}

// envOr 返回非空环境变量，否则返回默认值，用于覆盖本地开发配置。
func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
