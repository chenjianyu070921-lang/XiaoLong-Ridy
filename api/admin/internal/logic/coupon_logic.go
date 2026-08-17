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
	return &types.CouponSaveResponse{}, nil
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
