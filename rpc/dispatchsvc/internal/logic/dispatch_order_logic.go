package logic

import (
	"context"

	"XiaoLong-Ridy/common/constants"
	"XiaoLong-Ridy/rpc/dispatchsvc/internal/model"
	"XiaoLong-Ridy/rpc/dispatchsvc/internal/svc"
	"XiaoLong-Ridy/rpc/dispatchsvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

type DispatchOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDispatchOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DispatchOrderLogic {
	return &DispatchOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 对指定订单执行派单（P0 为 mock 直派），返回写入的派单记录。
func (l *DispatchOrderLogic) DispatchOrder(in *proto.DispatchOrderRequest) (*proto.DispatchOrderResponse, error) {
	if in.OrderId <= 0 {
		return nil, ErrInvalidOrderParams
	}

	candidates, err := l.svcCtx.DispatchEngine.FindCandidates(
		l.ctx,
		uint64(in.OrderId),
		in.FromLongitude,
		in.FromLatitude,
		in.CarType,
		in.CityCode,
	)
	if err != nil {
		return nil, err
	}

	list := make([]*proto.DispatchRecord, 0, len(candidates))
	for _, candidate := range candidates {
		record := &model.DispatchRecord{
			OrderId:      uint64(in.OrderId),
			DriverId:     candidate.DriverID,
			DispatchType: constants.DispatchTypeAuto,
			Status:       constants.DispatchStatusPending,
			MatchScore:   candidate.MatchScore,
			Remark:       "mock直派",
		}
		if err := l.svcCtx.DispatchRepository.Create(l.ctx, record); err != nil {
			return nil, err
		}
		list = append(list, &proto.DispatchRecord{
			Id:           int64(record.Id),
			OrderId:      int64(record.OrderId),
			DriverId:     int64(record.DriverId),
			DispatchType: int32(record.DispatchType),
			Status:       int32(record.Status),
			MatchScore:   record.MatchScore,
			Remark:       record.Remark,
			CreatedAt:    record.CreatedAt.Unix(),
			UpdatedAt:    record.UpdatedAt.Unix(),
		})
	}

	return &proto.DispatchOrderResponse{
		OrderId: in.OrderId,
		List:    list,
	}, nil
}
