package adminservicelogic

import (
	"context"
	"strings"
	"time"

	"XiaoLong-Ridy/rpc/adminsvc/adminsvc"
	"XiaoLong-Ridy/rpc/adminsvc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

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
		if operator.ID != in.GetOperatorAdminId() || operator.Role != 1 {
			return nil, status.Error(codes.PermissionDenied, "只有超级管理员可以新增管理员")
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
