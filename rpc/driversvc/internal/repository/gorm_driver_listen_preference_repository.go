package repository

import (
	"context"

	"XiaoLong-Ridy/rpc/driversvc/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type gormDriverListenPreferenceRepository struct {
	db *gorm.DB
}

func NewGormDriverListenPreferenceRepository(db *gorm.DB) DriverListenPreferenceRepository {
	return &gormDriverListenPreferenceRepository{db: db}
}

func (r *gormDriverListenPreferenceRepository) GetByDriverID(ctx context.Context, driverID uint64) (*model.DriverListenPreference, error) {
	var pref model.DriverListenPreference
	err := r.db.WithContext(ctx).Where("driver_id = ?", driverID).First(&pref).Error
	if err != nil {
		if errorsIsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return &pref, nil
}

func (r *gormDriverListenPreferenceRepository) Upsert(ctx context.Context, pref *model.DriverListenPreference) error {
	if pref == nil {
		return nil
	}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "driver_id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"accept_realtime":    pref.AcceptRealtime,
			"accept_reservation": pref.AcceptReservation,
		}),
	}).Create(pref).Error
}