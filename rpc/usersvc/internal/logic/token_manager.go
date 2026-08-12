package logic

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	accessTokenTTL  = 2 * time.Hour
	refreshTokenTTL = 7 * 24 * time.Hour
)

type tokenClaims struct {
	Subject string `json:"sub"`
	Phone   string `json:"phone"`
	Role    string `json:"role"`
	Issued  int64  `json:"iat"`
	Expire  int64  `json:"exp"`
	Issuer  string `json:"iss"`
}

type refreshSession struct {
	userID    uint64
	phone     string
	expiresAt time.Time
}

// TokenManager 管理本地开发阶段的 Access Token 和 Refresh Token。
type TokenManager struct {
	mu              sync.Mutex
	signingKey      []byte
	refreshSessions map[string]refreshSession
	revokedTokens   map[string]time.Time
}

// NewTokenManager 创建令牌管理器；生产环境应从安全配置读取 signingKey。
func NewTokenManager(signingKey string) *TokenManager {
	return &TokenManager{
		signingKey:      []byte(signingKey),
		refreshSessions: make(map[string]refreshSession),
		revokedTokens:   make(map[string]time.Time),
	}
}

// Issue 为指定用户签发 Access Token 与 Refresh Token。
func (m *TokenManager) Issue(userID uint64, phone string) (string, string, error) {
	accessToken, err := m.issueAccessToken(userID, phone)
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
		userID:    userID,
		phone:     phone,
		expiresAt: time.Now().Add(refreshTokenTTL),
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
	return m.Issue(session.userID, session.phone)
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
func (m *TokenManager) Validate(token string) (*tokenClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return nil, ErrInvalidToken
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil || !hmac.Equal([]byte(parts[1]), []byte(m.sign(parts[0]))) {
		return nil, ErrInvalidToken
	}

	var claims tokenClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, ErrInvalidToken
	}
	if time.Now().Unix() >= claims.Expire {
		return nil, ErrTokenExpired
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
	return &claims, nil
}

func (m *TokenManager) issueAccessToken(userID uint64, phone string) (string, error) {
	now := time.Now()
	payload, err := json.Marshal(tokenClaims{
		Subject: "user_" + formatUserID(userID),
		Phone:   maskTokenPhone(phone),
		Role:    "passenger",
		Issued:  now.Unix(),
		Expire:  now.Add(accessTokenTTL).Unix(),
		Issuer:  "huaxiaozhu-api",
	})
	if err != nil {
		return "", err
	}
	encoded := base64.RawURLEncoding.EncodeToString(payload)
	return encoded + "." + m.sign(encoded), nil
}

func (m *TokenManager) sign(payload string) string {
	mac := hmac.New(sha256.New, m.signingKey)
	_, _ = mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func randomToken() (string, error) {
	value := make([]byte, 32)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(value), nil
}

func formatUserID(userID uint64) string {
	return strconv.FormatUint(userID, 10)
}

func maskTokenPhone(phone string) string {
	if len(phone) != 11 {
		return phone
	}
	return phone[:3] + "****" + phone[7:]
}
