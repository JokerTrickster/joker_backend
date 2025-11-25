package request

// ActivityHistoryRequestDTO for activity history request
type ActivityHistoryRequestDTO struct {
	Month string `query:"month"` // Format: YYYY-MM
}