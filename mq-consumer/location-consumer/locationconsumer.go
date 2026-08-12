package main

import (
	"flag"
	"fmt"

	"XiaoLong-Ridy/mq-consumer/location-consumer/internal/config"
	"XiaoLong-Ridy/mq-consumer/location-consumer/internal/handler"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
)

var configFile = flag.String("f", "etc/location-consumer.yaml", "the config file")

func main() {
	flag.Parse()

	logx.DisableStat()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	serviceGroup := service.NewServiceGroup()
	defer serviceGroup.Stop()

	h := handler.NewLocationHandler()
	_ = h

	fmt.Println("Starting location-consumer...")
	fmt.Println("TODO: 对接 Kafka 消费司机位置上报消息")
	serviceGroup.Start()
}
