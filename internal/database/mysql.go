package database

import (
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

type MySQLConfig struct {
	Host            string
	Port            int
	Username        string
	Password        string
	Database        string
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime int
}

func InitMySQL(config *MySQLConfig) error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.Username,
		config.Password,
		config.Host,
		config.Port,
		config.Database,
	)

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return fmt.Errorf("连接MySQL失败: %v", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("获取MySQL连接池失败: %v", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(config.ConnMaxLifetime) * time.Second)

	return nil
} 