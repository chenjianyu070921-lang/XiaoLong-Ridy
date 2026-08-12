package datasource

import (
	"time"

	"XiaoLong-Ridy/common/config"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// NewMysqlClient 创建 gorm 客户端
func NewMysqlClient(c config.MysqlConf) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(c.Dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(c.MaxOpenConn)
	sqlDB.SetMaxIdleConns(c.MaxIdleConn)
	sqlDB.SetConnMaxLifetime(time.Duration(c.MaxLifeTime))

	return db, nil
}
