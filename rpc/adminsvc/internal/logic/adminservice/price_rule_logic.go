package adminservicelogic

import (
	"context"
	"fmt"

	"XiaoLong-Ridy/rpc/adminsvc/adminsvc"
	"XiaoLong-Ridy/rpc/adminsvc/internal/svc"
	priceclient "XiaoLong-Ridy/rpc/pricesvc/price"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ensurePriceService 用于本地最小服务集模式下校验 pricesvc 客户端是否已初始化。
// 入参为管理 RPC 服务上下文，返回 nil 表示可安全调用计价下游；否则返回可识别的 Unavailable 错误，
// 防止未启动 pricesvc 时对空客户端解引用并导致整个 adminsvc 进程退出。
func ensurePriceService(svcCtx *svc.ServiceContext) error {
	if svcCtx == nil || svcCtx.PricesSvc == nil {
		return status.Error(codes.Unavailable, "计价服务未启动")
	}
	return nil
}

// ListPriceRulesLogic 处理计价规则列表查询。
type ListPriceRulesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewListPriceRulesLogic 创建计价规则列表逻辑对象。
func NewListPriceRulesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListPriceRulesLogic {
	return &ListPriceRulesLogic{ctx: ctx, svcCtx: svcCtx}
}

// ListPriceRules 查询计价规则列表。
func (l *ListPriceRulesLogic) ListPriceRules(in *adminsvc.PriceRuleListRequest) (*adminsvc.PriceRuleListResponse, error) {
	if err := ensurePriceService(l.svcCtx); err != nil {
		return nil, err
	}
	resp, err := l.svcCtx.PricesSvc.ListPriceRules(l.ctx, &priceclient.PriceRuleListRequest{
		Page:     in.GetPage(),
		PageSize: in.GetPageSize(),
		Keyword:  in.GetKeyword(),
		CityCode: in.GetCityCode(),
		CarType:  in.GetCarType(),
		Status:   in.GetStatus(),
	})
	if err != nil {
		return nil, err
	}
	items := make([]*adminsvc.PriceRule, 0, len(resp.List))
	for _, item := range resp.List {
		items = append(items, priceRulePBToAdminPB(item))
	}
	return &adminsvc.PriceRuleListResponse{
		List:     items,
		Total:    resp.GetTotal(),
		Page:     resp.GetPage(),
		PageSize: resp.GetPageSize(),
	}, nil
}

// GetPriceRuleLogic 处理计价规则详情查询。
type GetPriceRuleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGetPriceRuleLogic 创建计价规则详情逻辑对象。
func NewGetPriceRuleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPriceRuleLogic {
	return &GetPriceRuleLogic{ctx: ctx, svcCtx: svcCtx}
}

// GetPriceRule 查询计价规则详情。
func (l *GetPriceRuleLogic) GetPriceRule(in *adminsvc.PriceRuleDetailRequest) (*adminsvc.PriceRule, error) {
	if err := ensurePriceService(l.svcCtx); err != nil {
		return nil, err
	}
	resp, err := l.svcCtx.PricesSvc.GetPriceRule(l.ctx, &priceclient.PriceRuleDetailRequest{Id: in.GetId()})
	if err != nil {
		return nil, err
	}
	return priceRulePBToAdminPB(resp), nil
}

// CreatePriceRuleLogic 处理计价规则创建。
type CreatePriceRuleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewCreatePriceRuleLogic 创建计价规则创建逻辑对象。
func NewCreatePriceRuleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreatePriceRuleLogic {
	return &CreatePriceRuleLogic{ctx: ctx, svcCtx: svcCtx}
}

// CreatePriceRule 创建计价规则并写入操作日志。
func (l *CreatePriceRuleLogic) CreatePriceRule(in *adminsvc.PriceRuleRequest) (*adminsvc.CreatePriceRuleResponse, error) {
	if err := ensurePriceService(l.svcCtx); err != nil {
		return nil, err
	}
	resp, err := l.svcCtx.PricesSvc.CreatePriceRule(l.ctx, priceRuleRequestToPB(in))
	if err != nil {
		return nil, err
	}
	detail := fmt.Sprintf("创建计价规则：%s", in.GetName())
	if err := writeAuditAfterCommitted(l.ctx, l.svcCtx, in.GetAdminId(), "price", "create", "price_rule", resp.GetId(), detail, in.GetIp()); err != nil {
		return nil, err
	}
	return &adminsvc.CreatePriceRuleResponse{Id: resp.GetId(), Message: "ok"}, nil
}

// UpdatePriceRuleLogic 处理计价规则更新。
type UpdatePriceRuleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewUpdatePriceRuleLogic 创建计价规则更新逻辑对象。
func NewUpdatePriceRuleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdatePriceRuleLogic {
	return &UpdatePriceRuleLogic{ctx: ctx, svcCtx: svcCtx}
}

// UpdatePriceRule 更新计价规则并写入操作日志。
func (l *UpdatePriceRuleLogic) UpdatePriceRule(in *adminsvc.PriceRuleRequest) (*adminsvc.CommonResponse, error) {
	if err := ensurePriceService(l.svcCtx); err != nil {
		return nil, err
	}
	if _, err := l.svcCtx.PricesSvc.UpdatePriceRule(l.ctx, priceRuleRequestToPB(in)); err != nil {
		return nil, err
	}
	if err := writeAuditAfterCommitted(l.ctx, l.svcCtx, in.GetAdminId(), "price", "update", "price_rule", in.GetId(), fmt.Sprintf("编辑计价规则：%s", in.GetName()), in.GetIp()); err != nil {
		return nil, err
	}
	return &adminsvc.CommonResponse{Message: "ok"}, nil
}

// EnablePriceRuleLogic 处理计价规则启用。
type EnablePriceRuleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewEnablePriceRuleLogic 创建计价规则启用逻辑对象。
func NewEnablePriceRuleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *EnablePriceRuleLogic {
	return &EnablePriceRuleLogic{ctx: ctx, svcCtx: svcCtx}
}

// EnablePriceRule 启用计价规则并写入操作日志。
func (l *EnablePriceRuleLogic) EnablePriceRule(in *adminsvc.PriceRuleStatusRequest) (*adminsvc.CommonResponse, error) {
	if err := ensurePriceService(l.svcCtx); err != nil {
		return nil, err
	}
	if _, err := l.svcCtx.PricesSvc.SetPriceRuleStatus(l.ctx, &priceclient.PriceRuleStatusRequest{Id: in.GetId(), Status: 1}); err != nil {
		return nil, err
	}
	if err := writeAuditAfterCommitted(l.ctx, l.svcCtx, in.GetAdminId(), "price", "enable", "price_rule", in.GetId(), "启用计价规则", in.GetIp()); err != nil {
		return nil, err
	}
	return &adminsvc.CommonResponse{Message: "ok"}, nil
}

// DisablePriceRuleLogic 处理计价规则停用。
type DisablePriceRuleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewDisablePriceRuleLogic 创建计价规则停用逻辑对象。
func NewDisablePriceRuleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DisablePriceRuleLogic {
	return &DisablePriceRuleLogic{ctx: ctx, svcCtx: svcCtx}
}

// DisablePriceRule 停用计价规则并写入操作日志。
func (l *DisablePriceRuleLogic) DisablePriceRule(in *adminsvc.PriceRuleStatusRequest) (*adminsvc.CommonResponse, error) {
	if err := ensurePriceService(l.svcCtx); err != nil {
		return nil, err
	}
	if _, err := l.svcCtx.PricesSvc.SetPriceRuleStatus(l.ctx, &priceclient.PriceRuleStatusRequest{Id: in.GetId(), Status: 2}); err != nil {
		return nil, err
	}
	if err := writeAuditAfterCommitted(l.ctx, l.svcCtx, in.GetAdminId(), "price", "disable", "price_rule", in.GetId(), "停用计价规则", in.GetIp()); err != nil {
		return nil, err
	}
	return &adminsvc.CommonResponse{Message: "ok"}, nil
}

// priceRulePBToAdminPB 将 pricesvc 的计价规则转换为 adminsvc 对象。
func priceRulePBToAdminPB(item *priceclient.PriceRule) *adminsvc.PriceRule {
	if item == nil {
		return nil
	}
	return &adminsvc.PriceRule{
		Id:               item.GetId(),
		Name:             item.GetName(),
		CityCode:         item.GetCityCode(),
		CarType:          item.GetCarType(),
		BasePrice:        item.GetBasePrice(),
		BaseDistanceKm:   item.GetBaseDistanceKm(),
		PerKmPrice:       item.GetPerKmPrice(),
		PerMinutePrice:   item.GetPerMinutePrice(),
		NightStartTime:   item.GetNightStartTime(),
		NightEndTime:     item.GetNightEndTime(),
		NightSurcharge:   item.GetNightSurcharge(),
		DynamicMaxFactor: item.GetDynamicMaxFactor(),
		Status:           item.GetStatus(),
		EffectiveAt:      item.GetEffectiveAt(),
		ExpireAt:         item.GetExpireAt(),
		CreatedAt:        item.GetCreatedAt(),
		UpdatedAt:        item.GetUpdatedAt(),
	}
}

// priceRuleRequestToPB 将 adminsvc 请求转换为 pricesvc 请求。
func priceRuleRequestToPB(in *adminsvc.PriceRuleRequest) *priceclient.PriceRuleRequest {
	if in == nil {
		return nil
	}
	return &priceclient.PriceRuleRequest{
		Id:               in.GetId(),
		Name:             in.GetName(),
		CityCode:         in.GetCityCode(),
		CarType:          in.GetCarType(),
		BasePrice:        in.GetBasePrice(),
		BaseDistanceKm:   in.GetBaseDistanceKm(),
		PerKmPrice:       in.GetPerKmPrice(),
		PerMinutePrice:   in.GetPerMinutePrice(),
		NightStartTime:   in.GetNightStartTime(),
		NightEndTime:     in.GetNightEndTime(),
		NightSurcharge:   in.GetNightSurcharge(),
		DynamicMaxFactor: in.GetDynamicMaxFactor(),
		Status:           in.GetStatus(),
		EffectiveAt:      in.GetEffectiveAt(),
		ExpireAt:         in.GetExpireAt(),
	}
}
