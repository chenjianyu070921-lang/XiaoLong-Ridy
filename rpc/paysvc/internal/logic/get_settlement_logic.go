package logic

import (
	"context"
	"errors"
	"fmt"

	"XiaoLong-Ridy/common/priceutil"
	"XiaoLong-Ridy/rpc/paysvc/internal/repository"
	"XiaoLong-Ridy/rpc/paysvc/internal/svc"
	"XiaoLong-Ridy/rpc/paysvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

// GetSettlementLogic 处理管理后台订单结算查询。
type GetSettlementLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewGetSettlementLogic 创建订单结算查询逻辑实例。
func NewGetSettlementLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSettlementLogic {
	return &GetSettlementLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// GetSettlement 按订单ID查询结算单，供管理后台展示，避免后台跨服务直读结算表。
func (l *GetSettlementLogic) GetSettlement(in *proto.GetSettlementRequest) (*proto.GetSettlementResponse, error) {
	if in.GetOrderId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "订单ID不能为空")
	}
	repo := repository.NewSettlementRepo(l.svcCtx.DB)
	s, err := repo.FindByOrderId(l.ctx, uint64(in.GetOrderId()))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, status.Error(codes.NotFound, "订单结算记录不存在")
	}
	if err != nil {
		return nil, err
	}
	var settledAt int64
	if s.SettledAt != nil {
		settledAt = s.SettledAt.Unix()
	}
	return &proto.GetSettlementResponse{
		SettlementId:            int64(s.Id),
		SettlementNo:            s.SettlementNo,
		OrderId:                 int64(s.OrderId),
		DriverId:                int64(s.DriverId),
		TotalAmountCents:        priceutil.YuanToCents(s.TotalAmount),
		PlatformCommissionRate:  fmt.Sprintf("%g", s.PlatformCommissionRate),
		PlatformCommissionCents: priceutil.YuanToCents(s.PlatformCommission),
		DriverIncomeCents:       priceutil.YuanToCents(s.DriverIncome),
		Status:                  int32(s.Status),
		SettledAt:               settledAt,
	}, nil
}
