package handler

import (
	"bytes"
	"fmt"
	"melogo/internal/config"
	"melogo/internal/i18n"
	"melogo/internal/middleware"
	"melogo/internal/model"
	"melogo/internal/services"
	"melogo/internal/utils"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

var userService *services.UserService
var appConfig *config.Config

// 使用music.go中定义的errorHandler，避免重复声明
// var errorHandler = utils.NewErrorHandler()

// InitUserHandler 初始化用户handler
func InitUserHandler(cfg *config.Config) {
	userService = services.NewUserService(services.DB)
	appConfig = cfg
}

// Register handles user registration
func Register(c *gin.Context) {
	if !appConfig.Auth.AllowRegistration {
		errorHandler.HandleForbidden(c, "注册功能已关闭")
		return
	}

	var req model.RegisterRequest

	// 绑定并验证请求数据
	if err := c.ShouldBindJSON(&req); err != nil {
		errorHandler.HandleBadRequest(c, "请求参数错误: "+err.Error(), err)
		return
	}

	// 调用服务层创建用户
	user, err := userService.Register(req.Username, req.Email, req.Password)
	if err != nil {
		errorHandler.HandleBadRequest(c, err.Error(), err)
		return
	}

	// 返回成功响应
	errorHandler.HandleCreated(c, model.SuccessResponse{
		Message: "注册成功",
		Data: model.UserProfile{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			Avatar:   user.Avatar,
			IsAdmin:  user.IsAdmin,
		},
	})
}

// Login handles user login
func Login(c *gin.Context) {
	var req model.LoginRequest

	// 绑定并验证请求数据
	if err := c.ShouldBindJSON(&req); err != nil {
		errorHandler.HandleBadRequest(c, "请求参数错误: "+err.Error(), err)
		return
	}

	// 调用服务层验证用户
	user, token, err := userService.Login(req.Username, req.Password)
	if err != nil {
		errorHandler.HandleUnauthorized(c, err.Error())
		return
	}

	// 设置Cookie
	c.SetCookie(
		"token", // name
		token,   // value
		3600*24, // maxAge (24小时)
		"/",     // path
		"",      // domain
		false,   // secure
		true,    // httpOnly
	)

	// 返回token和用户信息
	errorHandler.HandleOK(c, model.LoginResponse{
		Token: token,
		User: &model.UserProfile{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			Avatar:   user.Avatar,
			IsAdmin:  user.IsAdmin,
		},
	})
}

// Logout handles user logout
func Logout(c *gin.Context) {
	// 清除Cookie中的token
	c.SetCookie(
		"token", // name
		"",      // value
		-1,      // maxAge (立即过期)
		"/",     // path
		"",      // domain
		false,   // secure
		true,    // httpOnly
	)

	errorHandler.HandleOK(c, model.SuccessResponse{
		Message: "登出成功",
	})
}

// GetUserProfile retrieves user profile information
func GetUserProfile(c *gin.Context) {
	// 从context中获取当前用户ID
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		errorHandler.HandleUnauthorized(c, "未授权访问")
		return
	}

	// 查询用户信息
	user, err := userService.GetUserByID(userID)
	if err != nil {
		errorHandler.HandleNotFound(c, "用户不存在")
		return
	}

	errorHandler.HandleOK(c, model.UserProfile{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Avatar:   user.Avatar,
		IsAdmin:  user.IsAdmin,
	})
}

// UpdateUserProfile updates user profile information
func UpdateUserProfile(c *gin.Context) {
	// 从context中获取当前用户ID
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		errorHandler.HandleUnauthorized(c, "未授权访问")
		return
	}

	// 解析 multipart form
	// 10MB max memory
	if err := c.Request.ParseMultipartForm(10 << 20); err != nil {
		errorHandler.HandleBadRequest(c, "无法解析表单数据", err)
		return
	}

	email := c.PostForm("email")

	// 处理头像上传
	var avatarUpdated bool
	file, _, err := c.Request.FormFile("avatar_file")
	if err == nil {
		defer file.Close()

		// 读取文件内容到 byte slice
		buf := new(bytes.Buffer)
		if _, err := buf.ReadFrom(file); err != nil {
			errorHandler.HandleInternalServerError(c, "读取文件失败", err)
			return
		}
		fileBytes := buf.Bytes()

		// 更新数据库 BLOB
		if err := userService.UpdateUserAvatarData(userID, fileBytes); err != nil {
			errorHandler.HandleInternalServerError(c, "保存头像失败", err)
			return
		}
		avatarUpdated = true
	}

	// 更新其他信息 (Email)
	if email != "" {
		// 获取当前信息以保留旧值
		currentUser, err := userService.GetUserByID(userID)
		if err != nil {
			errorHandler.HandleInternalServerError(c, "用户查询失败", err)
			return
		}

		// 如果上传了新头像，更新URL指向头像接口
		newAvatarURL := currentUser.Avatar
		if avatarUpdated {
			newAvatarURL = fmt.Sprintf("/api/v1/user/%d/avatar", userID)
		}

		if err := userService.UpdateUser(userID, email, newAvatarURL); err != nil {
			errorHandler.HandleInternalServerError(c, "更新用户信息失败", err)
			return
		}
	} else if avatarUpdated {
		// 仅更新了头像，也需要更新URL字段指向头像接口
		newAvatarURL := fmt.Sprintf("/api/v1/user/%d/avatar", userID)
		// 获取当前Emial以保持不变
		currentUser, _ := userService.GetUserByID(userID)
		if err := userService.UpdateUser(userID, currentUser.Email, newAvatarURL); err != nil {
			utils.NewLogger().Errorf("Failed to update avatar URL: %v", err)
		}
	}

	errorHandler.HandleOK(c, model.SuccessResponse{
		Message: "用户信息更新成功",
		Data: gin.H{
			"avatar": fmt.Sprintf("/api/v1/user/%d/avatar", userID),
		},
	})
}

// GetUserAvatar 获取用户头像
func GetUserAvatar(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	data, err := userService.GetUserAvatarData(userID)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	if len(data) == 0 {
		// Redirect to default avatar or return 404
		c.Redirect(http.StatusFound, "/assets/images/default-avatar.png")
		return
	}

	// Detect content type or assume jpeg/png
	contentType := http.DetectContentType(data)
	c.Data(http.StatusOK, contentType, data)
}

// LoginPage 登录页面
func LoginPage(c *gin.Context) {
	i18n.HTML(c, http.StatusOK, "login.html", gin.H{
		"title":              "用户登录",
		"allow_registration": appConfig.Auth.AllowRegistration,
	})
}

// RegisterPage 注册页面
func RegisterPage(c *gin.Context) {
	i18n.HTML(c, http.StatusOK, "register.html", gin.H{
		"title":              "用户注册",
		"allow_registration": appConfig.Auth.AllowRegistration,
	})
}

// ProfilePage 用户信息页面
func ProfilePage(c *gin.Context) {
	i18n.HTML(c, http.StatusOK, "profile.html", gin.H{
		"title": "个人信息",
		"time":  time.Now().Format("2006-01-02 15:04:05"),
	})
}
