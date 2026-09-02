package adminservicelogic

import (
	"context"
	"sort"
	"strconv"
	"sync"
	"time"

	"XiaoLong-Ridy/rpc/adminsvc/adminsvc"
	"XiaoLong-Ridy/rpc/adminsvc/internal/svc"
	driverproto "XiaoLong-Ridy/rpc/driversvc/proto"
	locationsvc "XiaoLong-Ridy/rpc/locationsvc/locationsvc"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	defaultCapacityMapLimit = 100
	maxCapacityMapLimit     = 200
	capacityLocationWorkers = 16
)

// GetCapacityMapLogic 负责聚合司机服务和位置服务，生成后台运力地图快照。
type GetCapacityMapLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGetCapacityMapLogic 创建运力地图快照逻辑对象。
func NewGetCapacityMapLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCapacityMapLogic {
	return &GetCapacityMapLogic{ctx: ctx, svcCtx: svcCtx}
}

// GetCapacityMap 查询司机列表并并发读取最新位置。
// 单个司机位置失败只计数并跳过该点，避免少量脏数据阻断整张地图。
func (l *GetCapacityMapLogic) GetCapacityMap(in *adminsvc.CapacityMapRequest) (*adminsvc.CapacityMapResponse, error) {
	if err := requireAdminRoles(l.ctx, l.svcCtx, 1, 2, 3); err != nil {
		return nil, err
	}
	if l.svcCtx == nil || l.svcCtx.DriverSvc == nil || l.svcCtx.LocationSvc == nil {
		return nil, status.Error(codes.Unavailable, "司机或位置服务未启动")
	}
	limit := int(in.GetLimit())
	if limit <= 0 {
		limit = defaultCapacityMapLimit
	}
	if limit > maxCapacityMapLimit {
		limit = maxCapacityMapLimit
	}
	request := &driverproto.ListDriversRequest{Page: 1, PageSize: int32(limit)}
	if in.GetStatus() > 0 {
		statusValue := driverproto.DriverStatus(in.GetStatus())
		request.Status = &statusValue
	}
	list, err := l.svcCtx.DriverSvc.ListDrivers(l.ctx, request)
	if err != nil {
		return nil, err
	}
	items := list.GetDrivers()
	result := &adminsvc.CapacityMapResponse{Drivers: make([]*adminsvc.CapacityDriver, 0, len(items)), Total: int32(len(items))}
	var mu sync.Mutex
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, capacityLocationWorkers)
	for _, driver := range items {
		driver := driver
		if driver == nil || driver.GetId() <= 0 {
			result.PositionFailureCount++
			continue
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			semaphore <- struct{}{}
			location, locationErr := l.svcCtx.LocationSvc.GetDriverLocation(l.ctx, &locationsvc.GetDriverLocationReq{DriverId: driver.GetId()})
			<-semaphore
			if locationErr != nil || location == nil || !validCoordinate(location.GetLng(), location.GetLat()) {
				mu.Lock()
				result.PositionFailureCount++
				mu.Unlock()
				return
			}
			if in.GetOnlineStatus() > 0 && location.GetOnlineStatus() != in.GetOnlineStatus() {
				return
			}
			item := &adminsvc.CapacityDriver{DriverId: driver.GetId(), DriverStatus: int32(driver.GetStatus()), OnlineStatus: location.GetOnlineStatus(), Lng: strconv.FormatFloat(location.GetLng(), 'f', -1, 64), Lat: strconv.FormatFloat(location.GetLat(), 'f', -1, 64), Heading: strconv.Itoa(int(location.GetHeading())), SpeedKmh: strconv.FormatFloat(location.GetSpeedKmh(), 'f', -1, 64), ReportTime: formatCapacityTime(location.GetReportTime())}
			mu.Lock()
			result.Drivers = append(result.Drivers, item)
			mu.Unlock()
		}()
	}
	wg.Wait()
	sort.Slice(result.Drivers, func(i, j int) bool { return result.Drivers[i].GetDriverId() < result.Drivers[j].GetDriverId() })
	for _, item := range result.Drivers {
		switch item.GetOnlineStatus() {
		case 1:
			result.AvailableCount++
		case 2:
			result.InTripCount++
		default:
			result.OfflineCount++
		}
	}
	result.GeneratedAt = time.Now().Format(statisticsTimeLayout)
	return result, nil
}

// validCoordinate 校验经纬度范围，拒绝位置服务返回的空值或越界值。
func validCoordinate(lng, lat float64) bool {
	return lng >= -180 && lng <= 180 && lat >= -90 && lat <= 90 && !(lng == 0 && lat == 0)
}

// formatCapacityTime 将位置服务 Unix 秒转换为后台统一时间文本。
func formatCapacityTime(ts int64) string {
	if ts <= 0 {
		return ""
	}
	return time.Unix(ts, 0).Format(statisticsTimeLayout)
}
