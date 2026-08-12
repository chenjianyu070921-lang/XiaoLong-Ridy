package handler

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	"XiaoLong-Ridy/api/passenger/internal/types"
)

func writeSuccess(w http.ResponseWriter, data any) {
	writeResponse(w, http.StatusOK, 0, "success", data)
}

func writeError(w http.ResponseWriter, status, code int, message string) {
	writeResponse(w, status, code, message, nil)
}

func writeResponse(w http.ResponseWriter, status, code int, message string, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(types.Response{
		Code:      code,
		Message:   message,
		Data:      data,
		Timestamp: time.Now().Unix(),
		TraceID:   traceID(),
	})
}

func traceID() string {
	var value [8]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "trace_local"
	}
	return "trace_" + hex.EncodeToString(value[:])
}
