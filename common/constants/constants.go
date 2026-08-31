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
)
