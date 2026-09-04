package adminservicelogic

import (
	"context"
	"fmt"
	"strings"

	"XiaoLong-Ridy/rpc/adminsvc/adminsvc"
	"XiaoLong-Ridy/rpc/adminsvc/internal/svc"
	usersvc "XiaoLong-Ridy/rpc/usersvc/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// FreezeUserLogic 处理用户冻结 RPC。
type FreezeUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewFreezeUserLogic 创建用户冻结逻辑。
func NewFreezeUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FreezeUserLogic {
	return &FreezeUserLogic{ctx: ctx, svcCtx: svcCtx}
}

// FreezeUser 将用户状态更新为冻结，并记录操作日志。
func (l *FreezeUserLogic) FreezeUser(in *adminsvc.ChangeUserStatusRequest) (*adminsvc.CommonResponse, error) {
	return changeUserStatus(l.ctx, l.svcCtx, in, "freeze", "冻结用户")
}

// UnfreezeUserLogic 处理用户解封 RPC。
type UnfreezeUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewUnfreezeUserLogic 创建用户解封逻辑。
func NewUnfreezeUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UnfreezeUserLogic {
	return &UnfreezeUserLogic{ctx: ctx, svcCtx: svcCtx}
}

// UnfreezeUser 将用户状态恢复为正常，并记录操作日志。
func (l *UnfreezeUserLogic) UnfreezeUser(in *adminsvc.ChangeUserStatusRequest) (*adminsvc.CommonResponse, error) {
	return changeUserStatus(l.ctx, l.svcCtx, in, "unfreeze", "解封用户")
}

// changeUserStatus 通过 usersvc 执行用户冻结/解冻，并写入后台审计日志。
// 用户状态变更以 usersvc 为权威，adminsvc 不直接更新 user 表；与司机侧对齐，仅超级管理员可执行。
func changeUserStatus(ctx context.Context, svcCtx *svc.ServiceContext, in *adminsvc.ChangeUserStatusRequest, action, detailPrefix string) (*adminsvc.CommonResponse, error) {
	if in.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "用户ID不能为空")
	}
	if in.GetAdminId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "管理员ID不能为空")
	}
	// 用户冻结/解冻与司机冻结/解冻策略对齐，统一收归超级管理员，避免运营角色越权处置乘客账号。
	if err := requireAdminRoles(ctx, svcCtx, 1); err != nil {
		return nil, err
	}
	if svcCtx == nil || svcCtx.UsersSvc == nil {
		return nil, status.Error(codes.FailedPrecondition, "user service is not running or downstream RPC is disabled")
	}

	req := &usersvc.AdminFreezeUserRequest{
		UserId:     uint64(in.GetId()),
		Reason:     strings.TrimSpace(in.GetReason()),
		OperatorId: in.GetAdminId(),
		Ip:         in.GetIp(),
	}
	var err error
	if action == "unfreeze" {
		_, err = svcCtx.UsersSvc.AdminUnfreezeUser(ctx, req)
	} else {
		_, err = svcCtx.UsersSvc.AdminFreezeUser(ctx, req)
	}
	if err != nil {
		return nil, err
	}

	detail := fmt.Sprintf("%s：%s", detailPrefix, in.GetRemark())
	if in.GetReason() != "" {
		detail = fmt.Sprintf("%s，原因：%s", detail, in.GetReason())
	}
	if err := writeAuditAfterCommitted(ctx, svcCtx, in.GetAdminId(), "user", action, "user", in.GetId(), detail, in.GetIp()); err != nil {
		return nil, err
	}
	return &adminsvc.CommonResponse{Message: "ok"}, nil
}
