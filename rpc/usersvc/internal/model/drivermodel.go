package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ DriverModel = (*customDriverModel)(nil)

type (
	// DriverModel is an interface to be customized, add more methods here,
	// and implement the added methods in customDriverModel.
	DriverModel interface {
		driverModel
		withSession(session sqlx.Session) DriverModel
	}

	customDriverModel struct {
		*defaultDriverModel
	}
)

// NewDriverModel returns a model for the database table.
func NewDriverModel(conn sqlx.SqlConn) DriverModel {
	return &customDriverModel{
		defaultDriverModel: newDriverModel(conn),
	}
}

func (m *customDriverModel) withSession(session sqlx.Session) DriverModel {
	return NewDriverModel(sqlx.NewSqlConnFromSession(session))
}
