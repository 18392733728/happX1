package main

import (
	"fmt"
	"log"

	"happx1/internal/config"
	"happx1/internal/database"
	"happx1/internal/service"

	"github.com/gin-gonic/gin"
)

func main() {
	// 初始化配置
	if err := config.Init(); err != nil {
		log.Fatalf("初始化配置失败: %v", err)
	}

	// 初始化MySQL
	if err := database.InitMySQL(&config.GlobalConfig.MySQL); err != nil {
		log.Fatalf("初始化MySQL失败: %v", err)
	}

	// 初始化Redis
	if err := database.InitRedis(&config.GlobalConfig.Redis); err != nil {
		log.Fatalf("初始化Redis失败: %v", err)
	}

	// 设置gin模式
	gin.SetMode(config.GlobalConfig.Server.Mode)

	// 创建默认的gin引擎
	r := gin.Default()

	// 创建并注册处理器
	handler := service.NewHandler()
	handler.RegisterRoutes(r)

	// 启动服务器
	addr := fmt.Sprintf(":%d", config.GlobalConfig.Server.Port)
	if err := r.Run(addr); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}
