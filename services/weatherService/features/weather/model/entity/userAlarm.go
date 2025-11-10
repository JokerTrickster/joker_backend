package entity

import (
	"time"
)

type UserAlarm struct {
	ID        int        `gorm:"primaryKey;autoIncrement"`
	CreatedAt time.Time  `gorm:"autoCreateTime"`
	UpdatedAt time.Time  `gorm:"autoUpdateTime"`
	DeletedAt *time.Time `gorm:"index"`
	UserID    int        `gorm:"not null"`
	AlarmTime string     `gorm:"type:time;not null"` // HH:MM:SS 형식
	Region    string     `gorm:"type:varchar(255);not null"`
	IsEnabled bool       `gorm:"default:true"`
	LastSent  *time.Time
}

func (UserAlarm) TableName() string {
	return "user_alarms"
}
