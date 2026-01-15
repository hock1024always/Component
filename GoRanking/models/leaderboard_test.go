package models

import "testing"

func TestNewLeaderboard(t *testing.T) {
	lb := NewLeaderboard()
	if lb == nil {
		t.Fatal("NewLeaderboard returned nil")
	}
}

func TestLeaderboardUpdateScore(t *testing.T) {
	lb := NewLeaderboard()

	lb.UpdateScore("user1", "Alice", 100)
	rank, exists := lb.GetUserRank("user1")
	if !exists {
		t.Fatal("User should exist")
	}
	if rank != 1 {
		t.Errorf("Expected rank 1, got %d", rank)
	}
}

func TestGetTopN(t *testing.T) {
	lb := NewLeaderboard()

	lb.UpdateScore("user1", "Alice", 300)
	lb.UpdateScore("user2", "Bob", 200)
	lb.UpdateScore("user3", "Charlie", 100)

	top3 := lb.GetTopN(3)
	if len(top3) != 3 {
		t.Errorf("Expected 3 players, got %d", len(top3))
	}

	if top3[0].Score != 300 {
		t.Errorf("Expected top score 300, got %d", top3[0].Score)
	}
}

func TestGetUserRank(t *testing.T) {
	lb := NewLeaderboard()

	lb.UpdateScore("user1", "Alice", 100)
	lb.UpdateScore("user2", "Bob", 200)

	rank, exists := lb.GetUserRank("user2")
	if !exists {
		t.Fatal("User should exist")
	}
	if rank != 1 {
		t.Errorf("Expected rank 1, got %d", rank)
	}
}
