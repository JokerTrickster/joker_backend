package repository

import (
	"context"
	"fmt"

	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/entity"
	_interface "github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/interface"
	"gorm.io/gorm"
)

type FolderShareRepository struct {
	db *gorm.DB
}

func NewFolderShareRepository(db *gorm.DB) _interface.IFolderShareRepository {
	return &FolderShareRepository{
		db: db,
	}
}

// CreateFolderShare creates a new folder share
func (r *FolderShareRepository) CreateFolderShare(ctx context.Context, share *entity.FolderShare) error {
	if err := r.db.WithContext(ctx).Create(share).Error; err != nil {
		return fmt.Errorf("failed to create folder share: %w", err)
	}
	return nil
}

// GetFolderSharesByFolderID retrieves all shares for a folder
func (r *FolderShareRepository) GetFolderSharesByFolderID(ctx context.Context, folderID uint) ([]entity.FolderShare, error) {
	var shares []entity.FolderShare
	if err := r.db.WithContext(ctx).
		Preload("SharedWith").
		Where("folder_id = ? AND deleted_at IS NULL", folderID).
		Find(&shares).Error; err != nil {
		return nil, fmt.Errorf("failed to get folder shares: %w", err)
	}
	return shares, nil
}

// GetSharedFoldersByUserID retrieves all folders shared with a user
func (r *FolderShareRepository) GetSharedFoldersByUserID(ctx context.Context, userID int32) ([]entity.FolderShare, error) {
	var shares []entity.FolderShare
	if err := r.db.WithContext(ctx).
		Preload("Folder").
		Preload("Owner").
		Where("shared_with_id = ? AND deleted_at IS NULL", userID).
		Find(&shares).Error; err != nil {
		return nil, fmt.Errorf("failed to get shared folders: %w", err)
	}
	return shares, nil
}

// DeleteFolderShare removes a folder share
func (r *FolderShareRepository) DeleteFolderShare(ctx context.Context, folderID uint, sharedWithID int32, ownerID int32) error {
	result := r.db.WithContext(ctx).
		Where("folder_id = ? AND shared_with_id = ? AND owner_id = ? AND deleted_at IS NULL", folderID, sharedWithID, ownerID).
		Delete(&entity.FolderShare{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete folder share: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("folder share not found")
	}

	return nil
}

// HasFolderAccess checks if a user has access to a folder (either as owner or shared)
func (r *FolderShareRepository) HasFolderAccess(ctx context.Context, userID uint, folderID uint) (bool, error) {
	// Check if user is the owner
	var folder entity.Folder
	err := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ? AND deleted_at IS NULL", folderID, userID).
		First(&folder).Error

	if err == nil {
		return true, nil // User is the owner
	}

	if err != gorm.ErrRecordNotFound {
		return false, fmt.Errorf("failed to check folder ownership: %w", err)
	}

	// Check if folder is shared with user
	var share entity.FolderShare
	err = r.db.WithContext(ctx).
		Where("folder_id = ? AND shared_with_id = ? AND deleted_at IS NULL", folderID, userID).
		First(&share).Error

	if err == nil {
		return true, nil // User has shared access
	}

	if err == gorm.ErrRecordNotFound {
		return false, nil // No access
	}

	return false, fmt.Errorf("failed to check folder share: %w", err)
}

// GetUsersByEmails retrieves users by their email addresses
func (r *FolderShareRepository) GetUsersByEmails(ctx context.Context, emails []string) ([]entity.User, error) {
	var users []entity.User
	if err := r.db.WithContext(ctx).
		Where("email IN ? AND deleted_at IS NULL", emails).
		Find(&users).Error; err != nil {
		return nil, fmt.Errorf("failed to get users by emails: %w", err)
	}
	return users, nil
}
