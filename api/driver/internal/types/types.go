package types

import "encoding/json"

type Response struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data"`
	Timestamp int64       `json:"timestamp"`
	TraceID   string      `json:"traceId"`
}

type RegisterDriverRequest struct {
	Phone           string `json:"phone"`
	Password        string `json:"password"`
	RealName        string `json:"realName"`
	IdCardNo        string `json:"idCardNo"`
	DriverLicenseNo string `json:"driverLicenseNo"`
	AvatarURL       string `json:"avatarUrl"`
}

func (r *RegisterDriverRequest) UnmarshalJSON(data []byte) error {
	type alias RegisterDriverRequest
	var raw struct {
		alias
		PasswordHash          string `json:"password_hash"`
		RealNameLegacy        string `json:"real_name"`
		IdCardNoLegacy        string `json:"id_card_no"`
		DriverLicenseNoLegacy string `json:"driver_license_no"`
		AvatarURLLegacy       string `json:"avatar_url"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*r = RegisterDriverRequest(raw.alias)
	if r.Password == "" {
		r.Password = raw.PasswordHash
	}
	if r.RealName == "" {
		r.RealName = raw.RealNameLegacy
	}
	if r.IdCardNo == "" {
		r.IdCardNo = raw.IdCardNoLegacy
	}
	if r.DriverLicenseNo == "" {
		r.DriverLicenseNo = raw.DriverLicenseNoLegacy
	}
	if r.AvatarURL == "" {
		r.AvatarURL = raw.AvatarURLLegacy
	}
	return nil
}

type RegisterDriverResponse struct {
	ID        int64  `json:"id"`
	Status    string `json:"status"`
	CreatedAt int64  `json:"createdAt"`
}

type CreateDriverRequest struct {
	Phone           string `json:"phone"`
	Password        string `json:"password"`
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
	Password        *string `json:"password,omitempty"`
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

type GetDriverRequest struct {
	ID int64 `form:"id"`
}

type DriverDetail struct {
	ID              int64  `json:"id"`
	Phone           string `json:"phone"`
	RealName        string `json:"realName"`
	IdCardNo        string `json:"idCardNo"`
	DriverLicenseNo string `json:"driverLicenseNo"`
	AvatarURL       string `json:"avatarUrl"`
	Status          string `json:"status"`
	OnlineStatus    int    `json:"onlineStatus"`
	VehicleID       int64  `json:"vehicleId"`
	CreatedAt       int64  `json:"createdAt"`
	UpdatedAt       int64  `json:"updatedAt"`
}

type GetDriverResponse struct {
	Driver DriverDetail `json:"driver"`
}

type DeleteDriverRequest struct {
	ID int64 `form:"id"`
}

type DeleteResponse struct {
	ID      int64 `json:"id"`
	Success bool  `json:"success"`
}

type SendSMSCodeRequest struct {
	Phone string `json:"phone"`
}

// SendSMSCodeResponse 发码结果。联调阶段验证码仅打印到服务端日志，不随响应返回给客户端，
// 避免验证码明文泄露。
type SendSMSCodeResponse struct {
	Success  bool `json:"success"`
	ExpireIn int  `json:"expireIn"`
}

type LoginByPasswordRequest struct {
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

type LoginBySMSRequest struct {
	Phone string `json:"phone"`
	Code  string `json:"code"`
}

type DriverBrief struct {
	ID        int64  `json:"id"`
	Phone     string `json:"phone"`
	Status    string `json:"status"`
	VehicleID int64  `json:"vehicleId"`
}

type LoginResponse struct {
	Token    string      `json:"token"`
	ExpireIn int64       `json:"expireIn"`
	Driver   DriverBrief `json:"driver"`
}

type SetOnlineRequest struct {
	DeviceID  string  `json:"deviceId"`
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
}

type SetOfflineRequest struct {
	DeviceID  string  `json:"deviceId"`
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
}

type SetOnlineResponse struct {
	DriverID     int64 `json:"driverId"`
	OnlineStatus int   `json:"onlineStatus"`
	Kicked       bool  `json:"kicked"`
}

type SetOfflineResponse struct {
	DriverID     int64 `json:"driverId"`
	OnlineStatus int   `json:"onlineStatus"`
	Kicked       bool  `json:"kicked"`
}

type HeartbeatRequest struct {
	DeviceID  string  `json:"deviceId"`
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
}

type HeartbeatResponse struct {
	OnlineStatus int   `json:"onlineStatus"`
	Kicked       bool  `json:"kicked"`
	ServerTime   int64 `json:"serverTime"`
}

type ReportLocationRequest struct {
	DeviceID  string  `json:"deviceId"`
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
	Heading   int32   `json:"heading"`
	SpeedKmh  float64 `json:"speedKmh"`
	OrderID   int64   `json:"orderId"`
}

type ReportLocationResponse struct {
	DriverID     int64 `json:"driverId"`
	OnlineStatus int   `json:"onlineStatus"`
	Kicked       bool  `json:"kicked"`
	ReportTime   int64 `json:"reportTime"`
}

type AiScoreFactor struct {
	Key    string  `json:"key"`
	Label  string  `json:"label"`
	Value  float64 `json:"value"`
	Impact string  `json:"impact"`
	Hint   string  `json:"hint"`
}

type GetDriverAiScoreRequest struct {
	DriverID int64 `form:"id"`
}

type GetDriverAiScoreResponse struct {
	DriverID      int64           `json:"driverId"`
	AiScore       float64         `json:"aiScore"`
	Level         int32           `json:"level"`
	Factors       []AiScoreFactor `json:"factors"`
	Degraded      bool            `json:"degraded"`
	DegradeReason string          `json:"degradeReason"`
}

type UploadCertificationRequest struct {
	VehicleID      int64  `json:"vehicleId"`
	IdCardFront    string `json:"idCardFront"`
	IdCardBack     string `json:"idCardBack"`
	DriverLicense  string `json:"driverLicense"`
	VehicleLicense string `json:"vehicleLicense"`
}

type CertificationInfo struct {
	ID                int64  `json:"id"`
	DriverID          int64  `json:"driverId"`
	VehicleID         int64  `json:"vehicleId"`
	IdCardFrontURL    string `json:"idCardFrontUrl"`
	IdCardBackURL     string `json:"idCardBackUrl"`
	DriverLicenseURL  string `json:"driverLicenseUrl"`
	VehicleLicenseURL string `json:"vehicleLicenseUrl"`
	AuditStatus       int    `json:"auditStatus"`
	AuditRemark       string `json:"auditRemark"`
}

type UploadCertificationResponse struct {
	ID            int64              `json:"id"`
	Certification *CertificationInfo `json:"certification"`
}

type GetCertificationResponse struct {
	Certification *CertificationInfo `json:"certification"`
	Found         bool               `json:"found"`
}

type CreateVehicleRequest struct {
	PlateNo           string `json:"plateNo"`
	Brand             string `json:"brand"`
	Model             string `json:"model"`
	Color             string `json:"color"`
	VehicleType       int32  `json:"vehicleType"`
	RegistrationDate  int64  `json:"registrationDate,omitempty"`
	InsuranceNo       string `json:"insuranceNo"`
	InsuranceExpireAt int64  `json:"insuranceExpireAt,omitempty"`
}

type CreateVehicleResponse struct {
	ID        int64  `json:"id"`
	Status    string `json:"status"`
	CreatedAt int64  `json:"createdAt"`
}

type GetVehicleRequest struct {
	ID int64 `form:"id"`
}

type VehicleInfo struct {
	ID                int64  `json:"id"`
	DriverID          int64  `json:"driverId"`
	PlateNo           string `json:"plateNo"`
	Brand             string `json:"brand"`
	Model             string `json:"model"`
	Color             string `json:"color"`
	VehicleType       int32  `json:"vehicleType"`
	RegistrationDate  int64  `json:"registrationDate,omitempty"`
	InsuranceNo       string `json:"insuranceNo"`
	InsuranceExpireAt int64  `json:"insuranceExpireAt,omitempty"`
	Status            string `json:"status"`
	CreatedAt         int64  `json:"createdAt"`
	UpdatedAt         int64  `json:"updatedAt"`
}

type GetVehicleResponse struct {
	Vehicle VehicleInfo `json:"vehicle"`
}

type UpdateVehicleRequest struct {
	ID                int64   `json:"id"`
	DriverID          *int64  `json:"driverId,omitempty"`
	PlateNo           *string `json:"plateNo,omitempty"`
	Brand             *string `json:"brand,omitempty"`
	Model             *string `json:"model,omitempty"`
	Color             *string `json:"color,omitempty"`
	VehicleType       *int32  `json:"vehicleType,omitempty"`
	RegistrationDate  *int64  `json:"registrationDate,omitempty"`
	InsuranceNo       *string `json:"insuranceNo,omitempty"`
	InsuranceExpireAt *int64  `json:"insuranceExpireAt,omitempty"`
	Status            *string `json:"status,omitempty"`
}

type UpdateVehicleResponse struct {
	ID        int64  `json:"id"`
	Status    string `json:"status"`
	UpdatedAt int64  `json:"updatedAt"`
}

type DeleteVehicleRequest struct {
	ID int64 `json:"id" form:"id"`
}

type DeleteVehicleResponse struct {
	ID      int64 `json:"id"`
	Success bool  `json:"success"`
}

type GetDriverByPhoneRequest struct {
	Phone string `json:"phone" form:"phone"`
}

type GetDriverByPhoneResponse struct {
	Driver DriverDetail `json:"driver"`
}

type NearbyDriver struct {
	DriverID       int64   `json:"driverId"`
	Longitude      float64 `json:"longitude"`
	Latitude       float64 `json:"latitude"`
	DistanceMeters int32   `json:"distanceMeters"`
}

type ListNearbyDriversRequest struct {
	Longitude    float64 `json:"longitude"`
	Latitude     float64 `json:"latitude"`
	RadiusMeters float64 `json:"radiusMeters"`
	Limit        int32   `json:"limit"`
}

type ListNearbyDriversResponse struct {
	List    []NearbyDriver `json:"list"`
	Drivers []NearbyDriver `json:"drivers"`
}

type CommonResponse struct {
	Message string `json:"message"`
}

// ---- Withdraw and income ----

type CreateWithdrawRequest struct {
	Amount     float64 `json:"amount"`
	PayeeName  string  `json:"payeeName"`
	PayAccount string  `json:"payAccount"`
}

type CreateWithdrawResponse struct {
	ID         int64  `json:"id"`
	WithdrawNo string `json:"withdrawNo"`
	Status     int32  `json:"status"`
	CreatedAt  int64  `json:"createdAt"`
}

type ListWithdrawsRequest struct {
	Page     int32 `json:"page"`
	PageSize int32 `json:"pageSize"`
}

type WithdrawRecord struct {
	ID         int64   `json:"id"`
	WithdrawNo string  `json:"withdrawNo"`
	Amount     float64 `json:"amount"`
	PayeeName  string  `json:"payeeName"`
	PayAccount string  `json:"payAccount"`
	Status     int32   `json:"status"`
	Remark     string  `json:"remark"`
	AppliedAt  int64   `json:"appliedAt"`
	PaidAt     int64   `json:"paidAt"`
	CreatedAt  int64   `json:"createdAt"`
}

type ListWithdrawsResponse struct {
	List  []WithdrawRecord `json:"list"`
	Total int64            `json:"total"`
}

type GetIncomeSummaryResponse struct {
	DriverID         int64  `json:"driverId"`
	CompletedOrders  int64  `json:"completedOrders"`
	TotalIncomeCents int64  `json:"totalIncomeCents"`
	Source           string `json:"source"`
}

type PeriodIncomeResponse struct {
	DriverID         int64  `json:"driverId"`
	Period           string `json:"period"`
	CompletedOrders  int64  `json:"completedOrders"`
	TotalIncomeCents int64  `json:"totalIncomeCents"`
	StartAt          int64  `json:"startAt"`
	EndAt            int64  `json:"endAt"`
	Source           string `json:"source"`
}
type ListIncomeBillsRequest struct {
	Page     int32 `json:"page"`
	PageSize int32 `json:"pageSize"`
}

type IncomeBill struct {
	OrderID     int64  `json:"orderId"`
	OrderNo     string `json:"orderNo"`
	IncomeCents int64  `json:"incomeCents"`
	Status      int32  `json:"status"`
	CreatedAt   int64  `json:"createdAt"`
}

type ListIncomeBillsResponse struct {
	List     []IncomeBill `json:"list"`
	Total    int64        `json:"total"`
	Page     int32        `json:"page"`
	PageSize int32        `json:"pageSize"`
	Source   string       `json:"source"`
}

type AcceptOrderRequest struct {
	OrderID int64 `json:"orderId"`
}

type AcceptOrderResponse struct {
	OrderID int64 `json:"orderId"`
	Status  int32 `json:"status"`
}

type CancelOrderRequest struct {
	OrderID int64  `json:"orderId"`
	Reason  string `json:"reason"`
}

type CancelOrderResponse struct {
	OrderID int64 `json:"orderId"`
	Status  int32 `json:"status"`
}

type ConfirmArriveRequest struct {
	OrderID int64 `json:"orderId"`
}

type ConfirmArriveResponse struct {
	OrderID int64 `json:"orderId"`
	Status  int32 `json:"status"`
}

type StartTripRequest struct {
	OrderID int64 `json:"orderId"`
}

type StartTripResponse struct {
	OrderID int64 `json:"orderId"`
	Status  int32 `json:"status"`
}

type FinishTripRequest struct {
	OrderID         int64 `json:"orderId"`
	ActualDistanceM int64 `json:"actualDistanceM"`
	ActualDurationS int64 `json:"actualDurationS"`
}

type FinishTripResponse struct {
	OrderID            int64 `json:"orderId"`
	Status             int32 `json:"status"`
	PayableAmountCents int64 `json:"payableAmountCents"`
}

type RejectOrderRequest struct {
	OrderID int64  `json:"orderId"`
	Reason  string `json:"reason"`
}

type RejectOrderResponse struct {
	OrderID  int64 `json:"orderId"`
	DriverID int64 `json:"driverId"`
	Status   int32 `json:"status"`
}

type ListMyDispatchesRequest struct {
	DriverID int64 `json:"driverId,omitempty"`
	Page     int32 `json:"page"`
	PageSize int32 `json:"pageSize"`
	Status   int32 `json:"status"`
}

type DispatchRecord struct {
	ID           int64   `json:"id"`
	OrderID      int64   `json:"orderId"`
	DriverID     int64   `json:"driverId"`
	DispatchType int32   `json:"dispatchType"`
	Status       int32   `json:"status"`
	MatchScore   float64 `json:"matchScore"`
	Remark       string  `json:"remark"`
	RejectReason string  `json:"rejectReason"`
	CreatedAt    int64   `json:"createdAt"`
	UpdatedAt    int64   `json:"updatedAt"`
}

type OrderBrief struct {
	OrderID             int64  `json:"orderId"`
	OrderNo             string `json:"orderNo"`
	FromAddress         string `json:"fromAddress"`
	ToAddress           string `json:"toAddress"`
	Status              int32  `json:"status"`
	EstimatedPriceCents int64  `json:"estimatedPriceCents"`
	DistanceMeters      int64  `json:"distanceMeters"`
	CreatedAt           int64  `json:"createdAt"`
}
type MyDispatchItem struct {
	Dispatch DispatchRecord `json:"dispatch"`
	Order    OrderBrief     `json:"order"`
}

type ListMyDispatchesResponse struct {
	List         []MyDispatchItem `json:"list"`
	Total        int64            `json:"total"`
	Page         int32            `json:"page"`
	PageSize     int32            `json:"pageSize"`
	OrderQueryOk bool             `json:"orderQueryOk"`
}

type ListMyOrdersRequest struct {
	Page     int32 `json:"page"`
	PageSize int32 `json:"pageSize"`
	Status   int32 `json:"status"`
}

type ListMyOrdersResponse struct {
	List     []OrderBrief `json:"list"`
	Total    int64        `json:"total"`
	Page     int32        `json:"page"`
	PageSize int32        `json:"pageSize"`
}

type HeatmapRequest struct {
	Longitude    float64 `json:"longitude"`
	Latitude     float64 `json:"latitude"`
	RadiusMeters float64 `json:"radiusMeters"`
}

type HeatmapPoint struct {
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
	Weight    int64   `json:"weight"`
}

type HeatmapResponse struct {
	Points         []HeatmapPoint `json:"points"`
	RadiusMeters   float64        `json:"radiusMeters"`
	GridSizeMeters int64          `json:"gridSizeMeters"`
	Cached         bool           `json:"cached"`
}

type GetMyOrderDetailRequest struct {
	OrderID int64 `json:"orderId"`
}

type OrderDetail struct {
	OrderID             int64   `json:"orderId"`
	OrderNo             string  `json:"orderNo"`
	UserID              int64   `json:"userId"`
	DriverID            int64   `json:"driverId"`
	CarType             int32   `json:"carType"`
	FromAddress         string  `json:"fromAddress"`
	FromLongitude       float64 `json:"fromLongitude"`
	FromLatitude        float64 `json:"fromLatitude"`
	ToAddress           string  `json:"toAddress"`
	ToLongitude         float64 `json:"toLongitude"`
	ToLatitude          float64 `json:"toLatitude"`
	EstimatedDistanceM  int64   `json:"estimatedDistanceM"`
	EstimatedDurationS  int64   `json:"estimatedDurationS"`
	EstimatedPriceCents int64   `json:"estimatedPriceCents"`
	Status              int32   `json:"status"`
	CancelReason        string  `json:"cancelReason"`
	CancelBy            string  `json:"cancelBy"`
	CreatedAt           int64   `json:"createdAt"`
	UpdatedAt           int64   `json:"updatedAt"`
}

type GetMyOrderDetailResponse struct {
	Order OrderDetail `json:"order"`
}

type ListPassengerReviewsRequest struct {
	Page     int32 `json:"page"`
	PageSize int32 `json:"pageSize"`
}

type PassengerReview struct {
	OrderID   int64  `json:"orderId"`
	Rating    int32  `json:"rating"`
	Comment   string `json:"comment"`
	CreatedAt int64  `json:"createdAt"`
}

type ListPassengerReviewsResponse struct {
	List     []PassengerReview `json:"list"`
	Total    int64             `json:"total"`
	Page     int32             `json:"page"`
	PageSize int32             `json:"pageSize"`
	Degraded bool              `json:"degraded"`
	Message  string            `json:"message"`
}

type GetTripTrajectoryRequest struct {
	OrderID int64 `json:"orderId"`
}

type TrajectoryPoint struct {
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
	SpeedKmh  float64 `json:"speedKmh"`
	Heading   int32   `json:"heading"`
	CreatedAt int64   `json:"createdAt"`
}

type GetTripTrajectoryResponse struct {
	OrderID  int64             `json:"orderId"`
	Points   []TrajectoryPoint `json:"points"`
	Degraded bool              `json:"degraded"`
	Message  string            `json:"message"`
}

type AgentChatRequest struct {
	Question string `json:"question"`
}

type AgentObservation struct {
	ToolName string `json:"toolName"`
	Result   string `json:"result"`
}

type AgentChatResponse struct {
	Answer       string             `json:"answer"`
	LoopCount    int                `json:"loopCount"`
	Observations []AgentObservation `json:"observations"`
	Mode         string             `json:"mode"`
}
