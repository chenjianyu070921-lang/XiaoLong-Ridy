package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"strings"
	"syscall"
	"time"

	"XiaoLong-Ridy/api/driver/internal/handler"
	"XiaoLong-Ridy/api/driver/internal/middleware"
	"XiaoLong-Ridy/api/driver/internal/svc"
	commonconfig "XiaoLong-Ridy/common/config"
	qiniuutil "XiaoLong-Ridy/common/qiniu"

	"gopkg.in/yaml.v3"
)

const defaultHTTPAddress = ":8082"

// Driver-side backend services are expected to run locally (localhost) in dev;
// override via env (DRIVER_GRPC_ADDR etc.) or etc/driver.yaml to point at a remote instance.
const defaultDriverGRPCAddr = "127.0.0.1:50055"

const defaultOrderGRPCAddr = "127.0.0.1:50051"

const defaultPayGRPCAddr = "127.0.0.1:50054"

const defaultPriceGRPCAddr = "127.0.0.1:50053"

const defaultDispatchGRPCAddr = "127.0.0.1:50056"

const defaultLocationGRPCAddr = "127.0.0.1:50057"

const defaultRedisAddr = ""

type driverConfig struct {
	HTTPAddr           string                 `yaml:"httpAddr"`
	DriverGRPCAddr     string                 `yaml:"driverGrpcAddr"`
	OrderGRPCAddr      string                 `yaml:"orderGrpcAddr"`
	PayGRPCAddr        string                 `yaml:"payGrpcAddr"`
	PriceGRPCAddr      string                 `yaml:"priceGrpcAddr"`
	DispatchGRPCAddr   string                 `yaml:"dispatchGrpcAddr"`
	LocationGRPCAddr   string                 `yaml:"locationGrpcAddr"`
	CORSAllowedOrigins []string               `yaml:"corsAllowedOrigins"`
	RedisAddr          string                 `yaml:"redisAddr"`
	RedisPassword      string                 `yaml:"redisPassword"`
	Mysql              commonconfig.MysqlConf `yaml:"mysql"`
	InternalAuth       svc.InternalAuthConfig `yaml:"internalAuth"`
	QiniuAccessKey     string                 `yaml:"qiniuAccessKey"`
	QiniuSecretKey     string                 `yaml:"qiniuSecretKey"`
	QiniuBucket        string                 `yaml:"qiniuBucket"`
	QiniuDomain        string                 `yaml:"qiniuDomain"`
	QiniuUploadURL     string                 `yaml:"qiniuUploadURL"`
	// SigningKey 司机端 JWT 签名密钥；为空时回退到内置本地联调默认值。
	// 生产/联调跨服务校验必须与 rpc/driversvc/etc/driversvc.yaml 的 signingKey 保持一致，
	// 推荐通过环境变量 DRIVER_SIGNING_KEY 注入，避免明文入库。
	SigningKey string `yaml:"signingKey"`
}

func main() {
	configPath := flag.String("f", "etc/driver.yaml", "driver api config file")
	flag.Parse()

	cfg, err := loadDriverConfig(*configPath)
	if err != nil {
		panic(fmt.Errorf("load driver api config: %w", err))
	}
	address := envOr("DRIVER_HTTP_ADDR", cfg.HTTPAddr)
	driverGRPCAddr := envOr("DRIVER_GRPC_ADDR", cfg.DriverGRPCAddr)
	orderGRPCAddr := envOr("ORDER_GRPC_ADDR", cfg.OrderGRPCAddr)
	payGRPCAddr := envOr("PAY_GRPC_ADDR", cfg.PayGRPCAddr)
	priceGRPCAddr := envOr("PRICE_GRPC_ADDR", cfg.PriceGRPCAddr)
	dispatchGRPCAddr := envOr("DISPATCH_GRPC_ADDR", cfg.DispatchGRPCAddr)
	locationGRPCAddr := envOr("LOCATION_GRPC_ADDR", cfg.LocationGRPCAddr)
	redisAddr := envOr("DRIVER_REDIS_ADDR", cfg.RedisAddr)
	redisPassword := envOr("DRIVER_REDIS_PASSWORD", cfg.RedisPassword)
	if internalToken := envOr("DRIVER_INTERNAL_SERVICE_TOKEN", ""); internalToken != "" {
		cfg.InternalAuth.ServiceToken = internalToken
	}
	if mysqlDSN := envOr("DRIVER_MYSQL_DSN", ""); mysqlDSN != "" {
		cfg.Mysql.Dsn = mysqlDSN
	}

	svcCtx := svc.NewServiceContextWithStorage(
		driverGRPCAddr,
		orderGRPCAddr,
		dispatchGRPCAddr,
		locationGRPCAddr,
		commonconfig.RedisConf{Host: redisAddr, Pass: redisPassword},
		cfg.Mysql,
		payGRPCAddr,
		priceGRPCAddr,
	)
	svcCtx.InternalAuth = cfg.InternalAuth
	if qiniuClient, err := newDriverQiniuClient(cfg); err != nil {
		panic(fmt.Errorf("driver qiniu config: %w", err))
	} else {
		svcCtx.Qiniu = qiniuClient
	}
	// 司机端 JWT 签名密钥解析优先级：环境变量 DRIVER_SIGNING_KEY > 配置文件 signingKey > 内置本地联调默认值。
	// 三层兜底确保 dev 联调无需手动注入即可启动，密钥永远不会落为空（空密钥会在 ValidateSigningKey 触发 panic）。
	if key := strings.TrimSpace(cfg.SigningKey); key != "" {
		svcCtx.SigningKey = key
	}
	svcCtx.SigningKey = envOr("DRIVER_SIGNING_KEY", svcCtx.SigningKey)
	if err := svcCtx.ValidateSigningKey(); err != nil {
		panic(fmt.Errorf("driver api signing key check: %w", err))
	}
	if err := svcCtx.ValidateInternalAuth(); err != nil {
		panic(fmt.Errorf("driver api internal auth check: %w", err))
	}

	server := &http.Server{
		Addr:         address,
		Handler:      recoverMiddleware(withCORS(middleware.InternalServiceAuth(svcCtx)(handler.NewRouter(svcCtx)), driverCORSAllowedOrigins(cfg))),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	log.Printf("driver api started at http://%s  (driversvc gRPC: %s, ordersvc gRPC: %s, paysvc gRPC: %s, pricesvc gRPC: %s, dispatchsvc gRPC: %s, locationsvc gRPC: %s, redis: %s)", address, driverGRPCAddr, orderGRPCAddr, payGRPCAddr, priceGRPCAddr, dispatchGRPCAddr, locationGRPCAddr, redisAddr)
	stopCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		<-stopCtx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		panic(fmt.Errorf("start driver api: %w", err))
	}
	if svcCtx.RedisClient != nil {
		_ = svcCtx.RedisClient.Close()
	}
}

// recoverMiddleware 捕获 handler panic，返回 500 而非让连接异常断开。
func recoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("panic recovered: %v\n%s", rec, debug.Stack())
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`{"code":50000,"message":"internal server error","data":null,"timestamp":0,"traceId":""}`))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func newDriverQiniuClient(cfg driverConfig) (*qiniuutil.Client, error) {
	qiniuCfg := qiniuutil.Config{
		AccessKey: envOr("DRIVER_QINIU_ACCESS_KEY", envOr("PASSENGER_QINIU_ACCESS_KEY", cfg.QiniuAccessKey)),
		SecretKey: envOr("DRIVER_QINIU_SECRET_KEY", envOr("PASSENGER_QINIU_SECRET_KEY", cfg.QiniuSecretKey)),
		Bucket:    envOr("DRIVER_QINIU_BUCKET", envOr("PASSENGER_QINIU_BUCKET", cfg.QiniuBucket)),
		Domain:    envOr("DRIVER_QINIU_DOMAIN", envOr("PASSENGER_QINIU_DOMAIN", cfg.QiniuDomain)),
		UploadURL: envOr("DRIVER_QINIU_UPLOAD_URL", envOr("PASSENGER_QINIU_UPLOAD_URL", cfg.QiniuUploadURL)),
	}
	if qiniuCfg.AccessKey == "" && qiniuCfg.SecretKey == "" && qiniuCfg.Bucket == "" && qiniuCfg.Domain == "" {
		return nil, nil
	}
	return qiniuutil.NewClient(qiniuCfg)
}

func loadDriverConfig(path string) (driverConfig, error) {
	cfg := driverConfig{
		HTTPAddr:         defaultHTTPAddress,
		DriverGRPCAddr:   defaultDriverGRPCAddr,
		OrderGRPCAddr:    defaultOrderGRPCAddr,
		PayGRPCAddr:      defaultPayGRPCAddr,
		PriceGRPCAddr:    defaultPriceGRPCAddr,
		DispatchGRPCAddr: defaultDispatchGRPCAddr,
		LocationGRPCAddr: defaultLocationGRPCAddr,
		RedisAddr:        defaultRedisAddr,
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}
		return cfg, err
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}
	if cfg.HTTPAddr == "" {
		cfg.HTTPAddr = defaultHTTPAddress
	}
	if cfg.DriverGRPCAddr == "" {
		cfg.DriverGRPCAddr = defaultDriverGRPCAddr
	}
	if cfg.OrderGRPCAddr == "" {
		cfg.OrderGRPCAddr = defaultOrderGRPCAddr
	}
	if cfg.PayGRPCAddr == "" {
		cfg.PayGRPCAddr = defaultPayGRPCAddr
	}
	if cfg.PriceGRPCAddr == "" {
		cfg.PriceGRPCAddr = defaultPriceGRPCAddr
	}
	if cfg.DispatchGRPCAddr == "" {
		cfg.DispatchGRPCAddr = defaultDispatchGRPCAddr
	}
	if cfg.LocationGRPCAddr == "" {
		cfg.LocationGRPCAddr = defaultLocationGRPCAddr
	}
	return cfg, nil
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func driverCORSAllowedOrigins(cfg driverConfig) []string {
	if origins := splitCSVEnv("DRIVER_CORS_ALLOWED_ORIGINS"); len(origins) > 0 {
		return origins
	}
	return compactStrings(cfg.CORSAllowedOrigins)
}

func splitCSVEnv(key string) []string {
	return compactStrings(strings.Split(os.Getenv(key), ","))
}

func compactStrings(values []string) []string {
	out := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func withCORS(next http.Handler, allowedOrigins []string) http.Handler {
	allowed := make(map[string]struct{}, len(allowedOrigins))
	for _, origin := range compactStrings(allowedOrigins) {
		allowed[origin] = struct{}{}
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := strings.TrimSpace(r.Header.Get("Origin"))
		if origin == "" {
			next.ServeHTTP(w, r)
			return
		}
		if _, ok := allowed[origin]; !ok {
			next.ServeHTTP(w, r)
			return
		}

		header := w.Header()
		header.Set("Access-Control-Allow-Origin", origin)
		header.Add("Vary", "Origin")
		header.Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		header.Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Internal-Service-Token, X-Trace-Id")
		header.Set("Access-Control-Max-Age", "600")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}


