package handler

import (
	"fmt"
	"net/http"

	"XiaoLong-Ridy/api/driver/internal/svc"
	"XiaoLong-Ridy/rpc/locationsvc/locationsvc"

	"github.com/zeromicro/go-zero/core/logx"
)

// reportLocationReq 司机位置上报请求体（JSON，与全局接口约定一致）。
type reportLocationReq struct {
	DriverId     int64   `json:"driverId"`
	Lng          float64 `json:"lng"`
	Lat          float64 `json:"lat"`
	OnlineStatus int32   `json:"onlineStatus"`
	City         string  `json:"city"`
}

// ReportLocationHandler 司机端位置上报入口：把司机实时位置转发给 locationsvc，
// 由其写入 Redis GEO 与位置事件流，实现对下游 producer 的真实接入。
// 方法校验（POST）由 decodeJSON 统一处理，这里只负责解析与转发。
func ReportLocationHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req reportLocationReq
		if !decodeJSON(w, r, &req) {
			return
		}
		if req.DriverId <= 0 {
			writeParamError(w, fmt.Errorf("driverId 非法"))
			return
		}
		if req.City == "" {
			writeParamError(w, fmt.Errorf("city 不能为空"))
			return
		}

		resp, err := svcCtx.LocationClient.ReportLocation(r.Context(), &locationsvc.ReportLocationReq{
			DriverId:     req.DriverId,
			Lng:          req.Lng,
			Lat:          req.Lat,
			OnlineStatus: req.OnlineStatus,
			City:         req.City,
		})
		if err != nil {
			logx.Errorf("ReportLocation failed: %v", err)
			writeError(w, http.StatusInternalServerError, 50002, "位置上报失败")
			return
		}
		writeSuccess(w, map[string]bool{"success": resp.Success})
	}
}
