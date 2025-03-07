package config

import (
	"gomail/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	dsn := "root:212328@tcp(127.0.0.1:3306)/gomail?charset=utf8mb4&parseTime=True&loc=Local"

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database: " + err.Error())
	}

	DB.AutoMigrate(&models.User{})
}
