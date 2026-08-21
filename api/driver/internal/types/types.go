// Package types 定义司机端 API 的请求与响应数据结构。
package types

// ============ 通用响应 ============

// Response 是接口文档约定的统一响应格式（与 passenger 端保持一致）。
type Response struct {
	Code      int    `json:"code"`      // 业务码，0 表示成功
	Message   string `json:"message"`   // 提示信息
	Data      any    `json:"data"`      // 业务数据，可为任意结构
	Timestamp int64  `json:"timestamp"` // 响应生成时的秒级时间戳
	TraceID   string `json:"traceId"`   // 请求链路追踪 ID
}

// ============ 司机（增删改查四个核心接口） ============

// CreateDriverRequest 创建司机请求。
type CreateDriverRequest struct {
	Phone           string `json:"phone"`           // 手机号（登录账号）
	PasswordHash    string `json:"passwordHash"`    // 密码哈希
	RealName        string `json:"realName"`        // 真实姓名
	IdCardNo        string `json:"idCardNo"`        // 身份证号
	DriverLicenseNo string `json:"driverLicenseNo"` // 驾驶证号
	AvatarURL       string `json:"avatarUrl"`       // 头像地址
}

// CreateDriverResponse 创建司机响应。
type CreateDriverResponse struct {
	ID        int64  `json:"id"`        // 新建司机 ID
	Status    string `json:"status"`    // 初始状态（待审核）
	CreatedAt int64  `json:"createdAt"` // 创建时间（Unix 秒）
}

// UpdateDriverRequest 更新司机请求，除 id 外字段均为可选（指针表示可缺省）。
type UpdateDriverRequest struct {
	ID              int64   `json:"id"`              // 待更新司机 ID
	Phone           *string `json:"phone,omitempty"` // 可选手机号
	PasswordHash    *string `json:"passwordHash,omitempty"`    // 可选密码哈希
	RealName        *string `json:"realName,omitempty"`        // 可选真实姓名
	IdCardNo        *string `json:"idCardNo,omitempty"`        // 可选身份证号
	DriverLicenseNo *string `json:"driverLicenseNo,omitempty"` // 可选驾驶证号
	AvatarURL       *string `json:"avatarUrl,omitempty"`       // 可选头像地址
	Status          *string `json:"status,omitempty"`          // 可选账号状态
}

// UpdateDriverResponse 更新司机响应。
type UpdateDriverResponse struct {
	ID        int64  `json:"id"`        // 司机 ID
	Status    string `json:"status"`    // 更新后的状态
	UpdatedAt int64  `json:"updatedAt"` // 更新时间（Unix 秒）
}

// DriverDetail 司机完整信息。
type DriverDetail struct {
	ID              int64  `json:"id"`              // 司机 ID
	Phone           string `json:"phone"`           // 手机号
	RealName        string `json:"realName"`        // 真实姓名
	IdCardNo        string `json:"idCardNo"`        // 身份证号
	DriverLicenseNo string `json:"driverLicenseNo"` // 驾驶证号
	AvatarURL       string `json:"avatarUrl"`       // 头像地址
	Status          string `json:"status"`          // 账号状态
	OnlineStatus    int    `json:"onlineStatus"`    // 在线状态：0离线 1在线
	CreatedAt       int64  `json:"createdAt"`       // 创建时间（Unix 秒）
	UpdatedAt       int64  `json:"updatedAt"`       // 更新时间（Unix 秒）
}

// GetDriverResponse 查询司机详情响应。
type GetDriverResponse struct {
	Driver DriverDetail `json:"driver"` // 司机完整信息
}

// DeleteResponse 删除操作通用响应。
type DeleteResponse struct {
	ID      int64 `json:"id"`      // 资源 ID
	Success bool  `json:"success"` // 是否删除成功
}

// ============ 司机登录认证 ============

// SendSMSCodeRequest 发送登录短信验证码请求。
type SendSMSCodeRequest struct {
	Phone string `json:"phone"` // 手机号
}

// SendSMSCodeResponse 发送登录短信验证码响应。
type SendSMSCodeResponse struct {
	Success  bool `json:"success"`  // 是否发送成功
	ExpireIn int  `json:"expireIn"` // 验证码有效期（秒）
}

// LoginByPasswordRequest 手机号 + 密码登录请求。
type LoginByPasswordRequest struct {
	Phone    string `json:"phone"`    // 手机号
	Password string `json:"password"` // 明文密码
}

// LoginBySMSRequest 手机号 + 验证码登录请求。
type LoginBySMSRequest struct {
	Phone string `json:"phone"` // 手机号
	Code  string `json:"code"`  // 短信验证码
}

// LoginResponse 登录响应，返回 JWT 与司机基础信息。
type LoginResponse struct {
	Token   string      `json:"token"`   // JWT 登录凭证
	ExpireIn int64      `json:"expireIn"` // 有效期（秒）
	Driver  DriverBrief `json:"driver"`  // 司机基础信息
}

// DriverBrief 司机登录后可暴露的简要信息（脱敏）。
type DriverBrief struct {
	ID     int64  `json:"id"`     // 司机 ID
	Phone  string `json:"phone"`  // 脱敏手机号
	Status string `json:"status"` // 账号状态
}

// ============ 司机上下线 ============

// SetOnlineResponse 司机上线响应。
type SetOnlineResponse struct {
	DriverID     int64 `json:"driverId"`     // 司机 ID
	OnlineStatus int   `json:"onlineStatus"` // 上线后状态：1在线
}

// SetOfflineResponse 司机下线响应。
type SetOfflineResponse struct {
	DriverID     int64 `json:"driverId"`     // 司机 ID
	OnlineStatus int   `json:"onlineStatus"` // 下线后状态：0离线
}

// ============ 司机接单 / 行程 ============

// AcceptOrderRequest 司机接单请求。
type AcceptOrderRequest struct {
	OrderID int64 `json:"orderId"` // 待接订单 ID
}

// AcceptOrderResponse 司机接单响应。
type AcceptOrderResponse struct {
	OrderID int64 `json:"orderId"` // 订单 ID
	Status  int32 `json:"status"`  // 接单后订单状态（见 ordersvc OrderStatus 枚举）
}

// StartTripRequest 司机开始行程请求。
type StartTripRequest struct {
	OrderID int64 `json:"orderId"` // 订单 ID
}

// StartTripResponse 司机开始行程响应。
type StartTripResponse struct {
	OrderID int64 `json:"orderId"` // 订单 ID
	Status  int32 `json:"status"`  // 行程开始后订单状态
}

// ConfirmArriveRequest 司机确认到达请求。
type ConfirmArriveRequest struct {
	OrderID int64 `json:"orderId"` // 订单 ID
}

// ConfirmArriveResponse 司机确认到达响应。
type ConfirmArriveResponse struct {
	OrderID int64 `json:"orderId"` // 订单 ID
	Status  int32 `json:"status"`  // 确认到达后订单状态
}

// FinishTripRequest 司机结束行程请求。
type FinishTripRequest struct {
	OrderID           int64 `json:"orderId"`           // 订单 ID
	ActualDistanceM   int64 `json:"actualDistanceM"`   // 实际里程（米）
	ActualDurationS   int64 `json:"actualDurationS"`   // 实际时长（秒）
	ActualPriceCents  int64 `json:"actualPriceCents"`  // 实际金额（分）
}

// FinishTripResponse 司机结束行程响应。
type FinishTripResponse struct {
	OrderID            int64 `json:"orderId"`            // 订单 ID
	Status             int32 `json:"status"`             // 行程结束后订单状态
	PayableAmountCents int64 `json:"payableAmountCents"` // 应付金额（分）
}
