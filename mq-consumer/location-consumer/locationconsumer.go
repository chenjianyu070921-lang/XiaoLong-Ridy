package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"XiaoLong-Ridy/common/constants"
	"XiaoLong-Ridy/common/datasource"
	"XiaoLong-Ridy/common/events"
	"XiaoLong-Ridy/mq-consumer/location-consumer/internal/config"
	"XiaoLong-Ridy/mq-consumer/location-consumer/internal/handler"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
)

var configFile = flag.String("f", "etc/location-consumer.yaml", "the config file")

func main() {
	flag.Parse()

	logx.DisableStat()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	rdb := datasource.NewRedisClient(c.Redis)
	h := handler.NewLocationHandler(rdb)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		bus := events.NewRedisStreamBus(rdb, constants.DriverLocationStream)
		if err := bus.Consume(ctx, "location-consumer-group", h.Consume); err != nil {
			logx.Errorf("consume driver location failed: %v", err)
		}
	}()

	fmt.Println("Starting location-consumer...")
	fmt.Println("消费司机位置流: driver:location:stream")
	<-ctx.Done()
	fmt.Println("location-consumer stopped")
}
