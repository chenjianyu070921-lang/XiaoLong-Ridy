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
	"sync"
	"time"

	"XiaoLong-Ridy/rpc/adminsvc/adminsvc"
	"XiaoLong-Ridy/rpc/adminsvc/internal/svc"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	// maxConcurrentExportJobs 限制同一 adminsvc 实例内同时执行的导出任务数，防止大文件导出耗尽数据库连接和内存。
	maxConcurrentExportJobs = 4
	// exportJobQueueSize 限制尚未开始执行的导出任务数量，避免高峰请求无限积压在进程内。
	exportJobQueueSize = 64
)

// exportJob 表示已持久化、等待本实例执行的导出任务。
// 任务编号已写入数据库，worker 只负责驱动其状态机，不在内存中保存业务数据。
type exportJob struct {
	svcCtx *svc.ServiceContext
	taskNo string
}

var (
	exportJobQueue   = make(chan exportJob, exportJobQueueSize)
	exportWorkerOnce sync.Once
	// exportFileWriter 保留文件生成函数注入点，供测试覆盖 worker 的 panic 恢复路径。
	exportFileWriter   = writeExportTaskFile
	exportFileWriterMu sync.RWMutex
)

// startExportWorkers 按固定数量启动导出 worker。
// worker 数量固定，避免每次创建任务都启动一个 goroutine；队列容量同时限制等待任务数量。
func startExportWorkers() {
	exportWorkerOnce.Do(func() {
		for i := 0; i < maxConcurrentExportJobs; i++ {
			go func() {
				for job := range exportJobQueue {
					runExportTaskJob(job.svcCtx, job.taskNo)
				}
			}()
		}
	})
}

// couponIssueTargetConfig 表示后台发券任务的目标用户配置。
// P1 当前只落地显式用户 ID 发放，后续 crowd 人群包可由 MQ/Job 异步扩展。
type couponIssueTargetConfig struct {
	UserIDs []int64 `json:"user_ids"`
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
	// 优惠券状态约定：1 草稿、2 启用、3 停用；只有启用模板允许真正发券。
	if couponStatus != 2 {
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
	if err := createOperationLogTx(l.ctx, tx, in.GetAdminId(), "coupon", "issue", "coupon", in.GetCouponId(), fmt.Sprintf("创建发券任务：%s，成功%d，失败%d", taskNo, successCount, failCount), in.GetIp()); err != nil {
		return nil, err
	}
	if err := writeCouponPublishRecordTx(l.ctx, tx, in.GetCouponId(), taskNo, in.GetTargetConfig(), couponPublishStatus(taskStatus), failureReason, in.GetAdminId()); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
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

// writeCouponPublishRecordTx 在发券主事务中写入优惠券发布记录。
// 该记录与发券任务、用户券及领取数量保持原子一致，避免后台无法追溯已实际执行的发券动作。
func writeCouponPublishRecordTx(ctx context.Context, tx *sql.Tx, couponID int64, taskNo, targetConfig string, publishStatus int32, failureReason string, adminID int64) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO admin_coupon_publish_record
			(coupon_id, publish_version, publish_scope, target_config, status, failure_reason, operator_id)
		VALUES (?, ?, 'full', ?, ?, ?, ?)
	`, couponID, taskNo, targetConfig, publishStatus, failureReason, adminID)
	return err
}

// couponPublishStatus 将发券任务结果映射到优惠券发布记录状态。
// 全部成功记为发布成功；部分失败和全部失败均保留失败状态，并由 failure_reason 说明实际结果。
func couponPublishStatus(taskStatus int32) int32 {
	if taskStatus == 3 {
		return 2
	}
	return 3
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
	tx, err := l.svcCtx.MySQL.BeginTx(l.ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	res, err := tx.ExecContext(l.ctx, `
		INSERT INTO promotion_activity (name, type, config, start_at, end_at, status, created_by)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, in.GetName(), in.GetType(), in.GetConfig(), in.GetStartAt(), in.GetEndAt(), in.GetStatus(), in.GetAdminId())
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	if err := createOperationLogTx(l.ctx, tx, in.GetAdminId(), "promotion", "create", "promotion_activity", id, fmt.Sprintf("创建活动配置：%s", in.GetName()), in.GetIp()); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
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
	tx, err := l.svcCtx.MySQL.BeginTx(l.ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	res, err := tx.ExecContext(l.ctx, `
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
	if err := createOperationLogTx(l.ctx, tx, in.GetAdminId(), "promotion", "update", "promotion_activity", in.GetId(), fmt.Sprintf("编辑活动配置：%s", in.GetName()), in.GetIp()); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
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
		return nil, status.Error(codes.FailedPrecondition, "当前订单与支付表未保存城市编码，暂不支持按城市统计")
	}
	orderWhere, orderArgs := buildCreatedAtWhere("created_at", in.GetStartTime(), in.GetEndTime())
	payWhere, payArgs := buildCreatedAtWhere("created_at", in.GetStartTime(), in.GetEndTime())
	resp := &adminsvc.StatisticsOverviewResponse{}
	// 将全部聚合放入一次 SELECT，使 MySQL 在同一语句快照内计算指标，避免高并发写入造成指标互相不一致。
	args := make([]any, 0, len(orderArgs)*3+len(payArgs))
	args = append(args, orderArgs...)
	args = append(args, orderArgs...)
	args = append(args, orderArgs...)
	args = append(args, payArgs...)
	err := l.svcCtx.MySQL.QueryRowContext(l.ctx, `
		SELECT
			(SELECT COUNT(1) FROM user WHERE deleted_at IS NULL),
			(SELECT COUNT(1) FROM driver WHERE deleted_at IS NULL),
			(SELECT COUNT(1) FROM ride_order `+orderWhere+`),
			(SELECT COUNT(1) FROM ride_order `+appendWhere(orderWhere, "status = 5")+`),
			(SELECT COUNT(1) FROM ride_order `+appendWhere(orderWhere, "status = 6")+`),
			(SELECT COUNT(1) FROM user_coupon),
			(SELECT COUNT(1) FROM blacklist WHERE status = 1),
			(SELECT COALESCE(SUM(amount), 0) FROM payment `+appendWhere(payWhere, "status = 2")+`)
	`, args...).Scan(&resp.UserCount, &resp.DriverCount, &resp.OrderCount, &resp.CompletedOrderCount,
		&resp.AbnormalOrderCount, &resp.CouponIssueCount, &resp.BlacklistCount, &resp.Gmv)
	if err != nil {
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
		return nil, status.Error(codes.FailedPrecondition, "当前订单与支付表未保存城市编码，暂不支持按城市统计")
	}
	// 超时记录和支付记录本身的创建时间不代表订单统计时间。
	// 统一通过订单主表的 created_at 过滤，保证五项指标使用同一订单时间范围。
	where, args := buildCreatedAtWhere("ro.created_at", in.GetStartTime(), in.GetEndTime())
	resp := &adminsvc.OrderStatisticsResponse{}
	queryArgs := make([]any, 0, len(args)*5)
	queryArgs = append(queryArgs, args...)
	queryArgs = append(queryArgs, args...)
	queryArgs = append(queryArgs, args...)
	queryArgs = append(queryArgs, args...)
	queryArgs = append(queryArgs, args...)
	err := l.svcCtx.MySQL.QueryRowContext(l.ctx, `
		SELECT
			(SELECT COUNT(1) FROM ride_order ro `+where+`),
			(SELECT COUNT(1) FROM ride_order ro `+appendWhere(where, "ro.status = 5")+`),
			(SELECT COUNT(1) FROM ride_order ro `+appendWhere(where, "ro.status = 6")+`),
			(SELECT COUNT(1)
			 FROM dispatch_record dr
			 JOIN ride_order ro ON ro.id = dr.order_id
			 `+appendWhere(where, "dr.status = 4")+`),
			(SELECT COUNT(1)
			 FROM payment p
			 JOIN ride_order ro ON ro.id = p.order_id
			 `+appendWhere(where, "p.status = 3")+`)
	`, queryArgs...).Scan(&resp.OrderCount, &resp.CompletedOrderCount, &resp.CanceledOrderCount,
		&resp.TimeoutOrderCount, &resp.PaymentAbnormalCount)
	if err != nil {
		return nil, err
	}
	resp.CompletionRate = percentText(resp.CompletedOrderCount, resp.OrderCount)
	resp.CancelRate = percentText(resp.CanceledOrderCount, resp.OrderCount)
	return resp, nil
}

// GetDriverStatisticsLogic 处理司机经营统计。
type GetDriverStatisticsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGetDriverStatisticsLogic 创建司机经营统计逻辑对象。
func NewGetDriverStatisticsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDriverStatisticsLogic {
	return &GetDriverStatisticsLogic{ctx: ctx, svcCtx: svcCtx}
}

// GetDriverStatistics 汇总司机入驻、审核、完单、收入、提现和服务质量指标。
// 当前数据库没有司机在线状态和独立接单事件表，因此本接口只返回可由现有表可靠统计的字段，避免伪造运营数据。
func (l *GetDriverStatisticsLogic) GetDriverStatistics(in *adminsvc.StatisticsRequest) (*adminsvc.DriverStatisticsResponse, error) {
	if strings.TrimSpace(in.GetCityCode()) != "" {
		return nil, status.Error(codes.FailedPrecondition, "当前司机、订单与结算表未保存城市编码，暂不支持按城市统计")
	}
	driverWhere, driverArgs := buildCreatedAtWhere("d.created_at", in.GetStartTime(), in.GetEndTime())
	certWhere, certArgs := buildCreatedAtWhere("dc.created_at", in.GetStartTime(), in.GetEndTime())
	orderWhere, orderArgs := buildCreatedAtWhere("ro.created_at", in.GetStartTime(), in.GetEndTime())
	settlementWhere, settlementArgs := buildCreatedAtWhere("s.created_at", in.GetStartTime(), in.GetEndTime())
	withdrawWhere, withdrawArgs := buildCreatedAtWhere("dw.created_at", in.GetStartTime(), in.GetEndTime())
	scoreWhere, scoreArgs := buildCreatedAtWhere("ds.updated_at", in.GetStartTime(), in.GetEndTime())

	resp := &adminsvc.DriverStatisticsResponse{}
	// 将司机报表各指标放在同一条 SQL 中读取，减少多次查询带来的快照不一致。
	args := make([]any, 0, len(driverArgs)+len(certArgs)*2+len(orderArgs)+len(settlementArgs)+len(withdrawArgs)*3+len(scoreArgs)*2)
	args = append(args, driverArgs...)
	args = append(args, certArgs...)
	args = append(args, certArgs...)
	args = append(args, orderArgs...)
	args = append(args, settlementArgs...)
	args = append(args, withdrawArgs...)
	args = append(args, withdrawArgs...)
	args = append(args, withdrawArgs...)
	args = append(args, scoreArgs...)
	args = append(args, scoreArgs...)
	err := l.svcCtx.MySQL.QueryRowContext(l.ctx, `
		SELECT
			(SELECT COUNT(1) FROM driver WHERE deleted_at IS NULL),
			(SELECT COUNT(1) FROM driver d `+appendWhere(driverWhere, "d.deleted_at IS NULL")+`),
			(SELECT COUNT(1) FROM driver_certification dc `+appendWhere(certWhere, "dc.audit_status = 1")+`),
			(SELECT COUNT(1) FROM driver_certification dc `+appendWhere(certWhere, "dc.audit_status = 2")+`),
			(SELECT COUNT(1) FROM ride_order ro `+appendWhere(orderWhere, "ro.status = 5 AND ro.driver_id > 0")+`),
			(SELECT COALESCE(SUM(s.driver_income), 0) FROM settlement s `+appendWhere(settlementWhere, "s.status = 2")+`),
			(SELECT COALESCE(SUM(dw.amount), 0) FROM driver_withdraw dw `+appendWhere(withdrawWhere, "dw.status = 1")+`),
			(SELECT COALESCE(SUM(dw.amount), 0) FROM driver_withdraw dw `+appendWhere(withdrawWhere, "dw.status = 2")+`),
			(SELECT COUNT(1) FROM driver_withdraw dw `+appendWhere(withdrawWhere, "dw.status = 3")+`),
			(SELECT COALESCE(AVG(ds.score), 0) FROM driver_score ds `+scoreWhere+`),
			(SELECT COALESCE(SUM(ds.month_complaint_count), 0) FROM driver_score ds `+scoreWhere+`)
	`, args...).Scan(&resp.DriverTotal, &resp.NewDriverCount, &resp.PendingAuditCount,
		&resp.ApprovedDriverCount, &resp.CompletedOrderCount, &resp.DriverIncome,
		&resp.WithdrawPendingAmount, &resp.WithdrawSuccessAmount, &resp.WithdrawFailedCount,
		&resp.AverageScore, &resp.TotalComplaintCount)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// GetFinanceStatisticsLogic 处理后台财务统计。
type GetFinanceStatisticsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGetFinanceStatisticsLogic 创建财务统计逻辑对象。
func NewGetFinanceStatisticsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFinanceStatisticsLogic {
	return &GetFinanceStatisticsLogic{ctx: ctx, svcCtx: svcCtx}
}

// GetFinanceStatistics 汇总支付、退款、结算、平台抽佣、司机收入和平台补贴指标。
// 金额字段以字符串返回，保持 MySQL DECIMAL 精度，不在 Go 层转为浮点数。
func (l *GetFinanceStatisticsLogic) GetFinanceStatistics(in *adminsvc.StatisticsRequest) (*adminsvc.FinanceStatisticsResponse, error) {
	if strings.TrimSpace(in.GetCityCode()) != "" {
		return nil, status.Error(codes.FailedPrecondition, "当前支付、结算与价格表未保存城市编码，暂不支持按城市统计")
	}
	paymentWhere, paymentArgs := buildCreatedAtWhere("p.created_at", in.GetStartTime(), in.GetEndTime())
	settlementWhere, settlementArgs := buildCreatedAtWhere("s.created_at", in.GetStartTime(), in.GetEndTime())
	priceWhere, priceArgs := buildCreatedAtWhere("op.created_at", in.GetStartTime(), in.GetEndTime())

	resp := &adminsvc.FinanceStatisticsResponse{}
	// 财务报表按业务记录创建时间分别统计，避免支付、结算和补贴明细被订单时间错误截断。
	args := make([]any, 0, len(paymentArgs)*5+len(settlementArgs)*4+len(priceArgs))
	args = append(args, paymentArgs...)
	args = append(args, paymentArgs...)
	args = append(args, paymentArgs...)
	args = append(args, paymentArgs...)
	args = append(args, paymentArgs...)
	args = append(args, settlementArgs...)
	args = append(args, settlementArgs...)
	args = append(args, settlementArgs...)
	args = append(args, settlementArgs...)
	args = append(args, priceArgs...)
	err := l.svcCtx.MySQL.QueryRowContext(l.ctx, `
		SELECT
			(SELECT COUNT(1) FROM payment p `+appendWhere(paymentWhere, "p.status = 2")+`),
			(SELECT COALESCE(SUM(p.amount), 0) FROM payment p `+appendWhere(paymentWhere, "p.status = 2")+`),
			(SELECT COUNT(1) FROM payment p `+appendWhere(paymentWhere, "p.refund_amount > 0")+`),
			(SELECT COALESCE(SUM(p.refund_amount), 0) FROM payment p `+appendWhere(paymentWhere, "p.refund_amount > 0")+`),
			(SELECT COUNT(1) FROM payment p `+appendWhere(paymentWhere, "p.status = 3")+`),
			(SELECT COUNT(1) FROM settlement s `+appendWhere(settlementWhere, "s.status = 2")+`),
			(SELECT COALESCE(SUM(s.total_amount), 0) FROM settlement s `+appendWhere(settlementWhere, "s.status = 2")+`),
			(SELECT COALESCE(SUM(s.platform_commission), 0) FROM settlement s `+appendWhere(settlementWhere, "s.status = 2")+`),
			(SELECT COALESCE(SUM(s.driver_income), 0) FROM settlement s `+appendWhere(settlementWhere, "s.status = 2")+`),
			(SELECT COALESCE(SUM(op.platform_subsidy), 0) FROM order_price op `+priceWhere+`)
	`, args...).Scan(&resp.PaymentOrderCount, &resp.PaidAmount, &resp.RefundOrderCount,
		&resp.RefundAmount, &resp.PaymentFailedCount, &resp.SettlementOrderCount,
		&resp.SettlementTotalAmount, &resp.PlatformCommission, &resp.DriverIncome,
		&resp.PlatformSubsidy)
	if err != nil {
		return nil, err
	}
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
	if strings.TrimSpace(in.GetCityCode()) != "" {
		return nil, status.Error(codes.FailedPrecondition, "当前优惠券数据未保存城市编码，暂不支持按城市统计")
	}
	resp := &adminsvc.CouponStatisticsResponse{}
	err := l.svcCtx.MySQL.QueryRowContext(l.ctx, `
		SELECT
			(SELECT COUNT(1) FROM coupon),
			(SELECT COUNT(1) FROM coupon WHERE status = 2),
			(SELECT COUNT(1) FROM user_coupon),
			(SELECT COUNT(1) FROM user_coupon WHERE status = 2),
			(SELECT COUNT(1) FROM user_coupon WHERE status = 3 OR (status = 1 AND expire_at < NOW()))
	`).Scan(&resp.CouponCount, &resp.EnabledCouponCount, &resp.IssuedCouponCount,
		&resp.UsedCouponCount, &resp.ExpiredCouponCount)
	if err != nil {
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
	if !isSupportedExportType(exportType) {
		return nil, status.Error(codes.InvalidArgument, "导出类型暂不支持")
	}
	if _, err := parseExportFilters(exportType, in.GetFilters()); err != nil {
		return nil, err
	}
	taskNo := newAdminTaskNo("EX")
	if _, err := l.svcCtx.MySQL.ExecContext(l.ctx, `
		INSERT INTO admin_export_task (task_no, export_type, filters, status, admin_id, ip)
		VALUES (?, ?, ?, 'pending', ?, ?)
	`, taskNo, exportType, nullableJSON(in.GetFilters()), in.GetAdminId(), in.GetIp()); err != nil {
		return nil, err
	}
	_ = createOperationLog(l.ctx, l.svcCtx, in.GetAdminId(), "export", "create", exportType, 0, fmt.Sprintf("创建导出任务：%s", taskNo), in.GetIp())
	startExportWorkers()
	select {
	case exportJobQueue <- exportJob{svcCtx: l.svcCtx, taskNo: taskNo}:
		// 任务由固定数量 worker 异步执行，接口立即返回 pending 状态。
	case <-l.ctx.Done():
		markExportTaskFailed(l.svcCtx, taskNo, "导出任务投递已取消")
		return nil, status.Error(codes.Canceled, "导出任务投递已取消")
	default:
		// 内存队列满时不能让任务永久保持 pending；明确落失败状态，调用方可稍后重新创建任务。
		markExportTaskFailed(l.svcCtx, taskNo, "导出任务队列繁忙，请稍后重试")
		return nil, status.Error(codes.ResourceExhausted, "导出任务队列繁忙，请稍后重试")
	}
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
	if err := writeAuditAfterCommitted(l.ctx, l.svcCtx, in.GetAdminId(), "risk", "add_blacklist", "blacklist", id, fmt.Sprintf("新增黑名单：%s/%d", in.GetTargetType(), in.GetTargetId()), in.GetIp()); err != nil {
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
	if err := writeAuditAfterCommitted(l.ctx, l.svcCtx, in.GetAdminId(), "risk", "release_blacklist", "blacklist", in.GetId(), "解除黑名单", in.GetIp()); err != nil {
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

// HandleRiskHitRecordsLogic 处理风控命中记录人工处置。
type HandleRiskHitRecordsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewHandleRiskHitRecordsLogic 创建风控命中处置逻辑对象。
func NewHandleRiskHitRecordsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HandleRiskHitRecordsLogic {
	return &HandleRiskHitRecordsLogic{ctx: ctx, svcCtx: svcCtx}
}

// HandleRiskHitRecords 对风控命中记录执行复核通过、加入黑名单或转工单。
// risk_blacklist_hit_record 当前没有处理状态字段，因此复核通过只写操作日志；
// 拉黑和转工单会写入对应业务表，并在同一事务中记录审计，形成可追溯闭环。
func (l *HandleRiskHitRecordsLogic) HandleRiskHitRecords(in *adminsvc.RiskHitActionRequest) (*adminsvc.RiskHitActionResponse, error) {
	if err := validateRiskHitActionRequest(in); err != nil {
		return nil, err
	}
	resp := &adminsvc.RiskHitActionResponse{}
	for _, id := range uniquePositiveIDs(in.GetIds()) {
		workOrderID, err := l.handleOneRiskHit(id, in)
		if err != nil {
			resp.FailCount++
			resp.FailureReasons = append(resp.FailureReasons, fmt.Sprintf("命中记录%d：%s", id, err.Error()))
			continue
		}
		resp.SuccessCount++
		if workOrderID > 0 {
			resp.WorkOrderIds = append(resp.WorkOrderIds, workOrderID)
		}
	}
	return resp, nil
}

// handleOneRiskHit 在独立事务中处置单条命中记录，避免批量操作部分失败时回滚已成功记录。
func (l *HandleRiskHitRecordsLogic) handleOneRiskHit(id int64, in *adminsvc.RiskHitActionRequest) (int64, error) {
	tx, err := l.svcCtx.MySQL.BeginTx(l.ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	hit, err := scanRiskHitRecordRow(tx.QueryRowContext(l.ctx, `
		SELECT id, blacklist_id, target_type, target_id, scene, risk_level, hit_reason, request_id, created_at
		FROM risk_blacklist_hit_record
		WHERE id = ?
		FOR UPDATE
	`, id))
	if err != nil {
		return 0, err
	}
	var workOrderID int64
	switch in.GetAction() {
	case "review_pass":
		if err = createOperationLogTx(l.ctx, tx, in.GetAdminId(), "risk", "review_pass", "risk_hit_record", id, riskHitLogDetail(hit, in.GetReason()), in.GetIp()); err != nil {
			return 0, err
		}
	case "add_blacklist":
		res, err := tx.ExecContext(l.ctx, `
			INSERT INTO blacklist (target_type, target_id, reason, operator_id, status)
			VALUES (?, ?, ?, ?, 1)
		`, hit.GetTargetType(), hit.GetTargetId(), strings.TrimSpace(in.GetReason()), in.GetAdminId())
		if err != nil {
			return 0, err
		}
		blacklistID, _ := res.LastInsertId()
		if err = createOperationLogTx(l.ctx, tx, in.GetAdminId(), "risk", "hit_add_blacklist", "blacklist", blacklistID, riskHitLogDetail(hit, in.GetReason()), in.GetIp()); err != nil {
			return 0, err
		}
	case "create_work_order":
		workOrderID, err = insertRiskHitWorkOrderTx(l.ctx, tx, hit, in)
		if err != nil {
			return 0, err
		}
		if err = createOperationLogTx(l.ctx, tx, in.GetAdminId(), "risk", "hit_create_work_order", "work_order", workOrderID, riskHitLogDetail(hit, in.GetReason()), in.GetIp()); err != nil {
			return 0, err
		}
	default:
		return 0, status.Error(codes.InvalidArgument, "风控处置动作不合法")
	}
	return workOrderID, tx.Commit()
}

// insertRiskHitWorkOrderTx 将高风险命中转为后台工单，供运营继续仲裁、补证和结案。
func insertRiskHitWorkOrderTx(ctx context.Context, tx *sql.Tx, hit *adminsvc.RiskHitRecord, in *adminsvc.RiskHitActionRequest) (int64, error) {
	title := strings.TrimSpace(in.GetWorkOrderTitle())
	if title == "" {
		title = fmt.Sprintf("风控命中复核：%s/%d", hit.GetTargetType(), hit.GetTargetId())
	}
	priority := normalizeRiskWorkOrderPriority(in.GetPriority(), hit.GetRiskLevel())
	userID, driverID := int64(0), int64(0)
	if hit.GetTargetType() == "user" {
		userID = hit.GetTargetId()
	}
	if hit.GetTargetType() == "driver" {
		driverID = hit.GetTargetId()
	}
	content := strings.TrimSpace(in.GetReason())
	if content == "" {
		content = hit.GetHitReason()
	}
	res, err := tx.ExecContext(ctx, `
		INSERT INTO admin_complaint_work_order
			(work_order_no, work_order_type, source_type, source_id, order_id, user_id, driver_id, title, content, priority, status, created_by)
		VALUES (?, 3, ?, ?, 0, ?, ?, ?, ?, ?, 1, ?)
	`, newAdminTaskNo("WO"), hit.GetTargetType(), hit.GetTargetId(), userID, driverID, title, content, priority, in.GetAdminId())
	if err != nil {
		return 0, err
	}
	id, _ := res.LastInsertId()
	if _, err = tx.ExecContext(ctx, `INSERT INTO admin_work_order_flow (work_order_id, from_status, to_status, action, operator_id, content) VALUES (?, 0, 1, 'create_from_risk', ?, ?)`, id, in.GetAdminId(), content); err != nil {
		return 0, err
	}
	return id, nil
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

// changePromotionStatus 修改活动运行状态。
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

// changePromotionStatus 通过源状态限制更新，保证发布和回滚不会越过活动状态机。
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
// 任意步骤失败和 panic 都会尽力回写 failed，避免任务卡在 pending 或 running 状态。
func runExportTaskJob(svcCtx *svc.ServiceContext, taskNo string) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	defer func() {
		if recovered := recover(); recovered != nil {
			// goroutine 内 panic 不得传播到服务进程；失败回写本身仅尽力执行。
			markExportTaskFailed(svcCtx, taskNo, fmt.Sprintf("导出任务异常: %v", recovered))
		}
	}()

	task, err := loadExportTaskByNo(ctx, svcCtx, taskNo)
	if err != nil {
		markExportTaskFailed(svcCtx, taskNo, fmt.Sprintf("加载导出任务失败: %v", err))
		return
	}
	if err := updateExportTaskStatus(ctx, svcCtx, taskNo, "running", "", "", nil); err != nil {
		markExportTaskFailed(svcCtx, taskNo, fmt.Sprintf("更新导出任务运行状态失败: %v", err))
		return
	}
	filePath, err := callExportFileWriter(ctx, svcCtx, task)
	if err != nil {
		markExportTaskFailed(svcCtx, taskNo, fmt.Sprintf("生成导出文件失败: %v", err))
		return
	}
	if err := updateExportTaskStatus(ctx, svcCtx, taskNo, "success", filePath, "", time.Now().Add(7*24*time.Hour)); err != nil {
		markExportTaskFailed(svcCtx, taskNo, fmt.Sprintf("更新导出任务成功状态失败: %v", err))
		return
	}
	// 每次成功生成新文件后顺带清理已过期文件，避免本地导出目录无界增长。
	_ = cleanupExpiredExportFiles(ctx, svcCtx)
}

// callExportFileWriter 在读取可替换的文件生成器时提供并发保护。
// 正常运行固定使用 writeExportTaskFile，测试替换生成器时不会与异步 worker 发生数据竞争。
func callExportFileWriter(ctx context.Context, svcCtx *svc.ServiceContext, task *adminsvc.ExportTask) (string, error) {
	exportFileWriterMu.RLock()
	fileWriter := exportFileWriter
	exportFileWriterMu.RUnlock()
	return fileWriter(ctx, svcCtx, task)
}

// markExportTaskFailed 尽力将任务标记为失败。
// 失败回写使用独立的短超时上下文，避免导出主流程已超时时无法收口任务状态。
func markExportTaskFailed(svcCtx *svc.ServiceContext, taskNo, failureReason string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := updateExportTaskStatus(ctx, svcCtx, taskNo, "failed", "", failureReason, nil); err != nil {
		return
	}
}

// cleanupExpiredExportFiles 删除已过期任务在受控目录中的 CSV，并保留任务记录用于审计。
// 仅接受与任务号完全匹配的文件名，避免数据库异常值导致删除任意路径。
func cleanupExpiredExportFiles(ctx context.Context, svcCtx *svc.ServiceContext) error {
	rows, err := svcCtx.MySQL.QueryContext(ctx, `SELECT task_no, file_path FROM admin_export_task WHERE expires_at IS NOT NULL AND expires_at <= ? AND file_path <> ''`, time.Now())
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var taskNo, filePath string
		if err := rows.Scan(&taskNo, &filePath); err != nil {
			return err
		}
		if filepath.Base(filePath) != taskNo+".csv" {
			continue
		}
		if err := os.Remove(filepath.Join(".tmp-admin-exports", taskNo+".csv")); err != nil && !os.IsNotExist(err) {
			continue
		}
		if _, err := svcCtx.MySQL.ExecContext(ctx, `UPDATE admin_export_task SET file_path = '', file_url = '' WHERE task_no = ?`, taskNo); err != nil {
			return err
		}
	}
	return rows.Err()
}

// writeExportTaskFile 按导出类型分页读取业务表并生成 CSV 文件。
// 当前先覆盖后台核心审查场景：用户、司机认证、订单和操作日志；统计类导出输出实时聚合指标。
func writeExportTaskFile(ctx context.Context, svcCtx *svc.ServiceContext, task *adminsvc.ExportTask) (string, error) {
	filters, err := parseExportFilters(task.GetExportType(), task.GetFilters())
	if err != nil {
		return "", err
	}
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

// scanRiskHitRecord 扫描风控命中记录列表行。
func scanRiskHitRecord(rows *sql.Rows) (*adminsvc.RiskHitRecord, error) {
	return scanRiskHitRecordScanner(rows)
}

// scanRiskHitRecordRow 扫描单条风控命中记录，并将空结果转换为业务 NotFound。
func scanRiskHitRecordRow(row *sql.Row) (*adminsvc.RiskHitRecord, error) {
	item, err := scanRiskHitRecordScanner(row)
	if err == sql.ErrNoRows {
		return nil, status.Error(codes.NotFound, "风控命中记录不存在")
	}
	return item, err
}

// scanRiskHitRecordScanner 兼容 sql.Rows 和 sql.Row，避免列表和处置接口重复扫描字段。
func scanRiskHitRecordScanner(rows interface{ Scan(...any) error }) (*adminsvc.RiskHitRecord, error) {
	var item adminsvc.RiskHitRecord
	var createdAt sql.NullTime
	if err := rows.Scan(&item.Id, &item.BlacklistId, &item.TargetType, &item.TargetId, &item.Scene, &item.RiskLevel,
		&item.HitReason, &item.RequestId, &createdAt); err != nil {
		return nil, err
	}
	item.CreatedAt = formatNullTime(createdAt)
	return &item, nil
}

// uniquePositiveIDs 对批量请求中的 ID 去重并丢弃非正数，防止重复执行同一业务动作。
func uniquePositiveIDs(ids []int64) []int64 {
	seen := make(map[int64]struct{}, len(ids))
	result := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}

// normalizeRiskWorkOrderPriority 将风控风险等级映射为工单优先级，同时允许运营手动指定合法优先级。
func normalizeRiskWorkOrderPriority(priority, riskLevel int32) int32 {
	if priority >= 1 && priority <= 4 {
		return priority
	}
	switch riskLevel {
	case 3:
		return 4
	case 2:
		return 3
	default:
		return 2
	}
}

// riskHitLogDetail 生成风控命中处置审计详情，保留命中场景、对象和处置原因。
func riskHitLogDetail(hit *adminsvc.RiskHitRecord, reason string) string {
	detail := fmt.Sprintf("处置风控命中：%s/%d，场景：%s，原因：%s", hit.GetTargetType(), hit.GetTargetId(), hit.GetScene(), hit.GetHitReason())
	if strings.TrimSpace(reason) != "" {
		detail += "，处置说明：" + strings.TrimSpace(reason)
	}
	return detail
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
	case "users", "drivers", "orders", "operation_logs", "statistics":
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
func writeExportCSV(ctx context.Context, svcCtx *svc.ServiceContext, writer *csv.Writer, exportType string, filters exportFilters) error {
	switch exportType {
	case "users":
		return writeUsersCSV(ctx, svcCtx, writer, filters)
	case "drivers":
		return writeDriversCSV(ctx, svcCtx, writer, filters)
	case "orders":
		return writeOrdersCSV(ctx, svcCtx, writer, filters)
	case "operation_logs":
		return writeOperationLogsCSV(ctx, svcCtx, writer, filters)
	case "statistics":
		return writeStatisticsCSV(ctx, svcCtx, writer, filters)
	default:
		return status.Error(codes.InvalidArgument, "导出类型暂不支持")
	}
}

// writeUsersCSV 分页导出用户基础信息。
func writeUsersCSV(ctx context.Context, svcCtx *svc.ServiceContext, writer *csv.Writer, filters exportFilters) error {
	if err := writer.Write([]string{"id", "phone", "nickname", "real_name", "status", "created_at"}); err != nil {
		return err
	}
	where, args := exportWhere("deleted_at IS NULL", filters, map[string]string{"keyword": "phone", "status": "status", "created_at": "created_at"})
	rows, err := svcCtx.MySQL.QueryContext(ctx, `
		SELECT id, phone, nickname, real_name, status, created_at
		FROM user `+where+`
		ORDER BY id DESC
		LIMIT 5000
	`, args...)
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
func writeDriversCSV(ctx context.Context, svcCtx *svc.ServiceContext, writer *csv.Writer, filters exportFilters) error {
	if err := writer.Write([]string{"certification_id", "driver_id", "driver_name", "driver_phone", "plate_no", "audit_status", "audited_by", "created_at"}); err != nil {
		return err
	}
	where, args := exportWhere("1=1", filters, map[string]string{"keyword": "d.phone", "audit_status": "c.audit_status", "created_at": "c.created_at"})
	rows, err := svcCtx.MySQL.QueryContext(ctx, `
		SELECT c.id, c.driver_id, d.real_name, d.phone, v.plate_no, c.audit_status, c.audited_by, c.created_at
		FROM driver_certification c
		LEFT JOIN driver d ON d.id = c.driver_id
		LEFT JOIN vehicle v ON v.id = c.vehicle_id
		`+where+`
		ORDER BY c.id DESC
		LIMIT 5000
	`, args...)
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
func writeOrdersCSV(ctx context.Context, svcCtx *svc.ServiceContext, writer *csv.Writer, filters exportFilters) error {
	if err := writer.Write([]string{"id", "order_no", "user_id", "driver_id", "status", "estimated_price", "created_at"}); err != nil {
		return err
	}
	where, args := exportWhere("1=1", filters, map[string]string{"keyword": "order_no", "status": "status", "user_id": "user_id", "driver_id": "driver_id", "created_at": "created_at"})
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
func writeOperationLogsCSV(ctx context.Context, svcCtx *svc.ServiceContext, writer *csv.Writer, filters exportFilters) error {
	if err := writer.Write([]string{"id", "admin_id", "module", "action", "target_type", "target_id", "detail", "ip", "created_at"}); err != nil {
		return err
	}
	where, args := exportWhere("1=1", filters, map[string]string{"admin_id": "admin_id", "module": "module", "action": "action", "target_type": "target_type", "target_id": "target_id", "created_at": "created_at"})
	rows, err := svcCtx.MySQL.QueryContext(ctx, `
		SELECT id, admin_id, module, action, target_type, target_id, detail, ip, created_at
		FROM admin_operation_log `+where+`
		ORDER BY id DESC
		LIMIT 5000
	`, args...)
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
func writeStatisticsCSV(ctx context.Context, svcCtx *svc.ServiceContext, writer *csv.Writer, filters exportFilters) error {
	resp := &GetStatisticsOverviewLogic{ctx: ctx, svcCtx: svcCtx}
	overview, err := resp.GetStatisticsOverview(&adminsvc.StatisticsRequest{StartTime: filters.StartTime, EndTime: filters.EndTime})
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

// validateRiskHitActionRequest 校验风控命中处置请求，避免空批量或未知动作进入事务。
func validateRiskHitActionRequest(in *adminsvc.RiskHitActionRequest) error {
	if in.GetAdminId() <= 0 || len(uniquePositiveIDs(in.GetIds())) == 0 {
		return status.Error(codes.InvalidArgument, "命中记录ID列表和管理员ID不能为空")
	}
	switch in.GetAction() {
	case "review_pass":
		return nil
	case "add_blacklist", "create_work_order":
		if strings.TrimSpace(in.GetReason()) == "" {
			return status.Error(codes.InvalidArgument, "处置原因不能为空")
		}
		if in.GetPriority() < 0 || in.GetPriority() > 4 {
			return status.Error(codes.InvalidArgument, "工单优先级不合法")
		}
		return nil
	default:
		return status.Error(codes.InvalidArgument, "风控处置动作不合法")
	}
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
