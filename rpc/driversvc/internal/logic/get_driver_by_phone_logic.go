package logic

import (
	"context"
	"errors"
	"strings"

	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	"XiaoLong-Ridy/rpc/driversvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetDriverByPhoneLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetDriverByPhoneLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDriverByPhoneLogic {
	return &GetDriverByPhoneLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetDriverByPhone 根据手机号查询司机完整信息（登录场景使用）。
func (l *GetDriverByPhoneLogic) GetDriverByPhone(in *proto.GetDriverByPhoneRequest) (*proto.GetDriverByPhoneResponse, error) {
	if in == nil || strings.TrimSpace(in.Phone) == "" {
		return nil, errors.New("手机号不合法")
	}
	if l.svcCtx == nil || l.svcCtx.DriverRepository == nil {
		return nil, errors.New("driver repository not ready")
	}
	d, err := l.svcCtx.DriverRepository.GetByPhone(l.ctx, in.Phone)
	if err != nil {
		return nil, err
	}
	return &proto.GetDriverByPhoneResponse{
		Driver: &proto.Driver{
			Id:              int64(d.Id),
			Phone:           d.Phone,
			PasswordHash:    d.PasswordHash,
			RealName:        d.RealName,
			IdCardNo:        d.IdCardNo,
			DriverLicenseNo: d.DriverLicenseNo,
			AvatarUrl:       d.AvatarUrl,
			Status:          proto.DriverStatus(d.Status),
			CreatedAt:       d.CreatedAt.Unix(),
			UpdatedAt:       d.UpdatedAt.Unix(),
		},
	}, nil
}
