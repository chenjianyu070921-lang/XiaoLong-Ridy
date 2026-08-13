package logic

import (
	"context"
	"time"

	"XiaoLong-Ridy/rpc/driversvc/internal/model"
	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	"XiaoLong-Ridy/rpc/driversvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

// CreateWithdrawLogic 处理创建提现申请请求的逻辑结构体。
type CreateWithdrawLogic struct {
	ctx    context.Context     // ctx：请求上下文
	svcCtx *svc.ServiceContext // svcCtx：服务上下文，持有 DB 等依赖
	logx.Logger
}

// NewCreateWithdrawLogic 构造 CreateWithdrawLogic 实例。
func NewCreateWithdrawLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateWithdrawLogic {
	return &CreateWithdrawLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// CreateWithdraw 创建提现申请，状态初始为申请中（1），并记录申请时间。
func (l *CreateWithdrawLogic) CreateWithdraw(in *proto.CreateWithdrawRequest) (*proto.CreateWithdrawResponse, error) {
	now := time.Now()
	w := &model.DriverWithdraw{
		DriverId:  uint64(in.DriverId), // 所属司机 ID
		WithdrawNo: in.WithdrawNo,      // 提现单号
		Amount:    in.Amount,           // 提现金额
		PayeeName: in.PayeeName,        // 收款人姓名
		PayAccount: in.PayAccount,      // 收款账号
		Status:    1,                   // 初始状态：申请中
		AppliedAt: &now,                // 申请时间
	}
	// 写入数据库
	if err := l.svcCtx.DB.Create(w).Error; err != nil {
		return nil, err
	}
	// 返回新建提现 ID、单号与初始状态
	return &proto.CreateWithdrawResponse{
		Id:         int64(w.Id),
		WithdrawNo: w.WithdrawNo,
		Status:     int32(w.Status),
	}, nil
}
