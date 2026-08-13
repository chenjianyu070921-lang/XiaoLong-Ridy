package model

import "time"

// Driver 对应 driver 表：司机账号表，保存司机账号、实名信息和账号状态。
type Driver struct {
	Id              uint64     `gorm:"primaryKey;column:id" json:"id"`                          // Id：司机主键 ID（自增）
	Phone           string     `gorm:"column:phone;size:20" json:"phone"`                       // Phone：手机号，作为登录账号，唯一
	PasswordHash    string     `gorm:"column:password_hash;size:255;default:''" json:"passwordHash"` // PasswordHash：密码哈希值
	RealName        string     `gorm:"column:real_name;size:50;default:''" json:"realName"`     // RealName：司机真实姓名
	IdCardNo        string     `gorm:"column:id_card_no;size:30;default:''" json:"idCardNo"`    // IdCardNo：身份证号，唯一
	DriverLicenseNo string     `gorm:"column:driver_license_no;size:30;default:''" json:"driverLicenseNo"` // DriverLicenseNo：驾驶证号，唯一
	AvatarUrl       string     `gorm:"column:avatar_url;size:255;default:''" json:"avatarUrl"` // AvatarUrl：头像地址
	Status          int8       `gorm:"column:status;default:1" json:"status"`                  // Status：账号状态；1待审核 2正常 3冻结 4注销
	CreatedAt       time.Time  `gorm:"column:created_at" json:"createdAt"`                     // CreatedAt：创建时间
	UpdatedAt       time.Time  `gorm:"column:updated_at" json:"updatedAt"`                     // UpdatedAt：更新时间
	DeletedAt       *time.Time `gorm:"column:deleted_at" json:"deletedAt"`                     // DeletedAt：软删除时间，非空表示已删除
}

// TableName 返回对应的数据库表名。
func (Driver) TableName() string {
	return "driver"
}
