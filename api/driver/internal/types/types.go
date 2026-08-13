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

// ============ 司机 ============

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

// DriverSummary 司机列表摘要项。
type DriverSummary struct {
	ID              int64  `json:"id"`              // 司机 ID
	Phone           string `json:"phone"`           // 手机号
	RealName        string `json:"realName"`        // 真实姓名
	DriverLicenseNo string `json:"driverLicenseNo"` // 驾驶证号
	Status          string `json:"status"`          // 账号状态
	CreatedAt       int64  `json:"createdAt"`       // 创建时间（Unix 秒）
}

// ListDriversRequest 分页查询司机列表请求。
type ListDriversRequest struct {
	Status       string `json:"status,omitempty"`       // 按状态过滤（可选）
	PhoneKeyword string `json:"phoneKeyword,omitempty"` // 手机号模糊匹配（可选）
	Page         int32  `json:"page,omitempty"`         // 页码，默认 1
	PageSize     int32  `json:"pageSize,omitempty"`     // 每页条数，默认 20，上限 100
}

// ListDriversResponse 分页查询司机列表响应。
type ListDriversResponse struct {
	List     []DriverSummary `json:"list"`     // 司机摘要列表
	Total    int64           `json:"total"`    // 符合条件的总记录数
	Page     int32           `json:"page"`     // 当前页码
	PageSize int32           `json:"pageSize"` // 每页条数
}

// ============ 车辆 ============

// CreateVehicleRequest 创建车辆请求。
type CreateVehicleRequest struct {
	DriverID           int64  `json:"driverId"`           // 所属司机 ID
	PlateNo            string `json:"plateNo"`            // 车牌号
	Brand              string `json:"brand"`              // 品牌
	Model              string `json:"model"`              // 车型
	Color              string `json:"color"`              // 车身颜色
	VehicleType        int32  `json:"vehicleType"`        // 车辆类型：1特惠快车 2快车 3拼车
	RegistrationDate   string `json:"registrationDate"`   // 注册日期（YYYY-MM-DD）
	InsuranceNo        string `json:"insuranceNo"`        // 保险单号
	InsuranceExpireAt  string `json:"insuranceExpireAt"`  // 保险到期日（YYYY-MM-DD）
}

// CreateVehicleResponse 创建车辆响应。
type CreateVehicleResponse struct {
	ID     int64 `json:"id"`     // 新建车辆 ID
	Status int32 `json:"status"` // 初始状态（1待审核）
}

// UpdateVehicleRequest 更新车辆请求，除 id 外字段均为可选。
type UpdateVehicleRequest struct {
	ID                 int64   `json:"id"`                 // 待更新车辆 ID
	DriverID           *int64  `json:"driverId,omitempty"` // 可选所属司机 ID
	PlateNo            *string `json:"plateNo,omitempty"`  // 可选车牌号
	Brand              *string `json:"brand,omitempty"`    // 可选品牌
	Model              *string `json:"model,omitempty"`    // 可选车型
	Color              *string `json:"color,omitempty"`    // 可选车身颜色
	VehicleType        *int32  `json:"vehicleType,omitempty"`       // 可选车辆类型
	RegistrationDate   *string `json:"registrationDate,omitempty"`  // 可选注册日期
	InsuranceNo        *string `json:"insuranceNo,omitempty"`       // 可选保险单号
	InsuranceExpireAt  *string `json:"insuranceExpireAt,omitempty"` // 可选保险到期日
	Status             *int32  `json:"status,omitempty"`            // 可选状态
}

// UpdateVehicleResponse 更新车辆响应。
type UpdateVehicleResponse struct {
	ID        int64 `json:"id"`        // 车辆 ID
	Status    int32 `json:"status"`    // 更新后的状态
	UpdatedAt int64 `json:"updatedAt"` // 更新时间（Unix 秒）
}

// VehicleDetail 车辆完整信息。
type VehicleDetail struct {
	ID                int64  `json:"id"`                // 车辆 ID
	DriverID          int64  `json:"driverId"`          // 所属司机 ID
	PlateNo           string `json:"plateNo"`           // 车牌号
	Brand             string `json:"brand"`             // 品牌
	Model             string `json:"model"`             // 车型
	Color             string `json:"color"`             // 车身颜色
	VehicleType       int32  `json:"vehicleType"`       // 车辆类型
	RegistrationDate  string `json:"registrationDate"`  // 注册日期
	InsuranceNo       string `json:"insuranceNo"`       // 保险单号
	InsuranceExpireAt string `json:"insuranceExpireAt"` // 保险到期日
	Status            int32  `json:"status"`            // 状态
	CreatedAt         int64  `json:"createdAt"`         // 创建时间（Unix 秒）
	UpdatedAt         int64  `json:"updatedAt"`         // 更新时间（Unix 秒）
}

// GetVehicleResponse 查询车辆详情响应。
type GetVehicleResponse struct {
	Vehicle VehicleDetail `json:"vehicle"` // 车辆完整信息
}

// VehicleSummary 车辆列表摘要项。
type VehicleSummary struct {
	ID          int64  `json:"id"`          // 车辆 ID
	DriverID    int64  `json:"driverId"`    // 所属司机 ID
	PlateNo     string `json:"plateNo"`     // 车牌号
	Brand       string `json:"brand"`       // 品牌
	VehicleType int32  `json:"vehicleType"` // 车辆类型
	Status      int32  `json:"status"`      // 状态
}

// ListVehiclesRequest 分页查询车辆列表请求。
type ListVehiclesRequest struct {
	DriverID int64 `json:"driverId,omitempty"` // 按司机过滤（可选）
	Status   int32 `json:"status,omitempty"`   // 按状态过滤（可选）
	Page     int32 `json:"page,omitempty"`     // 页码，默认 1
	PageSize int32 `json:"pageSize,omitempty"` // 每页条数，默认 20，上限 100
}

// ListVehiclesResponse 分页查询车辆列表响应。
type ListVehiclesResponse struct {
	List     []VehicleSummary `json:"list"`     // 车辆摘要列表
	Total    int64            `json:"total"`    // 符合条件的总记录数
	Page     int32            `json:"page"`     // 当前页码
	PageSize int32            `json:"pageSize"` // 每页条数
}

// ============ 认证 ============

// CreateCertificationRequest 创建认证请求。
type CreateCertificationRequest struct {
	DriverID          int64  `json:"driverId"`          // 所属司机 ID
	VehicleID         int64  `json:"vehicleId"`         // 关联车辆 ID
	IdCardFrontURL    string `json:"idCardFrontUrl"`    // 身份证人像面
	IdCardBackURL     string `json:"idCardBackUrl"`     // 身份证国徽面
	DriverLicenseURL  string `json:"driverLicenseUrl"`  // 驾驶证照片
	VehicleLicenseURL string `json:"vehicleLicenseUrl"` // 行驶证照片
}

// CreateCertificationResponse 创建认证响应。
type CreateCertificationResponse struct {
	ID         int64 `json:"id"`         // 新建认证 ID
	AuditStatus int32 `json:"auditStatus"` // 初始审核状态（1待审核）
}

// UpdateCertificationRequest 更新认证请求，除 id 外字段均为可选。
type UpdateCertificationRequest struct {
	ID                int64   `json:"id"`                // 待更新认证 ID
	DriverID          *int64  `json:"driverId,omitempty"` // 可选所属司机 ID
	VehicleID         *int64  `json:"vehicleId,omitempty"` // 可选关联车辆 ID
	IdCardFrontURL    *string `json:"idCardFrontUrl,omitempty"`    // 可选身份证人像面
	IdCardBackURL     *string `json:"idCardBackUrl,omitempty"`     // 可选身份证国徽面
	DriverLicenseURL  *string `json:"driverLicenseUrl,omitempty"`  // 可选驾驶证照片
	VehicleLicenseURL *string `json:"vehicleLicenseUrl,omitempty"` // 可选行驶证照片
	AuditStatus       *int32  `json:"auditStatus,omitempty"`       // 可选审核状态
	AuditRemark       *string `json:"auditRemark,omitempty"`       // 可选驳回原因/审核备注
	AuditedBy         *int64  `json:"auditedBy,omitempty"`         // 可选审核人（管理员 ID）
	AuditedAt         *int64  `json:"auditedAt,omitempty"`         // 可选审核时间（Unix 秒）
}

// UpdateCertificationResponse 更新认证响应。
type UpdateCertificationResponse struct {
	ID         int64 `json:"id"`         // 认证 ID
	AuditStatus int32 `json:"auditStatus"` // 更新后的审核状态
	UpdatedAt  int64 `json:"updatedAt"`   // 更新时间（Unix 秒）
}

// CertificationDetail 认证完整信息。
type CertificationDetail struct {
	ID                int64  `json:"id"`                // 认证 ID
	DriverID          int64  `json:"driverId"`          // 所属司机 ID
	VehicleID         int64  `json:"vehicleId"`         // 关联车辆 ID
	IdCardFrontURL    string `json:"idCardFrontUrl"`    // 身份证人像面
	IdCardBackURL     string `json:"idCardBackUrl"`     // 身份证国徽面
	DriverLicenseURL  string `json:"driverLicenseUrl"`  // 驾驶证照片
	VehicleLicenseURL string `json:"vehicleLicenseUrl"` // 行驶证照片
	AuditStatus       int32  `json:"auditStatus"`       // 审核状态：1待审核 2通过 3驳回
	AuditRemark       string `json:"auditRemark"`       // 驳回原因/审核备注
	AuditedBy         int64  `json:"auditedBy"`         // 审核人（管理员 ID）
	AuditedAt         int64  `json:"auditedAt"`         // 审核时间（Unix 秒）
	CreatedAt         int64  `json:"createdAt"`         // 创建时间（Unix 秒）
	UpdatedAt         int64  `json:"updatedAt"`         // 更新时间（Unix 秒）
}

// GetCertificationResponse 查询认证详情响应。
type GetCertificationResponse struct {
	Certification CertificationDetail `json:"certification"` // 认证完整信息
}

// CertificationSummary 认证列表摘要项。
type CertificationSummary struct {
	ID          int64 `json:"id"`          // 认证 ID
	DriverID    int64 `json:"driverId"`    // 所属司机 ID
	VehicleID   int64 `json:"vehicleId"`   // 关联车辆 ID
	AuditStatus int32 `json:"auditStatus"` // 审核状态
	CreatedAt   int64 `json:"createdAt"`   // 创建时间（Unix 秒）
}

// ListCertificationsRequest 分页查询认证列表请求。
type ListCertificationsRequest struct {
	DriverID    int64 `json:"driverId,omitempty"`    // 按司机过滤（可选）
	AuditStatus int32 `json:"auditStatus,omitempty"` // 按审核状态过滤（可选）
	Page        int32 `json:"page,omitempty"`        // 页码，默认 1
	PageSize    int32 `json:"pageSize,omitempty"`    // 每页条数，默认 20，上限 100
}

// ListCertificationsResponse 分页查询认证列表响应。
type ListCertificationsResponse struct {
	List     []CertificationSummary `json:"list"`     // 认证摘要列表
	Total    int64                  `json:"total"`    // 符合条件的总记录数
	Page     int32                  `json:"page"`     // 当前页码
	PageSize int32                  `json:"pageSize"` // 每页条数
}

// ============ 服务分 ============

// CreateScoreRequest 创建服务分请求。
type CreateScoreRequest struct {
	DriverID            int64   `json:"driverId"`            // 所属司机 ID
	Score               float64 `json:"score"`               // 服务分（0~100）
	Level               int32   `json:"level"`               // 司机等级：1-5
	MonthOrders         int32   `json:"monthOrders"`         // 当月完单数
	MonthCancelRate     float64 `json:"monthCancelRate"`     // 当月取消率（%）
	MonthComplaintCount int32   `json:"monthComplaintCount"` // 当月投诉数
}

// CreateScoreResponse 创建服务分响应。
type CreateScoreResponse struct {
	ID       int64 `json:"id"`       // 新建记录 ID
	DriverID int64 `json:"driverId"` // 所属司机 ID
}

// UpdateScoreRequest 更新服务分请求，除 id 外字段均为可选。
type UpdateScoreRequest struct {
	ID                 int64    `json:"id"`                 // 待更新记录 ID
	DriverID           *int64   `json:"driverId,omitempty"` // 可选所属司机 ID
	Score              *float64 `json:"score,omitempty"`     // 可选服务分
	Level              *int32   `json:"level,omitempty"`     // 可选司机等级
	MonthOrders        *int32   `json:"monthOrders,omitempty"`        // 可选当月完单数
	MonthCancelRate    *float64 `json:"monthCancelRate,omitempty"`    // 可选当月取消率
	MonthComplaintCount *int32  `json:"monthComplaintCount,omitempty"` // 可选当月投诉数
}

// UpdateScoreResponse 更新服务分响应。
type UpdateScoreResponse struct {
	ID        int64 `json:"id"`        // 记录 ID
	DriverID  int64 `json:"driverId"`  // 所属司机 ID
	UpdatedAt int64 `json:"updatedAt"` // 更新时间（Unix 秒）
}

// ScoreDetail 服务分完整信息。
type ScoreDetail struct {
	ID                 int64   `json:"id"`                 // 主键 ID
	DriverID           int64   `json:"driverId"`           // 所属司机 ID
	Score              float64 `json:"score"`              // 服务分
	Level              int32   `json:"level"`              // 司机等级
	MonthOrders        int32   `json:"monthOrders"`        // 当月完单数
	MonthCancelRate    float64 `json:"monthCancelRate"`    // 当月取消率
	MonthComplaintCount int32  `json:"monthComplaintCount"` // 当月投诉数
	UpdatedAt          int64   `json:"updatedAt"`          // 更新时间（Unix 秒）
}

// GetScoreResponse 查询服务分详情响应。
type GetScoreResponse struct {
	Score ScoreDetail `json:"score"` // 服务分完整信息
}

// ScoreSummary 服务分列表摘要项。
type ScoreSummary struct {
	ID          int64   `json:"id"`          // 记录 ID
	DriverID    int64   `json:"driverId"`    // 所属司机 ID
	Score       float64 `json:"score"`       // 服务分
	Level       int32   `json:"level"`       // 司机等级
	MonthOrders int32   `json:"monthOrders"` // 当月完单数
}

// ListScoresRequest 分页查询服务分列表请求。
type ListScoresRequest struct {
	DriverID int64 `json:"driverId,omitempty"` // 按司机过滤（可选）
	Page     int32 `json:"page,omitempty"`     // 页码，默认 1
	PageSize int32 `json:"pageSize,omitempty"` // 每页条数，默认 20，上限 100
}

// ListScoresResponse 分页查询服务分列表响应。
type ListScoresResponse struct {
	List     []ScoreSummary `json:"list"`     // 服务分摘要列表
	Total    int64          `json:"total"`    // 符合条件的总记录数
	Page     int32          `json:"page"`     // 当前页码
	PageSize int32          `json:"pageSize"` // 每页条数
}

// ============ 提现 ============

// CreateWithdrawRequest 创建提现请求。
type CreateWithdrawRequest struct {
	DriverID   int64   `json:"driverId"`   // 所属司机 ID
	WithdrawNo string  `json:"withdrawNo"` // 提现单号
	Amount     float64 `json:"amount"`     // 提现金额
	PayeeName  string  `json:"payeeName"`  // 收款人姓名
	PayAccount string  `json:"payAccount"` // 收款账号
}

// CreateWithdrawResponse 创建提现响应。
type CreateWithdrawResponse struct {
	ID         int64  `json:"id"`         // 新建提现 ID
	WithdrawNo string `json:"withdrawNo"` // 提现单号
	Status     int32  `json:"status"`     // 初始状态（1申请中）
}

// UpdateWithdrawRequest 更新提现请求，除 id 外字段均为可选。
type UpdateWithdrawRequest struct {
	ID          int64    `json:"id"`          // 待更新提现 ID
	DriverID    *int64   `json:"driverId,omitempty"`    // 可选所属司机 ID
	WithdrawNo  *string  `json:"withdrawNo,omitempty"`  // 可选提现单号
	Amount      *float64 `json:"amount,omitempty"`      // 可选提现金额
	PayeeName   *string  `json:"payeeName,omitempty"`   // 可选收款人姓名
	PayAccount  *string  `json:"payAccount,omitempty"`  // 可选收款账号
	Status      *int32   `json:"status,omitempty"`      // 可选状态
	Remark      *string  `json:"remark,omitempty"`      // 可选失败原因/备注
	AppliedAt   *int64   `json:"appliedAt,omitempty"`   // 可选申请时间（Unix 秒）
	PaidAt      *int64   `json:"paidAt,omitempty"`      // 可选打款时间（Unix 秒）
}

// UpdateWithdrawResponse 更新提现响应。
type UpdateWithdrawResponse struct {
	ID     int64 `json:"id"`     // 提现 ID
	Status int32 `json:"status"` // 更新后的状态
}

// WithdrawDetail 提现完整信息。
type WithdrawDetail struct {
	ID          int64   `json:"id"`          // 提现 ID
	DriverID    int64   `json:"driverId"`    // 所属司机 ID
	WithdrawNo  string  `json:"withdrawNo"`  // 提现单号
	Amount      float64 `json:"amount"`      // 提现金额
	PayeeName   string  `json:"payeeName"`   // 收款人姓名
	PayAccount  string  `json:"payAccount"`  // 收款账号
	Status      int32   `json:"status"`      // 状态：1申请中 2打款成功 3打款失败
	Remark      string  `json:"remark"`      // 失败原因/备注
	AppliedAt   int64   `json:"appliedAt"`   // 申请时间（Unix 秒）
	PaidAt      int64   `json:"paidAt"`      // 打款时间（Unix 秒）
	CreatedAt   int64   `json:"createdAt"`   // 创建时间（Unix 秒）
}

// GetWithdrawResponse 查询提现详情响应。
type GetWithdrawResponse struct {
	Withdraw WithdrawDetail `json:"withdraw"` // 提现完整信息
}

// WithdrawSummary 提现列表摘要项。
type WithdrawSummary struct {
	ID         int64   `json:"id"`         // 提现 ID
	DriverID   int64   `json:"driverId"`   // 所属司机 ID
	WithdrawNo string  `json:"withdrawNo"` // 提现单号
	Amount     float64 `json:"amount"`     // 提现金额
	Status     int32   `json:"status"`     // 状态
	AppliedAt  int64   `json:"appliedAt"`  // 申请时间（Unix 秒）
}

// ListWithdrawsRequest 分页查询提现列表请求。
type ListWithdrawsRequest struct {
	DriverID int64 `json:"driverId,omitempty"` // 按司机过滤（可选）
	Status   int32 `json:"status,omitempty"`   // 按状态过滤（可选）
	Page     int32 `json:"page,omitempty"`     // 页码，默认 1
	PageSize int32 `json:"pageSize,omitempty"` // 每页条数，默认 20，上限 100
}

// ListWithdrawsResponse 分页查询提现列表响应。
type ListWithdrawsResponse struct {
	List     []WithdrawSummary `json:"list"`     // 提现摘要列表
	Total    int64             `json:"total"`    // 符合条件的总记录数
	Page     int32             `json:"page"`     // 当前页码
	PageSize int32             `json:"pageSize"` // 每页条数
}
