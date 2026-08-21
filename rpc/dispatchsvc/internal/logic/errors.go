package logic

import "errors"

var (
	// ErrInvalidOrderParams 表示订单入参不合法。
	ErrInvalidOrderParams = errors.New("invalid orderclient params")
)
