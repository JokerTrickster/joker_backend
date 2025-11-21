package usecase

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model"
	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/repository"
	sharedAws "github.com/JokerTrickster/joker_backend/shared/aws"
	"github.com/google/uuid"
)

const (
	DefaultUploadExpiration   = 15 * time.Minute
	DefaultDownloadExpiration = 1 * time.Hour
	MaxFileSize               = 100 * 1024 * 1024 // 100MB
	MaxBatchSize              = 30                // Maximum files per batch
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

type CloudRepositoryUsecase struct {
	repo *repository.CloudRepositoryRepository
}

func NewCloudRepositoryUsecase(repo *repository.CloudRepositoryRepository) *CloudRepositoryUsecase {
	return &CloudRepositoryUsecase{
		repo: repo,
	}
}

// generateS3Key generates a unique S3 key for a file
func (u *CloudRepositoryUsecase) generateS3Key(userID uint, fileType model.FileType, fileName string) string {
	// Generate UUID for uniqueness
	fileID := uuid.New().String()

	// Get file extension
	ext := filepath.Ext(fileName)
	if ext == "" {
		ext = ".bin"
	}

	// Clean extension
	ext = strings.ToLower(ext)

	// Generate random filename component
	randomName := sharedAws.FileNameGenerateRandom()

	// Format: {fileType}/{userID}/{uuid}_{random}{ext}
	return fmt.Sprintf("%s/%d/%s_%s%s", fileType, userID, fileID, randomName, ext)
}
