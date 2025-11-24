package response

// DownloadResponseDTO returns presigned download URL
type DownloadResponseDTO struct {
	DownloadURL string `json:"download_url"`
	FileName    string `json:"file_name"`
	ExpiresIn   int    `json:"expires_in"` // seconds
}
