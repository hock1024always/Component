package db

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB // 全局变量

func InitDB(dsn string) error {
	log.Printf("Connecting to MySQL: %s", dsn)

	// 直接连接到指定的数据库
	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Printf("Failed to connect to MySQL: %v", err)
		return fmt.Errorf("failed to connect to MySQL: %w", err)
	}

	if DB == nil {
		return fmt.Errorf("DB is nil after initialization")
	}

	log.Println("Connected to MySQL successfully")

	// 设置数据库连接池参数
	sqlDB, err := DB.DB()
	if err != nil {
		log.Printf("Failed to get database connection: %v", err)
		return fmt.Errorf("failed to get database connection: %w", err)
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// 自动迁移表结构
	log.Println("Running AutoMigrate...")
	if err := DB.AutoMigrate(&User{}, &Message{}).Error; err != nil {
		log.Printf("Failed to migrate database: %v", err)
		return fmt.Errorf("failed to migrate database: %w", err)
	}
	log.Println("Database connection established")
	return nil
}
