package logic

import (
	"context"
	"strings"

	"XiaoLong-Ridy/api/passenger/internal/svc"
	"XiaoLong-Ridy/api/passenger/internal/types"
	userproto "XiaoLong-Ridy/rpc/usersvc/proto"
)

// AddressLogic 封装乘客常用地址业务流程。
// API 层只负责解析登录态和调用 usersvc RPC，不直接操作地址仓储。
type AddressLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	token  string
}

// NewAddressLogic 创建常用地址业务逻辑实例。
func NewAddressLogic(ctx context.Context, svcCtx *svc.ServiceContext, token string) *AddressLogic {
	return &AddressLogic{ctx: ctx, svcCtx: svcCtx, token: token}
}

// CreateAddress 调用 usersvc.CreateAddress 新增当前乘客常用地址。
func (l *AddressLogic) CreateAddress(req *types.CreateAddressRequest) (*types.AddressInfo, error) {
	userID, err := currentUserID(l.svcCtx, l.token)
	if err != nil {
		return nil, err
	}
	if err := validateAddress(req.ContactName, req.ContactPhone, req.Address, req.Longitude, req.Latitude); err != nil {
		return nil, err
	}
	client, err := l.userClient()
	if err != nil {
		return nil, err
	}
	address, err := client.CreateAddress(l.ctx, &userproto.CreateAddressRequest{
		UserId:       userID,
		ContactName:  req.ContactName,
		ContactPhone: req.ContactPhone,
		Tag:          req.Tag,
		Address:      req.Address,
		Longitude:    req.Longitude,
		Latitude:     req.Latitude,
		IsDefault:    req.IsDefault,
		Sort:         int32(req.Sort),
	})
	if err != nil {
		return nil, err
	}
	return toAPIAddress(address), nil
}

// ListAddresses 调用 usersvc.ListAddresses 查询当前乘客常用地址列表。
func (l *AddressLogic) ListAddresses(_ *types.ListAddressesRequest) (*types.ListAddressesResponse, error) {
	userID, err := currentUserID(l.svcCtx, l.token)
	if err != nil {
		return nil, err
	}
	client, err := l.userClient()
	if err != nil {
		return nil, err
	}
	resp, err := client.ListAddresses(l.ctx, &userproto.ListAddressesRequest{UserId: userID})
	if err != nil {
		return nil, err
	}
	list := make([]types.AddressInfo, 0, len(resp.GetList()))
	for _, item := range resp.GetList() {
		// 下游异常返回空地址时跳过该条，避免 API 层因为解引用 nil 崩溃。
		if apiAddress := toAPIAddress(item); apiAddress != nil {
			list = append(list, *apiAddress)
		}
	}
	return &types.ListAddressesResponse{List: list}, nil
}

// UpdateAddress 调用 usersvc.UpdateAddress 更新当前乘客常用地址。
func (l *AddressLogic) UpdateAddress(req *types.UpdateAddressRequest) (*types.AddressInfo, error) {
	userID, err := currentUserID(l.svcCtx, l.token)
	if err != nil {
		return nil, err
	}
	if req.ID == 0 {
		return nil, ErrInvalidRequest
	}
	if err := validateAddress(req.ContactName, req.ContactPhone, req.Address, req.Longitude, req.Latitude); err != nil {
		return nil, err
	}
	client, err := l.userClient()
	if err != nil {
		return nil, err
	}
	address, err := client.UpdateAddress(l.ctx, &userproto.UpdateAddressRequest{
		Id:           req.ID,
		UserId:       userID,
		ContactName:  req.ContactName,
		ContactPhone: req.ContactPhone,
		Tag:          req.Tag,
		Address:      req.Address,
		Longitude:    req.Longitude,
		Latitude:     req.Latitude,
		IsDefault:    req.IsDefault,
		Sort:         int32(req.Sort),
	})
	if err != nil {
		return nil, err
	}
	return toAPIAddress(address), nil
}

// DeleteAddress 调用 usersvc.DeleteAddress 删除当前乘客常用地址。
func (l *AddressLogic) DeleteAddress(req *types.DeleteAddressRequest) (*types.DeleteAddressResponse, error) {
	userID, err := currentUserID(l.svcCtx, l.token)
	if err != nil {
		return nil, err
	}
	if req.ID == 0 {
		return nil, ErrInvalidRequest
	}
	client, err := l.userClient()
	if err != nil {
		return nil, err
	}
	resp, err := client.DeleteAddress(l.ctx, &userproto.DeleteAddressRequest{Id: req.ID, UserId: userID})
	if err != nil {
		return nil, err
	}
	return &types.DeleteAddressResponse{Success: resp.GetSuccess()}, nil
}

// validateAddress 校验常用地址的联系人、电话、详细地址和经纬度是否完整合法。
func validateAddress(contactName, contactPhone, address string, longitude, latitude float64) error {
	if strings.TrimSpace(contactName) == "" || strings.TrimSpace(contactPhone) == "" || strings.TrimSpace(address) == "" {
		return ErrInvalidRequest
	}
	if !isValidPassengerPhone(contactPhone) {
		return userproto.ErrInvalidAddressPhone
	}
	if !isValidLongitudeLatitude(longitude, latitude) {
		return userproto.ErrInvalidLongitudeLatitude
	}
	return nil
}

// userClient 获取 usersvc RPC 客户端。
func (l *AddressLogic) userClient() (svc.UserClient, error) {
	if l.svcCtx == nil || l.svcCtx.UserClient == nil {
		return nil, ErrUserClientNotConfigured
	}
	return l.svcCtx.UserClient, nil
}

// toAPIAddress 将 usersvc 地址 proto 响应转换为乘客端 API 响应结构。
func toAPIAddress(address *userproto.AddressInfo) *types.AddressInfo {
	if address == nil {
		return nil
	}
	return &types.AddressInfo{
		ID:           address.GetId(),
		ContactName:  address.GetContactName(),
		ContactPhone: address.GetContactPhone(),
		Tag:          address.GetTag(),
		Address:      address.GetAddress(),
		Longitude:    address.GetLongitude(),
		Latitude:     address.GetLatitude(),
		IsDefault:    address.GetIsDefault(),
		Sort:         int(address.GetSort()),
	}
}
