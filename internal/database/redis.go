package database

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
)

var RedisClient *redis.Client

type RedisConfig struct {
	Host         string
	Port         int
	Password     string
	DB           int
	PoolSize     int
	MinIdleConns int
}

func InitRedis(config *RedisConfig) error {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password:     config.Password,
		DB:           config.DB,
		PoolSize:     config.PoolSize,
		MinIdleConns: config.MinIdleConns,
	})

	// 测试连接
	ctx := context.Background()
	if err := RedisClient.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("连接Redis失败: %v", err)
	}

	return nil
} 