package repository

import (
	"context"
	"fmt"

	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/entity"
	_interface "github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/interface"
	"gorm.io/gorm"
)

type FolderRepository struct {
	db *gorm.DB
}

func NewFolderRepository(db *gorm.DB) _interface.IFolderRepository {
	return &FolderRepository{
		db: db,
	}
}

// CreateFolder creates a new folder
func (r *FolderRepository) CreateFolder(ctx context.Context, folder *entity.Folder) error {
	// Check if parent folder exists (if parent_folder_id is provided)
	if folder.ParentFolderID != nil {
		var parentFolder entity.Folder
		if err := r.db.WithContext(ctx).
			Where("id = ? AND user_id = ? AND deleted_at IS NULL", *folder.ParentFolderID, folder.UserID).
			First(&parentFolder).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return fmt.Errorf("parent folder not found")
			}
			return fmt.Errorf("failed to check parent folder: %w", err)
		}
	}

	// Create the folder
	if err := r.db.WithContext(ctx).Create(folder).Error; err != nil {
		return fmt.Errorf("failed to create folder: %w", err)
	}

	return nil
}

// GetFolderByID retrieves a folder by ID (owner only)
func (r *FolderRepository) GetFolderByID(ctx context.Context, id uint, userID int32) (*entity.Folder, error) {
	var folder entity.Folder
	if err := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ? AND deleted_at IS NULL", id, userID).
		First(&folder).Error; err != nil {
		return nil, err
	}
	return &folder, nil
}

// GetFolderByIDWithoutUserCheck retrieves a folder by ID without user check (use after permission check)
func (r *FolderRepository) GetFolderByIDWithoutUserCheck(ctx context.Context, id uint) (*entity.Folder, error) {
	var folder entity.Folder
	if err := r.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&folder).Error; err != nil {
		return nil, err
	}
	return &folder, nil
}

// GetFoldersByUserID retrieves all folders for a user
func (r *FolderRepository) GetFoldersByUserID(ctx context.Context, userID int32) ([]entity.Folder, error) {
	var folders []entity.Folder
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND deleted_at IS NULL", userID).
		Order("created_at ASC").
		Find(&folders).Error; err != nil {
		return nil, err
	}
	return folders, nil
}

// UpdateFolder updates folder name or parent folder
func (r *FolderRepository) UpdateFolder(ctx context.Context, folder *entity.Folder) error {
	// If parent_folder_id is being changed, validate it
	if folder.ParentFolderID != nil {
		// Check if parent folder exists
		var parentFolder entity.Folder
		if err := r.db.WithContext(ctx).
			Where("id = ? AND user_id = ? AND deleted_at IS NULL", *folder.ParentFolderID, folder.UserID).
			First(&parentFolder).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return fmt.Errorf("parent folder not found")
			}
			return fmt.Errorf("failed to check parent folder: %w", err)
		}

		// Prevent circular reference (folder cannot be its own parent or descendant)
		if *folder.ParentFolderID == folder.ID {
			return fmt.Errorf("folder cannot be its own parent")
		}
		// TODO: Add recursive check to prevent circular references
	}

	// Update the folder
	if err := r.db.WithContext(ctx).
		Model(&entity.Folder{}).
		Where("id = ? AND user_id = ? AND deleted_at IS NULL", folder.ID, folder.UserID).
		Updates(folder).Error; err != nil {
		return fmt.Errorf("failed to update folder: %w", err)
	}

	return nil
}

// DeleteFolder soft deletes a folder (sets deleted_at)
func (r *FolderRepository) DeleteFolder(ctx context.Context, id uint, userID int32) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Check if folder exists and belongs to user
		var folder entity.Folder
		if err := tx.Where("id = ? AND user_id = ? AND deleted_at IS NULL", id, userID).
			First(&folder).Error; err != nil {
			return err
		}

		// Check if folder has sub-folders
		var subFolderCount int64
		if err := tx.Model(&entity.Folder{}).
			Where("parent_folder_id = ? AND user_id = ? AND deleted_at IS NULL", id, userID).
			Count(&subFolderCount).Error; err != nil {
			return fmt.Errorf("failed to check sub-folders: %w", err)
		}

		if subFolderCount > 0 {
			return fmt.Errorf("cannot delete folder with %d sub-folders", subFolderCount)
		}

		// Soft delete the folder
		if err := tx.Model(&entity.Folder{}).
			Where("id = ? AND user_id = ?", id, userID).
			Update("deleted_at", gorm.Expr("NOW()")).Error; err != nil {
			return fmt.Errorf("failed to delete folder: %w", err)
		}

		// Move files in this folder to root (set folder_id to NULL)
		if err := tx.Model(&entity.CloudFile{}).
			Where("folder_id = ? AND user_id = ?", id, userID).
			Update("folder_id", nil).Error; err != nil {
			return fmt.Errorf("failed to move files to root: %w", err)
		}

		return nil
	})
}

// GetFolderFileCount returns the number of files in a folder
func (r *FolderRepository) GetFolderFileCount(ctx context.Context, folderID uint, userID int32) (int, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&entity.CloudFile{}).
		Where("folder_id = ? AND user_id = ? AND deleted_at IS NULL", folderID, userID).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return int(count), nil
}

// GetFilesByFolderID retrieves all files in a folder
func (r *FolderRepository) GetFilesByFolderID(ctx context.Context, folderID *uint, userID int32) ([]entity.CloudFile, error) {
	var files []entity.CloudFile
	query := r.db.WithContext(ctx).
		Where("user_id = ? AND deleted_at IS NULL", userID)

	if folderID == nil {
		// Get root files (files without folder)
		query = query.Where("folder_id IS NULL")
	} else {
		// Get files in specific folder
		query = query.Where("folder_id = ?", *folderID)
	}

	if err := query.Order("created_at DESC").Find(&files).Error; err != nil {
		return nil, err
	}
	return files, nil
}

// MoveFilesToFolder moves multiple files to a folder (or root if folderID is nil)
func (r *FolderRepository) MoveFilesToFolder(ctx context.Context, fileIDs []uint, folderID *uint, userID int32) (int, error) {
	// If folderID is provided, verify folder exists and belongs to user
	if folderID != nil {
		var folder entity.Folder
		if err := r.db.WithContext(ctx).
			Where("id = ? AND user_id = ? AND deleted_at IS NULL", *folderID, userID).
			First(&folder).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return 0, fmt.Errorf("target folder not found")
			}
			return 0, fmt.Errorf("failed to check target folder: %w", err)
		}
	}

	// Update files to new folder
	result := r.db.WithContext(ctx).
		Model(&entity.CloudFile{}).
		Where("id IN ? AND user_id = ? AND deleted_at IS NULL", fileIDs, userID).
		Update("folder_id", folderID)

	if result.Error != nil {
		return 0, fmt.Errorf("failed to move files: %w", result.Error)
	}

	return int(result.RowsAffected), nil
}
