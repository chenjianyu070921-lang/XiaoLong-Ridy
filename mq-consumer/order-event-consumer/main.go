package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"XiaoLong-Ridy/mq-consumer/order-event-consumer/internal/config"
	"XiaoLong-Ridy/mq-consumer/order-event-consumer/internal/consumer"
	"XiaoLong-Ridy/mq-consumer/order-event-consumer/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
)

var configFile = flag.String("f", "etc/orderclient-event-consumer.yaml", "the config file")

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

	fmt.Println("Starting orderclient-event-consumer...")
	fmt.Println("消费订单事件流: orderclient:event:stream")
	<-ctx.Done()
	fmt.Println("orderclient-event-consumer stopped")
}
