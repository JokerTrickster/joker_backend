package entity

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// LottoNumbers is a slice of 6 integers (1-45) for a lotto draw
type LottoNumbers [6]int

// Value implements driver.Valuer for GORM JSON storage
func (n LottoNumbers) Value() (driver.Value, error) {
	return json.Marshal(n[:])
}

// Scan implements sql.Scanner for GORM JSON storage
func (n *LottoNumbers) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("lotto numbers: expected []byte")
	}
	var arr []int
	if err := json.Unmarshal(bytes, &arr); err != nil {
		return err
	}
	if len(arr) != 6 {
		return errors.New("lotto numbers: expected 6 numbers")
	}
	copy(n[:], arr)
	return nil
}

// LottoDraw represents the drawn numbers for a round
type LottoDraw struct {
	ID        uint         `gorm:"primaryKey" json:"id"`
	RoundID   uint         `gorm:"not null;uniqueIndex" json:"round_id"`
	Numbers   LottoNumbers `gorm:"type:json;not null" json:"numbers"`
	CreatedAt time.Time    `gorm:"autoCreateTime" json:"created_at"`
}

// TableName specifies the table name for LottoDraw
func (LottoDraw) TableName() string {
	return "lotto_draws"
}
