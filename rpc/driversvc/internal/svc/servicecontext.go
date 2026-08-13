package svc

import (
	cfg "XiaoLong-Ridy/common/config"
	"XiaoLong-Ridy/common/datasource"
	"XiaoLong-Ridy/rpc/driversvc/internal/config"
	"time"

	"gorm.io/gorm"
)

// ServiceContext 保存服务的全局依赖，供各 Logic 使用。
type ServiceContext struct {
	Config config.Config // Config：服务配置
	DB     *gorm.DB      // DB：MySQL 连接实例（GORM 客户端）
}

// NewServiceContext 构建 ServiceContext，初始化 MySQL 连接。
func NewServiceContext(c config.Config) *ServiceContext {
	// 创建 mysql 连接，复用 common/datasource 的 GORM 客户端
	client, err := datasource.NewMysqlClient(cfg.MysqlConf{
		Dsn:         c.Mysql.DSN,        // Dsn：来自配置的数据库连接串
		MaxOpenConn: 200,                // MaxOpenConn：最大打开连接数
		MaxIdleConn: 30,                 // MaxIdleConn：最大空闲连接数
		MaxLifeTime: int(time.Hour),     // MaxLifeTime：连接最大存活时间（纳秒）
	})
	if err != nil {
		panic(err)
	}
	return &ServiceContext{
		Config: c,
		DB:     client,
	}
}
