package types

// AddressInfo 表示乘客常用地址的接口返回结构。
type AddressInfo struct {
	ID           uint64  `json:"id"`
	ContactName  string  `json:"contactName"`
	ContactPhone string  `json:"contactPhone"`
	Tag          string  `json:"tag"`
	Address      string  `json:"address"`
	Longitude    float64 `json:"longitude"`
	Latitude     float64 `json:"latitude"`
	IsDefault    bool    `json:"isDefault"`
	Sort         int     `json:"sort"`
}

// CreateAddressRequest 表示新增常用地址请求。
type CreateAddressRequest struct {
	ContactName  string  `json:"contactName"`
	ContactPhone string  `json:"contactPhone"`
	Tag          string  `json:"tag"`
	Address      string  `json:"address"`
	Longitude    float64 `json:"longitude"`
	Latitude     float64 `json:"latitude"`
	IsDefault    bool    `json:"isDefault"`
	Sort         int     `json:"sort"`
}

// ListAddressesRequest 表示查询常用地址列表请求，当前暂不需要额外字段。
type ListAddressesRequest struct{}

// ListAddressesResponse 表示常用地址列表响应。
type ListAddressesResponse struct {
	List []AddressInfo `json:"list"`
}

// UpdateAddressRequest 表示更新常用地址请求。
type UpdateAddressRequest struct {
	ID           uint64  `json:"id"`
	ContactName  string  `json:"contactName"`
	ContactPhone string  `json:"contactPhone"`
	Tag          string  `json:"tag"`
	Address      string  `json:"address"`
	Longitude    float64 `json:"longitude"`
	Latitude     float64 `json:"latitude"`
	IsDefault    bool    `json:"isDefault"`
	Sort         int     `json:"sort"`
}

// DeleteAddressRequest 表示删除常用地址请求。
type DeleteAddressRequest struct {
	ID uint64 `json:"id"`
}

// DeleteAddressResponse 表示删除常用地址响应。
type DeleteAddressResponse struct {
	Success bool `json:"success"`
}
