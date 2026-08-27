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

type ServiceContext struct {
	Config                           config.Config
	DB                               *gorm.DB
	DriverRepository                 repository.DriverRepository
	DriverVehicleRepository          repository.DriverVehicleRepository
	CertificationRepository          repository.CertificationRepository
	DriverWithdrawRepository         repository.DriverWithdrawRepository
	DriverListenPreferenceRepository repository.DriverListenPreferenceRepository
	RedisClient                      *redis.Client
	OnlineStore                      *onlinestore.Store
	MinioClient                      *minio.Client
}

func NewServiceContext(c config.Config) *ServiceContext {
	db, err := gorm.Open(mysql.Open(c.Mysql.DSN), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     c.DriverRedis.Host,
		Password: c.DriverRedis.Password,
		DB:       c.DriverRedis.DB,
	})

	mc, err := minio.New(c.Minio.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(c.Minio.AccessKey, c.Minio.SecretKey, ""),
		Secure: c.Minio.UseSSL,
	})
	if err != nil {
		panic(err)
	}

	return &ServiceContext{
		Config:                           c,
		DB:                               db,
		DriverRepository:                 repository.NewGormDriverRepository(db),
		DriverVehicleRepository:          repository.NewGormVehicleRepository(db),
		CertificationRepository:          repository.NewGormCertificationRepository(db),
		DriverWithdrawRepository:         repository.NewGormDriverWithdrawRepository(db),
		DriverListenPreferenceRepository: repository.NewGormDriverListenPreferenceRepository(db),
		RedisClient:                      rdb,
		OnlineStore:                      onlinestore.NewStore(rdb, 0),
		MinioClient:                      mc,
	}
}
