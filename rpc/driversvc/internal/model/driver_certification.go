package model

import "time"

// DriverCertification 对应 driver_certification 表：司机/车辆资质上传与审核记录。
type DriverCertification struct {
	Id                uint64     `gorm:"primaryKey;column:id" json:"id"`
	DriverId          uint64     `gorm:"column:driver_id" json:"driverId"`
	VehicleId         uint64     `gorm:"column:vehicle_id;default:0" json:"vehicleId"`
	IdCardFrontUrl    string     `gorm:"column:id_card_front_url;size:255;default:''" json:"idCardFrontUrl"`
	IdCardBackUrl     string     `gorm:"column:id_card_back_url;size:255;default:''" json:"idCardBackUrl"`
	DriverLicenseUrl  string     `gorm:"column:driver_license_url;size:255;default:''" json:"driverLicenseUrl"`
	VehicleLicenseUrl string     `gorm:"column:vehicle_license_url;size:255;default:''" json:"vehicleLicenseUrl"`
	AuditStatus       int8       `gorm:"column:audit_status;default:1" json:"auditStatus"`
	AuditRemark       string     `gorm:"column:audit_remark;size:255;default:''" json:"auditRemark"`
	AuditedBy         uint64     `gorm:"column:audited_by;default:0" json:"auditedBy"`
	AuditedAt         *time.Time `gorm:"column:audited_at" json:"auditedAt"`
	CreatedAt         time.Time  `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt         time.Time  `gorm:"column:updated_at" json:"updatedAt"`
}

// TableName 返回对应的数据库表名。
func (DriverCertification) TableName() string {
	return "driver_certification"
}
