package model

import "time"

// DriverListenPreference stores the order types a driver accepts while listening.
type DriverListenPreference struct {
	DriverID          uint64    `gorm:"primaryKey;column:driver_id" json:"driverId"`
	AcceptRealtime    bool      `gorm:"column:accept_realtime;not null;default:1" json:"acceptRealtime"`
	AcceptReservation bool      `gorm:"column:accept_reservation;not null;default:1" json:"acceptReservation"`
	CreatedAt         time.Time `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt         time.Time `gorm:"column:updated_at" json:"updatedAt"`
}

func (DriverListenPreference) TableName() string {
	return "driver_listen_preference"
}