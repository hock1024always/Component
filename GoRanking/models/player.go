package models

import "time"

type PlayerScore struct {
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	Score     int       `json:"score"`
	Rank      int       `json:"rank"`
	UpdatedAt time.Time `json:"updated_at"`
}

func NewPlayerScore(userID, username string, score int) *PlayerScore {
	return &PlayerScore{
		UserID:    userID,
		Username:  username,
		Score:     score,
		UpdatedAt: time.Now(),
	}
}

func (p *PlayerScore) UpdateScore(score int) {
	p.Score = score
	p.UpdatedAt = time.Now()
}
