package jwtx

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrTokenExpired = errors.New("token expired")
)

type UserClaims struct {
	AccountClaims
}

type UserTokenPayload struct {
	UserID     uint64
	Phone      string
	Role       string
	UserStatus int
	Issuer     string
	TTL        time.Duration
}

type AccountClaims struct {
	Subject       string `json:"sub"`
	AccountID     uint64 `json:"accountId"`
	AccountType   string `json:"accountType"`
	AccountStatus int    `json:"accountStatus"`
	Phone         string `json:"phone,omitempty"`
	Role          string `json:"role,omitempty"`
	IssuedAt      int64  `json:"iat"`
	ExpiresAt     int64  `json:"exp"`
	Issuer        string `json:"iss"`
}

type AccountTokenPayload struct {
	AccountID     uint64
	AccountType   string
	AccountStatus int
	Phone         string
	Role          string
	Issuer        string
	TTL           time.Duration
}

// SignAccountToken 签发通用账号 Token，适用于 passenger/driver/admin 等服务。
// 这里采用 header.payload.signature 三段式 HMAC-SHA256，便于后续替换标准 JWT 库。
func SignAccountToken(payload AccountTokenPayload, signingKey string) (string, error) {
	now := time.Now()
	claims := AccountClaims{
		Subject:       payload.AccountType + "_" + strconv.FormatUint(payload.AccountID, 10),
		AccountID:     payload.AccountID,
		AccountType:   payload.AccountType,
		AccountStatus: payload.AccountStatus,
		Phone:         MaskPhone(payload.Phone),
		Role:          payload.Role,
		IssuedAt:      now.Unix(),
		ExpiresAt:     now.Add(payload.TTL).Unix(),
		Issuer:        payload.Issuer,
	}

	header, err := json.Marshal(map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	})
	if err != nil {
		return "", err
	}
	body, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	unsigned := base64.RawURLEncoding.EncodeToString(header) + "." + base64.RawURLEncoding.EncodeToString(body)
	return unsigned + "." + sign(unsigned, signingKey), nil
}

// ParseAccountToken 解密/解析通用账号 Token，并校验签名和过期时间。
func ParseAccountToken(token, signingKey string) (*AccountClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidToken
	}

	unsigned := parts[0] + "." + parts[1]
	if !hmac.Equal([]byte(parts[2]), []byte(sign(unsigned, signingKey))) {
		return nil, ErrInvalidToken
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, ErrInvalidToken
	}
	var claims AccountClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, ErrInvalidToken
	}
	if time.Now().Unix() >= claims.ExpiresAt {
		return nil, ErrTokenExpired
	}
	return &claims, nil
}

// SignUserToken 兼容乘客用户场景，内部复用通用账号 Token。
func SignUserToken(payload UserTokenPayload, signingKey string) (string, error) {
	return SignAccountToken(AccountTokenPayload{
		AccountID:     payload.UserID,
		AccountType:   "passenger",
		AccountStatus: payload.UserStatus,
		Phone:         payload.Phone,
		Role:          payload.Role,
		Issuer:        payload.Issuer,
		TTL:           payload.TTL,
	}, signingKey)
}

// ParseUserToken 兼容乘客用户场景，返回历史 UserClaims 类型。
func ParseUserToken(token, signingKey string) (*UserClaims, error) {
	claims, err := ParseAccountToken(token, signingKey)
	if err != nil {
		return nil, err
	}
	return &UserClaims{AccountClaims: *claims}, nil
}

func sign(unsigned, signingKey string) string {
	mac := hmac.New(sha256.New, []byte(signingKey))
	_, _ = mac.Write([]byte(unsigned))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

// MaskPhone 对手机号做中间四位脱敏，用于 JWT claims 和接口返回。
func MaskPhone(phone string) string {
	if len(phone) != 11 {
		return phone
	}
	return phone[:3] + "****" + phone[7:]
}
