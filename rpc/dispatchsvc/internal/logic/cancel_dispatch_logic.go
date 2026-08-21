package logic

import (
	"context"

	"XiaoLong-Ridy/rpc/dispatchsvc/internal/svc"
	"XiaoLong-Ridy/rpc/dispatchsvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

type CancelDispatchLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCancelDispatchLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelDispatchLogic {
	return &CancelDispatchLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 订单取消/超时取消：将该订单全部待派单记录置为已取消（Pending -> Cancelled）。
func (l *CancelDispatchLogic) CancelDispatch(in *proto.CancelDispatchRequest) (*proto.CancelDispatchResponse, error) {
	if in.OrderId <= 0 {
		return nil, ErrInvalidOrderParams
	}
	reason := in.Reason
	if reason == "" {
		reason = "订单取消"
	}

	affected, err := l.svcCtx.DispatchRepository.CancelPendingByOrder(l.ctx, uint64(in.OrderId), reason)
	if err != nil {
		return nil, err
	}
	return &proto.CancelDispatchResponse{
		OrderId:  in.OrderId,
		Affected: affected,
	}, nil
}
