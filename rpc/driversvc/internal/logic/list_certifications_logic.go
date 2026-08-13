package logic

import (
	"context"

	"XiaoLong-Ridy/rpc/driversvc/internal/model"
	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	"XiaoLong-Ridy/rpc/driversvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

// ListCertificationsLogic 处理分页查询认证列表请求的逻辑结构体。
type ListCertificationsLogic struct {
	ctx    context.Context     // ctx：请求上下文
	svcCtx *svc.ServiceContext // svcCtx：服务上下文，持有 DB 等依赖
	logx.Logger
}

// NewListCertificationsLogic 构造 ListCertificationsLogic 实例。
func NewListCertificationsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListCertificationsLogic {
	return &ListCertificationsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ListCertifications 分页查询认证列表，支持按司机、审核状态过滤。
func (l *ListCertificationsLogic) ListCertifications(in *proto.ListCertificationsRequest) (*proto.ListCertificationsResponse, error) {
	// 解析分页参数，并做默认值与上限保护
	page := int(in.Page)
	if page <= 0 {
		page = 1 // 默认第 1 页
	}
	pageSize := int(in.PageSize)
	if pageSize <= 0 {
		pageSize = 20 // 默认每页 20 条
	}
	if pageSize > 100 {
		pageSize = 100 // 每页最多 100 条
	}

	// 构建查询条件
	query := l.svcCtx.DB.Model(&model.DriverCertification{})
	if in.DriverId != 0 {
		query = query.Where("driver_id = ?", in.DriverId) // 按司机过滤
	}
	if in.AuditStatus != 0 {
		query = query.Where("audit_status = ?", in.AuditStatus) // 按审核状态过滤
	}

	// 统计符合条件的总记录数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// 分页查询，按 ID 倒序返回
	var certs []model.DriverCertification
	if err := query.Order("id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&certs).Error; err != nil {
		return nil, err
	}

	// 转换为精简的认证摘要列表
	list := make([]*proto.CertificationSummary, 0, len(certs))
	for _, c := range certs {
		list = append(list, &proto.CertificationSummary{
			Id:          int64(c.Id),         // 认证 ID
			DriverId:    int64(c.DriverId),   // 所属司机 ID
			VehicleId:   int64(c.VehicleId),  // 关联车辆 ID
			AuditStatus: int32(c.AuditStatus), // 审核状态
			CreatedAt:   c.CreatedAt.Unix(),  // 创建时间
		})
	}

	// 返回列表、总数与分页信息
	return &proto.ListCertificationsResponse{
		List:     list,
		Total:    total,
		Page:     int32(page),
		PageSize: int32(pageSize),
	}, nil
}
