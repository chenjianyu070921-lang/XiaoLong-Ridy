package cryptox

import (
	"crypto/md5"
	"encoding/hex"

	"golang.org/x/crypto/bcrypt"
)

const DefaultBcryptCost = bcrypt.DefaultCost

// 本包是全项目通用加密工具，api/admin、api/driver、rpc/usersvc 等服务都可以复用。
// 新增密码字段时优先使用 bcrypt；MD5 仅保留给非密码摘要或历史兼容。

// MD5 返回普通 MD5 摘要，适合非密码场景的兼容签名或摘要。
// 密码存储不要使用 MD5，请使用 BcryptHash。
func MD5(plain string) string {
	sum := md5.Sum([]byte(plain))
	return hex.EncodeToString(sum[:])
}

// BcryptHash 使用 bcrypt 生成密码哈希，适合用户登录密码存储。
func BcryptHash(password string) (string, error) {
	return BcryptHashWithCost(password, DefaultBcryptCost)
}

// BcryptHashWithCost 使用指定 cost 生成 bcrypt 哈希；cost 越高计算越慢。
func BcryptHashWithCost(password string, cost int) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

// BcryptCompare 校验明文密码和数据库中的 bcrypt 哈希是否匹配。
func BcryptCompare(password, passwordHash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)) == nil
}
