package logic

import (
	"context"
	"errors"
	"strings"

	"XiaoLong-Ridy/rpc/usersvc/internal/model"
	"XiaoLong-Ridy/rpc/usersvc/internal/repository"
	"XiaoLong-Ridy/rpc/usersvc/internal/svc"
	userproto "XiaoLong-Ridy/rpc/usersvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

var (
	// ErrAddressRepositoryNotConfigured 表示 usersvc 未注入地址仓储。
	ErrAddressRepositoryNotConfigured = errors.New("address repository not configured")
	// ErrInvalidAddress 表示常用地址请求缺少必要字段。
	ErrInvalidAddress = errors.New("invalid address")
	// ErrInvalidAddressPhone 表示常用地址联系人手机号格式错误。
	ErrInvalidAddressPhone = userproto.ErrInvalidAddressPhone
	// ErrInvalidLongitudeLatitude 表示经纬度为 0 或无效。
	ErrInvalidLongitudeLatitude = userproto.ErrInvalidLongitudeLatitude
)

// CreateAddressLogic 处理新增乘客常用地址 RPC。
type CreateAddressLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewCreateAddressLogic 创建新增地址逻辑实例。
func NewCreateAddressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateAddressLogic {
	return &CreateAddressLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// CreateAddress 校验并保存乘客常用地址。
func (l *CreateAddressLogic) CreateAddress(in *userproto.CreateAddressRequest) (*userproto.AddressInfo, error) {
	if err := validateAddressInput(in.GetUserId(), in.GetContactName(), in.GetContactPhone(), in.GetAddress(), in.GetLongitude(), in.GetLatitude()); err != nil {
		return nil, err
	}
	addresses, err := addressRepository(l.svcCtx)
	if err != nil {
		return nil, err
	}
	address := &model.UserAddress{
		UserID:       in.GetUserId(),
		ContactName:  strings.TrimSpace(in.GetContactName()),
		ContactPhone: strings.TrimSpace(in.GetContactPhone()),
		Tag:          normalizeAddressTag(in.GetTag()),
		Address:      strings.TrimSpace(in.GetAddress()),
		Longitude:    in.GetLongitude(),
		Latitude:     in.GetLatitude(),
		IsDefault:    boolToDefaultFlag(in.GetIsDefault()),
		Sort:         int(in.GetSort()),
	}
	if err := addresses.Create(l.ctx, address); err != nil {
		return nil, err
	}
	return toAddressInfo(address), nil
}

// ListAddressesLogic 处理查询乘客常用地址列表 RPC。
type ListAddressesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewListAddressesLogic 创建查询地址列表逻辑实例。
func NewListAddressesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAddressesLogic {
	return &ListAddressesLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// ListAddresses 查询指定乘客的未删除常用地址。
func (l *ListAddressesLogic) ListAddresses(in *userproto.ListAddressesRequest) (*userproto.ListAddressesResponse, error) {
	if in.GetUserId() == 0 {
		return nil, ErrInvalidAddress
	}
	addresses, err := addressRepository(l.svcCtx)
	if err != nil {
		return nil, err
	}
	list, err := addresses.ListByUser(l.ctx, in.GetUserId())
	if err != nil {
		return nil, err
	}
	resp := &userproto.ListAddressesResponse{List: make([]*userproto.AddressInfo, 0, len(list))}
	for _, address := range list {
		resp.List = append(resp.List, toAddressInfo(address))
	}
	return resp, nil
}

// UpdateAddressLogic 处理更新乘客常用地址 RPC。
type UpdateAddressLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewUpdateAddressLogic 创建更新地址逻辑实例。
func NewUpdateAddressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateAddressLogic {
	return &UpdateAddressLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// UpdateAddress 校验归属后更新乘客常用地址。
func (l *UpdateAddressLogic) UpdateAddress(in *userproto.UpdateAddressRequest) (*userproto.AddressInfo, error) {
	if in.GetId() == 0 {
		return nil, ErrInvalidAddress
	}
	if err := validateAddressInput(in.GetUserId(), in.GetContactName(), in.GetContactPhone(), in.GetAddress(), in.GetLongitude(), in.GetLatitude()); err != nil {
		return nil, err
	}
	addresses, err := addressRepository(l.svcCtx)
	if err != nil {
		return nil, err
	}
	current, err := addresses.FindByID(l.ctx, in.GetUserId(), in.GetId())
	if err != nil {
		return nil, mapAddressRepositoryError(err)
	}
	current.ContactName = strings.TrimSpace(in.GetContactName())
	current.ContactPhone = strings.TrimSpace(in.GetContactPhone())
	current.Tag = normalizeAddressTag(in.GetTag())
	current.Address = strings.TrimSpace(in.GetAddress())
	current.Longitude = in.GetLongitude()
	current.Latitude = in.GetLatitude()
	current.IsDefault = boolToDefaultFlag(in.GetIsDefault())
	current.Sort = int(in.GetSort())
	if err := addresses.Update(l.ctx, current); err != nil {
		return nil, mapAddressRepositoryError(err)
	}
	return toAddressInfo(current), nil
}

// DeleteAddressLogic 处理删除乘客常用地址 RPC。
type DeleteAddressLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewDeleteAddressLogic 创建删除地址逻辑实例。
func NewDeleteAddressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteAddressLogic {
	return &DeleteAddressLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// DeleteAddress 软删除乘客自己的常用地址。
func (l *DeleteAddressLogic) DeleteAddress(in *userproto.DeleteAddressRequest) (*userproto.DeleteAddressResponse, error) {
	if in.GetId() == 0 || in.GetUserId() == 0 {
		return nil, ErrInvalidAddress
	}
	addresses, err := addressRepository(l.svcCtx)
	if err != nil {
		return nil, err
	}
	if err := addresses.Delete(l.ctx, in.GetUserId(), in.GetId()); err != nil {
		return nil, mapAddressRepositoryError(err)
	}
	return &userproto.DeleteAddressResponse{Success: true}, nil
}

// validateAddressInput 校验常用地址 RPC 的用户、联系人、电话、地址和坐标字段。
func validateAddressInput(userID uint64, contactName, contactPhone, address string, longitude, latitude float64) error {
	if userID == 0 || strings.TrimSpace(contactName) == "" || strings.TrimSpace(address) == "" {
		return ErrInvalidAddress
	}
	if !IsValidPhone(strings.TrimSpace(contactPhone)) {
		return ErrInvalidAddressPhone
	}
	if longitude == 0 || latitude == 0 {
		return ErrInvalidLongitudeLatitude
	}
	return nil
}

// addressRepository 获取 usersvc 地址仓储依赖。
func addressRepository(svcCtx *svc.ServiceContext) (repository.AddressRepository, error) {
	if svcCtx == nil || svcCtx.Addresses == nil {
		return nil, ErrAddressRepositoryNotConfigured
	}
	return svcCtx.Addresses, nil
}

// normalizeAddressTag 将外部标签统一归一为系统支持的 home/work/other。
func normalizeAddressTag(tag string) string {
	switch strings.TrimSpace(tag) {
	case model.UserAddressTagHome:
		return model.UserAddressTagHome
	case model.UserAddressTagWork:
		return model.UserAddressTagWork
	default:
		return model.UserAddressTagOther
	}
}

// boolToDefaultFlag 将 proto 布尔字段转换为模型层默认地址标记。
func boolToDefaultFlag(value bool) int8 {
	if value {
		return model.UserAddressIsDefault
	}
	return model.UserAddressNotDefault
}

// mapAddressRepositoryError 将仓储错误转换为 usersvc 对外错误。
func mapAddressRepositoryError(err error) error {
	if errors.Is(err, repository.ErrAddressNotFound) {
		return userproto.ErrAddressNotFound
	}
	return err
}

// toAddressInfo 将内部地址模型转换为 usersvc proto 响应结构。
func toAddressInfo(address *model.UserAddress) *userproto.AddressInfo {
	return &userproto.AddressInfo{
		Id:           address.ID,
		UserId:       address.UserID,
		ContactName:  address.ContactName,
		ContactPhone: address.ContactPhone,
		Tag:          address.Tag,
		Address:      address.Address,
		Longitude:    address.Longitude,
		Latitude:     address.Latitude,
		IsDefault:    address.IsDefault == model.UserAddressIsDefault,
		Sort:         int32(address.Sort),
	}
}
