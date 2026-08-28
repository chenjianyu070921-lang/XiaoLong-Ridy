package logic

import (
	"context"

	"XiaoLong-Ridy/api/admin/internal/model"
	"XiaoLong-Ridy/api/admin/internal/svc"
	"XiaoLong-Ridy/api/admin/internal/types"
	adminclient "XiaoLong-Ridy/rpc/adminsvc/client/adminservice"
)

// RiskLogic 负责风控黑名单 HTTP 请求到 adminsvc 的参数转换。
type RiskLogic struct {
	ctx *svc.ServiceContext
}

// NewRiskLogic 创建风控逻辑对象。
func NewRiskLogic(ctx *svc.ServiceContext) *RiskLogic {
	return &RiskLogic{ctx: ctx}
}

// ListBlacklists 查询风控黑名单列表。
func (l *RiskLogic) ListBlacklists(ctx context.Context, page, pageSize int, targetType string, targetID int64, status int32) (*types.PageResult, error) {
	resp, err := l.ctx.AdminSvc.ListBlacklists(ctx, &adminclient.BlacklistListRequest{
		Page:       int32(page),
		PageSize:   int32(pageSize),
		TargetType: targetType,
		TargetId:   targetID,
		Status:     status,
	})
	if err != nil {
		return nil, err
	}
	items := make([]types.BlacklistDTO, 0, len(resp.List))
	for _, item := range resp.List {
		items = append(items, types.BlacklistDTO{
			ID:         item.Id,
			TargetType: item.TargetType,
			TargetID:   item.TargetId,
			Reason:     item.Reason,
			OperatorID: item.OperatorId,
			Status:     item.Status,
			CreatedAt:  item.CreatedAt,
			UpdatedAt:  item.UpdatedAt,
		})
	}
	return &types.PageResult{List: items, Total: resp.Total, Page: int(resp.Page), PageSize: int(resp.PageSize)}, nil
}

// AddBlacklist 新增风控黑名单。
func (l *RiskLogic) AddBlacklist(ctx context.Context, req types.BlacklistRequest, session *model.AdminSession, ip string) error {
	_, err := l.ctx.AdminSvc.AddBlacklist(ctx, blacklistRequestToPB(0, req, session, ip))
	return err
}

// ReleaseBlacklist 解除风控黑名单。
func (l *RiskLogic) ReleaseBlacklist(ctx context.Context, id int64, req types.BlacklistRequest, session *model.AdminSession, ip string) error {
	_, err := l.ctx.AdminSvc.ReleaseBlacklist(ctx, blacklistRequestToPB(id, req, session, ip))
	return err
}

// ListRiskHitRecords 查询风控命中记录。
func (l *RiskLogic) ListRiskHitRecords(ctx context.Context, page, pageSize int, targetType string, targetID int64, scene string, riskLevel int32) (*types.PageResult, error) {
	resp, err := l.ctx.AdminSvc.ListRiskHitRecords(ctx, &adminclient.RiskHitRecordListRequest{
		Page:       int32(page),
		PageSize:   int32(pageSize),
		TargetType: targetType,
		TargetId:   targetID,
		Scene:      scene,
		RiskLevel:  riskLevel,
	})
	if err != nil {
		return nil, err
	}
	items := make([]types.RiskHitRecordDTO, 0, len(resp.List))
	for _, item := range resp.List {
		items = append(items, types.RiskHitRecordDTO{
			ID:           item.Id,
			BlacklistID:  item.BlacklistId,
			TargetType:   item.TargetType,
			TargetID:     item.TargetId,
			Scene:        item.Scene,
			RiskLevel:    item.RiskLevel,
			HitReason:    item.HitReason,
			RequestID:    item.RequestId,
			CreatedAt:    item.CreatedAt,
			HandleStatus: item.HandleStatus,
			HandleAction: item.HandleAction,
			HandledBy:    item.HandledBy,
			HandledAt:    item.HandledAt,
			WorkOrderID:  item.WorkOrderId,
		})
	}
	return &types.PageResult{List: items, Total: resp.Total, Page: int(resp.Page), PageSize: int(resp.PageSize)}, nil
}

// HandleRiskHitRecords 处置风控命中记录，支持复核通过、加入黑名单和转工单。
func (l *RiskLogic) HandleRiskHitRecords(ctx context.Context, req types.RiskHitActionRequest, session *model.AdminSession, ip string) (*adminclient.RiskHitActionResponse, error) {
	adminID := int64(0)
	if session != nil {
		adminID = session.AdminID
	}
	return l.ctx.AdminSvc.HandleRiskHitRecords(ctx, &adminclient.RiskHitActionRequest{
		Ids:            req.IDs,
		Action:         req.Action,
		Reason:         req.Reason,
		WorkOrderTitle: req.WorkOrderTitle,
		Priority:       req.Priority,
		AdminId:        adminID,
		Ip:             ip,
	})
}

// blacklistRequestToPB 将 HTTP 黑名单请求转换为 RPC 请求。
func blacklistRequestToPB(id int64, req types.BlacklistRequest, session *model.AdminSession, ip string) *adminclient.BlacklistRequest {
	adminID := int64(0)
	if session != nil {
		adminID = session.AdminID
	}
	return &adminclient.BlacklistRequest{
		Id:         id,
		TargetType: req.TargetType,
		TargetId:   req.TargetID,
		Reason:     req.Reason,
		AdminId:    adminID,
		Ip:         ip,
	}
}
