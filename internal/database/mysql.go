package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"happx1/internal/model"
)

// MySQLConfig MySQL配置结构体
type MySQLConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

// DB 数据库连接
type DB struct {
	*gorm.DB
}

// InitDB 初始化数据库连接
func InitDB(config *MySQLConfig) (*DB, error) {
	// 配置数据库连接
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.DBName,
	)

	// 配置GORM日志
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second, // 慢SQL阈值
			LogLevel:                  logger.Info, // 日志级别
			IgnoreRecordNotFoundError: true,        // 忽略记录未找到错误
			Colorful:                  true,        // 彩色打印
		},
	)

	// 连接数据库
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %v", err)
	}

	// 获取底层的sqlDB
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("获取sqlDB失败: %v", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(10)           // 设置空闲连接池中连接的最大数量
	sqlDB.SetMaxOpenConns(100)          // 设置打开数据库连接的最大数量
	sqlDB.SetConnMaxLifetime(time.Hour) // 设置连接可复用的最大时间

	// 自动迁移数据库表
	if err := db.AutoMigrate(
		&model.Task{},
		&model.TaskLog{},
		&model.TaskStats{},
	); err != nil {
		return nil, fmt.Errorf("自动迁移数据库表失败: %v", err)
	}

	return &DB{db}, nil
}

// Save 保存记录
func (db *DB) Save(value interface{}) *gorm.DB {
	return db.DB.Save(value)
}

// Create 创建记录
func (db *DB) Create(value interface{}) *gorm.DB {
	return db.DB.Create(value)
}

// Delete 删除记录
func (db *DB) Delete(value interface{}) *gorm.DB {
	return db.DB.Delete(value)
}

// First 获取第一条记录
func (db *DB) First(dest interface{}, conds ...interface{}) *gorm.DB {
	return db.DB.First(dest, conds...)
}

// Find 获取所有记录
func (db *DB) Find(dest interface{}, conds ...interface{}) *gorm.DB {
	return db.DB.Find(dest, conds...)
}

// Where 条件查询
func (db *DB) Where(query interface{}, args ...interface{}) *gorm.DB {
	return db.DB.Where(query, args...)
}

// Order 排序
func (db *DB) Order(value interface{}) *gorm.DB {
	return db.DB.Order(value)
}
