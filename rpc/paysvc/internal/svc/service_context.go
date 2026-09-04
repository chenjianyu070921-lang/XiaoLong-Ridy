package svc

import (
	"errors"
	"fmt"

	"XiaoLong-Ridy/common/alipay"
	cfg "XiaoLong-Ridy/common/config"
	"XiaoLong-Ridy/common/datasource"
	"XiaoLong-Ridy/common/mq"
	"XiaoLong-Ridy/rpc/paysvc/internal/channel"
	"XiaoLong-Ridy/rpc/paysvc/internal/config"
	"XiaoLong-Ridy/rpc/paysvc/internal/orderclient"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"gorm.io/gorm"
)

// ServiceContext 持有 paysvc 运行期所有依赖。
//   - alipayChannel / Verifier 强制要求密钥齐全（M5-3）。生产环境绝不允许 MockVerifier 兜底，
//     否则攻击者只需伪造回调即可绕过签名验证直接走到业务逻辑。
//   - Producer：Kafka 不可用时降级为 NoopProducer，让服务仍能启动；事件投递失败仅记日志（M5-11 由 paysvc.go 退出时统一 Close）。
//   - OrderClient：懒连接，启动期不要求 ordersvc 在线，运行期第一次 RPC 调用时再发现。
//   - DB：MySQL 不可达时启动失败（M5-7）。MySQL 是强依赖，没有它服务不应该启动。
type ServiceContext struct {
	Config        config.Config
	DB            *gorm.DB
	Producer      mq.Producer             // Kafka 生产者
	OrderClient   orderclient.OrderClient // 订单服务客户端
	Verifier      channel.SignVerifier    // 回调验签器
	alipayChannel *channel.AlipayChannel  // 支付宝真实渠道
}

// GetChannel 按渠道名返回支付渠道实现。
//
// M5-12：
//   - 生产模式：启动期已保证支付宝渠道配置了真实密钥（缺则 paysvc 启动失败），
//     此处直接返回真实 AlipayChannel，绝不做 mock 兜底。
//   - dev/test 模式：允许 alipayChannel 为 nil（本地联调缺沙箱密钥），降级 MockChannel。
//   - 余额渠道按真实降级到本地账户扣减（本期不动，仍返回 MockChannel 占位）。
func (s *ServiceContext) GetChannel(name string) channel.PayChannel {
	if name == channel.Alipay && s.alipayChannel != nil {
		return s.alipayChannel
	}
	return channel.NewMockChannel(name)
}

// NewServiceContext 创建服务依赖并完成启动期校验。
// 不再用 panic；启动失败返回 error，由 main 决定退出码与日志格式（M5-7）。
func NewServiceContext(c config.Config) (*ServiceContext, error) {
	// 1. MySQL 强依赖：连不上就启动失败，而不是带着坏掉的 DB 跑后续逻辑。
	db, err := datasource.NewMysqlClient(cfg.MysqlConf{
		Dsn:         c.Mysql.DSN,
		MaxOpenConn: 200,
		MaxIdleConn: 30,
		MaxLifeTime: defaultMaxLifeSeconds, // 单位：秒，datasource 内部转 time.Duration。
	})
	if err != nil {
		return nil, fmt.Errorf("init mysql: %w", err)
	}

	// 2. 验签器强校验（M5-3）：生产模式密钥不全立即启动失败，绝不允许 MockVerifier 在生产路径出现。
	//    仅显式配置 dev/test 模式时，允许 keys 缺失降级 MockVerifier，供本地联调模拟回调。
	verifier, err := newVerifier(c.Alipay, c.Mode)
	if err != nil {
		return nil, fmt.Errorf("init alipay verifier: %w", err)
	}

	// 3. 支付宝真实渠道（M5-12 硬校验）：生产模式缺密钥直接启动失败，不降级 Mock。
	//    仅显式配置 dev/test 模式时，允许密钥缺失降级（GetChannel 走 MockChannel），供本地联调。
	alipayCh, err := newAlipayChannel(c.Alipay, c.Mode)
	if err != nil {
		return nil, fmt.Errorf("init alipay channel: %w", err)
	}

	// 4. Kafka 生产者：失败降级 NoopProducer，让服务仍能启动；事件投递失败仅记日志。
	producer := newProducer(c.Kafka)

	// 5. 订单服务客户端：NonBlock=true 启动期不拨号（M5-7 优雅启动）。
	//    go-zero 默认（Block）会在 NewClient 时同步 dial 并等待连接，ordersvc 未启动就会 fatal。
	//    业务当前未在启动路径调用 ordersvc，因此允许"连接建立推迟到第一次 RPC 调用"。
	orderClient := orderclient.NewRpcOrderClient(zrpc.MustNewClient(zrpc.RpcClientConf{
		Target:   c.Ordersvc.Target,
		NonBlock: true,
	}))

	return &ServiceContext{
		Config:        c,
		DB:            db,
		Producer:      producer,
		OrderClient:   orderClient,
		Verifier:      verifier,
		alipayChannel: alipayCh,
	}, nil
}

// Close 优雅释放资源（M5-11）。
//  - DB: 关闭 sqlDB 释放连接池；
//  - Producer: Kafka 客户端关闭（NoopProducer.Close 是 noop）。
// 调用方 defer 在 s.Stop 之前执行。
func (s *ServiceContext) Close() error {
	var errs []error
	if s.DB != nil {
		if sqlDB, err := s.DB.DB(); err == nil {
			if err := sqlDB.Close(); err != nil {
				errs = append(errs, fmt.Errorf("close mysql: %w", err))
			}
		} else {
			errs = append(errs, fmt.Errorf("obtain sql.DB: %w", err))
		}
	}
	if s.Producer != nil {
		if err := s.Producer.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close producer: %w", err))
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}

// newAlipayChannel 创建支付宝真实渠道。
//
// M5-12：
//   - 生产模式（mode 非 dev/test）：缺核心三件套（appId/privateKey/alipayPublicKey）必须返回 error，
//     paysvc 启动失败，强制运维配置。绝不允许静默降级为 MockChannel：
//     生产环境如果 yaml 没填密钥，paysvc "假装上线" 但调 mock 渠道生成假 transaction_id，
//     会导致真实资金流断链。
//   - dev/test 模式：允许密钥缺失，返回 (nil, nil)，GetChannel 走 MockChannel 兜底（仅本地联调）。
//
// 密钥读取优先环境变量（ALIPAY_APP_ID/ALIPAY_PRIVATE_KEY/ALIPAY_PUBLIC_KEY）。
func newAlipayChannel(a alipay.Config, mode string) (*channel.AlipayChannel, error) {
	if !a.HasRealKeys() {
		if mode == service.TestMode || mode == service.DevMode {
			logx.Info("alipay keys missing, channel will fallback to mock (dev/test mode only)")
			return nil, nil
		}
		return nil, errors.New("alipay keys missing: appId/privateKey/alipayPublicKey must all be set in paysvc.yaml or via env ALIPAY_APP_ID/ALIPAY_PRIVATE_KEY/ALIPAY_PUBLIC_KEY; mock fallback is not allowed (M5-12)")
	}
	resolved := a.WithDefaults()
	ch, err := channel.NewAlipayChannel(resolved)
	if err != nil {
		return nil, fmt.Errorf("init alipay channel: %w", err)
	}
	logx.Info("alipay real channel enabled")
	return ch, nil
}

// newProducer 创建 Kafka 生产者，失败时降级为 NoopProducer。
func newProducer(k config.KafkaConf) mq.Producer {
	if len(k.Brokers) == 0 {
		logx.Info("kafka brokers empty, use NoopProducer")
		return &mq.NoopProducer{}
	}
	p, err := mq.NewKafkaProducer(k.Brokers)
	if err != nil {
		logx.Errorf("init kafka producer failed: %v, use NoopProducer", err)
		return &mq.NoopProducer{}
	}
	return p
}

// newVerifier 创建支付宝回调验签器。
//
// M5-3：生产模式强制配置，密钥不全立即返回 error，绝不允许 MockVerifier 在生产路径出现。
//   - 仅当显式配置 Mode=dev/test 且密钥缺失时，降级 MockVerifier（仅本地联调，模拟回调不验签）；
//   - 密钥读取优先环境变量（M5-8）。
//
// 注意：真实支付宝回调用支付宝私钥签名、商户用支付宝公钥验签，本地联调无法伪造合法签名，
// 因此联调环境（dev/test）降级 MockVerifier 是唯一可行的回调模拟方式。
func newVerifier(a alipay.Config, mode string) (channel.SignVerifier, error) {
	if !a.HasRealKeys() {
		if mode == service.TestMode || mode == service.DevMode {
			logx.Info("alipay keys missing, use MockVerifier (dev/test mode only)")
			return &channel.MockVerifier{}, nil
		}
		return nil, errors.New("alipay keys missing (appId/privateKey/alipayPublicKey must all be set); " +
			"set ALIPAY_APP_ID/ALIPAY_PRIVATE_KEY/ALIPAY_PUBLIC_KEY env, or fill paysvc.yaml.alipay")
	}
	resolved := a.WithDefaults()
	v, err := channel.NewAlipayVerifier(resolved.AppId, resolved.PrivateKey, resolved.AlipayPublicKey, resolved.Sandbox)
	if err != nil {
		return nil, fmt.Errorf("init alipay verifier: %w", err)
	}
	return v, nil
}

// defaultMaxLifeSeconds 默认的 mysql 连接最大生命周期（秒）：1 小时。
const defaultMaxLifeSeconds = 3600
