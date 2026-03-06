package repository

import (
	"context"
	"strings"
	"time"

	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/entity"
	_interface "github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/interface"
	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/request"
	sharedAws "github.com/JokerTrickster/joker_backend/shared/aws"
	"gorm.io/gorm"
)

type ListCloudRepositoryRepository struct {
	db     *gorm.DB
	bucket string
}

func escapeLike(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "%", "\\%")
	s = strings.ReplaceAll(s, "_", "\\_")
	return s
}

func NewListCloudRepositoryRepository(db *gorm.DB, bucket string) _interface.IListCloudRepositoryRepository {
	return &ListCloudRepositoryRepository{
		db:     db,
		bucket: bucket,
	}
}

// GetFilesByUserID retrieves all files for a user with filtering and pagination
func (r *ListCloudRepositoryRepository) GetFilesByUserID(ctx context.Context, userID uint, filter request.ListFilesRequestDTO) ([]entity.CloudFile, int64, error) {
	var files []entity.CloudFile
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.CloudFile{}).
		Preload("Tags"). // Eager load tags
		Where("user_id = ? AND deleted_at IS NULL", userID)

	// Filter by folder (important for showing files in correct location)
	// If folder_id is provided in query params, filter by it
	// - folder_id=0: root folder (folder_id IS NULL in DB)
	// - folder_id=N: specific folder N
	// - folder_id not provided: return all files (no folder filter)
	if filter.FolderID != nil {
		if *filter.FolderID == 0 {
			// Root folder: files with folder_id IS NULL
			query = query.Where("folder_id IS NULL")
		} else {
			// Specific folder
			query = query.Where("folder_id = ?", *filter.FolderID)
		}
	}

	// Apply filters
	if filter.FileType != "" {
		query = query.Where("file_type = ?", filter.FileType)
	}

	// Keyword search (filename OR tag name)
	if filter.Keyword != "" {
		kw := escapeLike(filter.Keyword)
		query = query.Where("file_name LIKE ? OR id IN (SELECT cloud_file_id FROM file_tags JOIN tags ON tags.id = file_tags.tag_id WHERE tags.name LIKE ?)", "%"+kw+"%", "%"+kw+"%")
	}

	// Tag filtering (files must have ALL specified tags)
	if len(filter.Tags) > 0 {
		for _, tagName := range filter.Tags {
			query = query.Where("id IN (SELECT cloud_file_id FROM file_tags JOIN tags ON tags.id = file_tags.tag_id WHERE tags.name = ?)", tagName)
		}
	}

	if filter.StartDate != "" {
		query = query.Where("created_at >= ?", filter.StartDate+" 00:00:00")
	}
	if filter.EndDate != "" {
		query = query.Where("created_at <= ?", filter.EndDate+" 23:59:59")
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting
	switch filter.Sort {
	case "oldest":
		query = query.Order("created_at ASC")
	case "name":
		query = query.Order("file_name ASC")
	case "size":
		query = query.Order("file_size DESC")
	default: // "latest" or empty
		query = query.Order("created_at DESC")
	}

	// Apply pagination
	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	offset := (page - 1) * pageSize
	err := query.Limit(pageSize).Offset(offset).Find(&files).Error
	if err != nil {
		return nil, 0, err
	}

	return files, total, nil
}

// GeneratePresignedDownloadURL generates a presigned URL for downloading
func (r *ListCloudRepositoryRepository) GeneratePresignedDownloadURL(ctx context.Context, s3Key string, expiration time.Duration) (string, error) {
	return sharedAws.GeneratePresignedDownloadURL(ctx, r.bucket, s3Key, expiration)
}
