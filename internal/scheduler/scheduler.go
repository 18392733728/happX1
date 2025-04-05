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
}
