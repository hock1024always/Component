package main

import (
	"github.com/gin-gonic/gin"
	"gomail/config"
	"gomail/controllers"
)

func main() {
	config.InitDB() // 初始化数据库连接

	r := gin.Default()

	r.POST("/register", controllers.Register)
	r.POST("/verify", controllers.Verify)

	r.Run(":9999")
}
