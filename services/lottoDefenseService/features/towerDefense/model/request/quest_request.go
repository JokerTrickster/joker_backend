package request

type UpdateQuestProgressRequest struct {
	Increment uint `json:"increment" validate:"required,min=1"`
}
