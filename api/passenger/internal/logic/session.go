package logic

import (
	"errors"
	"strings"

	"XiaoLong-Ridy/api/passenger/internal/svc"
	"XiaoLong-Ridy/common/jwtx"
)

var (
	// ErrInvalidRequest 表示接口请求参数不符合业务要求。
	ErrInvalidRequest = errors.New("invalid request")
	// ErrUnauthorized 表示请求缺少有效登录态。
	ErrUnauthorized = errors.New("unauthorized")
	// ErrForbidden 表示当前用户无权访问指定资源。
	ErrForbidden = errors.New("forbidden")
	// ErrOrderClientNotConfigured 表示订单服务客户端未配置。
	ErrOrderClientNotConfigured = errors.New("orderclient client not configured")
	// ErrPriceClientNotConfigured 表示价格服务客户端未配置。
	ErrPriceClientNotConfigured = errors.New("price client not configured")
	// ErrPayClientNotConfigured 表示支付服务客户端未配置。
	ErrPayClientNotConfigured = errors.New("pay client not configured")
	// ErrDispatchClientNotConfigured 表示派单服务客户端未配置。
	ErrDispatchClientNotConfigured = errors.New("dispatch client not configured")
	ErrLocationClientNotConfigured  = errors.New("location client not configured")
	// ErrOrderNotPayable 表示订单当前状态不能发起支付。
	ErrOrderNotPayable = errors.New("order not payable")
)

// currentUserID 从乘客 JWT 中解析当前用户 ID。
func currentUserID(svcCtx *svc.ServiceContext, token string) (uint64, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return 0, ErrUnauthorized
	}
	if svcCtx == nil || strings.TrimSpace(svcCtx.TokenSigningKey) == "" {
		return 0, ErrUnauthorized
	}
	claims, err := jwtx.ParseUserToken(token, strings.TrimSpace(svcCtx.TokenSigningKey))
	if err != nil {
		return 0, ErrUnauthorized
	}
	if claims.AccountID == 0 || claims.AccountType != "passenger" {
		return 0, ErrUnauthorized
	}
	return claims.AccountID, nil
}
