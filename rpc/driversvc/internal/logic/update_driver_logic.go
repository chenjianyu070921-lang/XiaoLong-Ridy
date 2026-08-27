package logic

import (
	"context"
	"errors"
	"strings"

	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	"XiaoLong-Ridy/rpc/driversvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateDriverLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateDriverLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateDriverLogic {
	return &UpdateDriverLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UpdateDriver 更新司机信息，仅修改请求中显式传入的字段（optional 字段为指针，nil 表示不更新）。
func (l *UpdateDriverLogic) UpdateDriver(in *proto.UpdateDriverRequest) (*proto.UpdateDriverResponse, error) {
	if in == nil || in.Id <= 0 {
		return nil, errors.New("司机ID不合法")
	}
	if in.Phone != nil && !driverPhoneRegexp.MatchString(in.GetPhone()) {
		return nil, errors.New("手机号格式不合法")
	}
	if in.PasswordHash != nil {
		if err := validateDriverPasswordHash(in.GetPasswordHash()); err != nil {
			return nil, err
		}
	}
	if in.RealName != nil && strings.TrimSpace(in.GetRealName()) == "" {
		return nil, errors.New("真实姓名不能为空")
	}
	if in.IdCardNo != nil && !driverIDCardRegexp.MatchString(in.GetIdCardNo()) {
		return nil, errors.New("身份证号格式不合法")
	}
	if in.DriverLicenseNo != nil && strings.TrimSpace(in.GetDriverLicenseNo()) == "" {
		return nil, errors.New("驾驶证号不能为空")
	}
	if in.Status != nil && (in.GetStatus() < proto.DriverStatus_DRIVER_STATUS_PENDING || in.GetStatus() > proto.DriverStatus_DRIVER_STATUS_CANCELLED) {
		return nil, errors.New("司机状态不合法")
	}
	d, err := l.svcCtx.DriverRepository.GetByID(l.ctx, uint64(in.Id))
	if err != nil {
		return nil, err
	}

	updates := map[string]interface{}{}
	if in.Phone != nil {
		updates["phone"] = in.GetPhone()
	}
	if in.PasswordHash != nil {
		updates["password_hash"] = in.GetPasswordHash()
	}
	if in.RealName != nil {
		updates["real_name"] = in.GetRealName()
	}
	if in.IdCardNo != nil {
		updates["id_card_no"] = in.GetIdCardNo()
	}
	if in.DriverLicenseNo != nil {
		updates["driver_license_no"] = in.GetDriverLicenseNo()
	}
	if in.AvatarUrl != nil {
		updates["avatar_url"] = in.GetAvatarUrl()
	}
	if in.Status != nil {
		updates["status"] = int8(in.GetStatus())
	}

	if err := l.svcCtx.DriverRepository.Update(l.ctx, uint64(in.Id), updates); err != nil {
		return nil, err
	}
	if d, err = l.svcCtx.DriverRepository.GetByID(l.ctx, uint64(in.Id)); err != nil {
		return nil, err
	}
	return &proto.UpdateDriverResponse{
		Id:        int64(d.Id),
		Status:    proto.DriverStatus(d.Status),
		UpdatedAt: d.UpdatedAt.Unix(),
	}, nil
}
