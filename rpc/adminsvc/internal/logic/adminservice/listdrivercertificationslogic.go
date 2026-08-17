package adminservicelogic

import (
	"context"

	"XiaoLong-Ridy/rpc/adminsvc/adminsvc"
	"XiaoLong-Ridy/rpc/adminsvc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

// ListDriverCertificationsLogic 处理司机认证审核列表 RPC。
type ListDriverCertificationsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewListDriverCertificationsLogic 创建司机认证审核列表逻辑对象。
func NewListDriverCertificationsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListDriverCertificationsLogic {
	return &ListDriverCertificationsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ListDriverCertifications 查询司机认证审核列表，并关联司机和车辆摘要信息。
func (l *ListDriverCertificationsLogic) ListDriverCertifications(in *adminsvc.DriverCertificationListRequest) (*adminsvc.DriverCertificationListResponse, error) {
	where, args := buildCertificationWhere(in)
	countSQL := `
		SELECT COUNT(1)
		FROM driver_certification c
		LEFT JOIN driver d ON d.id = c.driver_id
		LEFT JOIN driver_vehicle v ON v.id = c.vehicle_id
	` + where
	var total int64
	if err := l.svcCtx.MySQL.QueryRowContext(l.ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, err
	}
	limit := normalizePageSize(in.GetPageSize())
	queryArgs := append(args, limit, offset(in.GetPage(), in.GetPageSize()))
	rows, err := l.svcCtx.MySQL.QueryContext(l.ctx, `
		SELECT c.id, c.driver_id, c.vehicle_id,
		       COALESCE(d.phone, ''), COALESCE(d.real_name, ''), COALESCE(d.status, 0),
		       COALESCE(v.plate_no, ''), COALESCE(v.status, 0),
		       c.id_card_front_url, c.id_card_back_url, c.driver_license_url, c.vehicle_license_url,
		       c.audit_status, c.audit_remark, c.audited_by, c.audited_at, c.created_at, c.updated_at
		FROM driver_certification c
		LEFT JOIN driver d ON d.id = c.driver_id
		LEFT JOIN driver_vehicle v ON v.id = c.vehicle_id
	`+where+`
		ORDER BY c.id DESC
		LIMIT ? OFFSET ?
	`, queryArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]*adminsvc.DriverCertification, 0)
	for rows.Next() {
		item, err := scanCertificationRows(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, item)
	}
	return &adminsvc.DriverCertificationListResponse{List: list, Total: total, Page: normalizePage(in.GetPage()), PageSize: limit}, rows.Err()
}
