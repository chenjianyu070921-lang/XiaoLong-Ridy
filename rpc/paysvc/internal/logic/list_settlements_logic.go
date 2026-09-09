package logic

import (
	"context"
	"time"

	"XiaoLong-Ridy/common/priceutil"
	"XiaoLong-Ridy/rpc/paysvc/internal/model"
	"XiaoLong-Ridy/rpc/paysvc/internal/repository"
	"XiaoLong-Ridy/rpc/paysvc/internal/svc"
	"XiaoLong-Ridy/rpc/paysvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListSettlementsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListSettlementsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListSettlementsLogic {
	return &ListSettlementsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListSettlementsLogic) ListSettlements(in *proto.ListSettlementsRequest) (*proto.ListSettlementsResponse, error) {
	if in == nil || in.GetDriverId() <= 0 {
		return nil, ErrInvalidParam
	}
	if l.svcCtx == nil || l.svcCtx.DB == nil {
		return nil, ErrDBNotConfigured
	}

	filter := repository.SettlementListFilter{
		DriverID: uint64(in.GetDriverId()),
		Status:   in.GetStatus(),
		Page:     in.GetPage(),
		PageSize: in.GetPageSize(),
	}
	if in.GetStartAt() > 0 {
		startAt := time.Unix(in.GetStartAt(), 0)
		filter.StartAt = &startAt
	}
	if in.GetEndAt() > 0 {
		endAt := time.Unix(in.GetEndAt(), 0)
		filter.EndAt = &endAt
	}

	records, total, err := repository.NewSettlementRepo(l.svcCtx.DB).List(l.ctx, filter)
	if err != nil {
		return nil, err
	}

	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	resp := &proto.ListSettlementsResponse{
		Records:  make([]*proto.SettlementBill, 0, len(records)),
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}
	for _, record := range records {
		resp.Records = append(resp.Records, settlementToBill(record))
	}
	return resp, nil
}

func settlementToBill(record *model.Settlement) *proto.SettlementBill {
	if record == nil {
		return nil
	}
	settledAt := record.CreatedAt.Unix()
	if record.SettledAt != nil {
		settledAt = record.SettledAt.Unix()
	}
	return &proto.SettlementBill{
		SettlementId:            int64(record.Id),
		SettlementNo:            record.SettlementNo,
		OrderId:                 int64(record.OrderId),
		DriverId:                int64(record.DriverId),
		TotalAmountCents:        priceutil.YuanToCents(record.TotalAmount),
		PlatformCommissionCents: priceutil.YuanToCents(record.PlatformCommission),
		DriverIncomeCents:       priceutil.YuanToCents(record.DriverIncome),
		Status:                  int32(record.Status),
		SettledAt:               settledAt,
		CreatedAt:               record.CreatedAt.Unix(),
	}
}
