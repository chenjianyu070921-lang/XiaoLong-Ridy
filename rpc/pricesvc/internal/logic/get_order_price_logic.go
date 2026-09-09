package logic

import (
	"context"
	"errors"

	"XiaoLong-Ridy/common/priceutil"
	"XiaoLong-Ridy/rpc/pricesvc/internal/repository"
	"XiaoLong-Ridy/rpc/pricesvc/internal/svc"
	"XiaoLong-Ridy/rpc/pricesvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

// GetOrderPriceLogic 处理管理后台订单价格明细查询。
type GetOrderPriceLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewGetOrderPriceLogic 创建订单价格明细查询逻辑实例。
func NewGetOrderPriceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOrderPriceLogic {
	return &GetOrderPriceLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// GetOrderPrice 按订单ID查询 order_price 快照，供管理后台展示，避免后台跨服务直读价格表。
func (l *GetOrderPriceLogic) GetOrderPrice(in *proto.GetOrderPriceRequest) (*proto.OrderPriceInfo, error) {
	if in.GetOrderId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "订单ID不能为空")
	}
	repo := repository.NewOrderPriceRepo(l.svcCtx.DB)
	p, err := repo.FindByOrderId(l.ctx, uint64(in.GetOrderId()))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, status.Error(codes.NotFound, "订单价格明细不存在")
	}
	if err != nil {
		return nil, err
	}
	return &proto.OrderPriceInfo{
		Id:                   int64(p.Id),
		OrderId:              int64(p.OrderId),
		PriceRuleId:          int64(p.PriceRuleId),
		EstimatedPriceCents:  priceutil.YuanToCents(p.EstimatedPrice),
		ActualPriceCents:     priceutil.YuanToCents(p.ActualPrice),
		BaseFeeCents:         priceutil.YuanToCents(p.BaseFee),
		DistanceFeeCents:     priceutil.YuanToCents(p.DistanceFee),
		TimeFeeCents:         priceutil.YuanToCents(p.TimeFee),
		NightFeeCents:        priceutil.YuanToCents(p.NightFee),
		DynamicFeeCents:      priceutil.YuanToCents(p.DynamicFee),
		DiscountAmountCents:  priceutil.YuanToCents(p.DiscountAmount),
		PlatformSubsidyCents: priceutil.YuanToCents(p.PlatformSubsidy),
		PayableAmountCents:   priceutil.YuanToCents(p.PayableAmount),
		Status:               int32(p.Status),
	}, nil
}
