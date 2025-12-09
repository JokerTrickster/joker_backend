package repository

import (
	"context"
	"fmt"

	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/entity"
	_interface "github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/interface"
	"gorm.io/gorm"
)

type FileShareRepository struct {
	db *gorm.DB
}

func NewFileShareRepository(db *gorm.DB) _interface.IFileShareRepository {
	return &FileShareRepository{
		db: db,
	}
}

// CreateFileShare creates a new file share
func (r *FileShareRepository) CreateFileShare(ctx context.Context, share *entity.FileShare) error {
	if err := r.db.WithContext(ctx).Create(share).Error; err != nil {
		return fmt.Errorf("failed to create file share: %w", err)
	}
	return nil
}

// GetFileSharesByFileID retrieves all shares for a file
func (r *FileShareRepository) GetFileSharesByFileID(ctx context.Context, fileID uint) ([]entity.FileShare, error) {
	var shares []entity.FileShare
	if err := r.db.WithContext(ctx).
		Preload("SharedWith").
		Where("file_id = ? AND deleted_at IS NULL", fileID).
		Find(&shares).Error; err != nil {
		return nil, fmt.Errorf("failed to get file shares: %w", err)
	}
	return shares, nil
}

// GetSharedFilesByUserID retrieves all files shared with a user
func (r *FileShareRepository) GetSharedFilesByUserID(ctx context.Context, userID int32) ([]entity.FileShare, error) {
	var shares []entity.FileShare
	if err := r.db.WithContext(ctx).
		Preload("File").
		Preload("Owner").
		Where("shared_with_id = ? AND deleted_at IS NULL", userID).
		Find(&shares).Error; err != nil {
		return nil, fmt.Errorf("failed to get shared files: %w", err)
	}
	return shares, nil
}

// DeleteFileShare removes a file share
func (r *FileShareRepository) DeleteFileShare(ctx context.Context, fileID uint, sharedWithID int32, ownerID int32) error {
	result := r.db.WithContext(ctx).
		Where("file_id = ? AND shared_with_id = ? AND owner_id = ? AND deleted_at IS NULL", fileID, sharedWithID, ownerID).
		Delete(&entity.FileShare{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete file share: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("file share not found")
	}

	return nil
}

// HasFileAccess checks if a user has access to a file (owner, shared, or folder shared)
func (r *FileShareRepository) HasFileAccess(ctx context.Context, userID int32, fileID uint) (bool, error) {
	// Get the file
	var file entity.CloudFile
	err := r.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", fileID).
		First(&file).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, fmt.Errorf("failed to get file: %w", err)
	}

	// Check if user is the owner
	if file.UserID == uint(userID) {
		return true, nil
	}

	// Check if file is directly shared with user
	var fileShare entity.FileShare
	err = r.db.WithContext(ctx).
		Where("file_id = ? AND shared_with_id = ? AND deleted_at IS NULL", fileID, userID).
		First(&fileShare).Error

	if err == nil {
		return true, nil // User has direct file share access
	}

	if err != gorm.ErrRecordNotFound {
		return false, fmt.Errorf("failed to check file share: %w", err)
	}

	// Check if file's parent folder is shared with user
	if file.FolderID != nil {
		var folderShare entity.FolderShare
		err = r.db.WithContext(ctx).
			Where("folder_id = ? AND shared_with_id = ? AND deleted_at IS NULL", *file.FolderID, userID).
			First(&folderShare).Error

		if err == nil {
			return true, nil // User has access via folder share
		}

		if err != gorm.ErrRecordNotFound {
			return false, fmt.Errorf("failed to check folder share: %w", err)
		}
	}

	return false, nil // No access
}

// GetFilesSharedByUserID retrieves files that a user has shared with others
func (r *FileShareRepository) GetFilesSharedByUserID(ctx context.Context, ownerID int32) ([]entity.FileShare, error) {
	var shares []entity.FileShare
	if err := r.db.WithContext(ctx).
		Preload("File").
		Preload("SharedWith").
		Where("owner_id = ? AND deleted_at IS NULL", ownerID).
		Order("created_at DESC").
		Find(&shares).Error; err != nil {
		return nil, fmt.Errorf("failed to get files shared by user: %w", err)
	}
	return shares, nil
}
