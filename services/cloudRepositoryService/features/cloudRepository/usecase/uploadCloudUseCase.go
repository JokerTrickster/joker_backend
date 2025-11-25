package usecase

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/entity"
	_interface "github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/interface"
	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/request"
	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/response"
	"github.com/google/uuid"
)

const (
	DefaultUploadExpiration = 15 * time.Minute
	MaxFileSize             = 100 * 1024 * 1024 // 100MB
)

var (
	AllowedImageTypes = map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}
	AllowedVideoTypes = map[string]bool{
		"video/mp4":  true,
		"video/webm": true,
		"video/avi":  true,
		"video/mov":  true,
	}
)

type UploadCloudRepositoryUseCase struct {
	Repo           _interface.IUploadCloudRepositoryRepository
	ContextTimeout time.Duration
}

func NewUploadCloudRepositoryUseCase(repo _interface.IUploadCloudRepositoryRepository, timeout time.Duration) _interface.IUploadCloudRepositoryUseCase {
	return &UploadCloudRepositoryUseCase{
		Repo:           repo,
		ContextTimeout: timeout,
	}
}

// RequestUploadURL generates a presigned upload URL and creates a file record
func (u *UploadCloudRepositoryUseCase) RequestUploadURL(c context.Context, userID uint, req *request.UploadRequestDTO) (*response.UploadResponseDTO, error) {
	ctx, cancel := context.WithTimeout(c, u.ContextTimeout)
	defer cancel()
	// Validate file size
	if req.FileSize > MaxFileSize {
		return nil, fmt.Errorf("file size exceeds maximum allowed size of %d bytes", MaxFileSize)
	}

	// Validate content type
	fileType := entity.FileType(req.FileType)
	if fileType == entity.FileTypeImage && !AllowedImageTypes[req.ContentType] {
		return nil, fmt.Errorf("invalid image content type: %s", req.ContentType)
	}
	if fileType == entity.FileTypeVideo && !AllowedVideoTypes[req.ContentType] {
		return nil, fmt.Errorf("invalid video content type: %s", req.ContentType)
	}

	// Generate S3 key
	s3Key := u.generateS3Key(userID, fileType, req.FileName)

	// Create file record in database
	file := &entity.CloudFile{
		UserID:      userID,
		FileName:    req.FileName,
		S3Key:       s3Key,
		FileType:    fileType,
		ContentType: req.ContentType,
		FileSize:    req.FileSize,
	}

	if err := u.Repo.CreateFile(ctx, file); err != nil {
		return nil, fmt.Errorf("failed to create file record: %w", err)
	}

	// Generate presigned upload URL
	uploadURL, err := u.Repo.GeneratePresignedUploadURL(ctx, s3Key, req.ContentType, DefaultUploadExpiration)
	if err != nil {
		return nil, fmt.Errorf("failed to generate upload URL: %w", err)
	}

	return &response.UploadResponseDTO{
		FileID:    file.ID,
		UploadURL: uploadURL,
		S3Key:     s3Key,
		ExpiresIn: int(DefaultUploadExpiration.Seconds()),
	}, nil
}

// generateS3Key generates a unique S3 key for a file
func (u *UploadCloudRepositoryUseCase) generateS3Key(userID uint, fileType entity.FileType, fileName string) string {
	// Generate UUID for uniqueness
	fileID := uuid.New().String()

	// Get file extension
	ext := filepath.Ext(fileName)
	if ext == "" {
		ext = ".bin"
	}

	// Clean extension
	ext = strings.ToLower(ext)

	// Get filename without extension
	baseName := strings.TrimSuffix(fileName, ext)
	if baseName == "" {
		baseName = "file"
	}

	// Format: users/{userID}/files/{uuid}-{baseName}{ext}
	return fmt.Sprintf("users/%d/files/%s-%s%s", userID, fileID, baseName, ext)
}
