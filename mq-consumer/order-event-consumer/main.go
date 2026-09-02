package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"XiaoLong-Ridy/mq-consumer/order-event-consumer/internal/config"
	"XiaoLong-Ridy/mq-consumer/order-event-consumer/internal/consumer"
	"XiaoLong-Ridy/mq-consumer/order-event-consumer/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
)

// configFile 默认指向仓库中实际维护的消费者配置文件；仍可通过 -f 覆盖。
var configFile = flag.String("f", "etc/order-event-consumer.yaml", "the config file")

func main() {
	flag.Parse()

	logx.DisableStat()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	svcCtx := svc.NewServiceContext(c)
	orderConsumer := consumer.NewOrderConsumer(svcCtx)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := orderConsumer.Start(ctx); err != nil {
			logx.Errorf("orderclient event consumer failed: %v", err)
		}
	}()

	logx.Info("Starting orderclient-event-consumer...")
	logx.Info("消费订单事件流: orderclient:event:stream")
	<-ctx.Done()
	logx.Info("orderclient-event-consumer stopped")
}
