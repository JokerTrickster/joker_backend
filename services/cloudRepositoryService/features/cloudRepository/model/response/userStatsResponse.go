package response

// StorageInfoDTO represents storage usage information
type StorageInfoDTO struct {
	Used       int64   `json:"used"`       // Bytes
	Total      int64   `json:"total"`      // Bytes
	Percentage float64 `json:"percentage"`
}

// MonthlyStatsDTO represents monthly activity statistics
type MonthlyStatsDTO struct {
	Uploads     int `json:"uploads"`
	Downloads   int `json:"downloads"`
	TagsCreated int `json:"tagsCreated"`
}

// UserStatsResponseDTO for user statistics
type UserStatsResponseDTO struct {
	Storage      StorageInfoDTO  `json:"storage"`
	MonthlyStats MonthlyStatsDTO `json:"monthlyStats"`
}