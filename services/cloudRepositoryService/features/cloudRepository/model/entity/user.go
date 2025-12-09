package entity

import "time"

// User represents a user in the system (partial fields for sharing purposes)
type User struct {
	ID        int32      `gorm:"primaryKey" json:"id"`
	Name      string     `gorm:"size:255;not null" json:"name"`
	Email     string     `gorm:"size:255;not null;uniqueIndex" json:"email"`
	Provider  string     `gorm:"size:50;not null;default:local" json:"provider"`
	CreatedAt time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt *time.Time `gorm:"index" json:"deleted_at,omitempty"`
}

// TableName specifies the table name for User
func (User) TableName() string {
	return "users"
}
