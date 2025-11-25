package response

// DailyActivityDTO represents activity for a single day
type DailyActivityDTO struct {
	Uploads   int      `json:"uploads"`
	Downloads int      `json:"downloads"`
	Tags      []string `json:"tags"`
}

// ActivityHistoryResponseDTO for activity history
// Map key is date in "YYYY-MM-DD" format
type ActivityHistoryResponseDTO map[string]DailyActivityDTO