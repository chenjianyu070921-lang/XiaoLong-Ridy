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
