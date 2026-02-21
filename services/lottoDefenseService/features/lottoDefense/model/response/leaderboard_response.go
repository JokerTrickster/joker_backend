package response

// LeaderboardEntry is one row in the leaderboard
type LeaderboardEntry struct {
	Rank   int    `json:"rank"`
	UserID uint   `json:"user_id"`
	Name   string `json:"name,omitempty"`
	Score  uint   `json:"score"`
}

// LeaderboardResponse is the API response for leaderboard
type LeaderboardResponse struct {
	Entries []LeaderboardEntry `json:"entries"`
}
