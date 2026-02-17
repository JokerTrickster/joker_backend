package response

import "time"

type AuthResponse struct {
	User  *UserData `json:"user"`
	Token string    `json:"token"`
}

type UserData struct {
	ID        uint       `json:"id"`
	Username  string     `json:"username"`
	Email     string     `json:"email"`
	CreatedAt time.Time  `json:"created_at"`
	LastLogin *time.Time `json:"last_login,omitempty"`
}

type UserInfoResponse struct {
	User  *UserData       `json:"user"`
	Stats *UserStatsData  `json:"stats"`
}

type UserStatsData struct {
	SingleHighestRound uint `json:"single_highest_round"`
	SingleTotalGames   uint `json:"single_total_games"`
	SingleTotalKills   uint `json:"single_total_kills"`
	CoopHighestRound   uint `json:"coop_highest_round"`
	CoopTotalGames     uint `json:"coop_total_games"`
	CoopTotalKills     uint `json:"coop_total_kills"`
	CoopWins           uint `json:"coop_wins"`
	TotalGoldEarned    uint `json:"total_gold_earned"`
	CurrentGold        uint `json:"current_gold"`
	QuestsCompleted    uint `json:"quests_completed"`
}

type UserStatsResponse struct {
	Single SingleStats `json:"single"`
	Coop   CoopStats   `json:"coop"`
	Gold   GoldStats   `json:"gold"`
}

type SingleStats struct {
	HighestRound uint    `json:"highest_round"`
	TotalGames   uint    `json:"total_games"`
	TotalKills   uint    `json:"total_kills"`
	AverageRound float64 `json:"average_round"`
}

type CoopStats struct {
	HighestRound uint    `json:"highest_round"`
	TotalGames   uint    `json:"total_games"`
	TotalKills   uint    `json:"total_kills"`
	Wins         uint    `json:"wins"`
	WinRate      float64 `json:"win_rate"`
}

type GoldStats struct {
	TotalEarned uint `json:"total_earned"`
	Current     uint `json:"current"`
}
