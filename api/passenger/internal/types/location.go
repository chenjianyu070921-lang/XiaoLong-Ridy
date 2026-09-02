package types

// ReverseGeocodeRequest 表示乘客端 GPS 坐标反查地址请求。
type ReverseGeocodeRequest struct {
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
}

// ReverseGeocodeResponse 表示乘客端坐标反查出的地址信息。
type ReverseGeocodeResponse struct {
	Address   string  `json:"address"`
	Province  string  `json:"province"`
	CityName  string  `json:"cityName"`
	District  string  `json:"district"`
	PoiName   string  `json:"poiName"`
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
}

// GeocodeRequest 表示乘客端文字地址解析请求。
type GeocodeRequest struct {
	Address   string  `json:"address"`
	CityCode  string  `json:"cityCode"`
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
	Radius    int32   `json:"radius"`
}

// GeocodeResponse 表示乘客端文字地址解析出的首个候选位置。
type GeocodeResponse struct {
	Name      string  `json:"name"`
	Address   string  `json:"address"`
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
	Category  string  `json:"category"`
	Distance  int32   `json:"distance"`
	CityCode  string  `json:"cityCode"`
}
