package proto

import "errors"

var (
	// ErrAddressNotFound 表示常用地址不存在或不属于当前用户。
	ErrAddressNotFound = errors.New("address not found")
	// ErrInvalidAddressPhone 表示常用地址联系人手机号格式错误。
	ErrInvalidAddressPhone = errors.New("异常_电话格式错误")
	// ErrInvalidLongitudeLatitude 表示常用地址经纬度为 0。
	ErrInvalidLongitudeLatitude = errors.New("异常_经纬度为0")
)
