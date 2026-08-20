package logic

import (
	"context"

	"XiaoLong-Ridy/api/admin/internal/model"
	"XiaoLong-Ridy/api/admin/internal/svc"
	"XiaoLong-Ridy/api/admin/internal/types"
	adminclient "XiaoLong-Ridy/rpc/adminsvc/client/adminservice"
)

// PriceRuleLogic 负责后台计价规则 HTTP 请求到 adminsvc 的参数转换。
type PriceRuleLogic struct {
	ctx *svc.ServiceContext
}

// NewPriceRuleLogic 创建计价规则逻辑对象。
func NewPriceRuleLogic(ctx *svc.ServiceContext) *PriceRuleLogic {
	return &PriceRuleLogic{ctx: ctx}
}

// List 查询计价规则列表。
func (l *PriceRuleLogic) List(ctx context.Context, req types.PriceRuleListRequest) (*types.PageResult, error) {
	resp, err := l.ctx.AdminSvc.ListPriceRules(ctx, &adminclient.PriceRuleListRequest{
		Page:     int32(req.Page),
		PageSize: int32(req.PageSize),
		Keyword:  req.Keyword,
		CityCode: req.CityCode,
		CarType:  req.CarType,
		Status:   req.Status,
	})
	if err != nil {
		return nil, err
	}
	items := make([]types.PriceRuleDTO, 0, len(resp.List))
	for _, item := range resp.List {
		items = append(items, priceRulePBToDTO(item))
	}
	return &types.PageResult{List: items, Total: resp.Total, Page: int(resp.Page), PageSize: int(resp.PageSize)}, nil
}

// Detail 查询计价规则详情。
func (l *PriceRuleLogic) Detail(ctx context.Context, id int64) (*types.PriceRuleDTO, error) {
	resp, err := l.ctx.AdminSvc.GetPriceRule(ctx, &adminclient.PriceRuleDetailRequest{Id: id})
	if err != nil {
		return nil, err
	}
	dto := priceRulePBToDTO(resp)
	return &dto, nil
}

// Create 创建计价规则。
func (l *PriceRuleLogic) Create(ctx context.Context, req types.PriceRuleSaveRequest, session *model.AdminSession, ip string) error {
	_, err := l.ctx.AdminSvc.CreatePriceRule(ctx, priceRuleRequestToPB(0, req, session, ip))
	return err
}

// Update 更新计价规则。
func (l *PriceRuleLogic) Update(ctx context.Context, id int64, req types.PriceRuleSaveRequest, session *model.AdminSession, ip string) error {
	_, err := l.ctx.AdminSvc.UpdatePriceRule(ctx, priceRuleRequestToPB(id, req, session, ip))
	return err
}

// Enable 启用计价规则。
func (l *PriceRuleLogic) Enable(ctx context.Context, id int64, session *model.AdminSession, ip string) error {
	_, err := l.ctx.AdminSvc.EnablePriceRule(ctx, priceRuleStatusRequestToPB(id, 1, session, ip))
	return err
}

// Disable 停用计价规则。
func (l *PriceRuleLogic) Disable(ctx context.Context, id int64, session *model.AdminSession, ip string) error {
	_, err := l.ctx.AdminSvc.DisablePriceRule(ctx, priceRuleStatusRequestToPB(id, 2, session, ip))
	return err
}

// priceRuleRequestToPB 将 HTTP 请求转换为 adminsvc 请求。
func priceRuleRequestToPB(id int64, req types.PriceRuleSaveRequest, session *model.AdminSession, ip string) *adminclient.PriceRuleRequest {
	adminID := int64(0)
	if session != nil {
		adminID = session.AdminID
	}
	return &adminclient.PriceRuleRequest{
		Id:               id,
		Name:             req.Name,
		CityCode:         req.CityCode,
		CarType:          req.CarType,
		BasePrice:        req.BasePrice,
		BaseDistanceKm:   req.BaseDistanceKm,
		PerKmPrice:       req.PerKmPrice,
		PerMinutePrice:   req.PerMinutePrice,
		NightStartTime:   req.NightStartTime,
		NightEndTime:     req.NightEndTime,
		NightSurcharge:   req.NightSurcharge,
		DynamicMaxFactor: req.DynamicMaxFactor,
		Status:           req.Status,
		EffectiveAt:      req.EffectiveAt,
		ExpireAt:         req.ExpireAt,
		AdminId:          adminID,
		Ip:               ip,
	}
}

// priceRuleStatusRequestToPB 将状态修改请求转换为 adminsvc 请求。
func priceRuleStatusRequestToPB(id int64, status int32, session *model.AdminSession, ip string) *adminclient.PriceRuleStatusRequest {
	adminID := int64(0)
	if session != nil {
		adminID = session.AdminID
	}
	return &adminclient.PriceRuleStatusRequest{
		Id:      id,
		Status:  status,
		AdminId: adminID,
		Ip:      ip,
	}
}

// priceRulePBToDTO 将 RPC 对象转换为 HTTP DTO。
func priceRulePBToDTO(item *adminclient.PriceRule) types.PriceRuleDTO {
	if item == nil {
		return types.PriceRuleDTO{}
	}
	return types.PriceRuleDTO{
		ID:               item.Id,
		Name:             item.Name,
		CityCode:         item.CityCode,
		CarType:          item.CarType,
		BasePrice:        item.BasePrice,
		BaseDistanceKm:   item.BaseDistanceKm,
		PerKmPrice:       item.PerKmPrice,
		PerMinutePrice:   item.PerMinutePrice,
		NightStartTime:   item.NightStartTime,
		NightEndTime:     item.NightEndTime,
		NightSurcharge:   item.NightSurcharge,
		DynamicMaxFactor: item.DynamicMaxFactor,
		Status:           item.Status,
		EffectiveAt:      item.EffectiveAt,
		ExpireAt:         item.ExpireAt,
		CreatedAt:        item.CreatedAt,
		UpdatedAt:        item.UpdatedAt,
	}
}
