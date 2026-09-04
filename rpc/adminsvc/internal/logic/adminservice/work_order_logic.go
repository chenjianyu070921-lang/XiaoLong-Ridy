package adminservicelogic

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"XiaoLong-Ridy/rpc/adminsvc/adminsvc"
	"XiaoLong-Ridy/rpc/adminsvc/internal/svc"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// WorkOrderLogic 处理后台投诉和申诉工单的创建、查询与流转。
type WorkOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewWorkOrderLogic 创建工单逻辑实例。
func NewWorkOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *WorkOrderLogic {
	return &WorkOrderLogic{ctx: ctx, svcCtx: svcCtx}
}

// CreateWorkOrder 创建待处理工单，并在同一事务内写入创建流转和审计日志。
func (l *WorkOrderLogic) CreateWorkOrder(in *adminsvc.WorkOrderRequest) (*adminsvc.WorkOrder, error) {
	if err := validateWorkOrderRequest(in); err != nil {
		return nil, err
	}
	tx, err := l.svcCtx.MySQL.BeginTx(l.ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	no := newAdminTaskNo("WO")
	res, err := tx.ExecContext(l.ctx, `INSERT INTO admin_complaint_work_order (work_order_no, work_order_type, source_type, source_id, order_id, user_id, driver_id, title, content, priority, status, created_by) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 1, ?)`, no, in.GetWorkOrderType(), in.GetSourceType(), in.GetSourceId(), in.GetOrderId(), in.GetUserId(), in.GetDriverId(), strings.TrimSpace(in.GetTitle()), strings.TrimSpace(in.GetContent()), in.GetPriority(), in.GetAdminId())
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	if _, err = tx.ExecContext(l.ctx, `INSERT INTO admin_work_order_flow (work_order_id, from_status, to_status, action, operator_id, content) VALUES (?, 0, 1, 'create', ?, ?)`, id, in.GetAdminId(), strings.TrimSpace(in.GetContent())); err != nil {
		return nil, err
	}
	if err = createOperationLogTx(l.ctx, tx, in.GetAdminId(), "work_order", "create", "work_order", id, "创建工单："+no, in.GetIp()); err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return l.GetWorkOrder(&adminsvc.WorkOrderDetailRequest{Id: id})
}

// ListWorkOrders 分页查询后台工单。
func (l *WorkOrderLogic) ListWorkOrders(in *adminsvc.WorkOrderListRequest) (*adminsvc.WorkOrderListResponse, error) {
	where, args := " WHERE 1=1", []any{}
	if in.GetStatus() > 0 {
		where += " AND status = ?"
		args = append(args, in.GetStatus())
	}
	if in.GetAssigneeId() > 0 {
		where += " AND assignee_id = ?"
		args = append(args, in.GetAssigneeId())
	}
	if in.GetWorkOrderType() > 0 {
		where += " AND work_order_type = ?"
		args = append(args, in.GetWorkOrderType())
	}
	var total int64
	if err := l.svcCtx.MySQL.QueryRowContext(l.ctx, "SELECT COUNT(1) FROM admin_complaint_work_order"+where, args...).Scan(&total); err != nil {
		return nil, err
	}
	limit := normalizePageSize(in.GetPageSize())
	rows, err := l.svcCtx.MySQL.QueryContext(l.ctx, workOrderSelect+where+" ORDER BY id DESC LIMIT ? OFFSET ?", append(args, limit, offset(in.GetPage(), in.GetPageSize()))...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []*adminsvc.WorkOrder{}
	for rows.Next() {
		item, err := scanWorkOrder(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, item)
	}
	return &adminsvc.WorkOrderListResponse{List: list, Total: total, Page: normalizePage(in.GetPage()), PageSize: limit}, rows.Err()
}

const workOrderSelect = `SELECT id, work_order_no, work_order_type, source_type, source_id, order_id, user_id, driver_id, title, content, priority, status, assignee_id, arbitration_result, remark, version, created_by, created_at, updated_at, closed_at FROM admin_complaint_work_order`

// GetWorkOrder 查询单个工单详情。
func (l *WorkOrderLogic) GetWorkOrder(in *adminsvc.WorkOrderDetailRequest) (*adminsvc.WorkOrder, error) {
	if in.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "工单ID不能为空")
	}
	return scanWorkOrderRow(l.svcCtx.MySQL.QueryRowContext(l.ctx, workOrderSelect+" WHERE id = ?", in.GetId()))
}

// ActWorkOrder 按已定义状态机执行工单动作，并使用 version 防止并发覆盖。
func (l *WorkOrderLogic) ActWorkOrder(in *adminsvc.WorkOrderActionRequest) (*adminsvc.WorkOrder, error) {
	if in.GetId() <= 0 || in.GetAdminId() <= 0 || in.GetVersion() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "工单ID、管理员ID和版本不能为空")
	}
	to, err := workOrderTargetStatus(in)
	if err != nil {
		return nil, err
	}
	// RPC 拦截器只校验方法级权限；这里根据当前会话中的真实角色再次校验具体动作，
	// 防止调用方利用 ActWorkOrder 入口执行超出角色职责范围的状态流转。
	admin, err := ValidateAdminTokenFromContext(l.ctx, l.svcCtx)
	if err != nil {
		return nil, err
	}
	if admin.ID != in.GetAdminId() {
		return nil, status.Error(codes.PermissionDenied, "请求操作者与管理员会话不一致")
	}
	if err := validateWorkOrderActionRole(admin.Role, in.GetAction()); err != nil {
		return nil, err
	}
	tx, err := l.svcCtx.MySQL.BeginTx(l.ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	var from int32
	if err = tx.QueryRowContext(l.ctx, "SELECT status FROM admin_complaint_work_order WHERE id = ? FOR UPDATE", in.GetId()).Scan(&from); err == sql.ErrNoRows {
		return nil, status.Error(codes.NotFound, "工单不存在")
	}
	if err != nil {
		return nil, err
	}
	if !validWorkOrderTransition(from, to, in.GetAction()) {
		return nil, status.Error(codes.FailedPrecondition, "工单状态不允许执行该动作")
	}
	closedAt := any(nil)
	if in.GetAction() == "close" {
		closedAt = time.Now()
	}
	if in.GetAction() == "reopen" {
		closedAt = nil
	}
	res, err := tx.ExecContext(l.ctx, `UPDATE admin_complaint_work_order SET status=?, assignee_id=CASE WHEN ? > 0 THEN ? ELSE assignee_id END, arbitration_result=CASE WHEN ? <> '' THEN ? ELSE arbitration_result END, remark=CASE WHEN ? <> '' THEN ? ELSE remark END, closed_at=?, version=version+1 WHERE id=? AND version=?`, to, in.GetAssigneeId(), in.GetAssigneeId(), in.GetArbitrationResult(), in.GetArbitrationResult(), in.GetContent(), in.GetContent(), closedAt, in.GetId(), in.GetVersion())
	if err != nil {
		return nil, err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return nil, status.Error(codes.Aborted, "工单已被其他管理员更新")
	}
	if _, err = tx.ExecContext(l.ctx, `INSERT INTO admin_work_order_flow (work_order_id, from_status, to_status, action, operator_id, content) VALUES (?, ?, ?, ?, ?, ?)`, in.GetId(), from, to, in.GetAction(), in.GetAdminId(), strings.TrimSpace(in.GetContent())); err != nil {
		return nil, err
	}
	if err = createOperationLogTx(l.ctx, tx, in.GetAdminId(), "work_order", in.GetAction(), "work_order", in.GetId(), fmt.Sprintf("工单状态：%d->%d", from, to), in.GetIp()); err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return l.GetWorkOrder(&adminsvc.WorkOrderDetailRequest{Id: in.GetId()})
}

// BatchActWorkOrders 批量执行工单分配、跟进、仲裁、结案或重开。
// 批量入口不要求前端提交每条记录的 version，但每条工单仍在事务内 FOR UPDATE 加锁读取当前状态，
// 并逐条写入流转记录和审计日志，避免批量操作绕过状态机和审计约束。
func (l *WorkOrderLogic) BatchActWorkOrders(in *adminsvc.WorkOrderBatchActionRequest) (*adminsvc.WorkOrderBatchActionResponse, error) {
	if len(in.GetIds()) == 0 || in.GetAdminId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "工单ID列表和管理员ID不能为空")
	}
	to, err := workOrderTargetStatus(&adminsvc.WorkOrderActionRequest{
		Action:            in.GetAction(),
		AssigneeId:        in.GetAssigneeId(),
		Content:           in.GetContent(),
		ArbitrationResult: in.GetArbitrationResult(),
	})
	if err != nil {
		return nil, err
	}
	admin, err := ValidateAdminTokenFromContext(l.ctx, l.svcCtx)
	if err != nil {
		return nil, err
	}
	if admin.ID != in.GetAdminId() {
		return nil, status.Error(codes.PermissionDenied, "请求操作者与管理员会话不一致")
	}
	if err := validateWorkOrderActionRole(admin.Role, in.GetAction()); err != nil {
		return nil, err
	}
	resp := &adminsvc.WorkOrderBatchActionResponse{}
	for _, id := range uniquePositiveIDs(in.GetIds()) {
		if err := l.batchActOneWorkOrder(id, to, in); err != nil {
			resp.FailCount++
			resp.FailureReasons = append(resp.FailureReasons, fmt.Sprintf("工单%d：%s", id, err.Error()))
			continue
		}
		resp.SuccessCount++
	}
	return resp, nil
}

// batchActOneWorkOrder 在独立事务内处理单个工单，保证批量操作部分失败时不影响其他工单。
func (l *WorkOrderLogic) batchActOneWorkOrder(id int64, to int32, in *adminsvc.WorkOrderBatchActionRequest) error {
	tx, err := l.svcCtx.MySQL.BeginTx(l.ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var from int32
	if err = tx.QueryRowContext(l.ctx, "SELECT status FROM admin_complaint_work_order WHERE id = ? FOR UPDATE", id).Scan(&from); err == sql.ErrNoRows {
		return status.Error(codes.NotFound, "工单不存在")
	}
	if err != nil {
		return err
	}
	if !validWorkOrderTransition(from, to, in.GetAction()) {
		return status.Error(codes.FailedPrecondition, "工单状态不允许执行该动作")
	}
	closedAt := any(nil)
	if in.GetAction() == "close" {
		closedAt = time.Now()
	}
	if _, err = tx.ExecContext(l.ctx, `UPDATE admin_complaint_work_order SET status=?, assignee_id=CASE WHEN ? > 0 THEN ? ELSE assignee_id END, arbitration_result=CASE WHEN ? <> '' THEN ? ELSE arbitration_result END, remark=CASE WHEN ? <> '' THEN ? ELSE remark END, closed_at=?, version=version+1 WHERE id=?`, to, in.GetAssigneeId(), in.GetAssigneeId(), in.GetArbitrationResult(), in.GetArbitrationResult(), in.GetContent(), in.GetContent(), closedAt, id); err != nil {
		return err
	}
	if _, err = tx.ExecContext(l.ctx, `INSERT INTO admin_work_order_flow (work_order_id, from_status, to_status, action, operator_id, content) VALUES (?, ?, ?, ?, ?, ?)`, id, from, to, in.GetAction(), in.GetAdminId(), strings.TrimSpace(in.GetContent())); err != nil {
		return err
	}
	if err = createOperationLogTx(l.ctx, tx, in.GetAdminId(), "work_order", "batch_"+in.GetAction(), "work_order", id, fmt.Sprintf("批量工单状态：%d->%d", from, to), in.GetIp()); err != nil {
		return err
	}
	return tx.Commit()
}

// validateWorkOrderActionRole 校验管理员角色是否可以执行指定工单动作。
// 角色权限严格遵循工单设计：超管拥有全部动作权限，运营仅可分配和跟进，
// 客服仅可跟进；仲裁、结案和重开只能由超级管理员执行。
func validateWorkOrderActionRole(role int32, action string) error {
	switch role {
	case 1:
		return nil
	case 2:
		if action == "assign" || action == "follow" {
			return nil
		}
	case 3:
		if action == "follow" {
			return nil
		}
	}
	return status.Error(codes.PermissionDenied, "当前管理员无权执行该工单动作")
}

// AddWorkOrderEvidence 保存已存在资源或文本的证据索引，不处理文件上传。
func (l *WorkOrderLogic) AddWorkOrderEvidence(in *adminsvc.WorkOrderEvidenceRequest) (*adminsvc.WorkOrderEvidence, error) {
	if in.GetWorkOrderId() <= 0 || in.GetAdminId() <= 0 || !validEvidence(in) {
		return nil, status.Error(codes.InvalidArgument, "工单证据参数不合法")
	}
	res, err := l.svcCtx.MySQL.ExecContext(l.ctx, `INSERT INTO admin_work_order_evidence (work_order_id,evidence_type,evidence_url,content,uploaded_by) SELECT ?,?,?,?,? WHERE EXISTS (SELECT 1 FROM admin_complaint_work_order WHERE id=?)`, in.GetWorkOrderId(), in.GetEvidenceType(), in.GetEvidenceUrl(), in.GetContent(), in.GetAdminId(), in.GetWorkOrderId())
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	if id == 0 {
		return nil, status.Error(codes.NotFound, "工单不存在")
	}
	return l.getEvidence(id)
}

// ListWorkOrderEvidence 分页查询指定工单的证据索引。
func (l *WorkOrderLogic) ListWorkOrderEvidence(in *adminsvc.WorkOrderEvidenceListRequest) (*adminsvc.WorkOrderEvidenceListResponse, error) {
	if in.GetWorkOrderId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "工单ID不能为空")
	}
	var total int64
	if err := l.svcCtx.MySQL.QueryRowContext(l.ctx, "SELECT COUNT(1) FROM admin_work_order_evidence WHERE work_order_id=?", in.GetWorkOrderId()).Scan(&total); err != nil {
		return nil, err
	}
	limit := normalizePageSize(in.GetPageSize())
	rows, err := l.svcCtx.MySQL.QueryContext(l.ctx, "SELECT id,work_order_id,evidence_type,evidence_url,content,uploaded_by,created_at FROM admin_work_order_evidence WHERE work_order_id=? ORDER BY id DESC LIMIT ? OFFSET ?", in.GetWorkOrderId(), limit, offset(in.GetPage(), in.GetPageSize()))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []*adminsvc.WorkOrderEvidence{}
	for rows.Next() {
		item, err := scanEvidence(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, item)
	}
	return &adminsvc.WorkOrderEvidenceListResponse{List: list, Total: total, Page: normalizePage(in.GetPage()), PageSize: limit}, rows.Err()
}

// GetExportDownload 校验下载人身份和文件有效期，只返回受控文件名。
func (l *WorkOrderLogic) GetExportDownload(in *adminsvc.ExportDownloadRequest) (*adminsvc.ExportDownloadResponse, error) {
	if in.GetTaskNo() == "" || in.GetAdminId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "任务编号和管理员ID不能为空")
	}
	var owner int64
	var taskStatus, filePath string
	var expires sql.NullTime
	err := l.svcCtx.MySQL.QueryRowContext(l.ctx, "SELECT admin_id,status,file_path,expires_at FROM admin_export_task WHERE task_no=?", in.GetTaskNo()).Scan(&owner, &taskStatus, &filePath, &expires)
	if err == sql.ErrNoRows {
		return nil, status.Error(codes.NotFound, "导出任务不存在")
	}
	if err != nil {
		return nil, err
	}
	// 下载角色从服务端回查管理员账号真实角色，不信任客户端传入的 admin_role，防止伪造超管角色越权下载。
	role := int32(0)
	if admin, err := getAdminByID(l.ctx, l.svcCtx, in.GetAdminId()); err == nil && admin != nil {
		role = admin.Role
	}
	if role != 1 && owner != in.GetAdminId() {
		return nil, status.Error(codes.PermissionDenied, "无权下载该导出文件")
	}
	if taskStatus != "success" {
		return nil, status.Error(codes.FailedPrecondition, "导出任务尚未完成")
	}
	if !expires.Valid || !expires.Time.After(time.Now()) {
		return nil, status.Error(codes.OutOfRange, "导出文件已过期")
	}
	name := in.GetTaskNo() + ".csv"
	if filepath.Base(filePath) != name {
		return nil, status.Error(codes.FailedPrecondition, "导出文件无效")
	}
	return &adminsvc.ExportDownloadResponse{TaskNo: in.GetTaskNo(), FileName: name, ExpiresAt: expires.Time.Format("2006-01-02 15:04:05")}, nil
}

// OpenAuthorizedExportFile 按已校验下载元数据打开受控导出文件。
// 文件路径只在 adminsvc 进程内使用，调用方只能获得固定任务号对应的 CSV 内容流。
func (l *WorkOrderLogic) OpenAuthorizedExportFile(in *adminsvc.ExportDownloadRequest) (*os.File, *adminsvc.ExportDownloadResponse, error) {
	meta, err := l.GetExportDownload(in)
	if err != nil {
		return nil, nil, err
	}
	file, err := os.Open(filepath.Join(".tmp-admin-exports", meta.GetFileName()))
	if os.IsNotExist(err) {
		return nil, nil, status.Error(codes.NotFound, "导出文件不存在")
	}
	if err != nil {
		return nil, nil, err
	}
	return file, meta, nil
}

func validateWorkOrderRequest(in *adminsvc.WorkOrderRequest) error {
	if in.GetAdminId() <= 0 || in.GetWorkOrderType() < 1 || in.GetWorkOrderType() > 3 || !validSource(in.GetSourceType()) || strings.TrimSpace(in.GetTitle()) == "" || len([]rune(in.GetTitle())) > 100 || in.GetPriority() < 1 || in.GetPriority() > 4 {
		return status.Error(codes.InvalidArgument, "工单参数不合法")
	}
	return nil
}
func validSource(v string) bool { return v == "user" || v == "driver" || v == "order" || v == "system" }
func validEvidence(in *adminsvc.WorkOrderEvidenceRequest) bool {
	switch in.GetEvidenceType() {
	case "track", "audio", "chat", "payment", "image", "text":
	default:
		return false
	}
	return strings.TrimSpace(in.GetEvidenceUrl()) != "" || strings.TrimSpace(in.GetContent()) != ""
}
func workOrderTargetStatus(in *adminsvc.WorkOrderActionRequest) (int32, error) {
	switch in.GetAction() {
	case "assign", "follow", "reopen":
		return 2, nil
	case "arbitrate":
		if strings.TrimSpace(in.GetArbitrationResult()) == "" {
			return 0, status.Error(codes.InvalidArgument, "仲裁结果不能为空")
		}
		return 3, nil
	case "close":
		return 4, nil
	default:
		return 0, status.Error(codes.InvalidArgument, "工单动作不合法")
	}
}
func validWorkOrderTransition(from, to int32, action string) bool {
	switch action {
	case "assign":
		return from == 1 && to == 2
	case "follow":
		return from == 2 && to == 2
	case "arbitrate":
		return from == 2 && to == 3
	case "close":
		return from == 3 && to == 4
	case "reopen":
		return (from == 4 || from == 5) && to == 2
	}
	return false
}
func scanWorkOrder(rows interface{ Scan(...any) error }) (*adminsvc.WorkOrder, error) {
	var i adminsvc.WorkOrder
	var created, updated, closed sql.NullTime
	err := rows.Scan(&i.Id, &i.WorkOrderNo, &i.WorkOrderType, &i.SourceType, &i.SourceId, &i.OrderId, &i.UserId, &i.DriverId, &i.Title, &i.Content, &i.Priority, &i.Status, &i.AssigneeId, &i.ArbitrationResult, &i.Remark, &i.Version, &i.CreatedBy, &created, &updated, &closed)
	if err != nil {
		return nil, err
	}
	i.CreatedAt = formatNullTime(created)
	i.UpdatedAt = formatNullTime(updated)
	i.ClosedAt = formatNullTime(closed)
	return &i, nil
}
func scanWorkOrderRow(row *sql.Row) (*adminsvc.WorkOrder, error) {
	item, err := scanWorkOrder(row)
	if err == sql.ErrNoRows {
		return nil, status.Error(codes.NotFound, "工单不存在")
	}
	return item, err
}
func scanEvidence(rows interface{ Scan(...any) error }) (*adminsvc.WorkOrderEvidence, error) {
	var i adminsvc.WorkOrderEvidence
	var created sql.NullTime
	if err := rows.Scan(&i.Id, &i.WorkOrderId, &i.EvidenceType, &i.EvidenceUrl, &i.Content, &i.UploadedBy, &created); err != nil {
		return nil, err
	}
	i.CreatedAt = formatNullTime(created)
	return &i, nil
}
func (l *WorkOrderLogic) getEvidence(id int64) (*adminsvc.WorkOrderEvidence, error) {
	return scanEvidence(l.svcCtx.MySQL.QueryRowContext(l.ctx, "SELECT id,work_order_id,evidence_type,evidence_url,content,uploaded_by,created_at FROM admin_work_order_evidence WHERE id=?", id))
}
