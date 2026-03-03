package request

type ReqSubmitRanking struct {
	ClearTimeMs uint `json:"clearTimeMs" validate:"required"`
}

type ReqListRanking struct {
	Limit int `query:"limit"`
}
