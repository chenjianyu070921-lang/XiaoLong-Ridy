package jwtx

import (
	"errors"
	"time"

	"XiaoLong-Ridy/common/errorx"

	"github.com/golang-jwt/jwt/v5"
)

// Claims 自定义 JWT 载荷
type Claims struct {
	Uid   int64             `json:"uid"`
	Role  string            `json:"role"`
	Extra map[string]string `json:"extra,omitempty"`
	jwt.RegisteredClaims
}

// GenerateToken 生成 JWT，ttl<=0 时默认 7 天
func GenerateToken(secret string, uid int64, role string, ttl time.Duration) (string, error) {
	if ttl <= 0 {
		ttl = 7 * 24 * time.Hour
	}
	now := time.Now()
	claims := Claims{
		Uid:  uid,
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "xiao-long-ridy",
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
}

// ParseToken 解析并校验 JWT，过期/无效分别返回对应错误码
func ParseToken(secret, token string) (*Claims, error) {
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errorx.NewErr(errorx.TokenInvalid)
		}
		return []byte(secret), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, errorx.NewErr(errorx.TokenExpired)
		}
		return nil, errorx.NewErr(errorx.TokenInvalid)
	}
	return claims, nil
}
