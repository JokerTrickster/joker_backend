package mysql

import (
	"gorm.io/gorm"
)

type Tokens struct {
	gorm.Model
	UserID           uint   `json:"userID" gorm:"column:user_id"`
	AccessToken      string `json:"accessToken" gorm:"column:access_token"`
	RefreshToken     string `json:"refreshToken" gorm:"column:refresh_token"`
	RefreshExpiredAt int64  `json:"refreshExpiredAt" gorm:"column:refresh_expired_at"`
}

type Users struct {
	gorm.Model
	Name     string `json:"name" gorm:"column:name"`
	Email    string `json:"email" gorm:"uniqueIndex;column:email"`
	Password string `json:"password" gorm:"column:password"`
	Provider string `json:"provider" gorm:"column:provider"`
}
