package scheduler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
	"happx1/internal/database"
	"happx1/internal/model"
	"happx1/pkg/utils"
)

// Scheduler 调度器
type Scheduler struct {
	db   *database.DB
	cron *cron.Cron
}

// NewScheduler 创建调度器
func NewScheduler(db *database.DB) *Scheduler {
	s := &Scheduler{
		db:   db,
		cron: cron.New(cron.WithSeconds()),
	}
	s.cron.Start()
	return s
}

// Start 启动调度器
func (s *Scheduler) Start() error {
	// 自动迁移数据库表
	if err := s.db.AutoMigrate(&model.Task{}, &model.TaskLog{}); err != nil {
		return fmt.Errorf("数据库迁移失败: %v", err)
	}

	// 加载所有启用的任务
	var tasks []model.Task
	if err := s.db.Where("status = ?", 1).Find(&tasks).Error; err != nil {
		return fmt.Errorf("加载任务失败: %v", err)
	}

	// 添加任务到调度器
	for _, task := range tasks {
		if err := s.AddTask(&task); err != nil {
			log.Printf("添加任务失败 [%s]: %v", task.Name, err)
			continue
		}
	}

	// 启动调度器
	s.cron.Start()
	return nil
}

// Stop 停止调度器
func (s *Scheduler) Stop() {
	s.cron.Stop()
}

// AddTask 添加任务
func (s *Scheduler) AddTask(task *model.Task) error {
	// 检查任务是否已存在
	var count int64
	if err := s.db.Model(&model.Task{}).Where("name = ?", task.Name).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("任务已存在: %s", task.Name)
	}

	// 保存任务到数据库
	if err := s.db.Create(task).Error; err != nil {
		return err
	}

	// 根据任务类型添加到调度器
	switch task.Type {
	case model.TaskTypeOnce:
		// 解析执行时间
		execTime, err := time.Parse(time.RFC3339, task.Spec)
		if err != nil {
			return fmt.Errorf("解析执行时间失败: %v", err)
		}

		// 计算延迟时间
		delay := execTime.Sub(time.Now())
		if delay < 0 {
			return fmt.Errorf("执行时间已过期")
		}

		// 启动一次性任务
		go func() {
			defer utils.Recover(fmt.Sprintf("OnceTask-%d", task.ID), context.Background())
			time.Sleep(delay)
			s.ExecuteTask(task)
		}()

	case model.TaskTypeCron:
		// 添加到 cron 调度器
		_, err := s.cron.AddFunc(task.Spec, func() {
			go func() {
				defer utils.Recover(fmt.Sprintf("CronTask-%d", task.ID), context.Background())
				s.ExecuteTask(task)
			}()
		})
		if err != nil {
			return err
		}

	default:
		return fmt.Errorf("不支持的任务类型: %d", task.Type)
	}

	return nil
}

// ExecuteTask 执行任务
func (s *Scheduler) ExecuteTask(task *model.Task) {
	// 创建任务日志
	taskLog := &model.TaskLog{
		TaskID:    task.ID,
		StartTime: time.Now(),
		Status:    0,
	}

	// 执行任务（带重试）
	var output string
	var err error
	for i := 0; i <= task.RetryTimes; i++ {
		// 如果不是第一次尝试，等待重试延迟
		if i > 0 {
			time.Sleep(time.Duration(task.RetryDelay) * time.Second)
			log.Printf("任务 %d 第 %d 次重试", task.ID, i)
		}

		// 创建带超时的上下文
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(task.Timeout)*time.Second)
		defer cancel()

		// 根据执行类型执行不同的任务
		switch task.ExecType {
		case model.ExecTypeShell:
			output, err = s.executeShell(ctx, task)
		case model.ExecTypeHTTP:
			output, err = s.executeHTTP(ctx, task)
		default:
			err = fmt.Errorf("不支持的执行类型: %d", task.ExecType)
		}

		// 如果执行成功，跳出重试循环
		if err == nil {
			break
		}

		// 检查是否是超时错误
		if ctx.Err() == context.DeadlineExceeded {
			log.Printf("任务 %d 执行超时", task.ID)
			err = fmt.Errorf("任务执行超时（%d秒）", task.Timeout)
		}

		// 记录重试次数
		taskLog.RetryCount = i
	}

	// 更新任务日志
	taskLog.EndTime = time.Now()
	taskLog.Duration = int(taskLog.EndTime.Sub(taskLog.StartTime).Seconds())
	taskLog.Output = output

	if err != nil {
		taskLog.Status = 0
		taskLog.Error = err.Error()
	} else {
		taskLog.Status = 1
	}

	// 保存日志
	if err := s.db.Create(taskLog).Error; err != nil {
		log.Printf("保存任务日志失败: %v", err)
	}

	// 更新任务统计信息
	var stats model.TaskStats
	result := s.db.Where("task_id = ?", task.ID).First(&stats)

	// 如果统计记录不存在，创建新记录
	if result.Error != nil {
		stats = model.TaskStats{
			TaskID: task.ID,
		}
	}

	// 更新统计信息
	stats.TotalRuns++
	stats.TotalDuration += taskLog.Duration
	stats.AvgDuration = float64(stats.TotalDuration) / float64(stats.TotalRuns)
	stats.RetryCount += taskLog.RetryCount
	stats.UpdatedAt = time.Now()

	if taskLog.Status == 1 {
		stats.SuccessRuns++
		stats.LastSuccess = taskLog.EndTime
	} else {
		stats.FailedRuns++
		stats.LastFailure = taskLog.EndTime
		stats.LastError = taskLog.Error
	}

	// 检查是否是超时错误
	if strings.Contains(taskLog.Error, "任务执行超时") {
		stats.TimeoutCount++
	}

	// 保存或更新统计信息
	if result.Error != nil {
		if err := s.db.Create(&stats).Error; err != nil {
			log.Printf("创建任务统计信息失败: %v", err)
		}
	} else {
		if err := s.db.Save(&stats).Error; err != nil {
			log.Printf("更新任务统计信息失败: %v", err)
		}
	}

	// 更新任务状态
	task.LastRunTime = taskLog.StartTime
	if task.Type == model.TaskTypeOnce {
		// 一次性任务执行完成后禁用
		task.Status = 0
	} else {
		// 循环任务更新下次执行时间
		task.NextRunTime = s.cron.Entry(cron.EntryID(task.ID)).Next
	}
	if err := s.db.Save(task).Error; err != nil {
		log.Printf("更新任务状态失败: %v", err)
	}

	// 发送回调通知
	if task.CallbackURL != "" {
		go s.sendCallback(task, taskLog)
	}
}

// executeShell 执行 Shell 命令
func (s *Scheduler) executeShell(ctx context.Context, task *model.Task) (string, error) {
	cmd := exec.CommandContext(ctx, "sh", "-c", task.Command)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// executeHTTP 执行 HTTP 请求
func (s *Scheduler) executeHTTP(ctx context.Context, task *model.Task) (string, error) {
	// 创建 HTTP 客户端
	client := &http.Client{
		Timeout: time.Duration(task.Timeout) * time.Second,
	}

	// 解析请求头
	var headers map[string]string
	if task.Headers != "" {
		if err := json.Unmarshal([]byte(task.Headers), &headers); err != nil {
			return "", fmt.Errorf("解析请求头失败: %v", err)
		}
	}

	// 创建请求
	var reqBody io.Reader
	if task.Body != "" {
		reqBody = bytes.NewBufferString(task.Body)
	}

	req, err := http.NewRequestWithContext(ctx, task.Method, task.Command, reqBody)
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %v", err)
	}

	// 检查响应状态码
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return string(body), fmt.Errorf("请求失败，状态码: %d", resp.StatusCode)
	}

	return string(body), nil
}

// sendCallback 发送回调通知
func (s *Scheduler) sendCallback(task *model.Task, taskLog *model.TaskLog) {
	defer utils.Recover("sendCallback", context.Background())

	// 准备回调数据
	callbackData := map[string]interface{}{
		"task_id":    task.ID,
		"name":       task.Name,
		"status":     taskLog.Status,
		"output":     taskLog.Output,
		"error":      taskLog.Error,
		"start_time": taskLog.StartTime,
		"end_time":   taskLog.EndTime,
		"duration":   taskLog.Duration,
	}

	// 替换回调体模板中的变量
	body := task.CallbackBody
	if body != "" {
		for key, value := range callbackData {
			placeholder := fmt.Sprintf("${%s}", key)
			body = strings.ReplaceAll(body, placeholder, fmt.Sprintf("%v", value))
		}
	}

	// 解析请求头
	var headers map[string]string
	if task.CallbackHeaders != "" {
		if err := json.Unmarshal([]byte(task.CallbackHeaders), &headers); err != nil {
			log.Printf("解析回调请求头失败: %v", err)
			return
		}
	}

	// 创建请求
	var reqBody io.Reader
	if body != "" {
		reqBody = bytes.NewBufferString(body)
	}

	method := task.CallbackMethod
	if method == "" {
		method = "POST"
	}

	req, err := http.NewRequest(method, task.CallbackURL, reqBody)
	if err != nil {
		log.Printf("创建回调请求失败: %v", err)
		return
	}

	// 设置请求头
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	// 发送请求
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("发送回调请求失败: %v", err)
		return
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Printf("回调请求失败，状态码: %d", resp.StatusCode)
		return
	}

	log.Printf("任务 %d 回调通知发送成功", task.ID)
}

// RemoveTask 从调度器中移除任务
func (s *Scheduler) RemoveTask(task *model.Task) error {
	// 从cron调度器中移除
	if task.Type == model.TaskTypeCron {
		s.cron.Remove(cron.EntryID(task.ID))
	}

	// 更新任务状态
	task.Status = 0
	if err := s.db.Save(task).Error; err != nil {
		return fmt.Errorf("更新任务状态失败: %v", err)
	}

	return nil
}
