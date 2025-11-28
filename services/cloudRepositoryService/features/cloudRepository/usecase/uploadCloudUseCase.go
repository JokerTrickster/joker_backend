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
	"gorm.io/gorm"
)

const (
	DefaultUploadExpiration = 12 * time.Hour
)

var (
	AllowedImageTypes = map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}
	AllowedVideoTypes = map[string]bool{
		"video/mp4":        true,
		"video/webm":       true,
		"video/avi":        true,
		"video/x-msvideo":  true, // Alternative AVI MIME
		"video/mov":        true,
		"video/quicktime":  true, // Alternative MOV MIME
		"video/mpeg":       true, // MPEG files
		"video/x-matroska": true, // MKV files
		"video/3gpp":       true, // 3GP mobile video
	}
)

type UploadCloudRepositoryUseCase struct {
	Repo           _interface.IUploadCloudRepositoryRepository
	StatsRepo      _interface.IUserStatsCloudRepositoryRepository
	DB             *gorm.DB
	ContextTimeout time.Duration
}

func NewUploadCloudRepositoryUseCase(repo _interface.IUploadCloudRepositoryRepository, statsRepo _interface.IUserStatsCloudRepositoryRepository, db *gorm.DB, timeout time.Duration) _interface.IUploadCloudRepositoryUseCase {
	return &UploadCloudRepositoryUseCase{
		Repo:           repo,
		StatsRepo:      statsRepo,
		DB:             db,
		ContextTimeout: timeout,
	}
}

// RequestUploadURL generates a presigned upload URL and creates a file record
func (u *UploadCloudRepositoryUseCase) RequestUploadURL(c context.Context, userID uint, req *request.UploadRequestDTO) (*response.UploadResponseDTO, error) {
	ctx, cancel := context.WithTimeout(c, u.ContextTimeout)
	defer cancel()

	// Validate content type
	fileType := entity.FileType(req.FileType)
	if fileType == entity.FileTypeImage && !AllowedImageTypes[req.ContentType] {
		return nil, fmt.Errorf("invalid image content type: %s", req.ContentType)
	}
	if fileType == entity.FileTypeVideo && !AllowedVideoTypes[req.ContentType] {
		return nil, fmt.Errorf("invalid video content type: %s", req.ContentType)
	}

	// Generate S3 keys for original and thumbnail
	s3Key := u.generateS3Key(userID, fileType, req.FileName)
	thumbnailKey := ""

	// Generate thumbnail key for both images and videos
	if fileType == entity.FileTypeImage || fileType == entity.FileTypeVideo {
		thumbnailKey = u.generateThumbnailKey(userID, req.FileName)
	}

	// Process tags if provided
	tags := make([]entity.Tag, 0, len(req.Tags))
	if len(req.Tags) > 0 {
		for _, tagName := range req.Tags {
			if tagName == "" {
				continue
			}
			tag := entity.Tag{
				UserID: userID,
				Name:   tagName,
			}
			// Find or create tag
			if err := u.DB.WithContext(ctx).Where("user_id = ? AND name = ?", userID, tagName).FirstOrCreate(&tag).Error; err != nil {
				return nil, fmt.Errorf("failed to process tag %s: %w", tagName, err)
			}
			tags = append(tags, tag)
		}

		// Log tag activity
		if u.StatsRepo != nil {
			for _, tag := range tags {
				activity := &entity.ActivityLog{
					UserID:       userID,
					ActivityType: entity.ActivityTypeTagAdd,
					TagName:      tag.Name,
				}
				_ = u.StatsRepo.LogActivity(ctx, activity) // Don't fail on logging error
			}
		}
	}

	// Create file record in database
	file := &entity.CloudFile{
		UserID:       userID,
		FileName:     req.FileName,
		S3Key:        s3Key,
		ThumbnailKey: thumbnailKey,
		FileType:     fileType,
		ContentType:  req.ContentType,
		FileSize:     req.FileSize,
		Duration:     req.Duration,
		Tags:         tags,
	}

	if err := u.Repo.CreateFile(ctx, file); err != nil {
		return nil, fmt.Errorf("failed to create file record: %w", err)
	}

	// Log upload activity
	if u.StatsRepo != nil {
		activity := &entity.ActivityLog{
			UserID:       userID,
			FileID:       &file.ID,
			ActivityType: entity.ActivityTypeUpload,
		}
		_ = u.StatsRepo.LogActivity(ctx, activity) // Don't fail on logging error
	}

	// Generate presigned upload URL for original
	uploadURL, err := u.Repo.GeneratePresignedUploadURL(ctx, s3Key, req.ContentType, DefaultUploadExpiration)
	if err != nil {
		return nil, fmt.Errorf("failed to generate upload URL: %w", err)
	}

	// Generate presigned upload URL for thumbnail if it's an image
	thumbnailURL := ""
	if thumbnailKey != "" {
		thumbnailURL, err = u.Repo.GeneratePresignedUploadURL(ctx, thumbnailKey, req.ContentType, DefaultUploadExpiration)
		if err != nil {
			// Log error but don't fail the entire request
			thumbnailURL = ""
		}
	}

	return &response.UploadResponseDTO{
		FileID:       file.ID,
		UploadURL:    uploadURL,
		S3Key:        s3Key,
		ThumbnailURL: thumbnailURL,
		ThumbnailKey: thumbnailKey,
		ExpiresIn:    int(DefaultUploadExpiration.Seconds()),
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

// generateThumbnailKey generates a unique S3 key for a thumbnail
func (u *UploadCloudRepositoryUseCase) generateThumbnailKey(userID uint, fileName string) string {
	// Generate UUID for uniqueness
	fileID := uuid.New().String()

	// Get file extension
	ext := filepath.Ext(fileName)
	if ext == "" {
		ext = ".jpg" // Default to jpg for thumbnails
	}

	// Clean extension
	ext = strings.ToLower(ext)

	// Get filename without extension
	baseName := strings.TrimSuffix(fileName, ext)
	if baseName == "" {
		baseName = "file"
	}

	// Format: users/{userID}/thumbnails/{uuid}-{baseName}_thumb{ext}
	return fmt.Sprintf("users/%d/thumbnails/%s-%s_thumb%s", userID, fileID, baseName, ext)
}
