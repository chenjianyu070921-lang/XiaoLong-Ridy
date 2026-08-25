package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"XiaoLong-Ridy/api/admin/internal/logic"
	"XiaoLong-Ridy/api/admin/internal/model"
	"XiaoLong-Ridy/api/admin/internal/svc"
	adminclient "XiaoLong-Ridy/rpc/adminsvc/client/adminservice"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// testResponse 表示管理后台统一响应结构，测试中只关心错误码和提示信息。
type testResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

// decodeTestResponse 解析 HTTP 测试响应，保证路由层始终输出统一 JSON。
func decodeTestResponse(t *testing.T, rec *httptest.ResponseRecorder) testResponse {
	t.Helper()
	var resp testResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v, body=%s", err, rec.Body.String())
	}
	return resp
}

// fakeAdminService 是价格规则接口测试专用的 adminsvc 假客户端。
// 它通过内嵌接口满足完整 AdminService 合约，仅实现本组测试会触发的方法。
type fakeAdminService struct {
	adminclient.AdminService

	validateReq *adminclient.ValidateSessionRequest
	registerReq *adminclient.RegisterRequest
	loginReq    *adminclient.LoginRequest
	logoutReq   *adminclient.LogoutRequest
	meReq       *adminclient.MeRequest
	menusReq    *adminclient.MenusRequest
	exportReq   *adminclient.ExportTaskDetailRequest

	listReq    *adminclient.PriceRuleListRequest
	detailReq  *adminclient.PriceRuleDetailRequest
	createReq  *adminclient.PriceRuleRequest
	updateReq  *adminclient.PriceRuleRequest
	enableReq  *adminclient.PriceRuleStatusRequest
	disableReq *adminclient.PriceRuleStatusRequest
}

// ValidateSession 记录鉴权请求，并返回固定管理员，证明 HTTP 层不再直接访问 Redis。
func (f *fakeAdminService) ValidateSession(ctx context.Context, in *adminclient.ValidateSessionRequest, opts ...grpc.CallOption) (*adminclient.ValidateSessionResponse, error) {
	f.validateReq = in
	return &adminclient.ValidateSessionResponse{
		Admin: &adminclient.Admin{Id: 7, Username: "root", RealName: "超级管理员", Role: 1, Status: 1},
	}, nil
}

// Register 记录注册请求，验证 HTTP 层只透传参数和当前操作者信息。
func (f *fakeAdminService) Register(ctx context.Context, in *adminclient.RegisterRequest, opts ...grpc.CallOption) (*adminclient.AuthResponse, error) {
	f.registerReq = in
	return &adminclient.AuthResponse{
		Token:     "new-token",
		ExpiresIn: 86400,
		Admin:     &adminclient.Admin{Id: 8, Username: in.GetUsername(), RealName: in.GetRealName(), Role: in.GetRole(), Status: 1},
	}, nil
}

// Login 记录登录请求，真正密码校验和会话生成由 adminsvc 负责。
func (f *fakeAdminService) Login(ctx context.Context, in *adminclient.LoginRequest, opts ...grpc.CallOption) (*adminclient.AuthResponse, error) {
	f.loginReq = in
	return &adminclient.AuthResponse{
		Token:     "login-token",
		ExpiresIn: 86400,
		Admin:     &adminclient.Admin{Id: 7, Username: in.GetUsername(), RealName: "超级管理员", Role: 1, Status: 1},
	}, nil
}

// Logout 记录退出请求，证明 token 删除下沉到 adminsvc。
func (f *fakeAdminService) Logout(ctx context.Context, in *adminclient.LogoutRequest, opts ...grpc.CallOption) (*adminclient.CommonResponse, error) {
	f.logoutReq = in
	return &adminclient.CommonResponse{Message: "ok"}, nil
}

// Me 记录当前管理员查询请求，验证 token 由 HTTP 层透传到 adminsvc。
func (f *fakeAdminService) Me(ctx context.Context, in *adminclient.MeRequest, opts ...grpc.CallOption) (*adminclient.MeResponse, error) {
	f.meReq = in
	return &adminclient.MeResponse{Admin: &adminclient.Admin{Id: 7, Username: "root", RealName: "超级管理员", Role: 1, Status: 1}}, nil
}

// Menus 记录菜单查询请求，菜单规则由 adminsvc 根据 token 对应角色决定。
func (f *fakeAdminService) Menus(ctx context.Context, in *adminclient.MenusRequest, opts ...grpc.CallOption) (*adminclient.MenusResponse, error) {
	f.menusReq = in
	return &adminclient.MenusResponse{
		Items: []*adminclient.MenuItem{{Name: "工作台", Path: "/dashboard", Icon: "Home", Perm: "dashboard:view"}},
	}, nil
}

// GetExportTask 记录导出任务详情请求，验证 HTTP 详情路由透传 task_no 到 adminsvc。
func (f *fakeAdminService) GetExportTask(ctx context.Context, in *adminclient.ExportTaskDetailRequest, opts ...grpc.CallOption) (*adminclient.ExportTask, error) {
	f.exportReq = in
	return &adminclient.ExportTask{
		TaskNo:        in.GetTaskNo(),
		ExportType:    "orders",
		Filters:       `{"status":5}`,
		Status:        "success",
		FilePath:      ".tmp-admin-exports/" + in.GetTaskNo() + ".json",
		FailureReason: "",
		UpdatedAt:     "2026-08-20 12:00:01",
	}, nil
}

// ListPriceRules 记录列表请求，并返回一条固定计价规则用于验证响应结构。
func (f *fakeAdminService) ListPriceRules(ctx context.Context, in *adminclient.PriceRuleListRequest, opts ...grpc.CallOption) (*adminclient.PriceRuleListResponse, error) {
	f.listReq = in
	return &adminclient.PriceRuleListResponse{
		List: []*adminclient.PriceRule{{
			Id:               9,
			Name:             "标准快车",
			CityCode:         "110100",
			CarType:          1,
			BasePrice:        "12.50",
			BaseDistanceKm:   "3.00",
			PerKmPrice:       "2.40",
			PerMinutePrice:   "0.50",
			NightStartTime:   "22:00:00",
			NightEndTime:     "06:00:00",
			NightSurcharge:   "1.20",
			DynamicMaxFactor: "2.00",
			Status:           1,
			EffectiveAt:      "2026-08-20 00:00:00",
			CreatedAt:        "2026-08-20 00:00:00",
			UpdatedAt:        "2026-08-20 00:00:00",
		}},
		Total:    1,
		Page:     in.GetPage(),
		PageSize: in.GetPageSize(),
	}, nil
}

// GetPriceRule 记录详情请求，并返回固定计价规则。
func (f *fakeAdminService) GetPriceRule(ctx context.Context, in *adminclient.PriceRuleDetailRequest, opts ...grpc.CallOption) (*adminclient.PriceRule, error) {
	f.detailReq = in
	return &adminclient.PriceRule{Id: in.GetId(), Name: "标准快车", Status: 1}, nil
}

// CreatePriceRule 记录新增请求，验证 HTTP 层已补充管理员和 IP 信息。
func (f *fakeAdminService) CreatePriceRule(ctx context.Context, in *adminclient.PriceRuleRequest, opts ...grpc.CallOption) (*adminclient.CreatePriceRuleResponse, error) {
	f.createReq = in
	return &adminclient.CreatePriceRuleResponse{Id: 1, Message: "ok"}, nil
}

// UpdatePriceRule 记录编辑请求，验证路径 ID 会覆盖请求体 ID。
func (f *fakeAdminService) UpdatePriceRule(ctx context.Context, in *adminclient.PriceRuleRequest, opts ...grpc.CallOption) (*adminclient.CommonResponse, error) {
	f.updateReq = in
	return &adminclient.CommonResponse{Message: "ok"}, nil
}

// EnablePriceRule 记录启用请求，验证状态操作走 adminsvc 而非 HTTP 直连数据库。
func (f *fakeAdminService) EnablePriceRule(ctx context.Context, in *adminclient.PriceRuleStatusRequest, opts ...grpc.CallOption) (*adminclient.CommonResponse, error) {
	f.enableReq = in
	return &adminclient.CommonResponse{Message: "ok"}, nil
}

// DisablePriceRule 记录停用请求，验证状态操作走 adminsvc 而非 HTTP 直连数据库。
func (f *fakeAdminService) DisablePriceRule(ctx context.Context, in *adminclient.PriceRuleStatusRequest, opts ...grpc.CallOption) (*adminclient.CommonResponse, error) {
	f.disableReq = in
	return &adminclient.CommonResponse{Message: "ok"}, nil
}

// ListOperationLogs 返回操作日志列表，用于全量路由冒烟测试。
func (f *fakeAdminService) ListOperationLogs(ctx context.Context, in *adminclient.OperationLogListRequest, opts ...grpc.CallOption) (*adminclient.OperationLogListResponse, error) {
	return &adminclient.OperationLogListResponse{List: []*adminclient.OperationLog{{Id: 1, AdminId: 7, Module: "user", Action: "list", TargetType: "user", TargetId: 1, Detail: "ok", Ip: "127.0.0.1", CreatedAt: "2026-08-20 12:00:00"}}, Total: 1, Page: in.GetPage(), PageSize: in.GetPageSize()}, nil
}

// ListUsers 返回用户列表，用于验证用户列表路由。
func (f *fakeAdminService) ListUsers(ctx context.Context, in *adminclient.UserListRequest, opts ...grpc.CallOption) (*adminclient.UserListResponse, error) {
	return &adminclient.UserListResponse{List: []*adminclient.User{{Id: 1, Phone: "13800000000", Nickname: "用户", Status: 1, CreatedAt: "2026-08-20 12:00:00"}}, Total: 1, Page: in.GetPage(), PageSize: in.GetPageSize()}, nil
}

// GetUser 返回用户详情，用于验证用户详情路由。
func (f *fakeAdminService) GetUser(ctx context.Context, in *adminclient.UserDetailRequest, opts ...grpc.CallOption) (*adminclient.User, error) {
	return &adminclient.User{Id: in.GetId(), Phone: "13800000000", Nickname: "用户", Status: 1}, nil
}

// FreezeUser 返回冻结成功结果，用于验证用户状态操作路由。
func (f *fakeAdminService) FreezeUser(ctx context.Context, in *adminclient.ChangeUserStatusRequest, opts ...grpc.CallOption) (*adminclient.CommonResponse, error) {
	return &adminclient.CommonResponse{Message: "ok"}, nil
}

// UnfreezeUser 返回解封成功结果，用于验证用户状态操作路由。
func (f *fakeAdminService) UnfreezeUser(ctx context.Context, in *adminclient.ChangeUserStatusRequest, opts ...grpc.CallOption) (*adminclient.CommonResponse, error) {
	return &adminclient.CommonResponse{Message: "ok"}, nil
}

// ListDriverCertifications 返回司机审核列表，用于验证司机审核列表路由。
func (f *fakeAdminService) ListDriverCertifications(ctx context.Context, in *adminclient.DriverCertificationListRequest, opts ...grpc.CallOption) (*adminclient.DriverCertificationListResponse, error) {
	return &adminclient.DriverCertificationListResponse{List: []*adminclient.DriverCertification{{Id: 1, DriverId: 10, DriverPhone: "13900000000", DriverName: "司机", AuditStatus: 1}}, Total: 1, Page: in.GetPage(), PageSize: in.GetPageSize()}, nil
}

// GetDriverCertification 返回司机审核详情，用于验证详情路由。
func (f *fakeAdminService) GetDriverCertification(ctx context.Context, in *adminclient.DriverCertificationDetailRequest, opts ...grpc.CallOption) (*adminclient.DriverCertification, error) {
	return &adminclient.DriverCertification{Id: in.GetId(), DriverId: 10, DriverName: "司机", AuditStatus: 1}, nil
}

// ApproveDriverCertification 返回审核通过结果，用于验证司机审核通过路由。
func (f *fakeAdminService) ApproveDriverCertification(ctx context.Context, in *adminclient.AuditDriverCertificationRequest, opts ...grpc.CallOption) (*adminclient.CommonResponse, error) {
	return &adminclient.CommonResponse{Message: "ok"}, nil
}

// RejectDriverCertification 返回审核驳回结果，用于验证司机审核驳回路由。
func (f *fakeAdminService) RejectDriverCertification(ctx context.Context, in *adminclient.AuditDriverCertificationRequest, opts ...grpc.CallOption) (*adminclient.CommonResponse, error) {
	return &adminclient.CommonResponse{Message: "ok"}, nil
}

// ListOrders 返回订单列表，用于验证订单列表路由。
func (f *fakeAdminService) ListOrders(ctx context.Context, in *adminclient.OrderListRequest, opts ...grpc.CallOption) (*adminclient.OrderListResponse, error) {
	return &adminclient.OrderListResponse{List: []*adminclient.Order{{Id: 1, OrderNo: "ORD001", UserId: 1001, DriverId: 2002, Status: 4, EstimatedPrice: "52.00"}}, Total: 1, Page: in.GetPage(), PageSize: in.GetPageSize()}, nil
}

// GetOrder 返回订单详情，用于验证订单详情聚合路由。
func (f *fakeAdminService) GetOrder(ctx context.Context, in *adminclient.OrderDetailRequest, opts ...grpc.CallOption) (*adminclient.OrderDetail, error) {
	return &adminclient.OrderDetail{Order: &adminclient.Order{Id: in.GetId(), OrderNo: "ORD001", UserId: 1001, DriverId: 2002, Status: 4}}, nil
}

// CancelOrder 返回后台取消订单成功结果，用于验证取消路由。
func (f *fakeAdminService) CancelOrder(ctx context.Context, in *adminclient.AdminCancelOrderRequest, opts ...grpc.CallOption) (*adminclient.CommonResponse, error) {
	return &adminclient.CommonResponse{Message: "ok"}, nil
}

// ListAbnormalOrders 返回异常订单列表，用于验证异常订单路由。
func (f *fakeAdminService) ListAbnormalOrders(ctx context.Context, in *adminclient.AbnormalOrderListRequest, opts ...grpc.CallOption) (*adminclient.AbnormalOrderListResponse, error) {
	return &adminclient.AbnormalOrderListResponse{List: []*adminclient.AbnormalOrder{{Id: 1, OrderNo: "ORD001", UserId: 1001, AbnormalType: "payment", AbnormalReason: "支付异常"}}, Total: 1, Page: in.GetPage(), PageSize: in.GetPageSize()}, nil
}

// ListCoupons 返回优惠券列表，用于验证列表和创建后回查路由。
func (f *fakeAdminService) ListCoupons(ctx context.Context, in *adminclient.CouponListRequest, opts ...grpc.CallOption) (*adminclient.CouponListResponse, error) {
	return &adminclient.CouponListResponse{List: []*adminclient.Coupon{{Id: 1, Name: "新人券", Type: 1, FaceValue: "10.00", ThresholdAmount: "30.00", TotalCount: 100, Status: 1}}, Total: 1, Page: in.GetPage(), PageSize: in.GetPageSize()}, nil
}

// CreateCoupon 返回优惠券创建成功结果。
func (f *fakeAdminService) CreateCoupon(ctx context.Context, in *adminclient.CouponRequest, opts ...grpc.CallOption) (*adminclient.CreateCouponResponse, error) {
	return &adminclient.CreateCouponResponse{Id: 1, Message: "ok"}, nil
}

// UpdateCoupon 返回优惠券编辑成功结果。
func (f *fakeAdminService) UpdateCoupon(ctx context.Context, in *adminclient.CouponRequest, opts ...grpc.CallOption) (*adminclient.CommonResponse, error) {
	return &adminclient.CommonResponse{Message: "ok"}, nil
}

// DisableCoupon 返回优惠券下架成功结果。
func (f *fakeAdminService) DisableCoupon(ctx context.Context, in *adminclient.CouponRequest, opts ...grpc.CallOption) (*adminclient.CommonResponse, error) {
	return &adminclient.CommonResponse{Message: "ok"}, nil
}

// IssueCoupon 返回发券任务创建结果。
func (f *fakeAdminService) IssueCoupon(ctx context.Context, in *adminclient.CouponIssueRequest, opts ...grpc.CallOption) (*adminclient.CouponIssueResponse, error) {
	return &adminclient.CouponIssueResponse{TaskNo: "CI202608200001", TotalCount: 1, SuccessCount: 1, Status: "success"}, nil
}

// ListCouponIssueTasks 返回发券任务列表。
func (f *fakeAdminService) ListCouponIssueTasks(ctx context.Context, in *adminclient.CouponIssueTaskListRequest, opts ...grpc.CallOption) (*adminclient.CouponIssueTaskListResponse, error) {
	return &adminclient.CouponIssueTaskListResponse{List: []*adminclient.CouponIssueTask{{Id: 1, TaskNo: "CI202608200001", CouponId: 1, Status: "success"}}, Total: 1, Page: in.GetPage(), PageSize: in.GetPageSize()}, nil
}

// ListPromotionActivities 返回活动列表。
func (f *fakeAdminService) ListPromotionActivities(ctx context.Context, in *adminclient.PromotionActivityListRequest, opts ...grpc.CallOption) (*adminclient.PromotionActivityListResponse, error) {
	return &adminclient.PromotionActivityListResponse{List: []*adminclient.PromotionActivity{{Id: 1, Name: "暑期活动", Type: 1, Config: "{}", Status: 1}}, Total: 1, Page: in.GetPage(), PageSize: in.GetPageSize()}, nil
}

// CreatePromotionActivity 返回活动创建成功结果。
func (f *fakeAdminService) CreatePromotionActivity(ctx context.Context, in *adminclient.PromotionActivityRequest, opts ...grpc.CallOption) (*adminclient.CommonResponse, error) {
	return &adminclient.CommonResponse{Message: "ok"}, nil
}

// UpdatePromotionActivity 返回活动编辑成功结果。
func (f *fakeAdminService) UpdatePromotionActivity(ctx context.Context, in *adminclient.PromotionActivityRequest, opts ...grpc.CallOption) (*adminclient.CommonResponse, error) {
	return &adminclient.CommonResponse{Message: "ok"}, nil
}

// PublishPromotionActivity 返回活动发布成功结果。
func (f *fakeAdminService) PublishPromotionActivity(ctx context.Context, in *adminclient.PromotionActivityActionRequest, opts ...grpc.CallOption) (*adminclient.CommonResponse, error) {
	return &adminclient.CommonResponse{Message: "ok"}, nil
}

// RollbackPromotionActivity 返回活动回滚成功结果。
func (f *fakeAdminService) RollbackPromotionActivity(ctx context.Context, in *adminclient.PromotionActivityActionRequest, opts ...grpc.CallOption) (*adminclient.CommonResponse, error) {
	return &adminclient.CommonResponse{Message: "ok"}, nil
}

// GetStatisticsOverview 返回运营总览统计。
func (f *fakeAdminService) GetStatisticsOverview(ctx context.Context, in *adminclient.StatisticsRequest, opts ...grpc.CallOption) (*adminclient.StatisticsOverviewResponse, error) {
	return &adminclient.StatisticsOverviewResponse{UserCount: 10, DriverCount: 2, OrderCount: 3, Gmv: "100.00"}, nil
}

// GetOrderStatistics 返回订单统计。
func (f *fakeAdminService) GetOrderStatistics(ctx context.Context, in *adminclient.StatisticsRequest, opts ...grpc.CallOption) (*adminclient.OrderStatisticsResponse, error) {
	return &adminclient.OrderStatisticsResponse{OrderCount: 3, CompletedOrderCount: 2, CompletionRate: "66.67%"}, nil
}

// GetDriverStatistics 返回司机经营统计。
func (f *fakeAdminService) GetDriverStatistics(ctx context.Context, in *adminclient.StatisticsRequest, opts ...grpc.CallOption) (*adminclient.DriverStatisticsResponse, error) {
	return &adminclient.DriverStatisticsResponse{DriverTotal: 2, NewDriverCount: 1, CompletedOrderCount: 8, DriverIncome: "80.00"}, nil
}

// GetFinanceStatistics 返回财务收入统计。
func (f *fakeAdminService) GetFinanceStatistics(ctx context.Context, in *adminclient.StatisticsRequest, opts ...grpc.CallOption) (*adminclient.FinanceStatisticsResponse, error) {
	return &adminclient.FinanceStatisticsResponse{PaymentOrderCount: 3, PaidAmount: "100.00", DriverIncome: "80.00"}, nil
}

// GetCouponStatistics 返回优惠券统计。
func (f *fakeAdminService) GetCouponStatistics(ctx context.Context, in *adminclient.StatisticsRequest, opts ...grpc.CallOption) (*adminclient.CouponStatisticsResponse, error) {
	return &adminclient.CouponStatisticsResponse{CouponCount: 2, IssuedCouponCount: 10, UseRate: "20.00%"}, nil
}

// CreateExportTask 返回导出任务创建结果。
func (f *fakeAdminService) CreateExportTask(ctx context.Context, in *adminclient.ExportTaskRequest, opts ...grpc.CallOption) (*adminclient.ExportTaskResponse, error) {
	return &adminclient.ExportTaskResponse{TaskNo: "EX20260820120000000001", Status: "pending", Message: "ok"}, nil
}

// ListExportTasks 返回导出任务列表。
func (f *fakeAdminService) ListExportTasks(ctx context.Context, in *adminclient.ExportTaskListRequest, opts ...grpc.CallOption) (*adminclient.ExportTaskListResponse, error) {
	return &adminclient.ExportTaskListResponse{List: []*adminclient.ExportTask{{TaskNo: "EX20260820120000000001", ExportType: "orders", Status: "success"}}, Total: 1, Page: in.GetPage(), PageSize: in.GetPageSize()}, nil
}

// ListBlacklists 返回黑名单列表。
func (f *fakeAdminService) ListBlacklists(ctx context.Context, in *adminclient.BlacklistListRequest, opts ...grpc.CallOption) (*adminclient.BlacklistListResponse, error) {
	return &adminclient.BlacklistListResponse{List: []*adminclient.Blacklist{{Id: 1, TargetType: "user", TargetId: 1001, Reason: "risk", Status: 1}}, Total: 1, Page: in.GetPage(), PageSize: in.GetPageSize()}, nil
}

// AddBlacklist 返回新增黑名单成功结果。
func (f *fakeAdminService) AddBlacklist(ctx context.Context, in *adminclient.BlacklistRequest, opts ...grpc.CallOption) (*adminclient.CommonResponse, error) {
	return &adminclient.CommonResponse{Message: "ok"}, nil
}

// ReleaseBlacklist 返回解除黑名单成功结果。
func (f *fakeAdminService) ReleaseBlacklist(ctx context.Context, in *adminclient.BlacklistRequest, opts ...grpc.CallOption) (*adminclient.CommonResponse, error) {
	return &adminclient.CommonResponse{Message: "ok"}, nil
}

// ListRiskHitRecords 返回风控命中记录列表。
func (f *fakeAdminService) ListRiskHitRecords(ctx context.Context, in *adminclient.RiskHitRecordListRequest, opts ...grpc.CallOption) (*adminclient.RiskHitRecordListResponse, error) {
	return &adminclient.RiskHitRecordListResponse{List: []*adminclient.RiskHitRecord{{Id: 1, TargetType: "user", TargetId: 1001, Scene: "login", RiskLevel: 2, HitReason: "命中黑名单"}}, Total: 1, Page: in.GetPage(), PageSize: in.GetPageSize()}, nil
}

// CreateWorkOrder 返回工单创建结果。
func (f *fakeAdminService) CreateWorkOrder(ctx context.Context, in *adminclient.WorkOrderRequest, opts ...grpc.CallOption) (*adminclient.WorkOrder, error) {
	return &adminclient.WorkOrder{Id: 1, WorkOrderNo: "WO202608240001", Title: in.GetTitle(), Status: 1, Version: 1}, nil
}

// ListWorkOrders 返回工单列表。
func (f *fakeAdminService) ListWorkOrders(ctx context.Context, in *adminclient.WorkOrderListRequest, opts ...grpc.CallOption) (*adminclient.WorkOrderListResponse, error) {
	return &adminclient.WorkOrderListResponse{List: []*adminclient.WorkOrder{{Id: 1, WorkOrderNo: "WO202608240001", Title: "测试工单", Status: 1, Version: 1}}, Total: 1, Page: in.GetPage(), PageSize: in.GetPageSize()}, nil
}

// GetWorkOrder 返回工单详情。
func (f *fakeAdminService) GetWorkOrder(ctx context.Context, in *adminclient.WorkOrderDetailRequest, opts ...grpc.CallOption) (*adminclient.WorkOrder, error) {
	return &adminclient.WorkOrder{Id: in.GetId(), WorkOrderNo: "WO202608240001", Title: "测试工单", Status: 1, Version: 1}, nil
}

// ActWorkOrder 返回工单流转结果。
func (f *fakeAdminService) ActWorkOrder(ctx context.Context, in *adminclient.WorkOrderActionRequest, opts ...grpc.CallOption) (*adminclient.WorkOrder, error) {
	return &adminclient.WorkOrder{Id: in.GetId(), WorkOrderNo: "WO202608240001", Status: 2, Version: in.GetVersion() + 1}, nil
}

// BatchActWorkOrders 返回批量工单处理结果。
func (f *fakeAdminService) BatchActWorkOrders(ctx context.Context, in *adminclient.WorkOrderBatchActionRequest, opts ...grpc.CallOption) (*adminclient.WorkOrderBatchActionResponse, error) {
	return &adminclient.WorkOrderBatchActionResponse{SuccessCount: int64(len(in.GetIds()))}, nil
}

// AddWorkOrderEvidence 返回新增工单证据。
func (f *fakeAdminService) AddWorkOrderEvidence(ctx context.Context, in *adminclient.WorkOrderEvidenceRequest, opts ...grpc.CallOption) (*adminclient.WorkOrderEvidence, error) {
	return &adminclient.WorkOrderEvidence{Id: 1, WorkOrderId: in.GetWorkOrderId(), EvidenceType: in.GetEvidenceType(), Content: in.GetContent()}, nil
}

// ListWorkOrderEvidence 返回工单证据列表。
func (f *fakeAdminService) ListWorkOrderEvidence(ctx context.Context, in *adminclient.WorkOrderEvidenceListRequest, opts ...grpc.CallOption) (*adminclient.WorkOrderEvidenceListResponse, error) {
	return &adminclient.WorkOrderEvidenceListResponse{List: []*adminclient.WorkOrderEvidence{{Id: 1, WorkOrderId: in.GetWorkOrderId(), EvidenceType: "text"}}, Total: 1, Page: in.GetPage(), PageSize: in.GetPageSize()}, nil
}

// HandleRiskHitRecords 返回风控命中处置结果。
func (f *fakeAdminService) HandleRiskHitRecords(ctx context.Context, in *adminclient.RiskHitActionRequest, opts ...grpc.CallOption) (*adminclient.RiskHitActionResponse, error) {
	return &adminclient.RiskHitActionResponse{SuccessCount: int64(len(in.GetIds())), WorkOrderIds: []int64{1}}, nil
}

// newPriceRuleTestRouter 创建绕过鉴权中间件的 Router，用于直接测试价格规则 handler。
func newPriceRuleTestRouter(adminSvc adminclient.AdminService) *Router {
	return &Router{ctx: &svc.ServiceContext{AdminSvc: adminSvc}}
}

// newSessionRequest 为直接调用 handler 的请求注入管理员会话。
func newSessionRequest(method, target, body string) *http.Request {
	var reader *strings.Reader
	if body == "" {
		reader = strings.NewReader("")
	} else {
		reader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, target, reader)
	req.RemoteAddr = "127.0.0.1:12345"
	ctx := context.WithValue(req.Context(), sessionContextKey{}, &model.AdminSession{AdminID: 7})
	return req.WithContext(ctx)
}

// validPriceRuleBody 返回符合接口约束的计价规则保存请求体。
func validPriceRuleBody() string {
	return `{"name":"标准快车","city_code":"110100","car_type":1,"base_price":"12.50","base_distance_km":"3.00","per_km_price":"2.40","per_minute_price":"0.50","night_start_time":"22:00:00","night_end_time":"06:00:00","night_surcharge":"1.20","dynamic_max_factor":"2.00","status":1,"effective_at":"2026-08-20 00:00:00","expire_at":"2026-12-31 23:59:59"}`
}

// TestRouter_AuthRoutesUseAdminSvc 验证登录、注册、当前用户、菜单、退出均只调用 adminsvc。
func TestRouter_AuthRoutesUseAdminSvc(t *testing.T) {
	fakeSvc := &fakeAdminService{}
	router := NewRouter(&svc.ServiceContext{AdminSvc: fakeSvc})

	cases := []struct {
		name   string
		method string
		path   string
		body   string
		token  string
		assert func(t *testing.T)
	}{
		{name: "login", method: http.MethodPost, path: "/admin/v1/auth/login", body: `{"username":"root","password":"secret"}`, assert: func(t *testing.T) {
			if fakeSvc.loginReq == nil || fakeSvc.loginReq.GetUsername() != "root" || fakeSvc.loginReq.GetPassword() != "secret" {
				t.Fatalf("login request not mapped correctly: %+v", fakeSvc.loginReq)
			}
		}},
		{name: "register with token", method: http.MethodPost, path: "/admin/v1/auth/register", token: "root-token", body: `{"username":"ops","password":"secret","real_name":"运营","role":2}`, assert: func(t *testing.T) {
			if fakeSvc.validateReq == nil || fakeSvc.validateReq.GetToken() != "root-token" {
				t.Fatalf("validate request not mapped correctly: %+v", fakeSvc.validateReq)
			}
			if fakeSvc.registerReq == nil || fakeSvc.registerReq.GetUsername() != "ops" || fakeSvc.registerReq.GetOperatorAdminId() != 7 || fakeSvc.registerReq.GetOperatorRole() != 1 {
				t.Fatalf("register request not mapped correctly: %+v", fakeSvc.registerReq)
			}
		}},
		{name: "me", method: http.MethodGet, path: "/admin/v1/auth/me", token: "root-token", assert: func(t *testing.T) {
			if fakeSvc.meReq == nil || fakeSvc.meReq.GetToken() != "root-token" {
				t.Fatalf("me request not mapped correctly: %+v", fakeSvc.meReq)
			}
		}},
		{name: "menus", method: http.MethodGet, path: "/admin/v1/menus", token: "root-token", assert: func(t *testing.T) {
			if fakeSvc.menusReq == nil || fakeSvc.menusReq.GetToken() != "root-token" {
				t.Fatalf("menus request not mapped correctly: %+v", fakeSvc.menusReq)
			}
		}},
		{name: "logout", method: http.MethodPost, path: "/admin/v1/auth/logout", token: "root-token", assert: func(t *testing.T) {
			if fakeSvc.logoutReq == nil || fakeSvc.logoutReq.GetToken() != "root-token" {
				t.Fatalf("logout request not mapped correctly: %+v", fakeSvc.logoutReq)
			}
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			if tc.token != "" {
				req.Header.Set("Authorization", "Bearer "+tc.token)
			}
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			resp := decodeTestResponse(t, rec)
			if rec.Code != http.StatusOK || resp.Code != 0 {
				t.Fatalf("status/code = %d/%d, want %d/%d, body=%s", rec.Code, resp.Code, http.StatusOK, 0, rec.Body.String())
			}
			tc.assert(t)
		})
	}
}

// TestRouter_ExportTaskDetailRoute 验证导出任务详情路由会走 adminsvc 并返回状态机字段。
func TestRouter_ExportTaskDetailRoute(t *testing.T) {
	fakeSvc := &fakeAdminService{}
	router := NewRouter(&svc.ServiceContext{AdminSvc: fakeSvc})
	req := httptest.NewRequest(http.MethodGet, "/admin/v1/export-tasks/EX20260820120000000001", nil)
	req.Header.Set("Authorization", "Bearer root-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	resp := decodeTestResponse(t, rec)
	if rec.Code != http.StatusOK || resp.Code != 0 {
		t.Fatalf("status/code = %d/%d, want %d/%d, body=%s", rec.Code, resp.Code, http.StatusOK, 0, rec.Body.String())
	}
	if fakeSvc.validateReq == nil || fakeSvc.exportReq == nil || fakeSvc.exportReq.GetTaskNo() != "EX20260820120000000001" {
		t.Fatalf("export detail request not mapped correctly, validate=%+v detail=%+v", fakeSvc.validateReq, fakeSvc.exportReq)
	}
	if !bytes.Contains(resp.Data, []byte(`"file_path"`)) || !bytes.Contains(resp.Data, []byte(`"updated_at"`)) {
		t.Fatalf("response missing export state fields: %s", string(resp.Data))
	}
}

// TestRouter_AuthRequiredMissingToken 验证后台业务接口缺少 token 时返回统一未登录错误。
func TestRouter_AuthRequiredMissingToken(t *testing.T) {
	router := NewRouter(&svc.ServiceContext{})
	req := httptest.NewRequest(http.MethodGet, "/admin/v1/users", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	resp := decodeTestResponse(t, rec)
	if rec.Code != http.StatusUnauthorized || resp.Code != 40004 {
		t.Fatalf("status/code = %d/%d, want %d/%d", rec.Code, resp.Code, http.StatusUnauthorized, 40004)
	}
}

// TestRouter_AuthRequiredMissingTokenPriceRule 验证新增的计价规则路由同样经过统一鉴权中间件。
func TestRouter_AuthRequiredMissingTokenPriceRule(t *testing.T) {
	router := NewRouter(&svc.ServiceContext{})
	req := httptest.NewRequest(http.MethodGet, "/admin/v1/price-rules", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	resp := decodeTestResponse(t, rec)
	if rec.Code != http.StatusUnauthorized || resp.Code != 40004 {
		t.Fatalf("status/code = %d/%d, want %d/%d", rec.Code, resp.Code, http.StatusUnauthorized, 40004)
	}
}

// TestRouter_MethodNotAllowed 验证已注册接口使用错误 HTTP 方法时返回参数错误码。
func TestRouter_MethodNotAllowed(t *testing.T) {
	router := NewRouter(&svc.ServiceContext{})
	req := httptest.NewRequest(http.MethodGet, "/admin/v1/auth/login", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	resp := decodeTestResponse(t, rec)
	if rec.Code != http.StatusMethodNotAllowed || resp.Code != 40001 {
		t.Fatalf("status/code = %d/%d, want %d/%d", rec.Code, resp.Code, http.StatusMethodNotAllowed, 40001)
	}
}

// TestRouter_WriteBizErrorMapping 验证业务层和 RPC 层错误会映射为管理后台统一错误码。
func TestRouter_WriteBizErrorMapping(t *testing.T) {
	cases := []struct {
		name       string
		err        error
		httpStatus int
		code       int
	}{
		{name: "bad request", err: logic.ErrBadRequest, httpStatus: http.StatusBadRequest, code: 40001},
		{name: "grpc invalid argument", err: status.Error(codes.InvalidArgument, "bad"), httpStatus: http.StatusBadRequest, code: 40001},
		{name: "grpc not found", err: status.Error(codes.NotFound, "missing"), httpStatus: http.StatusNotFound, code: 40401},
		{name: "grpc permission denied", err: status.Error(codes.PermissionDenied, "forbidden"), httpStatus: http.StatusForbidden, code: 40003},
		{name: "grpc unauthenticated", err: status.Error(codes.Unauthenticated, "login"), httpStatus: http.StatusUnauthorized, code: 40004},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			(&Router{}).writeBizError(rec, tc.err)

			resp := decodeTestResponse(t, rec)
			if rec.Code != tc.httpStatus || resp.Code != tc.code {
				t.Fatalf("status/code = %d/%d, want %d/%d", rec.Code, resp.Code, tc.httpStatus, tc.code)
			}
		})
	}
}

// TestRouter_PriceRuleListRoute 验证价格规则列表接口会正确解析查询参数并调用 adminsvc。
func TestRouter_PriceRuleListRoute(t *testing.T) {
	fakeSvc := &fakeAdminService{}
	router := newPriceRuleTestRouter(fakeSvc)
	req := newSessionRequest(http.MethodGet, "/admin/v1/price-rules?page=2&page_size=30&keyword=快车&city_code=110100&car_type=1&status=1", "")
	rec := httptest.NewRecorder()

	router.handlePriceRules(rec, req)

	resp := decodeTestResponse(t, rec)
	if rec.Code != http.StatusOK || resp.Code != 0 {
		t.Fatalf("status/code = %d/%d, want %d/%d", rec.Code, resp.Code, http.StatusOK, 0)
	}
	if fakeSvc.listReq == nil || fakeSvc.listReq.GetPage() != 2 || fakeSvc.listReq.GetPageSize() != 30 || fakeSvc.listReq.GetKeyword() != "快车" || fakeSvc.listReq.GetCityCode() != "110100" || fakeSvc.listReq.GetCarType() != 1 || fakeSvc.listReq.GetStatus() != 1 {
		t.Fatalf("list request not mapped correctly: %+v", fakeSvc.listReq)
	}
	if !bytes.Contains(resp.Data, []byte(`"total":1`)) || !bytes.Contains(resp.Data, []byte(`"标准快车"`)) {
		t.Fatalf("unexpected response data: %s", string(resp.Data))
	}
}

// TestRouter_PriceRuleCreateRoute 验证新增计价规则接口会补齐管理员、客户端 IP 和请求体字段。
func TestRouter_PriceRuleCreateRoute(t *testing.T) {
	fakeSvc := &fakeAdminService{}
	router := newPriceRuleTestRouter(fakeSvc)
	req := newSessionRequest(http.MethodPost, "/admin/v1/price-rules", validPriceRuleBody())
	rec := httptest.NewRecorder()

	router.handlePriceRules(rec, req)

	resp := decodeTestResponse(t, rec)
	if rec.Code != http.StatusOK || resp.Code != 0 {
		t.Fatalf("status/code = %d/%d, want %d/%d", rec.Code, resp.Code, http.StatusOK, 0)
	}
	if fakeSvc.createReq == nil || fakeSvc.createReq.GetAdminId() != 7 || fakeSvc.createReq.GetIp() != "127.0.0.1" || fakeSvc.createReq.GetName() != "标准快车" || fakeSvc.createReq.GetStatus() != 1 {
		t.Fatalf("create request not mapped correctly: %+v", fakeSvc.createReq)
	}
}

// TestRouter_PriceRuleDetailUpdateAndStatusRoutes 验证详情、编辑、启用、停用四个带 ID 的计价规则接口。
func TestRouter_PriceRuleDetailUpdateAndStatusRoutes(t *testing.T) {
	fakeSvc := &fakeAdminService{}
	router := newPriceRuleTestRouter(fakeSvc)

	cases := []struct {
		name   string
		method string
		path   string
		body   string
		assert func(t *testing.T)
	}{
		{name: "detail", method: http.MethodGet, path: "/admin/v1/price-rules/9", assert: func(t *testing.T) {
			if fakeSvc.detailReq == nil || fakeSvc.detailReq.GetId() != 9 {
				t.Fatalf("detail request not mapped correctly: %+v", fakeSvc.detailReq)
			}
		}},
		{name: "update", method: http.MethodPut, path: "/admin/v1/price-rules/9", body: validPriceRuleBody(), assert: func(t *testing.T) {
			if fakeSvc.updateReq == nil || fakeSvc.updateReq.GetId() != 9 || fakeSvc.updateReq.GetAdminId() != 7 || fakeSvc.updateReq.GetName() != "标准快车" {
				t.Fatalf("update request not mapped correctly: %+v", fakeSvc.updateReq)
			}
		}},
		{name: "enable", method: http.MethodPost, path: "/admin/v1/price-rules/9/enable", assert: func(t *testing.T) {
			if fakeSvc.enableReq == nil || fakeSvc.enableReq.GetId() != 9 || fakeSvc.enableReq.GetAdminId() != 7 || fakeSvc.enableReq.GetStatus() != 1 {
				t.Fatalf("enable request not mapped correctly: %+v", fakeSvc.enableReq)
			}
		}},
		{name: "disable", method: http.MethodPost, path: "/admin/v1/price-rules/9/disable", assert: func(t *testing.T) {
			if fakeSvc.disableReq == nil || fakeSvc.disableReq.GetId() != 9 || fakeSvc.disableReq.GetAdminId() != 7 || fakeSvc.disableReq.GetStatus() != 2 {
				t.Fatalf("disable request not mapped correctly: %+v", fakeSvc.disableReq)
			}
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := newSessionRequest(tc.method, tc.path, tc.body)
			rec := httptest.NewRecorder()

			router.handlePriceRuleByID(rec, req)

			resp := decodeTestResponse(t, rec)
			if rec.Code != http.StatusOK || resp.Code != 0 {
				t.Fatalf("status/code = %d/%d, want %d/%d, body=%s", rec.Code, resp.Code, http.StatusOK, 0, rec.Body.String())
			}
			tc.assert(t)
		})
	}
}

// TestRouter_PriceRuleInvalidInput 验证价格规则接口对非法 ID、非法 JSON 和错误方法返回统一错误。
func TestRouter_PriceRuleInvalidInput(t *testing.T) {
	router := newPriceRuleTestRouter(&fakeAdminService{})
	cases := []struct {
		name       string
		method     string
		path       string
		body       string
		call       func(*Router, http.ResponseWriter, *http.Request)
		httpStatus int
		code       int
	}{
		{name: "invalid id", method: http.MethodGet, path: "/admin/v1/price-rules/abc", call: (*Router).handlePriceRuleByID, httpStatus: http.StatusBadRequest, code: 40001},
		{name: "invalid json", method: http.MethodPost, path: "/admin/v1/price-rules", body: "{", call: (*Router).handlePriceRules, httpStatus: http.StatusBadRequest, code: 40001},
		{name: "method not allowed", method: http.MethodDelete, path: "/admin/v1/price-rules", call: (*Router).handlePriceRules, httpStatus: http.StatusMethodNotAllowed, code: 40001},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := newSessionRequest(tc.method, tc.path, tc.body)
			rec := httptest.NewRecorder()

			tc.call(router, rec, req)

			resp := decodeTestResponse(t, rec)
			if rec.Code != tc.httpStatus || resp.Code != tc.code {
				t.Fatalf("status/code = %d/%d, want %d/%d", rec.Code, resp.Code, tc.httpStatus, tc.code)
			}
		})
	}
}

// TestRouter_AllImplementedAdminRoutesSmoke 覆盖当前 router.go 已注册的全部后台接口。
// 该用例使用 fake adminsvc 做最小成功返回，目标是验证 HTTP 路由、鉴权中间件、JSON 解析和统一响应不会断。
func TestRouter_AllImplementedAdminRoutesSmoke(t *testing.T) {
	router := NewRouter(&svc.ServiceContext{AdminSvc: &fakeAdminService{}})
	couponBody := `{"name":"新人券","type":1,"face_value":"10.00","discount":"","threshold_amount":"30.00","total_count":100,"per_user_limit":1,"valid_start_at":"2026-08-20 00:00:00","valid_end_at":"2026-12-31 23:59:59","status":1}`
	activityBody := `{"name":"暑期活动","type":1,"config":"{}","start_at":"2026-08-20 00:00:00","end_at":"2026-12-31 23:59:59","status":1}`
	activityActionBody := `{"publish_scope":"all","target_config":"{}"}`
	blacklistBody := `{"target_type":"user","target_id":1001,"reason":"risk"}`

	cases := []struct {
		name   string
		method string
		path   string
		body   string
		token  bool
	}{
		{name: "root", method: http.MethodGet, path: "/"},
		{name: "health", method: http.MethodGet, path: "/healthz"},
		{name: "register", method: http.MethodPost, path: "/admin/v1/auth/register", body: `{"username":"ops","password":"secret","real_name":"运营","role":2}`},
		{name: "login", method: http.MethodPost, path: "/admin/v1/auth/login", body: `{"username":"root","password":"secret"}`},
		{name: "logout", method: http.MethodPost, path: "/admin/v1/auth/logout", token: true},
		{name: "me", method: http.MethodGet, path: "/admin/v1/auth/me", token: true},
		{name: "menus", method: http.MethodGet, path: "/admin/v1/menus", token: true},
		{name: "operation logs", method: http.MethodGet, path: "/admin/v1/operation-logs?page=1&page_size=20&module=user", token: true},
		{name: "users list", method: http.MethodGet, path: "/admin/v1/users?page=1&page_size=20&keyword=138&status=1", token: true},
		{name: "user detail", method: http.MethodGet, path: "/admin/v1/users/1", token: true},
		{name: "user freeze", method: http.MethodPost, path: "/admin/v1/users/1/freeze", body: `{"reason":"risk","remark":"冻结测试"}`, token: true},
		{name: "user unfreeze", method: http.MethodPost, path: "/admin/v1/users/1/unfreeze", body: `{"reason":"ok","remark":"解封测试"}`, token: true},
		{name: "driver cert list", method: http.MethodGet, path: "/admin/v1/driver-certifications?page=1&page_size=20&audit_status=1", token: true},
		{name: "driver cert detail", method: http.MethodGet, path: "/admin/v1/driver-certifications/1", token: true},
		{name: "driver cert approve", method: http.MethodPost, path: "/admin/v1/driver-certifications/1/approve", body: `{"remark":"通过"}`, token: true},
		{name: "driver cert reject", method: http.MethodPost, path: "/admin/v1/driver-certifications/1/reject", body: `{"remark":"资料不完整"}`, token: true},
		{name: "orders list", method: http.MethodGet, path: "/admin/v1/orders?page=1&page_size=20&status=4", token: true},
		{name: "abnormal orders", method: http.MethodGet, path: "/admin/v1/orders/abnormal?page=1&page_size=20&abnormal_type=payment", token: true},
		{name: "order detail", method: http.MethodGet, path: "/admin/v1/orders/1", token: true},
		{name: "order cancel", method: http.MethodPost, path: "/admin/v1/orders/1/cancel", body: `{"reason":"客服取消"}`, token: true},
		{name: "coupons list", method: http.MethodGet, path: "/admin/v1/coupons?page=1&page_size=20&type=1&status=1", token: true},
		{name: "coupon create", method: http.MethodPost, path: "/admin/v1/coupons", body: couponBody, token: true},
		{name: "coupon update", method: http.MethodPut, path: "/admin/v1/coupons/1", body: couponBody, token: true},
		{name: "coupon disable", method: http.MethodPost, path: "/admin/v1/coupons/1/disable", token: true},
		{name: "coupon issue", method: http.MethodPost, path: "/admin/v1/coupons/1/issue", body: `{"target_type":"explicit_users","target_config":"{\"user_ids\":[1001]}"}`, token: true},
		{name: "coupon issue tasks", method: http.MethodGet, path: "/admin/v1/coupon-issue-tasks?page=1&page_size=20&coupon_id=1&status=success", token: true},
		{name: "price rules list", method: http.MethodGet, path: "/admin/v1/price-rules?page=1&page_size=20&city_code=110100", token: true},
		{name: "price rule create", method: http.MethodPost, path: "/admin/v1/price-rules", body: validPriceRuleBody(), token: true},
		{name: "price rule detail", method: http.MethodGet, path: "/admin/v1/price-rules/9", token: true},
		{name: "price rule update", method: http.MethodPut, path: "/admin/v1/price-rules/9", body: validPriceRuleBody(), token: true},
		{name: "price rule enable", method: http.MethodPost, path: "/admin/v1/price-rules/9/enable", token: true},
		{name: "price rule disable", method: http.MethodPost, path: "/admin/v1/price-rules/9/disable", token: true},
		{name: "promotion list", method: http.MethodGet, path: "/admin/v1/promotion-activities?page=1&page_size=20&type=1", token: true},
		{name: "promotion create", method: http.MethodPost, path: "/admin/v1/promotion-activities", body: activityBody, token: true},
		{name: "promotion update", method: http.MethodPut, path: "/admin/v1/promotion-activities/1", body: activityBody, token: true},
		{name: "promotion publish", method: http.MethodPost, path: "/admin/v1/promotion-activities/1/publish", body: activityActionBody, token: true},
		{name: "promotion rollback", method: http.MethodPost, path: "/admin/v1/promotion-activities/1/rollback", body: activityActionBody, token: true},
		{name: "statistics overview", method: http.MethodGet, path: "/admin/v1/statistics/overview?city_code=110100", token: true},
		{name: "statistics orders", method: http.MethodGet, path: "/admin/v1/statistics/orders?city_code=110100", token: true},
		{name: "statistics coupons", method: http.MethodGet, path: "/admin/v1/statistics/coupons?city_code=110100", token: true},
		{name: "export tasks list", method: http.MethodGet, path: "/admin/v1/export-tasks?page=1&page_size=20&export_type=orders", token: true},
		{name: "export task create", method: http.MethodPost, path: "/admin/v1/export-tasks", body: `{"export_type":"orders","filters":"{\"status\":5}"}`, token: true},
		{name: "export task detail", method: http.MethodGet, path: "/admin/v1/export-tasks/EX20260820120000000001", token: true},
		{name: "blacklist list", method: http.MethodGet, path: "/admin/v1/blacklist?page=1&page_size=20&target_type=user&target_id=1001", token: true},
		{name: "blacklist add", method: http.MethodPost, path: "/admin/v1/blacklist", body: blacklistBody, token: true},
		{name: "blacklist release", method: http.MethodPost, path: "/admin/v1/blacklist/1/release", body: blacklistBody, token: true},
		{name: "risk hit records", method: http.MethodGet, path: "/admin/v1/risk/hit-records?page=1&page_size=20&target_type=user&scene=login", token: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			req.RemoteAddr = "127.0.0.1:12345"
			if tc.token {
				req.Header.Set("Authorization", "Bearer root-token")
			}
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			resp := decodeTestResponse(t, rec)
			if rec.Code != http.StatusOK || resp.Code != 0 {
				t.Fatalf("%s %s status/code = %d/%d, want %d/%d, body=%s", tc.method, tc.path, rec.Code, resp.Code, http.StatusOK, 0, rec.Body.String())
			}
		})
	}
}
