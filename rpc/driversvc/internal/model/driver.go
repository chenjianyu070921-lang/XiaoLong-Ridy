package model

import "time"

// Driver 对应 driver 表：司机账号表。
type Driver struct {
	Id              uint64     `gorm:"primaryKey;column:id" json:"id"`
	Phone           string     `gorm:"column:phone;size:20" json:"phone"`
	PasswordHash    string     `gorm:"column:password_hash;size:255;default:''" json:"passwordHash"`
	RealName        string     `gorm:"column:real_name;size:50;default:''" json:"realName"`
	IdCardNo        string     `gorm:"column:id_card_no;size:30;default:''" json:"idCardNo"`
	DriverLicenseNo string     `gorm:"column:driver_license_no;size:30;default:''" json:"driverLicenseNo"`
	AvatarUrl       string     `gorm:"column:avatar_url;size:255;default:''" json:"avatarUrl"`
	Status          int8       `gorm:"column:status;default:1" json:"status"`
	CreatedAt       time.Time  `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt       time.Time  `gorm:"column:updated_at" json:"updatedAt"`
	DeletedAt       *time.Time `gorm:"column:deleted_at" json:"deletedAt"`
}

// TableName 返回对应的数据库表名。
func (Driver) TableName() string {
	return "driver"
}
