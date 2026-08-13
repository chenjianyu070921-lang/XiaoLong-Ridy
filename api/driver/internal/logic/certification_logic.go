package logic

import (
	"context"
	"errors"

	"XiaoLong-Ridy/api/driver/internal/svc"
	"XiaoLong-Ridy/api/driver/internal/types"
	driversproto "XiaoLong-Ridy/rpc/driversvc/proto"
)

type CertificationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCertificationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CertificationLogic {
	return &CertificationLogic{ctx: ctx, svcCtx: svcCtx}
}

func (l *CertificationLogic) CreateCertification(req *types.CreateCertificationRequest) (*types.CreateCertificationResponse, error) {
	if req.DriverID <= 0 {
		return nil, errors.New("司机ID不合法")
	}
	if req.VehicleID <= 0 {
		return nil, errors.New("车辆ID不合法")
	}
	if req.IdCardFrontURL == "" || req.IdCardBackURL == "" || req.DriverLicenseURL == "" || req.VehicleLicenseURL == "" {
		return nil, errors.New("四张认证图片地址均不能为空")
	}
	client, err := l.client()
	if err != nil {
		return nil, err
	}
	resp, err := client.CreateCertification(l.ctx, &driversproto.CreateCertificationRequest{
		DriverId:          req.DriverID,
		VehicleId:         req.VehicleID,
		IdCardFrontUrl:    req.IdCardFrontURL,
		IdCardBackUrl:     req.IdCardBackURL,
		DriverLicenseUrl:  req.DriverLicenseURL,
		VehicleLicenseUrl: req.VehicleLicenseURL,
	})
	if err != nil {
		return nil, err
	}
	return &types.CreateCertificationResponse{ID: resp.GetId(), AuditStatus: resp.GetAuditStatus()}, nil
}

func (l *CertificationLogic) UpdateCertification(req *types.UpdateCertificationRequest) (*types.UpdateCertificationResponse, error) {
	if req.ID <= 0 {
		return nil, errors.New("认证ID不合法")
	}
	if req.AuditStatus != nil && (*req.AuditStatus < 1 || *req.AuditStatus > 3) {
		return nil, errors.New("审核状态不合法(1待审核 2通过 3驳回)")
	}
	client, err := l.client()
	if err != nil {
		return nil, err
	}
	resp, err := client.UpdateCertification(l.ctx, &driversproto.UpdateCertificationRequest{
		Id:                req.ID,
		DriverId:          req.DriverID,
		VehicleId:         req.VehicleID,
		IdCardFrontUrl:    req.IdCardFrontURL,
		IdCardBackUrl:     req.IdCardBackURL,
		DriverLicenseUrl:  req.DriverLicenseURL,
		VehicleLicenseUrl: req.VehicleLicenseURL,
		AuditStatus:       req.AuditStatus,
		AuditRemark:       req.AuditRemark,
		AuditedBy:         req.AuditedBy,
		AuditedAt:         req.AuditedAt,
	})
	if err != nil {
		return nil, err
	}
	return &types.UpdateCertificationResponse{ID: resp.GetId(), AuditStatus: resp.GetAuditStatus(), UpdatedAt: resp.GetUpdatedAt()}, nil
}

func (l *CertificationLogic) GetCertification(id int64) (*types.GetCertificationResponse, error) {
	if id <= 0 {
		return nil, errors.New("认证ID不合法")
	}
	client, err := l.client()
	if err != nil {
		return nil, err
	}
	resp, err := client.GetCertification(l.ctx, &driversproto.GetCertificationRequest{Id: id})
	if err != nil {
		return nil, err
	}
	c := resp.GetCertification()
	return &types.GetCertificationResponse{Certification: types.CertificationDetail{
		ID: c.GetId(), DriverID: c.GetDriverId(), VehicleID: c.GetVehicleId(),
		IdCardFrontURL: c.GetIdCardFrontUrl(), IdCardBackURL: c.GetIdCardBackUrl(),
		DriverLicenseURL: c.GetDriverLicenseUrl(), VehicleLicenseURL: c.GetVehicleLicenseUrl(),
		AuditStatus: c.GetAuditStatus(), AuditRemark: c.GetAuditRemark(),
		AuditedBy: c.GetAuditedBy(), AuditedAt: c.GetAuditedAt(),
		CreatedAt: c.GetCreatedAt(), UpdatedAt: c.GetUpdatedAt(),
	}}, nil
}

func (l *CertificationLogic) DeleteCertification(id int64) (*types.DeleteResponse, error) {
	if id <= 0 {
		return nil, errors.New("认证ID不合法")
	}
	client, err := l.client()
	if err != nil {
		return nil, err
	}
	resp, err := client.DeleteCertification(l.ctx, &driversproto.DeleteCertificationRequest{Id: id})
	if err != nil {
		return nil, err
	}
	return &types.DeleteResponse{ID: resp.GetId(), Success: resp.GetSuccess()}, nil
}

func (l *CertificationLogic) ListCertifications(req *types.ListCertificationsRequest) (*types.ListCertificationsResponse, error) {
	page, pageSize := clampPage(req.Page, req.PageSize)
	client, err := l.client()
	if err != nil {
		return nil, err
	}
	resp, err := client.ListCertifications(l.ctx, &driversproto.ListCertificationsRequest{
		DriverId: req.DriverID, AuditStatus: req.AuditStatus, Page: page, PageSize: pageSize,
	})
	if err != nil {
		return nil, err
	}
	list := make([]types.CertificationSummary, 0, len(resp.GetList()))
	for _, s := range resp.GetList() {
		list = append(list, types.CertificationSummary{
			ID: s.GetId(), DriverID: s.GetDriverId(), VehicleID: s.GetVehicleId(),
			AuditStatus: s.GetAuditStatus(), CreatedAt: s.GetCreatedAt(),
		})
	}
	return &types.ListCertificationsResponse{List: list, Total: resp.GetTotal(), Page: resp.GetPage(), PageSize: resp.GetPageSize()}, nil
}

func (l *CertificationLogic) client() (svc.DriverClient, error) {
	if l.svcCtx == nil || l.svcCtx.DriverClient == nil {
		return nil, ErrDriverClientNotConfigured
	}
	return l.svcCtx.DriverClient, nil
}
