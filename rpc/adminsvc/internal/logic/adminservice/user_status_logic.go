package adminservicelogic

import (
	"context"
	"fmt"
	"time"

	"XiaoLong-Ridy/rpc/adminsvc/adminsvc"
	"XiaoLong-Ridy/rpc/adminsvc/internal/svc"

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
	return changeUserStatus(l.ctx, l.svcCtx, in, 2, "freeze", "冻结用户")
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
	return changeUserStatus(l.ctx, l.svcCtx, in, 1, "unfreeze", "解封用户")
}

// changeUserStatus 执行用户状态变更的事务性更新。
func changeUserStatus(ctx context.Context, svcCtx *svc.ServiceContext, in *adminsvc.ChangeUserStatusRequest, statusValue int32, action, detailPrefix string) (*adminsvc.CommonResponse, error) {
	if in.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "用户ID不能为空")
	}
	if in.GetAdminId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "管理员ID不能为空")
	}
	tx, err := svcCtx.MySQL.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	now := time.Now()
	res, err := tx.ExecContext(ctx, `
		UPDATE user
		SET status = ?, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`, statusValue, now, in.GetId())
	if err != nil {
		return nil, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}
	if affected == 0 {
		return nil, status.Error(codes.NotFound, "用户不存在")
	}
	detail := fmt.Sprintf("%s：%s", detailPrefix, in.GetRemark())
	if in.GetReason() != "" {
		detail = fmt.Sprintf("%s，原因：%s", detail, in.GetReason())
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO admin_operation_log (admin_id, module, action, target_type, target_id, detail, ip)
		VALUES (?, 'user', ?, 'user', ?, ?, ?)
	`, in.GetAdminId(), action, in.GetId(), detail, in.GetIp()); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &adminsvc.CommonResponse{Message: "ok"}, nil
}
