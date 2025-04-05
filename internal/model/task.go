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

// ExecType 执行类型
type ExecType int

const (
	ExecTypeShell ExecType = iota + 1 // Shell 命令
	ExecTypeHTTP                      // HTTP 接口
)

// Task 定时任务模型
type Task struct {
	gorm.Model
	Name        string    `gorm:"type:varchar(100);not null;unique" json:"name"`         // 任务名称
	Type        TaskType  `gorm:"type:tinyint;not null;default:1" json:"type"`           // 任务类型：1-一次性任务，2-循环任务
	ExecType    ExecType  `gorm:"type:tinyint;not null;default:1" json:"exec_type"`      // 执行类型：1-Shell命令，2-HTTP接口
	Spec        string    `gorm:"type:varchar(100);not null" json:"spec"`                // cron 表达式或执行时间
	Command     string    `gorm:"type:text;not null" json:"command"`                     // 执行的命令或URL
	Method      string    `gorm:"type:varchar(10);default:'GET'" json:"method"`          // HTTP方法：GET, POST, PUT, DELETE
	Headers     string    `gorm:"type:text" json:"headers"`                              // HTTP请求头，JSON格式
	Body        string    `gorm:"type:text" json:"body"`                                 // HTTP请求体，JSON格式
	Status      int       `gorm:"type:tinyint;not null;default:1" json:"status"`        // 状态：1-启用，0-禁用
	LastRunTime time.Time `json:"last_run_time"`                                         // 上次运行时间
	NextRunTime time.Time `json:"next_run_time"`                                         // 下次运行时间
	Timeout     int       `gorm:"type:int;not null;default:60" json:"timeout"`          // 超时时间（秒）
	RetryTimes  int       `gorm:"type:int;not null;default:3" json:"retry_times"`       // 重试次数
	RetryDelay  int       `gorm:"type:int;not null;default:5" json:"retry_delay"`       // 重试延迟（秒）
	Description string    `gorm:"type:varchar(500)" json:"description"`                  // 任务描述
	CallbackURL string    `gorm:"type:varchar(500)" json:"callback_url"`                 // 回调通知URL
	CallbackMethod string `gorm:"type:varchar(10)" json:"callback_method"`               // 回调请求方法
	CallbackHeaders string `gorm:"type:text" json:"callback_headers"`                    // 回调请求头（JSON格式）
	CallbackBody string    `gorm:"type:text" json:"callback_body"`                       // 回调请求体模板（支持变量替换）
}

// TaskLog 任务执行日志
type TaskLog struct {
	gorm.Model
	TaskID    uint      `gorm:"not null" json:"task_id"`                              // 任务ID
	Status    int       `gorm:"type:tinyint;not null" json:"status"`                  // 状态：1-成功，0-失败
	StartTime time.Time `gorm:"not null" json:"start_time"`                           // 开始时间
	EndTime   time.Time `json:"end_time"`                                             // 结束时间
	Duration  int       `gorm:"type:int;not null" json:"duration"`                    // 执行时长（秒）
	Output    string    `gorm:"type:text" json:"output"`                              // 输出结果
	Error     string    `gorm:"type:text" json:"error"`                               // 错误信息
	RetryCount int      `gorm:"type:int;not null;default:0" json:"retry_count"`       // 重试次数
}

// TaskStats 任务执行统计
type TaskStats struct {
	gorm.Model
	TaskID        uint      `gorm:"not null;uniqueIndex" json:"task_id"`                // 任务ID
	TotalRuns     int       `gorm:"not null;default:0" json:"total_runs"`               // 总执行次数
	SuccessRuns   int       `gorm:"not null;default:0" json:"success_runs"`             // 成功次数
	FailedRuns    int       `gorm:"not null;default:0" json:"failed_runs"`              // 失败次数
	TotalDuration int       `gorm:"not null;default:0" json:"total_duration"`           // 总执行时长（秒）
	AvgDuration   float64   `gorm:"not null;default:0" json:"avg_duration"`             // 平均执行时长（秒）
	LastSuccess   time.Time `json:"last_success"`                                        // 最后一次成功时间
	LastFailure   time.Time `json:"last_failure"`                                        // 最后一次失败时间
	LastError     string    `gorm:"type:text" json:"last_error"`                        // 最后一次错误信息
	RetryCount    int       `gorm:"not null;default:0" json:"retry_count"`              // 总重试次数
	TimeoutCount  int       `gorm:"not null;default:0" json:"timeout_count"`            // 超时次数
	UpdatedAt     time.Time `json:"updated_at"`                                          // 最后更新时间
}
