package logic

import (
	"crypto/rand"
	"encoding/base64"
	"sync"
	"time"

	"XiaoLong-Ridy/common/jwtx"
)

const (
	accessTokenTTL  = 2 * time.Hour
	refreshTokenTTL = 7 * 24 * time.Hour
)

type refreshSession struct {
	userID     uint64
	phone      string
	userStatus int
	expiresAt  time.Time
}

// TokenManager 管理本地开发阶段的 Access Token 和 Refresh Token。
type TokenManager struct {
	mu              sync.Mutex
	signingKey      string
	refreshSessions map[string]refreshSession
	revokedTokens   map[string]time.Time
}

// NewTokenManager 创建令牌管理器；生产环境应从安全配置读取 signingKey。
func NewTokenManager(signingKey string) *TokenManager {
	return &TokenManager{
		signingKey:      signingKey,
		refreshSessions: make(map[string]refreshSession),
		revokedTokens:   make(map[string]time.Time),
	}
}

// Issue 为指定用户签发 Access Token 与 Refresh Token。
func (m *TokenManager) Issue(userID uint64, phone string, userStatus int) (string, string, error) {
	accessToken, err := jwtx.SignUserToken(jwtx.UserTokenPayload{
		UserID:     userID,
		Phone:      phone,
		Role:       "passenger",
		UserStatus: userStatus,
		Issuer:     "huaxiaozhu-api",
		TTL:        accessTokenTTL,
	}, m.signingKey)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := randomToken()
	if err != nil {
		return "", "", err
	}
	m.mu.Lock()
	// Refresh Token 仅保存服务端会话；客户端拿到的旧令牌在刷新后会被删除。
	m.refreshSessions[refreshToken] = refreshSession{
		userID:     userID,
		phone:      phone,
		userStatus: userStatus,
		expiresAt:  time.Now().Add(refreshTokenTTL),
	}
	m.mu.Unlock()
	return accessToken, refreshToken, nil
}

// Refresh 校验刷新令牌并轮换新的令牌对。
func (m *TokenManager) Refresh(refreshToken string) (string, string, error) {
	m.mu.Lock()
	session, ok := m.refreshSessions[refreshToken]
	// 先删除旧令牌，确保同一个 Refresh Token 不能被并发或重复使用。
	delete(m.refreshSessions, refreshToken)
	m.mu.Unlock()
	if !ok {
		return "", "", ErrInvalidToken
	}
	if time.Now().After(session.expiresAt) {
		return "", "", ErrTokenExpired
	}
	return m.Issue(session.userID, session.phone, session.userStatus)
}

// Revoke 使当前 Access Token 立即失效。
func (m *TokenManager) Revoke(token string) error {
	if _, err := m.Validate(token); err != nil {
		return err
	}
	m.mu.Lock()
	// 注销记录的保存时间与 Access Token 有效期一致，过期后可自动清理。
	m.revokedTokens[token] = time.Now().Add(accessTokenTTL)
	m.mu.Unlock()
	return nil
}

// Validate 校验 Access Token 的格式、签名、过期时间和注销状态。
func (m *TokenManager) Validate(token string) (*jwtx.UserClaims, error) {
	claims, err := jwtx.ParseUserToken(token, m.signingKey)
	if err == jwtx.ErrInvalidToken {
		return nil, ErrInvalidToken
	}
	if err == jwtx.ErrTokenExpired {
		return nil, ErrTokenExpired
	}
	if err != nil {
		return nil, err
	}

	// 校验注销黑名单，避免用户登出后旧 Token 仍可继续访问。
	m.mu.Lock()
	revokedUntil, revoked := m.revokedTokens[token]
	if revoked && time.Now().After(revokedUntil) {
		delete(m.revokedTokens, token)
		revoked = false
	}
	m.mu.Unlock()
	if revoked {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

func randomToken() (string, error) {
	value := make([]byte, 32)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(value), nil
}
