package adminservicelogic

import (
	"context"

	"XiaoLong-Ridy/rpc/adminsvc/adminsvc"
	"XiaoLong-Ridy/rpc/adminsvc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MeLogic 处理当前管理员信息查询 RPC。
type MeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewMeLogic 创建当前管理员查询逻辑对象。
func NewMeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MeLogic {
	return &MeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Me 根据管理员 ID 查询有效管理员信息。
func (l *MeLogic) Me(in *adminsvc.MeRequest) (*adminsvc.MeResponse, error) {
	if in.GetToken() != "" {
		admin, err := validateSession(l.ctx, l.svcCtx, in.GetToken())
		if err != nil {
			return nil, err
		}
		return &adminsvc.MeResponse{Admin: toAdminPB(admin)}, nil
	}
	if in.GetAdminId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "管理员ID不能为空")
	}
	admin, err := getAdminByID(l.ctx, l.svcCtx, in.GetAdminId())
	if err != nil {
		return nil, err
	}
	return &adminsvc.MeResponse{Admin: toAdminPB(admin)}, nil
}
