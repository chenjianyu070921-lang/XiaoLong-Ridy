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

// ListDriverCertifications 通过 driversvc 查询司机认证审核列表，并关联司机和车辆摘要信息。
// 认证数据以 driversvc 为权威数据源，adminsvc 不直接读取司机域表。
func (l *ListDriverCertificationsLogic) ListDriverCertifications(in *adminsvc.DriverCertificationListRequest) (*adminsvc.DriverCertificationListResponse, error) {
	if l.svcCtx == nil || l.svcCtx.DriverSvc == nil {
		return nil, status.Error(codes.FailedPrecondition, "driver service is not running or downstream RPC is disabled")
	}
	req := &driverproto.AdminListCertificationsRequest{
		Page:      in.GetPage(),
		PageSize:  in.GetPageSize(),
		Keyword:   in.GetKeyword(),
		StartTime: in.GetStartTime(),
		EndTime:   in.GetEndTime(),
	}
	if in.GetAuditStatus() > 0 {
		auditStatus := in.GetAuditStatus()
		req.AuditStatus = &auditStatus
	}
	resp, err := l.svcCtx.DriverSvc.AdminListCertifications(l.ctx, req)
	if err != nil {
		return nil, err
	}
	list := make([]*adminsvc.DriverCertification, 0, len(resp.GetList()))
	for _, item := range resp.GetList() {
		list = append(list, adminCertificationFromDriverSvc(item))
	}
	return &adminsvc.DriverCertificationListResponse{
		List:     list,
		Total:    resp.GetTotal(),
		Page:     normalizePage(in.GetPage()),
		PageSize: normalizePageSize(in.GetPageSize()),
	}, nil
}
