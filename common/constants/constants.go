package constants

import "fmt"

// 订单状态
const (
	OrderStatusWaitAccept = 1 // 待接单
	OrderStatusAccepted   = 2 // 已接单
	OrderStatusOnTrip     = 3 // 行程中
	OrderStatusWaitPay    = 4 // 待支付
	OrderStatusCompleted  = 5 // 已完成
	OrderStatusCancelled  = 6 // 已取消
)

// 订单操作方
const (
	OperatorUser   = "user"
	OperatorDriver = "driver"
	OperatorSystem = "system"
	OperatorAdmin  = "admin"
)

// Redis Key 模板（用的时候 fmt.Sprintf 填入 ID）
const (
	RedisDriverPos = "driver:pos:%d" // 司机位置
	RedisOrderInfo = "order:info:%d" // 订单信息
	RedisSmsCode   = "sms:code:%s"   // 验证码
)

// 司机位置相关 Redis Key（locationsvc / job / location-consumer 共用，避免各模块硬编码不一致）
const (
	DriverGeoKey      = "driver:geo"             // 司机实时位置 GEO 集合默认 key
	DriverOnlineKey   = "driver:online"          // 在线司机集合（SADD/SREM 维护）
	LocationStreamKey = "driver:location:stream" // 司机位置事件流（Redis Stream）
)

// DriverGeoKeyOf 返回按城市分桶的 GEO key。city 为空时回退到默认 DriverGeoKey，
// 保证多城市派单互不干扰，同时兼容旧调用。
func DriverGeoKeyOf(city string) string {
	if city == "" {
		return DriverGeoKey
	}
	return fmt.Sprintf("driver:geo:%s", city)
}

// Kafka 消息主题
const (
	TopicLocation = "location-report" // 司机位置上报
	TopicOrder    = "order-event"     // 订单事件
)

// 过期时间（秒）
const (
	RedisPosExpire   = 300 // 司机位置在 Redis 的存活时间，5 分钟
	OrderWaitTimeout = 300 // 订单等待接单超时时间，5 分钟
)

// JWT 相关
const (
	JwtSecret = "xiao-long-ridy-secret"
	JwtExpire = 7 * 24 * 3600 // 7 天，单位秒
	JwtIssuer = "xiao-long-ridy"
)
