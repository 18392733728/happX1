package utils

import (
	"fmt"
	"strings"

	"github.com/robfig/cron/v3"
)

// ParseCron 解析cron表达式
func ParseCron(spec string) (*cron.Schedule, error) {
	// 移除多余的空格
	spec = strings.TrimSpace(spec)

	// 检查是否为空
	if spec == "" {
		return nil, fmt.Errorf("cron表达式不能为空")
	}

	// 解析cron表达式
	schedule, err := cron.ParseStandard(spec)
	if err != nil {
		return nil, fmt.Errorf("解析cron表达式失败: %v", err)
	}

	return &schedule, nil
}

// ValidateCronSpec 验证cron表达式格式
func ValidateCronSpec(spec string) error {
	// 移除多余的空格
	spec = strings.TrimSpace(spec)

	// 检查是否为空
	if spec == "" {
		return fmt.Errorf("cron表达式不能为空")
	}

	// 检查基本格式
	parts := strings.Fields(spec)
	if len(parts) < 5 || len(parts) > 6 {
		return fmt.Errorf("cron表达式格式错误，应为5-6个字段")
	}

	// 检查每个字段的格式
	for i, part := range parts {
		if err := validateCronField(part, i); err != nil {
			return err
		}
	}

	return nil
}

// validateCronField 验证cron表达式的单个字段
func validateCronField(field string, position int) error {
	// 字段名称
	fieldNames := []string{"分钟", "小时", "日期", "月份", "星期", "秒"}
	if position >= len(fieldNames) {
		return fmt.Errorf("无效的字段位置: %d", position)
	}

	// 检查字段值
	if field == "*" {
		return nil
	}

	// 分割字段值
	values := strings.Split(field, ",")
	for _, value := range values {
		// 检查范围
		if strings.Contains(value, "-") {
			parts := strings.Split(value, "-")
			if len(parts) != 2 {
				return fmt.Errorf("%s字段范围格式错误: %s", fieldNames[position], value)
			}
			if err := validateCronValue(parts[0], position); err != nil {
				return err
			}
			if err := validateCronValue(parts[1], position); err != nil {
				return err
			}
			continue
		}

		// 检查步长
		if strings.Contains(value, "/") {
			parts := strings.Split(value, "/")
			if len(parts) != 2 {
				return fmt.Errorf("%s字段步长格式错误: %s", fieldNames[position], value)
			}
			if parts[0] != "*" {
				if err := validateCronValue(parts[0], position); err != nil {
					return err
				}
			}
			if err := validateCronValue(parts[1], position); err != nil {
				return err
			}
			continue
		}

		// 检查单个值
		if err := validateCronValue(value, position); err != nil {
			return err
		}
	}

	return nil
}

// validateCronValue 验证cron表达式的单个值
func validateCronValue(value string, position int) error {
	// 字段名称和范围
	fieldRanges := []struct {
		name  string
		min   int
		max   int
		names map[string]int // 用于月份和星期的名称映射
	}{
		{"分钟", 0, 59, nil},
		{"小时", 0, 23, nil},
		{"日期", 1, 31, nil},
		{"月份", 1, 12, map[string]int{
			"JAN": 1, "FEB": 2, "MAR": 3, "APR": 4, "MAY": 5, "JUN": 6,
			"JUL": 7, "AUG": 8, "SEP": 9, "OCT": 10, "NOV": 11, "DEC": 12,
		}},
		{"星期", 0, 6, map[string]int{
			"SUN": 0, "MON": 1, "TUE": 2, "WED": 3, "THU": 4, "FRI": 5, "SAT": 6,
		}},
		{"秒", 0, 59, nil},
	}

	if position >= len(fieldRanges) {
		return fmt.Errorf("无效的字段位置: %d", position)
	}

	rangeInfo := fieldRanges[position]

	// 检查是否为数字
	if value == "*" {
		return nil
	}

	// 检查是否为名称（月份或星期）
	if rangeInfo.names != nil {
		if num, ok := rangeInfo.names[strings.ToUpper(value)]; ok {
			if num < rangeInfo.min || num > rangeInfo.max {
				return fmt.Errorf("%s字段值超出范围: %s", rangeInfo.name, value)
			}
			return nil
		}
	}

	// 检查是否为数字
	var num int
	_, err := fmt.Sscanf(value, "%d", &num)
	if err != nil {
		return fmt.Errorf("%s字段值格式错误: %s", rangeInfo.name, value)
	}

	// 检查范围
	if num < rangeInfo.min || num > rangeInfo.max {
		return fmt.Errorf("%s字段值超出范围: %d (应为 %d-%d)", rangeInfo.name, num, rangeInfo.min, rangeInfo.max)
	}

	return nil
}
