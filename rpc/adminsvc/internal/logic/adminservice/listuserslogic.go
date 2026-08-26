package adminservicelogic

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"XiaoLong-Ridy/rpc/adminsvc/adminsvc"
	"XiaoLong-Ridy/rpc/adminsvc/internal/svc"
	usersvc "XiaoLong-Ridy/rpc/usersvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

// ListUsersLogic 处理用户列表查询 RPC。
type ListUsersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewListUsersLogic 创建用户列表查询逻辑对象。
func NewListUsersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListUsersLogic {
	return &ListUsersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ListUsers 按分页和筛选条件查询乘客用户列表。
func (l *ListUsersLogic) ListUsers(in *adminsvc.UserListRequest) (*adminsvc.UserListResponse, error) {
	canViewSensitive := canViewUserSensitive(l.ctx, l.svcCtx)
	// usersvc 当前支持状态分页查询；关键字和时间范围继续走兼容查询，避免丢失后台筛选能力。
	if l.svcCtx != nil && l.svcCtx.UsersSvc != nil && userListCanUseUsersRPC(in) {
		return l.listUsersFromRPC(in, canViewSensitive)
	}

	where, args := buildUserWhere(in)
	var total int64
	if err := l.svcCtx.MySQL.QueryRowContext(l.ctx, `SELECT COUNT(1) FROM user `+where, args...).Scan(&total); err != nil {
		return nil, err
	}
	limit := normalizePageSize(in.GetPageSize())
	queryArgs := append(args, limit, offset(in.GetPage(), in.GetPageSize()))
	rows, err := l.svcCtx.MySQL.QueryContext(l.ctx, `
		SELECT id, phone, nickname, avatar_url, gender, real_name, id_card_no,
		       register_source, status, created_at, updated_at
		FROM user `+where+`
		ORDER BY id DESC
		LIMIT ? OFFSET ?
	`, queryArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]*adminsvc.User, 0)
	for rows.Next() {
		item, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, filterAdminUserSensitive(item, canViewSensitive))
	}
	return &adminsvc.UserListResponse{List: list, Total: total, Page: normalizePage(in.GetPage()), PageSize: limit}, rows.Err()
}

// userListCanUseUsersRPC 判断筛选条件是否可由 usersvc.AdminListUsers 表达。
func userListCanUseUsersRPC(in *adminsvc.UserListRequest) bool {
	if in == nil {
		return false
	}
	return strings.TrimSpace(in.GetKeyword()) == "" && strings.TrimSpace(in.GetStartTime()) == "" && strings.TrimSpace(in.GetEndTime()) == ""
}

// listUsersFromRPC 通过 usersvc 查询真实用户列表并转换为后台协议结构。
func (l *ListUsersLogic) listUsersFromRPC(in *adminsvc.UserListRequest, canViewSensitive bool) (*adminsvc.UserListResponse, error) {
	resp, err := l.svcCtx.UsersSvc.AdminListUsers(l.ctx, &usersvc.AdminUserListRequest{Page: in.GetPage(), PageSize: in.GetPageSize(), Status: in.GetStatus()})
	if err != nil {
		return nil, err
	}
	list := make([]*adminsvc.User, 0, len(resp.GetList()))
	for _, item := range resp.GetList() {
		list = append(list, filterAdminUserSensitive(adminUserFromUsersRPC(item), canViewSensitive))
	}
	return &adminsvc.UserListResponse{List: list, Total: resp.GetTotal(), Page: resp.GetPage(), PageSize: resp.GetPageSize()}, nil
}

// adminUserFromUsersRPC 将 usersvc 后台用户对象转换为 adminsvc 响应对象。
// adminsvc 是后台聚合边界，统一在这里做协议转换，敏感字段是否脱敏由调用入口根据管理员会话决定。
func adminUserFromUsersRPC(item *usersvc.AdminUser) *adminsvc.User {
	if item == nil {
		return nil
	}
	return &adminsvc.User{
		Id:             int64(item.GetId()),
		Phone:          item.GetPhone(),
		Nickname:       item.GetNickname(),
		AvatarUrl:      item.GetAvatarUrl(),
		Gender:         item.GetGender(),
		RealName:       item.GetRealName(),
		IdCardNo:       item.GetIdCardNo(),
		RegisterSource: item.GetRegisterSource(),
		Status:         item.GetStatus(),
		CreatedAt:      unixText(item.GetCreatedAt()),
		UpdatedAt:      unixText(item.GetUpdatedAt()),
	}
}

// canViewUserSensitive 根据管理员真实会话判断是否允许查看完整手机号和身份证号。
// 权限设计当前固定角色为 1=超级管理员、2=运营、3=客服；只有超管和运营可查看完整敏感信息。
// 若调用方缺少 token、会话失效、Redis 不可用或账号状态异常，统一按无权限处理，保证读接口不会泄露明文。
func canViewUserSensitive(ctx context.Context, svcCtx *svc.ServiceContext) bool {
	if svcCtx == nil || svcCtx.Redis == nil || svcCtx.MySQL == nil {
		return false
	}
	admin, err := ValidateAdminTokenFromContext(ctx, svcCtx)
	if err != nil || admin == nil {
		return false
	}
	return admin.Role == 1 || admin.Role == 2
}

// filterAdminUserSensitive 按管理员权限裁剪后台用户敏感信息。
// canViewSensitive 为 true 时保留 usersvc 或兼容 SQL 返回的完整值；否则返回服务端脱敏值。
func filterAdminUserSensitive(item *adminsvc.User, canViewSensitive bool) *adminsvc.User {
	if item == nil {
		return nil
	}
	if canViewSensitive {
		return item
	}
	item.Phone = maskPhone(item.Phone)
	item.IdCardNo = maskIDCard(item.IdCardNo)
	return item
}

// maskPhone 脱敏中国大陆手机号，保留前三后四位便于客服核对。
func maskPhone(phone string) string {
	phone = strings.TrimSpace(phone)
	if len(phone) < 7 {
		return phone
	}
	return phone[:3] + "****" + phone[len(phone)-4:]
}

// maskIDCard 脱敏身份证号，保留前六后四位用于人工识别。
func maskIDCard(idCardNo string) string {
	idCardNo = strings.TrimSpace(idCardNo)
	if len(idCardNo) <= 10 {
		return idCardNo
	}
	return idCardNo[:6] + strings.Repeat("*", len(idCardNo)-10) + idCardNo[len(idCardNo)-4:]
}

// buildUserWhere 组装用户列表筛选条件。
func buildUserWhere(in *adminsvc.UserListRequest) (string, []any) {
	parts := []string{"deleted_at IS NULL"}
	args := make([]any, 0)
	if in.GetKeyword() != "" {
		parts = append(parts, "(phone LIKE ? OR nickname LIKE ? OR real_name LIKE ?)")
		kw := "%" + in.GetKeyword() + "%"
		args = append(args, kw, kw, kw)
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

// scanUser 将用户查询结果转换为 protobuf 用户对象。
func scanUser(rows *sql.Rows) (*adminsvc.User, error) {
	var item adminsvc.User
	var createdAt, updatedAt sql.NullTime
	if err := rows.Scan(
		&item.Id, &item.Phone, &item.Nickname, &item.AvatarUrl, &item.Gender,
		&item.RealName, &item.IdCardNo, &item.RegisterSource, &item.Status,
		&createdAt, &updatedAt,
	); err != nil {
		return nil, fmt.Errorf("scan user row: %w", err)
	}
	item.CreatedAt = formatNullTime(createdAt)
	item.UpdatedAt = formatNullTime(updatedAt)
	return &item, nil
}
