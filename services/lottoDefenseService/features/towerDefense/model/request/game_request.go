package request

type SaveGameResultRequest struct {
	GameMode            string `json:"game_mode" validate:"required,oneof=single coop"`
	RoundsReached       uint   `json:"rounds_reached" validate:"required,min=1"`
	MonstersKilled      uint   `json:"monsters_killed" validate:"required"`
	GoldEarned          uint   `json:"gold_earned" validate:"required"`
	SurvivalTimeSeconds *uint  `json:"survival_time_seconds,omitempty"`
	FinalArmyValue      *uint  `json:"final_army_value,omitempty"`
	Result              string `json:"result" validate:"required,oneof=victory defeat disconnect"`
}

type GameHistoryRequest struct {
	GameMode string `json:"game_mode" form:"mode"`
	Limit    int    `json:"limit" form:"limit"`
	Offset   int    `json:"offset" form:"offset"`
}
