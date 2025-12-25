package middleware

import (
	"melogo/internal/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware JWT认证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头或Cookie中获取token
		tokenString := getTokenFromRequest(c)

		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "未授权访问，请先登录",
			})
			c.Abort()
			return
		}

		// 解析token
		claims, err := utils.ParseToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Token无效或已过期",
			})
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)

		c.Next()
	}
}

// getTokenFromRequest 从请求中提取token
func getTokenFromRequest(c *gin.Context) string {
	// 优先从Authorization header获取
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		// 格式: Bearer <token>
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && parts[0] == "Bearer" {
			return parts[1]
		}
	}

	// 从Cookie获取
	token, err := c.Cookie("token")
	if err == nil {
		return token
	}

	// 从查询参数获取（可选）
	return c.Query("token")
}

// GetCurrentUserID 从上下文中获取当前用户ID
func GetCurrentUserID(c *gin.Context) (int, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}
	return userID.(int), true
}

// GetCurrentUsername 从上下文中获取当前用户名
func GetCurrentUsername(c *gin.Context) (string, bool) {
	username, exists := c.Get("username")
	if !exists {
		return "", false
	}
	return username.(string), true
}
