package logic

import (
	"context"

	"XiaoLong-Ridy/rpc/driversvc/internal/model"
	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	"XiaoLong-Ridy/rpc/driversvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

// CreateDriverLogic 处理创建司机请求的逻辑结构体。
type CreateDriverLogic struct {
	ctx    context.Context       // ctx：请求上下文
	svcCtx *svc.ServiceContext   // svcCtx：服务上下文，持有 DB 等依赖
	logx.Logger
}

// NewCreateDriverLogic 构造 CreateDriverLogic 实例。
func NewCreateDriverLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateDriverLogic {
	return &CreateDriverLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// CreateDriver 创建司机账号，状态初始为待审核（PENDING）。
func (l *CreateDriverLogic) CreateDriver(in *proto.CreateDriverRequest) (*proto.CreateDriverResponse, error) {
	// 组装司机记录，状态默认置为待审核(1)
	d := &model.Driver{
		Phone:           in.Phone,           // 手机号（登录账号）
		PasswordHash:    in.PasswordHash,    // 密码哈希
		RealName:        in.RealName,        // 真实姓名
		IdCardNo:        in.IdCardNo,        // 身份证号
		DriverLicenseNo: in.DriverLicenseNo, // 驾驶证号
		AvatarUrl:       in.AvatarUrl,       // 头像地址
		Status:          int8(proto.DriverStatus_DRIVER_STATUS_PENDING), // 初始状态：待审核
	}
	// 写入数据库
	if err := l.svcCtx.DB.Create(d).Error; err != nil {
		return nil, err
	}
	// 返回新建司机的 ID、状态与创建时间
	return &proto.CreateDriverResponse{
		Id:        int64(d.Id),
		Status:    proto.DriverStatus_DRIVER_STATUS_PENDING,
		CreatedAt: d.CreatedAt.Unix(),
	}, nil
}
