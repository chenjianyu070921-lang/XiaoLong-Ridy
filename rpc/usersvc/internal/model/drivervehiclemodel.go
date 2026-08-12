package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ DriverVehicleModel = (*customDriverVehicleModel)(nil)

type (
	// DriverVehicleModel is an interface to be customized, add more methods here,
	// and implement the added methods in customDriverVehicleModel.
	DriverVehicleModel interface {
		driverVehicleModel
		withSession(session sqlx.Session) DriverVehicleModel
	}

	customDriverVehicleModel struct {
		*defaultDriverVehicleModel
	}
)

// NewDriverVehicleModel returns a model for the database table.
func NewDriverVehicleModel(conn sqlx.SqlConn) DriverVehicleModel {
	return &customDriverVehicleModel{
		defaultDriverVehicleModel: newDriverVehicleModel(conn),
	}
}

func (m *customDriverVehicleModel) withSession(session sqlx.Session) DriverVehicleModel {
	return NewDriverVehicleModel(sqlx.NewSqlConnFromSession(session))
}
