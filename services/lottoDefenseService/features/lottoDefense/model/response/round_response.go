package response

import "time"

// RoundResponse is the API response for a single round
type RoundResponse struct {
	ID        uint       `json:"id"`
	UserID    uint       `json:"user_id"`
	Status    string     `json:"status"`
	Score     *uint      `json:"score,omitempty"`
	StartedAt *time.Time `json:"started_at,omitempty"`
	EndedAt   *time.Time `json:"ended_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// RoundWithDrawResponse includes lotto draw when round is completed
type RoundWithDrawResponse struct {
	RoundResponse
	Numbers []int `json:"numbers,omitempty"`
}
