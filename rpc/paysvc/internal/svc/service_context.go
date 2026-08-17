package svc

import (
	"XiaoLong-Ridy/common/alipay"
	cfg "XiaoLong-Ridy/common/config"
	"XiaoLong-Ridy/common/datasource"
	"XiaoLong-Ridy/common/mq"
	"XiaoLong-Ridy/rpc/paysvc/internal/channel"
	"XiaoLong-Ridy/rpc/paysvc/internal/config"
	"XiaoLong-Ridy/rpc/paysvc/internal/orderclient"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config      config.Config
	DB          *gorm.DB
	Producer    mq.Producer             // Kafka 生产者
	OrderClient orderclient.OrderClient // 订单服务客户端
	Verifier    channel.SignVerifier    // 回调验签器
}

func NewServiceContext(c config.Config) *ServiceContext {
	// mysql 连接
	db, err := datasource.NewMysqlClient(cfg.MysqlConf{
		Dsn:         c.Mysql.DSN,
		MaxOpenConn: 200,
		MaxIdleConn: 30,
		MaxLifeTime: int(time.Hour),
	})
	if err != nil {
		panic(err)
	}

	svcCtx := &ServiceContext{
		Config: c,
		DB:     db,
	}

	// Kafka 生产者（容错：未启动时降级为 NoopProducer）
	svcCtx.Producer = newProducer(c.Kafka)

	// 订单服务客户端（直连，懒连接）
	svcCtx.OrderClient = orderclient.NewRpcOrderClient(zrpc.MustNewClient(zrpc.RpcClientConf{
		Target: c.Ordersvc.Target,
	}))

	// 回调验签器（容错：密钥为空时降级为 MockVerifier）
	svcCtx.Verifier = newVerifier(c.Alipay)

	return svcCtx
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

// newVerifier 创建回调验签器，密钥缺失时降级为 MockVerifier。
func newVerifier(a alipay.Config) channel.SignVerifier {
	if a.AppId == "" || a.PrivateKey == "" || a.AlipayPublicKey == "" {
		logx.Info("alipay keys empty, use MockVerifier")
		return &channel.MockVerifier{}
	}
	v, err := channel.NewAlipayVerifier(a.AppId, a.PrivateKey, a.AlipayPublicKey, a.Sandbox)
	if err != nil {
		logx.Errorf("init alipay verifier failed: %v, use MockVerifier", err)
		return &channel.MockVerifier{}
	}
	return v
}
