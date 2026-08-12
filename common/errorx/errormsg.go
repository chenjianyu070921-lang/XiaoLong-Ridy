package errorx

// 统一错误码定义
const (
	OK           = 0
	ServerError  = 500
	ParamError   = 400
	Unauthorized = 401
	Forbidden    = 403
	NotFound     = 404

	// 业务错误码 (1001-1999)
	ErrUserNotFound       = 1001
	ErrUserExist          = 1002
	ErrPasswordWrong      = 1003
	ErrAuthFailed         = 1004
	ErrTokenExpired       = 1005
	ErrOrderNotFound      = 1010
	ErrOrderStatusInvalid = 1011
	ErrDriverNotAvailable = 1020
	ErrLocationFailed     = 1030
	ErrPushFailed         = 1040
	ErrSMSFailed          = 1041
)

var errorMsgMap = map[int]string{
	OK:                    "success",
	ServerError:           "服务器内部错误",
	ParamError:            "参数错误",
	Unauthorized:          "未授权",
	Forbidden:             "无权限",
	NotFound:              "资源不存在",
	ErrUserNotFound:       "用户不存在",
	ErrUserExist:          "用户已存在",
	ErrPasswordWrong:      "密码错误",
	ErrAuthFailed:         "认证失败",
	ErrTokenExpired:       "Token已过期",
	ErrOrderNotFound:      "订单不存在",
	ErrOrderStatusInvalid: "订单状态异常",
	ErrDriverNotAvailable: "司机不可用",
	ErrLocationFailed:     "位置服务失败",
	ErrPushFailed:         "推送失败",
	ErrSMSFailed:          "短信发送失败",
}

func GetMsg(code int) string {
	if msg, ok := errorMsgMap[code]; ok {
		return msg
	}
	return "未知错误"
}

func NewCodeMsgError(code int) error {
	return NewCodeError(code, GetMsg(code))
}
