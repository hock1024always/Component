package services

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

type ConnectionManager struct {
	sync.RWMutex
	connections map[*websocket.Conn]bool
	broadcast   chan []byte
}

func NewConnectionManager() *ConnectionManager {
	return &ConnectionManager{
		connections: make(map[*websocket.Conn]bool),
		broadcast:   make(chan []byte, 100),
	}
}

func (cm *ConnectionManager) Register(conn *websocket.Conn) {
	cm.Lock()
	defer cm.Unlock()
	cm.connections[conn] = true
}

func (cm *ConnectionManager) Unregister(conn *websocket.Conn) {
	cm.Lock()
	defer cm.Unlock()
	if _, exists := cm.connections[conn]; exists {
		delete(cm.connections, conn)
		conn.Close()
	}
}

func (cm *ConnectionManager) BroadcastMessage(message interface{}) {
	jsonMsg, err := json.Marshal(message)
	if err != nil {
		log.Printf("Failed to marshal message: %v", err)
		return
	}
	cm.broadcast <- jsonMsg
}

func (cm *ConnectionManager) Run() {
	for {
		message := <-cm.broadcast

		cm.RLock()
		for conn := range cm.connections {
			err := conn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				log.Printf("Error writing message: %v", err)
				cm.RUnlock()
				cm.Unregister(conn)
				cm.RLock()
			}
		}
		cm.RUnlock()
	}
}
