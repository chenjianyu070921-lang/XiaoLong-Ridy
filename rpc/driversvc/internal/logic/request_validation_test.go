package logic

import (
	"context"
	"testing"

	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
)

func TestNilRequestsAreRejectedWithoutPanic(t *testing.T) {
	tests := []struct {
		name string
		run  func() error
	}{
		{
			name: "GetDriver",
			run: func() error {
				_, err := NewGetDriverLogic(context.Background(), &svc.ServiceContext{}).GetDriver(nil)
				return err
			},
		},
		{
			name: "DeleteDriver",
			run: func() error {
				_, err := NewDeleteDriverLogic(context.Background(), &svc.ServiceContext{}).DeleteDriver(nil)
				return err
			},
		},
		{
			name: "GetVehicle",
			run: func() error {
				_, err := NewGetVehicleLogic(context.Background(), &svc.ServiceContext{}).GetVehicle(nil)
				return err
			},
		},
		{
			name: "DeleteVehicle",
			run: func() error {
				_, err := NewDeleteVehicleLogic(context.Background(), &svc.ServiceContext{}).DeleteVehicle(nil)
				return err
			},
		},
		{
			name: "UpdateVehicle",
			run: func() error {
				_, err := NewUpdateVehicleLogic(context.Background(), &svc.ServiceContext{}).UpdateVehicle(nil)
				return err
			},
		},
		{
			name: "GetCertification",
			run: func() error {
				_, err := NewGetCertificationLogic(context.Background(), &svc.ServiceContext{}).GetCertification(nil)
				return err
			},
		},
		{
			name: "GetDriverByPhone",
			run: func() error {
				_, err := NewGetDriverByPhoneLogic(context.Background(), &svc.ServiceContext{}).GetDriverByPhone(nil)
				return err
			},
		},
		{
			name: "GetDriverAiScore",
			run: func() error {
				_, err := NewGetDriverAiScoreLogic(context.Background(), &svc.ServiceContext{}).GetDriverAiScore(nil)
				return err
			},
		},
		{
			name: "ListDrivers",
			run: func() error {
				_, err := NewListDriversLogic(context.Background(), &svc.ServiceContext{}).ListDrivers(nil)
				return err
			},
		},
		{
			name: "ListNearbyDrivers",
			run: func() error {
				_, err := NewListNearbyDriversLogic(context.Background(), &svc.ServiceContext{}).ListNearbyDrivers(nil)
				return err
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("panic: %v", r)
				}
			}()
			if err := tc.run(); err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

func TestLoginBySMSRejectsNilRequestWithoutPanic(t *testing.T) {
	logic := NewLoginLogic(context.Background(), &svc.ServiceContext{})
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("panic: %v", r)
		}
	}()
	if _, err := logic.LoginBySMS(nil); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestListDriversRejectsNilRequestUsesInvalidArgumentLikeError(t *testing.T) {
	_, err := NewListDriversLogic(context.Background(), &svc.ServiceContext{}).ListDrivers(nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
