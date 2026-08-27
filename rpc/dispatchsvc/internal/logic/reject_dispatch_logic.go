package logic

import (
	"context"
	"errors"
	"strings"

	"XiaoLong-Ridy/rpc/dispatchsvc/internal/repository"
	"XiaoLong-Ridy/rpc/dispatchsvc/internal/svc"
	"XiaoLong-Ridy/rpc/dispatchsvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

type RejectDispatchLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRejectDispatchLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RejectDispatchLogic {
	return &RejectDispatchLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 司机拒单：将指定司机的待派单记录置为已拒绝。
func (l *RejectDispatchLogic) RejectDispatch(in *proto.RejectDispatchRequest) (*proto.RejectDispatchResponse, error) {
	if in.OrderId <= 0 || in.DriverId <= 0 {
		return nil, ErrInvalidOrderParams
	}
	reason := strings.TrimSpace(in.Reason)
	if reason == "" {
		reason = "司机拒单"
	}

	record, err := l.svcCtx.DispatchRepository.RejectByOrderAndDriver(l.ctx, uint64(in.OrderId), uint64(in.DriverId), reason)
	if errors.Is(err, repository.ErrDispatchRecordNotFound) {
		return nil, err
	}
	if err != nil {
		return nil, err
	}

	return &proto.RejectDispatchResponse{
		OrderId:  int64(record.OrderId),
		DriverId: int64(record.DriverId),
		Status:   int32(record.Status),
	}, nil
}
