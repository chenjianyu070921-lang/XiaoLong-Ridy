package adminservicelogic

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"XiaoLong-Ridy/rpc/adminsvc/adminsvc"
	"XiaoLong-Ridy/rpc/adminsvc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GetUserLogic 处理用户详情查询 RPC。
type GetUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewGetUserLogic 创建用户详情查询逻辑对象。
func NewGetUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserLogic {
	return &GetUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetUser 根据用户 ID 查询乘客用户详情。
func (l *GetUserLogic) GetUser(in *adminsvc.UserDetailRequest) (*adminsvc.User, error) {
	if in.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "用户ID不能为空")
	}
	row := l.svcCtx.MySQL.QueryRowContext(l.ctx, `
		SELECT id, phone, nickname, avatar_url, gender, real_name, id_card_no,
		       register_source, status, created_at, updated_at
		FROM user
		WHERE id = ? AND deleted_at IS NULL
	`, in.GetId())
	var item adminsvc.User
	var createdAt, updatedAt sql.NullTime
	err := row.Scan(&item.Id, &item.Phone, &item.Nickname, &item.AvatarUrl, &item.Gender, &item.RealName, &item.IdCardNo, &item.RegisterSource, &item.Status, &createdAt, &updatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, status.Error(codes.NotFound, "用户不存在")
	}
	if err != nil {
		return nil, fmt.Errorf("scan user: %w", err)
	}
	item.CreatedAt = formatNullTime(createdAt)
	item.UpdatedAt = formatNullTime(updatedAt)
	return &item, nil
}
