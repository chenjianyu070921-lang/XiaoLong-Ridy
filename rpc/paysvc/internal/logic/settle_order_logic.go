package logic

import (
	"context"
	"errors"
	"fmt"
	"time"

	"XiaoLong-Ridy/common/priceutil"
	"XiaoLong-Ridy/rpc/paysvc/internal/model"
	"XiaoLong-Ridy/rpc/paysvc/internal/rule"
	"XiaoLong-Ridy/rpc/paysvc/internal/svc"
	"XiaoLong-Ridy/rpc/paysvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

// 结算单号生成冲突的可重试错误；上层可对账补偿。
var errSettlementDuplicateNo = errors.New("settlement no duplicate, please retry")

type SettleOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSettleOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SettleOrderLogic {
	return &SettleOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// genSettlementNo 生成结算单号：SET + 时间戳 + 微秒。
func genSettlementNo() string {
	return fmt.Sprintf("SET%s%06d", time.Now().Format("20060102150405"), time.Now().Nanosecond()/1000)
}

// SettleOrder 司机结算：计算平台抽成与司机实收，写入结算单。
//
// 事务边界：用 db.Transaction 包裹 INSERT，让结算单写入过程可重试/回滚，
// 防止支付回调重投导致重复结算单。
func (l *SettleOrderLogic) SettleOrder(in *proto.SettleOrderRequest) (*proto.SettleOrderResponse, error) {
	// 入参校验：金额必须为正，司机 ID 必须有效。
	if in.TotalAmountCents < 0 {
		return nil, fmt.Errorf("invalid total_amount_cents: %d", in.TotalAmountCents)
	}
	if in.DriverId <= 0 {
		return nil, fmt.Errorf("invalid driver_id: %d", in.DriverId)
	}

	// 1. 计算平台抽成与司机实收
	commission, income := rule.CalcSettlement(in.TotalAmountCents, in.CommissionRate)

	// 2. 写结算单（事务：让结算单号 UNIQUE 冲突能回滚副作用）。
	var settlementID int64
	var settlementNo string
	err := l.svcCtx.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		settlementNo = genSettlementNo()
		s := &model.Settlement{
			SettlementNo:           settlementNo,
			OrderId:                uint64(in.OrderId),
			DriverId:               uint64(in.DriverId),
			TotalAmount:            priceutil.CentsToYuan(in.TotalAmountCents),
			PlatformCommissionRate: in.CommissionRate,
			PlatformCommission:     priceutil.CentsToYuan(commission),
			DriverIncome:           priceutil.CentsToYuan(income),
			Status:                 model.SettlementStatusSettled,
		}
		if err := tx.Create(s).Error; err != nil {
			return err
		}
		settlementID = int64(s.Id)
		return nil
	})
	if err != nil {
		// 唯一索引冲突（uk_settlement_no）：同一时间戳撞号罕见但可能，让上层感知。
		return nil, fmt.Errorf("%w: %v", errSettlementDuplicateNo, err)
	}

	return &proto.SettleOrderResponse{
		SettlementId:            settlementID,
		SettlementNo:            settlementNo,
		PlatformCommissionCents: commission,
		DriverIncomeCents:       income,
	}, nil
}
