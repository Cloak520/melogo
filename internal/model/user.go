package model

import (
	"time"
)

// User represents a user in the system
type User struct {
	ID        int       `json:"id" db:"id"`
	Username  string    `json:"username" db:"username"`
	Email     string    `json:"email" db:"email"`
	Password  string    `json:"-" db:"password_hash"`
	Avatar    string    `json:"avatar" db:"avatar"`
	IsAdmin   int       `json:"is_admin" db:"is_admin"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// UserProfile represents user profile information
type UserProfile struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Avatar   string `json:"avatar"`
	IsAdmin  int    `json:"is_admin"`
}

// RegisterRequest 用户注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"omitempty,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// LoginRequest 用户登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token string       `json:"token"`
	User  *UserProfile `json:"user"`
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	Error string `json:"error"`
}

// SuccessResponse 成功响应
type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
