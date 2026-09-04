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

type GetDriverLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetDriverLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDriverLogic {
	return &GetDriverLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetDriverLogic) GetDriver(in *adminsvc.DriverDetailRequest) (*adminsvc.Driver, error) {
	if in.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "driver id cannot be empty")
	}
	if l.svcCtx == nil || l.svcCtx.DriverSvc == nil {
		return nil, status.Error(codes.FailedPrecondition, "driver service is not running or downstream RPC is disabled")
	}
	canViewSensitive := false
	var sensitiveAdminID int64
	if in.GetSensitive() {
		admin, err := requireSensitiveFieldPermission(l.ctx, l.svcCtx)
		if err != nil {
			return nil, err
		}
		canViewSensitive = true
		sensitiveAdminID = admin.ID
	}

	// 管理后台司机详情必须通过 driversvc 获取权威司机资料，禁止绕过服务边界直查司机表。
	resp, err := l.svcCtx.DriverSvc.GetDriver(l.ctx, &driverproto.GetDriverRequest{Id: in.GetId()})
	if err != nil {
		return nil, err
	}
	if resp.GetDriver() == nil {
		return nil, status.Error(codes.NotFound, "driver not found")
	}
	if canViewSensitive {
		if err := writeAuditAfterCommitted(l.ctx, l.svcCtx, sensitiveAdminID, "driver", "view_sensitive", "driver", in.GetId(), "查看司机完整手机号、身份证号和驾驶证号", ""); err != nil {
			return nil, err
		}
	}
	return filterAdminDriverSensitive(mapDriverFromDriverSvc(resp.GetDriver()), canViewSensitive), nil
}
