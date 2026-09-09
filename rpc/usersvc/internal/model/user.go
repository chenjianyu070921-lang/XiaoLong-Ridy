package model

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

const (
	UserStatusNormal = 1
	UserStatusFrozen = 2

	RealNameStatusUnverified = "UNVERIFIED"
	RealNameStatusVerified   = "VERIFIED"

	RegisterSourcePhone  = "phone"
	RegisterSourceWechat = "wechat"
	RegisterSourceAlipay = "alipay"
)

var (
	// ErrInvalidPhone 表示用户提交的手机号格式不符合业务校验规则。
	ErrInvalidPhone = errors.New("invalid phone")
	// ErrInvalidSMSCode 表示用户提交的短信验证码与系统记录不一致。
	ErrInvalidSMSCode = errors.New("invalid sms code")
	// ErrSMSCodeExpired 表示短信验证码不存在或者已经超过有效期。
	ErrSMSCodeExpired = errors.New("sms code expired")
	// ErrSMSCodeSendTooFrequent 表示同一手机号仍处于短信发送冷却期。
	ErrSMSCodeSendTooFrequent = errors.New("sms code send too frequent")
	// ErrPhoneAlreadyExists 表示手机号已经关联用户账号，不能重复注册。
	ErrPhoneAlreadyExists = errors.New("phone already registered")
	// ErrUnsupportedRegister 表示当前注册来源不在用户服务支持范围内。
	ErrUnsupportedRegister = errors.New("unsupported register source")
	// ErrAccountFrozen 表示用户账号已被冻结，不能继续执行登录操作。
	ErrAccountFrozen = errors.New("account frozen")
	// ErrInvalidToken 表示用户令牌格式、签名或者服务端状态无效。
	ErrInvalidToken = errors.New("invalid token")
	// ErrTokenExpired 表示用户令牌已经超过有效期。
	ErrTokenExpired = errors.New("token expired")
	// ErrInvalidRealNameInfo 表示实名资料缺少姓名或证件号。
	ErrInvalidRealNameInfo = errors.New("invalid real name info")
	// ErrRealNameVerifyFailed 表示腾讯云实名认证校验未通过或调用失败。
	ErrRealNameVerifyFailed = errors.New("real name verification failed")
)

// User 对应 user 表：保存乘客账号、认证状态和基础资料。
type User struct {
	ID             uint64         `gorm:"primaryKey;column:id" json:"id"`
	Phone          string         `gorm:"column:phone;size:20" json:"phone"`
	PasswordHash   string         `gorm:"column:password_hash;size:255" json:"passwordHash"`
	Nickname       string         `gorm:"column:nickname;size:50" json:"nickname"`
	AvatarURL      string         `gorm:"column:avatar_url;size:255" json:"avatarUrl"`
	Gender         int            `gorm:"column:gender;default:0" json:"gender"`
	RealName       string         `gorm:"column:real_name;size:50" json:"realName"`
	IDCardNo       string         `gorm:"column:id_card_no;size:32" json:"idCardNo"`
	RegisterSource string         `gorm:"column:register_source;size:20" json:"registerSource"`
	Status         int            `gorm:"column:status;default:1" json:"status"`
	CreatedAt      time.Time      `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt      time.Time      `gorm:"column:updated_at" json:"updatedAt"`
	DeletedAt      gorm.DeletedAt `gorm:"column:deleted_at;index" json:"deletedAt"`
}

// TableName 返回用户模型对应的数据表名称，供 GORM 映射 user 表。
func (User) TableName() string {
	return "user"
}
