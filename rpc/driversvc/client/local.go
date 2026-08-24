package client

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"time"

	driversproto "XiaoLong-Ridy/rpc/driversvc/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// LocalClient 是本地开发和测试使用的 driversvc 内存实现。
type LocalClient struct {
	mu                  sync.RWMutex
	nextDriverID        uint64
	nextVehicleID       uint64
	nextCertificationID uint64
	drivers             map[uint64]*driversproto.Driver
	driverByPhone       map[string]uint64
	vehicles            map[uint64]*driversproto.Vehicle
	certByID            map[uint64]*driversproto.CertificationInfo
	certByDriver        map[uint64]uint64
	devices             map[uint64]string
	locations           map[uint64]*localDriverLocation
}

type localDriverLocation struct {
	longitude   float64
	latitude    float64
	reportTime  int64
	onlineState int32
}

// NewLocalClient 创建本地联调用的 driversvc 客户端。
func NewLocalClient() *LocalClient {
	return &LocalClient{
		nextDriverID:        1,
		nextVehicleID:       1,
		nextCertificationID: 1,
		drivers:             make(map[uint64]*driversproto.Driver),
		driverByPhone:       make(map[string]uint64),
		vehicles:            make(map[uint64]*driversproto.Vehicle),
		certByID:            make(map[uint64]*driversproto.CertificationInfo),
		certByDriver:        make(map[uint64]uint64),
		devices:             make(map[uint64]string),
		locations:           make(map[uint64]*localDriverLocation),
	}
}

// CreateDriver 创建司机账号。
func (c *LocalClient) CreateDriver(_ context.Context, req *driversproto.CreateDriverRequest) (*driversproto.CreateDriverResponse, error) {
	return c.createDriver(req)
}

// RegisterDriver 创建司机自注册账号。
func (c *LocalClient) RegisterDriver(_ context.Context, req *driversproto.CreateDriverRequest) (*driversproto.CreateDriverResponse, error) {
	return c.createDriver(req)
}

func (c *LocalClient) createDriver(req *driversproto.CreateDriverRequest) (*driversproto.CreateDriverResponse, error) {
	if req == nil || strings.TrimSpace(req.GetPhone()) == "" {
		return nil, status.Error(codes.InvalidArgument, "phone is required")
	}
	now := time.Now().Unix()

	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.driverByPhone[req.GetPhone()]; exists {
		return nil, status.Error(codes.AlreadyExists, "driver already exists")
	}

	id := c.nextDriverID
	c.nextDriverID++
	driver := &driversproto.Driver{
		Id:              int64(id),
		Phone:           req.GetPhone(),
		PasswordHash:    req.GetPasswordHash(),
		RealName:        req.GetRealName(),
		IdCardNo:        req.GetIdCardNo(),
		DriverLicenseNo: req.GetDriverLicenseNo(),
		AvatarUrl:       req.GetAvatarUrl(),
		Status:          driversproto.DriverStatus_DRIVER_STATUS_PENDING,
		CreatedAt:       now,
		UpdatedAt:       now,
		OnlineStatus:    0,
	}
	c.drivers[id] = cloneDriver(driver)
	c.driverByPhone[driver.GetPhone()] = id
	return &driversproto.CreateDriverResponse{
		Id:        int64(id),
		Status:    driver.GetStatus(),
		CreatedAt: now,
	}, nil
}

// UpdateDriver 更新司机信息。
func (c *LocalClient) UpdateDriver(_ context.Context, req *driversproto.UpdateDriverRequest) (*driversproto.UpdateDriverResponse, error) {
	if req == nil || req.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "driver id is required")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	driver, ok := c.drivers[uint64(req.GetId())]
	if !ok {
		return nil, status.Error(codes.NotFound, "driver not found")
	}
	if req.Phone != nil && *req.Phone != "" {
		if oldID, exists := c.driverByPhone[*req.Phone]; exists && oldID != uint64(req.GetId()) {
			return nil, status.Error(codes.AlreadyExists, "driver phone already exists")
		}
		delete(c.driverByPhone, driver.GetPhone())
		driver.Phone = *req.Phone
		c.driverByPhone[driver.GetPhone()] = uint64(req.GetId())
	}
	if req.PasswordHash != nil {
		driver.PasswordHash = *req.PasswordHash
	}
	if req.RealName != nil {
		driver.RealName = *req.RealName
	}
	if req.IdCardNo != nil {
		driver.IdCardNo = *req.IdCardNo
	}
	if req.DriverLicenseNo != nil {
		driver.DriverLicenseNo = *req.DriverLicenseNo
	}
	if req.AvatarUrl != nil {
		driver.AvatarUrl = *req.AvatarUrl
	}
	if req.Status != nil {
		driver.Status = *req.Status
	}
	driver.UpdatedAt = time.Now().Unix()
	c.drivers[uint64(req.GetId())] = cloneDriver(driver)
	return &driversproto.UpdateDriverResponse{
		Id:        req.GetId(),
		Status:    driver.GetStatus(),
		UpdatedAt: driver.GetUpdatedAt(),
	}, nil
}

// DeleteDriver 删除司机账号。
func (c *LocalClient) DeleteDriver(_ context.Context, req *driversproto.DeleteDriverRequest) (*driversproto.DeleteDriverResponse, error) {
	if req == nil || req.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "driver id is required")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	driver, ok := c.drivers[uint64(req.GetId())]
	if !ok {
		return nil, status.Error(codes.NotFound, "driver not found")
	}
	delete(c.driverByPhone, driver.GetPhone())
	delete(c.devices, uint64(req.GetId()))
	delete(c.locations, uint64(req.GetId()))
	delete(c.drivers, uint64(req.GetId()))
	return &driversproto.DeleteDriverResponse{Id: req.GetId(), Success: true}, nil
}

// GetDriver 按 ID 查询司机详情。
func (c *LocalClient) GetDriver(_ context.Context, req *driversproto.GetDriverRequest) (*driversproto.GetDriverResponse, error) {
	if req == nil || req.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "driver id is required")
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	driver, ok := c.drivers[uint64(req.GetId())]
	if !ok {
		return nil, status.Error(codes.NotFound, "driver not found")
	}
	return &driversproto.GetDriverResponse{Driver: cloneDriver(driver)}, nil
}

// GetDriverByPhone 按手机号查询司机。
func (c *LocalClient) GetDriverByPhone(_ context.Context, req *driversproto.GetDriverByPhoneRequest) (*driversproto.GetDriverByPhoneResponse, error) {
	if req == nil || strings.TrimSpace(req.GetPhone()) == "" {
		return nil, status.Error(codes.InvalidArgument, "phone is required")
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	id, ok := c.driverByPhone[req.GetPhone()]
	if !ok {
		return nil, status.Error(codes.NotFound, "driver not found")
	}
	return &driversproto.GetDriverByPhoneResponse{Driver: cloneDriver(c.drivers[id])}, nil
}

// SetDriverOnline 将司机置为在线。
func (c *LocalClient) SetDriverOnline(_ context.Context, req *driversproto.SetDriverOnlineRequest) (*driversproto.SetDriverOnlineResponse, error) {
	driverID, onlineStatus, kicked, err := c.setOnlineState(req.GetDriverId(), req.GetDeviceId(), req.GetLongitude(), req.GetLatitude(), driversproto.DriverStatus_DRIVER_STATUS_NORMAL, 1)
	if err != nil {
		return nil, err
	}
	return &driversproto.SetDriverOnlineResponse{
		DriverId:     driverID,
		OnlineStatus: onlineStatus,
		Kicked:       kicked,
	}, nil
}

// SetDriverOffline 将司机置为离线。
func (c *LocalClient) SetDriverOffline(_ context.Context, req *driversproto.SetDriverOfflineRequest) (*driversproto.SetDriverOfflineResponse, error) {
	driverID, onlineStatus, kicked, err := c.setOnlineState(req.GetDriverId(), req.GetDeviceId(), req.GetLongitude(), req.GetLatitude(), driversproto.DriverStatus_DRIVER_STATUS_NORMAL, 0)
	if err != nil {
		return nil, err
	}
	return &driversproto.SetDriverOfflineResponse{
		DriverId:     driverID,
		OnlineStatus: onlineStatus,
		Kicked:       kicked,
	}, nil
}

// ReportLocation 上报司机位置。
func (c *LocalClient) ReportLocation(_ context.Context, req *driversproto.ReportLocationRequest) (*driversproto.ReportLocationResponse, error) {
	if req == nil || req.GetDriverId() <= 0 || strings.TrimSpace(req.GetDeviceId()) == "" {
		return nil, status.Error(codes.InvalidArgument, "driver id and device id are required")
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	driver, ok := c.drivers[uint64(req.GetDriverId())]
	if !ok {
		return nil, status.Error(codes.NotFound, "driver not found")
	}
	kicked := false
	if prev := c.devices[uint64(req.GetDriverId())]; prev != "" && prev != req.GetDeviceId() {
		kicked = true
	}
	c.devices[uint64(req.GetDriverId())] = req.GetDeviceId()
	driver.OnlineStatus = 1
	now := time.Now().Unix()
	c.locations[uint64(req.GetDriverId())] = &localDriverLocation{
		longitude:   req.GetLongitude(),
		latitude:    req.GetLatitude(),
		reportTime:  now,
		onlineState: 1,
	}
	driver.UpdatedAt = now
	c.drivers[uint64(req.GetDriverId())] = cloneDriver(driver)
	return &driversproto.ReportLocationResponse{
		DriverId:     req.GetDriverId(),
		OnlineStatus: 1,
		Kicked:       kicked,
		ReportTime:   now,
	}, nil
}

// SetDriverServiceStatus 同步司机服务状态。
func (c *LocalClient) SetDriverServiceStatus(_ context.Context, req *driversproto.SetDriverServiceStatusRequest) (*driversproto.SetDriverServiceStatusResponse, error) {
	if req == nil || req.GetDriverId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "driver id is required")
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	driver, ok := c.drivers[uint64(req.GetDriverId())]
	if !ok {
		return nil, status.Error(codes.NotFound, "driver not found")
	}
	driver.OnlineStatus = req.GetOnlineStatus()
	driver.UpdatedAt = time.Now().Unix()
	c.drivers[uint64(req.GetDriverId())] = cloneDriver(driver)
	if loc, ok := c.locations[uint64(req.GetDriverId())]; ok {
		loc.onlineState = req.GetOnlineStatus()
		loc.reportTime = driver.GetUpdatedAt()
	}
	return &driversproto.SetDriverServiceStatusResponse{
		DriverId:     req.GetDriverId(),
		OnlineStatus: req.GetOnlineStatus(),
		UpdatedAt:    driver.GetUpdatedAt(),
	}, nil
}

// Heartbeat 刷新在线状态并判定多端互踢。
func (c *LocalClient) Heartbeat(_ context.Context, req *driversproto.HeartbeatRequest) (*driversproto.HeartbeatResponse, error) {
	if req == nil || req.GetDriverId() <= 0 || strings.TrimSpace(req.GetDeviceId()) == "" {
		return nil, status.Error(codes.InvalidArgument, "driver id and device id are required")
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	driver, ok := c.drivers[uint64(req.GetDriverId())]
	if !ok {
		return nil, status.Error(codes.NotFound, "driver not found")
	}
	kicked := false
	if prev := c.devices[uint64(req.GetDriverId())]; prev != "" && prev != req.GetDeviceId() {
		kicked = true
	}
	c.devices[uint64(req.GetDriverId())] = req.GetDeviceId()
	now := time.Now().Unix()
	driver.OnlineStatus = 1
	driver.UpdatedAt = now
	c.locations[uint64(req.GetDriverId())] = &localDriverLocation{
		longitude:   req.GetLongitude(),
		latitude:    req.GetLatitude(),
		reportTime:  now,
		onlineState: 1,
	}
	c.drivers[uint64(req.GetDriverId())] = cloneDriver(driver)
	return &driversproto.HeartbeatResponse{
		OnlineStatus: 1,
		Kicked:       kicked,
		ServerTime:   now,
	}, nil
}

func (c *LocalClient) setOnlineState(driverID int64, deviceID string, longitude, latitude float64, _ driversproto.DriverStatus, onlineStatus int32) (int64, int32, bool, error) {
	if driverID <= 0 || strings.TrimSpace(deviceID) == "" {
		return 0, 0, false, status.Error(codes.InvalidArgument, "driver id and device id are required")
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	driver, ok := c.drivers[uint64(driverID)]
	if !ok {
		return 0, 0, false, status.Error(codes.NotFound, "driver not found")
	}
	kicked := false
	if prev := c.devices[uint64(driverID)]; prev != "" && prev != deviceID {
		kicked = true
	}
	c.devices[uint64(driverID)] = deviceID
	now := time.Now().Unix()
	driver.OnlineStatus = onlineStatus
	driver.UpdatedAt = now
	c.locations[uint64(driverID)] = &localDriverLocation{
		longitude:   longitude,
		latitude:    latitude,
		reportTime:  now,
		onlineState: onlineStatus,
	}
	c.drivers[uint64(driverID)] = cloneDriver(driver)
	return driverID, onlineStatus, kicked, nil
}

// CreateVehicle 创建车辆。
func (c *LocalClient) CreateVehicle(_ context.Context, req *driversproto.CreateVehicleRequest) (*driversproto.CreateVehicleResponse, error) {
	if req == nil || req.GetDriverId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "driver id is required")
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.drivers[uint64(req.GetDriverId())]; !ok {
		return nil, status.Error(codes.NotFound, "driver not found")
	}
	id := c.nextVehicleID
	c.nextVehicleID++
	now := time.Now().Unix()
	vehicle := &driversproto.Vehicle{
		Id:          int64(id),
		DriverId:    req.GetDriverId(),
		PlateNo:     req.GetPlateNo(),
		Brand:       req.GetBrand(),
		Model:       req.GetModel(),
		Color:       req.GetColor(),
		VehicleType: req.GetVehicleType(),
		InsuranceNo: req.GetInsuranceNo(),
		Status:      driversproto.VehicleStatus_VEHICLE_STATUS_PENDING,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if req.RegistrationDate != nil {
		registrationDate := *req.RegistrationDate
		vehicle.RegistrationDate = &registrationDate
	}
	if req.InsuranceExpireAt != nil {
		insuranceExpireAt := *req.InsuranceExpireAt
		vehicle.InsuranceExpireAt = &insuranceExpireAt
	}
	c.vehicles[id] = cloneVehicle(vehicle)
	return &driversproto.CreateVehicleResponse{Id: int64(id), Status: vehicle.GetStatus(), CreatedAt: now}, nil
}

// UpdateVehicle 更新车辆信息。
func (c *LocalClient) UpdateVehicle(_ context.Context, req *driversproto.UpdateVehicleRequest) (*driversproto.UpdateVehicleResponse, error) {
	if req == nil || req.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "vehicle id is required")
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	vehicle, ok := c.vehicles[uint64(req.GetId())]
	if !ok {
		return nil, status.Error(codes.NotFound, "vehicle not found")
	}
	if req.DriverId != nil {
		vehicle.DriverId = *req.DriverId
	}
	if req.PlateNo != nil {
		vehicle.PlateNo = *req.PlateNo
	}
	if req.Brand != nil {
		vehicle.Brand = *req.Brand
	}
	if req.Model != nil {
		vehicle.Model = *req.Model
	}
	if req.Color != nil {
		vehicle.Color = *req.Color
	}
	if req.VehicleType != nil {
		vehicle.VehicleType = *req.VehicleType
	}
	if req.RegistrationDate != nil {
		registrationDate := *req.RegistrationDate
		vehicle.RegistrationDate = &registrationDate
	}
	if req.InsuranceNo != nil {
		vehicle.InsuranceNo = *req.InsuranceNo
	}
	if req.InsuranceExpireAt != nil {
		insuranceExpireAt := *req.InsuranceExpireAt
		vehicle.InsuranceExpireAt = &insuranceExpireAt
	}
	if req.Status != nil {
		vehicle.Status = *req.Status
	}
	vehicle.UpdatedAt = time.Now().Unix()
	c.vehicles[uint64(req.GetId())] = cloneVehicle(vehicle)
	return &driversproto.UpdateVehicleResponse{Id: req.GetId(), Status: vehicle.GetStatus(), UpdatedAt: vehicle.GetUpdatedAt()}, nil
}

// DeleteVehicle 删除车辆。
func (c *LocalClient) DeleteVehicle(_ context.Context, req *driversproto.DeleteVehicleRequest) (*driversproto.DeleteVehicleResponse, error) {
	if req == nil || req.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "vehicle id is required")
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.vehicles[uint64(req.GetId())]; !ok {
		return nil, status.Error(codes.NotFound, "vehicle not found")
	}
	delete(c.vehicles, uint64(req.GetId()))
	return &driversproto.DeleteVehicleResponse{Id: req.GetId(), Success: true}, nil
}

// GetVehicle 查询车辆详情。
func (c *LocalClient) GetVehicle(_ context.Context, req *driversproto.GetVehicleRequest) (*driversproto.GetVehicleResponse, error) {
	if req == nil || req.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "vehicle id is required")
	}
	c.mu.RLock()
	defer c.mu.RUnlock()

	vehicle, ok := c.vehicles[uint64(req.GetId())]
	if !ok {
		return nil, status.Error(codes.NotFound, "vehicle not found")
	}
	return &driversproto.GetVehicleResponse{Vehicle: cloneVehicle(vehicle)}, nil
}

// ListDrivers 分页查询司机列表。
func (c *LocalClient) ListDrivers(_ context.Context, req *driversproto.ListDriversRequest) (*driversproto.ListDriversResponse, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	list := make([]*driversproto.Driver, 0, len(c.drivers))
	for _, driver := range c.drivers {
		if req != nil && req.Status != nil && driver.GetStatus() != *req.Status {
			continue
		}
		if keyword := strings.TrimSpace(req.GetKeyword()); keyword != "" {
			if !strings.Contains(driver.GetPhone(), keyword) &&
				!strings.Contains(driver.GetRealName(), keyword) &&
				!strings.Contains(driver.GetDriverLicenseNo(), keyword) {
				continue
			}
		}
		list = append(list, cloneDriver(driver))
	}
	sort.Slice(list, func(i, j int) bool { return list[i].GetId() > list[j].GetId() })

	page := req.GetPage()
	if page <= 0 {
		page = 1
	}
	pageSize := req.GetPageSize()
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	start := int((page - 1) * pageSize)
	if start > len(list) {
		start = len(list)
	}
	end := start + int(pageSize)
	if end > len(list) {
		end = len(list)
	}
	return &driversproto.ListDriversResponse{
		Drivers: list[start:end],
		Total:   int64(len(list)),
	}, nil
}

// Login 密码登录。
func (c *LocalClient) Login(_ context.Context, req *driversproto.LoginRequest) (*driversproto.LoginResponse, error) {
	if req == nil || strings.TrimSpace(req.GetPhone()) == "" {
		return nil, status.Error(codes.InvalidArgument, "phone is required")
	}
	c.mu.RLock()
	defer c.mu.RUnlock()

	id, ok := c.driverByPhone[req.GetPhone()]
	if !ok {
		return nil, status.Error(codes.NotFound, "driver not found")
	}
	driver := c.drivers[id]
	if driver.GetStatus() == driversproto.DriverStatus_DRIVER_STATUS_FROZEN ||
		driver.GetStatus() == driversproto.DriverStatus_DRIVER_STATUS_CANCELLED {
		return nil, status.Error(codes.PermissionDenied, "driver unavailable")
	}
	if req.GetPassword() != "" && driver.GetPasswordHash() != "" && req.GetPassword() != driver.GetPasswordHash() {
		return nil, status.Error(codes.Unauthenticated, "password mismatch")
	}
	token := fmt.Sprintf("local-driver-token-%d", driver.GetId())
	return &driversproto.LoginResponse{
		Token:    token,
		ExpireIn: 7200,
		Driver:   cloneDriver(driver),
	}, nil
}

// LoginBySMS 短信登录。
func (c *LocalClient) LoginBySMS(_ context.Context, req *driversproto.LoginBySMSRequest) (*driversproto.LoginResponse, error) {
	if req == nil || strings.TrimSpace(req.GetPhone()) == "" {
		return nil, status.Error(codes.InvalidArgument, "phone is required")
	}
	c.mu.RLock()
	defer c.mu.RUnlock()

	id, ok := c.driverByPhone[req.GetPhone()]
	if !ok {
		return nil, status.Error(codes.NotFound, "driver not found")
	}
	driver := c.drivers[id]
	if driver.GetStatus() == driversproto.DriverStatus_DRIVER_STATUS_FROZEN ||
		driver.GetStatus() == driversproto.DriverStatus_DRIVER_STATUS_CANCELLED {
		return nil, status.Error(codes.PermissionDenied, "driver unavailable")
	}
	return &driversproto.LoginResponse{
		Token:    fmt.Sprintf("local-driver-token-%d", driver.GetId()),
		ExpireIn: 7200,
		Driver:   cloneDriver(driver),
	}, nil
}

// ListNearbyDrivers 查询附近司机。
func (c *LocalClient) ListNearbyDrivers(_ context.Context, req *driversproto.ListNearbyDriversRequest) (*driversproto.ListNearbyDriversResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	radius := req.GetRadiusMeters()
	if radius <= 0 {
		radius = 3000
	}
	limit := req.GetLimit()
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	type nearby struct {
		driverID uint64
		distance float64
	}
	candidates := make([]nearby, 0)
	for id, loc := range c.locations {
		if loc == nil || loc.onlineState != 1 {
			continue
		}
		distance := haversineMeters(req.GetLongitude(), req.GetLatitude(), loc.longitude, loc.latitude)
		if distance <= radius {
			candidates = append(candidates, nearby{driverID: id, distance: distance})
		}
	}
	sort.Slice(candidates, func(i, j int) bool { return candidates[i].distance < candidates[j].distance })
	if len(candidates) > int(limit) {
		candidates = candidates[:limit]
	}

	list := make([]*driversproto.NearbyDriver, 0, len(candidates))
	for _, item := range candidates {
		loc := c.locations[item.driverID]
		list = append(list, &driversproto.NearbyDriver{
			DriverId:       int64(item.driverID),
			Longitude:      loc.longitude,
			Latitude:       loc.latitude,
			DistanceMeters: int32(math.Round(item.distance)),
		})
	}
	return &driversproto.ListNearbyDriversResponse{Drivers: list}, nil
}

// GetDriverAiScore 查询司机 AI 推荐分。
func (c *LocalClient) GetDriverAiScore(_ context.Context, req *driversproto.GetDriverAiScoreRequest) (*driversproto.GetDriverAiScoreResponse, error) {
	if req == nil || req.GetDriverId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "driver id is required")
	}
	c.mu.RLock()
	defer c.mu.RUnlock()

	driver, ok := c.drivers[uint64(req.GetDriverId())]
	if !ok {
		return nil, status.Error(codes.NotFound, "driver not found")
	}
	score := 60.0
	if driver.GetStatus() == driversproto.DriverStatus_DRIVER_STATUS_NORMAL {
		score += 15
	}
	if driver.GetOnlineStatus() > 0 {
		score += 10
	}
	if _, ok := c.certByDriver[uint64(req.GetDriverId())]; ok {
		score += 5
	}
	if score > 100 {
		score = 100
	}
	return &driversproto.GetDriverAiScoreResponse{
		DriverId: req.GetDriverId(),
		AiScore:  score,
		Level:    int32(score/20) + 1,
		Factors: []*driversproto.AiScoreFactor{
			{Key: "status", Label: "账号状态", Value: float64(driver.GetStatus()), Impact: "neutral", Hint: "保持正常状态"},
			{Key: "online", Label: "在线状态", Value: float64(driver.GetOnlineStatus()), Impact: "positive", Hint: "在线可提升接单机会"},
		},
		Degraded:      false,
		DegradeReason: "",
	}, nil
}

// UploadCertification 上传司机资质。
func (c *LocalClient) UploadCertification(_ context.Context, req *driversproto.UploadCertificationRequest) (*driversproto.UploadCertificationResponse, error) {
	if req == nil || req.GetDriverId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "driver id is required")
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.drivers[uint64(req.GetDriverId())]; !ok {
		return nil, status.Error(codes.NotFound, "driver not found")
	}
	id := c.nextCertificationID
	c.nextCertificationID++
	cert := &driversproto.CertificationInfo{
		Id:                int64(id),
		DriverId:          req.GetDriverId(),
		VehicleId:         req.GetVehicleId(),
		IdCardFrontUrl:    localCertURL(req.GetDriverId(), "id-card-front", req.GetIdCardFront()),
		IdCardBackUrl:     localCertURL(req.GetDriverId(), "id-card-back", req.GetIdCardBack()),
		DriverLicenseUrl:  localCertURL(req.GetDriverId(), "driver-license", req.GetDriverLicense()),
		VehicleLicenseUrl: localCertURL(req.GetDriverId(), "vehicle-license", req.GetVehicleLicense()),
		AuditStatus:       1,
		AuditRemark:       "",
	}
	c.certByID[id] = cloneCertification(cert)
	c.certByDriver[uint64(req.GetDriverId())] = id
	return &driversproto.UploadCertificationResponse{Id: int64(id), Certification: cloneCertification(cert)}, nil
}

// GetCertification 查询司机资质。
func (c *LocalClient) GetCertification(_ context.Context, req *driversproto.GetCertificationRequest) (*driversproto.GetCertificationResponse, error) {
	if req == nil || req.GetDriverId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "driver id is required")
	}
	c.mu.RLock()
	defer c.mu.RUnlock()

	certID, ok := c.certByDriver[uint64(req.GetDriverId())]
	if !ok {
		return &driversproto.GetCertificationResponse{Found: false}, nil
	}
	return &driversproto.GetCertificationResponse{Certification: cloneCertification(c.certByID[certID]), Found: true}, nil
}

// ApproveCertification 审核通过司机资质。
func (c *LocalClient) ApproveCertification(_ context.Context, req *driversproto.AuditCertificationRequest) (*driversproto.CommonResponse, error) {
	if req == nil || req.GetCertificationId() <= 0 || req.GetOperatorId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "certification id and operator id are required")
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	cert, ok := c.certByID[uint64(req.GetCertificationId())]
	if !ok {
		return nil, status.Error(codes.NotFound, "driver certification not found")
	}
	cert.AuditStatus = 2
	cert.AuditRemark = req.GetRemark()
	c.certByID[uint64(req.GetCertificationId())] = cloneCertification(cert)
	if driverID := cert.GetDriverId(); driverID > 0 {
		if driver, ok := c.drivers[uint64(driverID)]; ok {
			driver.Status = driversproto.DriverStatus_DRIVER_STATUS_NORMAL
			c.drivers[uint64(driverID)] = cloneDriver(driver)
		}
	}
	if cert.GetVehicleId() > 0 {
		if vehicle, ok := c.vehicles[uint64(cert.GetVehicleId())]; ok {
			vehicle.Status = driversproto.VehicleStatus_VEHICLE_STATUS_NORMAL
			c.vehicles[uint64(cert.GetVehicleId())] = cloneVehicle(vehicle)
		}
	}
	return &driversproto.CommonResponse{Message: "ok"}, nil
}

// RejectCertification 驳回司机资质审核。
func (c *LocalClient) RejectCertification(_ context.Context, req *driversproto.AuditCertificationRequest) (*driversproto.CommonResponse, error) {
	if req == nil || req.GetCertificationId() <= 0 || req.GetOperatorId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "certification id and operator id are required")
	}
	if strings.TrimSpace(req.GetRemark()) == "" {
		return nil, status.Error(codes.InvalidArgument, "reject remark is required")
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	cert, ok := c.certByID[uint64(req.GetCertificationId())]
	if !ok {
		return nil, status.Error(codes.NotFound, "driver certification not found")
	}
	cert.AuditStatus = 3
	cert.AuditRemark = req.GetRemark()
	c.certByID[uint64(req.GetCertificationId())] = cloneCertification(cert)
	return &driversproto.CommonResponse{Message: "ok"}, nil
}

func cloneDriver(driver *driversproto.Driver) *driversproto.Driver {
	if driver == nil {
		return nil
	}
	return proto.Clone(driver).(*driversproto.Driver)
}

func cloneVehicle(vehicle *driversproto.Vehicle) *driversproto.Vehicle {
	if vehicle == nil {
		return nil
	}
	return proto.Clone(vehicle).(*driversproto.Vehicle)
}

func cloneCertification(cert *driversproto.CertificationInfo) *driversproto.CertificationInfo {
	if cert == nil {
		return nil
	}
	return proto.Clone(cert).(*driversproto.CertificationInfo)
}

func localCertURL(driverID int64, kind, payload string) string {
	if strings.TrimSpace(payload) == "" {
		return ""
	}
	return fmt.Sprintf("local://driver/%d/%s", driverID, kind)
}

func haversineMeters(lon1, lat1, lon2, lat2 float64) float64 {
	const earthRadius = 6371000.0
	toRad := func(v float64) float64 { return v * math.Pi / 180 }
	dLat := toRad(lat2 - lat1)
	dLon := toRad(lon2 - lon1)
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(toRad(lat1))*math.Cos(toRad(lat2))*math.Sin(dLon/2)*math.Sin(dLon/2)
	return 2 * earthRadius * math.Asin(math.Sqrt(a))
}
