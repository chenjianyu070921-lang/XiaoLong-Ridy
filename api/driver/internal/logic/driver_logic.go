// Package logic 实现 driver API 的司机业务逻辑层。
package logic

import (
	"context" // 用于在不同层之间传递请求上下文
	"errors"  // 用于返回业务校验错误

	"XiaoLong-Ridy/api/driver/internal/svc"          // 服务上下文，提供 driversvc 客户端
	"XiaoLong-Ridy/api/driver/internal/types"        // API 层使用的请求/响应类型
	"XiaoLong-Ridy/common/cryptox"                   // 密码哈希工具
	"XiaoLong-Ridy/common/jwtx"                      // 手机号脱敏工具
	driversproto "XiaoLong-Ridy/rpc/driversvc/proto" // driversvc 的 gRPC 请求/响应类型
)

// DriverLogic 司机业务逻辑处理器，持有请求上下文与下游 driversvc 客户端，负责司机增删改查的编排与参数校验。
type DriverLogic struct {
	ctx    context.Context     // 当前请求上下文
	svcCtx *svc.ServiceContext // 全局服务上下文（含 driversvc 客户端）
}

// NewDriverLogic 构造司机逻辑处理器实例，注入请求上下文与服务上下文。
func NewDriverLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DriverLogic {
	// 注入上下文与服务上下文。
	return &DriverLogic{ctx: ctx, svcCtx: svcCtx}
}

// enumDriverStatus 将可选状态字符串（如 "DRIVER_STATUS_NORMAL"）转为 proto 枚举指针；
// nil 或空串返回 nil（表示不更新该字段），未知值映射为 UNSPECIFIED 由底层忽略。
func enumDriverStatus(s *string) *driversproto.DriverStatus {
	// 入参为 nil 或空串时返回 nil，调用方据此跳过该可选字段。
	if s == nil || *s == "" {
		return nil
	}
	// 声明局部变量承载映射后的枚举值。
	var v driversproto.DriverStatus
	// 按字符串值映射到对应的 proto 枚举。
	switch *s {
	case "DRIVER_STATUS_PENDING": // 待审核
		v = driversproto.DriverStatus_DRIVER_STATUS_PENDING
	case "DRIVER_STATUS_NORMAL": // 正常
		v = driversproto.DriverStatus_DRIVER_STATUS_NORMAL
	case "DRIVER_STATUS_FROZEN": // 冻结
		v = driversproto.DriverStatus_DRIVER_STATUS_FROZEN
	case "DRIVER_STATUS_CANCELLED": // 注销
		v = driversproto.DriverStatus_DRIVER_STATUS_CANCELLED
	default: // 未知值映射为未指定（由底层忽略）。
		v = driversproto.DriverStatus_DRIVER_STATUS_UNSPECIFIED
	}
	// 返回枚举值的指针，以匹配 proto 的 optional 字段语义。
	return &v
}

// RegisterDriver 司机自注册：在 API 层校验手机号/姓名/身份证/驾驶证，
// 校验通过后调用 driversvc.RegisterDriver，返回新建司机 ID 与状态。
func (l *DriverLogic) RegisterDriver(req *types.RegisterDriverRequest) (*types.RegisterDriverResponse, error) {
	if req == nil {
		return nil, errors.New("请求参数不能为空")
	}
	normalizeRegisterDriverRequest(req)
	// 校验手机号格式。
	if !validPhone(req.Phone) {
		return nil, errors.New("手机号格式不合法")
	}
	if !validPassword(req.Password) {
		return nil, errors.New("密码长度必须为8到72字节")
	}
	// 校验真实姓名非空。
	if req.RealName == "" {
		return nil, errors.New("真实姓名不能为空")
	}
	// 校验身份证号格式。
	if !validIDCard(req.IdCardNo) {
		return nil, errors.New("身份证号格式不合法")
	}
	// 校验驾驶证号非空。
	if req.DriverLicenseNo == "" {
		return nil, errors.New("驾驶证号不能为空")
	}
	// 获取 driversvc 客户端（可能为配置错误）。
	client, err := l.driverClient()
	if err != nil {
		return nil, err
	}
	passwordHash, err := cryptox.BcryptHash(req.Password)
	if err != nil {
		return nil, err
	}
	// 调用下游独立的注册接口，并将 API 入参映射为 proto 请求。
	resp, err := client.RegisterDriver(l.ctx, &driversproto.CreateDriverRequest{
		Phone:           req.Phone,
		PasswordHash:    passwordHash,
		RealName:        req.RealName,
		IdCardNo:        req.IdCardNo,
		DriverLicenseNo: req.DriverLicenseNo,
		AvatarUrl:       req.AvatarURL,
	})
	if err != nil {
		return nil, err
	}
	return &types.RegisterDriverResponse{ID: resp.GetId(), Status: resp.GetStatus().String(), CreatedAt: resp.GetCreatedAt()}, nil
}

// UpdateDriver 更新司机：校验司机 ID、可选身份字段和密码，密码先在 API 层生成 bcrypt 哈希，再调用 driversvc 更新。
func (l *DriverLogic) UpdateDriver(req *types.UpdateDriverRequest) (*types.UpdateDriverResponse, error) {
	if req == nil {
		return nil, errors.New("请求参数不能为空")
	}
	// 校验司机 ID 合法性。
	if req.ID <= 0 {
		return nil, errors.New("司机ID不合法")
	}
	// 若传入手机号，则校验其格式。
	if req.Phone != nil && !validPhone(*req.Phone) {
		return nil, errors.New("手机号格式不合法")
	}
	// 若传入身份证号，则校验其格式。
	if req.IdCardNo != nil && !validIDCard(*req.IdCardNo) {
		return nil, errors.New("身份证号格式不合法")
	}
	if req.Status != nil && !validDriverStatus(*req.Status) {
		return nil, errors.New("司机状态不合法")
	}
	var passwordHash *string
	if req.Password != nil {
		if !validPassword(*req.Password) {
			return nil, errors.New("密码长度必须为8到72字节")
		}
		hash, err := cryptox.BcryptHash(*req.Password)
		if err != nil {
			return nil, err
		}
		passwordHash = &hash
	}
	// 获取 driversvc 客户端。
	client, err := l.driverClient()
	if err != nil {
		return nil, err
	}
	// 调用下游更新接口；可选字段直接透传指针，状态经枚举转换。
	resp, err := client.UpdateDriver(l.ctx, &driversproto.UpdateDriverRequest{
		Id:              req.ID,                       // 司机 ID
		Phone:           req.Phone,                    // 可选手机号
		PasswordHash:    passwordHash,                 // API 层已完成 bcrypt 哈希
		RealName:        req.RealName,                 // 可选姓名
		IdCardNo:        req.IdCardNo,                 // 可选身份证号
		DriverLicenseNo: req.DriverLicenseNo,          // 可选驾驶证号
		AvatarUrl:       req.AvatarURL,                // 可选头像
		Status:          enumDriverStatus(req.Status), // 可选状态（字符串转枚举指针）
	})
	if err != nil {
		return nil, err
	}
	// 转换响应并返回。
	return &types.UpdateDriverResponse{ID: resp.GetId(), Status: resp.GetStatus().String(), UpdatedAt: resp.GetUpdatedAt()}, nil
}

// GetDriver 查询司机详情：校验 ID 合法性后调用 driversvc 查询，将 proto 实体映射为 API 的司机详情（手机号、身份证号统一脱敏后返回）。
func (l *DriverLogic) GetDriver(id int64) (*types.GetDriverResponse, error) {
	// 校验司机 ID 合法性。
	if id <= 0 {
		return nil, errors.New("司机ID不合法")
	}
	// 获取 driversvc 客户端。
	client, err := l.driverClient()
	if err != nil {
		return nil, err
	}
	// 调用下游查询接口。
	resp, err := client.GetDriver(l.ctx, &driversproto.GetDriverRequest{Id: id})
	if err != nil {
		return nil, err
	}
	// 取出 proto 中的司机实体。
	d := resp.GetDriver()
	// 将 proto 实体映射为 API 的司机详情结构并返回。
	// 注意：手机号、身份证号属敏感字段，统一脱敏后再返回（与列表/按手机号接口保持一致）。
	return &types.GetDriverResponse{Driver: types.DriverDetail{
		ID:              d.GetId(),
		Phone:           jwtx.MaskPhone(d.GetPhone()),
		RealName:        d.GetRealName(),
		IdCardNo:        maskIDCard(d.GetIdCardNo()),
		DriverLicenseNo: d.GetDriverLicenseNo(),
		AvatarURL:       d.GetAvatarUrl(),
		Status:          d.GetStatus().String(),
		OnlineStatus:    int(d.GetOnlineStatus()),
		VehicleID:       d.GetVehicleId(),
		CreatedAt:       d.GetCreatedAt(),
		UpdatedAt:       d.GetUpdatedAt(),
	}}, nil
}

// maskIDCard 对身份证号做脱敏：保留前 4 位与后 2 位，中间以 * 代替。
// 长度不足时直接返回原值，避免产生误导性的脱敏结果。
func maskIDCard(id string) string {
	const keepHead, keepTail = 4, 2
	if len(id) <= keepHead+keepTail {
		return id
	}
	masked := make([]byte, 0, len(id))
	masked = append(masked, id[:keepHead]...)
	for i := keepHead; i < len(id)-keepTail; i++ {
		masked = append(masked, '*')
	}
	masked = append(masked, id[len(id)-keepTail:]...)
	return string(masked)
}

// driverClient 从服务上下文中安全取出 driversvc 客户端。
func (l *DriverLogic) driverClient() (svc.DriverClient, error) {
	// 防御性校验：上下文或服务客户端为空时返回配置错误。
	if l.svcCtx == nil || l.svcCtx.DriverClient == nil {
		return nil, ErrDriverClientNotConfigured
	}
	// 返回已配置的客户端。
	return l.svcCtx.DriverClient, nil
}

// GetDriverAiScore 查询司机 AI 推荐得分：调用 driversvc.GetDriverAiScore，将 proto 响应映射为 API 响应。
// 降级情况下（degraded=true）透传降级标记与原因，由前端提示「已回退距离优先」。
func (l *DriverLogic) GetDriverAiScore(driverID int64) (*types.GetDriverAiScoreResponse, error) {
	client, err := l.driverClient()
	if err != nil {
		return nil, err
	}
	resp, err := client.GetDriverAiScore(l.ctx, &driversproto.GetDriverAiScoreRequest{DriverId: driverID})
	if err != nil {
		return nil, err
	}
	factors := make([]types.AiScoreFactor, 0, len(resp.GetFactors()))
	for _, f := range resp.GetFactors() {
		factors = append(factors, types.AiScoreFactor{
			Key:    f.GetKey(),
			Label:  f.GetLabel(),
			Value:  f.GetValue(),
			Impact: f.GetImpact(),
			Hint:   f.GetHint(),
		})
	}
	return &types.GetDriverAiScoreResponse{
		DriverID:      resp.GetDriverId(),
		AiScore:       resp.GetAiScore(),
		Level:         resp.GetLevel(),
		Factors:       factors,
		Degraded:      resp.GetDegraded(),
		DegradeReason: resp.GetDegradeReason(),
	}, nil
}
