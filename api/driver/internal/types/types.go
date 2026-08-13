package types

// Response 是接口文档约定的通用响应格式（与 passenger 端保持一致）。
type Response struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	Data      any    `json:"data"`
	Timestamp int64  `json:"timestamp"`
	TraceID   string `json:"traceId"`
}

// ============ 司机 ============

type CreateDriverRequest struct {
	Phone           string `json:"phone"`
	PasswordHash    string `json:"passwordHash"`
	RealName        string `json:"realName"`
	IdCardNo        string `json:"idCardNo"`
	DriverLicenseNo string `json:"driverLicenseNo"`
	AvatarURL       string `json:"avatarUrl"`
}

type CreateDriverResponse struct {
	ID        int64  `json:"id"`
	Status    string `json:"status"`
	CreatedAt int64  `json:"createdAt"`
}

type UpdateDriverRequest struct {
	ID              int64   `json:"id"`
	Phone           *string `json:"phone,omitempty"`
	PasswordHash    *string `json:"passwordHash,omitempty"`
	RealName        *string `json:"realName,omitempty"`
	IdCardNo        *string `json:"idCardNo,omitempty"`
	DriverLicenseNo *string `json:"driverLicenseNo,omitempty"`
	AvatarURL       *string `json:"avatarUrl,omitempty"`
	Status          *string `json:"status,omitempty"`
}

type UpdateDriverResponse struct {
	ID        int64  `json:"id"`
	Status    string `json:"status"`
	UpdatedAt int64  `json:"updatedAt"`
}

type DriverDetail struct {
	ID              int64  `json:"id"`
	Phone           string `json:"phone"`
	RealName        string `json:"realName"`
	IdCardNo        string `json:"idCardNo"`
	DriverLicenseNo string `json:"driverLicenseNo"`
	AvatarURL       string `json:"avatarUrl"`
	Status          string `json:"status"`
	CreatedAt       int64  `json:"createdAt"`
	UpdatedAt       int64  `json:"updatedAt"`
}

type GetDriverResponse struct {
	Driver DriverDetail `json:"driver"`
}

type DeleteResponse struct {
	ID      int64 `json:"id"`
	Success bool  `json:"success"`
}

type DriverSummary struct {
	ID              int64  `json:"id"`
	Phone           string `json:"phone"`
	RealName        string `json:"realName"`
	DriverLicenseNo string `json:"driverLicenseNo"`
	Status          string `json:"status"`
	CreatedAt       int64  `json:"createdAt"`
}

type ListDriversRequest struct {
	Status       string `json:"status,omitempty"`
	PhoneKeyword string `json:"phoneKeyword,omitempty"`
	Page         int32  `json:"page,omitempty"`
	PageSize     int32  `json:"pageSize,omitempty"`
}

type ListDriversResponse struct {
	List     []DriverSummary `json:"list"`
	Total    int64           `json:"total"`
	Page     int32           `json:"page"`
	PageSize int32           `json:"pageSize"`
}

// ============ 车辆 ============

type CreateVehicleRequest struct {
	DriverID           int64  `json:"driverId"`
	PlateNo            string `json:"plateNo"`
	Brand              string `json:"brand"`
	Model              string `json:"model"`
	Color              string `json:"color"`
	VehicleType        int32  `json:"vehicleType"`
	RegistrationDate   string `json:"registrationDate"`
	InsuranceNo        string `json:"insuranceNo"`
	InsuranceExpireAt  string `json:"insuranceExpireAt"`
}

type CreateVehicleResponse struct {
	ID     int64 `json:"id"`
	Status int32 `json:"status"`
}

type UpdateVehicleRequest struct {
	ID                 int64   `json:"id"`
	DriverID           *int64  `json:"driverId,omitempty"`
	PlateNo            *string `json:"plateNo,omitempty"`
	Brand              *string `json:"brand,omitempty"`
	Model              *string `json:"model,omitempty"`
	Color              *string `json:"color,omitempty"`
	VehicleType        *int32  `json:"vehicleType,omitempty"`
	RegistrationDate   *string `json:"registrationDate,omitempty"`
	InsuranceNo        *string `json:"insuranceNo,omitempty"`
	InsuranceExpireAt  *string `json:"insuranceExpireAt,omitempty"`
	Status             *int32  `json:"status,omitempty"`
}

type UpdateVehicleResponse struct {
	ID        int64 `json:"id"`
	Status    int32 `json:"status"`
	UpdatedAt int64 `json:"updatedAt"`
}

type VehicleDetail struct {
	ID                int64  `json:"id"`
	DriverID          int64  `json:"driverId"`
	PlateNo           string `json:"plateNo"`
	Brand             string `json:"brand"`
	Model             string `json:"model"`
	Color             string `json:"color"`
	VehicleType       int32  `json:"vehicleType"`
	RegistrationDate  string `json:"registrationDate"`
	InsuranceNo       string `json:"insuranceNo"`
	InsuranceExpireAt string `json:"insuranceExpireAt"`
	Status            int32  `json:"status"`
	CreatedAt         int64  `json:"createdAt"`
	UpdatedAt         int64  `json:"updatedAt"`
}

type GetVehicleResponse struct {
	Vehicle VehicleDetail `json:"vehicle"`
}

type VehicleSummary struct {
	ID          int64  `json:"id"`
	DriverID    int64  `json:"driverId"`
	PlateNo     string `json:"plateNo"`
	Brand       string `json:"brand"`
	VehicleType int32  `json:"vehicleType"`
	Status      int32  `json:"status"`
}

type ListVehiclesRequest struct {
	DriverID  int64 `json:"driverId,omitempty"`
	Status    int32 `json:"status,omitempty"`
	Page      int32 `json:"page,omitempty"`
	PageSize  int32 `json:"pageSize,omitempty"`
}

type ListVehiclesResponse struct {
	List     []VehicleSummary `json:"list"`
	Total    int64            `json:"total"`
	Page     int32            `json:"page"`
	PageSize int32            `json:"pageSize"`
}

// ============ 认证 ============

type CreateCertificationRequest struct {
	DriverID          int64  `json:"driverId"`
	VehicleID         int64  `json:"vehicleId"`
	IdCardFrontURL    string `json:"idCardFrontUrl"`
	IdCardBackURL     string `json:"idCardBackUrl"`
	DriverLicenseURL  string `json:"driverLicenseUrl"`
	VehicleLicenseURL string `json:"vehicleLicenseUrl"`
}

type CreateCertificationResponse struct {
	ID         int64 `json:"id"`
	AuditStatus int32 `json:"auditStatus"`
}

type UpdateCertificationRequest struct {
	ID                int64   `json:"id"`
	DriverID          *int64  `json:"driverId,omitempty"`
	VehicleID         *int64  `json:"vehicleId,omitempty"`
	IdCardFrontURL    *string `json:"idCardFrontUrl,omitempty"`
	IdCardBackURL     *string `json:"idCardBackUrl,omitempty"`
	DriverLicenseURL  *string `json:"driverLicenseUrl,omitempty"`
	VehicleLicenseURL *string `json:"vehicleLicenseUrl,omitempty"`
	AuditStatus       *int32  `json:"auditStatus,omitempty"`
	AuditRemark       *string `json:"auditRemark,omitempty"`
	AuditedBy         *int64  `json:"auditedBy,omitempty"`
	AuditedAt         *int64  `json:"auditedAt,omitempty"`
}

type UpdateCertificationResponse struct {
	ID         int64 `json:"id"`
	AuditStatus int32 `json:"auditStatus"`
	UpdatedAt  int64 `json:"updatedAt"`
}

type CertificationDetail struct {
	ID                int64  `json:"id"`
	DriverID          int64  `json:"driverId"`
	VehicleID         int64  `json:"vehicleId"`
	IdCardFrontURL    string `json:"idCardFrontUrl"`
	IdCardBackURL     string `json:"idCardBackUrl"`
	DriverLicenseURL  string `json:"driverLicenseUrl"`
	VehicleLicenseURL string `json:"vehicleLicenseUrl"`
	AuditStatus       int32  `json:"auditStatus"`
	AuditRemark       string `json:"auditRemark"`
	AuditedBy         int64  `json:"auditedBy"`
	AuditedAt         int64  `json:"auditedAt"`
	CreatedAt         int64  `json:"createdAt"`
	UpdatedAt         int64  `json:"updatedAt"`
}

type GetCertificationResponse struct {
	Certification CertificationDetail `json:"certification"`
}

type CertificationSummary struct {
	ID          int64 `json:"id"`
	DriverID    int64 `json:"driverId"`
	VehicleID   int64 `json:"vehicleId"`
	AuditStatus int32 `json:"auditStatus"`
	CreatedAt   int64 `json:"createdAt"`
}

type ListCertificationsRequest struct {
	DriverID    int64 `json:"driverId,omitempty"`
	AuditStatus int32 `json:"auditStatus,omitempty"`
	Page        int32 `json:"page,omitempty"`
	PageSize    int32 `json:"pageSize,omitempty"`
}

type ListCertificationsResponse struct {
	List     []CertificationSummary `json:"list"`
	Total    int64                  `json:"total"`
	Page     int32                  `json:"page"`
	PageSize int32                  `json:"pageSize"`
}

// ============ 服务分 ============

type CreateScoreRequest struct {
	DriverID           int64   `json:"driverId"`
	Score              float64 `json:"score"`
	Level              int32   `json:"level"`
	MonthOrders        int32   `json:"monthOrders"`
	MonthCancelRate    float64 `json:"monthCancelRate"`
	MonthComplaintCount int32  `json:"monthComplaintCount"`
}

type CreateScoreResponse struct {
	ID       int64 `json:"id"`
	DriverID int64 `json:"driverId"`
}

type UpdateScoreRequest struct {
	ID                 int64    `json:"id"`
	DriverID           *int64   `json:"driverId,omitempty"`
	Score              *float64 `json:"score,omitempty"`
	Level              *int32   `json:"level,omitempty"`
	MonthOrders        *int32   `json:"monthOrders,omitempty"`
	MonthCancelRate    *float64 `json:"monthCancelRate,omitempty"`
	MonthComplaintCount *int32  `json:"monthComplaintCount,omitempty"`
}

type UpdateScoreResponse struct {
	ID        int64 `json:"id"`
	DriverID  int64 `json:"driverId"`
	UpdatedAt int64 `json:"updatedAt"`
}

type ScoreDetail struct {
	ID                 int64   `json:"id"`
	DriverID           int64   `json:"driverId"`
	Score              float64 `json:"score"`
	Level              int32   `json:"level"`
	MonthOrders        int32   `json:"monthOrders"`
	MonthCancelRate    float64 `json:"monthCancelRate"`
	MonthComplaintCount int32  `json:"monthComplaintCount"`
	UpdatedAt          int64   `json:"updatedAt"`
}

type GetScoreResponse struct {
	Score ScoreDetail `json:"score"`
}

type ScoreSummary struct {
	ID          int64   `json:"id"`
	DriverID    int64   `json:"driverId"`
	Score       float64 `json:"score"`
	Level       int32   `json:"level"`
	MonthOrders int32   `json:"monthOrders"`
}

type ListScoresRequest struct {
	DriverID int64 `json:"driverId,omitempty"`
	Page     int32 `json:"page,omitempty"`
	PageSize int32 `json:"pageSize,omitempty"`
}

type ListScoresResponse struct {
	List     []ScoreSummary `json:"list"`
	Total    int64          `json:"total"`
	Page     int32          `json:"page"`
	PageSize int32          `json:"pageSize"`
}

// ============ 提现 ============

type CreateWithdrawRequest struct {
	DriverID   int64   `json:"driverId"`
	WithdrawNo string  `json:"withdrawNo"`
	Amount     float64 `json:"amount"`
	PayeeName  string  `json:"payeeName"`
	PayAccount string  `json:"payAccount"`
}

type CreateWithdrawResponse struct {
	ID         int64  `json:"id"`
	WithdrawNo string `json:"withdrawNo"`
	Status     int32  `json:"status"`
}

type UpdateWithdrawRequest struct {
	ID          int64    `json:"id"`
	DriverID    *int64   `json:"driverId,omitempty"`
	WithdrawNo  *string  `json:"withdrawNo,omitempty"`
	Amount      *float64 `json:"amount,omitempty"`
	PayeeName   *string  `json:"payeeName,omitempty"`
	PayAccount  *string  `json:"payAccount,omitempty"`
	Status      *int32   `json:"status,omitempty"`
	Remark      *string  `json:"remark,omitempty"`
	AppliedAt   *int64   `json:"appliedAt,omitempty"`
	PaidAt      *int64   `json:"paidAt,omitempty"`
}

type UpdateWithdrawResponse struct {
	ID     int64 `json:"id"`
	Status int32 `json:"status"`
}

type WithdrawDetail struct {
	ID          int64   `json:"id"`
	DriverID    int64   `json:"driverId"`
	WithdrawNo  string  `json:"withdrawNo"`
	Amount      float64 `json:"amount"`
	PayeeName   string  `json:"payeeName"`
	PayAccount  string  `json:"payAccount"`
	Status      int32   `json:"status"`
	Remark      string  `json:"remark"`
	AppliedAt   int64   `json:"appliedAt"`
	PaidAt      int64   `json:"paidAt"`
	CreatedAt   int64   `json:"createdAt"`
}

type GetWithdrawResponse struct {
	Withdraw WithdrawDetail `json:"withdraw"`
}

type WithdrawSummary struct {
	ID         int64   `json:"id"`
	DriverID   int64   `json:"driverId"`
	WithdrawNo string  `json:"withdrawNo"`
	Amount     float64 `json:"amount"`
	Status     int32   `json:"status"`
	AppliedAt  int64   `json:"appliedAt"`
}

type ListWithdrawsRequest struct {
	DriverID int64 `json:"driverId,omitempty"`
	Status   int32 `json:"status,omitempty"`
	Page     int32 `json:"page,omitempty"`
	PageSize int32 `json:"pageSize,omitempty"`
}

type ListWithdrawsResponse struct {
	List     []WithdrawSummary `json:"list"`
	Total    int64             `json:"total"`
	Page     int32             `json:"page"`
	PageSize int32             `json:"pageSize"`
}
