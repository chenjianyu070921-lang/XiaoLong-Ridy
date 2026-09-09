package logic

import (
	"context"
	"errors"

	"XiaoLong-Ridy/rpc/driversvc/internal/repository"
	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	"XiaoLong-Ridy/rpc/driversvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

// GetCertificationLogic 司机资质查询业务逻辑。
// 作用：按司机 ID 返回其资质记录与审核状态，供司机端展示上传进度与审核结果。
type GetCertificationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewGetCertificationLogic 构造司机资质查询逻辑处理器。
func NewGetCertificationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCertificationLogic {
	return &GetCertificationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetCertification 按司机 ID 查询资质记录；无记录时返回 found=false。
func (l *GetCertificationLogic) GetCertification(in *proto.GetCertificationRequest) (*proto.GetCertificationResponse, error) {
	if in == nil || in.DriverId <= 0 {
		return nil, errInvalidDriverID
	}
	if l.svcCtx == nil || l.svcCtx.CertificationRepository == nil {
		return nil, errors.New("certification repository not ready")
	}
	cert, err := l.svcCtx.CertificationRepository.GetByDriverID(l.ctx, uint64(in.DriverId))
	if err != nil {
		// 无记录不算错误，返回 found=false 让前端提示"尚未上传资质"。
		if errors.Is(err, repository.ErrCertificationNotFound) {
			return &proto.GetCertificationResponse{Found: false}, nil
		}
		return nil, err
	}
	return &proto.GetCertificationResponse{
		Certification: toCertificationInfo(cert),
		Found:         true,
	}, nil
}
