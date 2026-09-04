package adminservicelogic

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"XiaoLong-Ridy/rpc/adminsvc/adminsvc"
	"XiaoLong-Ridy/rpc/adminsvc/internal/svc"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// driverPunishmentActions 定义第一期允许下发给司机域的处罚动作。
// fine 仅登记待结算罚款，不会在 adminsvc 内直接修改司机余额。
var driverPunishmentActions = map[string]struct{}{
	"no_dispatch":  {},
	"freeze":       {},
	"deduct_score": {},
	"fine":         {},
	"downgrade":    {},
}

// DriverPunishmentLogic 负责处罚规则、处罚单和申诉审核的后台编排。
// 所有写操作均在同一事务内写业务状态、操作审计和领域 outbox，禁止直接写 driversvc 数据表。
type DriverPunishmentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewDriverPunishmentLogic 创建司机处罚后台逻辑对象。
func NewDriverPunishmentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DriverPunishmentLogic {
	return &DriverPunishmentLogic{ctx: ctx, svcCtx: svcCtx}
}

// ListRules 查询处罚规则分页列表，客服只读权限由服务端拦截器统一判定。
func (l *DriverPunishmentLogic) ListRules(in *adminsvc.DriverPunishmentRuleListRequest) (*adminsvc.DriverPunishmentRuleListResponse, error) {
	where, args := buildPunishmentRuleWhere(in)
	var total int64
	if err := l.svcCtx.MySQL.QueryRowContext(l.ctx, "SELECT COUNT(1) FROM admin_driver_punishment_rule "+where, args...).Scan(&total); err != nil {
		return nil, err
	}
	page, pageSize := normalizePage(in.GetPage()), normalizePageSize(in.GetPageSize())
	queryArgs := append(append([]any{}, args...), pageSize, offset(page, pageSize))
	rows, err := l.svcCtx.MySQL.QueryContext(l.ctx, `
		SELECT id, name, violation_type, CAST(actions AS CHAR), penalty_cents, score_delta,
		       priority_weight_delta, status, version, created_by, updated_by, created_at, updated_at
		FROM admin_driver_punishment_rule `+where+`
		ORDER BY id DESC LIMIT ? OFFSET ?`, queryArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*adminsvc.DriverPunishmentRule, 0)
	for rows.Next() {
		item, err := scanDriverPunishmentRule(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, item)
	}
	return &adminsvc.DriverPunishmentRuleListResponse{List: list, Total: total, Page: page, PageSize: pageSize}, rows.Err()
}

// CreateRule 创建处罚规则。运营可保存规则，规则发布启停仍由服务端权限矩阵约束为超管。
func (l *DriverPunishmentLogic) CreateRule(in *adminsvc.DriverPunishmentRuleRequest) (*adminsvc.DriverPunishmentRule, error) {
	if err := requireAdminRoles(l.ctx, l.svcCtx, 1, 2); err != nil {
		return nil, err
	}
	if err := validateDriverPunishmentRule(in); err != nil {
		return nil, err
	}
	tx, err := l.svcCtx.MySQL.BeginTx(l.ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	res, err := tx.ExecContext(l.ctx, `
		INSERT INTO admin_driver_punishment_rule
			(name, violation_type, actions, penalty_cents, score_delta, priority_weight_delta, status, version, created_by, updated_by)
		VALUES (?, ?, ?, ?, ?, ?, ?, 1, ?, ?)`,
		strings.TrimSpace(in.GetName()), strings.TrimSpace(in.GetViolationType()), in.GetActions(), in.GetPenaltyCents(),
		in.GetScoreDelta(), in.GetPriorityWeightDelta(), in.GetStatus(), in.GetAdminId(), in.GetAdminId())
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	if err := createOperationLogTx(l.ctx, tx, in.GetAdminId(), "driver_punishment", "rule_create", "driver_punishment_rule", id, "创建司机处罚规则："+strings.TrimSpace(in.GetName()), in.GetIp()); err != nil {
		return nil, err
	}
	item, err := scanDriverPunishmentRule(tx.QueryRowContext(l.ctx, `
		SELECT id, name, violation_type, CAST(actions AS CHAR), penalty_cents, score_delta,
		       priority_weight_delta, status, version, created_by, updated_by, created_at, updated_at
		FROM admin_driver_punishment_rule WHERE id = ?`, id))
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return item, nil
}

// UpdateRule 编辑处罚规则并递增版本，已创建处罚单保留创建时的动作和金额快照。
func (l *DriverPunishmentLogic) UpdateRule(in *adminsvc.DriverPunishmentRuleRequest) (*adminsvc.DriverPunishmentRule, error) {
	if err := requireAdminRoles(l.ctx, l.svcCtx, 1, 2); err != nil {
		return nil, err
	}
	if in.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "处罚规则ID不能为空")
	}
	if err := validateDriverPunishmentRule(in); err != nil {
		return nil, err
	}
	tx, err := l.svcCtx.MySQL.BeginTx(l.ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	res, err := tx.ExecContext(l.ctx, `
		UPDATE admin_driver_punishment_rule
		SET name=?, violation_type=?, actions=?, penalty_cents=?, score_delta=?, priority_weight_delta=?,
		    status=?, version=version+1, updated_by=?
		WHERE id=?`,
		strings.TrimSpace(in.GetName()), strings.TrimSpace(in.GetViolationType()), in.GetActions(), in.GetPenaltyCents(),
		in.GetScoreDelta(), in.GetPriorityWeightDelta(), in.GetStatus(), in.GetAdminId(), in.GetId())
	if err != nil {
		return nil, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}
	if affected == 0 {
		return nil, status.Error(codes.NotFound, "处罚规则不存在")
	}
	if err := createOperationLogTx(l.ctx, tx, in.GetAdminId(), "driver_punishment", "rule_update", "driver_punishment_rule", in.GetId(), "更新司机处罚规则："+strings.TrimSpace(in.GetName()), in.GetIp()); err != nil {
		return nil, err
	}
	item, err := scanDriverPunishmentRule(tx.QueryRowContext(l.ctx, `
		SELECT id, name, violation_type, CAST(actions AS CHAR), penalty_cents, score_delta,
		       priority_weight_delta, status, version, created_by, updated_by, created_at, updated_at
		FROM admin_driver_punishment_rule WHERE id = ?`, in.GetId()))
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return item, nil
}

// SetRuleStatus 启停处罚规则。停用后不可用于新建处罚单，但不会影响既有处罚单。
func (l *DriverPunishmentLogic) SetRuleStatus(in *adminsvc.DriverPunishmentRuleStatusRequest) (*adminsvc.CommonResponse, error) {
	if err := requireAdminRoles(l.ctx, l.svcCtx, 1); err != nil {
		return nil, err
	}
	if in.GetId() <= 0 || (in.GetStatus() != 1 && in.GetStatus() != 2) {
		return nil, status.Error(codes.InvalidArgument, "处罚规则状态参数不合法")
	}
	tx, err := l.svcCtx.MySQL.BeginTx(l.ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	res, err := tx.ExecContext(l.ctx, "UPDATE admin_driver_punishment_rule SET status=?, version=version+1, updated_by=? WHERE id=?", in.GetStatus(), in.GetAdminId(), in.GetId())
	if err != nil {
		return nil, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}
	if affected == 0 {
		return nil, status.Error(codes.NotFound, "处罚规则不存在")
	}
	action := "rule_enable"
	if in.GetStatus() == 2 {
		action = "rule_disable"
	}
	if err := createOperationLogTx(l.ctx, tx, in.GetAdminId(), "driver_punishment", action, "driver_punishment_rule", in.GetId(), "更新司机处罚规则状态", in.GetIp()); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &adminsvc.CommonResponse{Message: "ok"}, nil
}

// ListPunishments 查询处罚单列表和当前异步执行状态。
func (l *DriverPunishmentLogic) ListPunishments(in *adminsvc.DriverPunishmentListRequest) (*adminsvc.DriverPunishmentListResponse, error) {
	where, args := buildPunishmentWhere(in)
	var total int64
	if err := l.svcCtx.MySQL.QueryRowContext(l.ctx, "SELECT COUNT(1) FROM admin_driver_punishment "+where, args...).Scan(&total); err != nil {
		return nil, err
	}
	page, pageSize := normalizePage(in.GetPage()), normalizePageSize(in.GetPageSize())
	queryArgs := append(append([]any{}, args...), pageSize, offset(page, pageSize))
	rows, err := l.svcCtx.MySQL.QueryContext(l.ctx, punishmentSelectSQL+where+" ORDER BY id DESC LIMIT ? OFFSET ?", queryArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*adminsvc.DriverPunishment, 0)
	for rows.Next() {
		item, err := scanDriverPunishment(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, item)
	}
	return &adminsvc.DriverPunishmentListResponse{List: list, Total: total, Page: page, PageSize: pageSize}, rows.Err()
}

// GetPunishment 查询单条处罚单，供后台展示失败原因和撤销入口。
func (l *DriverPunishmentLogic) GetPunishment(in *adminsvc.DriverPunishmentDetailRequest) (*adminsvc.DriverPunishment, error) {
	if in.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "处罚单ID不能为空")
	}
	item, err := scanDriverPunishment(l.svcCtx.MySQL.QueryRowContext(l.ctx, punishmentSelectSQL+" WHERE id=?", in.GetId()))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, status.Error(codes.NotFound, "处罚单不存在")
	}
	return item, err
}

// CreatePunishment 以规则快照或人工参数创建处罚单，并可靠下发给司机域消费者。
func (l *DriverPunishmentLogic) CreatePunishment(in *adminsvc.DriverPunishmentRequest) (*adminsvc.DriverPunishment, error) {
	if err := requireAdminRoles(l.ctx, l.svcCtx, 1, 2); err != nil {
		return nil, err
	}
	if in.GetDriverId() <= 0 || in.GetAdminId() <= 0 || strings.TrimSpace(in.GetReason()) == "" || strings.TrimSpace(in.GetRequestId()) == "" {
		return nil, status.Error(codes.InvalidArgument, "司机、管理员、处罚原因和request_id不能为空")
	}
	tx, err := l.svcCtx.MySQL.BeginTx(l.ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	actions, penaltyCents, ruleID, err := resolvePunishmentSnapshot(l.ctx, tx, in)
	if err != nil {
		return nil, err
	}
	punishmentNo := newAdminTaskNo("DP")
	res, err := tx.ExecContext(l.ctx, `
		INSERT INTO admin_driver_punishment
			(punishment_no, driver_id, order_id, rule_id, actions, reason, penalty_cents, status, request_id, created_by)
		VALUES (?, ?, ?, ?, ?, ?, ?, 'pending', ?, ?)`,
		punishmentNo, in.GetDriverId(), in.GetOrderId(), ruleID, actions, strings.TrimSpace(in.GetReason()), penaltyCents, strings.TrimSpace(in.GetRequestId()), in.GetAdminId())
	if err != nil {
		if isDuplicateKeyError(err) {
			return l.getPunishmentByRequestID(in.GetDriverId(), in.GetRequestId())
		}
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	eventNo, err := recordDomainOutboxTx(l.ctx, tx, AdminDomainEvent{
		EventType: "admin.driver.punishment.requested", AggregateType: "driver_punishment", AggregateID: id, RequestID: in.GetRequestId(),
		Payload: map[string]any{"punishment_no": punishmentNo, "driver_id": in.GetDriverId(), "order_id": in.GetOrderId(), "actions": json.RawMessage(actions), "penalty_cents": penaltyCents},
	})
	if err != nil {
		return nil, err
	}
	if err := createOperationLogTx(l.ctx, tx, in.GetAdminId(), "driver_punishment", "create", "driver_punishment", id, fmt.Sprintf("创建司机处罚单：%s，event_no=%s", punishmentNo, eventNo), in.GetIp()); err != nil {
		return nil, err
	}
	item, err := scanDriverPunishment(tx.QueryRowContext(l.ctx, punishmentSelectSQL+" WHERE id=?", id))
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return item, nil
}

// CancelPunishment 撤销未生效处罚单，并写入反向事件；已生效处罚只能走申诉复核以避免状态错乱。
func (l *DriverPunishmentLogic) CancelPunishment(in *adminsvc.DriverPunishmentActionRequest) (*adminsvc.CommonResponse, error) {
	if err := requireAdminRoles(l.ctx, l.svcCtx, 1); err != nil {
		return nil, err
	}
	if in.GetId() <= 0 || strings.TrimSpace(in.GetRequestId()) == "" || strings.TrimSpace(in.GetReason()) == "" {
		return nil, status.Error(codes.InvalidArgument, "处罚单ID、撤销原因和request_id不能为空")
	}
	tx, err := l.svcCtx.MySQL.BeginTx(l.ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	item, err := scanDriverPunishment(tx.QueryRowContext(l.ctx, punishmentSelectSQL+" WHERE id=? FOR UPDATE", in.GetId()))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "处罚单不存在")
		}
		return nil, err
	}
	if item.GetStatus() != "pending" && item.GetStatus() != "failed" {
		return nil, status.Error(codes.FailedPrecondition, "当前处罚状态不允许直接撤销")
	}
	res, err := tx.ExecContext(l.ctx, `
		UPDATE admin_driver_punishment
		SET status='cancelled', cancelled_by=?, cancelled_at=NOW(), failure_reason=''
		WHERE id=? AND status IN ('pending','failed')`, in.GetAdminId(), in.GetId())
	if err != nil {
		return nil, err
	}
	affected, err := res.RowsAffected()
	if err != nil || affected == 0 {
		return nil, status.Error(codes.FailedPrecondition, "处罚状态已变化，请刷新后重试")
	}
	eventNo, err := recordDomainOutboxTx(l.ctx, tx, AdminDomainEvent{
		EventType: "admin.driver.punishment.reversed", AggregateType: "driver_punishment", AggregateID: in.GetId(), RequestID: in.GetRequestId(),
		Payload: map[string]any{"punishment_no": item.GetPunishmentNo(), "driver_id": item.GetDriverId(), "reason": strings.TrimSpace(in.GetReason())},
	})
	if err != nil {
		return nil, err
	}
	if err := createOperationLogTx(l.ctx, tx, in.GetAdminId(), "driver_punishment", "cancel", "driver_punishment", in.GetId(), "撤销处罚单，event_no="+eventNo, in.GetIp()); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &adminsvc.CommonResponse{Message: "ok"}, nil
}

// ListAppeals 查询处罚申诉与复核结论。
func (l *DriverPunishmentLogic) ListAppeals(in *adminsvc.PunishmentAppealListRequest) (*adminsvc.PunishmentAppealListResponse, error) {
	where, args := buildAppealWhere(in)
	var total int64
	if err := l.svcCtx.MySQL.QueryRowContext(l.ctx, "SELECT COUNT(1) FROM admin_punishment_appeal "+where, args...).Scan(&total); err != nil {
		return nil, err
	}
	page, pageSize := normalizePage(in.GetPage()), normalizePageSize(in.GetPageSize())
	queryArgs := append(append([]any{}, args...), pageSize, offset(page, pageSize))
	rows, err := l.svcCtx.MySQL.QueryContext(l.ctx, appealSelectSQL+where+" ORDER BY id DESC LIMIT ? OFFSET ?", queryArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*adminsvc.PunishmentAppeal, 0)
	for rows.Next() {
		item, err := scanPunishmentAppeal(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, item)
	}
	return &adminsvc.PunishmentAppealListResponse{List: list, Total: total, Page: page, PageSize: pageSize}, rows.Err()
}

// CreateAppeal 创建司机处罚申诉，并以申诉编号和 request_id 保证重复提交幂等。
func (l *DriverPunishmentLogic) CreateAppeal(in *adminsvc.PunishmentAppealRequest) (*adminsvc.PunishmentAppeal, error) {
	if in.GetPunishmentId() <= 0 || in.GetDriverId() <= 0 || strings.TrimSpace(in.GetContent()) == "" || strings.TrimSpace(in.GetRequestId()) == "" {
		return nil, status.Error(codes.InvalidArgument, "处罚单、司机、申诉内容和request_id不能为空")
	}
	tx, err := l.svcCtx.MySQL.BeginTx(l.ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	var punishmentDriver int64
	var punishmentStatus string
	if err := tx.QueryRowContext(l.ctx, "SELECT driver_id, status FROM admin_driver_punishment WHERE id=? FOR UPDATE", in.GetPunishmentId()).Scan(&punishmentDriver, &punishmentStatus); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "处罚单不存在")
		}
		return nil, err
	}
	if punishmentDriver != in.GetDriverId() {
		return nil, status.Error(codes.InvalidArgument, "处罚单与司机不匹配")
	}
	if punishmentStatus == "cancelled" {
		return nil, status.Error(codes.FailedPrecondition, "已撤销处罚不能申诉")
	}
	appealNo := newAdminTaskNo("PA")
	if _, err := tx.ExecContext(l.ctx, `
		INSERT INTO admin_punishment_appeal
			(appeal_no, punishment_id, driver_id, content, evidence_config, status, request_id)
		VALUES (?, ?, ?, ?, ?, 'pending', ?)`,
		appealNo, in.GetPunishmentId(), in.GetDriverId(), strings.TrimSpace(in.GetContent()), nullableJSON(in.GetEvidenceConfig()), strings.TrimSpace(in.GetRequestId())); err != nil {
		if isDuplicateKeyError(err) {
			return l.getAppealByRequestID(in.GetPunishmentId(), in.GetRequestId())
		}
		return nil, err
	}
	id, err := lastInsertID(tx, l.ctx, "admin_punishment_appeal", appealNo)
	if err != nil {
		return nil, err
	}
	item, err := scanPunishmentAppeal(tx.QueryRowContext(l.ctx, appealSelectSQL+" WHERE id=?", id))
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return item, nil
}

// ReviewAppeal 完成处罚申诉复核。revoked 会在同一事务内将处罚单置为 cancelled 并写反向事件。
func (l *DriverPunishmentLogic) ReviewAppeal(in *adminsvc.PunishmentAppealReviewRequest) (*adminsvc.CommonResponse, error) {
	if err := requireAdminRoles(l.ctx, l.svcCtx, 1, 2); err != nil {
		return nil, err
	}
	if in.GetId() <= 0 || strings.TrimSpace(in.GetRequestId()) == "" || strings.TrimSpace(in.GetReviewResult()) == "" {
		return nil, status.Error(codes.InvalidArgument, "申诉ID、审核结论和request_id不能为空")
	}
	action := strings.TrimSpace(in.GetAction())
	if action != "upheld" && action != "revoked" && action != "rejected" {
		return nil, status.Error(codes.InvalidArgument, "申诉审核动作不合法")
	}
	tx, err := l.svcCtx.MySQL.BeginTx(l.ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	appeal, err := scanPunishmentAppeal(tx.QueryRowContext(l.ctx, appealSelectSQL+" WHERE id=? FOR UPDATE", in.GetId()))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "处罚申诉不存在")
		}
		return nil, err
	}
	if appeal.GetStatus() != "pending" && appeal.GetStatus() != "reviewing" {
		return nil, status.Error(codes.FailedPrecondition, "申诉已完成复核")
	}
	if _, err := tx.ExecContext(l.ctx, `UPDATE admin_punishment_appeal SET status=?, review_result=?, reviewed_by=?, reviewed_at=NOW() WHERE id=? AND status IN ('pending','reviewing')`,
		action, strings.TrimSpace(in.GetReviewResult()), in.GetAdminId(), in.GetId()); err != nil {
		return nil, err
	}
	if action == "revoked" {
		punishment, err := scanDriverPunishment(tx.QueryRowContext(l.ctx, punishmentSelectSQL+" WHERE id=? FOR UPDATE", appeal.GetPunishmentId()))
		if err != nil {
			return nil, err
		}
		if punishment.GetStatus() != "cancelled" {
			if _, err := tx.ExecContext(l.ctx, `UPDATE admin_driver_punishment SET status='cancelled', cancelled_by=?, cancelled_at=NOW() WHERE id=?`, in.GetAdminId(), punishment.GetId()); err != nil {
				return nil, err
			}
			if _, err := recordDomainOutboxTx(l.ctx, tx, AdminDomainEvent{
				EventType: "admin.driver.punishment.reversed", AggregateType: "driver_punishment", AggregateID: punishment.GetId(), RequestID: in.GetRequestId(),
				Payload: map[string]any{"punishment_no": punishment.GetPunishmentNo(), "driver_id": punishment.GetDriverId(), "reason": "处罚申诉复核撤销"},
			}); err != nil {
				return nil, err
			}
		}
	}
	if _, err := recordDomainOutboxTx(l.ctx, tx, AdminDomainEvent{
		EventType: "admin.driver.punishment.appeal.reviewed", AggregateType: "punishment_appeal", AggregateID: appeal.GetId(), RequestID: in.GetRequestId(),
		Payload: map[string]any{"punishment_id": appeal.GetPunishmentId(), "driver_id": appeal.GetDriverId(), "action": action},
	}); err != nil {
		return nil, err
	}
	if err := createOperationLogTx(l.ctx, tx, in.GetAdminId(), "driver_punishment", "appeal_"+action, "punishment_appeal", in.GetId(), "完成处罚申诉复核："+strings.TrimSpace(in.GetReviewResult()), in.GetIp()); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &adminsvc.CommonResponse{Message: "ok"}, nil
}

// getAppealByRequestID 返回重复申诉请求对应的已有记录。
func (l *DriverPunishmentLogic) getAppealByRequestID(punishmentID int64, requestID string) (*adminsvc.PunishmentAppeal, error) {
	item, err := scanPunishmentAppeal(l.svcCtx.MySQL.QueryRowContext(l.ctx, appealSelectSQL+" WHERE punishment_id=? AND request_id=?", punishmentID, requestID))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, status.Error(codes.AlreadyExists, "申诉请求已存在")
	}
	return item, err
}

// lastInsertID 读取指定业务编号刚创建的记录 ID，避免依赖驱动返回值的非一致行为。
func lastInsertID(tx *sql.Tx, ctx context.Context, table, no string) (int64, error) {
	var id int64
	if err := tx.QueryRowContext(ctx, "SELECT id FROM "+table+" WHERE "+map[string]string{"admin_punishment_appeal": "appeal_no", "admin_driver_punishment": "punishment_no"}[table]+"=?", no).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

const punishmentSelectSQL = `SELECT id, punishment_no, driver_id, order_id, rule_id, CAST(actions AS CHAR), reason, penalty_cents, status, request_id, failure_reason, created_by, cancelled_by, cancelled_at, effective_at, created_at, updated_at FROM admin_driver_punishment `
const appealSelectSQL = `SELECT id, appeal_no, punishment_id, driver_id, content, COALESCE(CAST(evidence_config AS CHAR), ''), status, review_result, reviewed_by, reviewed_at, request_id, created_at, updated_at FROM admin_punishment_appeal `

// resolvePunishmentSnapshot 固化处罚规则快照，避免规则日后编辑影响已创建处罚单。
func resolvePunishmentSnapshot(ctx context.Context, tx *sql.Tx, in *adminsvc.DriverPunishmentRequest) (string, int64, int64, error) {
	if in.GetRuleId() <= 0 {
		if err := validatePunishmentActions(in.GetActions()); err != nil {
			return "", 0, 0, err
		}
		if in.GetPenaltyCents() < 0 {
			return "", 0, 0, status.Error(codes.InvalidArgument, "罚款金额不能小于0")
		}
		return in.GetActions(), in.GetPenaltyCents(), 0, nil
	}
	var actions string
	var penaltyCents int64
	err := tx.QueryRowContext(ctx, `SELECT CAST(actions AS CHAR), penalty_cents FROM admin_driver_punishment_rule WHERE id=? AND status=1`, in.GetRuleId()).Scan(&actions, &penaltyCents)
	if errors.Is(err, sql.ErrNoRows) {
		return "", 0, 0, status.Error(codes.FailedPrecondition, "处罚规则不存在或已停用")
	}
	if err != nil {
		return "", 0, 0, err
	}
	return actions, penaltyCents, in.GetRuleId(), nil
}

// validateDriverPunishmentRule 校验处罚规则字段与动作组合。
func validateDriverPunishmentRule(in *adminsvc.DriverPunishmentRuleRequest) error {
	if in == nil || in.GetAdminId() <= 0 || strings.TrimSpace(in.GetName()) == "" || strings.TrimSpace(in.GetViolationType()) == "" {
		return status.Error(codes.InvalidArgument, "处罚规则名称、违规类型和管理员不能为空")
	}
	if in.GetStatus() != 1 && in.GetStatus() != 2 {
		return status.Error(codes.InvalidArgument, "处罚规则状态不合法")
	}
	if in.GetPenaltyCents() < 0 || in.GetScoreDelta() > 0 || in.GetPriorityWeightDelta() > 0 {
		return status.Error(codes.InvalidArgument, "处罚规则数值不合法")
	}
	return validatePunishmentActions(in.GetActions())
}

// validatePunishmentActions 校验动作 JSON 为非空字符串数组，并仅允许第一期动作集合。
func validatePunishmentActions(raw string) error {
	var actions []string
	if err := json.Unmarshal([]byte(raw), &actions); err != nil || len(actions) == 0 {
		return status.Error(codes.InvalidArgument, "处罚动作必须为非空JSON数组")
	}
	seen := make(map[string]struct{}, len(actions))
	for _, action := range actions {
		if _, ok := driverPunishmentActions[action]; !ok {
			return status.Error(codes.InvalidArgument, "处罚动作不合法")
		}
		if _, duplicated := seen[action]; duplicated {
			return status.Error(codes.InvalidArgument, "处罚动作不能重复")
		}
		seen[action] = struct{}{}
	}
	return nil
}

// buildPunishmentRuleWhere 构造规则列表的白名单筛选条件。
func buildPunishmentRuleWhere(in *adminsvc.DriverPunishmentRuleListRequest) (string, []any) {
	parts, args := []string{"WHERE 1=1"}, []any{}
	if keyword := strings.TrimSpace(in.GetKeyword()); keyword != "" {
		parts, args = append(parts, "(name LIKE ? OR violation_type LIKE ?)"), append(args, "%"+keyword+"%", "%"+keyword+"%")
	}
	if violationType := strings.TrimSpace(in.GetViolationType()); violationType != "" {
		parts, args = append(parts, "violation_type=?"), append(args, violationType)
	}
	if in.GetStatus() > 0 {
		parts, args = append(parts, "status=?"), append(args, in.GetStatus())
	}
	return strings.Join(parts, " AND "), args
}

// buildPunishmentWhere 构造处罚单列表的白名单筛选条件。
func buildPunishmentWhere(in *adminsvc.DriverPunishmentListRequest) (string, []any) {
	parts, args := []string{"WHERE 1=1"}, []any{}
	if in.GetDriverId() > 0 {
		parts, args = append(parts, "driver_id=?"), append(args, in.GetDriverId())
	}
	if in.GetOrderId() > 0 {
		parts, args = append(parts, "order_id=?"), append(args, in.GetOrderId())
	}
	if value := strings.TrimSpace(in.GetStatus()); value != "" {
		parts, args = append(parts, "status=?"), append(args, value)
	}
	if value := strings.TrimSpace(in.GetRequestId()); value != "" {
		parts, args = append(parts, "request_id=?"), append(args, value)
	}
	return strings.Join(parts, " AND "), args
}

// buildAppealWhere 构造处罚申诉列表的白名单筛选条件。
func buildAppealWhere(in *adminsvc.PunishmentAppealListRequest) (string, []any) {
	parts, args := []string{"WHERE 1=1"}, []any{}
	if in.GetPunishmentId() > 0 {
		parts, args = append(parts, "punishment_id=?"), append(args, in.GetPunishmentId())
	}
	if in.GetDriverId() > 0 {
		parts, args = append(parts, "driver_id=?"), append(args, in.GetDriverId())
	}
	if value := strings.TrimSpace(in.GetStatus()); value != "" {
		parts, args = append(parts, "status=?"), append(args, value)
	}
	return strings.Join(parts, " AND "), args
}

// scanDriverPunishmentRule 将规则查询行转换为 RPC DTO。
func scanDriverPunishmentRule(scanner interface{ Scan(...any) error }) (*adminsvc.DriverPunishmentRule, error) {
	item := &adminsvc.DriverPunishmentRule{}
	var createdAt, updatedAt sql.NullTime
	if err := scanner.Scan(&item.Id, &item.Name, &item.ViolationType, &item.Actions, &item.PenaltyCents, &item.ScoreDelta, &item.PriorityWeightDelta, &item.Status, &item.Version, &item.CreatedBy, &item.UpdatedBy, &createdAt, &updatedAt); err != nil {
		return nil, err
	}
	item.CreatedAt, item.UpdatedAt = formatNullTime(createdAt), formatNullTime(updatedAt)
	return item, nil
}

// scanDriverPunishment 将处罚单查询行转换为 RPC DTO。
func scanDriverPunishment(scanner interface{ Scan(...any) error }) (*adminsvc.DriverPunishment, error) {
	item := &adminsvc.DriverPunishment{}
	var cancelledAt, effectiveAt, createdAt, updatedAt sql.NullTime
	if err := scanner.Scan(&item.Id, &item.PunishmentNo, &item.DriverId, &item.OrderId, &item.RuleId, &item.Actions, &item.Reason, &item.PenaltyCents, &item.Status, &item.RequestId, &item.FailureReason, &item.CreatedBy, &item.CancelledBy, &cancelledAt, &effectiveAt, &createdAt, &updatedAt); err != nil {
		return nil, err
	}
	item.CancelledAt, item.EffectiveAt, item.CreatedAt, item.UpdatedAt = formatNullTime(cancelledAt), formatNullTime(effectiveAt), formatNullTime(createdAt), formatNullTime(updatedAt)
	return item, nil
}

// scanPunishmentAppeal 将处罚申诉查询行转换为 RPC DTO。
func scanPunishmentAppeal(scanner interface{ Scan(...any) error }) (*adminsvc.PunishmentAppeal, error) {
	item := &adminsvc.PunishmentAppeal{}
	var reviewedAt, createdAt, updatedAt sql.NullTime
	if err := scanner.Scan(&item.Id, &item.AppealNo, &item.PunishmentId, &item.DriverId, &item.Content, &item.EvidenceConfig, &item.Status, &item.ReviewResult, &item.ReviewedBy, &reviewedAt, &item.RequestId, &createdAt, &updatedAt); err != nil {
		return nil, err
	}
	item.ReviewedAt, item.CreatedAt, item.UpdatedAt = formatNullTime(reviewedAt), formatNullTime(createdAt), formatNullTime(updatedAt)
	return item, nil
}

// getPunishmentByRequestID 返回已创建处罚单，供重复 request_id 请求幂等返回。
func (l *DriverPunishmentLogic) getPunishmentByRequestID(driverID int64, requestID string) (*adminsvc.DriverPunishment, error) {
	item, err := scanDriverPunishment(l.svcCtx.MySQL.QueryRowContext(l.ctx, punishmentSelectSQL+" WHERE driver_id=? AND request_id=?", driverID, requestID))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, status.Error(codes.AlreadyExists, "处罚请求已存在")
	}
	return item, err
}

// isDuplicateKeyError 识别 MySQL 唯一键冲突，保证幂等请求返回已有记录而不是重复写入。
func isDuplicateKeyError(err error) bool {
	return err != nil && strings.Contains(strings.ToLower(err.Error()), "duplicate")
}
