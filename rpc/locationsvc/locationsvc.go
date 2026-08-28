package main

import (
	"context"
	"flag"
	"fmt"
	"os"

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

	// ========== 配置加载：本地 yaml 优先，保证任何环境(含无 etcd)都能直接启动 ==========
	// 原来用 configcenter.MustNewConfigCenter 强制连 etcd，导致没有 etcd 配置的机器启动即 panic。
	// 现改为：先加载本地 yaml；etcd 配置中心仅作为「可选热更新」，连不上不影响启动。
	var c config.Config
	conf.MustLoad(*configFile, &c)

	// 高德 ApiKey 允许通过环境变量覆盖，便于不同环境/开发者使用各自的 Key（不写死在仓库）
	if envKey := os.Getenv("AMAP_API_KEY"); envKey != "" {
		c.MapService.ApiKey = envKey
	}
	fmt.Printf("[配置] 本地配置加载成功: Provider=%s BaseUrl=%s\n", c.MapService.Provider, c.MapService.BaseUrl)

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

	// ========== 可选：etcd 配置中心热更新（连不上/无配置则静默跳过，不影响启动） ==========
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logx.Errorf("[配置中心] etcd 热更新初始化失败(已跳过，使用本地配置): %v", r)
			}
		}()
		cc := configcenter.MustNewConfigCenter[config.Config](
			configcenter.Config{Type: "yaml"},
			subscriber.MustNewEtcdSubscriber(subscriber.EtcdConf{
				Hosts: []string{"127.0.0.1:2379"},
				Key:   "locationsvc.yaml",
			}),
		)
		if remote, e := cc.GetConfig(); e == nil {
			ctx.UpdateConfig(remote)
			fmt.Printf("[配置中心] 已加载远程配置，地图ApiKey=%s\n", remote.MapService.ApiKey)
		} else {
			logx.Infof("[配置中心] 远程配置不可用，使用本地配置: %v", e)
		}
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
	}()

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
