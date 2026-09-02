package types

// POISearchRequest 表示目的地 POI 搜索请求。
// 由乘客端前端直接调用，不依赖 AMap JS API key，避免在浏览器暴露后端高德 key。
type POISearchRequest struct {
	Keyword string  `json:"keyword"`
	City    string  `json:"city"`
	Lat     float64 `json:"lat"`
	Lng     float64 `json:"lng"`
	Radius  int32   `json:"radius"`
	Page    int32   `json:"page"`
	Size    int32   `json:"size"`
}

// POIDTO 表示一条搜索结果条目。
type POIDTO struct {
	Name     string  `json:"name"`
	Address  string  `json:"address"`
	Lat      float64 `json:"lat"`
	Lng      float64 `json:"lng"`
	Category string  `json:"category"`
	Distance int32   `json:"distance"`
}

// POISearchResponse 表示目的地 POI 搜索响应。
type POISearchResponse struct {
	Items []POIDTO `json:"items"`
	Total int32    `json:"total"`
}