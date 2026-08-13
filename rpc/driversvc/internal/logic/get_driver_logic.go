package logic

import (
	"context"
	"errors"

	"XiaoLong-Ridy/rpc/driversvc/internal/model"
	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	"XiaoLong-Ridy/rpc/driversvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

// GetDriverLogic 处理查询司机详情请求的逻辑结构体。
type GetDriverLogic struct {
	ctx    context.Context      // ctx：请求上下文
	svcCtx *svc.ServiceContext  // svcCtx：服务上下文，持有 DB 等依赖
	logx.Logger
}

// NewGetDriverLogic 构造 GetDriverLogic 实例。
func NewGetDriverLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDriverLogic {
	return &GetDriverLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetDriver 根据司机 ID 查询司机完整信息。
func (l *GetDriverLogic) GetDriver(in *proto.GetDriverRequest) (*proto.GetDriverResponse, error) {
	// 按 ID 查询司机
	var d model.Driver
	err := l.svcCtx.DB.First(&d, in.Id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("driver not found") // 司机不存在
	}
	if err != nil {
		return nil, err
	}
	// 组装并返回司机详情
	return &proto.GetDriverResponse{
		Driver: &proto.Driver{
			Id:              int64(d.Id),              // 司机 ID
			Phone:           d.Phone,                 // 手机号
			PasswordHash:    d.PasswordHash,          // 密码哈希
			RealName:        d.RealName,              // 真实姓名
			IdCardNo:        d.IdCardNo,              // 身份证号
			DriverLicenseNo: d.DriverLicenseNo,       // 驾驶证号
			AvatarUrl:       d.AvatarUrl,             // 头像地址
			Status:          proto.DriverStatus(d.Status), // 账号状态
			CreatedAt:       d.CreatedAt.Unix(),      // 创建时间（Unix 秒）
			UpdatedAt:       d.UpdatedAt.Unix(),      // 更新时间（Unix 秒）
		},
	}, nil
}
