package svc

import (
	"XiaoLong-Ridy/job/internal/config"
	order "XiaoLong-Ridy/rpc/ordersvc/order"

	"github.com/zeromicro/go-zero/zrpc"
)

// ServiceContext job 任务上下文，持有下游 RPC 客户端。
type ServiceContext struct {
	Config      config.Config
	OrderClient order.Order
}

// NewServiceContext 初始化下游 RPC 客户端；空配置时兜底本地默认地址。
func NewServiceContext(c config.Config) *ServiceContext {
	orderRPC := c.OrderRPC
	if orderRPC.Target == "" && len(orderRPC.Endpoints) == 0 {
		orderRPC.Target = "127.0.0.1:50051"
	}
	orderClient, err := zrpc.NewClient(orderRPC)
	if err != nil {
		panic(err)
	}
	return &ServiceContext{
		Config:      c,
		OrderClient: order.NewOrder(orderClient),
	}
}
