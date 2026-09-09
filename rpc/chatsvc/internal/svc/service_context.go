package svc

import (
	"strings"

	"XiaoLong-Ridy/rpc/chatsvc/internal/config"
	"XiaoLong-Ridy/rpc/chatsvc/internal/model"
	orderproto "XiaoLong-Ridy/rpc/ordersvc/proto"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/zrpc"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// ServiceContext 承载 chatsvc 的依赖：gorm 库、司机推送 Redis、ordersvc 客户端、敏感词集合。
type ServiceContext struct {
	Config       config.Config
	DB           *gorm.DB
	RedisClient  *redis.Client
	OrderClient  orderproto.OrderClient
	SensitiveSet map[string]struct{}
}

func NewServiceContext(c config.Config) *ServiceContext {
	db, err := gorm.Open(mysql.Open(c.Mysql.DSN), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     c.DriverRedis.Host,
		Password: c.DriverRedis.Password,
		DB:       c.DriverRedis.DB,
	})

	orderClient := orderproto.NewOrderClient(zrpc.MustNewClient(c.OrderRpc).Conn())

	set := make(map[string]struct{}, len(c.SensitiveWords))
	for _, w := range c.SensitiveWords {
		if w = strings.TrimSpace(w); w != "" {
			set[w] = struct{}{}
		}
	}

	return &ServiceContext{
		Config:       c,
		DB:           db,
		RedisClient:  rdb,
		OrderClient:  orderClient,
		SensitiveSet: set,
	}
}

// ContainsSensitive 判断消息是否命中敏感词（线下交易/联系方式等）。
func (s *ServiceContext) ContainsSensitive(content string) bool {
	for w := range s.SensitiveSet {
		if w != "" && strings.Contains(content, w) {
			return true
		}
	}
	return false
}

var _ = model.IMConversation{}
