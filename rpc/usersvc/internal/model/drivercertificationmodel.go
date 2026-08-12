package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ DriverCertificationModel = (*customDriverCertificationModel)(nil)

type (
	// DriverCertificationModel is an interface to be customized, add more methods here,
	// and implement the added methods in customDriverCertificationModel.
	DriverCertificationModel interface {
		driverCertificationModel
		withSession(session sqlx.Session) DriverCertificationModel
	}

	customDriverCertificationModel struct {
		*defaultDriverCertificationModel
	}
)

// NewDriverCertificationModel returns a model for the database table.
func NewDriverCertificationModel(conn sqlx.SqlConn) DriverCertificationModel {
	return &customDriverCertificationModel{
		defaultDriverCertificationModel: newDriverCertificationModel(conn),
	}
}

func (m *customDriverCertificationModel) withSession(session sqlx.Session) DriverCertificationModel {
	return NewDriverCertificationModel(sqlx.NewSqlConnFromSession(session))
}
