package logic

import (
	"context"
	"errors"

	"XiaoLong-Ridy/common/priceutil"
	"XiaoLong-Ridy/rpc/pricesvc/internal/model"
	"XiaoLong-Ridy/rpc/pricesvc/internal/repository"
	"XiaoLong-Ridy/rpc/pricesvc/internal/svc"
	"XiaoLong-Ridy/rpc/pricesvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

var (
	ErrOrderIdInvalid     = errors.New("orderclient id must be positive")
	ErrActualPriceInvalid = errors.New("actual price must be non-negative")
)

type SaveActualOrderPriceLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSaveActualOrderPriceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveActualOrderPriceLogic {
	return &SaveActualOrderPriceLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 实际费用落库：行程结束时由订单模块调用，将实际费用快照写入 order_price。
//
// 行为：
//   - 若 order_price 已存在（此前预估过），仅更新实际费用与分项明细，状态置「已确认(2)」；
//   - 若不存在，则新建一条（estimated_price 暂按实际总价填充）。
func (l *SaveActualOrderPriceLogic) SaveActualOrderPrice(in *proto.SaveActualOrderPriceRequest) (*proto.SaveActualOrderPriceResponse, error) {
	if in.OrderId <= 0 {
		return nil, ErrOrderIdInvalid
	}
	if in.ActualPriceCents < 0 {
		return nil, ErrActualPriceInvalid
	}

	actualCents := in.ActualPriceCents
	// 若请求未显式给总价，但给了明细，则以明细 total 为准
	if actualCents == 0 && in.Detail != nil && in.Detail.TotalCents > 0 {
		actualCents = in.Detail.TotalCents
	}

	// 分项费用（分）
	var baseFee, distanceFee, timeFee, nightFee, dynamicFee int64
	if in.Detail != nil {
		baseFee = in.Detail.BaseFeeCents
		distanceFee = in.Detail.DistanceFeeCents
		timeFee = in.Detail.TimeFeeCents
		nightFee = in.Detail.NightFeeCents
		dynamicFee = in.Detail.DynamicFeeCents
	}

	repo := repository.NewOrderPriceRepo(l.svcCtx.DB)

	op, err := repo.FindByOrderId(l.ctx, uint64(in.OrderId))
	switch {
	case err == nil:
		// 已存在：更新实际费用与分项，状态置「已确认」
		op.PriceRuleId = uint64(in.PriceRuleId)
		op.ActualPrice = priceutil.CentsToYuan(actualCents)
		op.BaseFee = priceutil.CentsToYuan(baseFee)
		op.DistanceFee = priceutil.CentsToYuan(distanceFee)
		op.TimeFee = priceutil.CentsToYuan(timeFee)
		op.NightFee = priceutil.CentsToYuan(nightFee)
		op.DynamicFee = priceutil.CentsToYuan(dynamicFee)
		op.Status = 2 // 已确认
		if err := repo.Update(l.ctx, op); err != nil {
			return nil, err
		}
		return &proto.SaveActualOrderPriceResponse{
			Success:      true,
			OrderPriceId: int64(op.Id),
		}, nil

	case errors.Is(err, gorm.ErrRecordNotFound):
		// 不存在：新建实际费用快照
		op = &model.OrderPrice{
			OrderId:        uint64(in.OrderId),
			PriceRuleId:    uint64(in.PriceRuleId),
			EstimatedPrice: priceutil.CentsToYuan(actualCents), // 首次落库以实际总价兜底
			ActualPrice:    priceutil.CentsToYuan(actualCents),
			BaseFee:        priceutil.CentsToYuan(baseFee),
			DistanceFee:    priceutil.CentsToYuan(distanceFee),
			TimeFee:        priceutil.CentsToYuan(timeFee),
			NightFee:       priceutil.CentsToYuan(nightFee),
			DynamicFee:     priceutil.CentsToYuan(dynamicFee),
			Status:         2, // 已确认
		}
		if err := repo.Create(l.ctx, op); err != nil {
			return nil, err
		}
		return &proto.SaveActualOrderPriceResponse{
			Success:      true,
			OrderPriceId: int64(op.Id),
		}, nil

	default:
		return nil, err
	}
}
