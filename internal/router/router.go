package router

import (
	"github.com/gin-gonic/gin"
	"happx1/internal/service"
)

// RegisterRoutes 注册路由
func RegisterRoutes(r *gin.Engine, taskHandler *service.TaskHandler) {
	// 注册任务相关路由
	taskHandler.RegisterRoutes(r)
}
