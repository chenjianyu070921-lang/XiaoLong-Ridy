package logic

import (
	"context"
	"fmt"
	"time"

	"XiaoLong-Ridy/common/priceutil"
	"XiaoLong-Ridy/rpc/paysvc/internal/model"
	"XiaoLong-Ridy/rpc/paysvc/internal/repository"
	"XiaoLong-Ridy/rpc/paysvc/internal/rule"
	"XiaoLong-Ridy/rpc/paysvc/internal/svc"
	"XiaoLong-Ridy/rpc/paysvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

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

// 司机结算：计算平台抽成与司机实收，写入结算单。
func (l *SettleOrderLogic) SettleOrder(in *proto.SettleOrderRequest) (*proto.SettleOrderResponse, error) {
	// 1. 计算平台抽成与司机实收
	commission, income := rule.CalcSettlement(in.TotalAmountCents, in.CommissionRate)

	// 2. 生成结算单号并落库
	settlementNo := genSettlementNo()
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

	repo := repository.NewSettlementRepo(l.svcCtx.DB)
	if err := repo.Create(l.ctx, s); err != nil {
		return nil, err
	}

	return &proto.SettleOrderResponse{
		SettlementId:           int64(s.Id),
		SettlementNo:           settlementNo,
		PlatformCommissionCents: commission,
		DriverIncomeCents:      income,
	}, nil
}
