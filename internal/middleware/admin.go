package middleware

import (
	"melogo/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AdminMiddleware 管理员权限中间件
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := GetCurrentUserID(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "未授权访问，请先登录",
			})
			c.Abort()
			return
		}

		// 创建UserService实例并查询用户是否为管理员
		userService := services.NewUserService(services.DB)
		user, err := userService.GetUserByID(userID)
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "用户不存在",
			})
			c.Abort()
			return
		}

		if user.IsAdmin == 0 {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "需要管理员权限才能访问",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
