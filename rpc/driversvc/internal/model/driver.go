package model

import (
	"time"

	"gorm.io/gorm"
)

// Driver 对应 driver 表：司机账号表。
type Driver struct {
	Id              uint64         `gorm:"primaryKey;column:id" json:"id"`                                     // 司机 ID
	Phone           string         `gorm:"column:phone;size:20" json:"phone"`                                  // 登录手机号
	PasswordHash    string         `gorm:"column:password_hash;size:255;default:''" json:"passwordHash"`       // bcrypt 密码哈希
	RealName        string         `gorm:"column:real_name;size:50;default:''" json:"realName"`                // 真实姓名
	IdCardNo        string         `gorm:"column:id_card_no;size:30;default:''" json:"idCardNo"`               // 身份证号码
	DriverLicenseNo string         `gorm:"column:driver_license_no;size:30;default:''" json:"driverLicenseNo"` // 驾驶证号码
	AvatarUrl       string         `gorm:"column:avatar_url;size:255;default:''" json:"avatarUrl"`             // 头像地址
	Status          int8           `gorm:"column:status;default:1" json:"status"`                              // 账号状态：1待审核、2正常、3冻结、4注销
	OnlineStatus    int8           `gorm:"column:online_status;default:0" json:"onlineStatus"`                 // 在线状态：0离线、1在线、2行程中
	CreatedAt       time.Time      `gorm:"column:created_at" json:"createdAt"`                                 // 创建时间
	UpdatedAt       time.Time      `gorm:"column:updated_at" json:"updatedAt"`                                 // 更新时间
	DeletedAt       gorm.DeletedAt `gorm:"column:deleted_at" json:"deletedAt"`                                 // 软删除时间
}

// TableName 返回对应的数据库表名。
func (Driver) TableName() string {
	return "driver"
}
