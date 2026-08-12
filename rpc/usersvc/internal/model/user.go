package model

import "time"

const (
	UserStatusNormal = 1
	UserStatusFrozen = 2

	RegisterSourcePhone  = "phone"
	RegisterSourceWechat = "wechat"
	RegisterSourceAlipay = "alipay"
)

// User 表示 usersvc 领域中的乘客账号聚合。
type User struct {
	ID             uint64
	Phone          string
	PasswordHash   string
	Nickname       string
	AvatarURL      string
	Gender         int
	RealName       string
	IDCardNo       string
	RegisterSource string
	Status         int
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      *time.Time
}
