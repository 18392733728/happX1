package service

import (
	"context"
	"fmt"
	"gorm.io/gorm"
	"happx1/internal/model"
	"happx1/internal/scheduler"
	"happx1/pkg/utils"
)

type TaskService struct {
	scheduler *scheduler.Scheduler
	db        *gorm.DB
}

func NewTaskService(scheduler *scheduler.Scheduler, db *gorm.DB) *TaskService {
	return &TaskService{
		scheduler: scheduler,
		db:        db,
	}
}

// CreateTask 创建任务
func (s *TaskService) CreateTask(task *model.Task) error {
	return s.scheduler.AddTask(task)
}

// ListTasks 获取任务列表
func (s *TaskService) ListTasks() ([]model.Task, error) {
	var tasks []model.Task
	if err := s.db.Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

// GetTask 获取任务详情
func (s *TaskService) GetTask(id uint) (*model.Task, error) {
	var task model.Task
	if err := s.db.First(&task, id).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

// UpdateTask 更新任务
func (s *TaskService) UpdateTask(task *model.Task) error {
	return s.db.Save(task).Error
}

// DeleteTask 删除任务
func (s *TaskService) DeleteTask(id uint) error {
	return s.db.Delete(&model.Task{}, id).Error
}

// RunTask 立即执行任务
func (s *TaskService) RunTask(task *model.Task) {
	go func() {
		defer utils.Recover(fmt.Sprintf("ManualTask-%d", task.ID), context.Background())
		s.scheduler.ExecuteTask(task)
	}()
}

// GetTaskLogs 获取任务执行日志
func (s *TaskService) GetTaskLogs(taskID uint) ([]model.TaskLog, error) {
	var logs []model.TaskLog
	if err := s.db.Where("task_id = ?", taskID).Order("created_at desc").Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}
