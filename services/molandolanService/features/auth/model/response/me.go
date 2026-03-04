package response

import "time"

type ResMe struct {
	ID           string    `json:"id"`
	Nickname     string    `json:"nickname"`
	Email        string    `json:"email"`
	ProfileImage *string   `json:"profileImage"`
	Role         string    `json:"role"`
	Provider     string    `json:"provider"`
	CreatedAt    time.Time `json:"createdAt"`
}
