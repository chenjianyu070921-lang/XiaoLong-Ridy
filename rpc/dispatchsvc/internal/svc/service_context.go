package svc

import (
	"context"
	"strconv"
	"time"

	cfg "XiaoLong-Ridy/common/config"
	"XiaoLong-Ridy/common/constants"
	"XiaoLong-Ridy/common/datasource"
	"XiaoLong-Ridy/common/events"
	"XiaoLong-Ridy/rpc/dispatchsvc/internal/config"
	"XiaoLong-Ridy/rpc/dispatchsvc/internal/engine"
	"XiaoLong-Ridy/rpc/dispatchsvc/internal/repository"
	"XiaoLong-Ridy/rpc/ordersvc/orderclient"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config             config.Config
	DB                 *gorm.DB
	Redis              *redis.Client
	EventBus           events.Bus
	DispatchRepository repository.DispatchRepository
	DispatchEngine     engine.DispatchEngine
	// OrderStatusVerifier 派单前复核订单当前状态的函数（P0-M4-1），
	// nil 表示未配置订单服务（单测/离线场景），此时跳过状态复核。
	OrderStatusVerifier func(ctx context.Context, orderID int64) (int32, error)
}

func NewServiceContext(c config.Config) *ServiceContext {
	client, err := datasource.NewMysqlClient(cfg.MysqlConf{
		Dsn:         c.Mysql.DSN,
		MaxOpenConn: 200,
		MaxIdleConn: 30,
		MaxLifeTime: int((time.Hour * 1).Seconds()),
	})
	if err != nil {
		panic(err)
	}

	redisClient := datasource.NewRedisClient(c.Redis)

	// 注入订单状态复核器：派单前确认订单仍为待接单，防止取消/超时订单被竞态派单（P0-M4-1）。
	var orderStatusVerifier func(ctx context.Context, orderID int64) (int32, error)
	if c.OrderRPC.Target != "" || len(c.OrderRPC.Endpoints) > 0 {
		conn, err := zrpc.NewClient(c.OrderRPC)
		if err != nil {
			panic(err)
		}
		oc := orderclient.NewOrder(conn)
		orderStatusVerifier = func(ctx context.Context, orderID int64) (int32, error) {
			resp, err := oc.GetOrder(ctx, &orderclient.GetOrderRequest{OrderId: orderID})
			if err != nil {
				return 0, err
			}
			if resp == nil || resp.OrderId <= 0 {
				return 0, gorm.ErrRecordNotFound
			}
			return int32(resp.Status), nil
		}
	}

	var dispatchEngine engine.DispatchEngine
	if c.Redis.Host != "" {
		// 注入真实司机评分提供器：从 driver_score 读取服务分与完单率，替换写死权重。
		repo := repository.NewGormDispatchRepository(client)
		scoreProvider := func(ctx context.Context, driverID uint64) (float64, float64) {
			s, err := repo.GetDriverScore(ctx, driverID)
			if err != nil || s == nil {
				// 评分数据缺失时降级为 0（加权失效），告警以便排查 driver_score 是否未初始化（P2-M4-6）。
				logx.Errorf("driver score not found for driverId=%d, fallback to zero weight: err=%v", driverID, err)
				return 0, 0
			}
			rating := s.Score / 20 // 服务分(0~100) 归一化到 rating(0~5)
			if rating > 5 {
				rating = 5
			}
			completion := 1 - s.MonthCancelRate/100 // 取消率(%) 反推完单率(0~1)
			if completion < 0 {
				completion = 0
			}
			return rating, completion
		}
		// 注入司机可用性过滤：只派单给"在线且未忙碌"的司机（P1-M4-8）。
		// 在线集合由 location-consumer 位置上报写入；忙碌集合由订单状态机维护（见 ordersvc Accept/Finish/Cancel）。
		availability := func(ctx context.Context, driverID uint64) (online, busy bool) {
			member := strconv.FormatUint(driverID, 10)
			if v, err := redisClient.SIsMember(ctx, constants.RedisDriverOnline, member).Result(); err == nil {
				online = v
			}
			if v, err := redisClient.SIsMember(ctx, constants.RedisDriverBusy, member).Result(); err == nil {
				busy = v
			}
			return
		}
		// 默认城市键从配置读取，消除硬编码 "default"，与 locationsvc 写入的 GEO key 保持对齐（P1-M4-5）。
		defaultCity := c.DefaultCityCode
		if defaultCity == "" {
			defaultCity = "default"
		}
		dispatchEngine = engine.NewGeoDispatchEngineWithScoreAndAvailability(redisClient, defaultCity, c.EnableMockDispatch, scoreProvider, availability)
	} else {
		dispatchEngine = engine.NewMockDispatchEngine()
	}

	var eventBus events.Bus
	if c.Redis.Host != "" {
		eventBus = events.NewRedisStreamBus(redisClient, constants.OrderEventStream)
	}

	return &ServiceContext{
		Config:              c,
		DB:                  client,
		Redis:               redisClient,
		EventBus:            eventBus,
		DispatchRepository:  repository.NewGormDispatchRepository(client),
		DispatchEngine:      dispatchEngine,
		OrderStatusVerifier: orderStatusVerifier,
	}
}
