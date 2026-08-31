package logic

import (
	"context"

	driversproto "XiaoLong-Ridy/rpc/driversvc/proto"
)

type fakeDriverClient struct {
	setOnlineRequest            *driversproto.SetDriverOnlineRequest
	setOfflineRequest           *driversproto.SetDriverOfflineRequest
	reportLocationRequest       *driversproto.ReportLocationRequest
	serviceStatusRequests       []*driversproto.SetDriverServiceStatusRequest
	loginRequest                *driversproto.LoginRequest
	loginBySMSRequest           *driversproto.LoginBySMSRequest
	createVehicleRequest        *driversproto.CreateVehicleRequest
	updateVehicleRequest        *driversproto.UpdateVehicleRequest
	deleteVehicleRequest        *driversproto.DeleteVehicleRequest
	getVehicleRequest           *driversproto.GetVehicleRequest
	getDriverByPhoneRequest     *driversproto.GetDriverByPhoneRequest
	nearbyDriversRequest        *driversproto.ListNearbyDriversRequest
	reportLocationResponse      *driversproto.ReportLocationResponse
	setServiceStatusResponse    *driversproto.SetDriverServiceStatusResponse
	loginResponse               *driversproto.LoginResponse
	loginBySMSResponse          *driversproto.LoginResponse
	createVehicleResponse       *driversproto.CreateVehicleResponse
	updateVehicleResponse       *driversproto.UpdateVehicleResponse
	deleteVehicleResponse       *driversproto.DeleteVehicleResponse
	getVehicleResponse          *driversproto.GetVehicleResponse
	getDriverByPhoneResponse    *driversproto.GetDriverByPhoneResponse
	nearbyDriversResponse       *driversproto.ListNearbyDriversResponse
}

func (f *fakeDriverClient) CreateDriver(context.Context, *driversproto.CreateDriverRequest) (*driversproto.CreateDriverResponse, error) {
	return nil, nil
}

func (f *fakeDriverClient) RegisterDriver(context.Context, *driversproto.CreateDriverRequest) (*driversproto.CreateDriverResponse, error) {
	return nil, nil
}

func (f *fakeDriverClient) UpdateDriver(context.Context, *driversproto.UpdateDriverRequest) (*driversproto.UpdateDriverResponse, error) {
	return nil, nil
}

func (f *fakeDriverClient) GetDriver(context.Context, *driversproto.GetDriverRequest) (*driversproto.GetDriverResponse, error) {
	return nil, nil
}

func (f *fakeDriverClient) SetDriverOnline(_ context.Context, req *driversproto.SetDriverOnlineRequest) (*driversproto.SetDriverOnlineResponse, error) {
	f.setOnlineRequest = req
	return &driversproto.SetDriverOnlineResponse{
		DriverId:     req.GetDriverId(),
		OnlineStatus: 1,
	}, nil
}

func (f *fakeDriverClient) SetDriverOffline(_ context.Context, req *driversproto.SetDriverOfflineRequest) (*driversproto.SetDriverOfflineResponse, error) {
	f.setOfflineRequest = req
	return &driversproto.SetDriverOfflineResponse{
		DriverId:     req.GetDriverId(),
		OnlineStatus: 0,
	}, nil
}

func (f *fakeDriverClient) Login(_ context.Context, req *driversproto.LoginRequest) (*driversproto.LoginResponse, error) {
	f.loginRequest = req
	if f.loginResponse != nil {
		return f.loginResponse, nil
	}
	return &driversproto.LoginResponse{
		Token:    "token",
		ExpireIn: 7200,
		Driver: &driversproto.Driver{
			Id:     25,
			Phone:  "13800000001",
			Status: driversproto.DriverStatus_DRIVER_STATUS_NORMAL,
		},
	}, nil
}

func (f *fakeDriverClient) LoginBySMS(_ context.Context, req *driversproto.LoginBySMSRequest) (*driversproto.LoginResponse, error) {
	f.loginBySMSRequest = req
	if f.loginBySMSResponse != nil {
		return f.loginBySMSResponse, nil
	}
	return &driversproto.LoginResponse{
		Token:    "sms-token",
		ExpireIn: 7200,
		Driver: &driversproto.Driver{
			Id:     25,
			Phone:  "13800000001",
			Status: driversproto.DriverStatus_DRIVER_STATUS_NORMAL,
		},
	}, nil
}

func (f *fakeDriverClient) DeleteDriver(context.Context, *driversproto.DeleteDriverRequest) (*driversproto.DeleteDriverResponse, error) {
	return nil, nil
}

func (f *fakeDriverClient) GetDriverAiScore(context.Context, *driversproto.GetDriverAiScoreRequest) (*driversproto.GetDriverAiScoreResponse, error) {
	return nil, nil
}

func (f *fakeDriverClient) UploadCertification(context.Context, *driversproto.UploadCertificationRequest) (*driversproto.UploadCertificationResponse, error) {
	return nil, nil
}

func (f *fakeDriverClient) GetCertification(context.Context, *driversproto.GetCertificationRequest) (*driversproto.GetCertificationResponse, error) {
	return nil, nil
}

func (f *fakeDriverClient) CreateWithdraw(context.Context, *driversproto.CreateWithdrawRequest) (*driversproto.CreateWithdrawResponse, error) {
	return nil, nil
}

func (f *fakeDriverClient) ListWithdraws(context.Context, *driversproto.ListWithdrawsRequest) (*driversproto.ListWithdrawsResponse, error) {
	return nil, nil
}

func (f *fakeDriverClient) CreateVehicle(_ context.Context, req *driversproto.CreateVehicleRequest) (*driversproto.CreateVehicleResponse, error) {
	f.createVehicleRequest = req
	if f.createVehicleResponse != nil {
		return f.createVehicleResponse, nil
	}
	return &driversproto.CreateVehicleResponse{
		Id:        77,
		Status:    driversproto.VehicleStatus_VEHICLE_STATUS_PENDING,
		CreatedAt: 123,
	}, nil
}

func (f *fakeDriverClient) UpdateVehicle(_ context.Context, req *driversproto.UpdateVehicleRequest) (*driversproto.UpdateVehicleResponse, error) {
	f.updateVehicleRequest = req
	if f.updateVehicleResponse != nil {
		return f.updateVehicleResponse, nil
	}
	return &driversproto.UpdateVehicleResponse{
		Id:        req.GetId(),
		Status:    driversproto.VehicleStatus_VEHICLE_STATUS_NORMAL,
		UpdatedAt: 456,
	}, nil
}

func (f *fakeDriverClient) DeleteVehicle(_ context.Context, req *driversproto.DeleteVehicleRequest) (*driversproto.DeleteVehicleResponse, error) {
	f.deleteVehicleRequest = req
	if f.deleteVehicleResponse != nil {
		return f.deleteVehicleResponse, nil
	}
	return &driversproto.DeleteVehicleResponse{Id: req.GetId(), Success: true}, nil
}

func (f *fakeDriverClient) GetVehicle(_ context.Context, req *driversproto.GetVehicleRequest) (*driversproto.GetVehicleResponse, error) {
	f.getVehicleRequest = req
	if f.getVehicleResponse != nil {
		return f.getVehicleResponse, nil
	}
	return &driversproto.GetVehicleResponse{
		Vehicle: &driversproto.Vehicle{
			Id:       req.GetId(),
			DriverId: 25,
			PlateNo:  "粤B12345",
			Brand:    "BYD",
			Model:    "Han",
			Status:   driversproto.VehicleStatus_VEHICLE_STATUS_PENDING,
		},
	}, nil
}

func (f *fakeDriverClient) GetDriverByPhone(_ context.Context, req *driversproto.GetDriverByPhoneRequest) (*driversproto.GetDriverByPhoneResponse, error) {
	f.getDriverByPhoneRequest = req
	if f.getDriverByPhoneResponse != nil {
		return f.getDriverByPhoneResponse, nil
	}
	return &driversproto.GetDriverByPhoneResponse{
		Driver: &driversproto.Driver{
			Id:     25,
			Phone:  req.GetPhone(),
			Status: driversproto.DriverStatus_DRIVER_STATUS_NORMAL,
		},
	}, nil
}

func (f *fakeDriverClient) ListNearbyDrivers(_ context.Context, req *driversproto.ListNearbyDriversRequest) (*driversproto.ListNearbyDriversResponse, error) {
	f.nearbyDriversRequest = req
	if f.nearbyDriversResponse != nil {
		return f.nearbyDriversResponse, nil
	}
	return &driversproto.ListNearbyDriversResponse{
		Drivers: []*driversproto.NearbyDriver{{
			DriverId:       25,
			Longitude:      req.GetLongitude(),
			Latitude:       req.GetLatitude(),
			DistanceMeters: 1200,
		}},
	}, nil
}

func (f *fakeDriverClient) Heartbeat(context.Context, *driversproto.HeartbeatRequest) (*driversproto.HeartbeatResponse, error) {
	return nil, nil
}

func (f *fakeDriverClient) ReportLocation(_ context.Context, req *driversproto.ReportLocationRequest) (*driversproto.ReportLocationResponse, error) {
	f.reportLocationRequest = req
	if f.reportLocationResponse != nil {
		return f.reportLocationResponse, nil
	}
	return &driversproto.ReportLocationResponse{
		DriverId:     req.GetDriverId(),
		OnlineStatus: 1,
		ReportTime:   123,
	}, nil
}

func (f *fakeDriverClient) SetDriverServiceStatus(_ context.Context, req *driversproto.SetDriverServiceStatusRequest) (*driversproto.SetDriverServiceStatusResponse, error) {
	f.serviceStatusRequests = append(f.serviceStatusRequests, req)
	if f.setServiceStatusResponse != nil {
		return f.setServiceStatusResponse, nil
	}
	return &driversproto.SetDriverServiceStatusResponse{
		DriverId:     req.GetDriverId(),
		OnlineStatus: req.GetOnlineStatus(),
		UpdatedAt:    123,
	}, nil
}
