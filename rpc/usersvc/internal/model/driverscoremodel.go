package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ DriverScoreModel = (*customDriverScoreModel)(nil)

type (
	// DriverScoreModel is an interface to be customized, add more methods here,
	// and implement the added methods in customDriverScoreModel.
	DriverScoreModel interface {
		driverScoreModel
		withSession(session sqlx.Session) DriverScoreModel
	}

	customDriverScoreModel struct {
		*defaultDriverScoreModel
	}
)

// NewDriverScoreModel returns a model for the database table.
func NewDriverScoreModel(conn sqlx.SqlConn) DriverScoreModel {
	return &customDriverScoreModel{
		defaultDriverScoreModel: newDriverScoreModel(conn),
	}
}

func (m *customDriverScoreModel) withSession(session sqlx.Session) DriverScoreModel {
	return NewDriverScoreModel(sqlx.NewSqlConnFromSession(session))
}
