package logic

import (
	"context"

	"XiaoLong-Ridy/api/admin/internal/model"
	"XiaoLong-Ridy/api/admin/internal/svc"
	"XiaoLong-Ridy/api/admin/internal/types"
	adminclient "XiaoLong-Ridy/rpc/adminsvc/client/adminservice"
)

// CouponLogic 负责优惠券 HTTP 请求到 adminsvc 的参数转换。
type CouponLogic struct {
	ctx *svc.ServiceContext
}

// NewCouponLogic 创建优惠券逻辑。
func NewCouponLogic(ctx *svc.ServiceContext) *CouponLogic {
	return &CouponLogic{ctx: ctx}
}

// List 查询优惠券模板列表。
func (l *CouponLogic) List(ctx context.Context, req types.CouponListRequest) (*types.PageResult, error) {
	resp, err := l.ctx.AdminSvc.ListCoupons(ctx, &adminclient.CouponListRequest{
		Page: int32(req.Page), PageSize: int32(req.PageSize), Keyword: req.Keyword,
		Type: req.Type, Status: req.Status, StartTime: req.StartTime, EndTime: req.EndTime,
	})
	if err != nil {
		return nil, err
	}
	items := make([]types.CouponDTO, 0, len(resp.List))
	for _, item := range resp.List {
		items = append(items, couponPBToDTO(item))
	}
	return &types.PageResult{List: items, Total: resp.Total, Page: int(resp.Page), PageSize: int(resp.PageSize)}, nil
}

// Create 新增优惠券模板。
func (l *CouponLogic) Create(ctx context.Context, req types.CouponSaveRequest, session *model.AdminSession, ip string) (*types.CouponSaveResponse, error) {
	_, err := l.ctx.AdminSvc.CreateCoupon(ctx, couponRequestToPB(0, req, session, ip))
	if err != nil {
		return nil, err
	}
	// adminsvc 当前创建接口保持通用 CommonResponse，未在 RPC 响应中直接返回自增 ID。
	// HTTP 层创建成功后按本次名称回查最新模板，确保前端仍能拿到新建优惠券 ID。
	resp, err := l.ctx.AdminSvc.ListCoupons(ctx, &adminclient.CouponListRequest{
		Page:     1,
		PageSize: 1,
		Keyword:  req.Name,
	})
	if err != nil {
		return nil, err
	}
	if len(resp.List) == 0 {
		return &types.CouponSaveResponse{}, nil
	}
	return &types.CouponSaveResponse{ID: resp.List[0].Id}, nil
}

// Update 编辑优惠券模板。
func (l *CouponLogic) Update(ctx context.Context, id int64, req types.CouponSaveRequest, session *model.AdminSession, ip string) error {
	_, err := l.ctx.AdminSvc.UpdateCoupon(ctx, couponRequestToPB(id, req, session, ip))
	return err
}

// Disable 下架优惠券模板。
func (l *CouponLogic) Disable(ctx context.Context, id int64, session *model.AdminSession, ip string) error {
	adminID := int64(0)
	if session != nil {
		adminID = session.AdminID
	}
	_, err := l.ctx.AdminSvc.DisableCoupon(ctx, &adminclient.CouponRequest{Id: id, AdminId: adminID, Ip: ip})
	return err
}

// Issue 创建优惠券发放任务。
func (l *CouponLogic) Issue(ctx context.Context, couponID int64, req types.CouponIssueRequest, session *model.AdminSession, ip string) (*types.CouponIssueResponse, error) {
	adminID := int64(0)
	if session != nil {
		adminID = session.AdminID
	}
	resp, err := l.ctx.AdminSvc.IssueCoupon(ctx, &adminclient.CouponIssueRequest{
		CouponId:     couponID,
		TargetType:   req.TargetType,
		TargetConfig: req.TargetConfig,
		AdminId:      adminID,
		Ip:           ip,
	})
	if err != nil {
		return nil, err
	}
	return &types.CouponIssueResponse{
		TaskNo:       resp.TaskNo,
		TotalCount:   resp.TotalCount,
		SuccessCount: resp.SuccessCount,
		FailCount:    resp.FailCount,
		Status:       resp.Status,
	}, nil
}

// ListIssueTasks 查询优惠券发放任务。
func (l *CouponLogic) ListIssueTasks(ctx context.Context, req types.CouponListRequest, couponID int64, status string) (*types.PageResult, error) {
	resp, err := l.ctx.AdminSvc.ListCouponIssueTasks(ctx, &adminclient.CouponIssueTaskListRequest{
		Page:      int32(req.Page),
		PageSize:  int32(req.PageSize),
		CouponId:  couponID,
		Status:    status,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
	})
	if err != nil {
		return nil, err
	}
	items := make([]types.CouponIssueTaskDTO, 0, len(resp.List))
	for _, item := range resp.List {
		items = append(items, types.CouponIssueTaskDTO{
			ID:            item.Id,
			TaskNo:        item.TaskNo,
			CouponID:      item.CouponId,
			TargetType:    item.TargetType,
			TargetConfig:  item.TargetConfig,
			TotalCount:    item.TotalCount,
			SuccessCount:  item.SuccessCount,
			FailCount:     item.FailCount,
			Status:        item.Status,
			FailureReason: item.FailureReason,
			OperatorID:    item.OperatorId,
			CreatedAt:     item.CreatedAt,
			UpdatedAt:     item.UpdatedAt,
		})
	}
	return &types.PageResult{List: items, Total: resp.Total, Page: int(resp.Page), PageSize: int(resp.PageSize)}, nil
}

// couponRequestToPB 将 HTTP 请求体转换为 RPC 请求体。
func couponRequestToPB(id int64, req types.CouponSaveRequest, session *model.AdminSession, ip string) *adminclient.CouponRequest {
	adminID := int64(0)
	if session != nil {
		adminID = session.AdminID
	}
	return &adminclient.CouponRequest{
		Id:              id,
		Name:            req.Name,
		Type:            req.Type,
		FaceValue:       req.FaceValue,
		Discount:        req.Discount,
		ThresholdAmount: req.ThresholdAmount,
		TotalCount:      req.TotalCount,
		PerUserLimit:    int32(req.PerUserLimit),
		ValidStartAt:    req.ValidStartAt,
		ValidEndAt:      req.ValidEndAt,
		Status:          req.Status,
		AdminId:         adminID,
		Ip:              ip,
	}
}

// couponPBToDTO 将 RPC 优惠券对象转换为 HTTP DTO。
func couponPBToDTO(item *adminclient.Coupon) types.CouponDTO {
	return types.CouponDTO{
		ID:              item.Id,
		Name:            item.Name,
		Type:            item.Type,
		FaceValue:       item.FaceValue,
		Discount:        item.Discount,
		ThresholdAmount: item.ThresholdAmount,
		TotalCount:      item.TotalCount,
		ReceivedCount:   item.ReceivedCount,
		PerUserLimit:    int64(item.PerUserLimit),
		ValidStartAt:    item.ValidStartAt,
		ValidEndAt:      item.ValidEndAt,
		Status:          item.Status,
		CreatedAt:       item.CreatedAt,
		UpdatedAt:       item.UpdatedAt,
	}
}
