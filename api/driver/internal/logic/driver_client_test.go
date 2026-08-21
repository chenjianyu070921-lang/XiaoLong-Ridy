package logic

import (
	"context"

	driversproto "XiaoLong-Ridy/rpc/driversvc/proto"
)

type fakeDriverClient struct {
	reportLocationRequest    *driversproto.ReportLocationRequest
	serviceStatusRequests    []*driversproto.SetDriverServiceStatusRequest
	reportLocationResponse   *driversproto.ReportLocationResponse
	setServiceStatusResponse *driversproto.SetDriverServiceStatusResponse
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

func (f *fakeDriverClient) GetDriverByPhone(context.Context, *driversproto.GetDriverByPhoneRequest) (*driversproto.GetDriverByPhoneResponse, error) {
	return nil, nil
}

func (f *fakeDriverClient) SetDriverOnline(context.Context, *driversproto.SetDriverOnlineRequest) (*driversproto.SetDriverOnlineResponse, error) {
	return nil, nil
}

func (f *fakeDriverClient) SetDriverOffline(context.Context, *driversproto.SetDriverOfflineRequest) (*driversproto.SetDriverOfflineResponse, error) {
	return nil, nil
}

func (f *fakeDriverClient) Login(context.Context, *driversproto.LoginRequest) (*driversproto.LoginResponse, error) {
	return nil, nil
}

func (f *fakeDriverClient) LoginBySMS(context.Context, *driversproto.LoginBySMSRequest) (*driversproto.LoginResponse, error) {
	return nil, nil
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
