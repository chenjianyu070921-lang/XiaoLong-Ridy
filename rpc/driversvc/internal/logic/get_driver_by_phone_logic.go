package logic

import (
	"context"

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
	d, err := l.svcCtx.DriverRepository.GetByPhone(l.ctx, in.Phone)
	if err != nil {
		return nil, err
	}
	return &proto.GetDriverByPhoneResponse{
		Driver: &proto.Driver{
			Id:              int64(d.Id),
			Phone:           d.Phone,
			PasswordHash:    "", // 不返回密码哈希，避免凭据泄露
			RealName:        d.RealName,
			IdCardNo:        maskIDCard(d.IdCardNo), // 身份证脱敏返回
			DriverLicenseNo: d.DriverLicenseNo,
			AvatarUrl:       d.AvatarUrl,
			Status:          proto.DriverStatus(d.Status),
			CreatedAt:       d.CreatedAt.Unix(),
			UpdatedAt:       d.UpdatedAt.Unix(),
		},
	}, nil
}

// maskIDCard 对身份证号做脱敏：保留前 4 位与后 2 位，中间以 * 替代。
func maskIDCard(no string) string {
	if len(no) <= 6 {
		return no
	}
	return no[:4] + "**********" + no[len(no)-2:]
}
