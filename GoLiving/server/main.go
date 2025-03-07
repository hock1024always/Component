package main

import (
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"log"
	"online_meeting/models"
	"online_meeting/server/router"
)

// @title Online Meeting API
// @version 1.0
// @description 这是一个在线会议服务的API文档。
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /

func main() {
	models.NewDB()
	e := router.Router()
	err := e.Run(":8080") // 监听并在 0.0.0.0:8080 上启动服务

	// 添加Swagger UI路由
	e.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	log.Fatal(e.Run(":8080")) // 监听并服务

	if err != nil {
		//这个是在启动阶段的常见写法
		log.Fatalln("run err.", err) //记录错误日志并终止程序
		return
	}
}
