package adminservicelogic

import (
	"context"
	"strings"

	"XiaoLong-Ridy/rpc/adminsvc/adminsvc"
	"XiaoLong-Ridy/rpc/adminsvc/internal/svc"

	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AdminManagementLogic 负责超级管理员对后台管理员账号的统一管理。
type AdminManagementLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewAdminManagementLogic 创建管理员管理逻辑对象。
func NewAdminManagementLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminManagementLogic {
	return &AdminManagementLogic{ctx: ctx, svcCtx: svcCtx}
}

// ListAdmins 查询管理员分页列表，仅允许超级管理员调用。
func (l *AdminManagementLogic) ListAdmins(in *adminsvc.AdminListRequest) (*adminsvc.AdminListResponse, error) {
	operator, err := l.requireSuperAdmin()
	if err != nil {
		return nil, err
	}
	page, pageSize := normalizePage(in.GetPage()), normalizePageSize(in.GetPageSize())
	args := []any{}
	where := []string{"deleted_at IS NULL"}
	if keyword := strings.TrimSpace(in.GetKeyword()); keyword != "" {
		where = append(where, "(username LIKE ? OR real_name LIKE ?)")
		like := "%" + keyword + "%"
		args = append(args, like, like)
	}
	if in.GetRole() > 0 {
		where = append(where, "role = ?")
		args = append(args, in.GetRole())
	}
	if in.GetStatus() > 0 {
		where = append(where, "status = ?")
		args = append(args, in.GetStatus())
	}
	whereSQL := strings.Join(where, " AND ")
	var total int64
	if err := l.svcCtx.MySQL.QueryRowContext(l.ctx, "SELECT COUNT(1) FROM admin_user WHERE "+whereSQL, args...).Scan(&total); err != nil {
		return nil, err
	}
	queryArgs := append([]any{}, args...)
	queryArgs = append(queryArgs, int64((page-1)*pageSize), int64(pageSize))
	rows, err := l.svcCtx.MySQL.QueryContext(l.ctx, "SELECT id, username, real_name, role, status FROM admin_user WHERE "+whereSQL+" ORDER BY id DESC LIMIT ?, ?", queryArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]*adminsvc.Admin, 0)
	for rows.Next() {
		var item adminRow
		if err := rows.Scan(&item.ID, &item.Username, &item.RealName, &item.Role, &item.Status); err != nil {
			return nil, err
		}
		items = append(items, toAdminPB(&item))
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	_ = operator
	return &adminsvc.AdminListResponse{List: items, Total: total, Page: page, PageSize: pageSize}, nil
}

// CreateAdmin 新增管理员账号并记录审计日志。
func (l *AdminManagementLogic) CreateAdmin(in *adminsvc.AdminSaveRequest) (*adminsvc.Admin, error) {
	operator, err := l.requireSuperAdmin()
	if err != nil {
		return nil, err
	}
	if err := validateAdminSave(in, true); err != nil {
		return nil, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(in.GetPassword()), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	result, err := l.svcCtx.MySQL.ExecContext(l.ctx, "INSERT INTO admin_user (username,password_hash,real_name,role,status,created_at,updated_at) VALUES (?,?,?,?,1,NOW(),NOW())", in.GetUsername(), string(hash), in.GetRealName(), in.GetRole())
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") {
			return nil, status.Error(codes.AlreadyExists, "管理员账号已存在")
		}
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	admin, err := getAdminByID(l.ctx, l.svcCtx, id)
	if err != nil {
		return nil, err
	}
	if err := createOperationLog(l.ctx, l.svcCtx, operator.ID, "admin", "create", "admin_user", id, "新增管理员", in.GetIp()); err != nil {
		return nil, err
	}
	return toAdminPB(admin), nil
}

// UpdateAdmin 编辑管理员姓名、角色和账号状态以外的资料。
func (l *AdminManagementLogic) UpdateAdmin(in *adminsvc.AdminSaveRequest) (*adminsvc.Admin, error) {
	operator, err := l.requireSuperAdmin()
	if err != nil {
		return nil, err
	}
	if in.GetId() <= 0 || strings.TrimSpace(in.GetRealName()) == "" || in.GetRole() < 1 || in.GetRole() > 3 {
		return nil, status.Error(codes.InvalidArgument, "管理员参数不合法")
	}
	if in.GetId() == operator.ID {
		return nil, status.Error(codes.InvalidArgument, "不能编辑当前超级管理员")
	}
	if _, err := getAdminByID(l.ctx, l.svcCtx, in.GetId()); err != nil {
		return nil, err
	}
	_, err = l.svcCtx.MySQL.ExecContext(l.ctx, "UPDATE admin_user SET real_name=?, role=?, updated_at=NOW() WHERE id=? AND deleted_at IS NULL", in.GetRealName(), in.GetRole(), in.GetId())
	if err != nil {
		return nil, err
	}
	admin, err := getAdminByID(l.ctx, l.svcCtx, in.GetId())
	if err != nil {
		return nil, err
	}
	if err := createOperationLog(l.ctx, l.svcCtx, operator.ID, "admin", "update", "admin_user", in.GetId(), "编辑管理员", in.GetIp()); err != nil {
		return nil, err
	}
	return toAdminPB(admin), nil
}

// SetAdminStatus 启用或停用管理员。
func (l *AdminManagementLogic) SetAdminStatus(in *adminsvc.AdminStatusRequest) (*adminsvc.CommonResponse, error) {
	operator, err := l.requireSuperAdmin()
	if err != nil {
		return nil, err
	}
	if in.GetId() <= 0 || (in.GetStatus() != 1 && in.GetStatus() != 2) {
		return nil, status.Error(codes.InvalidArgument, "管理员状态参数不合法")
	}
	if in.GetId() == operator.ID {
		return nil, status.Error(codes.InvalidArgument, "不能停用当前超级管理员")
	}
	if _, err := getAdminByID(l.ctx, l.svcCtx, in.GetId()); err != nil {
		return nil, err
	}
	if _, err := l.svcCtx.MySQL.ExecContext(l.ctx, "UPDATE admin_user SET status=?, updated_at=NOW() WHERE id=? AND deleted_at IS NULL", in.GetStatus(), in.GetId()); err != nil {
		return nil, err
	}
	if err := createOperationLog(l.ctx, l.svcCtx, operator.ID, "admin", map[int32]string{1: "enable", 2: "disable"}[in.GetStatus()], "admin_user", in.GetId(), in.GetReason(), in.GetIp()); err != nil {
		return nil, err
	}
	return &adminsvc.CommonResponse{Message: "ok"}, nil
}

// ResetAdminPassword 重置指定管理员密码。
func (l *AdminManagementLogic) ResetAdminPassword(in *adminsvc.AdminPasswordResetRequest) (*adminsvc.CommonResponse, error) {
	operator, err := l.requireSuperAdmin()
	if err != nil {
		return nil, err
	}
	if in.GetId() <= 0 || len(strings.TrimSpace(in.GetPassword())) < 6 {
		return nil, status.Error(codes.InvalidArgument, "密码长度不能少于 6 位")
	}
	if in.GetId() == operator.ID {
		return nil, status.Error(codes.InvalidArgument, "不能重置当前超级管理员密码")
	}
	if _, err := getAdminByID(l.ctx, l.svcCtx, in.GetId()); err != nil {
		return nil, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(in.GetPassword()), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	if _, err := l.svcCtx.MySQL.ExecContext(l.ctx, "UPDATE admin_user SET password_hash=?, updated_at=NOW() WHERE id=? AND deleted_at IS NULL", string(hash), in.GetId()); err != nil {
		return nil, err
	}
	if err := createOperationLog(l.ctx, l.svcCtx, operator.ID, "admin", "reset_password", "admin_user", in.GetId(), "重置管理员密码", in.GetIp()); err != nil {
		return nil, err
	}
	return &adminsvc.CommonResponse{Message: "ok"}, nil
}

func (l *AdminManagementLogic) requireSuperAdmin() (*adminRow, error) {
	operator, err := ValidateAdminTokenFromContext(l.ctx, l.svcCtx)
	if err != nil {
		return nil, err
	}
	if operator.Role != 1 {
		return nil, status.Error(codes.PermissionDenied, "只有超级管理员可以管理管理员账号")
	}
	return operator, nil
}

func validateAdminSave(in *adminsvc.AdminSaveRequest, create bool) error {
	if create && strings.TrimSpace(in.GetUsername()) == "" {
		return status.Error(codes.InvalidArgument, "管理员账号不能为空")
	}
	if create && len(strings.TrimSpace(in.GetPassword())) < 6 {
		return status.Error(codes.InvalidArgument, "密码长度不能少于 6 位")
	}
	if strings.TrimSpace(in.GetRealName()) == "" || in.GetRole() < 1 || in.GetRole() > 3 {
		return status.Error(codes.InvalidArgument, "管理员资料参数不合法")
	}
	return nil
}
