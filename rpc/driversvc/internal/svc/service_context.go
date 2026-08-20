package svc

import (
	"XiaoLong-Ridy/rpc/driversvc/internal/config"
	"XiaoLong-Ridy/rpc/driversvc/internal/onlinestore"
	"XiaoLong-Ridy/rpc/driversvc/internal/repository"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// ServiceContext 持有 driversvc 运行所需的依赖：配置、数据库仓储、Redis 与 MinIO 客户端。
type ServiceContext struct {
	Config                  config.Config                          // 服务全局配置（监听地址、MySQL DSN、JWT 密钥等）
	DB                      *gorm.DB                               // 原始 GORM 连接：供资质审核等需要事务的直接操作使用
	DriverRepository        repository.DriverRepository            // 司机主表仓储
	DriverVehicleRepository repository.DriverVehicleRepository     // 司机车辆仓储
	CertificationRepository repository.CertificationRepository     // 司机资质仓储
	RedisClient             *redis.Client                          // Redis 客户端：维护在线状态与多端互踢
	OnlineStore             *onlinestore.Store                     // 在线状态存储：封装在线/心跳/互踢判定
	MinioClient             *minio.Client                          // MinIO 客户端：司机资质图片对象存储
}

// NewServiceContext 构造服务上下文：建立 MySQL / Redis / MinIO 连接，注入仓储与客户端实例。
func NewServiceContext(c config.Config) *ServiceContext {
	// 使用配置中的 DSN 建立 MySQL 连接（GORM + mysql 驱动）。
	db, err := gorm.Open(mysql.Open(c.Mysql.DSN), &gorm.Config{})
	if err != nil {
		// 连接失败直接 panic，避免带着 nil 仓储启动导致后续请求崩溃。
		panic(err)
	}

	// 建立 Redis 连接，用于司机在线状态维护与多端互踢判定。
	rdb := redis.NewClient(&redis.Options{
		Addr:     c.DriverRedis.Host,
		Password: c.DriverRedis.Password,
		DB:       c.DriverRedis.DB,
	})

	// 建立 MinIO 连接，用于司机资质图片上传与读取。
	mc, err := minio.New(c.Minio.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(c.Minio.AccessKey, c.Minio.SecretKey, ""),
		Secure: c.Minio.UseSSL,
	})
	if err != nil {
		panic(err)
	}

	return &ServiceContext{
		Config:                  c,
		DB:                      db,
		DriverRepository:        repository.NewGormDriverRepository(db),
		DriverVehicleRepository: repository.NewGormVehicleRepository(db),
		CertificationRepository: repository.NewGormCertificationRepository(db),
		RedisClient:             rdb,
		OnlineStore:             onlinestore.NewStore(rdb, 0),
		MinioClient:             mc,
	}
}
