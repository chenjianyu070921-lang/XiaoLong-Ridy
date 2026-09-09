package logic

import (
	"context"
	"strings"
	"time"

	"XiaoLong-Ridy/rpc/driversvc/internal/repository"
	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	"XiaoLong-Ridy/rpc/driversvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AdminListCertificationsLogic 管理后台司机认证审核列表查询。
type AdminListCertificationsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewAdminListCertificationsLogic 创建管理后台认证列表逻辑实例。
func NewAdminListCertificationsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminListCertificationsLogic {
	return &AdminListCertificationsLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// AdminListCertifications 按关键词/审核状态/时间范围分页查询认证记录，供 adminsvc 消费。
func (l *AdminListCertificationsLogic) AdminListCertifications(in *proto.AdminListCertificationsRequest) (*proto.AdminListCertificationsResponse, error) {
	if in == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	if l.svcCtx == nil || l.svcCtx.CertificationRepository == nil {
		return nil, status.Error(codes.FailedPrecondition, "certification repository not ready")
	}
	filter := repository.AdminCertificationFilter{
		Page:     int(in.GetPage()),
		PageSize: int(in.GetPageSize()),
		Keyword:  strings.TrimSpace(in.GetKeyword()),
	}
	if in.AuditStatus != nil {
		if in.GetAuditStatus() <= 0 || in.GetAuditStatus() > 3 {
			return nil, status.Error(codes.InvalidArgument, "审核状态筛选不合法")
		}
		filter.AuditStatus = int8(in.GetAuditStatus())
	}
	if ts := strings.TrimSpace(in.GetStartTime()); ts != "" {
		t, err := time.ParseInLocation("2006-01-02 15:04:05", ts, time.Local)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "start_time 格式不合法")
		}
		filter.StartTime = t
	}
	if ts := strings.TrimSpace(in.GetEndTime()); ts != "" {
		t, err := time.ParseInLocation("2006-01-02 15:04:05", ts, time.Local)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "end_time 格式不合法")
		}
		filter.EndTime = t
	}

	rows, total, err := l.svcCtx.CertificationRepository.AdminList(l.ctx, filter)
	if err != nil {
		return nil, err
	}
	resp := &proto.AdminListCertificationsResponse{Total: total}
	for _, row := range rows {
		resp.List = append(resp.List, toAdminCertification(row))
	}
	return resp, nil
}
