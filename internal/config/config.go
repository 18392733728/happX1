package config

import (
	"fmt"

	"github.com/spf13/viper"
	"happx1/internal/database"
)

type Config struct {
	MySQL database.MySQLConfig
	Redis database.RedisConfig
	Server struct {
		Port int
		Mode string
	}
}

var GlobalConfig Config

func Init() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("读取配置文件失败: %v", err)
	}

	if err := viper.Unmarshal(&GlobalConfig); err != nil {
		return fmt.Errorf("解析配置文件失败: %v", err)
	}

	return nil
} 