package models

import (
	"gorm.io/gorm"
)

type RoomUser struct {
	gorm.Model
	Rid uint `gorm:"column:rid;type:int(11);not null" json:"rid"` //房间ID
	Uid uint `gorm:"column:uid;type:int(11);not null" json:"uid"` //用户ID
}

func (table *RoomUser) TableName() string {
	return "room_user"
}

//// Meeting represents a meeting entity.
//type Meeting struct {
//	gorm.Model
//	Identity string    `gorm:"column:identity;type:varchar(36);uniqueIndex;not null" json:"identity"`
//	Name     string    `gorm:"column:name;type:varchar(100);not null" json:"name"`
//	BeginAt  time.Time `gorm:"column:begin_at;type:datetime;" json:"begin_at"`
//	EndAt    time.Time `gorm:"column:end_at;type:datetime;" json:"end_at"`
//	CreateId uint      `gorm:"column:create_id;type:int(11);" json:"create_id"` // 创建者ID
//}
