package logic

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"XiaoLong-Ridy/rpc/driversvc/internal/model"
	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	"XiaoLong-Ridy/rpc/driversvc/proto"

	"github.com/minio/minio-go/v7"
	"github.com/zeromicro/go-zero/core/logx"
)

// 资质图片类型校验：仅允许常见图片扩展名。
var (
	// errEmptyCertification 表示未提供任何资质图片。
	errEmptyCertification = errors.New("至少上传一张资质图片")
	// errInvalidImage 表示图片 base64 解码失败或文件过大。
	errInvalidImage = errors.New("资质图片数据非法")
)

// maxCertImageBytes 单张资质图片大小上限（5MB）。
const maxCertImageBytes = 5 * 1024 * 1024

// UploadCertificationLogic 司机资质上传业务逻辑。
// 作用：将司机上传的身份证/驾驶证/行驶证图片（base64）直传 MinIO，并落库 driver_certification 表。
type UploadCertificationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewUploadCertificationLogic 构造司机资质上传逻辑处理器。
func NewUploadCertificationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UploadCertificationLogic {
	return &UploadCertificationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UploadCertification 处理资质上传：解码校验图片、上传 MinIO、落库并返回访问 URL。
func (l *UploadCertificationLogic) UploadCertification(in *proto.UploadCertificationRequest) (*proto.UploadCertificationResponse, error) {
	// 校验司机 ID 与至少一张图片。
	if in.DriverId <= 0 {
		return nil, errInvalidDriverID
	}
	if in.IdCardFront == "" && in.IdCardBack == "" && in.DriverLicense == "" && in.VehicleLicense == "" {
		return nil, errEmptyCertification
	}

	// 逐类上传图片到 MinIO，得到访问 URL。
	idCardFrontURL, err := l.uploadImage(in.DriverId, "id_card_front", in.IdCardFront)
	if err != nil {
		return nil, err
	}
	idCardBackURL, err := l.uploadImage(in.DriverId, "id_card_back", in.IdCardBack)
	if err != nil {
		return nil, err
	}
	driverLicenseURL, err := l.uploadImage(in.DriverId, "driver_license", in.DriverLicense)
	if err != nil {
		return nil, err
	}
	vehicleLicenseURL, err := l.uploadImage(in.DriverId, "vehicle_license", in.VehicleLicense)
	if err != nil {
		return nil, err
	}

	// 读取已有资质记录（无则新建），保留已上传的其他字段。
	var existing *model.DriverCertification
	if ex, err := l.svcCtx.CertificationRepository.GetByDriverID(l.ctx, uint64(in.DriverId)); err == nil {
		existing = ex
	}
	cert := &model.DriverCertification{
		DriverId:  uint64(in.DriverId),
		VehicleId: uint64(in.VehicleId),
	}
	if existing != nil {
		cert.IdCardFrontUrl = existing.IdCardFrontUrl
		cert.IdCardBackUrl = existing.IdCardBackUrl
		cert.DriverLicenseUrl = existing.DriverLicenseUrl
		cert.VehicleLicenseUrl = existing.VehicleLicenseUrl
	}
	if idCardFrontURL != "" {
		cert.IdCardFrontUrl = idCardFrontURL
	}
	if idCardBackURL != "" {
		cert.IdCardBackUrl = idCardBackURL
	}
	if driverLicenseURL != "" {
		cert.DriverLicenseUrl = driverLicenseURL
	}
	if vehicleLicenseURL != "" {
		cert.VehicleLicenseUrl = vehicleLicenseURL
	}
	// 有新图片上传则置为待审核。
	cert.AuditStatus = AuditStatusPending

	saved, err := l.svcCtx.CertificationRepository.Upsert(l.ctx, cert)
	if err != nil {
		return nil, err
	}

	return &proto.UploadCertificationResponse{
		Id: int64(saved.Id),
		Certification: toCertificationInfo(saved),
	}, nil
}

// uploadImage 将一张 base64 图片上传到 MinIO，返回公开访问 URL；base64 为空时返回空串。
func (l *UploadCertificationLogic) uploadImage(driverID int64, kind, base64Data string) (string, error) {
	if base64Data == "" {
		return "", nil
	}
	// 解码 base64 并做类型/大小校验。
	data, err := decodeImage(base64Data)
	if err != nil {
		return "", err
	}
	// 生成对象键：drivers/{driverID}/{kind}-{unix}.ext。
	ext := guessImageExt(base64Data)
	objectKey := fmt.Sprintf("drivers/%d/%s-%d%s", driverID, kind, time.Now().UnixNano(), ext)
	// 上传到 MinIO（服务端直传），内容类型固定为 image。
	_, err = l.svcCtx.MinioClient.PutObject(l.ctx, l.svcCtx.Config.Minio.Bucket, objectKey, bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{ContentType: "image/" + strings.TrimPrefix(ext, ".")})
	if err != nil {
		return "", err
	}
	// 返回可访问 URL：endpoint/bucket/objectKey。
	return fmt.Sprintf("%s/%s/%s", l.svcCtx.Config.Minio.Endpoint, l.svcCtx.Config.Minio.Bucket, objectKey), nil
}
