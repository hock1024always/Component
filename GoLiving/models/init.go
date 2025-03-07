package models

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

// 创建数据库连接
func NewDB() {
	// 数据库连接字符串
	dsn := "root:212328@tcp(127.0.0.1:3306)/meeting?charset=utf8mb4&parseTime=True&loc=Local"
	// 使用gorm.Open函数打开数据库连接
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	// 如果连接失败，则抛出异常
	if err != nil {
		panic("数据库连接失败")
	}
	// 自动迁移模式
	db.AutoMigrate(&RoomBasic{}, &RoomUser{}, &UserBasic{})
	// 将数据库连接赋值给全局变量DB
	DB = db
}
