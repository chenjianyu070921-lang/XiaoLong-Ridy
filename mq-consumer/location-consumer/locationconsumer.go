package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"time"

	commonconfig "XiaoLong-Ridy/common/config"
	"XiaoLong-Ridy/common/constants"
	"XiaoLong-Ridy/mq-consumer/location-consumer/internal/config"
	"XiaoLong-Ridy/mq-consumer/location-consumer/internal/handler"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
)

const (
	consumerGroup = "location-consumer-group"
	consumerName  = "location-consumer-1"
)

var configFile = flag.String("f", "etc/location-consumer.yaml", "the config file")

func main() {
	flag.Parse()
	logx.DisableStat()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	rdb := newRedisClient(c.RedisConf)

	ctx := context.Background()
	// 创建消费组（流不存在则一并创建），已存在时忽略 BUSYGROUP
	if err := rdb.XGroupCreateMkStream(ctx, constants.LocationStreamKey, consumerGroup, "0").Err(); err != nil {
		logx.Infof("XGroupCreateMkStream: %v（可能已存在，可忽略）", err)
	}

	serviceGroup := service.NewServiceGroup()
	defer serviceGroup.Stop()

	h := handler.NewLocationHandler(rdb)
	go runConsumer(ctx, rdb, h)

	fmt.Println("Starting location-consumer, consuming stream:", constants.LocationStreamKey)
	serviceGroup.Start()
}

// newRedisClient 根据配置构造 redis 客户端，超时缺失时使用默认值
func newRedisClient(cfg commonconfig.RedisConf) *redis.Client {
	dialTimeout := time.Duration(cfg.DialTimeout) * time.Second
	if dialTimeout <= 0 {
		dialTimeout = 5 * time.Second
	}
	readTimeout := time.Duration(cfg.ReadTimeout) * time.Second
	if readTimeout <= 0 {
		readTimeout = 3 * time.Second
	}
	writeTimeout := time.Duration(cfg.WriteTimeout) * time.Second
	if writeTimeout <= 0 {
		writeTimeout = 3 * time.Second
	}
	return redis.NewClient(&redis.Options{
		Addr:         cfg.Host,
		Password:     cfg.Pass,
		DB:           cfg.Db,
		PoolSize:     cfg.PoolSize,
		DialTimeout:  dialTimeout,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	})
}

// runConsumer 循环消费位置事件流，维护在线司机集合
func runConsumer(ctx context.Context, rdb *redis.Client, h *handler.LocationHandler) {
	for {
		res := rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    consumerGroup,
			Consumer: consumerName,
			Streams:  []string{constants.LocationStreamKey, ">"},
			Count:    10,
			Block:    2 * time.Second,
		})
		if err := res.Err(); err != nil {
			if errors.Is(err, redis.Nil) {
				continue
			}
			logx.Errorf("XReadGroup failed: %v", err)
			time.Sleep(time.Second)
			continue
		}
		for _, stream := range res.Val() {
			for _, msg := range stream.Messages {
				if err := h.Consume(ctx, msg.ID, msg.Values); err != nil {
					logx.Errorf("consume msg %s failed: %v", msg.ID, err)
					continue
				}
				rdb.XAck(ctx, constants.LocationStreamKey, consumerGroup, msg.ID)
			}
		}
	}
}
