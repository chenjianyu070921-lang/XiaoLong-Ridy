package svc

import (
	"usersvc/internal/config"
	"usersvc/internal/model"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config                config.Config
	DriverModel          model.DriverModel
	DriverVehicleModel   model.DriverVehicleModel
	DriverCertificationModel model.DriverCertificationModel
	DriverScoreModel     model.DriverScoreModel
	DriverWithdrawModel  model.DriverWithdrawModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.Mysql.DataSource)
	return &ServiceContext{
		Config:                c,
		DriverModel:          model.NewDriverModel(conn),
		DriverVehicleModel:   model.NewDriverVehicleModel(conn),
		DriverCertificationModel: model.NewDriverCertificationModel(conn),
		DriverScoreModel:     model.NewDriverScoreModel(conn),
		DriverWithdrawModel:  model.NewDriverWithdrawModel(conn),
	}
}
