package repository

import (
	"context"

	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model"
)

// GetFilesByUserID retrieves all files for a user with filtering and pagination
func (r *CloudRepositoryRepository) GetFilesByUserID(ctx context.Context, userID uint, filter model.ListFilesRequestDTO) ([]model.CloudFile, int64, error) {
	var files []model.CloudFile
	var total int64

	query := r.db.WithContext(ctx).Model(&model.CloudFile{}).
		Preload("Tags"). // Eager load tags
		Where("user_id = ? AND deleted_at IS NULL", userID)
	
	// Apply filters
	if filter.FileType != "" {
		query = query.Where("file_type = ?", filter.FileType)
	}
	
	// Keyword search (filename OR tag name)
	if filter.Keyword != "" {
		query = query.Where("file_name LIKE ? OR id IN (SELECT cloud_file_id FROM file_tags JOIN tags ON tags.id = file_tags.tag_id WHERE tags.name LIKE ?)", "%"+filter.Keyword+"%", "%"+filter.Keyword+"%")
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
	
	offset := (page - 1) * pageSize
	err := query.Limit(pageSize).Offset(offset).Find(&files).Error
	if err != nil {
		return nil, 0, err
	}

	return files, total, nil
}
