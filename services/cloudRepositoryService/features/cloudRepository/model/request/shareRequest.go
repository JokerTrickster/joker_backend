package request

// ShareFolderRequestDTO for sharing a folder with users
type ShareFolderRequestDTO struct {
	UserEmails []string `json:"user_emails" validate:"required,min=1,dive,email"`
	Permission string   `json:"permission" validate:"omitempty,oneof=read write"` // Optional: "read" or "write", defaults to "read"
}

// ShareFileRequestDTO for sharing a file with users
type ShareFileRequestDTO struct {
	UserEmails []string `json:"user_emails" validate:"required,min=1,dive,email"`
	Permission string   `json:"permission" validate:"omitempty,oneof=read write"` // Optional: "read" or "write", defaults to "read"
}
