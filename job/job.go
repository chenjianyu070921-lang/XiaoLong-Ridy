package main

import (
	"context"
	"flag"
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
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			if err := h.RetryRefundEvents(); err != nil {
				logx.Errorf("RetryRefundEvents failed: %v", err)
			}
		}
	}()

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
		// 服务启动时先执行一次超时取消扫描，避免任务启动后必须等待完整一分钟。
		if err := h.TimeoutCancelOrders(); err != nil {
			logx.Errorf("initial TimeoutCancelOrders failed: %v", err)
		}

		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			if err := h.TimeoutCancelOrders(); err != nil {
				logx.Errorf("TimeoutCancelOrders failed: %v", err)
			}
		}
	}()

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			if err := h.RescheduleExpiredDispatches(); err != nil {
				logx.Errorf("RescheduleExpiredDispatches failed: %v", err)
			}
		}
	}()

	go func() {
		// 派单失败补偿队列消费：10s 粒度可覆盖 5s/15s/45s 的退避窗口（P1-M4-2）。
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			if err := h.RetryPendingDispatches(); err != nil {
				logx.Errorf("RetryPendingDispatches failed: %v", err)
			}
		}
	}()

	go func() {
		// 管理后台 outbox 补偿：覆盖审计重写、司机冻结重放和 pushsvc 通知重试。
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			if err := h.RetryAdminAuditOutbox(); err != nil {
				logx.Errorf("RetryAdminAuditOutbox failed: %v", err)
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

	logx.Info("Starting job scheduler...")
	logx.Info("定时任务:")
	logx.Info("  - 每小时: 清理过期位置数据")
	logx.Info("  - 每10秒: 派单失败补偿重试")
	logx.Info("  - 每30秒: 管理后台 outbox 补偿重试")
	logx.Info("  - 每1分钟: 超时未接单订单自动取消")
	logx.Info("  - 每30秒: 派单超时重派")
	logx.Info("  - 每日凌晨1点: 生成统计报表")

	select {}
}

var _ = context.Background
