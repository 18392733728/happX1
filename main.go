package main

import (
	"fmt"
	"log"

	"happx1/internal/config"
	"happx1/internal/database"
	"happx1/internal/scheduler"
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

	// 初始化调度器
	scheduler := scheduler.NewScheduler()
	if err := scheduler.Start(); err != nil {
		log.Fatalf("启动调度器失败: %v", err)
	}
	defer scheduler.Stop()

	// 设置gin模式
	gin.SetMode(config.GlobalConfig.Server.Mode)

	// 创建默认的gin引擎
	r := gin.Default()

	// 创建服务层
	taskService := service.NewTaskService(scheduler, database.DB)

	// 创建并注册处理器
	taskHandler := service.NewTaskHandler(taskService)
	taskHandler.RegisterRoutes(r)

	// 启动服务器
	addr := fmt.Sprintf(":%d", config.GlobalConfig.Server.Port)
	if err := r.Run(addr); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}
