package logic

import (
	"context"

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
