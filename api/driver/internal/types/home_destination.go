package types

// HomeDestinationResponse 司机回家目的地与回家顺路模式状态。
type HomeDestinationResponse struct {
	HasSetting      bool    `json:"hasSetting"`
	HomeAddr        string  `json:"homeAddr"`
	HomeLng         float64 `json:"homeLng"`
	HomeLat         float64 `json:"homeLat"`
	IsHomeModeOpen  bool    `json:"isHomeModeOpen"`
	MaxDetourRatio  float64 `json:"maxDetourRatio"`
}

// SetHomeDestinationRequest 保存回家目的地的请求；Open 为 true 时保存后同时开启回家模式。
type SetHomeDestinationRequest struct {
	HomeAddr       string  `json:"homeAddr"`
	HomeLng        float64 `json:"homeLng"`
	HomeLat        float64 `json:"homeLat"`
	Open           bool    `json:"open"`
	MaxDetourRatio float64 `json:"maxDetourRatio"`
}

// SetHomeModeRequest 仅切换回家顺路模式开关。
type SetHomeModeRequest struct {
	Open bool `json:"open"`
}
