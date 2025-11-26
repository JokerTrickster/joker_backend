package repository

import (
	"context"
	"strings"

	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/entity"
	_interface "github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/interface"
	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/request"
	"gorm.io/gorm"
)

type FavoriteRepository struct {
	db *gorm.DB
}

func NewFavoriteRepository(db *gorm.DB) _interface.IFavoriteRepository {
	return &FavoriteRepository{
		db: db,
	}
}

// AddFavorite creates a new favorite record (idempotent using FirstOrCreate)
func (r *FavoriteRepository) AddFavorite(ctx context.Context, userID, fileID uint) (*entity.Favorite, error) {
	favorite := &entity.Favorite{
		UserID: userID,
		FileID: fileID,
	}
	// FirstOrCreate is idempotent - returns existing or creates new
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND file_id = ?", userID, fileID).
		FirstOrCreate(favorite).Error
	return favorite, err
}

// RemoveFavorite deletes a favorite record (idempotent - no error if not exists)
func (r *FavoriteRepository) RemoveFavorite(ctx context.Context, userID, fileID uint) error {
	// Delete returns no error even if record doesn't exist (0 rows affected)
	return r.db.WithContext(ctx).
		Where("user_id = ? AND file_id = ?", userID, fileID).
		Delete(&entity.Favorite{}).Error
}

// GetFavoritesByUserID retrieves all favorited files for a user with filtering and pagination
func (r *FavoriteRepository) GetFavoritesByUserID(ctx context.Context, userID uint, filter request.ListFavoritesRequestDTO) ([]entity.CloudFile, int64, error) {
	var files []entity.CloudFile
	var total int64

	// Build query with JOIN to get file details
	query := r.db.WithContext(ctx).
		Model(&entity.CloudFile{}).
		Preload("Tags"). // Eager load tags
		Joins("INNER JOIN favorites ON cloud_files.id = favorites.file_id").
		Where("favorites.user_id = ? AND cloud_files.deleted_at IS NULL", userID)

	// Apply filename search filter
	if filter.Q != "" {
		query = query.Where("cloud_files.file_name LIKE ?", "%"+filter.Q+"%")
	}

	// Apply extension filter
	if filter.Ext != "" {
		// Extension should match the file extension (case-insensitive)
		ext := strings.TrimPrefix(filter.Ext, ".")
		query = query.Where("LOWER(cloud_files.file_name) LIKE ?", "%."+strings.ToLower(ext))
	}

	// Apply tag filter
	if filter.Tag != "" {
		query = query.Where("cloud_files.id IN (SELECT cloud_file_id FROM file_tags JOIN tags ON tags.id = file_tags.tag_id WHERE tags.name = ?)", filter.Tag)
	}

	// Get total count before pagination
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting
	sortField := "favorites.favorited_at" // Default sort by when favorited
	if filter.Sort == "fileName" {
		sortField = "cloud_files.file_name"
	} else if filter.Sort == "uploadDate" {
		sortField = "cloud_files.created_at"
	}

	sortOrder := "DESC"
	if filter.Order == "asc" {
		sortOrder = "ASC"
	}
	query = query.Order(sortField + " " + sortOrder)

	// Apply pagination
	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.Size
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

// CheckIsFavorited checks if a file is favorited by a user
func (r *FavoriteRepository) CheckIsFavorited(ctx context.Context, userID, fileID uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entity.Favorite{}).
		Where("user_id = ? AND file_id = ?", userID, fileID).
		Count(&count).Error
	return count > 0, err
}
