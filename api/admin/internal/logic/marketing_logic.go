package logic

import (
	"context"

	"XiaoLong-Ridy/api/admin/internal/model"
	"XiaoLong-Ridy/api/admin/internal/svc"
	"XiaoLong-Ridy/api/admin/internal/types"
	adminclient "XiaoLong-Ridy/rpc/adminsvc/client/adminservice"
)

// MarketingLogic 负责营销活动 HTTP 请求到 adminsvc 的参数转换。
type MarketingLogic struct {
	ctx *svc.ServiceContext
}

// NewMarketingLogic 创建营销活动逻辑对象。
func NewMarketingLogic(ctx *svc.ServiceContext) *MarketingLogic {
	return &MarketingLogic{ctx: ctx}
}

// ListActivities 查询活动配置列表。
func (l *MarketingLogic) ListActivities(ctx context.Context, req types.PromotionActivityListRequest) (*types.PageResult, error) {
	resp, err := l.ctx.AdminSvc.ListPromotionActivities(ctx, &adminclient.PromotionActivityListRequest{
		Page:      int32(req.Page),
		PageSize:  int32(req.PageSize),
		Keyword:   req.Keyword,
		Type:      req.Type,
		Status:    req.Status,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
	})
	if err != nil {
		return nil, err
	}
	items := make([]types.PromotionActivityDTO, 0, len(resp.List))
	for _, item := range resp.List {
		items = append(items, promotionActivityPBToDTO(item))
	}
	return &types.PageResult{List: items, Total: resp.Total, Page: int(resp.Page), PageSize: int(resp.PageSize)}, nil
}

// CreateActivity 新增活动配置。
func (l *MarketingLogic) CreateActivity(ctx context.Context, req types.PromotionActivitySaveRequest, session *model.AdminSession, ip string) error {
	_, err := l.ctx.AdminSvc.CreatePromotionActivity(ctx, promotionActivityRequestToPB(0, req, session, ip))
	return err
}

// UpdateActivity 编辑活动配置。
func (l *MarketingLogic) UpdateActivity(ctx context.Context, id int64, req types.PromotionActivitySaveRequest, session *model.AdminSession, ip string) error {
	_, err := l.ctx.AdminSvc.UpdatePromotionActivity(ctx, promotionActivityRequestToPB(id, req, session, ip))
	return err
}

// PublishActivity 发布活动配置。
func (l *MarketingLogic) PublishActivity(ctx context.Context, id int64, req types.PromotionActivityActionRequest, session *model.AdminSession, ip string) error {
	_, err := l.ctx.AdminSvc.PublishPromotionActivity(ctx, promotionActivityActionToPB(id, req, session, ip))
	return err
}

// RollbackActivity 回滚活动配置。
func (l *MarketingLogic) RollbackActivity(ctx context.Context, id int64, req types.PromotionActivityActionRequest, session *model.AdminSession, ip string) error {
	_, err := l.ctx.AdminSvc.RollbackPromotionActivity(ctx, promotionActivityActionToPB(id, req, session, ip))
	return err
}

// promotionActivityRequestToPB 将 HTTP 活动请求转换为 RPC 请求。
func promotionActivityRequestToPB(id int64, req types.PromotionActivitySaveRequest, session *model.AdminSession, ip string) *adminclient.PromotionActivityRequest {
	adminID := int64(0)
	if session != nil {
		adminID = session.AdminID
	}
	return &adminclient.PromotionActivityRequest{
		Id:      id,
		Name:    req.Name,
		Type:    req.Type,
		Config:  req.Config,
		StartAt: req.StartAt,
		EndAt:   req.EndAt,
		Status:  req.Status,
		AdminId: adminID,
		Ip:      ip,
	}
}

// promotionActivityActionToPB 将 HTTP 活动动作请求转换为 RPC 请求。
func promotionActivityActionToPB(id int64, req types.PromotionActivityActionRequest, session *model.AdminSession, ip string) *adminclient.PromotionActivityActionRequest {
	adminID := int64(0)
	if session != nil {
		adminID = session.AdminID
	}
	return &adminclient.PromotionActivityActionRequest{
		Id:           id,
		PublishScope: req.PublishScope,
		TargetConfig: req.TargetConfig,
		AdminId:      adminID,
		Ip:           ip,
	}
}

// promotionActivityPBToDTO 将 RPC 活动对象转换为 HTTP DTO。
func promotionActivityPBToDTO(item *adminclient.PromotionActivity) types.PromotionActivityDTO {
	return types.PromotionActivityDTO{
		ID:        item.Id,
		Name:      item.Name,
		Type:      item.Type,
		Config:    item.Config,
		StartAt:   item.StartAt,
		EndAt:     item.EndAt,
		Status:    item.Status,
		CreatedBy: item.CreatedBy,
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}
}
