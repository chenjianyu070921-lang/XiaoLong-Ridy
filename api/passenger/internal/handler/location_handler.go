package handler

import (
	"net/http"

	"XiaoLong-Ridy/api/passenger/internal/logic"
	"XiaoLong-Ridy/api/passenger/internal/svc"
	"XiaoLong-Ridy/api/passenger/internal/types"
)

// SearchPOIHandler 处理 POST /api/passenger/v1/location/poi-search。
// 公开接口（无需登录），供乘客端 Home/AddressList 等页面做目的地关键词搜索。
// 调用方：乘客端 H5 前端。
func SearchPOIHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.POISearchRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		resp, err := logic.NewLocationLogic(r.Context(), svcCtx).POISearch(&req)
		if err != nil {
			writeBusinessError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}