package utils

import (
	"context"
	"log"
	"runtime/debug"
)

// Recover 用于恢复协程中的 panic
func Recover(name string, ctx context.Context) {
	if err := recover(); err != nil {
		// 获取堆栈信息
		stack := debug.Stack()

		// 记录错误日志
		log.Printf("[PANIC] %s: %v\n%s", name, err, string(stack))

		// 这里可以添加告警通知，比如发送邮件、钉钉等
		// TODO: 实现告警通知
	}
}
