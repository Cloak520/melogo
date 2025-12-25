package utils

import (
	"golang.org/x/crypto/bcrypt"
)

// HashPassword 使用bcrypt加密密码
func HashPassword(password string) (string, error) {
	// 使用默认成本参数（10）
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// VerifyPassword 验证密码是否匹配
func VerifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
