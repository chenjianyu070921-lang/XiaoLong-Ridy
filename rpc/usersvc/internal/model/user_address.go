package model

import (
	"time"

	"gorm.io/gorm"
)

const (
	// UserAddressTagHome 表示家庭地址。
	UserAddressTagHome = "home"
	// UserAddressTagWork 表示公司地址。
	UserAddressTagWork = "work"
	// UserAddressTagOther 表示其他类型的常用地址。
	UserAddressTagOther = "other"

	// UserAddressNotDefault 表示该地址不是用户的默认地址。
	UserAddressNotDefault int8 = 0
	// UserAddressIsDefault 表示该地址是用户的默认地址。
	UserAddressIsDefault int8 = 1
)

// UserAddress 对应 user_address 表，保存乘客的联系人、常用地址及地理坐标。
type UserAddress struct {
	ID           uint64         `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	UserID       uint64         `gorm:"column:user_id;not null;index:idx_user_id;index:idx_user_tag,priority:1;index:idx_user_default,priority:1" json:"userId"`
	ContactName  string         `gorm:"column:contact_name;size:50;not null;default:''" json:"contactName"`
	ContactPhone string         `gorm:"column:contact_phone;size:20;not null;default:''" json:"contactPhone"`
	Tag          string         `gorm:"column:tag;size:20;not null;default:other;index:idx_user_tag,priority:2" json:"tag"`
	Address      string         `gorm:"column:address;size:255;not null" json:"address"`
	Longitude    float64        `gorm:"column:longitude;type:decimal(10,6);not null" json:"longitude"`
	Latitude     float64        `gorm:"column:latitude;type:decimal(10,6);not null" json:"latitude"`
	IsDefault    int8           `gorm:"column:is_default;not null;default:0;index:idx_user_default,priority:2" json:"isDefault"`
	Sort         int            `gorm:"column:sort;not null;default:0" json:"sort"`
	CreatedAt    time.Time      `gorm:"column:created_at;not null;autoCreateTime" json:"createdAt"`
	UpdatedAt    time.Time      `gorm:"column:updated_at;not null;autoUpdateTime" json:"updatedAt"`
	DeletedAt    gorm.DeletedAt `gorm:"column:deleted_at;index" json:"deletedAt"`
}

// TableName 返回常用地址模型对应的数据表名称，供 GORM 映射 user_address 表。
func (UserAddress) TableName() string {
	return "user_address"
}
