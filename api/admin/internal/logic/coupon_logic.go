package logic

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"XiaoLong-Ridy/api/admin/internal/model"
	"XiaoLong-Ridy/api/admin/internal/repository"
	"XiaoLong-Ridy/api/admin/internal/svc"
	"XiaoLong-Ridy/api/admin/internal/types"
)

const adminTimeLayout = "2006-01-02 15:04:05"

// CouponLogic 封装管理后台优惠券配置业务逻辑。
// 它负责参数校验、调用仓储读写 coupon 表，并为新增和编辑动作写入后台操作日志。
type CouponLogic struct {
	ctx *svc.ServiceContext
}

// NewCouponLogic 创建优惠券业务逻辑对象。
// handler 层每次请求创建轻量实例，共用 ServiceContext 中的连接池和仓储。
func NewCouponLogic(ctx *svc.ServiceContext) *CouponLogic {
	return &CouponLogic{ctx: ctx}
}

// List 查询优惠券模板分页列表。
// 该接口只读取模板配置，不读取用户券明细，避免列表页查询过重。
func (l *CouponLogic) List(ctx context.Context, req types.CouponListRequest) (*types.PageResult, error) {
	list, total, err := l.ctx.CouponRepository.List(ctx, req)
	if err != nil {
		return nil, err
	}
	items := make([]types.CouponDTO, 0, len(list))
	for _, item := range list {
		items = append(items, toCouponDTO(item))
	}
	return &types.PageResult{
		List:     items,
		Total:    total,
		Page:     normalizePage(req.Page),
		PageSize: normalizePageSize(req.PageSize),
	}, nil
}

// Create 新增优惠券模板。
// 创建成功后写入 admin_operation_log，便于运营配置变更追溯。
func (l *CouponLogic) Create(ctx context.Context, req types.CouponSaveRequest, session *model.AdminSession, ip string) (*types.CouponSaveResponse, error) {
	if err := validateCouponSaveRequest(req); err != nil {
		return nil, err
	}
	id, err := l.ctx.CouponRepository.Create(ctx, req)
	if err != nil {
		return nil, err
	}
	if err := l.ctx.OperationLogRepository.Create(ctx, repository.CreateOperationLogInput{
		AdminID:    session.AdminID,
		Module:     "coupon",
		Action:     "create",
		TargetType: "coupon",
		TargetID:   id,
		Detail:     fmt.Sprintf("创建优惠券模板：%s", req.Name),
		IP:         ip,
	}); err != nil {
		return nil, err
	}
	return &types.CouponSaveResponse{ID: id}, nil
}

// Update 编辑优惠券模板。
// 编辑接口不直接操作 user_coupon，防止后台配置修改影响已经发放给用户的券实例。
func (l *CouponLogic) Update(ctx context.Context, id int64, req types.CouponSaveRequest, session *model.AdminSession, ip string) error {
	if id <= 0 {
		return ErrBadRequest
	}
	if err := validateCouponSaveRequest(req); err != nil {
		return err
	}
	if err := l.ctx.CouponRepository.Update(ctx, id, req); err != nil {
		return err
	}
	return l.ctx.OperationLogRepository.Create(ctx, repository.CreateOperationLogInput{
		AdminID:    session.AdminID,
		Module:     "coupon",
		Action:     "update",
		TargetType: "coupon",
		TargetID:   id,
		Detail:     fmt.Sprintf("编辑优惠券模板：%s", req.Name),
		IP:         ip,
	})
}

// validateCouponSaveRequest 校验优惠券新增和编辑请求。
// 这里做基础业务约束校验，复杂的活动冲突检测后续应交给计价/营销 RPC 服务处理。
func validateCouponSaveRequest(req types.CouponSaveRequest) error {
	if strings.TrimSpace(req.Name) == "" {
		return ErrBadRequest
	}
	if req.Type < 1 || req.Type > 3 {
		return ErrBadRequest
	}
	if req.Status < 1 || req.Status > 2 {
		return ErrBadRequest
	}
	if req.TotalCount < 0 || req.PerUserLimit <= 0 {
		return ErrBadRequest
	}
	if !validDecimal(req.FaceValue) || !validDecimal(req.Discount) || !validDecimal(req.ThresholdAmount) {
		return ErrBadRequest
	}
	start, err := time.ParseInLocation(adminTimeLayout, req.ValidStartAt, time.Local)
	if err != nil {
		return ErrBadRequest
	}
	end, err := time.ParseInLocation(adminTimeLayout, req.ValidEndAt, time.Local)
	if err != nil {
		return ErrBadRequest
	}
	if !end.After(start) {
		return ErrBadRequest
	}
	return nil
}

// validDecimal 校验金额和折扣字段是否为合法十进制字符串。
// 业务层只校验格式和非负，实际精度由 MySQL DECIMAL 字段兜底约束。
func validDecimal(value string) bool {
	if strings.TrimSpace(value) == "" {
		return false
	}
	parsed, err := strconv.ParseFloat(value, 64)
	return err == nil && parsed >= 0
}

// toCouponDTO 将优惠券数据库模型转换为接口返回对象。
// 时间统一格式化为 yyyy-MM-dd HH:mm:ss，便于前端直接展示。
func toCouponDTO(item model.Coupon) types.CouponDTO {
	return types.CouponDTO{
		ID:              item.ID,
		Name:            item.Name,
		Type:            item.Type,
		FaceValue:       item.FaceValue,
		Discount:        item.Discount,
		ThresholdAmount: item.ThresholdAmount,
		TotalCount:      item.TotalCount,
		ReceivedCount:   item.ReceivedCount,
		PerUserLimit:    item.PerUserLimit,
		ValidStartAt:    repository.FormatTime(item.ValidStartAt),
		ValidEndAt:      repository.FormatTime(item.ValidEndAt),
		Status:          item.Status,
		CreatedAt:       repository.FormatTime(item.CreatedAt),
		UpdatedAt:       repository.FormatTime(item.UpdatedAt),
	}
}
