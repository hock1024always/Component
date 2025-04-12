package main

import (
	"chatroom/db"
	"github.com/gin-gonic/gin"
	"log"
)

//func main() {
//	// 初始化数据库
//	dsn := "root:212328@(127.0.0.1:3306)/chatroom?charset=utf8mb4&parseTime=True&loc=Local"
//	if err := db.InitDB(dsn); err != nil {
//		log.Fatalf("Failed to initialize database: %v", err)
//	}
//	// 初始化Gin
//	r := gin.Default()
//
//	r.LoadHTMLGlob("templates/*.html")
//	r.Static("/static", "./static")
//
//	// 路由
//	r.GET("/", func(c *gin.Context) {
//		c.Redirect(http.StatusMovedPermanently, "/login")
//	})
//	r.GET("/login", func(c *gin.Context) {
//		c.HTML(http.StatusOK, "login.html", nil)
//	})
//	r.GET("/register", func(c *gin.Context) {
//		c.HTML(http.StatusOK, "register.html", nil)
//	})
//	r.GET("/chat", func(c *gin.Context) {
//		c.HTML(http.StatusOK, "chat.html", nil)
//	})
//	// 用户注册和登录
//	r.POST("/register", handlers.Register)
//	r.POST("/login", handlers.Login)
//
//	// 消息相关接口
//	r.POST("/message", handlers.SendMessage)
//	r.GET("/messages/:receiverID", handlers.GetMessages)
//
//	// 启动服务
//	r.Run(":8080")
//}

func main() {
	// 初始化数据库
	dsn := "root:212328@(127.0.0.1:3306)/chatroom?charset=utf8mb4&parseTime=True&loc=Local"
	if err := db.InitDB(dsn); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// 初始化 Gin
	r := gin.Default()

	// 注册路由
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Hello, World!"})
	})

	// 启动服务
	log.Println("Starting server on :8080")
	r.Run(":8080")
}
