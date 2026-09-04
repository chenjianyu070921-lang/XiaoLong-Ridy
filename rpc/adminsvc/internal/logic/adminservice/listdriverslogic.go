package adminservicelogic

import (
	"context"
	"time"

	"XiaoLong-Ridy/rpc/adminsvc/adminsvc"
	"XiaoLong-Ridy/rpc/adminsvc/internal/svc"
	driverproto "XiaoLong-Ridy/rpc/driversvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ListDriversLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListDriversLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListDriversLogic {
	return &ListDriversLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListDriversLogic) ListDrivers(in *adminsvc.DriverListRequest) (*adminsvc.DriverListResponse, error) {
	if l.svcCtx == nil || l.svcCtx.DriverSvc == nil {
		return nil, status.Error(codes.FailedPrecondition, "driver service is not running or downstream RPC is disabled")
	}

	// 管理后台司机列表必须以 driversvc 为权威数据源，避免跨服务直查司机库导致数据边界失效。
	req := &driverproto.ListDriversRequest{
		Page:     in.GetPage(),
		PageSize: in.GetPageSize(),
		Keyword:  in.GetKeyword(),
	}
	if in.GetStatus() > 0 {
		driverStatus := driverproto.DriverStatus(in.GetStatus())
		req.Status = &driverStatus
	}

	result, err := l.svcCtx.DriverSvc.ListDrivers(l.ctx, req)
	if err != nil {
		return nil, err
	}
	list := make([]*adminsvc.Driver, 0, len(result.GetDrivers()))
	for _, item := range result.GetDrivers() {
		list = append(list, filterAdminDriverSensitive(mapDriverFromDriverSvc(item), false))
	}
	return &adminsvc.DriverListResponse{
		List:     list,
		Total:    result.GetTotal(),
		Page:     normalizePage(in.GetPage()),
		PageSize: normalizePageSize(in.GetPageSize()),
	}, nil
}

// filterAdminDriverSensitive 按后台敏感字段权限裁剪司机手机号、身份证号和驾驶证号。
// canViewSensitive 为 true 时保留 driversvc 返回的完整值；否则返回服务端脱敏值。
func filterAdminDriverSensitive(item *adminsvc.Driver, canViewSensitive bool) *adminsvc.Driver {
	if item == nil {
		return nil
	}
	if canViewSensitive {
		return item
	}
	item.Phone = maskPhone(item.Phone)
	item.IdCardNo = maskIDCard(item.IdCardNo)
	item.DriverLicenseNo = maskIDCard(item.DriverLicenseNo)
	return item
}

// mapDriverFromDriverSvc 将 driversvc 返回的司机基础资料转换为管理后台展示模型。
// 当前 driversvc 的 ListDrivers/GetDriver 只返回司机基础字段，车辆和认证聚合字段保持零值，
// 后续如需展示车辆、认证、服务统计，应先由 driversvc 扩展后台聚合 RPC，再由 adminsvc 适配。
func mapDriverFromDriverSvc(item *driverproto.Driver) *adminsvc.Driver {
	if item == nil {
		return nil
	}
	return &adminsvc.Driver{
		Id:              item.GetId(),
		Phone:           item.GetPhone(),
		RealName:        item.GetRealName(),
		IdCardNo:        item.GetIdCardNo(),
		DriverLicenseNo: item.GetDriverLicenseNo(),
		AvatarUrl:       item.GetAvatarUrl(),
		Status:          int32(item.GetStatus()),
		OnlineStatus:    item.GetOnlineStatus(),
		VehicleId:       item.GetVehicleId(),
		PlateNo:         item.GetPlateNo(),
		VehicleStatus:   item.GetVehicleStatus(),
		CertificationId: item.GetCertificationId(),
		AuditStatus:     item.GetAuditStatus(),
		AuditRemark:     item.GetAuditRemark(),
		CreatedAt:       formatUnixSecond(item.GetCreatedAt()),
		UpdatedAt:       formatUnixSecond(item.GetUpdatedAt()),
	}
}

// formatUnixSecond 将下游 RPC 的秒级时间戳格式化为后台统一时间字符串。
func formatUnixSecond(ts int64) string {
	if ts <= 0 {
		return ""
	}
	return formatTime(time.Unix(ts, 0))
}
