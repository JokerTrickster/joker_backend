package entity

import (
	"time"
)

type WeatherServiceToken struct {
	ID        int        `gorm:"primaryKey;autoIncrement"`
	CreatedAt time.Time  `gorm:"autoCreateTime"`
	UpdatedAt time.Time  `gorm:"autoUpdateTime"`
	DeletedAt *time.Time `gorm:"index"`
	UserID    int        `gorm:"not null"`
	FCMToken  string     `gorm:"type:varchar(500);not null"`
	DeviceID  string     `gorm:"type:varchar(255)"`
}

func (WeatherServiceToken) TableName() string {
	return "weather_service_tokens"
}
