package logic

import (
	"context"
	"errors"

	"XiaoLong-Ridy/common/priceutil"
	"XiaoLong-Ridy/rpc/pricesvc/internal/model"
	"XiaoLong-Ridy/rpc/pricesvc/internal/svc"
	"XiaoLong-Ridy/rpc/pricesvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

var (
	ErrOrderIdInvalid     = errors.New("order id must be positive")
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

// SaveActualOrderPrice 实际费用落库：行程结束时由订单模块调用，将实际费用快照写入 order_price。
//
// 行为：
//   - 若 order_price 已存在（此前预估过），仅更新必要列（UpdateSelective）并把状态置「已确认(2)」；
//   - 若不存在，则新建一条（INSERT）；
//   - 整段读写都包在 db.Transaction() 里，避免对账与"价格已确认、订单未完结"等竞态（M5-2 / M5-5）。
func (l *SaveActualOrderPriceLogic) SaveActualOrderPrice(in *proto.SaveActualOrderPriceRequest) (*proto.SaveActualOrderPriceResponse, error) {
	if in.OrderId <= 0 {
		return nil, ErrOrderIdInvalid
	}
	if in.ActualPriceCents < 0 {
		return nil, ErrActualPriceInvalid
	}

	actualCents := in.ActualPriceCents
	// 若请求未显式给总价但给了明细，则以明细 total 为准
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

	// 单一变量 result：事务闭包传出主键与是否新建。
	type result struct {
		id      int64
		created bool
	}
	var res result

	err := l.svcCtx.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		var existing model.OrderPrice
		findErr := tx.Where("order_id = ?", uint64(in.OrderId)).First(&existing).Error

		switch {
		case findErr == nil:
			// 已存在：仅更新必要列（Updates(map)），不改 created_at / 其它无关列。
			updates := map[string]interface{}{
				"price_rule_id": uint64(in.PriceRuleId),
				"actual_price":  priceutil.CentsToYuan(actualCents),
				"base_fee":      priceutil.CentsToYuan(baseFee),
				"distance_fee":  priceutil.CentsToYuan(distanceFee),
				"time_fee":      priceutil.CentsToYuan(timeFee),
				"night_fee":     priceutil.CentsToYuan(nightFee),
				"dynamic_fee":   priceutil.CentsToYuan(dynamicFee),
				"status":        2, // 已确认
			}
			upd := tx.Model(&model.OrderPrice{}).
				Where("id = ?", existing.Id).
				Updates(updates)
			if upd.Error != nil {
				return upd.Error
			}
			res.id = int64(existing.Id)
			res.created = false
			return nil

		case errors.Is(findErr, gorm.ErrRecordNotFound):
			// 不存在：新建一条（INSERT）。
			op := &model.OrderPrice{
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
			if err := tx.Create(op).Error; err != nil {
				return err
			}
			res.id = int64(op.Id)
			res.created = true
			return nil

		default:
			return findErr
		}
	})
	if err != nil {
		return nil, err
	}

	_ = res.created // 留口用于将来按 created 分支发不同事件

	return &proto.SaveActualOrderPriceResponse{
		Success:      true,
		OrderPriceId: res.id,
	}, nil
}
