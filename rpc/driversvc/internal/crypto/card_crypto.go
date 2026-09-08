package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"strings"
)

// 银行卡号加解密与脱敏工具。
// 合规要求：银行卡号属于敏感信息，库内以 AES-GCM 密文存储，接口只允许返回脱敏视图。

// deriveKey 将配置密钥 SHA-256 后作为 AES-256 密钥（任意长度配置串均可使用）。
func deriveKey(key string) []byte {
	sum := sha256.Sum256([]byte(key))
	return sum[:]
}

// EncryptCardNo 使用 AES-GCM 加密银行卡号，返回 hex(ciphertext+nonce)。
func EncryptCardNo(plain, key string) (string, error) {
	plain = strings.TrimSpace(plain)
	if plain == "" {
		return "", errors.New("card no is empty")
	}
	block, err := aes.NewCipher(deriveKey(key))
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	sealed := gcm.Seal(nil, nonce, []byte(plain), nil)
	return hex.EncodeToString(append(nonce, sealed...)), nil
}

// DecryptCardNo 解密 EncryptCardNo 产生的密文。
func DecryptCardNo(enc, key string) (string, error) {
	raw, err := hex.DecodeString(strings.TrimSpace(enc))
	if err != nil {
		return "", errors.New("invalid cipher text")
	}
	block, err := aes.NewCipher(deriveKey(key))
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := gcm.NonceSize()
	if len(raw) < nonceSize {
		return "", errors.New("cipher text too short")
	}
	plain, err := gcm.Open(nil, raw[:nonceSize], raw[nonceSize:], nil)
	if err != nil {
		return "", errors.New("decrypt failed")
	}
	return string(plain), nil
}

// CardNoHash 返回卡号 SHA-256 十六进制摘要，用于同一司机重复绑卡查重与唯一约束。
func CardNoHash(plain string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(plain)))
	return hex.EncodeToString(sum[:])
}

// MaskCardNo 脱敏卡号：仅保留前 3 位与后 3 位，中间以 * 填充。
func MaskCardNo(plain string) string {
	plain = strings.TrimSpace(plain)
	if len(plain) <= 6 {
		return plain
	}
	return plain[:3] + "****" + plain[len(plain)-3:]
}
