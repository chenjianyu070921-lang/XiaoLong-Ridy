package adminservicelogic

import (
	"context"

	"XiaoLong-Ridy/rpc/adminsvc/adminsvc"
	"XiaoLong-Ridy/rpc/adminsvc/internal/svc"

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

// GetDriverCertification 根据审核记录 ID 查询司机认证详情。
func (l *GetDriverCertificationLogic) GetDriverCertification(in *adminsvc.DriverCertificationDetailRequest) (*adminsvc.DriverCertification, error) {
	if in.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "审核记录ID不能为空")
	}
	row := l.svcCtx.MySQL.QueryRowContext(l.ctx, `
		SELECT c.id, c.driver_id, c.vehicle_id,
		       COALESCE(d.phone, ''), COALESCE(d.real_name, ''), COALESCE(d.status, 0),
		       COALESCE(v.plate_no, ''), COALESCE(v.status, 0),
		       c.id_card_front_url, c.id_card_back_url, c.driver_license_url, c.vehicle_license_url,
		       c.audit_status, c.audit_remark, c.audited_by, c.audited_at, c.created_at, c.updated_at
		FROM driver_certification c
		LEFT JOIN driver d ON d.id = c.driver_id
		LEFT JOIN driver_vehicle v ON v.id = c.vehicle_id
		WHERE c.id = ?
	`, in.GetId())
	return scanCertificationRow(row)
}
