package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	commonconfig "XiaoLong-Ridy/common/config"
	"XiaoLong-Ridy/common/datasource"
	commonRealName "XiaoLong-Ridy/common/realname"
	commonSMS "XiaoLong-Ridy/common/sms"
	"XiaoLong-Ridy/rpc/usersvc/internal/config"
	"XiaoLong-Ridy/rpc/usersvc/internal/logic"
	"XiaoLong-Ridy/rpc/usersvc/internal/repository"
	"XiaoLong-Ridy/rpc/usersvc/internal/server"
	"XiaoLong-Ridy/rpc/usersvc/internal/svc"
	userproto "XiaoLong-Ridy/rpc/usersvc/proto"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/usersvc.yaml", "the config file")

// main 是 usersvc 的 go-zero RPC 启动入口，负责加载配置、组装依赖并注册 gRPC 服务。
func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx := newServiceContext(c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		userproto.RegisterUserServer(grpcServer, server.NewUserServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}

// newServiceContext 创建 usersvc 运行时依赖；正式服务统一使用 MySQL 和 Redis 持久化。
func newServiceContext(c config.Config) *svc.ServiceContext {
	signingKey := firstNonEmpty(os.Getenv("USERSVC_TOKEN_SIGNING_KEY"), os.Getenv("JWT_SIGNING_KEY"), c.TokenAuth.SigningKey)
	if signingKey == "" {
		panic("usersvc token signing key is required")
	}

	mysqlConf := normalizeMysqlConf(c.Mysql)
	db, err := datasource.NewMysqlClient(commonconfig.MysqlConf{
		Dsn:         mysqlConf.DSN,
		MaxOpenConn: mysqlConf.MaxOpenConn,
		MaxIdleConn: mysqlConf.MaxIdleConn,
		MaxLifeTime: mysqlConf.MaxLifeTime,
	})
	if err != nil {
		panic(fmt.Errorf("连接 usersvc mysql 失败: %w", err))
	}

	redisConf := normalizeRedisConf(c.CacheRedis)
	redisClient := datasource.NewRedisClient(commonconfig.RedisConf{
		Host:     redisConf.Host,
		Pass:     redisConf.Pass,
		Db:       redisConf.DB,
		PoolSize: redisConf.PoolSize,
	})
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		panic(fmt.Errorf("连接 usersvc redis 失败: %w", err))
	}

	users := repository.NewGormUserRepository(db)
	addresses := repository.NewGormAddressRepository(db)
	coupons := repository.NewGormCouponRepository(db)
	smsMessageSender, err := newSMSMessageSender(c.SMS)
	if err != nil {
		panic(err)
	}
	smsService := logic.NewRedisSMSCodeService(redisClient, smsMessageSender, func(phone, code string) {
		if smsMessageSender != nil {
			log.Printf("短信验证码已提交腾讯云发送：phone=%s", phone)
			return
		}
		log.Printf("本地短信验证码：phone=%s code=%s", phone, code)
	})
	tokens := logic.NewRedisTokenManager(redisClient, signingKey)

	// 初始化腾讯云实名认证服务（配置为空时返回 nil，SubmitRealName 将跳过核验）
	realNameVer, err := newRealNameVerifier(c.TencentCloud)
	if err != nil {
		panic(fmt.Errorf("初始化腾讯云实名认证失败: %w", err))
	}

	return svc.NewServiceContext(c, users, addresses, coupons, smsService, smsService, tokens, realNameVer)
}

// newRealNameVerifier 根据配置创建实名认证实例；未配置时返回 nil 表示跳过核验。
func newRealNameVerifier(c commonRealName.TencentCloudConfig) (commonRealName.Verifier, error) {
	c = normalizeTencentCloudConf(c)
	if c.SecretID == "" || c.SecretKey == "" {
		log.Println("未配置腾讯云密钥，实名认证将跳过核验")
		return nil, nil
	}
	return commonRealName.NewTencentCloudRealNameVerifier(c), nil
}

// normalizeTencentCloudConf 从环境变量补齐腾讯云配置。
func normalizeTencentCloudConf(c commonRealName.TencentCloudConfig) commonRealName.TencentCloudConfig {
	c.SecretID = firstNonEmpty(os.Getenv("TENCENTCLOUD_SECRET_ID"), c.SecretID)
	c.SecretKey = firstNonEmpty(os.Getenv("TENCENTCLOUD_SECRET_KEY"), c.SecretKey)
	c.Region = firstNonEmpty(os.Getenv("TENCENTCLOUD_REGION"), c.Region)
	if c.Region == "" {
		c.Region = "ap-beijing"
	}
	return c
}

// newSMSMessageSender 根据配置创建真实短信发送器；腾讯云配置齐全时作为首选通道。
func newSMSMessageSender(c config.SMSConf) (commonSMS.Sender, error) {
	c = normalizeSMSConf(c)
	provider := strings.ToLower(strings.TrimSpace(c.Provider))
	if provider == "" && hasTencentSMSConfig(c) {
		provider = "tencent"
	}

	switch provider {
	case "", "local", "log":
		return nil, nil
	case "tencent":
		return commonSMS.NewTencentSender(commonSMS.TencentConfig{
			SecretID:    c.SecretID,
			SecretKey:   c.SecretKey,
			Region:      c.Region,
			SmsSdkAppID: c.SmsSdkAppID,
			SignName:    c.SignName,
			TemplateID:  c.TemplateID,
		})
	default:
		return nil, fmt.Errorf("unsupported usersvc sms provider: %s", c.Provider)
	}
}

// normalizeSMSConf 从环境变量补齐腾讯云短信配置，避免密钥硬编码在配置文件或代码里。
func normalizeSMSConf(c config.SMSConf) config.SMSConf {
	c.Provider = firstNonEmpty(os.Getenv("USERSVC_SMS_PROVIDER"), c.Provider)
	c.Region = firstNonEmpty(os.Getenv("TENCENTCLOUD_REGION"), c.Region)
	c.SecretID = firstNonEmpty(os.Getenv("TENCENTCLOUD_SECRET_ID"), c.SecretID)
	c.SecretKey = firstNonEmpty(os.Getenv("TENCENTCLOUD_SECRET_KEY"), c.SecretKey)
	c.SmsSdkAppID = firstNonEmpty(os.Getenv("TENCENTCLOUD_SMS_SDK_APP_ID"), c.SmsSdkAppID)
	c.SignName = firstNonEmpty(os.Getenv("TENCENTCLOUD_SMS_SIGN_NAME"), c.SignName)
	c.TemplateID = firstNonEmpty(os.Getenv("TENCENTCLOUD_SMS_TEMPLATE_ID"), c.TemplateID)
	return c
}

// hasTencentSMSConfig 判断是否已经提供任一腾讯云短信关键配置，用于自动选择腾讯云通道。
func hasTencentSMSConfig(c config.SMSConf) bool {
	return strings.TrimSpace(c.SecretID) != "" ||
		strings.TrimSpace(c.SecretKey) != "" ||
		strings.TrimSpace(c.SmsSdkAppID) != "" ||
		strings.TrimSpace(c.SignName) != "" ||
		strings.TrimSpace(c.TemplateID) != ""
}

// firstNonEmpty 返回第一个非空字符串，统一处理配置和环境变量优先级。
func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

// normalizeMysqlConf 校验并补齐 usersvc MySQL 连接池默认值。
func normalizeMysqlConf(c config.MysqlConf) config.MysqlConf {
	c.DSN = strings.TrimSpace(c.DSN)
	if c.DSN == "" {
		panic("usersvc mysql dsn is required")
	}
	if c.MaxOpenConn <= 0 {
		c.MaxOpenConn = 200
	}
	if c.MaxIdleConn <= 0 {
		c.MaxIdleConn = 30
	}
	if c.MaxLifeTime <= 0 {
		c.MaxLifeTime = 3600
	}
	return c
}

// normalizeRedisConf 校验并补齐 usersvc Redis 连接池默认值。
func normalizeRedisConf(c config.RedisConf) config.RedisConf {
	c.Host = strings.TrimSpace(c.Host)
	if c.Host == "" {
		panic("usersvc redis host is required")
	}
	if c.PoolSize <= 0 {
		c.PoolSize = 20
	}
	return c
}
