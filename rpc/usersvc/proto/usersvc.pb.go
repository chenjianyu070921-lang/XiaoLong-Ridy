package proto

// 本文件临时承载 usersvc.proto 对应的 Go 类型。
// 接入 protoc/goctl 后，可用生成文件替换。

type SendSMSCodeRequest struct {
	Phone string
}

func (x *SendSMSCodeRequest) GetPhone() string {
	if x != nil {
		return x.Phone
	}
	return ""
}

type SendSMSCodeResponse struct {
	Success  bool
	ExpireIn int64
}

func (x *SendSMSCodeResponse) GetSuccess() bool {
	if x != nil {
		return x.Success
	}
	return false
}

func (x *SendSMSCodeResponse) GetExpireIn() int64 {
	if x != nil {
		return x.ExpireIn
	}
	return 0
}

type LoginBySMSRequest struct {
	Phone string
	Code  string
}

func (x *LoginBySMSRequest) GetPhone() string {
	if x != nil {
		return x.Phone
	}
	return ""
}

func (x *LoginBySMSRequest) GetCode() string {
	if x != nil {
		return x.Code
	}
	return ""
}

type UserInfo struct {
	UserId         uint64
	Phone          string
	Nickname       string
	AvatarUrl      string
	RealNameStatus string
}

func (x *UserInfo) GetUserId() uint64 {
	if x != nil {
		return x.UserId
	}
	return 0
}

func (x *UserInfo) GetPhone() string {
	if x != nil {
		return x.Phone
	}
	return ""
}

func (x *UserInfo) GetNickname() string {
	if x != nil {
		return x.Nickname
	}
	return ""
}

func (x *UserInfo) GetAvatarUrl() string {
	if x != nil {
		return x.AvatarUrl
	}
	return ""
}

func (x *UserInfo) GetRealNameStatus() string {
	if x != nil {
		return x.RealNameStatus
	}
	return ""
}

type LoginBySMSResponse struct {
	Token        string
	RefreshToken string
	IsNewUser    bool
	User         *UserInfo
}

func (x *LoginBySMSResponse) GetToken() string {
	if x != nil {
		return x.Token
	}
	return ""
}

func (x *LoginBySMSResponse) GetRefreshToken() string {
	if x != nil {
		return x.RefreshToken
	}
	return ""
}

func (x *LoginBySMSResponse) GetIsNewUser() bool {
	if x != nil {
		return x.IsNewUser
	}
	return false
}

func (x *LoginBySMSResponse) GetUser() *UserInfo {
	if x != nil {
		return x.User
	}
	return nil
}

type RefreshTokenRequest struct {
	RefreshToken string
}

func (x *RefreshTokenRequest) GetRefreshToken() string {
	if x != nil {
		return x.RefreshToken
	}
	return ""
}

type RefreshTokenResponse struct {
	Token        string
	RefreshToken string
}

func (x *RefreshTokenResponse) GetToken() string {
	if x != nil {
		return x.Token
	}
	return ""
}

func (x *RefreshTokenResponse) GetRefreshToken() string {
	if x != nil {
		return x.RefreshToken
	}
	return ""
}

type LogoutRequest struct {
	Token string
}

func (x *LogoutRequest) GetToken() string {
	if x != nil {
		return x.Token
	}
	return ""
}

type LogoutResponse struct {
	Success bool
}

func (x *LogoutResponse) GetSuccess() bool {
	if x != nil {
		return x.Success
	}
	return false
}
