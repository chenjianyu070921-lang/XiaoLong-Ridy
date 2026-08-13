package server

import (
	"context"

	"XiaoLong-Ridy/rpc/usersvc/internal/logic"
)

type UserService struct {
	registerLogic *logic.RegisterLogic
}

// NewUserService 创建供传输层调用的 usersvc 门面。
func NewUserService(registerLogic *logic.RegisterLogic) *UserService {
	return &UserService{registerLogic: registerLogic}
}

// Register 将注册用例暴露给外部调用方。
func (s *UserService) Register(ctx context.Context, req logic.RegisterRequest) (*logic.RegisterResponse, error) {
	return s.registerLogic.Register(ctx, req)
}
