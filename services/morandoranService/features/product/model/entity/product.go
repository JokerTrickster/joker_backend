package entity

import (
	"time"

	"gorm.io/gorm"
)

type Product struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	Name          string         `gorm:"size:100;not null" json:"name"`
	Price         uint           `gorm:"not null" json:"price"`
	OriginalPrice *uint          `json:"originalPrice"`
	Description   string         `gorm:"type:text;not null" json:"description"`
	Image         string         `gorm:"size:512;not null" json:"image"`
	Category      string         `gorm:"size:50;not null" json:"category"`
	Badge         *string        `gorm:"size:20" json:"badge"`
	InStock       bool           `gorm:"not null;default:true" json:"inStock"`
	CreatedAt     time.Time      `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime" json:"updatedAt"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Product) TableName() string {
	return "products"
}
