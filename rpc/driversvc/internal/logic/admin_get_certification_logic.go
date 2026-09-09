package logic

import (
	"context"
	"errors"

	"XiaoLong-Ridy/rpc/driversvc/internal/repository"
	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	"XiaoLong-Ridy/rpc/driversvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AdminGetCertificationLogic 管理后台司机认证详情查询。
type AdminGetCertificationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewAdminGetCertificationLogic 创建管理后台认证详情逻辑实例。
func NewAdminGetCertificationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminGetCertificationLogic {
	return &AdminGetCertificationLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// AdminGetCertification 按认证记录 ID 查询详情，供 adminsvc 消费，避免后台直连司机域表。
func (l *AdminGetCertificationLogic) AdminGetCertification(in *proto.AdminGetCertificationRequest) (*proto.AdminCertification, error) {
	if in == nil || in.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "认证记录ID不能为空")
	}
	if l.svcCtx == nil || l.svcCtx.CertificationRepository == nil {
		return nil, status.Error(codes.FailedPrecondition, "certification repository not ready")
	}
	row, err := l.svcCtx.CertificationRepository.AdminGetByID(l.ctx, uint64(in.GetId()))
	if errors.Is(err, repository.ErrCertificationNotFound) {
		return nil, status.Error(codes.NotFound, "司机认证记录不存在")
	}
	if err != nil {
		return nil, err
	}
	return toAdminCertification(row), nil
}
