package request

// EndRoundRequest is the body for ending a round (submitting score)
type EndRoundRequest struct {
	Score uint `json:"score" validate:"required"`
}
