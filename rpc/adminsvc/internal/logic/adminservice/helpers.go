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

// deleteSession 删除 Redis 中的管理员会话。
func deleteSession(ctx context.Context, svcCtx *svc.ServiceContext, token string) error {
	if strings.TrimSpace(token) == "" {
		return status.Error(codes.InvalidArgument, "token不能为空")
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

// mapMenus 返回 P0 阶段固定的后台菜单。
func mapMenus(role int32) []*adminsvc.MenuItem {
	items := []*adminsvc.MenuItem{
		{Name: "用户管理", Path: "/users", Icon: "User", Perm: "user:list"},
		{Name: "司机审核", Path: "/driver-certifications", Icon: "BadgeCheck", Perm: "driver:audit"},
		{Name: "订单监控", Path: "/orders", Icon: "ClipboardList", Perm: "order:list"},
		{Name: "操作日志", Path: "/operation-logs", Icon: "ScrollText", Perm: "log:list"},
	}
	if role == 1 {
		items = append([]*adminsvc.MenuItem{{Name: "管理员", Path: "/admins", Icon: "Shield", Perm: "admin:manage"}}, items...)
	}
	return items
}
