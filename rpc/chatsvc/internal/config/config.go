package config

import (
	cfg "XiaoLong-Ridy/common/config"

	"github.com/zeromicro/go-zero/zrpc"
)

// Config 是 chatsvc 的运行配置。
// 统一方案 V1.0：chatsvc 仅依赖 MySQL（会话/消息持久化）与 ordersvc（订单归属与状态）。
type Config struct {
	zrpc.RpcServerConf

	// Mysql 司乘聊天库，与业务库同实例不同表（im_conversation / im_message）。
	Mysql cfg.MysqlConf `yaml:"mysql" json:"mysql"`
	// OrderRPCAddr 订单服务 gRPC 地址，用于读取订单归属与状态。
	OrderRPCAddr string `yaml:"orderRpcAddr" json:"orderRpcAddr"`
	// Redis 实时推送通道（driver:push:%d）的发布端；留空则不使用实时推送（前端轮询兜底）。
	// 必须与 api/driver 使用同一实例，否则司机 WS 收不到推送。
	RedisHost string `yaml:"redisHost" json:"redisHost"`
	RedisPass string `yaml:"redisPass" json:"redisPass"`
}
