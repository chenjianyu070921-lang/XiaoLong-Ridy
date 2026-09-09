package model

import "time"

// DriverVehicle 对应 driver_vehicle 表：司机绑定的车辆信息。
type DriverVehicle struct {
	Id                uint64     `gorm:"primaryKey;column:id" json:"id"`                            // 车辆 ID
	DriverId          uint64     `gorm:"column:driver_id" json:"driverId"`                          // 绑定的司机 ID
	PlateNo           string     `gorm:"column:plate_no;size:20" json:"plateNo"`                    // 车牌号码
	Brand             string     `gorm:"column:brand;size:50;default:''" json:"brand"`              // 车辆品牌
	Model             string     `gorm:"column:model;size:50;default:''" json:"model"`              // 车辆型号
	Color             string     `gorm:"column:color;size:20;default:''" json:"color"`              // 车身颜色
	VehicleType       int8       `gorm:"column:vehicle_type;default:1" json:"vehicleType"`          // 车辆类型：1特惠快车、2快车、3拼车
	RegistrationDate  *time.Time `gorm:"column:registration_date" json:"registrationDate"`          // 注册日期
	InsuranceNo       string     `gorm:"column:insurance_no;size:50;default:''" json:"insuranceNo"` // 保险单号
	InsuranceExpireAt *time.Time `gorm:"column:insurance_expire_at" json:"insuranceExpireAt"`       // 保险到期时间
	Status            int8       `gorm:"column:status;default:1" json:"status"`                     // 状态：1有效、2停用
	CreatedAt         time.Time  `gorm:"column:created_at" json:"createdAt"`                        // 创建时间
	UpdatedAt         time.Time  `gorm:"column:updated_at" json:"updatedAt"`                        // 更新时间
}

// TableName 返回对应的数据库表名。
func (DriverVehicle) TableName() string {
	return "driver_vehicle"
}
