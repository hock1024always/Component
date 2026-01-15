package models

import (
	"sort"
	"sync"
)

type Leaderboard struct {
	sync.RWMutex
	scores     map[string]*PlayerScore
	sortedKeys []string
}

func NewLeaderboard() *Leaderboard {
	return &Leaderboard{
		scores:     make(map[string]*PlayerScore),
		sortedKeys: make([]string, 0),
	}
}

func (lb *Leaderboard) UpdateScore(userID, username string, score int) {
	lb.Lock()
	defer lb.Unlock()

	if player, exists := lb.scores[userID]; exists {
		player.UpdateScore(score)
	} else {
		lb.scores[userID] = NewPlayerScore(userID, username, score)
	}

	lb.resort()
}

func (lb *Leaderboard) resort() {
	players := make([]*PlayerScore, 0, len(lb.scores))
	for _, player := range lb.scores {
		players = append(players, player)
	}

	sort.Slice(players, func(i, j int) bool {
		if players[i].Score == players[j].Score {
			return players[i].UpdatedAt.Before(players[j].UpdatedAt)
		}
		return players[i].Score > players[j].Score
	})

	lb.sortedKeys = make([]string, len(players))
	for i, player := range players {
		player.Rank = i + 1
		lb.sortedKeys[i] = player.UserID
	}
}

func (lb *Leaderboard) GetTopN(n int) []*PlayerScore {
	lb.RLock()
	defer lb.RUnlock()

	if n > len(lb.sortedKeys) {
		n = len(lb.sortedKeys)
	}

	result := make([]*PlayerScore, n)
	for i := 0; i < n; i++ {
		userID := lb.sortedKeys[i]
		result[i] = lb.scores[userID]
	}
	return result
}

func (lb *Leaderboard) GetUserRank(userID string) (int, bool) {
	lb.RLock()
	defer lb.RUnlock()

	player, exists := lb.scores[userID]
	if !exists {
		return 0, false
	}
	return player.Rank, true
}
