package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"happx1/internal/database"
	"happx1/internal/router"
	"happx1/internal/scheduler"
	"happx1/internal/service"
)

func main() {
	// 初始化数据库连接
	db, err := database.InitDB()
	if err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}

	// 初始化调度器
	scheduler := scheduler.NewScheduler(db)

	// 初始化服务层
	taskService := service.NewTaskService(db, scheduler)

	// 初始化处理器
	taskHandler := service.NewTaskHandler(taskService)

	// 初始化路由
	r := gin.Default()
	router.RegisterRoutes(r, taskHandler)

	// 启动服务器
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("启动服务器失败: %v", err)
	}
}
