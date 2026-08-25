package logic

import (
	"context"
	"strings"

	"XiaoLong-Ridy/api/passenger/internal/svc"
	"XiaoLong-Ridy/api/passenger/internal/types"
	userproto "XiaoLong-Ridy/rpc/usersvc/proto"
)

// ProfileLogic 封装乘客个人中心业务流程。
// API 层只负责解析登录态和调用 usersvc RPC，不直接读写用户资料。
type ProfileLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	token  string
}

// NewProfileLogic 创建个人中心业务逻辑实例。
func NewProfileLogic(ctx context.Context, svcCtx *svc.ServiceContext, token string) *ProfileLogic {
	return &ProfileLogic{ctx: ctx, svcCtx: svcCtx, token: token}
}

// GetProfile 调用 usersvc.GetProfile 查询当前乘客个人中心资料。
func (l *ProfileLogic) GetProfile(_ *types.GetProfileRequest) (*types.GetProfileResponse, error) {
	userID, err := currentUserID(l.svcCtx, l.token)
	if err != nil {
		return nil, err
	}
	client, err := l.userClient()
	if err != nil {
		return nil, err
	}
	resp, err := client.GetProfile(l.ctx, &userproto.GetProfileRequest{UserId: userID})
	if err != nil {
		return nil, err
	}
	return &types.GetProfileResponse{User: toAPIUserInfo(resp.GetUser())}, nil
}

// SubmitRealName 调用 usersvc.SubmitRealName 提交当前乘客实名资料。
func (l *ProfileLogic) SubmitRealName(req *types.SubmitRealNameRequest) (*types.SubmitRealNameResponse, error) {
	userID, err := currentUserID(l.svcCtx, l.token)
	if err != nil {
		return nil, err
	}
	if req == nil || strings.TrimSpace(req.RealName) == "" || strings.TrimSpace(req.IDCardNo) == "" {
		return nil, ErrInvalidRequest
	}
	client, err := l.userClient()
	if err != nil {
		return nil, err
	}
	resp, err := client.SubmitRealName(l.ctx, &userproto.SubmitRealNameRequest{
		UserId:   userID,
		RealName: strings.TrimSpace(req.RealName),
		IdCardNo: strings.TrimSpace(req.IDCardNo),
	})
	if err != nil {
		return nil, err
	}
	return &types.SubmitRealNameResponse{User: toAPIUserInfo(resp.GetUser())}, nil
}

// UpdateProfile 调用 usersvc.UpdateProfile 更新当前乘客昵称与头像。
func (l *ProfileLogic) UpdateProfile(req *types.UpdateProfileRequest) (*types.UpdateProfileResponse, error) {
	userID, err := currentUserID(l.svcCtx, l.token)
	if err != nil {
		return nil, err
	}
	if req == nil || (strings.TrimSpace(req.Nickname) == "" && strings.TrimSpace(req.AvatarURL) == "") {
		return nil, ErrInvalidRequest
	}
	client, err := l.userClient()
	if err != nil {
		return nil, err
	}
	resp, err := client.UpdateProfile(l.ctx, &userproto.UpdateProfileRequest{
		UserId:    userID,
		Nickname:  strings.TrimSpace(req.Nickname),
		AvatarUrl: strings.TrimSpace(req.AvatarURL),
	})
	if err != nil {
		return nil, err
	}
	return &types.UpdateProfileResponse{User: toAPIUserInfo(resp.GetUser())}, nil
}

// userClient 获取 usersvc RPC 客户端。
func (l *ProfileLogic) userClient() (svc.UserClient, error) {
	if l.svcCtx == nil || l.svcCtx.UserClient == nil {
		return nil, ErrUserClientNotConfigured
	}
	return l.svcCtx.UserClient, nil
}
