package adminservicelogic

import (
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"XiaoLong-Ridy/rpc/adminsvc/adminsvc"
	"XiaoLong-Ridy/rpc/adminsvc/internal/svc"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// couponIssueTargetConfig 表示后台发券任务的目标用户配置。
// P1 当前只落地显式用户 ID 发放，后续 crowd 人群包可由 MQ/Job 异步扩展。
type couponIssueTargetConfig struct {
	UserIDs []int64 `json:"user_ids"`
}

// orderExportFilters 是订单 CSV 唯一允许的筛选字段集合。
type orderExportFilters struct {
	Status    *int32 `json:"status,omitempty"`
	StartTime string `json:"start_time,omitempty"`
	EndTime   string `json:"end_time,omitempty"`
	UserID    *int64 `json:"user_id,omitempty"`
	DriverID  *int64 `json:"driver_id,omitempty"`
	raw       string
}

// parseOrderExportFilters 校验订单导出筛选条件，拒绝城市条件和未知字段。
func parseOrderExportFilters(exportType, raw string) (orderExportFilters, error) {
	if exportType != "orders" {
		return orderExportFilters{}, status.Error(codes.InvalidArgument, "当前仅支持订单导出")
	}
	filters := orderExportFilters{raw: raw}
	if strings.TrimSpace(raw) == "" {
		return filters, nil
	}
	var values map[string]json.RawMessage
	if err := json.Unmarshal([]byte(raw), &values); err != nil {
		return orderExportFilters{}, status.Error(codes.InvalidArgument, "导出筛选条件必须是JSON对象")
	}
	for key, value := range values {
		switch key {
		case "status":
			var statusCode int32
			if err := json.Unmarshal(value, &statusCode); err != nil || statusCode <= 0 {
				return orderExportFilters{}, status.Error(codes.InvalidArgument, "status必须是正整数")
			}
			filters.Status = &statusCode
		case "user_id", "driver_id":
			var id int64
			if err := json.Unmarshal(value, &id); err != nil || id <= 0 {
				return orderExportFilters{}, status.Error(codes.InvalidArgument, key+"必须是正整数")
			}
			if key == "user_id" {
				filters.UserID = &id
			} else {
				filters.DriverID = &id
			}
		case "start_time", "end_time":
			var timeText string
			if err := json.Unmarshal(value, &timeText); err != nil {
				return orderExportFilters{}, status.Error(codes.InvalidArgument, key+"必须是时间字符串")
			}
			if _, err := time.ParseInLocation(couponTimeLayout, timeText, time.Local); err != nil {
				return orderExportFilters{}, status.Error(codes.InvalidArgument, key+"格式错误")
			}
			if key == "start_time" {
				filters.StartTime = timeText
			} else {
				filters.EndTime = timeText
			}
		case "city_code":
			return orderExportFilters{}, status.Error(codes.InvalidArgument, "当前订单数据未保存城市编码，暂不支持城市筛选")
		default:
			return orderExportFilters{}, status.Error(codes.InvalidArgument, "不支持的导出筛选字段："+key)
		}
	}
	if filters.StartTime != "" && filters.EndTime != "" && filters.EndTime < filters.StartTime {
		return orderExportFilters{}, status.Error(codes.InvalidArgument, "结束时间不能早于开始时间")
	}
	return filters, nil
}

// where 返回订单导出 SQL 的参数化筛选条件。
func (f orderExportFilters) where() (string, []any) {
	parts := []string{"deleted_at IS NULL"}
	args := make([]any, 0, 5)
	if f.Status != nil {
		parts = append(parts, "status = ?")
		args = append(args, *f.Status)
	}
	if f.UserID != nil {
		parts = append(parts, "user_id = ?")
		args = append(args, *f.UserID)
	}
	if f.DriverID != nil {
		parts = append(parts, "driver_id = ?")
		args = append(args, *f.DriverID)
	}
	if f.StartTime != "" {
		parts = append(parts, "created_at >= ?")
		args = append(args, f.StartTime)
	}
	if f.EndTime != "" {
		parts = append(parts, "created_at <= ?")
		args = append(args, f.EndTime)
	}
	return "WHERE " + strings.Join(parts, " AND "), args
}

// IssueCouponLogic 处理优惠券发放任务创建和同步发券。
type IssueCouponLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewIssueCouponLogic 创建优惠券发放逻辑对象。
func NewIssueCouponLogic(ctx context.Context, svcCtx *svc.ServiceContext) *IssueCouponLogic {
	return &IssueCouponLogic{ctx: ctx, svcCtx: svcCtx}
}

// IssueCoupon 创建发券任务，并在当前请求内完成 user_coupon 写入。
func (l *IssueCouponLogic) IssueCoupon(in *adminsvc.CouponIssueRequest) (*adminsvc.CouponIssueResponse, error) {
	if in.GetCouponId() <= 0 || in.GetAdminId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "优惠券ID和管理员ID不能为空")
	}
	cfg, err := parseCouponIssueTarget(in.GetTargetType(), in.GetTargetConfig())
	if err != nil {
		return nil, err
	}
	taskNo := newAdminTaskNo("CI")
	tx, err := l.svcCtx.MySQL.BeginTx(l.ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	var validEndAt time.Time
	var couponStatus int32
	var totalCount, receivedCount, perUserLimit int64
	err = tx.QueryRowContext(l.ctx, `
		SELECT valid_end_at, status, total_count, received_count, per_user_limit
		FROM coupon
		WHERE id = ?
		FOR UPDATE
	`, in.GetCouponId()).Scan(&validEndAt, &couponStatus, &totalCount, &receivedCount, &perUserLimit)
	if err == sql.ErrNoRows {
		return nil, status.Error(codes.NotFound, "优惠券不存在")
	}
	if err != nil {
		return nil, err
	}
	if couponStatus != 1 {
		return nil, status.Error(codes.FailedPrecondition, "优惠券未启用，不能发放")
	}
	if time.Now().After(validEndAt) {
		return nil, status.Error(codes.FailedPrecondition, "优惠券已过期，不能发放")
	}

	successCount, failCount := int64(0), int64(0)
	for _, userID := range cfg.UserIDs {
		if userID <= 0 {
			failCount++
			continue
		}
		if totalCount > 0 && receivedCount+successCount >= totalCount {
			failCount++
			continue
		}
		var userCouponCount int64
		if err := tx.QueryRowContext(l.ctx, `SELECT COUNT(1) FROM user_coupon WHERE user_id = ? AND coupon_id = ?`, userID, in.GetCouponId()).Scan(&userCouponCount); err != nil {
			return nil, err
		}
		if perUserLimit > 0 && userCouponCount >= perUserLimit {
			failCount++
			continue
		}
		if _, err := tx.ExecContext(l.ctx, `
			INSERT INTO user_coupon (user_id, coupon_id, order_id, status, received_at, expire_at)
			VALUES (?, ?, 0, 1, ?, ?)
		`, userID, in.GetCouponId(), time.Now(), validEndAt); err != nil {
			return nil, err
		}
		successCount++
	}
	taskStatus := int32(3)
	failureReason := ""
	if successCount == 0 {
		taskStatus = 5
		failureReason = "全部用户发放失败，请检查用户ID、库存和单用户领取上限"
	} else if failCount > 0 {
		taskStatus = 4
		failureReason = "部分用户因参数、库存或领取上限发放失败"
	}
	if _, err := tx.ExecContext(l.ctx, `
		INSERT INTO admin_coupon_issue_task
			(task_no, coupon_id, target_type, target_config, total_count, success_count, fail_count, status, failure_reason, operator_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, taskNo, in.GetCouponId(), in.GetTargetType(), in.GetTargetConfig(), len(cfg.UserIDs), successCount, failCount, taskStatus, failureReason, in.GetAdminId()); err != nil {
		return nil, err
	}
	if _, err := tx.ExecContext(l.ctx, `UPDATE coupon SET received_count = received_count + ? WHERE id = ?`, successCount, in.GetCouponId()); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	if err := createOperationLog(l.ctx, l.svcCtx, in.GetAdminId(), "coupon", "issue", "coupon", in.GetCouponId(), fmt.Sprintf("创建发券任务：%s，成功%d，失败%d", taskNo, successCount, failCount), in.GetIp()); err != nil {
		return nil, err
	}
	return &adminsvc.CouponIssueResponse{
		TaskNo:       taskNo,
		TotalCount:   int64(len(cfg.UserIDs)),
		SuccessCount: successCount,
		FailCount:    failCount,
		Status:       couponIssueStatusText(taskStatus),
	}, nil
}

// ListCouponIssueTasksLogic 处理发券任务列表查询。
type ListCouponIssueTasksLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewListCouponIssueTasksLogic 创建发券任务列表逻辑对象。
func NewListCouponIssueTasksLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListCouponIssueTasksLogic {
	return &ListCouponIssueTasksLogic{ctx: ctx, svcCtx: svcCtx}
}

// ListCouponIssueTasks 查询后台发券任务列表。
func (l *ListCouponIssueTasksLogic) ListCouponIssueTasks(in *adminsvc.CouponIssueTaskListRequest) (*adminsvc.CouponIssueTaskListResponse, error) {
	where, args := buildIssueTaskWhere(in)
	var total int64
	if err := l.svcCtx.MySQL.QueryRowContext(l.ctx, `SELECT COUNT(1) FROM admin_coupon_issue_task `+where, args...).Scan(&total); err != nil {
		return nil, err
	}
	limit := normalizePageSize(in.GetPageSize())
	queryArgs := append(args, limit, offset(in.GetPage(), in.GetPageSize()))
	rows, err := l.svcCtx.MySQL.QueryContext(l.ctx, `
		SELECT id, task_no, coupon_id, target_type, target_config, total_count, success_count,
		       fail_count, status, failure_reason, operator_id, created_at, updated_at
		FROM admin_coupon_issue_task `+where+`
		ORDER BY id DESC
		LIMIT ? OFFSET ?
	`, queryArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*adminsvc.CouponIssueTask, 0)
	for rows.Next() {
		item, err := scanIssueTask(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, item)
	}
	return &adminsvc.CouponIssueTaskListResponse{List: list, Total: total, Page: normalizePage(in.GetPage()), PageSize: limit}, rows.Err()
}

// ListPromotionActivitiesLogic 处理活动配置列表查询。
type ListPromotionActivitiesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewListPromotionActivitiesLogic 创建活动配置列表逻辑对象。
func NewListPromotionActivitiesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListPromotionActivitiesLogic {
	return &ListPromotionActivitiesLogic{ctx: ctx, svcCtx: svcCtx}
}

// ListPromotionActivities 查询活动配置列表。
func (l *ListPromotionActivitiesLogic) ListPromotionActivities(in *adminsvc.PromotionActivityListRequest) (*adminsvc.PromotionActivityListResponse, error) {
	where, args := buildPromotionWhere(in)
	var total int64
	if err := l.svcCtx.MySQL.QueryRowContext(l.ctx, `SELECT COUNT(1) FROM promotion_activity `+where, args...).Scan(&total); err != nil {
		return nil, err
	}
	limit := normalizePageSize(in.GetPageSize())
	queryArgs := append(args, limit, offset(in.GetPage(), in.GetPageSize()))
	rows, err := l.svcCtx.MySQL.QueryContext(l.ctx, `
		SELECT id, name, type, config, start_at, end_at, status, created_by, created_at, updated_at
		FROM promotion_activity `+where+`
		ORDER BY id DESC
		LIMIT ? OFFSET ?
	`, queryArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*adminsvc.PromotionActivity, 0)
	for rows.Next() {
		item, err := scanPromotionActivity(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, item)
	}
	return &adminsvc.PromotionActivityListResponse{List: list, Total: total, Page: normalizePage(in.GetPage()), PageSize: limit}, rows.Err()
}

// CreatePromotionActivityLogic 处理活动配置创建。
type CreatePromotionActivityLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewCreatePromotionActivityLogic 创建活动配置新增逻辑对象。
func NewCreatePromotionActivityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreatePromotionActivityLogic {
	return &CreatePromotionActivityLogic{ctx: ctx, svcCtx: svcCtx}
}

// CreatePromotionActivity 新增活动配置草稿或待开始活动。
func (l *CreatePromotionActivityLogic) CreatePromotionActivity(in *adminsvc.PromotionActivityRequest) (*adminsvc.CommonResponse, error) {
	if err := validatePromotionActivity(in); err != nil {
		return nil, err
	}
	res, err := l.svcCtx.MySQL.ExecContext(l.ctx, `
		INSERT INTO promotion_activity (name, type, config, start_at, end_at, status, created_by)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, in.GetName(), in.GetType(), in.GetConfig(), in.GetStartAt(), in.GetEndAt(), in.GetStatus(), in.GetAdminId())
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	if err := createOperationLog(l.ctx, l.svcCtx, in.GetAdminId(), "promotion", "create", "promotion_activity", id, fmt.Sprintf("创建活动配置：%s", in.GetName()), in.GetIp()); err != nil {
		return nil, err
	}
	return &adminsvc.CommonResponse{Message: "ok"}, nil
}

// UpdatePromotionActivityLogic 处理活动配置更新。
type UpdatePromotionActivityLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewUpdatePromotionActivityLogic 创建活动配置更新逻辑对象。
func NewUpdatePromotionActivityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdatePromotionActivityLogic {
	return &UpdatePromotionActivityLogic{ctx: ctx, svcCtx: svcCtx}
}

// UpdatePromotionActivity 更新活动配置。
func (l *UpdatePromotionActivityLogic) UpdatePromotionActivity(in *adminsvc.PromotionActivityRequest) (*adminsvc.CommonResponse, error) {
	if in.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "活动ID不能为空")
	}
	if err := validatePromotionActivity(in); err != nil {
		return nil, err
	}
	res, err := l.svcCtx.MySQL.ExecContext(l.ctx, `
		UPDATE promotion_activity
		SET name = ?, type = ?, config = ?, start_at = ?, end_at = ?, status = ?
		WHERE id = ?
	`, in.GetName(), in.GetType(), in.GetConfig(), in.GetStartAt(), in.GetEndAt(), in.GetStatus(), in.GetId())
	if err != nil {
		return nil, err
	}
	if affected, _ := res.RowsAffected(); affected == 0 {
		return nil, status.Error(codes.NotFound, "活动配置不存在")
	}
	if err := createOperationLog(l.ctx, l.svcCtx, in.GetAdminId(), "promotion", "update", "promotion_activity", in.GetId(), fmt.Sprintf("编辑活动配置：%s", in.GetName()), in.GetIp()); err != nil {
		return nil, err
	}
	return &adminsvc.CommonResponse{Message: "ok"}, nil
}

// PublishPromotionActivityLogic 处理活动发布。
type PublishPromotionActivityLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewPublishPromotionActivityLogic 创建活动发布逻辑对象。
func NewPublishPromotionActivityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PublishPromotionActivityLogic {
	return &PublishPromotionActivityLogic{ctx: ctx, svcCtx: svcCtx}
}

// PublishPromotionActivity 将活动状态置为运行中，并写入操作日志。
func (l *PublishPromotionActivityLogic) PublishPromotionActivity(in *adminsvc.PromotionActivityActionRequest) (*adminsvc.CommonResponse, error) {
	if err := validatePromotionAction(in, true); err != nil {
		return nil, err
	}
	if err := changePromotionStatus(l.ctx, l.svcCtx, in.GetId(), 1, 2); err != nil {
		return nil, err
	}
	if err := writeAuditAfterCommitted(l.ctx, l.svcCtx, in.GetAdminId(), "promotion", "publish", "promotion_activity", in.GetId(), fmt.Sprintf("发布活动，范围：%s，配置：%s", in.GetPublishScope(), in.GetTargetConfig()), in.GetIp()); err != nil {
		return nil, err
	}
	return &adminsvc.CommonResponse{Message: "ok"}, nil
}

// RollbackPromotionActivityLogic 处理活动回滚。
type RollbackPromotionActivityLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewRollbackPromotionActivityLogic 创建活动回滚逻辑对象。
func NewRollbackPromotionActivityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RollbackPromotionActivityLogic {
	return &RollbackPromotionActivityLogic{ctx: ctx, svcCtx: svcCtx}
}

// RollbackPromotionActivity 将活动回滚为未开始状态，并写入操作日志。
func (l *RollbackPromotionActivityLogic) RollbackPromotionActivity(in *adminsvc.PromotionActivityActionRequest) (*adminsvc.CommonResponse, error) {
	if err := validatePromotionAction(in, false); err != nil {
		return nil, err
	}
	if err := changePromotionStatus(l.ctx, l.svcCtx, in.GetId(), 2, 1); err != nil {
		return nil, err
	}
	if err := writeAuditAfterCommitted(l.ctx, l.svcCtx, in.GetAdminId(), "promotion", "rollback", "promotion_activity", in.GetId(), fmt.Sprintf("回滚活动，配置：%s", in.GetTargetConfig()), in.GetIp()); err != nil {
		return nil, err
	}
	return &adminsvc.CommonResponse{Message: "ok"}, nil
}

// GetStatisticsOverviewLogic 处理运营总览统计。
type GetStatisticsOverviewLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGetStatisticsOverviewLogic 创建运营总览统计逻辑对象。
func NewGetStatisticsOverviewLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetStatisticsOverviewLogic {
	return &GetStatisticsOverviewLogic{ctx: ctx, svcCtx: svcCtx}
}

// GetStatisticsOverview 汇总用户、司机、订单、GMV、优惠券和黑名单指标。
func (l *GetStatisticsOverviewLogic) GetStatisticsOverview(in *adminsvc.StatisticsRequest) (*adminsvc.StatisticsOverviewResponse, error) {
	if strings.TrimSpace(in.GetCityCode()) != "" {
		return nil, status.Error(codes.InvalidArgument, "当前订单数据未保存城市编码，暂不支持城市维度统计")
	}
	orderWhere, orderArgs := buildCreatedAtWhere("created_at", in.GetStartTime(), in.GetEndTime())
	resp := &adminsvc.StatisticsOverviewResponse{}
	var err error
	if resp.UserCount, err = countSQL(l.ctx, l.svcCtx.MySQL, `SELECT COUNT(1) FROM user WHERE deleted_at IS NULL`); err != nil {
		return nil, err
	}
	if resp.DriverCount, err = countSQL(l.ctx, l.svcCtx.MySQL, `SELECT COUNT(1) FROM driver WHERE deleted_at IS NULL`); err != nil {
		return nil, err
	}
	if resp.OrderCount, err = countSQL(l.ctx, l.svcCtx.MySQL, `SELECT COUNT(1) FROM ride_order `+orderWhere, orderArgs...); err != nil {
		return nil, err
	}
	if resp.CompletedOrderCount, err = countSQL(l.ctx, l.svcCtx.MySQL, `SELECT COUNT(1) FROM ride_order `+appendWhere(orderWhere, "status = 5"), orderArgs...); err != nil {
		return nil, err
	}
	if resp.AbnormalOrderCount, err = countSQL(l.ctx, l.svcCtx.MySQL, `SELECT COUNT(1) FROM ride_order `+appendWhere(orderWhere, "status = 6"), orderArgs...); err != nil {
		return nil, err
	}
	if resp.CouponIssueCount, err = countSQL(l.ctx, l.svcCtx.MySQL, `SELECT COUNT(1) FROM user_coupon`); err != nil {
		return nil, err
	}
	if resp.BlacklistCount, err = countSQL(l.ctx, l.svcCtx.MySQL, `SELECT COUNT(1) FROM blacklist WHERE status = 1`); err != nil {
		return nil, err
	}
	gmvWhere := appendWhere(replaceWhereColumn(orderWhere, "o.created_at"), "p.status = 2")
	if resp.Gmv, err = sumSQL(l.ctx, l.svcCtx.MySQL, `SELECT COALESCE(SUM(p.amount), 0) FROM payment p JOIN ride_order o ON o.id = p.order_id `+gmvWhere, orderArgs...); err != nil {
		return nil, err
	}
	return resp, nil
}

// GetOrderStatisticsLogic 处理订单统计。
type GetOrderStatisticsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGetOrderStatisticsLogic 创建订单统计逻辑对象。
func NewGetOrderStatisticsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOrderStatisticsLogic {
	return &GetOrderStatisticsLogic{ctx: ctx, svcCtx: svcCtx}
}

// GetOrderStatistics 统计订单完成、取消、超时和支付异常指标。
func (l *GetOrderStatisticsLogic) GetOrderStatistics(in *adminsvc.StatisticsRequest) (*adminsvc.OrderStatisticsResponse, error) {
	if strings.TrimSpace(in.GetCityCode()) != "" {
		return nil, status.Error(codes.InvalidArgument, "当前订单数据未保存城市编码，暂不支持城市维度统计")
	}
	where, args := buildCreatedAtWhere("created_at", in.GetStartTime(), in.GetEndTime())
	resp := &adminsvc.OrderStatisticsResponse{}
	var err error
	if resp.OrderCount, err = countSQL(l.ctx, l.svcCtx.MySQL, `SELECT COUNT(1) FROM ride_order `+where, args...); err != nil {
		return nil, err
	}
	if resp.CompletedOrderCount, err = countSQL(l.ctx, l.svcCtx.MySQL, `SELECT COUNT(1) FROM ride_order `+appendWhere(where, "status = 5"), args...); err != nil {
		return nil, err
	}
	if resp.CanceledOrderCount, err = countSQL(l.ctx, l.svcCtx.MySQL, `SELECT COUNT(1) FROM ride_order `+appendWhere(where, "status = 6"), args...); err != nil {
		return nil, err
	}
	joinedWhere := replaceWhereColumn(where, "o.created_at")
	if resp.TimeoutOrderCount, err = countSQL(l.ctx, l.svcCtx.MySQL, `SELECT COUNT(DISTINCT o.id) FROM ride_order o JOIN dispatch_record d ON d.order_id = o.id `+appendWhere(joinedWhere, "d.status = 4"), args...); err != nil {
		return nil, err
	}
	if resp.PaymentAbnormalCount, err = countSQL(l.ctx, l.svcCtx.MySQL, `SELECT COUNT(DISTINCT o.id) FROM ride_order o JOIN payment p ON p.order_id = o.id `+appendWhere(joinedWhere, "p.status = 3"), args...); err != nil {
		return nil, err
	}
	resp.CompletionRate = percentText(resp.CompletedOrderCount, resp.OrderCount)
	resp.CancelRate = percentText(resp.CanceledOrderCount, resp.OrderCount)
	return resp, nil
}

// GetCouponStatisticsLogic 处理优惠券统计。
type GetCouponStatisticsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGetCouponStatisticsLogic 创建优惠券统计逻辑对象。
func NewGetCouponStatisticsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCouponStatisticsLogic {
	return &GetCouponStatisticsLogic{ctx: ctx, svcCtx: svcCtx}
}

// GetCouponStatistics 统计优惠券模板、发放、使用和过期指标。
func (l *GetCouponStatisticsLogic) GetCouponStatistics(in *adminsvc.StatisticsRequest) (*adminsvc.CouponStatisticsResponse, error) {
	resp := &adminsvc.CouponStatisticsResponse{}
	var err error
	if resp.CouponCount, err = countSQL(l.ctx, l.svcCtx.MySQL, `SELECT COUNT(1) FROM coupon`); err != nil {
		return nil, err
	}
	if resp.EnabledCouponCount, err = countSQL(l.ctx, l.svcCtx.MySQL, `SELECT COUNT(1) FROM coupon WHERE status = 1`); err != nil {
		return nil, err
	}
	if resp.IssuedCouponCount, err = countSQL(l.ctx, l.svcCtx.MySQL, `SELECT COUNT(1) FROM user_coupon`); err != nil {
		return nil, err
	}
	if resp.UsedCouponCount, err = countSQL(l.ctx, l.svcCtx.MySQL, `SELECT COUNT(1) FROM user_coupon WHERE status = 2`); err != nil {
		return nil, err
	}
	if resp.ExpiredCouponCount, err = countSQL(l.ctx, l.svcCtx.MySQL, `SELECT COUNT(1) FROM user_coupon WHERE status = 3 OR (status = 1 AND expire_at < NOW())`); err != nil {
		return nil, err
	}
	resp.UseRate = percentText(resp.UsedCouponCount, resp.IssuedCouponCount)
	return resp, nil
}

// CreateExportTaskLogic 处理导出任务创建。
type CreateExportTaskLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewCreateExportTaskLogic 创建导出任务逻辑对象。
func NewCreateExportTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateExportTaskLogic {
	return &CreateExportTaskLogic{ctx: ctx, svcCtx: svcCtx}
}

// CreateExportTask 创建导出任务并启动后台 goroutine 异步生成文件。
// 任务状态写入 admin_export_task，admin_operation_log 只保留审计用途，避免再用日志 detail 承载状态机。
func (l *CreateExportTaskLogic) CreateExportTask(in *adminsvc.ExportTaskRequest) (*adminsvc.ExportTaskResponse, error) {
	exportType := strings.TrimSpace(in.GetExportType())
	if exportType == "" || in.GetAdminId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "导出类型和管理员ID不能为空")
	}
	filters, err := parseOrderExportFilters(exportType, in.GetFilters())
	if err != nil {
		return nil, err
	}
	if !isSupportedExportType(exportType) {
		return nil, status.Error(codes.InvalidArgument, "导出类型暂不支持")
	}
	taskNo := newAdminTaskNo("EX")
	if _, err := l.svcCtx.MySQL.ExecContext(l.ctx, `
		INSERT INTO admin_export_task (task_no, export_type, filters, status, admin_id, ip)
		VALUES (?, ?, ?, 'pending', ?, ?)
	`, taskNo, exportType, nullableJSON(filters.raw), in.GetAdminId(), in.GetIp()); err != nil {
		return nil, err
	}
	_ = createOperationLog(l.ctx, l.svcCtx, in.GetAdminId(), "export", "create", exportType, 0, fmt.Sprintf("创建导出任务：%s", taskNo), in.GetIp())
	go runExportTaskJob(l.svcCtx, taskNo)
	return &adminsvc.ExportTaskResponse{TaskNo: taskNo, Status: "pending", Message: "导出任务已创建，后台正在异步生成文件"}, nil
}

// ListExportTasksLogic 处理导出任务列表查询。
type ListExportTasksLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewListExportTasksLogic 创建导出任务列表逻辑对象。
func NewListExportTasksLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListExportTasksLogic {
	return &ListExportTasksLogic{ctx: ctx, svcCtx: svcCtx}
}

// ListExportTasks 从 admin_export_task 独立任务表读取导出状态列表。
func (l *ListExportTasksLogic) ListExportTasks(in *adminsvc.ExportTaskListRequest) (*adminsvc.ExportTaskListResponse, error) {
	where := "WHERE 1=1"
	args := make([]any, 0)
	if in.GetExportType() != "" {
		where += " AND export_type = ?"
		args = append(args, in.GetExportType())
	}
	var total int64
	if err := l.svcCtx.MySQL.QueryRowContext(l.ctx, `SELECT COUNT(1) FROM admin_export_task `+where, args...).Scan(&total); err != nil {
		return nil, err
	}
	limit := normalizePageSize(in.GetPageSize())
	queryArgs := append(args, limit, offset(in.GetPage(), in.GetPageSize()))
	rows, err := l.svcCtx.MySQL.QueryContext(l.ctx, `
		SELECT task_no, export_type, COALESCE(CAST(filters AS CHAR), ''), status, admin_id,
		       file_path, file_url, failure_reason, created_at, updated_at, expires_at
		FROM admin_export_task `+where+`
		ORDER BY id DESC
		LIMIT ? OFFSET ?
	`, queryArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*adminsvc.ExportTask, 0)
	for rows.Next() {
		item, err := scanExportTask(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, item)
	}
	return &adminsvc.ExportTaskListResponse{List: list, Total: total, Page: normalizePage(in.GetPage()), PageSize: limit}, rows.Err()
}

// GetExportTaskLogic 处理导出任务详情查询。
type GetExportTaskLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGetExportTaskLogic 创建导出任务详情逻辑对象。
func NewGetExportTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetExportTaskLogic {
	return &GetExportTaskLogic{ctx: ctx, svcCtx: svcCtx}
}

// GetExportTask 按任务编号查询导出任务当前状态。
func (l *GetExportTaskLogic) GetExportTask(in *adminsvc.ExportTaskDetailRequest) (*adminsvc.ExportTask, error) {
	taskNo := strings.TrimSpace(in.GetTaskNo())
	if taskNo == "" {
		return nil, status.Error(codes.InvalidArgument, "导出任务编号不能为空")
	}
	row := l.svcCtx.MySQL.QueryRowContext(l.ctx, `
		SELECT task_no, export_type, COALESCE(CAST(filters AS CHAR), ''), status, admin_id,
		       file_path, file_url, failure_reason, created_at, updated_at, expires_at
		FROM admin_export_task
		WHERE task_no = ?
	`, taskNo)
	return scanExportTaskRow(row)
}

// ListBlacklistsLogic 处理黑名单列表查询。
type ListBlacklistsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewListBlacklistsLogic 创建黑名单列表逻辑对象。
func NewListBlacklistsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListBlacklistsLogic {
	return &ListBlacklistsLogic{ctx: ctx, svcCtx: svcCtx}
}

// ListBlacklists 查询风控黑名单列表。
func (l *ListBlacklistsLogic) ListBlacklists(in *adminsvc.BlacklistListRequest) (*adminsvc.BlacklistListResponse, error) {
	where, args := buildBlacklistWhere(in)
	var total int64
	if err := l.svcCtx.MySQL.QueryRowContext(l.ctx, `SELECT COUNT(1) FROM blacklist `+where, args...).Scan(&total); err != nil {
		return nil, err
	}
	limit := normalizePageSize(in.GetPageSize())
	queryArgs := append(args, limit, offset(in.GetPage(), in.GetPageSize()))
	rows, err := l.svcCtx.MySQL.QueryContext(l.ctx, `
		SELECT id, target_type, target_id, reason, operator_id, status, created_at, updated_at
		FROM blacklist `+where+`
		ORDER BY id DESC
		LIMIT ? OFFSET ?
	`, queryArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*adminsvc.Blacklist, 0)
	for rows.Next() {
		item, err := scanBlacklist(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, item)
	}
	return &adminsvc.BlacklistListResponse{List: list, Total: total, Page: normalizePage(in.GetPage()), PageSize: limit}, rows.Err()
}

// AddBlacklistLogic 处理黑名单新增。
type AddBlacklistLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewAddBlacklistLogic 创建黑名单新增逻辑对象。
func NewAddBlacklistLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddBlacklistLogic {
	return &AddBlacklistLogic{ctx: ctx, svcCtx: svcCtx}
}

// AddBlacklist 新增或重新激活风控黑名单。
func (l *AddBlacklistLogic) AddBlacklist(in *adminsvc.BlacklistRequest) (*adminsvc.CommonResponse, error) {
	if err := validateBlacklistRequest(in, false); err != nil {
		return nil, err
	}
	res, err := l.svcCtx.MySQL.ExecContext(l.ctx, `
		INSERT INTO blacklist (target_type, target_id, reason, operator_id, status)
		VALUES (?, ?, ?, ?, 1)
	`, in.GetTargetType(), in.GetTargetId(), in.GetReason(), in.GetAdminId())
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	if err := createOperationLog(l.ctx, l.svcCtx, in.GetAdminId(), "risk", "add_blacklist", "blacklist", id, fmt.Sprintf("新增黑名单：%s/%d", in.GetTargetType(), in.GetTargetId()), in.GetIp()); err != nil {
		return nil, err
	}
	return &adminsvc.CommonResponse{Message: "ok"}, nil
}

// ReleaseBlacklistLogic 处理黑名单解除。
type ReleaseBlacklistLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewReleaseBlacklistLogic 创建黑名单解除逻辑对象。
func NewReleaseBlacklistLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReleaseBlacklistLogic {
	return &ReleaseBlacklistLogic{ctx: ctx, svcCtx: svcCtx}
}

// ReleaseBlacklist 将黑名单状态置为已解除。
func (l *ReleaseBlacklistLogic) ReleaseBlacklist(in *adminsvc.BlacklistRequest) (*adminsvc.CommonResponse, error) {
	if err := validateBlacklistRequest(in, true); err != nil {
		return nil, err
	}
	res, err := l.svcCtx.MySQL.ExecContext(l.ctx, `UPDATE blacklist SET status = 2, reason = ? WHERE id = ? AND status = 1`, in.GetReason(), in.GetId())
	if err != nil {
		return nil, err
	}
	if affected, _ := res.RowsAffected(); affected == 0 {
		return nil, status.Error(codes.NotFound, "黑名单不存在或已解除")
	}
	if err := createOperationLog(l.ctx, l.svcCtx, in.GetAdminId(), "risk", "release_blacklist", "blacklist", in.GetId(), "解除黑名单", in.GetIp()); err != nil {
		return nil, err
	}
	return &adminsvc.CommonResponse{Message: "ok"}, nil
}

// ListRiskHitRecordsLogic 处理风控命中记录查询。
type ListRiskHitRecordsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewListRiskHitRecordsLogic 创建风控命中记录查询逻辑对象。
func NewListRiskHitRecordsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListRiskHitRecordsLogic {
	return &ListRiskHitRecordsLogic{ctx: ctx, svcCtx: svcCtx}
}

// ListRiskHitRecords 查询风控黑名单命中记录。
func (l *ListRiskHitRecordsLogic) ListRiskHitRecords(in *adminsvc.RiskHitRecordListRequest) (*adminsvc.RiskHitRecordListResponse, error) {
	where, args := buildRiskHitWhere(in)
	var total int64
	if err := l.svcCtx.MySQL.QueryRowContext(l.ctx, `SELECT COUNT(1) FROM risk_blacklist_hit_record `+where, args...).Scan(&total); err != nil {
		return nil, err
	}
	limit := normalizePageSize(in.GetPageSize())
	queryArgs := append(args, limit, offset(in.GetPage(), in.GetPageSize()))
	rows, err := l.svcCtx.MySQL.QueryContext(l.ctx, `
		SELECT id, blacklist_id, target_type, target_id, scene, risk_level, hit_reason, request_id, created_at
		FROM risk_blacklist_hit_record `+where+`
		ORDER BY id DESC
		LIMIT ? OFFSET ?
	`, queryArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*adminsvc.RiskHitRecord, 0)
	for rows.Next() {
		item, err := scanRiskHitRecord(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, item)
	}
	return &adminsvc.RiskHitRecordListResponse{List: list, Total: total, Page: normalizePage(in.GetPage()), PageSize: limit}, rows.Err()
}

// parseCouponIssueTarget 解析发券目标配置。
func parseCouponIssueTarget(targetType, targetConfig string) (*couponIssueTargetConfig, error) {
	if targetType != "user" && targetType != "batch" && targetType != "crowd" {
		return nil, status.Error(codes.InvalidArgument, "发券目标类型不合法")
	}
	var cfg couponIssueTargetConfig
	if err := json.Unmarshal([]byte(targetConfig), &cfg); err != nil {
		return nil, status.Error(codes.InvalidArgument, "发券目标配置必须是合法JSON")
	}
	if len(cfg.UserIDs) == 0 {
		return nil, status.Error(codes.InvalidArgument, "发券目标用户不能为空")
	}
	return &cfg, nil
}

// validatePromotionActivity 校验活动配置参数。
func validatePromotionActivity(in *adminsvc.PromotionActivityRequest) error {
	if strings.TrimSpace(in.GetName()) == "" || in.GetAdminId() <= 0 {
		return status.Error(codes.InvalidArgument, "活动名称和管理员ID不能为空")
	}
	if in.GetType() < 1 || in.GetType() > 3 || in.GetStatus() < 1 || in.GetStatus() > 3 {
		return status.Error(codes.InvalidArgument, "活动类型或状态不合法")
	}
	if !json.Valid([]byte(in.GetConfig())) {
		return status.Error(codes.InvalidArgument, "活动配置必须是合法JSON")
	}
	start, err := time.ParseInLocation(couponTimeLayout, in.GetStartAt(), time.Local)
	if err != nil {
		return status.Error(codes.InvalidArgument, "活动开始时间格式不合法")
	}
	end, err := time.ParseInLocation(couponTimeLayout, in.GetEndAt(), time.Local)
	if err != nil {
		return status.Error(codes.InvalidArgument, "活动结束时间格式不合法")
	}
	if !end.After(start) {
		return status.Error(codes.InvalidArgument, "活动结束时间必须晚于开始时间")
	}
	return nil
}

// validatePromotionAction 校验活动发布与回滚的操作参数。
func validatePromotionAction(in *adminsvc.PromotionActivityActionRequest, publish bool) error {
	if in.GetId() <= 0 || in.GetAdminId() <= 0 {
		return status.Error(codes.InvalidArgument, "活动ID和管理员ID不能为空")
	}
	if publish && in.GetPublishScope() != "gray" && in.GetPublishScope() != "full" {
		return status.Error(codes.InvalidArgument, "活动发布范围仅支持gray或full")
	}
	if !json.Valid([]byte(in.GetTargetConfig())) {
		return status.Error(codes.InvalidArgument, "活动目标配置必须是合法JSON")
	}
	return nil
}

// changePromotionStatus 通过源状态限制更新，保证发布和回滚不会跨过活动状态机。
func changePromotionStatus(ctx context.Context, svcCtx *svc.ServiceContext, id int64, expectedStatus, targetStatus int32) error {
	if id <= 0 {
		return status.Error(codes.InvalidArgument, "活动ID不能为空")
	}
	res, err := svcCtx.MySQL.ExecContext(ctx, `UPDATE promotion_activity SET status = ? WHERE id = ? AND status = ?`, targetStatus, id, expectedStatus)
	if err != nil {
		return err
	}
	if affected, _ := res.RowsAffected(); affected == 0 {
		return status.Error(codes.FailedPrecondition, "活动当前状态不允许执行该操作")
	}
	return nil
}

// buildIssueTaskWhere 组装发券任务筛选条件。
func buildIssueTaskWhere(in *adminsvc.CouponIssueTaskListRequest) (string, []any) {
	parts := []string{"1=1"}
	args := make([]any, 0)
	if in.GetCouponId() > 0 {
		parts = append(parts, "coupon_id = ?")
		args = append(args, in.GetCouponId())
	}
	if statusCode, ok := couponIssueStatusCode(in.GetStatus()); ok {
		parts = append(parts, "status = ?")
		args = append(args, statusCode)
	}
	if in.GetStartTime() != "" {
		parts = append(parts, "created_at >= ?")
		args = append(args, in.GetStartTime())
	}
	if in.GetEndTime() != "" {
		parts = append(parts, "created_at <= ?")
		args = append(args, in.GetEndTime())
	}
	return "WHERE " + strings.Join(parts, " AND "), args
}

// buildPromotionWhere 组装活动配置筛选条件。
func buildPromotionWhere(in *adminsvc.PromotionActivityListRequest) (string, []any) {
	parts := []string{"1=1"}
	args := make([]any, 0)
	if in.GetKeyword() != "" {
		parts = append(parts, "name LIKE ?")
		args = append(args, "%"+in.GetKeyword()+"%")
	}
	if in.GetType() > 0 {
		parts = append(parts, "type = ?")
		args = append(args, in.GetType())
	}
	if in.GetStatus() > 0 {
		parts = append(parts, "status = ?")
		args = append(args, in.GetStatus())
	}
	if in.GetStartTime() != "" {
		parts = append(parts, "created_at >= ?")
		args = append(args, in.GetStartTime())
	}
	if in.GetEndTime() != "" {
		parts = append(parts, "created_at <= ?")
		args = append(args, in.GetEndTime())
	}
	return "WHERE " + strings.Join(parts, " AND "), args
}

// buildBlacklistWhere 组装黑名单筛选条件。
func buildBlacklistWhere(in *adminsvc.BlacklistListRequest) (string, []any) {
	parts := []string{"1=1"}
	args := make([]any, 0)
	if in.GetTargetType() != "" {
		parts = append(parts, "target_type = ?")
		args = append(args, in.GetTargetType())
	}
	if in.GetTargetId() > 0 {
		parts = append(parts, "target_id = ?")
		args = append(args, in.GetTargetId())
	}
	if in.GetStatus() > 0 {
		parts = append(parts, "status = ?")
		args = append(args, in.GetStatus())
	}
	return "WHERE " + strings.Join(parts, " AND "), args
}

// buildRiskHitWhere 组装风控命中记录筛选条件。
func buildRiskHitWhere(in *adminsvc.RiskHitRecordListRequest) (string, []any) {
	parts := []string{"1=1"}
	args := make([]any, 0)
	if in.GetTargetType() != "" {
		parts = append(parts, "target_type = ?")
		args = append(args, in.GetTargetType())
	}
	if in.GetTargetId() > 0 {
		parts = append(parts, "target_id = ?")
		args = append(args, in.GetTargetId())
	}
	if in.GetScene() != "" {
		parts = append(parts, "scene = ?")
		args = append(args, in.GetScene())
	}
	if in.GetRiskLevel() > 0 {
		parts = append(parts, "risk_level = ?")
		args = append(args, in.GetRiskLevel())
	}
	return "WHERE " + strings.Join(parts, " AND "), args
}

// buildCreatedAtWhere 组装创建时间范围筛选条件。
func buildCreatedAtWhere(column, start, end string) (string, []any) {
	parts := make([]string, 0, 2)
	args := make([]any, 0, 2)
	if start != "" {
		parts = append(parts, column+" >= ?")
		args = append(args, start)
	}
	if end != "" {
		parts = append(parts, column+" <= ?")
		args = append(args, end)
	}
	if len(parts) == 0 {
		return "", args
	}
	return "WHERE " + strings.Join(parts, " AND "), args
}

// appendWhere 在已有 WHERE 条件后追加一个 AND 条件。
func appendWhere(where, condition string) string {
	if strings.TrimSpace(where) == "" {
		return "WHERE " + condition
	}
	return where + " AND " + condition
}

// replaceWhereColumn 将通用订单时间条件替换为 JOIN 查询使用的表别名列。
func replaceWhereColumn(where, column string) string {
	return strings.ReplaceAll(where, "created_at", column)
}

// scanIssueTask 扫描发券任务行。
func scanIssueTask(rows *sql.Rows) (*adminsvc.CouponIssueTask, error) {
	var item adminsvc.CouponIssueTask
	var statusCode int32
	var createdAt, updatedAt sql.NullTime
	if err := rows.Scan(&item.Id, &item.TaskNo, &item.CouponId, &item.TargetType, &item.TargetConfig, &item.TotalCount,
		&item.SuccessCount, &item.FailCount, &statusCode, &item.FailureReason, &item.OperatorId, &createdAt, &updatedAt); err != nil {
		return nil, err
	}
	item.Status = couponIssueStatusText(statusCode)
	item.CreatedAt = formatNullTime(createdAt)
	item.UpdatedAt = formatNullTime(updatedAt)
	return &item, nil
}

// scanPromotionActivity 扫描活动配置行。
func scanPromotionActivity(rows *sql.Rows) (*adminsvc.PromotionActivity, error) {
	var item adminsvc.PromotionActivity
	var startAt, endAt, createdAt, updatedAt sql.NullTime
	if err := rows.Scan(&item.Id, &item.Name, &item.Type, &item.Config, &startAt, &endAt, &item.Status, &item.CreatedBy, &createdAt, &updatedAt); err != nil {
		return nil, err
	}
	item.StartAt = formatNullTime(startAt)
	item.EndAt = formatNullTime(endAt)
	item.CreatedAt = formatNullTime(createdAt)
	item.UpdatedAt = formatNullTime(updatedAt)
	return &item, nil
}

// scanExportTask 扫描 admin_export_task 列表行。
func scanExportTask(rows *sql.Rows) (*adminsvc.ExportTask, error) {
	var item adminsvc.ExportTask
	var fileURL string
	var createdAt, updatedAt, expiresAt sql.NullTime
	if err := rows.Scan(&item.TaskNo, &item.ExportType, &item.Filters, &item.Status, &item.AdminId,
		&item.FilePath, &fileURL, &item.FailureReason, &createdAt, &updatedAt, &expiresAt); err != nil {
		return nil, err
	}
	item.CreatedAt = formatNullTime(createdAt)
	item.UpdatedAt = formatNullTime(updatedAt)
	return &item, nil
}

// scanExportTaskRow 扫描 admin_export_task 单条任务记录。
func scanExportTaskRow(row *sql.Row) (*adminsvc.ExportTask, error) {
	var item adminsvc.ExportTask
	var fileURL string
	var createdAt, updatedAt, expiresAt sql.NullTime
	err := row.Scan(&item.TaskNo, &item.ExportType, &item.Filters, &item.Status, &item.AdminId,
		&item.FilePath, &fileURL, &item.FailureReason, &createdAt, &updatedAt, &expiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, status.Error(codes.NotFound, "导出任务不存在")
	}
	if err != nil {
		return nil, err
	}
	item.CreatedAt = formatNullTime(createdAt)
	item.UpdatedAt = formatNullTime(updatedAt)
	return &item, nil
}

// updateExportTaskStatus 更新导出任务状态、文件路径和失败原因。
// 该函数只修改 admin_export_task，不再回写 admin_operation_log.detail。
func updateExportTaskStatus(ctx context.Context, svcCtx *svc.ServiceContext, taskNo, statusText, filePath, failureReason string, expiresAt any) error {
	res, err := svcCtx.MySQL.ExecContext(ctx, `
		UPDATE admin_export_task
		SET status = ?, file_path = ?, file_url = ?, failure_reason = ?, expires_at = ?
		WHERE task_no = ?
	`, statusText, filePath, filePath, failureReason, expiresAt, taskNo)
	if err != nil {
		return err
	}
	if affected, _ := res.RowsAffected(); affected == 0 {
		return status.Error(codes.NotFound, "导出任务不存在")
	}
	return nil
}

// runExportTaskJob 在后台执行导出文件生成，并把 running/success/failed 状态写回独立任务表。
func runExportTaskJob(svcCtx *svc.ServiceContext, taskNo string) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	task, err := loadExportTaskByNo(ctx, svcCtx, taskNo)
	if err != nil {
		return
	}
	if err := updateExportTaskStatus(ctx, svcCtx, taskNo, "running", "", "", nil); err != nil {
		return
	}
	filePath, err := writeExportTaskFile(ctx, svcCtx, task)
	if err != nil {
		_ = updateExportTaskStatus(ctx, svcCtx, taskNo, "failed", "", err.Error(), nil)
		return
	}
	_ = updateExportTaskStatus(ctx, svcCtx, taskNo, "success", filePath, "", time.Now().Add(7*24*time.Hour))
}

// writeExportTaskFile 按导出类型分页读取业务表并生成 CSV 文件。
// 当前先覆盖后台核心审查场景：用户、司机认证、订单和操作日志；统计类导出输出实时聚合指标。
func writeExportTaskFile(ctx context.Context, svcCtx *svc.ServiceContext, task *adminsvc.ExportTask) (string, error) {
	dir := filepath.Join(".tmp-admin-exports")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	filePath := filepath.Join(dir, task.GetTaskNo()+".csv")
	file, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	writer := csv.NewWriter(file)
	filters, err := parseOrderExportFilters(task.GetExportType(), task.GetFilters())
	if err != nil {
		return "", err
	}
	if err := writeExportCSV(ctx, svcCtx, writer, task.GetExportType(), filters); err != nil {
		return "", err
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return "", err
	}
	return filePath, nil
}

// scanBlacklist 扫描黑名单行。
func scanBlacklist(rows *sql.Rows) (*adminsvc.Blacklist, error) {
	var item adminsvc.Blacklist
	var createdAt, updatedAt sql.NullTime
	if err := rows.Scan(&item.Id, &item.TargetType, &item.TargetId, &item.Reason, &item.OperatorId, &item.Status, &createdAt, &updatedAt); err != nil {
		return nil, err
	}
	item.CreatedAt = formatNullTime(createdAt)
	item.UpdatedAt = formatNullTime(updatedAt)
	return &item, nil
}

// scanRiskHitRecord 扫描风控命中记录行。
func scanRiskHitRecord(rows *sql.Rows) (*adminsvc.RiskHitRecord, error) {
	var item adminsvc.RiskHitRecord
	var createdAt sql.NullTime
	if err := rows.Scan(&item.Id, &item.BlacklistId, &item.TargetType, &item.TargetId, &item.Scene, &item.RiskLevel,
		&item.HitReason, &item.RequestId, &createdAt); err != nil {
		return nil, err
	}
	item.CreatedAt = formatNullTime(createdAt)
	return &item, nil
}

// nullableJSON 将空筛选条件写成 SQL NULL，非空条件按 JSON 字符串写入。
func nullableJSON(filters string) any {
	if strings.TrimSpace(filters) == "" {
		return nil
	}
	return filters
}

// isSupportedExportType 判断后台当前支持的真实导出类型。
func isSupportedExportType(exportType string) bool {
	switch exportType {
	case "orders":
		return true
	default:
		return false
	}
}

// loadExportTaskByNo 从 admin_export_task 读取后台任务信息，供异步任务生成 CSV 使用。
func loadExportTaskByNo(ctx context.Context, svcCtx *svc.ServiceContext, taskNo string) (*adminsvc.ExportTask, error) {
	row := svcCtx.MySQL.QueryRowContext(ctx, `
		SELECT task_no, export_type, COALESCE(CAST(filters AS CHAR), ''), status, admin_id,
		       file_path, file_url, failure_reason, created_at, updated_at, expires_at
		FROM admin_export_task
		WHERE task_no = ?
	`, taskNo)
	return scanExportTaskRow(row)
}

// writeExportCSV 根据导出类型分派到不同业务查询器。
func writeExportCSV(ctx context.Context, svcCtx *svc.ServiceContext, writer *csv.Writer, exportType string, filters orderExportFilters) error {
	switch exportType {
	case "orders":
		return writeOrdersCSV(ctx, svcCtx, writer, filters)
	default:
		return status.Error(codes.InvalidArgument, "导出类型暂不支持")
	}
}

// writeUsersCSV 分页导出用户基础信息。
func writeUsersCSV(ctx context.Context, svcCtx *svc.ServiceContext, writer *csv.Writer) error {
	if err := writer.Write([]string{"id", "phone", "nickname", "real_name", "status", "created_at"}); err != nil {
		return err
	}
	rows, err := svcCtx.MySQL.QueryContext(ctx, `
		SELECT id, phone, nickname, real_name, status, created_at
		FROM user
		WHERE deleted_at IS NULL
		ORDER BY id DESC
		LIMIT 5000
	`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		var statusCode int32
		var phone, nickname, realName sql.NullString
		var createdAt sql.NullTime
		if err := rows.Scan(&id, &phone, &nickname, &realName, &statusCode, &createdAt); err != nil {
			return err
		}
		if err := writer.Write([]string{strconv.FormatInt(id, 10), phone.String, nickname.String, realName.String, strconv.FormatInt(int64(statusCode), 10), formatNullTime(createdAt)}); err != nil {
			return err
		}
	}
	return rows.Err()
}

// writeDriversCSV 分页导出司机认证审核信息。
func writeDriversCSV(ctx context.Context, svcCtx *svc.ServiceContext, writer *csv.Writer) error {
	if err := writer.Write([]string{"certification_id", "driver_id", "driver_name", "driver_phone", "plate_no", "audit_status", "audited_by", "created_at"}); err != nil {
		return err
	}
	rows, err := svcCtx.MySQL.QueryContext(ctx, `
		SELECT c.id, c.driver_id, d.real_name, d.phone, v.plate_no, c.audit_status, c.audited_by, c.created_at
		FROM driver_certification c
		LEFT JOIN driver d ON d.id = c.driver_id
		LEFT JOIN vehicle v ON v.id = c.vehicle_id
		ORDER BY c.id DESC
		LIMIT 5000
	`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var id, driverID, auditedBy int64
		var auditStatus int32
		var name, phone, plateNo sql.NullString
		var createdAt sql.NullTime
		if err := rows.Scan(&id, &driverID, &name, &phone, &plateNo, &auditStatus, &auditedBy, &createdAt); err != nil {
			return err
		}
		if err := writer.Write([]string{strconv.FormatInt(id, 10), strconv.FormatInt(driverID, 10), name.String, phone.String, plateNo.String, strconv.FormatInt(int64(auditStatus), 10), strconv.FormatInt(auditedBy, 10), formatNullTime(createdAt)}); err != nil {
			return err
		}
	}
	return rows.Err()
}

// writeOrdersCSV 分页导出后台订单监控核心字段。
func writeOrdersCSV(ctx context.Context, svcCtx *svc.ServiceContext, writer *csv.Writer, filters orderExportFilters) error {
	if err := writer.Write([]string{"id", "order_no", "user_id", "driver_id", "status", "estimated_price", "created_at"}); err != nil {
		return err
	}
	where, args := filters.where()
	rows, err := svcCtx.MySQL.QueryContext(ctx, `
		SELECT id, order_no, user_id, driver_id, status, estimated_price, created_at
		FROM ride_order `+where+`
		ORDER BY id DESC
		LIMIT 5000
	`, args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var id, userID, driverID int64
		var statusCode int32
		var orderNo, estimatedPrice sql.NullString
		var createdAt sql.NullTime
		if err := rows.Scan(&id, &orderNo, &userID, &driverID, &statusCode, &estimatedPrice, &createdAt); err != nil {
			return err
		}
		if err := writer.Write([]string{strconv.FormatInt(id, 10), orderNo.String, strconv.FormatInt(userID, 10), strconv.FormatInt(driverID, 10), strconv.FormatInt(int64(statusCode), 10), estimatedPrice.String, formatNullTime(createdAt)}); err != nil {
			return err
		}
	}
	return rows.Err()
}

// writeOperationLogsCSV 分页导出后台审计日志。
func writeOperationLogsCSV(ctx context.Context, svcCtx *svc.ServiceContext, writer *csv.Writer) error {
	if err := writer.Write([]string{"id", "admin_id", "module", "action", "target_type", "target_id", "detail", "ip", "created_at"}); err != nil {
		return err
	}
	rows, err := svcCtx.MySQL.QueryContext(ctx, `
		SELECT id, admin_id, module, action, target_type, target_id, detail, ip, created_at
		FROM admin_operation_log
		ORDER BY id DESC
		LIMIT 5000
	`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var id, adminID, targetID int64
		var module, action, targetType, detail, ip sql.NullString
		var createdAt sql.NullTime
		if err := rows.Scan(&id, &adminID, &module, &action, &targetType, &targetID, &detail, &ip, &createdAt); err != nil {
			return err
		}
		if err := writer.Write([]string{strconv.FormatInt(id, 10), strconv.FormatInt(adminID, 10), module.String, action.String, targetType.String, strconv.FormatInt(targetID, 10), detail.String, ip.String, formatNullTime(createdAt)}); err != nil {
			return err
		}
	}
	return rows.Err()
}

// writeStatisticsCSV 导出当前运营统计聚合结果。
func writeStatisticsCSV(ctx context.Context, svcCtx *svc.ServiceContext, writer *csv.Writer) error {
	resp := &GetStatisticsOverviewLogic{ctx: ctx, svcCtx: svcCtx}
	overview, err := resp.GetStatisticsOverview(&adminsvc.StatisticsRequest{})
	if err != nil {
		return err
	}
	if err := writer.Write([]string{"metric", "value"}); err != nil {
		return err
	}
	return writer.WriteAll([][]string{
		{"user_count", strconv.FormatInt(overview.GetUserCount(), 10)},
		{"driver_count", strconv.FormatInt(overview.GetDriverCount(), 10)},
		{"order_count", strconv.FormatInt(overview.GetOrderCount(), 10)},
		{"completed_order_count", strconv.FormatInt(overview.GetCompletedOrderCount(), 10)},
		{"abnormal_order_count", strconv.FormatInt(overview.GetAbnormalOrderCount(), 10)},
		{"gmv", overview.GetGmv()},
		{"coupon_issue_count", strconv.FormatInt(overview.GetCouponIssueCount(), 10)},
		{"blacklist_count", strconv.FormatInt(overview.GetBlacklistCount(), 10)},
	})
}

// validateBlacklistRequest 校验黑名单新增或解除请求。
func validateBlacklistRequest(in *adminsvc.BlacklistRequest, release bool) error {
	if in.GetAdminId() <= 0 {
		return status.Error(codes.InvalidArgument, "管理员ID不能为空")
	}
	if release {
		if in.GetId() <= 0 {
			return status.Error(codes.InvalidArgument, "黑名单ID不能为空")
		}
		return nil
	}
	if in.GetTargetType() != "user" && in.GetTargetType() != "driver" && in.GetTargetType() != "device" && in.GetTargetType() != "phone" {
		return status.Error(codes.InvalidArgument, "黑名单目标类型不合法")
	}
	if in.GetTargetId() <= 0 || strings.TrimSpace(in.GetReason()) == "" {
		return status.Error(codes.InvalidArgument, "黑名单目标ID和原因不能为空")
	}
	return nil
}

// couponIssueStatusText 将发券任务状态码转换成前端可读文本。
func couponIssueStatusText(statusCode int32) string {
	switch statusCode {
	case 1:
		return "pending"
	case 2:
		return "running"
	case 3:
		return "success"
	case 4:
		return "partial_failed"
	case 5:
		return "failed"
	default:
		return "unknown"
	}
}

// couponIssueStatusCode 将查询参数中的状态转换为数据库状态码。
func couponIssueStatusCode(statusText string) (int32, bool) {
	switch strings.TrimSpace(statusText) {
	case "1", "pending":
		return 1, true
	case "2", "running":
		return 2, true
	case "3", "success":
		return 3, true
	case "4", "partial_failed":
		return 4, true
	case "5", "failed":
		return 5, true
	default:
		return 0, false
	}
}

// countSQL 执行 COUNT 查询。
func countSQL(ctx context.Context, db *sql.DB, query string, args ...any) (int64, error) {
	var count int64
	err := db.QueryRowContext(ctx, query, args...).Scan(&count)
	return count, err
}

// sumSQL 执行金额 SUM 查询，并统一返回字符串，避免浮点精度暴露给调用方。
func sumSQL(ctx context.Context, db *sql.DB, query string, args ...any) (string, error) {
	var value sql.NullString
	if err := db.QueryRowContext(ctx, query, args...).Scan(&value); err != nil {
		return "0.00", err
	}
	if !value.Valid || value.String == "" {
		return "0.00", nil
	}
	return value.String, nil
}

// percentText 计算百分比字符串。
func percentText(numerator, denominator int64) string {
	if denominator <= 0 {
		return "0.00%"
	}
	return fmt.Sprintf("%.2f%%", float64(numerator)*100/float64(denominator))
}

// newAdminTaskNo 生成后台任务编号。
func newAdminTaskNo(prefix string) string {
	return fmt.Sprintf("%s%s%s", prefix, time.Now().Format("20060102150405"), strconv.FormatInt(time.Now().UnixNano()%1000000, 10))
}
