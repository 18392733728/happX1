package model

import (
	"time"

	"gorm.io/gorm"
)

// TaskType 任务类型
type TaskType int

const (
	TaskTypeOnce TaskType = iota + 1 // 一次性任务
	TaskTypeCron                     // 循环任务
)

// Task 定时任务模型
type Task struct {
	gorm.Model
	Name        string    `gorm:"type:varchar(100);not null;unique" json:"name"`  // 任务名称
	Type        TaskType  `gorm:"type:tinyint;not null;default:1" json:"type"`    // 任务类型：1-一次性任务，2-循环任务
	Spec        string    `gorm:"type:varchar(100);not null" json:"spec"`         // cron 表达式或执行时间
	Command     string    `gorm:"type:text;not null" json:"command"`              // 执行的命令
	Status      int       `gorm:"type:tinyint;not null;default:1" json:"status"`  // 状态：1-启用，0-禁用
	LastRunTime time.Time `json:"last_run_time"`                                  // 上次运行时间
	NextRunTime time.Time `json:"next_run_time"`                                  // 下次运行时间
	Timeout     int       `gorm:"type:int;not null;default:60" json:"timeout"`    // 超时时间（秒）
	RetryTimes  int       `gorm:"type:int;not null;default:3" json:"retry_times"` // 重试次数
	RetryDelay  int       `gorm:"type:int;not null;default:5" json:"retry_delay"` // 重试延迟（秒）
	Description string    `gorm:"type:varchar(500)" json:"description"`           // 任务描述
}

// TaskLog 任务执行日志
type TaskLog struct {
	gorm.Model
	TaskID     uint      `gorm:"not null" json:"task_id"`                        // 任务ID
	Status     int       `gorm:"type:tinyint;not null" json:"status"`            // 状态：1-成功，0-失败
	StartTime  time.Time `gorm:"not null" json:"start_time"`                     // 开始时间
	EndTime    time.Time `json:"end_time"`                                       // 结束时间
	Duration   int       `gorm:"type:int;not null" json:"duration"`              // 执行时长（秒）
	Output     string    `gorm:"type:text" json:"output"`                        // 输出结果
	Error      string    `gorm:"type:text" json:"error"`                         // 错误信息
	RetryCount int       `gorm:"type:int;not null;default:0" json:"retry_count"` // 重试次数
}
