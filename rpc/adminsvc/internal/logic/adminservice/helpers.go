package adminservicelogic

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"XiaoLong-Ridy/rpc/adminsvc/adminsvc"
	adminconfig "XiaoLong-Ridy/rpc/adminsvc/internal/config"
	"XiaoLong-Ridy/rpc/adminsvc/internal/svc"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// adminSession 是保存到 Redis 的管理员登录会话。
type adminSession struct {
	AdminID  int64  `json:"admin_id"`
	Username string `json:"username"`
	RealName string `json:"real_name"`
	Role     int32  `json:"role"`
	Status   int32  `json:"status"`
	Token    string `json:"token"`
}

// adminRow 表示 admin_user 表的管理员查询结果。
type adminRow struct {
	ID           int64
	Username     string
	PasswordHash string
	RealName     string
	Role         int32
	Status       int32
}

// normalizePage 统一页码默认值。
func normalizePage(page int32) int32 {
	if page <= 0 {
		return 1
	}
	return page
}

// normalizePageSize 统一限制分页大小，避免一次拉取过多数据。
func normalizePageSize(pageSize int32) int32 {
	if pageSize <= 0 {
		return 20
	}
	if pageSize > 100 {
		return 100
	}
	return pageSize
}

// offset 根据页码和分页大小计算 SQL offset。
func offset(page, pageSize int32) int32 {
	return (normalizePage(page) - 1) * normalizePageSize(pageSize)
}

// formatTime 将时间格式化为统一字符串。
func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
}

// formatNullTime 格式化数据库可空时间。
func formatNullTime(t sql.NullTime) string {
	if !t.Valid {
		return ""
	}
	return formatTime(t.Time)
}

// randomToken 生成管理员登录 token。
func randomToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

// sessionKey 统一生成 Redis 会话 key。
func sessionKey(svcCtx *svc.ServiceContext, token string) string {
	prefix := svcCtx.Config.Session.TokenPrefix
	if prefix == "" {
		prefix = "admin:sess:"
	}
	return prefix + token
}

// saveSession 将管理员会话写入 Redis。
func saveSession(ctx context.Context, svcCtx *svc.ServiceContext, sess adminSession) error {
	b, err := json.Marshal(sess)
	if err != nil {
		return fmt.Errorf("marshal session: %w", err)
	}
	ttlHours := svcCtx.Config.Session.SessionTTLHours
	if ttlHours <= 0 {
		ttlHours = 24
	}
	return svcCtx.Redis.Set(ctx, sessionKey(svcCtx, sess.Token), b, time.Duration(ttlHours)*time.Hour).Err()
}

// readSession 从 Redis 中读取并解析管理员会话。
// 该函数是 adminsvc 对外承担会话边界的统一入口，HTTP 层不再直接访问 Redis。
func readSession(ctx context.Context, svcCtx *svc.ServiceContext, token string) (*adminSession, error) {
	if strings.TrimSpace(token) == "" {
		return nil, status.Error(codes.Unauthenticated, "token不能为空")
	}
	raw, err := svcCtx.Redis.Get(ctx, sessionKey(svcCtx, token)).Result()
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "登录会话无效或已过期")
	}
	var sess adminSession
	if err := json.Unmarshal([]byte(raw), &sess); err != nil {
		return nil, status.Error(codes.Unauthenticated, "登录会话格式异常")
	}
	if sess.AdminID <= 0 {
		return nil, status.Error(codes.Unauthenticated, "登录会话缺少管理员信息")
	}
	return &sess, nil
}

// validateSession 读取 Redis 会话后回查 admin_user，确保账号仍存在且未停用。
// 这样管理员被停用后，旧 token 在下一次请求时会立即失效。
func validateSession(ctx context.Context, svcCtx *svc.ServiceContext, token string) (*adminRow, error) {
	sess, err := readSession(ctx, svcCtx, token)
	if err != nil {
		return nil, err
	}
	admin, err := getAdminByID(ctx, svcCtx, sess.AdminID)
	if err != nil {
		return nil, err
	}
	if admin.Status != 1 {
		return nil, status.Error(codes.PermissionDenied, "管理员账号已停用")
	}
	return admin, nil
}

// deleteSession 删除 Redis 中的管理员会话。
func deleteSession(ctx context.Context, svcCtx *svc.ServiceContext, token string) error {
	if strings.TrimSpace(token) == "" {
		return status.Error(codes.Unauthenticated, "token不能为空")
	}
	return svcCtx.Redis.Del(ctx, sessionKey(svcCtx, token)).Err()
}

// countAdmins 统计有效管理员数量，用于判断是否允许首次注册。
func countAdmins(ctx context.Context, svcCtx *svc.ServiceContext) (int64, error) {
	var count int64
	err := svcCtx.MySQL.QueryRowContext(ctx, `SELECT COUNT(1) FROM admin_user WHERE deleted_at IS NULL`).Scan(&count)
	return count, err
}

// getAdminByUsername 按用户名查询管理员。
func getAdminByUsername(ctx context.Context, svcCtx *svc.ServiceContext, username string) (*adminRow, error) {
	row := svcCtx.MySQL.QueryRowContext(ctx, `
		SELECT id, username, password_hash, real_name, role, status
		FROM admin_user
		WHERE username = ? AND deleted_at IS NULL
	`, username)
	return scanAdmin(row)
}

// getAdminByID 按管理员 ID 查询管理员。
func getAdminByID(ctx context.Context, svcCtx *svc.ServiceContext, id int64) (*adminRow, error) {
	row := svcCtx.MySQL.QueryRowContext(ctx, `
		SELECT id, username, password_hash, real_name, role, status
		FROM admin_user
		WHERE id = ? AND deleted_at IS NULL
	`, id)
	return scanAdmin(row)
}

// scanAdmin 将 SQL 查询结果转换为管理员结构。
func scanAdmin(row *sql.Row) (*adminRow, error) {
	var item adminRow
	if err := row.Scan(&item.ID, &item.Username, &item.PasswordHash, &item.RealName, &item.Role, &item.Status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "管理员不存在")
		}
		return nil, err
	}
	return &item, nil
}

// toAdminPB 将管理员结构转换为 protobuf 响应对象。
func toAdminPB(item *adminRow) *adminsvc.Admin {
	if item == nil {
		return nil
	}
	return &adminsvc.Admin{
		Id:       item.ID,
		Username: item.Username,
		RealName: item.RealName,
		Role:     item.Role,
		Status:   item.Status,
	}
}

// createOperationLog 写入后台操作日志。
func createOperationLog(ctx context.Context, svcCtx *svc.ServiceContext, adminID int64, module, action, targetType string, targetID int64, detail, ip string) error {
	_, err := svcCtx.MySQL.ExecContext(ctx, `
		INSERT INTO admin_operation_log (admin_id, module, action, target_type, target_id, detail, ip)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, adminID, module, action, targetType, targetID, detail, ip)
	return err
}

// createOperationLogTx 在业务事务内写入后台操作日志，保证同库业务变更与审计记录原子提交。
func createOperationLogTx(ctx context.Context, tx *sql.Tx, adminID int64, module, action, targetType string, targetID int64, detail, ip string) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO admin_operation_log (admin_id, module, action, target_type, target_id, detail, ip)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, adminID, module, action, targetType, targetID, detail, ip)
	return err
}

// writeAuditAfterCommitted 为已提交或跨服务业务操作补写审计日志。
// 审计失败但补偿任务创建成功时仍返回 nil，避免调用方把已成功的业务动作误判为失败。
func writeAuditAfterCommitted(ctx context.Context, svcCtx *svc.ServiceContext, adminID int64, module, action, targetType string, targetID int64, detail, ip string) error {
	if err := createOperationLog(ctx, svcCtx, adminID, module, action, targetType, targetID, detail, ip); err != nil {
		if outboxErr := recordAuditOutbox(ctx, svcCtx, adminID, module, action, targetType, targetID, detail, ip, err); outboxErr != nil {
			return fmt.Errorf("业务操作已成功，但审计日志和补偿任务均写入失败: audit=%w; outbox=%v", err, outboxErr)
		}
	}
	return nil
}

// recordAuditOutbox 写入审计补偿任务。
// 跨服务操作已经提交但本地审计日志写入失败时，adminsvc 用该表记录待补偿事件，
// 后续可由 job 或 MQ 消费者重放到 admin_operation_log 或独立 auditsvc。
func recordAuditOutbox(ctx context.Context, svcCtx *svc.ServiceContext, adminID int64, module, action, targetType string, targetID int64, detail, ip string, cause error) error {
	eventNo := newAdminTaskNo("AO")
	failureReason := ""
	if cause != nil {
		failureReason = cause.Error()
	}
	_, err := svcCtx.MySQL.ExecContext(ctx, `
		INSERT INTO admin_audit_outbox
			(event_no, module, action, target_type, target_id, admin_id, detail, ip, status, failure_reason)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'pending', ?)
	`, eventNo, module, action, targetType, targetID, adminID, detail, ip, failureReason)
	return err
}

// recordOperationOrOutbox 优先写操作审计；跨服务业务已提交时，审计失败改写补偿任务。
func recordOperationOrOutbox(ctx context.Context, svcCtx *svc.ServiceContext, adminID int64, module, action, targetType string, targetID int64, detail, ip string) error {
	if err := createOperationLog(ctx, svcCtx, adminID, module, action, targetType, targetID, detail, ip); err != nil {
		return recordAuditOutbox(ctx, svcCtx, adminID, module, action, targetType, targetID, detail, ip, err)
	}
	return nil
}

// mapMenus 将服务配置中声明的角色菜单转换为 RPC 返回结构。
// 未配置角色不返回菜单，避免因默认菜单导致前端暴露未授权入口。
func mapMenus(role int32, menuRoles map[int32][]adminconfig.MenuItemConfig) []*adminsvc.MenuItem {
	return mapMenuItems(menuRoles[role])
}

// mapMenuItems 递归转换菜单配置，并构造新的返回对象，防止调用方修改影响服务配置。
func mapMenuItems(configs []adminconfig.MenuItemConfig) []*adminsvc.MenuItem {
	items := make([]*adminsvc.MenuItem, 0, len(configs))
	for _, item := range configs {
		items = append(items, &adminsvc.MenuItem{
			Name:     item.Name,
			Path:     item.Path,
			Icon:     item.Icon,
			Perm:     item.Perm,
			Children: mapMenuItems(item.Children),
		})
	}
	return items
}
