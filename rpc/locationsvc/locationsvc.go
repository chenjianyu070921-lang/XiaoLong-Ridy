package main

import (
	"context"
	"flag"
	"fmt"

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

	// ========== 配置中心：优先从 etcd 拉取配置，etcd 不可用或读取失败时降级本地 yaml ==========
	// 注意：etcd 可用时需先把配置写入 etcd，key 为 locationsvc.yaml
	// docker exec etcd etcdctl put locationsvc.yaml < locationsvc.yaml
	// 本地无 etcd 时（本地联调），subscriber.NewEtcdSubscriber 会返回错误，
	// 直接回退 conf.MustLoad 本地文件，避免 MustNewEtcdSubscriber 直接 panic 中断启动。
	var c config.Config
	var cc configcenter.Configurator[config.Config]
	sub, err := subscriber.NewEtcdSubscriber(subscriber.EtcdConf{
		Hosts: []string{"127.0.0.1:2379"},
		Key:   "locationsvc.yaml",
	})
	if err != nil {
		logx.Errorf("etcd 配置中心不可用(%v)，降级加载本地文件 %s", err, *configFile)
		conf.MustLoad(*configFile, &c)
	} else {
		client, e := configcenter.NewConfigCenter[config.Config](configcenter.Config{Type: "yaml"}, sub)
		if e != nil {
			logx.Errorf("从配置中心加载配置失败(%v)，降级加载本地文件 %s", e, *configFile)
			conf.MustLoad(*configFile, &c)
		} else {
			cc = client
			if c, e = cc.GetConfig(); e != nil {
				logx.Errorf("从配置中心读取配置失败(%v)，降级加载本地文件 %s", e, *configFile)
				conf.MustLoad(*configFile, &c)
			}
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

	// ========== 配置热更新：仅配置中心可用时监听；本地降级模式无热更新 ==========
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
