package response

// LoginResponse represents login response
type LoginResponse struct {
	Token   string          `json:"token"`
	UserID  string          `json:"userId"`
	Profile ProfileResponse `json:"profile"`
}

// ProfileResponse represents player profile in responses
type ProfileResponse struct {
	Nickname   string      `json:"nickname"`
	AvatarID   string      `json:"avatarId"`
	Level      int         `json:"level"`
	Experience int         `json:"experience"`
	Stats      *StatsResponse `json:"stats,omitempty"`
}

// StatsResponse represents player statistics
type StatsResponse struct {
	GamesPlayed     int   `json:"gamesPlayed"`
	Victories       int   `json:"victories"`
	TotalScore      int64 `json:"totalScore"`
	HighestScore    int64 `json:"highestScore"`
	HighestWave     int   `json:"highestWave"`
	UnitsPlaced     int   `json:"unitsPlaced"`
	EnemiesDefeated int   `json:"enemiesDefeated"`
}

// RefreshTokenResponse represents refresh token response
type RefreshTokenResponse struct {
	Token string `json:"token"`
}