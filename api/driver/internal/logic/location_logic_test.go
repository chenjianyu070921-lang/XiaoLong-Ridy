package logic

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"

	"XiaoLong-Ridy/api/driver/internal/svc"
	"XiaoLong-Ridy/api/driver/internal/types"
	driversproto "XiaoLong-Ridy/rpc/driversvc/proto"
)

func TestReportLocationForwardsCurrentDriverAndLocation(t *testing.T) {
	driverClient := &fakeDriverClient{}
	logic := NewLocationLogic(context.Background(), &svc.ServiceContext{DriverClient: driverClient})

	resp, err := logic.ReportLocation(25, &types.ReportLocationRequest{
		DeviceID:  "device-1",
		Longitude: 116.397,
		Latitude:  39.908,
	})
	if err != nil {
		t.Fatalf("ReportLocation() error = %v", err)
	}
	if resp.DriverID != 25 || resp.OnlineStatus != 1 || resp.ReportTime != 123 {
		t.Fatalf("ReportLocation() response = %+v", resp)
	}
	req := driverClient.reportLocationRequest
	if req.GetDriverId() != 25 || req.GetDeviceId() != "device-1" ||
		req.GetLongitude() != 116.397 || req.GetLatitude() != 39.908 {
		t.Fatalf("ReportLocation() request = %+v", req)
	}
}

func TestReportLocationRejectsInvalidInput(t *testing.T) {
	logic := NewLocationLogic(context.Background(), &svc.ServiceContext{DriverClient: &fakeDriverClient{}})

	cases := []struct {
		name     string
		driverID int64
		req      *types.ReportLocationRequest
	}{
		{name: "missing driver", driverID: 0, req: &types.ReportLocationRequest{DeviceID: "d", Longitude: 1, Latitude: 1}},
		{name: "missing device", driverID: 25, req: &types.ReportLocationRequest{Longitude: 1, Latitude: 1}},
		{name: "bad longitude", driverID: 25, req: &types.ReportLocationRequest{DeviceID: "d", Longitude: 181, Latitude: 1}},
		{name: "bad latitude", driverID: 25, req: &types.ReportLocationRequest{DeviceID: "d", Longitude: 1, Latitude: -91}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := logic.ReportLocation(tc.driverID, tc.req); err != ErrInvalidParam {
				t.Fatalf("ReportLocation() error = %v, want %v", err, ErrInvalidParam)
			}
		})
	}
}

func TestReportLocationParallelStress(t *testing.T) {
	client := &benchmarkDriverClient{}
	logic := NewLocationLogic(context.Background(), &svc.ServiceContext{DriverClient: client})
	req := &types.ReportLocationRequest{
		DeviceID:  "device-stress",
		Longitude: 116.397,
		Latitude:  39.908,
	}

	const goroutines = 64
	const callsPerGoroutine = 100
	var wg sync.WaitGroup
	errCh := make(chan error, goroutines*callsPerGoroutine)

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(driverID int64) {
			defer wg.Done()
			for j := 0; j < callsPerGoroutine; j++ {
				resp, err := logic.ReportLocation(driverID, req)
				if err != nil {
					errCh <- err
					continue
				}
				if resp.DriverID != driverID || resp.OnlineStatus != 1 {
					errCh <- ErrInvalidParam
				}
			}
		}(int64(i + 1))
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Fatalf("ReportLocation() parallel stress error = %v", err)
	}
	if got, want := client.calls.Load(), int64(goroutines*callsPerGoroutine); got != want {
		t.Fatalf("ReportLocation() calls = %d, want %d", got, want)
	}
}

func BenchmarkReportLocation(b *testing.B) {
	client := &benchmarkDriverClient{}
	logic := NewLocationLogic(context.Background(), &svc.ServiceContext{DriverClient: client})
	req := &types.ReportLocationRequest{
		DeviceID:  "device-benchmark",
		Longitude: 116.397,
		Latitude:  39.908,
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := logic.ReportLocation(25, req); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkReportLocationParallel(b *testing.B) {
	client := &benchmarkDriverClient{}
	logic := NewLocationLogic(context.Background(), &svc.ServiceContext{DriverClient: client})
	req := &types.ReportLocationRequest{
		DeviceID:  "device-benchmark",
		Longitude: 116.397,
		Latitude:  39.908,
	}

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if _, err := logic.ReportLocation(25, req); err != nil {
				b.Fatal(err)
			}
		}
	})
}

type benchmarkDriverClient struct {
	fakeDriverClient
	calls atomic.Int64
}

func (f *benchmarkDriverClient) ReportLocation(_ context.Context, req *driversproto.ReportLocationRequest) (*driversproto.ReportLocationResponse, error) {
	f.calls.Add(1)
	return &driversproto.ReportLocationResponse{
		DriverId:     req.GetDriverId(),
		OnlineStatus: 1,
		ReportTime:   123,
	}, nil
}
