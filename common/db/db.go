package db

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

// 本地默认配置（改这里就行）
var (
	// MySQL 连接信息
	MySQLAddr     = "127.0.0.1:3306"
	MySQLUser     = "root"
	MySQLPass     = "root"
	MySQLDBName   = "xiaolongridy"

	// Redis 连接信息
	RedisAddr = "127.0.0.1:6379"
	RedisPass = ""
)

// MySQL 连接
var mysqlDB *sql.DB

// GetMySQL 获取 MySQL 连接（第一次调用时自动连接）
func GetMySQL() *sql.DB {
	if mysqlDB == nil {
		dsn := MySQLUser + ":" + MySQLPass + "@tcp(" + MySQLAddr + ")/" + MySQLDBName + "?charset=utf8mb4&parseTime=true"
		var err error
		mysqlDB, err = sql.Open("mysql", dsn)
		if err != nil {
			log.Fatal("MySQL 连接失败:", err)
		}
		// 测试连接
		err = mysqlDB.Ping()
		if err != nil {
			log.Fatal("MySQL Ping 失败:", err)
		}
		log.Println("MySQL 连接成功!", MySQLDBName)
	}
	return mysqlDB
}

// Redis 连接
var redisDB *redis.Redis

// GetRedis 获取 Redis 连接（第一次调用时自动连接）
func GetRedis() *redis.Redis {
	if redisDB == nil {
		var err error
		redisDB, err = redis.NewRedis(redis.RedisConf{
			Host: RedisAddr,
			Pass: RedisPass,
			Type: redis.NodeType,
		})
		if err != nil {
			log.Fatal("Redis 连接失败:", err)
		}
		log.Println("Redis 连接成功!")
	}
	return redisDB
}
