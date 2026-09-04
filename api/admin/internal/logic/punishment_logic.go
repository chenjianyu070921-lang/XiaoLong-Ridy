package logic

import (
	"context"

	"XiaoLong-Ridy/api/admin/internal/model"
	"XiaoLong-Ridy/api/admin/internal/svc"
	"XiaoLong-Ridy/api/admin/internal/types"
	adminclient "XiaoLong-Ridy/rpc/adminsvc/client/adminservice"
)

// PunishmentLogic 负责处罚域 HTTP 请求与 adminsvc RPC 之间的参数适配。
// 该层不承载处罚业务规则，所有状态变更、权限和幂等校验由 adminsvc 完成。
type PunishmentLogic struct {
	ctx *svc.ServiceContext
}

// NewPunishmentLogic 创建处罚域 HTTP 逻辑对象。
func NewPunishmentLogic(ctx *svc.ServiceContext) *PunishmentLogic {
	return &PunishmentLogic{ctx: ctx}
}

// ListRules 查询处罚规则分页列表。
func (l *PunishmentLogic) ListRules(ctx context.Context, req types.DriverPunishmentRuleListRequest) (*types.PageResult, error) {
	resp, err := l.ctx.AdminSvc.ListDriverPunishmentRules(ctx, &adminclient.DriverPunishmentRuleListRequest{
		Page: int32(req.Page), PageSize: int32(req.PageSize), Keyword: req.Keyword,
		ViolationType: req.ViolationType, Status: req.Status,
	})
	if err != nil {
		return nil, err
	}
	items := make([]types.DriverPunishmentRuleDTO, 0, len(resp.GetList()))
	for _, item := range resp.GetList() {
		items = append(items, types.DriverPunishmentRuleDTO{
			ID: item.GetId(), Name: item.GetName(), ViolationType: item.GetViolationType(),
			Actions: item.GetActions(), PenaltyCents: item.GetPenaltyCents(), ScoreDelta: item.GetScoreDelta(),
			PriorityWeightDelta: item.GetPriorityWeightDelta(), Status: item.GetStatus(), Version: item.GetVersion(),
			CreatedBy: item.GetCreatedBy(), UpdatedBy: item.GetUpdatedBy(), CreatedAt: item.GetCreatedAt(), UpdatedAt: item.GetUpdatedAt(),
		})
	}
	return &types.PageResult{List: items, Total: resp.GetTotal(), Page: int(resp.GetPage()), PageSize: int(resp.GetPageSize())}, nil
}

// SaveRule 新增或编辑处罚规则。
func (l *PunishmentLogic) SaveRule(ctx context.Context, id int64, req types.DriverPunishmentRuleSaveRequest, session *model.AdminSession, ip string) (any, error) {
	adminID := session.AdminID
	in := &adminclient.DriverPunishmentRuleRequest{
		Id: id, Name: req.Name, ViolationType: req.ViolationType, Actions: req.Actions,
		PenaltyCents: req.PenaltyCents, ScoreDelta: req.ScoreDelta, PriorityWeightDelta: req.PriorityWeightDelta,
		Status: req.Status, AdminId: adminID, Ip: ip,
	}
	if id > 0 {
		return l.ctx.AdminSvc.UpdateDriverPunishmentRule(ctx, in)
	}
	return l.ctx.AdminSvc.CreateDriverPunishmentRule(ctx, in)
}

// SetRuleStatus 修改处罚规则启停状态。
func (l *PunishmentLogic) SetRuleStatus(ctx context.Context, id int64, req types.DriverPunishmentRuleStatusRequest, session *model.AdminSession, ip string) error {
	_, err := l.ctx.AdminSvc.SetDriverPunishmentRuleStatus(ctx, &adminclient.DriverPunishmentRuleStatusRequest{
		Id: id, Status: req.Status, AdminId: session.AdminID, Ip: ip,
	})
	return err
}

// ListPunishments 查询处罚单分页列表。
func (l *PunishmentLogic) ListPunishments(ctx context.Context, req types.DriverPunishmentListRequest) (*types.PageResult, error) {
	resp, err := l.ctx.AdminSvc.ListDriverPunishments(ctx, &adminclient.DriverPunishmentListRequest{
		Page: int32(req.Page), PageSize: int32(req.PageSize), DriverId: req.DriverID,
		OrderId: req.OrderID, Status: req.Status, RequestId: req.RequestID,
	})
	if err != nil {
		return nil, err
	}
	items := make([]types.DriverPunishmentDTO, 0, len(resp.GetList()))
	for _, item := range resp.GetList() {
		items = append(items, types.DriverPunishmentDTO{
			ID: item.GetId(), PunishmentNo: item.GetPunishmentNo(), DriverID: item.GetDriverId(), OrderID: item.GetOrderId(),
			RuleID: item.GetRuleId(), Actions: item.GetActions(), Reason: item.GetReason(), PenaltyCents: item.GetPenaltyCents(),
			Status: item.GetStatus(), RequestID: item.GetRequestId(), FailureReason: item.GetFailureReason(), CreatedBy: item.GetCreatedBy(),
			CancelledBy: item.GetCancelledBy(), CancelledAt: item.GetCancelledAt(), EffectiveAt: item.GetEffectiveAt(),
			CreatedAt: item.GetCreatedAt(), UpdatedAt: item.GetUpdatedAt(),
		})
	}
	return &types.PageResult{List: items, Total: resp.GetTotal(), Page: int(resp.GetPage()), PageSize: int(resp.GetPageSize())}, nil
}

// GetPunishment 查询处罚单详情。
func (l *PunishmentLogic) GetPunishment(ctx context.Context, id int64) (*types.DriverPunishmentDTO, error) {
	item, err := l.ctx.AdminSvc.GetDriverPunishment(ctx, &adminclient.DriverPunishmentDetailRequest{Id: id})
	if err != nil {
		return nil, err
	}
	return &types.DriverPunishmentDTO{
		ID: item.GetId(), PunishmentNo: item.GetPunishmentNo(), DriverID: item.GetDriverId(), OrderID: item.GetOrderId(),
		RuleID: item.GetRuleId(), Actions: item.GetActions(), Reason: item.GetReason(), PenaltyCents: item.GetPenaltyCents(),
		Status: item.GetStatus(), RequestID: item.GetRequestId(), FailureReason: item.GetFailureReason(), CreatedBy: item.GetCreatedBy(),
		CancelledBy: item.GetCancelledBy(), CancelledAt: item.GetCancelledAt(), EffectiveAt: item.GetEffectiveAt(),
		CreatedAt: item.GetCreatedAt(), UpdatedAt: item.GetUpdatedAt(),
	}, nil
}

// CreatePunishment 创建处罚单。
func (l *PunishmentLogic) CreatePunishment(ctx context.Context, req types.DriverPunishmentCreateRequest, session *model.AdminSession, ip string) (*types.DriverPunishmentDTO, error) {
	item, err := l.ctx.AdminSvc.CreateDriverPunishment(ctx, &adminclient.DriverPunishmentRequest{
		DriverId: req.DriverID, OrderId: req.OrderID, RuleId: req.RuleID, Actions: req.Actions, Reason: req.Reason,
		PenaltyCents: req.PenaltyCents, RequestId: req.RequestID, AdminId: session.AdminID, Ip: ip,
	})
	if err != nil {
		return nil, err
	}
	return l.toPunishment(item), nil
}

// CancelPunishment 撤销未生效处罚单。
func (l *PunishmentLogic) CancelPunishment(ctx context.Context, id int64, req types.DriverPunishmentActionRequest, session *model.AdminSession, ip string) error {
	_, err := l.ctx.AdminSvc.CancelDriverPunishment(ctx, &adminclient.DriverPunishmentActionRequest{
		Id: id, Reason: req.Reason, RequestId: req.RequestID, AdminId: session.AdminID, Ip: ip,
	})
	return err
}

// ListAppeals 查询处罚申诉。
func (l *PunishmentLogic) ListAppeals(ctx context.Context, req types.PunishmentAppealListRequest) (*types.PageResult, error) {
	resp, err := l.ctx.AdminSvc.ListPunishmentAppeals(ctx, &adminclient.PunishmentAppealListRequest{
		Page: int32(req.Page), PageSize: int32(req.PageSize), PunishmentId: req.PunishmentID,
		DriverId: req.DriverID, Status: req.Status,
	})
	if err != nil {
		return nil, err
	}
	items := make([]types.PunishmentAppealDTO, 0, len(resp.GetList()))
	for _, item := range resp.GetList() {
		items = append(items, l.toAppeal(item))
	}
	return &types.PageResult{List: items, Total: resp.GetTotal(), Page: int(resp.GetPage()), PageSize: int(resp.GetPageSize())}, nil
}

// CreateAppeal 创建司机处罚申诉入口，供司机端或客服代录使用。
func (l *PunishmentLogic) CreateAppeal(ctx context.Context, req struct {
	PunishmentID   int64  `json:"punishment_id"`
	DriverID       int64  `json:"driver_id"`
	Content        string `json:"content"`
	EvidenceConfig string `json:"evidence_config"`
	RequestID      string `json:"request_id"`
}) (*types.PunishmentAppealDTO, error) {
	item, err := l.ctx.AdminSvc.CreatePunishmentAppeal(ctx, &adminclient.PunishmentAppealRequest{
		PunishmentId: req.PunishmentID, DriverId: req.DriverID, Content: req.Content,
		EvidenceConfig: req.EvidenceConfig, RequestId: req.RequestID,
	})
	if err != nil {
		return nil, err
	}
	appeal := l.toAppeal(item)
	return &appeal, nil
}

// ReviewAppeal 审核处罚申诉。
func (l *PunishmentLogic) ReviewAppeal(ctx context.Context, id int64, req types.PunishmentAppealReviewRequest, session *model.AdminSession, ip string) error {
	_, err := l.ctx.AdminSvc.ReviewPunishmentAppeal(ctx, &adminclient.PunishmentAppealReviewRequest{
		Id: id, Action: req.Action, ReviewResult: req.ReviewResult, RequestId: req.RequestID,
		AdminId: session.AdminID, Ip: ip,
	})
	return err
}

func (l *PunishmentLogic) toPunishment(item *adminclient.DriverPunishment) *types.DriverPunishmentDTO {
	if item == nil {
		return &types.DriverPunishmentDTO{}
	}
	return &types.DriverPunishmentDTO{
		ID: item.GetId(), PunishmentNo: item.GetPunishmentNo(), DriverID: item.GetDriverId(), OrderID: item.GetOrderId(),
		RuleID: item.GetRuleId(), Actions: item.GetActions(), Reason: item.GetReason(), PenaltyCents: item.GetPenaltyCents(),
		Status: item.GetStatus(), RequestID: item.GetRequestId(), FailureReason: item.GetFailureReason(), CreatedBy: item.GetCreatedBy(),
		CancelledBy: item.GetCancelledBy(), CancelledAt: item.GetCancelledAt(), EffectiveAt: item.GetEffectiveAt(),
		CreatedAt: item.GetCreatedAt(), UpdatedAt: item.GetUpdatedAt(),
	}
}

func (l *PunishmentLogic) toAppeal(item *adminclient.PunishmentAppeal) types.PunishmentAppealDTO {
	if item == nil {
		return types.PunishmentAppealDTO{}
	}
	return types.PunishmentAppealDTO{
		ID: item.GetId(), AppealNo: item.GetAppealNo(), PunishmentID: item.GetPunishmentId(), DriverID: item.GetDriverId(),
		Content: item.GetContent(), EvidenceConfig: item.GetEvidenceConfig(), Status: item.GetStatus(), ReviewResult: item.GetReviewResult(),
		ReviewedBy: item.GetReviewedBy(), ReviewedAt: item.GetReviewedAt(), RequestID: item.GetRequestId(),
		CreatedAt: item.GetCreatedAt(), UpdatedAt: item.GetUpdatedAt(),
	}
}
