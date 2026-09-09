package constants

const (
	OrderTypeRealtime    = 1
	OrderTypeReservation = 2
)

const (
	OrderStatusWaitAccept = 1
	OrderStatusAccepted   = 2
	OrderStatusOnTrip     = 3
	OrderStatusWaitPay    = 4
	OrderStatusCompleted  = 5
	OrderStatusCancelled  = 6
	OrderStatusRefunded   = 7
)

const (
	DispatchTypeAuto = 1
)

const (
	DispatchStatusPending   = 1
	DispatchStatusAccepted  = 2
	DispatchStatusRejected  = 3
	DispatchStatusTimeout   = 4
	DispatchStatusCancelled = 5
)

const (
	OperatorUser               = "user"
	OperatorDriver             = "driver"
	OperatorSystem             = "system"
	OperatorAdmin              = "admin"
	RedisDriverPos             = "driver:pos:%d"
	RedisOrderInfo             = "order:info:%d"
	RedisSmsCode               = "sms:code:%s"
	RedisOrderLock             = "r:lock:order:%d"
	RedisDriverGeo             = "driver:geo:%s"
	RedisDriverOnline          = "driver:online"
	RedisDriverBusy            = "driver:busy"
	RedisDriverAvailable       = "driver:available:%d"
	RedisDriverPush            = "driver:push:%d"
	RedisDriverPrefRealtime    = "driver:pref:realtime"
	RedisDriverPrefReservation = "driver:pref:reservation"
	// RedisChatOnline 聊天相关 Redis key：跨网关 Pub/Sub 转发核心。
	RedisChatOnline      = "chat:online:user:%d"  // 用户在线 / 所在 WS 实例
	RedisChatRoomMembers = "chat:room:%d:members" // 聊天室成员
	RedisChatPub         = "chat:pub:%d"          // 实时消息 Pub/Sub 频道（roomId=orderId）
)

const (
	TopicLocation           = "location-report"
	TopicOrder              = "order-event"
	TopicOrderCreated       = "order.created"
	TopicOrderStatusChanged = "order.status.changed"
	TopicOrderCancelled     = "order.canceled"
	TopicDispatchNew        = "dispatch.new"
	TopicDispatchResult     = "dispatch.result"
	TopicOrderPaid          = "order.paid"
	TopicOrderRefunded      = "order.refunded" // 閫€娆炬垚鍔?)
	// TopicAdminDomain 承载管理后台领域可靠事件，消费者按事件体中的 event_type 分发具体业务动作。
	TopicAdminDomain = "admin.domain"
	// TopicChatMessage 聊天消息落库后的异步事件，供重试 / 死信（chat.message.dlq）消费。
	TopicChatMessage = "chat.message"
)

const (
	OrderEventStream     = "order:event:stream"
	DriverLocationStream = "driver:location:stream"
)

const (
	DispatchRetryQueueKey   = "dispatch:retry:orders"
	MaxDispatchRetryAttempt = 3
	RefundRetryQueueKey     = "refund:retry:events"
	MaxRefundRetryAttempt   = 5
	// PaymentRetryQueueKey 行程结束后创建支付单失败的重试队列。
	// P0-3 修复：FinishTrip 先把订单状态改成 WaitPay 再调 createPayment，后者失败时无补偿会导致订单永久卡 WaitPay。
	PaymentRetryQueueKey   = "payment:retry:orders"
	MaxPaymentRetryAttempt = 5
)
