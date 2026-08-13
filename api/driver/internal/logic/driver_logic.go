package logic

import (
	"context"
	"errors"

	"XiaoLong-Ridy/api/driver/internal/svc"
	"XiaoLong-Ridy/api/driver/internal/types"
	driversproto "XiaoLong-Ridy/rpc/driversvc/proto"
)

type DriverLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDriverLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DriverLogic {
	return &DriverLogic{ctx: ctx, svcCtx: svcCtx}
}

// CreateDriver 创建司机，校验必填项与手机号/身份证格式。
func (l *DriverLogic) CreateDriver(req *types.CreateDriverRequest) (*types.CreateDriverResponse, error) {
	if !validPhone(req.Phone) {
		return nil, errors.New("手机号格式不合法")
	}
	if req.RealName == "" {
		return nil, errors.New("真实姓名不能为空")
	}
	if !validIDCard(req.IdCardNo) {
		return nil, errors.New("身份证号格式不合法")
	}
	if req.DriverLicenseNo == "" {
		return nil, errors.New("驾驶证号不能为空")
	}
	client, err := l.driverClient()
	if err != nil {
		return nil, err
	}
	resp, err := client.CreateDriver(l.ctx, &driversproto.CreateDriverRequest{
		Phone:           req.Phone,
		PasswordHash:    req.PasswordHash,
		RealName:        req.RealName,
		IdCardNo:        req.IdCardNo,
		DriverLicenseNo: req.DriverLicenseNo,
		AvatarUrl:       req.AvatarURL,
	})
	if err != nil {
		return nil, err
	}
	return &types.CreateDriverResponse{ID: resp.GetId(), Status: resp.GetStatus().String(), CreatedAt: resp.GetCreatedAt()}, nil
}

// UpdateDriver 更新司机，校验存在性后转发可选字段。
func (l *DriverLogic) UpdateDriver(req *types.UpdateDriverRequest) (*types.UpdateDriverResponse, error) {
	if req.ID <= 0 {
		return nil, errors.New("司机ID不合法")
	}
	if req.Phone != nil && !validPhone(*req.Phone) {
		return nil, errors.New("手机号格式不合法")
	}
	if req.IdCardNo != nil && !validIDCard(*req.IdCardNo) {
		return nil, errors.New("身份证号格式不合法")
	}
	client, err := l.driverClient()
	if err != nil {
		return nil, err
	}
	resp, err := client.UpdateDriver(l.ctx, &driversproto.UpdateDriverRequest{
		Id:              req.ID,
		Phone:           req.Phone,
		PasswordHash:    req.PasswordHash,
		RealName:        req.RealName,
		IdCardNo:        req.IdCardNo,
		DriverLicenseNo: req.DriverLicenseNo,
		AvatarUrl:       req.AvatarURL,
		Status:          enumDriverStatus(req.Status),
	})
	if err != nil {
		return nil, err
	}
	return &types.UpdateDriverResponse{ID: resp.GetId(), Status: resp.GetStatus().String(), UpdatedAt: resp.GetUpdatedAt()}, nil
}

// GetDriver 查询司机详情。
func (l *DriverLogic) GetDriver(id int64) (*types.GetDriverResponse, error) {
	if id <= 0 {
		return nil, errors.New("司机ID不合法")
	}
	client, err := l.driverClient()
	if err != nil {
		return nil, err
	}
	resp, err := client.GetDriver(l.ctx, &driversproto.GetDriverRequest{Id: id})
	if err != nil {
		return nil, err
	}
	d := resp.GetDriver()
	return &types.GetDriverResponse{Driver: types.DriverDetail{
		ID: d.GetId(), Phone: d.GetPhone(), RealName: d.GetRealName(), IdCardNo: d.GetIdCardNo(),
		DriverLicenseNo: d.GetDriverLicenseNo(), AvatarURL: d.GetAvatarUrl(), Status: d.GetStatus().String(),
		CreatedAt: d.GetCreatedAt(), UpdatedAt: d.GetUpdatedAt(),
	}}, nil
}

// DeleteDriver 删除（软删）司机。
func (l *DriverLogic) DeleteDriver(id int64) (*types.DeleteResponse, error) {
	if id <= 0 {
		return nil, errors.New("司机ID不合法")
	}
	client, err := l.driverClient()
	if err != nil {
		return nil, err
	}
	resp, err := client.DeleteDriver(l.ctx, &driversproto.DeleteDriverRequest{Id: id})
	if err != nil {
		return nil, err
	}
	return &types.DeleteResponse{ID: resp.GetId(), Success: resp.GetSuccess()}, nil
}

// ListDrivers 分页查询司机列表。
func (l *DriverLogic) ListDrivers(req *types.ListDriversRequest) (*types.ListDriversResponse, error) {
	page, pageSize := clampPage(req.Page, req.PageSize)
	client, err := l.driverClient()
	if err != nil {
		return nil, err
	}
	resp, err := client.ListDrivers(l.ctx, &driversproto.ListDriversRequest{
		Status:       enumDriverStatusStr(req.Status),
		PhoneKeyword: req.PhoneKeyword,
		Page:         page,
		PageSize:     pageSize,
	})
	if err != nil {
		return nil, err
	}
	list := make([]types.DriverSummary, 0, len(resp.GetList()))
	for _, s := range resp.GetList() {
		list = append(list, types.DriverSummary{
			ID: s.GetId(), Phone: s.GetPhone(), RealName: s.GetRealName(),
			DriverLicenseNo: s.GetDriverLicenseNo(), Status: s.GetStatus().String(), CreatedAt: s.GetCreatedAt(),
		})
	}
	return &types.ListDriversResponse{List: list, Total: resp.GetTotal(), Page: resp.GetPage(), PageSize: resp.GetPageSize()}, nil
}

func (l *DriverLogic) driverClient() (svc.DriverClient, error) {
	if l.svcCtx == nil || l.svcCtx.DriverClient == nil {
		return nil, ErrDriverClientNotConfigured
	}
	return l.svcCtx.DriverClient, nil
}
