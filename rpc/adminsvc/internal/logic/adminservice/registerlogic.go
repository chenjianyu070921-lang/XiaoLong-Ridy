package adminservicelogic

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"XiaoLong-Ridy/rpc/adminsvc/adminsvc"
	"XiaoLong-Ridy/rpc/adminsvc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// adminRegisterLockName 是保护首个管理员注册临界区的 MySQL 命名锁名称。
// 命名锁在同一 MySQL 实例内全局生效，即使 adminsvc 多实例部署也能互斥。
const adminRegisterLockName = "admin_register_bootstrap"

// acquireBootstrapLock 获取首个管理员注册临界区的命名锁。
// 命名锁绑定在连接上，必须使用专用 *sql.Conn 持有，避免连接归还连接池后
// 被其他查询复用导致锁会话错位。锁等待上限 5 秒，超时返回系统繁忙。
func acquireBootstrapLock(ctx context.Context, svcCtx *svc.ServiceContext) (*sql.Conn, error) {
	conn, err := svcCtx.MySQL.Conn(ctx)
	if err != nil {
		return nil, err
	}
	var acquired int
	if err := conn.QueryRowContext(ctx, `SELECT GET_LOCK(?, 5)`, adminRegisterLockName).Scan(&acquired); err != nil {
		_ = conn.Close()
		return nil, err
	}
	if acquired != 1 {
		_ = conn.Close()
		return nil, status.Error(codes.ResourceExhausted, "系统繁忙，请稍后重试")
	}
	return conn, nil
}

// releaseBootstrapLock 释放注册临界区命名锁并归还专用连接。
// 释放失败仅记录日志，不影响注册结果；连接关闭由 defer 保证执行。
func releaseBootstrapLock(ctx context.Context, conn *sql.Conn) {
	if conn == nil {
		return
	}
	defer func() { _ = conn.Close() }()
	if _, err := conn.ExecContext(ctx, `SELECT RELEASE_LOCK(?)`, adminRegisterLockName); err != nil {
		logx.Errorf("release admin register lock: %v", err)
	}
}

// RegisterLogic 处理管理员注册 RPC。
type RegisterLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewRegisterLogic 创建管理员注册逻辑对象。
func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Register 创建管理员账号；首个管理员允许直接注册，后续注册要求操作者为超级管理员。
func (l *RegisterLogic) Register(in *adminsvc.RegisterRequest) (*adminsvc.AuthResponse, error) {
	if strings.TrimSpace(in.GetUsername()) == "" || strings.TrimSpace(in.GetPassword()) == "" {
		return nil, status.Error(codes.InvalidArgument, "用户名和密码不能为空")
	}
	// 首个管理员免登录注册依赖"管理员总数为 0"这一条件，存在并发双注册竞态：
	// 两个请求可能同时读到 count=0 并各自创建账号。用命名锁串行化
	// "计数检查 + 插入"临界区，保证全库仅能产生一次免登录引导注册。
	bootstrapConn, err := acquireBootstrapLock(l.ctx, l.svcCtx)
	if err != nil {
		return nil, err
	}
	defer releaseBootstrapLock(l.ctx, bootstrapConn)

	count, err := countAdmins(l.ctx, l.svcCtx)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		// 已存在管理员时必须从 metadata 中复核操作者会话；不能信任客户端填写的 operator_role。
		operator, err := ValidateAdminTokenFromContext(l.ctx, l.svcCtx)
		if err != nil {
			return nil, err
		}
		if err := validateRegisterOperator(operator, in.GetOperatorAdminId()); err != nil {
			return nil, err
		}
	}
	role := in.GetRole()
	if role == 0 {
		role = 2
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(in.GetPassword()), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	res, err := l.svcCtx.MySQL.ExecContext(l.ctx, `
		INSERT INTO admin_user (username, password_hash, real_name, role, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, 1, ?, ?)
	`, in.GetUsername(), string(hash), in.GetRealName(), role, now, now)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	admin := &adminRow{ID: id, Username: in.GetUsername(), RealName: in.GetRealName(), Role: role, Status: 1}
	token, err := randomToken()
	if err != nil {
		return nil, err
	}
	if err := saveSession(l.ctx, l.svcCtx, adminSession{AdminID: id, Username: admin.Username, RealName: admin.RealName, Role: role, Status: 1, Token: token}); err != nil {
		return nil, err
	}
	operatorID := in.GetOperatorAdminId()
	if operatorID == 0 {
		operatorID = id
	}
	_ = createOperationLog(l.ctx, l.svcCtx, operatorID, "auth", "register", "admin_user", id, "管理员注册", "")
	return &adminsvc.AuthResponse{
		Token:     token,
		ExpiresIn: int64(l.svcCtx.Config.Session.SessionTTLHours) * 3600,
		Admin:     toAdminPB(admin),
	}, nil
}

// validateRegisterOperator 校验新增管理员操作的服务端操作者身份。
// 入参 operator 必须来自已验证的会话，requestOperatorID 为客户端请求中声明的操作者 ID。
// 返回 PermissionDenied 表示角色不足或请求冒用其他管理员身份。
func validateRegisterOperator(operator *adminRow, requestOperatorID int64) error {
	if operator == nil || operator.ID <= 0 || operator.ID != requestOperatorID || operator.Role != 1 {
		return status.Error(codes.PermissionDenied, "只有超级管理员可以新增管理员")
	}
	return nil
}
