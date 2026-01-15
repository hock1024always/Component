package main

import (
	"log"
	"net/http"

	"ranking/handlers"
	"ranking/models"
	"ranking/services"
)

func main() {
	// 创建排行榜
	leaderboard := models.NewLeaderboard()

	// 创建连接管理器
	manager := services.NewConnectionManager()

	// 创建处理器
	wsHandler := handlers.NewWebSocketHandler(leaderboard, manager)
	apiHandler := handlers.NewAPIHandler(leaderboard, manager)

	// 启动WebSocket广播协程
	go manager.Run()

	// 设置路由
	http.HandleFunc("/ws", wsHandler.HandleWebSocket)
	http.HandleFunc("/api/update-score", apiHandler.HandleUpdateScore)
	http.HandleFunc("/api/top", apiHandler.HandleGetTop)
	http.Handle("/", http.FileServer(http.Dir("./static")))

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
