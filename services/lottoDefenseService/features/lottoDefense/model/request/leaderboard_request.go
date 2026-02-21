package request

// LeaderboardQuery holds query params for leaderboard
type LeaderboardQuery struct {
	Limit int `query:"limit"` // default 10, max 100
}
