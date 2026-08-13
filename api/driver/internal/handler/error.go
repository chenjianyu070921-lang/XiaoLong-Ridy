package handler

import (
	"errors"
	"net/http"

	"XiaoLong-Ridy/api/driver/internal/logic"
)

// writeParamError 将 logic 层返回的业务/校验错误映射为统一 HTTP 响应。
func writeParamError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, logic.ErrDriverClientNotConfigured):
		writeError(w, http.StatusBadGateway, 50001, "下游 driversvc 不可用")
	case errors.Is(err, logic.ErrInvalidParam):
		writeError(w, http.StatusBadRequest, 50000, err.Error())
	default:
		// logic 层直接返回的校验错误带有中文描述，归为参数错误。
		writeError(w, http.StatusBadRequest, 50000, err.Error())
	}
}
