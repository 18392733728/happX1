package main

import (
	"log"

	"happx1/internal/service"

	"github.com/gin-gonic/gin"
)

func main() {
	// 创建默认的gin引擎
	r := gin.Default()

	// 创建并注册处理器
	handler := service.NewHandler()
	handler.RegisterRoutes(r)

	// 启动服务器
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}
