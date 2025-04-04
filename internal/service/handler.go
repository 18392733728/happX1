package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Handler 处理所有的HTTP请求
type Handler struct{}

// NewHandler 创建一个新的Handler实例
func NewHandler() *Handler {
	return &Handler{}
}

// RegisterRoutes 注册所有的路由
func (h *Handler) RegisterRoutes(r *gin.Engine) {
	// 健康检查
	r.GET("/health", h.HealthCheck)

	// API v1 路由组
	v1 := r.Group("/api/v1")
	{
		v1.GET("/hello", h.Hello)
	}
}

// HealthCheck 健康检查处理器
func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

// Hello 示例处理器
func (h *Handler) Hello(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Welcome to HappX1 API!",
	})
}
