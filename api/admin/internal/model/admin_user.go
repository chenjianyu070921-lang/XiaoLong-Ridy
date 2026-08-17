package model

import "time"

// AdminUser 对应 admin_user 表。
// 该表保存后台管理员账号、密码哈希、角色、状态和登录时间。
type AdminUser struct {
	ID           int64      `db:"id" json:"id"`
	Username     string     `db:"username" json:"username"`
	PasswordHash string     `db:"password_hash" json:"-"`
	RealName     string     `db:"real_name" json:"real_name"`
	Role         int32      `db:"role" json:"role"`
	Status       int32      `db:"status" json:"status"`
	LastLoginAt  *time.Time `db:"last_login_at" json:"last_login_at,omitempty"`
	CreatedAt    time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at" json:"updated_at"`
	DeletedAt    *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`
}

const (
	// AdminRoleSuper 表示超级管理员，拥有管理员注册等高权限操作能力。
	AdminRoleSuper int32 = 1
	// AdminRoleOps 表示运营人员，可访问大部分运营管理能力。
	AdminRoleOps int32 = 2
	// AdminRoleCS 表示客服人员，主要用于查询用户和订单。
	AdminRoleCS int32 = 3

	// AdminStatusNormal 表示管理员账号正常。
	AdminStatusNormal int32 = 1
	// AdminStatusFrozen 表示管理员账号被禁用。
	AdminStatusFrozen int32 = 2
)

// AdminSession 表示保存在 Redis 中的管理员登录会话。
// 服务端通过 token 查询该结构，从而完成后续接口鉴权。
type AdminSession struct {
	Token     string    `json:"token"`
	AdminID   int64     `json:"admin_id"`
	Username  string    `json:"username"`
	RealName  string    `json:"real_name"`
	Role      int32     `json:"role"`
	ExpiresAt time.Time `json:"expires_at"`
	IssuedAt  time.Time `json:"issued_at"`
}
