package logic

import (
	"context"
	"errors"
	"time"

	"XiaoLong-Ridy/rpc/driversvc/internal/model"
	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	"XiaoLong-Ridy/rpc/driversvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

// UpdateWithdrawLogic 处理更新提现请求的逻辑结构体。
type UpdateWithdrawLogic struct {
	ctx    context.Context     // ctx：请求上下文
	svcCtx *svc.ServiceContext // svcCtx：服务上下文，持有 DB 等依赖
	logx.Logger
}

// NewUpdateWithdrawLogic 构造 UpdateWithdrawLogic 实例。
func NewUpdateWithdrawLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateWithdrawLogic {
	return &UpdateWithdrawLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UpdateWithdraw 更新提现记录，仅修改请求中显式传入的字段；状态变为打款成功时记录打款时间。
func (l *UpdateWithdrawLogic) UpdateWithdraw(in *proto.UpdateWithdrawRequest) (*proto.UpdateWithdrawResponse, error) {
	// 先按 ID 查询记录是否存在
	var w model.DriverWithdraw
	err := l.svcCtx.DB.First(&w, in.Id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("withdraw not found") // 记录不存在
	}
	if err != nil {
		return nil, err
	}

	// 仅更新显式提供的字段（optional 字段为指针，nil 表示不更新）
	updates := map[string]interface{}{}
	if in.DriverId != nil {
		updates["driver_id"] = in.GetDriverId() // 所属司机 ID
	}
	if in.WithdrawNo != nil {
		updates["withdraw_no"] = in.GetWithdrawNo() // 提现单号
	}
	if in.Amount != nil {
		updates["amount"] = in.GetAmount() // 提现金额
	}
	if in.PayeeName != nil {
		updates["payee_name"] = in.GetPayeeName() // 收款人姓名
	}
	if in.PayAccount != nil {
		updates["pay_account"] = in.GetPayAccount() // 收款账号
	}
	if in.Status != nil {
		updates["status"] = int8(in.GetStatus()) // 状态
	}
	if in.Remark != nil {
		updates["remark"] = in.GetRemark() // 失败原因/备注
	}
	if in.AppliedAt != nil {
		t := time.Unix(in.GetAppliedAt(), 0)
		updates["applied_at"] = &t // 申请时间
	}
	if in.PaidAt != nil {
		t := time.Unix(in.GetPaidAt(), 0)
		updates["paid_at"] = &t // 打款时间
	}

	// 执行更新
	if err := l.svcCtx.DB.Model(&w).Updates(updates).Error; err != nil {
		return nil, err
	}
	// 重新读取更新后的记录
	if err := l.svcCtx.DB.First(&w, in.Id).Error; err != nil {
		return nil, err
	}
	// 返回更新后的 ID 与状态
	return &proto.UpdateWithdrawResponse{
		Id:     int64(w.Id),
		Status: int32(w.Status),
	}, nil
}
