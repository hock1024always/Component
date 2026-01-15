package models

import (
	"testing"
	"time"
)

func TestNewPlayerScore(t *testing.T) {
	player := NewPlayerScore("user1", "Alice", 100)
	if player.UserID != "user1" {
		t.Errorf("Expected UserID user1, got %s", player.UserID)
	}
	if player.Username != "Alice" {
		t.Errorf("Expected Username Alice, got %s", player.Username)
	}
	if player.Score != 100 {
		t.Errorf("Expected Score 100, got %d", player.Score)
	}
	if player.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should be set")
	}
}

func TestUpdateScore(t *testing.T) {
	player := NewPlayerScore("user1", "Alice", 100)
	oldTime := player.UpdatedAt

	time.Sleep(1 * time.Millisecond)

	player.UpdateScore(200)

	if player.Score != 200 {
		t.Errorf("Expected Score 200, got %d", player.Score)
	}
	if !player.UpdatedAt.After(oldTime) {
		t.Error("UpdatedAt should be updated")
	}
}
