package adminservicelogic

import (
	"context"
	"strings"

	"XiaoLong-Ridy/rpc/adminsvc/adminsvc"
	"XiaoLong-Ridy/rpc/adminsvc/internal/svc"
	driverproto "XiaoLong-Ridy/rpc/driversvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// HandleDriverWithdrawLogic 处理管理后台审核司机提现申请 RPC。
type HandleDriverWithdrawLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewHandleDriverWithdrawLogic 创建司机提现审核逻辑对象。
func NewHandleDriverWithdrawLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HandleDriverWithdrawLogic {
	return &HandleDriverWithdrawLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// HandleDriverWithdraw 仅超管可执行资金审核：approve=true 打款成功，approve=false 打款失败。
// 审核结果落库后写入审计并联动司机端通知；通知失败写补偿 outbox，不回滚审核结果。
func (l *HandleDriverWithdrawLogic) HandleDriverWithdraw(in *adminsvc.DriverWithdrawHandleRequest) (*adminsvc.CommonResponse, error) {
	if err := requireAdminRoles(l.ctx, l.svcCtx, 1); err != nil {
		return nil, err
	}
	if in.GetId() <= 0 || in.GetAdminId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "提现ID和管理员ID不能为空")
	}
	remark := strings.TrimSpace(in.GetRemark())
	if !in.GetApprove() && remark == "" {
		return nil, status.Error(codes.InvalidArgument, "打款失败时必须填写原因")
	}
	if l.svcCtx == nil || l.svcCtx.DriverSvc == nil {
		return nil, status.Error(codes.FailedPrecondition, "driver service is not running or downstream RPC is disabled")
	}
	resp, err := l.svcCtx.DriverSvc.AuditWithdraw(l.ctx, &driverproto.AuditWithdrawRequest{
		WithdrawId: in.GetId(),
		Approve:    in.GetApprove(),
		Remark:     remark,
		OperatorId: in.GetAdminId(),
		Ip:         in.GetIp(),
	})
	if err != nil {
		return nil, err
	}
	action := "approve"
	result := "打款成功"
	if !in.GetApprove() {
		action = "reject"
		result = "打款失败"
	}
	detail := "提现审核：" + result
	if remark != "" {
		detail += "，原因：" + remark
	}
	if err := writeAuditAfterCommitted(l.ctx, l.svcCtx, in.GetAdminId(), "driver_withdraw", action, "driver_withdraw", in.GetId(), detail, in.GetIp()); err != nil {
		return nil, err
	}
	if driverID := resp.GetDriverId(); driverID > 0 {
		title := "提现打款成功"
		content := "您的提现申请已打款成功"
		if !in.GetApprove() {
			title = "提现打款失败"
			content = "您的提现申请打款失败：" + remark
		}
		if err := notifyDriverBestEffort(l.ctx, l.svcCtx, driverID, in.GetAdminId(), title, content, "withdraw_"+action, in.GetIp()); err != nil {
			return nil, err
		}
	}
	return &adminsvc.CommonResponse{Message: "ok"}, nil
}
