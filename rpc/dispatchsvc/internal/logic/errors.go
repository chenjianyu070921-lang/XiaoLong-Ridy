package logic

import "errors"

var (
	// ErrInvalidOrderParams 表示订单入参不合法。
	ErrInvalidOrderParams = errors.New("invalid orderclient params")
	// ErrNoAvailableDriver 表示所有半径内都没有可用司机，需要触发上游重试队列（P0-6 修复）。
	ErrNoAvailableDriver = errors.New("no available driver found within all search radii")
)
