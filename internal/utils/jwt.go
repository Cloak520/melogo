package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// jwtSecret will be set from config
var jwtSecret []byte

// Claims 定义JWT claims结构
type Claims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// GenerateToken 生成JWT token
func GenerateToken(userID int, username string) (string, error) {
	// 设置token过期时间为24小时
	expirationTime := time.Now().Add(24 * time.Hour)

	claims := &Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "melogo",
		},
	}

	// 使用HS256签名算法
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ParseToken 解析JWT token
func ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// SetJWTSecret 设置JWT密钥（应该从配置文件读取）
func SetJWTSecret(secret string) {
	jwtSecret = []byte(secret)
}
