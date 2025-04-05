package scheduler

import (
	"context"
	"fmt"
	"gorm.io/gorm"
	"log"
	"os/exec"
	"time"

	"github.com/robfig/cron/v3"
	"happx1/internal/database"
	"happx1/internal/model"
	"happx1/pkg/utils"
)

type Scheduler struct {
	cron *cron.Cron
	db   *gorm.DB
}

func NewScheduler() *Scheduler {
	return &Scheduler{
		cron: cron.New(cron.WithSeconds()),
		db:   database.DB,
	}
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

	// 添加到调度器
	_, err := s.cron.AddFunc(task.Spec, func() {
		go func() {
			defer utils.Recover(fmt.Sprintf("Task-%d", task.ID), context.Background())
			s.ExecuteTask(task)
		}()
	})
	if err != nil {
		return err
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

	// 执行命令
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(task.Timeout)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sh", "-c", task.Command)
	output, err := cmd.CombinedOutput()

	// 更新任务日志
	taskLog.EndTime = time.Now()
	taskLog.Duration = int(taskLog.EndTime.Sub(taskLog.StartTime).Seconds())
	taskLog.Output = string(output)

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

	// 更新任务状态
	task.LastRunTime = taskLog.StartTime
	task.NextRunTime = s.cron.Entry(cron.EntryID(task.ID)).Next
	if err := s.db.Save(task).Error; err != nil {
		log.Printf("更新任务状态失败: %v", err)
	}
}
