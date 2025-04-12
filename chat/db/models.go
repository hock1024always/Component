package db

import "time"

type User struct {
	ID        uint   `gorm:"primaryKey;autoIncrement"`
	Username  string `gorm:"type:varchar(255);uniqueIndex;notNull"`
	Password  string `gorm:"type:varchar(255);notNull"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time `gorm:"index"`
}

type Message struct {
	ID         uint      `gorm:"primaryKey;autoIncrement"`
	SenderID   uint      `gorm:"notNull"`
	ReceiverID uint      `gorm:"notNull"`
	Content    string    `gorm:"type:varchar(1024);notNull"`
	SendTime   time.Time `gorm:"autoCreateTime"`
}
