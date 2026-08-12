package errorx

// 错误码（简单记）
const (
	OK     = 0
	FAIL   = 500
	PARAM  = 400
	NOTLOG = 401 // 未登录
)

// 错误消息
var msgs = map[int]string{
	OK:     "成功",
	FAIL:   "失败",
	PARAM:  "参数不对",
	NOTLOG: "未登录",
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

type Err struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func (e *Err) Error() string {
	return e.Msg
}
