package adminservicelogic

import (
	"context"

	"XiaoLong-Ridy/rpc/adminsvc/adminsvc"
	"XiaoLong-Ridy/rpc/adminsvc/internal/svc"
	driverproto "XiaoLong-Ridy/rpc/driversvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GetDriverCertificationLogic 处理司机认证审核详情 RPC。
type GetDriverCertificationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewGetDriverCertificationLogic 创建司机认证详情逻辑对象。
func NewGetDriverCertificationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDriverCertificationLogic {
	return &GetDriverCertificationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetDriverCertification 通过 driversvc 按审核记录 ID 查询司机认证详情。
func (l *GetDriverCertificationLogic) GetDriverCertification(in *adminsvc.DriverCertificationDetailRequest) (*adminsvc.DriverCertification, error) {
	if in.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "审核记录ID不能为空")
	}
	if l.svcCtx == nil || l.svcCtx.DriverSvc == nil {
		return nil, status.Error(codes.FailedPrecondition, "driver service is not running or downstream RPC is disabled")
	}
	item, err := l.svcCtx.DriverSvc.AdminGetCertification(l.ctx, &driverproto.AdminGetCertificationRequest{Id: in.GetId()})
	if err != nil {
		return nil, err
	}
	return adminCertificationFromDriverSvc(item), nil
}
