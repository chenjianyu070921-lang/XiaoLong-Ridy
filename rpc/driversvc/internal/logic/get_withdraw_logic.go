package logic

import (
	"context"
	"errors"

	"XiaoLong-Ridy/rpc/driversvc/internal/model"
	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	"XiaoLong-Ridy/rpc/driversvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

// GetWithdrawLogic 处理查询提现详情请求的逻辑结构体。
type GetWithdrawLogic struct {
	ctx    context.Context     // ctx：请求上下文
	svcCtx *svc.ServiceContext // svcCtx：服务上下文，持有 DB 等依赖
	logx.Logger
}

// NewGetWithdrawLogic 构造 GetWithdrawLogic 实例。
func NewGetWithdrawLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetWithdrawLogic {
	return &GetWithdrawLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetWithdraw 根据提现 ID 查询提现完整信息。
func (l *GetWithdrawLogic) GetWithdraw(in *proto.GetWithdrawRequest) (*proto.GetWithdrawResponse, error) {
	// 按 ID 查询记录
	var w model.DriverWithdraw
	err := l.svcCtx.DB.First(&w, in.Id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("withdraw not found") // 记录不存在
	}
	if err != nil {
		return nil, err
	}
	// 组装并返回提现详情
	return &proto.GetWithdrawResponse{
		Withdraw: &proto.Withdraw{
			Id:          int64(w.Id),                // 提现 ID
			DriverId:    int64(w.DriverId),          // 所属司机 ID
			WithdrawNo:  w.WithdrawNo,               // 提现单号
			Amount:      w.Amount,                  // 提现金额
			PayeeName:   w.PayeeName,               // 收款人姓名
			PayAccount:  w.PayAccount,              // 收款账号
			Status:      int32(w.Status),           // 状态
			Remark:      w.Remark,                 // 失败原因/备注
			AppliedAt:   unixOrZero(w.AppliedAt),   // 申请时间
			PaidAt:      unixOrZero(w.PaidAt),      // 打款时间
			CreatedAt:   w.CreatedAt.Unix(),        // 创建时间
		},
	}, nil
}
