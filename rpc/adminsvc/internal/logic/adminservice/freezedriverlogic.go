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

// FreezeDriverLogic 处理管理后台冻结司机 RPC。
type FreezeDriverLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewFreezeDriverLogic 创建冻结司机逻辑对象。
func NewFreezeDriverLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FreezeDriverLogic {
	return &FreezeDriverLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// FreezeDriver 校验后台权限后调用 driversvc 冻结司机，并写入审计与司机端通知。
// 司机状态修改以 driversvc 为权威，adminsvc 不直接更新司机表。
func (l *FreezeDriverLogic) FreezeDriver(in *adminsvc.FreezeDriverRequest) (*adminsvc.CommonResponse, error) {
	if err := requireAdminRoles(l.ctx, l.svcCtx, 1); err != nil {
		return nil, err
	}
	if in.GetId() <= 0 || in.GetAdminId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "司机ID和管理员ID不能为空")
	}
	reason := strings.TrimSpace(in.GetReason())
	if reason == "" {
		return nil, status.Error(codes.InvalidArgument, "冻结原因不能为空")
	}
	if l.svcCtx == nil || l.svcCtx.DriverSvc == nil {
		return nil, status.Error(codes.FailedPrecondition, "driver service is not running or downstream RPC is disabled")
	}
	if _, err := l.svcCtx.DriverSvc.FreezeDriver(l.ctx, &driverproto.FreezeDriverRequest{
		DriverId:   in.GetId(),
		Reason:     reason,
		OperatorId: in.GetAdminId(),
		Ip:         in.GetIp(),
	}); err != nil {
		return nil, err
	}
	detail := "冻结司机：" + reason
	if strings.TrimSpace(in.GetRemark()) != "" {
		detail += "，备注：" + strings.TrimSpace(in.GetRemark())
	}
	if err := writeAuditAfterCommitted(l.ctx, l.svcCtx, in.GetAdminId(), "driver", "freeze", "driver", in.GetId(), detail, in.GetIp()); err != nil {
		return nil, err
	}
	if err := notifyDriverBestEffort(l.ctx, l.svcCtx, in.GetId(), in.GetAdminId(), "账号已被冻结", reason, "freeze", in.GetIp()); err != nil {
		return nil, err
	}
	return &adminsvc.CommonResponse{Message: "ok"}, nil
}
