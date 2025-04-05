package service

import (
	"context"
	"encoding/json"
	"fmt"
	"gorm.io/gorm"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"happx1/internal/database"
	"happx1/internal/model"
	"happx1/internal/scheduler"
	"happx1/pkg/utils"
)

// TaskService 任务服务
type TaskService struct {
	db        *database.DB
	scheduler *scheduler.Scheduler
}

// NewTaskService 创建任务服务
func NewTaskService(db *database.DB, scheduler *scheduler.Scheduler) *TaskService {
	return &TaskService{
		db:        db,
		scheduler: scheduler,
	}
}

// TaskHandler 任务处理器
type TaskHandler struct {
	taskService *TaskService
}

// NewTaskHandler 创建任务处理器
func NewTaskHandler(taskService *TaskService) *TaskHandler {
	return &TaskHandler{
		taskService: taskService,
	}
}

// RegisterRoutes 注册路由
func (h *TaskHandler) RegisterRoutes(r *gin.Engine) {
	tasks := r.Group("/api/tasks")
	{
		// 创建任务
		tasks.POST("", h.CreateTask)
		// 获取任务列表
		tasks.GET("", h.ListTasks)
		// 获取任务详情
		tasks.GET("/:id", h.GetTask)
		// 更新任务
		tasks.POST("/:id/update", h.UpdateTask)
		// 删除任务
		tasks.POST("/:id/delete", h.DeleteTask)
		// 立即执行任务
		tasks.POST("/:id/run", h.RunTask)
		// 获取任务执行日志
		tasks.GET("/:id/logs", h.GetTaskLogs)
		// 获取任务统计信息
		tasks.GET("/stats/:id", h.GetTaskStats)
		// 获取所有任务统计信息
		tasks.GET("/stats", h.GetAllTaskStats)
	}
}

// CreateTask 创建任务
func (h *TaskHandler) CreateTask(c *gin.Context) {
	var task model.Task
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.taskService.CreateTask(&task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, task)
}

// ListTasks 获取任务列表
func (h *TaskHandler) ListTasks(c *gin.Context) {
	tasks, err := h.taskService.ListTasks()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tasks)
}

// GetTask 获取任务详情
func (h *TaskHandler) GetTask(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的任务ID"})
		return
	}

	task, err := h.taskService.GetTask(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "任务不存在"})
		return
	}

	c.JSON(http.StatusOK, task)
}

// UpdateTask 更新任务
func (h *TaskHandler) UpdateTask(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的任务ID"})
		return
	}

	task, err := h.taskService.GetTask(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "任务不存在"})
		return
	}

	if err := c.ShouldBindJSON(task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.taskService.UpdateTask(task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, task)
}

// DeleteTask 删除任务
func (h *TaskHandler) DeleteTask(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的任务ID"})
		return
	}

	if err := h.taskService.DeleteTask(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// RunTask 立即执行任务
func (h *TaskHandler) RunTask(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的任务ID"})
		return
	}

	task, err := h.taskService.GetTask(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "任务不存在"})
		return
	}

	h.taskService.RunTask(task)
	c.Status(http.StatusAccepted)
}

// GetTaskLogs 获取任务执行日志
func (h *TaskHandler) GetTaskLogs(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的任务ID"})
		return
	}

	logs, err := h.taskService.GetTaskLogs(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, logs)
}

// GetTaskStats 获取任务统计信息
func (h *TaskHandler) GetTaskStats(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的任务ID"})
		return
	}

	stats, err := h.taskService.GetTaskStats(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetAllTaskStats 获取所有任务的统计信息
func (h *TaskHandler) GetAllTaskStats(c *gin.Context) {
	stats, err := h.taskService.GetAllTaskStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// CreateTask 创建任务
func (s *TaskService) CreateTask(task *model.Task) error {
	// 验证任务类型
	if task.Type != model.TaskTypeOnce && task.Type != model.TaskTypeCron {
		return fmt.Errorf("不支持的任务类型: %d", task.Type)
	}

	// 验证执行类型
	if task.ExecType != model.ExecTypeShell && task.ExecType != model.ExecTypeHTTP {
		return fmt.Errorf("不支持的执行类型: %d", task.ExecType)
	}

	// 验证执行类型相关的字段
	switch task.ExecType {
	case model.ExecTypeShell:
		if task.Command == "" {
			return fmt.Errorf("Shell命令不能为空")
		}
	case model.ExecTypeHTTP:
		if task.Command == "" {
			return fmt.Errorf("HTTP URL不能为空")
		}
		if task.Method == "" {
			task.Method = "GET"
		}
		// 验证请求头格式
		if task.Headers != "" {
			var headers map[string]string
			if err := json.Unmarshal([]byte(task.Headers), &headers); err != nil {
				return fmt.Errorf("请求头格式错误: %v", err)
			}
		}
	}

	// 验证执行时间
	if task.Type == model.TaskTypeOnce {
		execTime, err := time.Parse(time.RFC3339, task.Spec)
		if err != nil {
			return fmt.Errorf("执行时间格式错误: %v", err)
		}
		if execTime.Before(time.Now()) {
			return fmt.Errorf("执行时间不能早于当前时间")
		}
		task.NextRunTime = execTime
	} else {
		// 验证 cron 表达式
		if _, err := utils.ParseCron(task.Spec); err != nil {
			return fmt.Errorf("cron表达式格式错误: %v", err)
		}
	}

	// 验证回调相关字段
	if task.CallbackURL != "" {
		// 验证回调URL格式
		if _, err := url.Parse(task.CallbackURL); err != nil {
			return fmt.Errorf("回调URL格式错误: %v", err)
		}

		// 验证回调请求方法
		if task.CallbackMethod != "" && task.CallbackMethod != "GET" && task.CallbackMethod != "POST" {
			return fmt.Errorf("不支持的回调请求方法: %s", task.CallbackMethod)
		}

		// 验证回调请求头格式
		if task.CallbackHeaders != "" {
			var headers map[string]string
			if err := json.Unmarshal([]byte(task.CallbackHeaders), &headers); err != nil {
				return fmt.Errorf("回调请求头格式错误: %v", err)
			}
		}

		// 验证回调体模板中的变量
		if task.CallbackBody != "" {
			// 检查变量格式
			matches := regexp.MustCompile(`\${[^}]+}`).FindAllString(task.CallbackBody, -1)
			for _, match := range matches {
				// 去掉${}，获取变量名
				varName := match[2 : len(match)-1]
				// 检查是否是支持的变量
				supportedVars := map[string]bool{
					"task_id":    true,
					"name":       true,
					"status":     true,
					"output":     true,
					"error":      true,
					"start_time": true,
					"end_time":   true,
					"duration":   true,
				}
				if !supportedVars[varName] {
					return fmt.Errorf("不支持的回调变量: %s", varName)
				}
			}
		}
	}

	// 设置默认值
	if task.Timeout <= 0 {
		task.Timeout = 60
	}
	if task.RetryTimes < 0 {
		task.RetryTimes = 3
	}
	if task.RetryDelay < 0 {
		task.RetryDelay = 5
	}

	// 保存任务
	if err := s.db.Create(task).Error; err != nil {
		return fmt.Errorf("创建任务失败: %v", err)
	}

	// 添加到调度器
	if err := s.scheduler.AddTask(task); err != nil {
		return fmt.Errorf("添加任务到调度器失败: %v", err)
	}

	return nil
}

// UpdateTask 更新任务
func (s *TaskService) UpdateTask(task *model.Task) error {
	// 获取原任务
	var oldTask model.Task
	if err := s.db.First(&oldTask, task.ID).Error; err != nil {
		return fmt.Errorf("任务不存在: %v", err)
	}

	// 验证任务类型
	if task.Type != model.TaskTypeOnce && task.Type != model.TaskTypeCron {
		return fmt.Errorf("不支持的任务类型: %d", task.Type)
	}

	// 验证执行类型
	if task.ExecType != model.ExecTypeShell && task.ExecType != model.ExecTypeHTTP {
		return fmt.Errorf("不支持的执行类型: %d", task.ExecType)
	}

	// 验证执行类型相关的字段
	switch task.ExecType {
	case model.ExecTypeShell:
		if task.Command == "" {
			return fmt.Errorf("Shell命令不能为空")
		}
	case model.ExecTypeHTTP:
		if task.Command == "" {
			return fmt.Errorf("HTTP URL不能为空")
		}
		if task.Method == "" {
			task.Method = "GET"
		}
		// 验证请求头格式
		if task.Headers != "" {
			var headers map[string]string
			if err := json.Unmarshal([]byte(task.Headers), &headers); err != nil {
				return fmt.Errorf("请求头格式错误: %v", err)
			}
		}
	}

	// 验证执行时间
	if task.Type == model.TaskTypeOnce {
		execTime, err := time.Parse(time.RFC3339, task.Spec)
		if err != nil {
			return fmt.Errorf("执行时间格式错误: %v", err)
		}
		if execTime.Before(time.Now()) {
			return fmt.Errorf("执行时间不能早于当前时间")
		}
		task.NextRunTime = execTime
	} else {
		// 验证 cron 表达式
		if _, err := utils.ParseCron(task.Spec); err != nil {
			return fmt.Errorf("cron表达式格式错误: %v", err)
		}
	}

	// 验证回调相关字段
	if task.CallbackURL != "" {
		// 验证回调URL格式
		if _, err := url.Parse(task.CallbackURL); err != nil {
			return fmt.Errorf("回调URL格式错误: %v", err)
		}

		// 验证回调请求方法
		if task.CallbackMethod != "" && task.CallbackMethod != "GET" && task.CallbackMethod != "POST" {
			return fmt.Errorf("不支持的回调请求方法: %s", task.CallbackMethod)
		}

		// 验证回调请求头格式
		if task.CallbackHeaders != "" {
			var headers map[string]string
			if err := json.Unmarshal([]byte(task.CallbackHeaders), &headers); err != nil {
				return fmt.Errorf("回调请求头格式错误: %v", err)
			}
		}

		// 验证回调体模板中的变量
		if task.CallbackBody != "" {
			// 检查变量格式
			matches := regexp.MustCompile(`\${[^}]+}`).FindAllString(task.CallbackBody, -1)
			for _, match := range matches {
				// 去掉${}，获取变量名
				varName := match[2 : len(match)-1]
				// 检查是否是支持的变量
				supportedVars := map[string]bool{
					"task_id":    true,
					"name":       true,
					"status":     true,
					"output":     true,
					"error":      true,
					"start_time": true,
					"end_time":   true,
					"duration":   true,
				}
				if !supportedVars[varName] {
					return fmt.Errorf("不支持的回调变量: %s", varName)
				}
			}
		}
	}

	// 设置默认值
	if task.Timeout <= 0 {
		task.Timeout = 60
	}
	if task.RetryTimes < 0 {
		task.RetryTimes = 3
	}
	if task.RetryDelay < 0 {
		task.RetryDelay = 5
	}

	// 更新任务
	if err := s.db.Save(task).Error; err != nil {
		return fmt.Errorf("更新任务失败: %v", err)
	}

	// 如果任务正在运行，需要重新添加到调度器
	if task.Status == 1 {
		if err := s.scheduler.AddTask(task); err != nil {
			return fmt.Errorf("更新任务到调度器失败: %v", err)
		}
	}

	return nil
}

// ListTasks 获取任务列表
func (s *TaskService) ListTasks() ([]model.Task, error) {
	var tasks []model.Task
	if err := s.db.Find(&tasks).Error; err != nil {
		return nil, fmt.Errorf("获取任务列表失败: %v", err)
	}
	return tasks, nil
}

// GetTask 获取任务详情
func (s *TaskService) GetTask(id uint) (*model.Task, error) {
	var task model.Task
	if err := s.db.First(&task, id).Error; err != nil {
		return nil, fmt.Errorf("获取任务详情失败: %v", err)
	}
	return &task, nil
}

// DeleteTask 删除任务
func (s *TaskService) DeleteTask(id uint) error {
	// 获取任务
	task, err := s.GetTask(id)
	if err != nil {
		return err
	}

	// 从调度器中移除
	if task.Status == 1 {
		if err := s.scheduler.RemoveTask(task); err != nil {
			return fmt.Errorf("从调度器移除任务失败: %v", err)
		}
	}

	// 删除任务
	if err := s.db.Delete(task).Error; err != nil {
		return fmt.Errorf("删除任务失败: %v", err)
	}

	return nil
}

// RunTask 立即执行任务
func (s *TaskService) RunTask(task *model.Task) {
	go func() {
		defer utils.Recover("RunTask", context.Background())
		s.scheduler.ExecuteTask(task)
	}()
}

// GetTaskLogs 获取任务执行日志
func (s *TaskService) GetTaskLogs(taskID uint) ([]model.TaskLog, error) {
	var logs []model.TaskLog
	if err := s.db.Where("task_id = ?", taskID).Order("start_time DESC").Find(&logs).Error; err != nil {
		return nil, fmt.Errorf("获取任务日志失败: %v", err)
	}
	return logs, nil
}

// GetTaskStats 获取任务统计信息
func (s *TaskService) GetTaskStats(taskID uint) (*model.TaskStats, error) {
	var stats model.TaskStats
	if err := s.db.Where("task_id = ?", taskID).First(&stats).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("任务统计信息不存在")
		}
		return nil, fmt.Errorf("获取任务统计信息失败: %v", err)
	}
	return &stats, nil
}

// GetAllTaskStats 获取所有任务的统计信息
func (s *TaskService) GetAllTaskStats() ([]model.TaskStats, error) {
	var stats []model.TaskStats
	if err := s.db.Find(&stats).Error; err != nil {
		return nil, fmt.Errorf("获取任务统计信息失败: %v", err)
	}
	return stats, nil
}
