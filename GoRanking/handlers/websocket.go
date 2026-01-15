package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"ranking/models"
	"ranking/services"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WebSocketHandler struct {
	leaderboard *models.Leaderboard
	manager     *services.ConnectionManager
}

func NewWebSocketHandler(leaderboard *models.Leaderboard, manager *services.ConnectionManager) *WebSocketHandler {
	return &WebSocketHandler{
		leaderboard: leaderboard,
		manager:     manager,
	}
}

func (h *WebSocketHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer h.manager.Unregister(conn)

	h.manager.Register(conn)

	initialData := map[string]interface{}{
		"type":    "initial",
		"top10":   h.leaderboard.GetTopN(10),
		"updated": time.Now().Unix(),
	}
	jsonData, _ := json.Marshal(initialData)
	conn.WriteMessage(websocket.TextMessage, jsonData)

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			log.Printf("WebSocket connection closed: %v", err)
			break
		}
	}
}
