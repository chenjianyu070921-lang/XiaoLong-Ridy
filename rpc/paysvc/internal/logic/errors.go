package logic

import "errors"

var (
	ErrPaymentNotFound = errors.New("payment not found")
	ErrInvalidParam    = errors.New("invalid param")
	ErrDBNotConfigured = errors.New("paysvc database not configured")
)
