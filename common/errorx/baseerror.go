package errorx

// 通用错误码
const (
	OK     = 0
	FAIL   = 500
	PARAM  = 400
	NOTLOG = 401 // 未登录
	NOAUTH = 403 // 无权限
)

// 业务错误码（1000 起）
const (
	DriverNotExist   = 1001 // 司机不存在
	NoNearbyDriver   = 1002 // 附近没有司机
	OrderTimeout     = 1003 // 订单已超时
	TokenExpired     = 1004 // 登录已过期
	TokenInvalid     = 1005 // 凭证无效
	RedisUnavailable = 1006 // 缓存服务不可用
	DbError          = 1007 // 数据库异常
	KafkaSendFailed  = 1008 // 消息发送失败
)

// 错误消息
var msgs = map[int]string{
	OK:               "成功",
	FAIL:             "失败",
	PARAM:            "参数不对",
	NOTLOG:           "未登录",
	NOAUTH:           "无权限",
	DriverNotExist:   "司机不存在",
	NoNearbyDriver:   "附近没有司机",
	OrderTimeout:     "订单已超时",
	TokenExpired:     "登录已过期",
	TokenInvalid:     "凭证无效",
	RedisUnavailable: "缓存服务不可用",
	DbError:          "数据库异常",
	KafkaSendFailed:  "消息发送失败",
}

func Msg(code int) string {
	if m, ok := msgs[code]; ok {
		return m
	}
	return "未知错误"
}

// NewErr 创建一个错误
func NewErr(code int) error {
	return &Err{Code: code, Msg: Msg(code)}
}

// NewErrMsg 创建带自定义消息的错误
func NewErrMsg(code int, msg string) error {
	return &Err{Code: code, Msg: msg}
}

type Err struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func (e *Err) Error() string {
	return e.Msg
}
