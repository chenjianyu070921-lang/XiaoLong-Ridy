package logic

import (
	"XiaoLong-Ridy/api/passenger/internal/types"
	userproto "XiaoLong-Ridy/rpc/usersvc/proto"
)

// toAPIUserInfo 将 usersvc 的用户资料响应转换为 passenger API 的统一用户结构。
func toAPIUserInfo(user *userproto.UserInfo) types.UserInfo {
	if user == nil {
		return types.UserInfo{}
	}
	return types.UserInfo{
		UserID:         user.GetUserId(),
		Phone:          user.GetPhone(),
		Nickname:       user.GetNickname(),
		AvatarURL:      user.GetAvatarUrl(),
		RealNameStatus: user.GetRealNameStatus(),
	}
}
