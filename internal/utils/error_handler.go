package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorHandler 是错误处理工具
type ErrorHandler struct {
	logger *Logger
}

// NewErrorHandler 创建新的错误处理器
func NewErrorHandler() *ErrorHandler {
	return &ErrorHandler{
		logger: NewLogger(),
	}
}

// HandleBadRequest 处理400错误
func (eh *ErrorHandler) HandleBadRequest(c *gin.Context, message string, err error) {
	if err != nil {
		eh.logger.Errorf("%s: %v", message, err)
	}
	c.JSON(http.StatusBadRequest, gin.H{"error": message})
}

// HandleUnauthorized 处理401错误
func (eh *ErrorHandler) HandleUnauthorized(c *gin.Context, message string) {
	eh.logger.Warningf("Unauthorized access attempt: %s", message)
	c.JSON(http.StatusUnauthorized, gin.H{"error": message})
}

// HandleForbidden 处理403错误
func (eh *ErrorHandler) HandleForbidden(c *gin.Context, message string) {
	eh.logger.Warningf("Forbidden access attempt: %s", message)
	c.JSON(http.StatusForbidden, gin.H{"error": message})
}

// HandleNotFound 处理404错误
func (eh *ErrorHandler) HandleNotFound(c *gin.Context, message string) {
	eh.logger.Warningf("Resource not found: %s", message)
	c.JSON(http.StatusNotFound, gin.H{"error": message})
}

// HandleInternalServerError 处理500错误
func (eh *ErrorHandler) HandleInternalServerError(c *gin.Context, message string, err error) {
	if err != nil {
		eh.logger.Errorf("%s: %v", message, err)
	} else {
		eh.logger.Error(message)
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": message})
}

// HandleSuccess 处理成功响应
func (eh *ErrorHandler) HandleSuccess(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(statusCode, data)
}

// HandleCreated 处理创建成功响应
func (eh *ErrorHandler) HandleCreated(c *gin.Context, data interface{}) {
	eh.HandleSuccess(c, http.StatusCreated, data)
}

// HandleOK 处理成功响应
func (eh *ErrorHandler) HandleOK(c *gin.Context, data interface{}) {
	eh.HandleSuccess(c, http.StatusOK, data)
}
