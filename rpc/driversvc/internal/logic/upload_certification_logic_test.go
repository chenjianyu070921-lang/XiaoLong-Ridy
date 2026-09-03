package logic

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	driverconfig "XiaoLong-Ridy/rpc/driversvc/internal/config"
	"XiaoLong-Ridy/rpc/driversvc/internal/model"
	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	"XiaoLong-Ridy/rpc/driversvc/proto"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestUploadCertificationRequiresVehicleID(t *testing.T) {
	logic := NewUploadCertificationLogic(context.Background(), &svc.ServiceContext{})

	if _, err := logic.UploadCertification(&proto.UploadCertificationRequest{
		DriverId:    25,
		IdCardFront: "not-decoded-before-vehicle-check",
	}); err == nil {
		t.Fatal("UploadCertification() accepted missing vehicle id")
	}
}

func TestUploadCertificationRejectsVehicleOwnershipMismatch(t *testing.T) {
	logic := NewUploadCertificationLogic(context.Background(), &svc.ServiceContext{
		DriverVehicleRepository: &updateVehicleRepository{vehicle: &model.DriverVehicle{
			Id:       77,
			DriverId: 26,
		}},
	})

	if _, err := logic.UploadCertification(&proto.UploadCertificationRequest{
		DriverId:    25,
		VehicleId:   77,
		IdCardFront: "not-decoded-before-ownership-check",
	}); err == nil {
		t.Fatal("UploadCertification() accepted vehicle ownership mismatch")
	} else if st, ok := status.FromError(err); !ok || st.Code() != codes.PermissionDenied {
		t.Fatalf("UploadCertification() error = %v, want PermissionDenied", err)
	}
}

func TestUploadCertificationFallsBackToLocalStorageWhenMinioUnavailable(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DRIVER_CERT_LOCAL_DIR", dir)
	t.Setenv("DRIVER_CERT_PUBLIC_PREFIX", "/api/driver/v1/certification-files")
	certRepo := &uploadCertificationRepository{}
	logic := NewUploadCertificationLogic(context.Background(), &svc.ServiceContext{
		DriverVehicleRepository: &updateVehicleRepository{vehicle: &model.DriverVehicle{
			Id:       77,
			DriverId: 25,
		}},
		CertificationRepository: certRepo,
	})

	resp, err := logic.UploadCertification(&proto.UploadCertificationRequest{
		DriverId:    25,
		VehicleId:   77,
		IdCardFront: "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+/p9sAAAAASUVORK5CYII=",
	})
	if err != nil {
		t.Fatalf("UploadCertification() error = %v", err)
	}
	cert := resp.GetCertification()
	if cert.GetId() != 99 || cert.GetVehicleId() != 77 || cert.GetAuditStatus() != int32(AuditStatusPending) {
		t.Fatalf("UploadCertification() response = %+v", resp)
	}
	if !strings.HasPrefix(cert.GetIdCardFrontUrl(), "/api/driver/v1/certification-files/drivers/25/id_card_front-") ||
		!strings.HasSuffix(cert.GetIdCardFrontUrl(), ".png") {
		t.Fatalf("local certification URL = %q", cert.GetIdCardFrontUrl())
	}
	if len(certRepo.saved) != 1 {
		t.Fatalf("saved certifications = %d", len(certRepo.saved))
	}
	entries, err := os.ReadDir(dir + "/drivers/25")
	if err != nil {
		t.Fatalf("read local certification dir: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("local certification file count = %d", len(entries))
	}
}

func TestUploadCertificationMinioTimeoutDoesNotCancelDatabaseSave(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DRIVER_CERT_LOCAL_DIR", dir)
	stalledMinio := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer stalledMinio.Close()
	minioClient, err := minio.New(strings.TrimPrefix(stalledMinio.URL, "http://"), &minio.Options{
		Creds: credentials.NewStaticV4("access", "secret", ""),
	})
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	certRepo := &uploadCertificationRepository{rejectCanceledContext: true}
	logic := NewUploadCertificationLogic(ctx, &svc.ServiceContext{
		Config: driverconfig.Config{
			Minio: driverconfig.MinioConf{
				Endpoint: strings.TrimPrefix(stalledMinio.URL, "http://"),
				Bucket:   "driver-cert",
			},
		},
		DriverVehicleRepository: &updateVehicleRepository{vehicle: &model.DriverVehicle{
			Id:       77,
			DriverId: 25,
		}},
		CertificationRepository: certRepo,
		MinioClient:             minioClient,
	})

	if _, err := logic.UploadCertification(&proto.UploadCertificationRequest{
		DriverId:    25,
		VehicleId:   77,
		IdCardFront: "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+/p9sAAAAASUVORK5CYII=",
	}); err != nil {
		t.Fatalf("UploadCertification() error = %v", err)
	}
	if len(certRepo.saved) != 1 {
		t.Fatalf("saved certifications = %d", len(certRepo.saved))
	}
}

type uploadCertificationRepository struct {
	saved                 []*model.DriverCertification
	rejectCanceledContext bool
}

func (r *uploadCertificationRepository) Upsert(ctx context.Context, cert *model.DriverCertification) (*model.DriverCertification, error) {
	if r.rejectCanceledContext && ctx.Err() != nil {
		return nil, ctx.Err()
	}
	cert.Id = 99
	cert.CreatedAt = time.Unix(123, 0)
	cert.UpdatedAt = time.Unix(456, 0)
	r.saved = append(r.saved, cert)
	return cert, nil
}

func (r *uploadCertificationRepository) GetByDriverID(ctx context.Context, _ uint64) (*model.DriverCertification, error) {
	if r.rejectCanceledContext && ctx.Err() != nil {
		return nil, ctx.Err()
	}
	return nil, nil
}

func (r *uploadCertificationRepository) UpdateAudit(context.Context, int64, int64, string, int8) error {
	return nil
}
