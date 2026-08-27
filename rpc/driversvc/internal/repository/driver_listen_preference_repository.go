package repository

import (
	"context"

	"XiaoLong-Ridy/rpc/driversvc/internal/model"
)

type DriverListenPreferenceRepository interface {
	GetByDriverID(ctx context.Context, driverID uint64) (*model.DriverListenPreference, error)
	Upsert(ctx context.Context, pref *model.DriverListenPreference) error
}