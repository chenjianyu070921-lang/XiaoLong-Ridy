package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ DriverWithdrawModel = (*customDriverWithdrawModel)(nil)

type (
	// DriverWithdrawModel is an interface to be customized, add more methods here,
	// and implement the added methods in customDriverWithdrawModel.
	DriverWithdrawModel interface {
		driverWithdrawModel
		withSession(session sqlx.Session) DriverWithdrawModel
	}

	customDriverWithdrawModel struct {
		*defaultDriverWithdrawModel
	}
)

// NewDriverWithdrawModel returns a model for the database table.
func NewDriverWithdrawModel(conn sqlx.SqlConn) DriverWithdrawModel {
	return &customDriverWithdrawModel{
		defaultDriverWithdrawModel: newDriverWithdrawModel(conn),
	}
}

func (m *customDriverWithdrawModel) withSession(session sqlx.Session) DriverWithdrawModel {
	return NewDriverWithdrawModel(sqlx.NewSqlConnFromSession(session))
}
