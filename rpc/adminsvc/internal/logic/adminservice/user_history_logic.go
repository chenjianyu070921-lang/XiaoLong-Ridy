package adminservicelogic

import (
	"context"
	"fmt"
	"time"

	"XiaoLong-Ridy/rpc/adminsvc/adminsvc"
	"XiaoLong-Ridy/rpc/adminsvc/internal/svc"
	ordersvc "XiaoLong-Ridy/rpc/ordersvc/proto"
	usersvc "XiaoLong-Ridy/rpc/usersvc/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UserHistoryLogic 负责管理后台用户订单与优惠券历史的跨服务只读查询。
type UserHistoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewUserHistoryLogic 创建只读查询逻辑实例。
func NewUserHistoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserHistoryLogic {
	return &UserHistoryLogic{ctx: ctx, svcCtx: svcCtx}
}

// ListUserOrders 通过 ordersvc 查询指定用户的订单历史，不直接读取订单表。
func (l *UserHistoryLogic) ListUserOrders(in *adminsvc.UserHistoryRequest) (*adminsvc.OrderListResponse, error) {
	if in.GetUserId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "用户 ID 必须大于 0")
	}
	if l.svcCtx.OrdersSvc == nil {
		return nil, status.Error(codes.Unavailable, "订单服务未启动")
	}
	resp, err := l.svcCtx.OrdersSvc.ListOrders(l.ctx, &ordersvc.ListOrdersRequest{UserId: in.GetUserId(), Status: ordersvc.OrderStatus(in.GetStatus()), Page: in.GetPage(), PageSize: in.GetPageSize()})
	if err != nil {
		return nil, err
	}
	list := make([]*adminsvc.Order, 0, len(resp.GetList()))
	for _, item := range resp.GetList() {
		list = append(list, &adminsvc.Order{Id: item.GetOrderId(), OrderNo: item.GetOrderNo(), UserId: in.GetUserId(), FromAddress: item.GetFromAddress(), ToAddress: item.GetToAddress(), Status: int32(item.GetStatus()), EstimatedPrice: formatCents(item.GetEstimatedPriceCents()), CreatedAt: unixText(item.GetCreatedAt())})
	}
	return &adminsvc.OrderListResponse{List: list, Total: resp.GetTotal(), Page: resp.GetPage(), PageSize: resp.GetPageSize()}, nil
}

// ListUserCoupons 通过 usersvc 查询用户券，并将分页参数透传到下游数据库查询。
func (l *UserHistoryLogic) ListUserCoupons(in *adminsvc.UserCouponHistoryRequest) (*adminsvc.UserCouponHistoryResponse, error) {
	if in.GetUserId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "用户 ID 必须大于 0")
	}
	if l.svcCtx.UsersSvc == nil {
		return nil, status.Error(codes.Unavailable, "用户服务未启动")
	}
	page, size := normalizePage(in.GetPage()), normalizePageSize(in.GetPageSize())
	resp, err := l.svcCtx.UsersSvc.ListMyCoupons(l.ctx, &usersvc.ListMyCouponsRequest{
		UserId:   uint64(in.GetUserId()),
		Status:   in.GetStatus(),
		Page:     page,
		PageSize: size,
	})
	if err != nil {
		return nil, err
	}
	list := make([]*adminsvc.UserCouponHistory, 0, len(resp.GetList()))
	for _, item := range resp.GetList() {
		list = append(list, &adminsvc.UserCouponHistory{UserCouponId: int64(item.GetUserCouponId()), CouponId: int64(item.GetCouponId()), Name: item.GetName(), Type: item.GetType(), FaceValueCents: item.GetFaceValueCents(), Discount: item.GetDiscount(), ThresholdCents: item.GetThresholdCents(), CarType: item.GetCarType(), CityCode: item.GetCityCode(), Status: item.GetStatus(), ReceivedAt: item.GetReceivedAt(), ExpireAt: item.GetExpireAt()})
	}
	return &adminsvc.UserCouponHistoryResponse{List: list, Total: resp.GetTotal(), Page: resp.GetPage(), PageSize: resp.GetPageSize()}, nil
}

// formatCents 将分单位金额转换为后台展示使用的两位小数字符串。
func formatCents(cents int64) string { return fmt.Sprintf("%d.%02d", cents/100, cents%100) }

// unixText 统一将下游 Unix 秒时间戳转换为管理端接口约定的时间字符串。
func unixText(timestamp int64) string {
	if timestamp <= 0 {
		return ""
	}
	return time.Unix(timestamp, 0).Format("2006-01-02 15:04:05")
}
