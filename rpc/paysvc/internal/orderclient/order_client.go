// Package orderclient 封装对订单服务（ordersvc）的 RPC 调用。
package orderclient

import (
	"context"

	"XiaoLong-Ridy/rpc/ordersvc/proto"

	"github.com/zeromicro/go-zero/zrpc"
)

// OrderClient 订单服务客户端抽象，便于业务层解耦与测试替换。
type OrderClient interface {
	// GetDriverId 查询订单的司机ID。
	GetDriverId(ctx context.Context, orderId int64) (int64, error)
	// ConfirmPaid 通知订单服务支付成功并完成订单。
	ConfirmPaid(ctx context.Context, in *proto.ConfirmPaidRequest) (*proto.ConfirmPaidResponse, error)
}

// RpcOrderClient 基于 ordersvc gRPC 的实现。
type RpcOrderClient struct {
	client proto.OrderClient
}

// NewRpcOrderClient 创建订单 RPC 客户端。
func NewRpcOrderClient(c zrpc.Client) *RpcOrderClient {
	return &RpcOrderClient{client: proto.NewOrderClient(c.Conn())}
}

// GetDriverId 调用 GetOrder 并返回司机ID。
func (r *RpcOrderClient) GetDriverId(ctx context.Context, orderId int64) (int64, error) {
	resp, err := r.client.GetOrder(ctx, &proto.GetOrderRequest{OrderId: orderId})
	if err != nil {
		return 0, err
	}
	return resp.DriverId, nil
}

// ConfirmPaid 调用 ordersvc 的支付成功确认接口。
func (r *RpcOrderClient) ConfirmPaid(ctx context.Context, in *proto.ConfirmPaidRequest) (*proto.ConfirmPaidResponse, error) {
	return r.client.ConfirmPaid(ctx, in)
}
