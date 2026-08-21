package svc

import (
	"fmt"
	"time"

	commoncfg "XiaoLong-Ridy/common/config"
	"XiaoLong-Ridy/common/datasource"
	"XiaoLong-Ridy/rpc/pricesvc/internal/config"
	"gorm.io/gorm"
)

// ServiceContext 持有 pricesvc 运行期依赖。
//   - DB 是强依赖：MySQL 连不上直接启动失败（M5-7），不要带着坏掉的 DB 跑后续逻辑。
//   - 价格是只读高敏感数据，连错库会让用户付错钱。
type ServiceContext struct {
	Config config.Config
	DB     *gorm.DB
}

// NewServiceContext 创建 ServiceContext。
// 与 paysvc 对齐：MySQL 失败立即返回 error 而不是 panic（M5-7），由 main 负责退出码与日志。
func NewServiceContext(c config.Config) (*ServiceContext, error) {
	client, err := datasource.NewMysqlClient(commoncfg.MysqlConf{
		Dsn:         c.Mysql.DSN,
		MaxOpenConn: 200,
		MaxIdleConn: 30,
		MaxLifeTime: int(time.Hour / time.Second), // 1 小时，单位秒
	})
	if err != nil {
		return nil, fmt.Errorf("init mysql: %w", err)
	}

	return &ServiceContext{
		Config: c,
		DB:     client,
	}, nil
}

// Close 优雅释放数据库连接池（M5-11）。
// 与 paysvc 的 Close 对齐，由 pricesvc.go 在 s.Stop 之前 defer 调用。
func (s *ServiceContext) Close() error {
	if s.DB == nil {
		return nil
	}
	sqlDB, err := s.DB.DB()
	if err != nil {
		return fmt.Errorf("obtain sql.DB: %w", err)
	}
	return sqlDB.Close()
}
