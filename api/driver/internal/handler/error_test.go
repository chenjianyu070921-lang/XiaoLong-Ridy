package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestWriteParamErrorMapsDriverRPCFailures(t *testing.T) {
	cases := []struct {
		name       string
		err        error
		httpStatus int
		code       int
		message    string
	}{
		{"validation", status.Error(codes.Unknown, "真实姓名不能为空"), http.StatusBadRequest, 50000, "真实姓名不能为空"},
		{"duplicate", status.Error(codes.Unknown, "driver already exists"), http.StatusConflict, codeDriverAlreadyExists, "手机号或驾驶证号已存在"},
		{"unavailable", status.Error(codes.Unavailable, "connection refused"), http.StatusBadGateway, 50001, "下游 driversvc 不可用"},
		{"not found", status.Error(codes.NotFound, "司机不存在"), http.StatusNotFound, codeDriverNotFound, "司机不存在"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			writeParamError(recorder, testCase.err)

			var response struct {
				Code    int    `json:"code"`
				Message string `json:"message"`
			}
			if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
				t.Fatalf("json.Unmarshal() error = %v", err)
			}
			if recorder.Code != testCase.httpStatus || response.Code != testCase.code || response.Message != testCase.message {
				t.Fatalf("response = status %d, %#v", recorder.Code, response)
			}
		})
	}
}
