package types

// LoginRequest 表示后台管理员登录请求。
// 该结构体由 /admin/v1/auth/login 接口使用，用于接收用户名和明文密码。
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// RegisterRequest 表示后台管理员注册请求。
// 首个管理员允许无 token 注册，后续管理员注册需要超级管理员身份。
type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	RealName string `json:"real_name"`
	Role     int32  `json:"role"`
}

// AuthResponse 表示登录或注册成功后的统一返回数据。
// token 存储在 Redis 中，前端后续通过 Authorization 头携带。
type AuthResponse struct {
	Token     string   `json:"token"`      // 登录凭证
	ExpiresIn int64    `json:"expires_in"` // token 有效期，单位秒
	Admin     AdminDTO `json:"admin"`      // 当前管理员信息
}

// AdminDTO 是对外暴露的管理员信息。
// 注意这里不包含 password_hash，避免敏感信息泄漏。
type AdminDTO struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	RealName string `json:"real_name"`
	Role     int32  `json:"role"`
	Status   int32  `json:"status"`
}

// MeResponse 表示当前登录管理员信息。
type MeResponse struct {
	Admin AdminDTO `json:"admin"`
}

// MenuItem 表示后台菜单和按钮权限节点。
// P0 阶段先返回固定菜单，后续可演进为角色权限表驱动。
type MenuItem struct {
	Name     string     `json:"name"`
	Path     string     `json:"path"`
	Icon     string     `json:"icon,omitempty"`
	Perm     string     `json:"perm,omitempty"`
	Children []MenuItem `json:"children,omitempty"`
}

// CommonResponse 表示简单操作成功后的返回结构。
type CommonResponse struct {
	Message string `json:"message"`
}

// PageRequest 是后台列表查询的通用分页参数。
type PageRequest struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

// PageResult 是后台列表接口的统一分页返回格式。
type PageResult struct {
	List     any   `json:"list"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
}

// UserListRequest 表示用户列表查询条件。
type UserListRequest struct {
	Page      int
	PageSize  int
	Keyword   string
	Status    int32
	StartTime string
	EndTime   string
}

// UserDTO 表示管理后台用户列表和详情中的用户信息。
type UserDTO struct {
	ID             int64  `json:"id"`
	Phone          string `json:"phone"`
	Nickname       string `json:"nickname"`
	AvatarURL      string `json:"avatar_url"`
	Gender         int32  `json:"gender"`
	RealName       string `json:"real_name"`
	IDCardNo       string `json:"id_card_no"`
	RegisterSource string `json:"register_source"`
	Status         int32  `json:"status"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}

// DriverCertificationListRequest 表示司机审核列表查询条件。
type DriverCertificationListRequest struct {
	Page        int
	PageSize    int
	Keyword     string
	AuditStatus int32
	StartTime   string
	EndTime     string
}

// DriverCertificationDTO 表示司机认证审核记录。
type DriverCertificationDTO struct {
	ID                int64  `json:"id"`
	DriverID          int64  `json:"driver_id"`
	VehicleID         int64  `json:"vehicle_id"`
	DriverPhone       string `json:"driver_phone"`
	DriverName        string `json:"driver_name"`
	DriverStatus      int32  `json:"driver_status"`
	PlateNo           string `json:"plate_no"`
	VehicleStatus     int32  `json:"vehicle_status"`
	IDCardFrontURL    string `json:"id_card_front_url"`
	IDCardBackURL     string `json:"id_card_back_url"`
	DriverLicenseURL  string `json:"driver_license_url"`
	VehicleLicenseURL string `json:"vehicle_license_url"`
	AuditStatus       int32  `json:"audit_status"`
	AuditRemark       string `json:"audit_remark"`
	AuditedBy         int64  `json:"audited_by"`
	AuditedAt         string `json:"audited_at"`
	CreatedAt         string `json:"created_at"`
	UpdatedAt         string `json:"updated_at"`
}

// AuditRequest 表示司机审核通过或驳回请求。
type AuditRequest struct {
	Remark string `json:"remark"`
}

// OrderListRequest 表示订单列表查询条件。
type OrderListRequest struct {
	Page      int
	PageSize  int
	Keyword   string
	Status    int32
	UserID    int64
	DriverID  int64
	StartTime string
	EndTime   string
}

// OrderDTO 表示订单主信息。
type OrderDTO struct {
	ID                 int64  `json:"id"`
	OrderNo            string `json:"order_no"`
	UserID             int64  `json:"user_id"`
	DriverID           int64  `json:"driver_id"`
	CarType            int32  `json:"car_type"`
	FromAddress        string `json:"from_address"`
	FromLongitude      string `json:"from_longitude"`
	FromLatitude       string `json:"from_latitude"`
	ToAddress          string `json:"to_address"`
	ToLongitude        string `json:"to_longitude"`
	ToLatitude         string `json:"to_latitude"`
	EstimatedDistanceM int64  `json:"estimated_distance_m"`
	EstimatedDurationS int64  `json:"estimated_duration_s"`
	EstimatedPrice     string `json:"estimated_price"`
	Status             int32  `json:"status"`
	CancelReason       string `json:"cancel_reason"`
	CancelBy           string `json:"cancel_by"`
	CreatedAt          string `json:"created_at"`
	UpdatedAt          string `json:"updated_at"`
}

// OrderDetailDTO 表示订单详情聚合信息。
type OrderDetailDTO struct {
	Order           OrderDTO         `json:"order"`
	StatusLogs      []OrderStatusLog `json:"status_logs"`
	DispatchRecords []DispatchRecord `json:"dispatch_records"`
	Price           *OrderPrice      `json:"price,omitempty"`
	Payment         *Payment         `json:"payment,omitempty"`
	Settlement      *Settlement      `json:"settlement,omitempty"`
}

// OrderStatusLog 表示订单状态流转日志。
type OrderStatusLog struct {
	ID           int64  `json:"id"`
	OrderID      int64  `json:"order_id"`
	FromStatus   int32  `json:"from_status"`
	ToStatus     int32  `json:"to_status"`
	OperatorType string `json:"operator_type"`
	OperatorID   int64  `json:"operator_id"`
	Remark       string `json:"remark"`
	CreatedAt    string `json:"created_at"`
}

// DispatchRecord 表示派单记录。
type DispatchRecord struct {
	ID           int64  `json:"id"`
	OrderID      int64  `json:"order_id"`
	DriverID     int64  `json:"driver_id"`
	DispatchType int32  `json:"dispatch_type"`
	Status       int32  `json:"status"`
	MatchScore   string `json:"match_score"`
	Remark       string `json:"remark"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

// OrderPrice 表示订单价格明细。
type OrderPrice struct {
	ID              int64  `json:"id"`
	OrderID         int64  `json:"order_id"`
	PriceRuleID     int64  `json:"price_rule_id"`
	EstimatedPrice  string `json:"estimated_price"`
	ActualPrice     string `json:"actual_price"`
	BaseFee         string `json:"base_fee"`
	DistanceFee     string `json:"distance_fee"`
	TimeFee         string `json:"time_fee"`
	NightFee        string `json:"night_fee"`
	DynamicFee      string `json:"dynamic_fee"`
	DiscountAmount  string `json:"discount_amount"`
	PlatformSubsidy string `json:"platform_subsidy"`
	PayableAmount   string `json:"payable_amount"`
	Status          int32  `json:"status"`
}

// Payment 表示支付单信息。
type Payment struct {
	ID            int64  `json:"id"`
	PaymentNo     string `json:"payment_no"`
	OrderID       int64  `json:"order_id"`
	UserID        int64  `json:"user_id"`
	Amount        string `json:"amount"`
	Channel       string `json:"channel"`
	Status        int32  `json:"status"`
	TransactionID string `json:"transaction_id"`
	RefundAmount  string `json:"refund_amount"`
	PaidAt        string `json:"paid_at"`
}

// Settlement 表示结算单信息。
type Settlement struct {
	ID                     int64  `json:"id"`
	SettlementNo           string `json:"settlement_no"`
	OrderID                int64  `json:"order_id"`
	DriverID               int64  `json:"driver_id"`
	TotalAmount            string `json:"total_amount"`
	PlatformCommissionRate string `json:"platform_commission_rate"`
	PlatformCommission     string `json:"platform_commission"`
	DriverIncome           string `json:"driver_income"`
	Status                 int32  `json:"status"`
	SettledAt              string `json:"settled_at"`
}

// OperationLogListRequest 表示操作日志列表查询条件。
type OperationLogListRequest struct {
	Page       int
	PageSize   int
	AdminID    int64
	Module     string
	Action     string
	TargetType string
	TargetID   int64
	StartTime  string
	EndTime    string
}

// OperationLogDTO 表示后台操作日志信息。
type OperationLogDTO struct {
	ID         int64  `json:"id"`
	AdminID    int64  `json:"admin_id"`
	Module     string `json:"module"`
	Action     string `json:"action"`
	TargetType string `json:"target_type"`
	TargetID   int64  `json:"target_id"`
	Detail     string `json:"detail"`
	IP         string `json:"ip"`
	CreatedAt  string `json:"created_at"`
}
