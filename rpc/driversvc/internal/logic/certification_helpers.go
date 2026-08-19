package logic

import (
	"encoding/base64"
	"strings"

	"XiaoLong-Ridy/rpc/driversvc/internal/model"
	"XiaoLong-Ridy/rpc/driversvc/proto"
)

// 资质审核状态常量（对齐 driver_certification.audit_status）。
const (
	// AuditStatusPending 待审核。
	AuditStatusPending int8 = 1
	// AuditStatusPassed 审核通过。
	AuditStatusPassed int8 = 2
	// AuditStatusRejected 审核驳回。
	AuditStatusRejected int8 = 3
)

// decodeImage 校验并解码 base64 图片数据。
// 支持可选 data URI 前缀（data:image/png;base64,...），解码后校验大小上限。
func decodeImage(b64 string) ([]byte, error) {
	// 去除 data URI 前缀（如果存在）。
	if idx := strings.Index(b64, ","); strings.HasPrefix(b64, "data:") && idx >= 0 {
		b64 = b64[idx+1:]
	}
	data, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return nil, errInvalidImage
	}
	if len(data) == 0 || len(data) > maxCertImageBytes {
		return nil, errInvalidImage
	}
	return data, nil
}

// guessImageExt 根据 data URI 或 base64 头部魔数猜测图片扩展名，未知时默认 .jpg。
func guessImageExt(b64 string) string {
	lower := strings.ToLower(b64)
	// 优先按 data URI 声明判断。
	if strings.HasPrefix(lower, "data:image/png") {
		return ".png"
	}
	if strings.HasPrefix(lower, "data:image/jpeg") || strings.HasPrefix(lower, "data:image/jpg") {
		return ".jpg"
	}
	// 无 data URI 时，用 base64 头魔数推断：png 以 "iVBOR" 开头，jpeg 以 "/9j/" 开头。
	head := lower
	if len(head) > 8 {
		head = head[:8]
	}
	if strings.HasPrefix(head, "ivbor") {
		return ".png"
	}
	if strings.HasPrefix(head, "/9j/") {
		return ".jpg"
	}
	return ".jpg"
}

// toCertificationInfo 将 model 资质记录映射为 proto 响应结构。
func toCertificationInfo(c *model.DriverCertification) *proto.CertificationInfo {
	if c == nil {
		return nil
	}
	return &proto.CertificationInfo{
		Id:                 int64(c.Id),
		DriverId:           int64(c.DriverId),
		VehicleId:          int64(c.VehicleId),
		IdCardFrontUrl:     c.IdCardFrontUrl,
		IdCardBackUrl:      c.IdCardBackUrl,
		DriverLicenseUrl:   c.DriverLicenseUrl,
		VehicleLicenseUrl:  c.VehicleLicenseUrl,
		AuditStatus:        int32(c.AuditStatus),
		AuditRemark:        c.AuditRemark,
	}
}
