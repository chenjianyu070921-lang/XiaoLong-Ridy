package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"XiaoLong-Ridy/job/internal/config"
	"XiaoLong-Ridy/job/internal/handler"
	"XiaoLong-Ridy/job/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
)

var configFile = flag.String("f", "etc/job.yaml", "the config file")

func main() {
	flag.Parse()

	logx.DisableStat()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	svcCtx := svc.NewServiceContext(c)
	h := handler.NewCleanupHandler(svcCtx)

	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			if err := h.CleanExpiredLocation(); err != nil {
				logx.Errorf("CleanExpiredLocation failed: %v", err)
			}
		}
	}()

	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			if err := h.SyncOrderStatus(); err != nil {
				logx.Errorf("SyncOrderStatus failed: %v", err)
			}
		}
	}()

	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			if err := h.TimeoutCancelOrders(); err != nil {
				logx.Errorf("TimeoutCancelOrders failed: %v", err)
			}
		}
	}()

	go func() {
		for {
			now := time.Now()
			next := time.Date(now.Year(), now.Month(), now.Day()+1, 1, 0, 0, 0, now.Location())
			time.Sleep(next.Sub(now))
			if err := h.DailyReport(); err != nil {
				logx.Errorf("DailyReport failed: %v", err)
			}
		}
	}()

	fmt.Println("Starting job scheduler...")
	fmt.Println("定时任务:")
	fmt.Println("  - 每小时: 清理过期位置数据")
	fmt.Println("  - 每10分钟: 同步异常订单状态")
	fmt.Println("  - 每1分钟: 超时未接单订单自动取消")
	fmt.Println("  - 每日凌晨1点: 生成统计报表")

	select {}
}

var _ = context.Background
