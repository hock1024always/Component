package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"ranking/models"
	"ranking/services"
)

type APIHandler struct {
	leaderboard *models.Leaderboard
	manager     *services.ConnectionManager
}

func NewAPIHandler(leaderboard *models.Leaderboard, manager *services.ConnectionManager) *APIHandler {
	return &APIHandler{
		leaderboard: leaderboard,
		manager:     manager,
	}
}

func (h *APIHandler) HandleUpdateScore(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		UserID   string `json:"user_id"`
		Username string `json:"username"`
		Score    int    `json:"score"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.UserID == "" || req.Username == "" || req.Score < 0 {
		http.Error(w, "Invalid parameters", http.StatusBadRequest)
		return
	}

	h.leaderboard.UpdateScore(req.UserID, req.Username, req.Score)

	broadcastMsg := map[string]interface{}{
		"type":    "update",
		"top10":   h.leaderboard.GetTopN(10),
		"updated": time.Now().Unix(),
	}

	h.manager.BroadcastMessage(broadcastMsg)

	currentRank, _ := h.leaderboard.GetUserRank(req.UserID)
	response := map[string]interface{}{
		"success": true,
		"rank":    currentRank,
		"user_id": req.UserID,
		"message": "Score updated successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *APIHandler) HandleGetTop(w http.ResponseWriter, r *http.Request) {
	top := h.leaderboard.GetTopN(10)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(top)
}
