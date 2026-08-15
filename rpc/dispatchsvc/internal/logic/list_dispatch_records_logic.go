package logic

import (
	"context"

	"XiaoLong-Ridy/rpc/dispatchsvc/internal/svc"
	"XiaoLong-Ridy/rpc/dispatchsvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListDispatchRecordsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListDispatchRecordsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListDispatchRecordsLogic {
	return &ListDispatchRecordsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 分页查询订单的派单记录。
func (l *ListDispatchRecordsLogic) ListDispatchRecords(in *proto.ListDispatchRecordsRequest) (*proto.ListDispatchRecordsResponse, error) {
	if in.OrderId <= 0 {
		return nil, ErrInvalidOrderParams
	}
	page := in.Page
	if page <= 0 {
		page = 1
	}
	pageSize := in.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	list, total, err := l.svcCtx.DispatchRepository.ListByOrder(l.ctx, uint64(in.OrderId), page, pageSize)
	if err != nil {
		return nil, err
	}
	items := make([]*proto.DispatchRecord, 0, len(list))
	for _, record := range list {
		items = append(items, &proto.DispatchRecord{
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

	return &proto.ListDispatchRecordsResponse{
		List:     items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}
