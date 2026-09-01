package logic

import (
	"context"
	"strings"

	"XiaoLong-Ridy/api/admin/internal/model"
	"XiaoLong-Ridy/api/admin/internal/svc"
	"XiaoLong-Ridy/api/admin/internal/types"
	adminclient "XiaoLong-Ridy/rpc/adminsvc/client/adminservice"
)

// DriverLogic 负责管理后台司机审核相关动作的 HTTP 适配。
type DriverLogic struct {
	ctx *svc.ServiceContext
}

// NewDriverLogic 创建司机审核逻辑。
func NewDriverLogic(ctx *svc.ServiceContext) *DriverLogic {
	return &DriverLogic{ctx: ctx}
}

// ListDrivers 查询司机基础资料列表。
func (l *DriverLogic) ListDrivers(ctx context.Context, req types.DriverListRequest) (*types.PageResult, error) {
	resp, err := l.ctx.AdminSvc.ListDrivers(ctx, &adminclient.DriverListRequest{
		Page:     int32(req.Page),
		PageSize: int32(req.PageSize),
		Keyword:  req.Keyword,
		Status:   req.Status,
	})
	if err != nil {
		return nil, err
	}
	items := make([]types.DriverDTO, 0, len(resp.List))
	for _, item := range resp.List {
		items = append(items, toDriverDTO(item))
	}
	return &types.PageResult{List: items, Total: resp.Total, Page: int(resp.Page), PageSize: int(resp.PageSize)}, nil
}

// DriverDetail 查询司机基础资料详情。
// sensitive 为 true 时表示后台管理员显式申请查看完整手机号、身份证号和驾驶证号，由 adminsvc 做二次授权与审计。
func (l *DriverLogic) DriverDetail(ctx context.Context, id int64, sensitive bool) (*types.DriverDTO, error) {
	resp, err := l.ctx.AdminSvc.GetDriver(ctx, &adminclient.DriverDetailRequest{Id: id, Sensitive: sensitive})
	if err != nil {
		return nil, err
	}
	item := toDriverDTO(resp)
	return &item, nil
}

// FreezeDriver 冻结司机账号，HTTP 层只做参数适配，状态变更由 adminsvc 转发到 driversvc。
func (l *DriverLogic) FreezeDriver(ctx context.Context, id int64, req types.DriverFreezeRequest, session *model.AdminSession, ip string) error {
	adminID := int64(0)
	if session != nil {
		adminID = session.AdminID
	}
	_, err := l.ctx.AdminSvc.FreezeDriver(ctx, &adminclient.FreezeDriverRequest{
		Id:      id,
		Reason:  req.Reason,
		Remark:  req.Remark,
		AdminId: adminID,
		Ip:      ip,
	})
	return err
}

// UnfreezeDriver 解冻司机账号，HTTP 层只做参数适配，状态变更由 adminsvc 转发到 driversvc。
func (l *DriverLogic) UnfreezeDriver(ctx context.Context, id int64, req types.DriverFreezeRequest, session *model.AdminSession, ip string) error {
	adminID := int64(0)
	if session != nil {
		adminID = session.AdminID
	}
	_, err := l.ctx.AdminSvc.UnfreezeDriver(ctx, &adminclient.FreezeDriverRequest{
		Id:      id,
		Reason:  req.Reason,
		Remark:  req.Remark,
		AdminId: adminID,
		Ip:      ip,
	})
	return err
}

// ListWithdrawals 查询司机提现申请列表，HTTP 层只做参数适配与类型转换。
func (l *DriverLogic) ListWithdrawals(ctx context.Context, page, pageSize int32, status int32, driverID int64, keyword string) (*types.PageResult, error) {
	resp, err := l.ctx.AdminSvc.ListDriverWithdrawals(ctx, &adminclient.DriverWithdrawListRequest{
		Page:     page,
		PageSize: pageSize,
		Status:   status,
		DriverId: driverID,
		Keyword:  keyword,
	})
	if err != nil {
		return nil, err
	}
	items := make([]types.DriverWithdrawalDTO, 0, len(resp.GetList()))
	for _, w := range resp.GetList() {
		items = append(items, types.DriverWithdrawalDTO{
			ID:         w.GetId(),
			DriverID:   w.GetDriverId(),
			WithdrawNo: w.GetWithdrawNo(),
			Amount:     w.GetAmount(),
			PayeeName:  w.GetPayeeName(),
			PayAccount: w.GetPayAccount(),
			Status:     w.GetStatus(),
			Remark:     w.GetRemark(),
			AppliedAt:  w.GetAppliedAt(),
			PaidAt:     w.GetPaidAt(),
			CreatedAt:  w.GetCreatedAt(),
		})
	}
	return &types.PageResult{List: items, Total: resp.GetTotal(), Page: int(resp.GetPage()), PageSize: int(resp.GetPageSize())}, nil
}

// HandleWithdrawal 审核司机提现申请，approve=true 打款成功，approve=false 打款失败。
func (l *DriverLogic) HandleWithdrawal(ctx context.Context, id int64, approve bool, req types.DriverWithdrawHandleRequest, session *model.AdminSession, ip string) error {
	adminID := int64(0)
	if session != nil {
		adminID = session.AdminID
	}
	_, err := l.ctx.AdminSvc.HandleDriverWithdraw(ctx, &adminclient.DriverWithdrawHandleRequest{
		Id:      id,
		AdminId: adminID,
		Approve: approve,
		Remark:  req.Remark,
		Ip:      ip,
	})
	return err
}

func toDriverDTO(item *adminclient.Driver) types.DriverDTO {
	if item == nil {
		return types.DriverDTO{}
	}
	return types.DriverDTO{
		ID:              item.Id,
		Phone:           item.Phone,
		RealName:        item.RealName,
		IDCardNo:        item.IdCardNo,
		DriverLicenseNo: item.DriverLicenseNo,
		AvatarURL:       item.AvatarUrl,
		Status:          item.Status,
		OnlineStatus:    item.OnlineStatus,
		VehicleID:       item.VehicleId,
		PlateNo:         item.PlateNo,
		VehicleStatus:   item.VehicleStatus,
		CertificationID: item.CertificationId,
		AuditStatus:     item.AuditStatus,
		AuditRemark:     item.AuditRemark,
		CreatedAt:       item.CreatedAt,
		UpdatedAt:       item.UpdatedAt,
	}
}

// ListCertifications 查询司机审核列表。
func (l *DriverLogic) ListCertifications(ctx context.Context, req types.DriverCertificationListRequest) (*types.PageResult, error) {
	resp, err := l.ctx.AdminSvc.ListDriverCertifications(ctx, &adminclient.DriverCertificationListRequest{
		Page:        int32(req.Page),
		PageSize:    int32(req.PageSize),
		Keyword:     req.Keyword,
		AuditStatus: req.AuditStatus,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
	})
	if err != nil {
		return nil, err
	}
	items := make([]types.DriverCertificationDTO, 0, len(resp.List))
	for _, item := range resp.List {
		items = append(items, types.DriverCertificationDTO{
			ID:                item.Id,
			DriverID:          item.DriverId,
			VehicleID:         item.VehicleId,
			DriverPhone:       item.DriverPhone,
			DriverName:        item.DriverName,
			DriverStatus:      item.DriverStatus,
			PlateNo:           item.PlateNo,
			VehicleStatus:     item.VehicleStatus,
			IDCardFrontURL:    item.IdCardFrontUrl,
			IDCardBackURL:     item.IdCardBackUrl,
			DriverLicenseURL:  item.DriverLicenseUrl,
			VehicleLicenseURL: item.VehicleLicenseUrl,
			AuditStatus:       item.AuditStatus,
			AuditRemark:       item.AuditRemark,
			AuditedBy:         item.AuditedBy,
			AuditedAt:         item.AuditedAt,
			CreatedAt:         item.CreatedAt,
			UpdatedAt:         item.UpdatedAt,
		})
	}
	return &types.PageResult{List: items, Total: resp.Total, Page: int(resp.Page), PageSize: int(resp.PageSize)}, nil
}

// CertificationDetail 查询司机审核详情。
func (l *DriverLogic) CertificationDetail(ctx context.Context, id int64) (*types.DriverCertificationDTO, error) {
	resp, err := l.ctx.AdminSvc.GetDriverCertification(ctx, &adminclient.DriverCertificationDetailRequest{Id: id})
	if err != nil {
		return nil, err
	}
	return &types.DriverCertificationDTO{
		ID:                resp.Id,
		DriverID:          resp.DriverId,
		VehicleID:         resp.VehicleId,
		DriverPhone:       resp.DriverPhone,
		DriverName:        resp.DriverName,
		DriverStatus:      resp.DriverStatus,
		PlateNo:           resp.PlateNo,
		VehicleStatus:     resp.VehicleStatus,
		IDCardFrontURL:    resp.IdCardFrontUrl,
		IDCardBackURL:     resp.IdCardBackUrl,
		DriverLicenseURL:  resp.DriverLicenseUrl,
		VehicleLicenseURL: resp.VehicleLicenseUrl,
		AuditStatus:       resp.AuditStatus,
		AuditRemark:       resp.AuditRemark,
		AuditedBy:         resp.AuditedBy,
		AuditedAt:         resp.AuditedAt,
		CreatedAt:         resp.CreatedAt,
		UpdatedAt:         resp.UpdatedAt,
	}, nil
}

// ApproveCertification 审核通过司机认证。
func (l *DriverLogic) ApproveCertification(ctx context.Context, id int64, req types.AuditRequest, session *model.AdminSession, ip string) error {
	remark := strings.TrimSpace(req.Remark)
	if remark == "" {
		remark = "审核通过"
	}
	_, err := l.ctx.AdminSvc.ApproveDriverCertification(ctx, &adminclient.AuditDriverCertificationRequest{
		Id:      id,
		Remark:  remark,
		AdminId: session.AdminID,
		Ip:      ip,
	})
	return err
}

// RejectCertification 驳回司机认证。
func (l *DriverLogic) RejectCertification(ctx context.Context, id int64, req types.AuditRequest, session *model.AdminSession, ip string) error {
	remark := strings.TrimSpace(req.Remark)
	if remark == "" {
		remark = "资料不完整"
	}
	_, err := l.ctx.AdminSvc.RejectDriverCertification(ctx, &adminclient.AuditDriverCertificationRequest{
		Id:      id,
		Remark:  remark,
		AdminId: session.AdminID,
		Ip:      ip,
	})
	return err
}
