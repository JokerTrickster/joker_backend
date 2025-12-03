package repository

import (
	"context"
	"fmt"

	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/entity"
	_interface "github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/interface"
	"gorm.io/gorm"
)

type TagRepository struct {
	db *gorm.DB
}

func NewTagRepository(db *gorm.DB) _interface.ITagRepository {
	return &TagRepository{
		db: db,
	}
}

// GetFileByID retrieves a file with its tags
func (r *TagRepository) GetFileByID(ctx context.Context, id uint, userID uint) (*entity.CloudFile, error) {
	var file entity.CloudFile
	if err := r.db.WithContext(ctx).
		Preload("Tags").
		Where("id = ? AND user_id = ? AND deleted_at IS NULL", id, userID).
		First(&file).Error; err != nil {
		return nil, err
	}
	return &file, nil
}

// UpdateFileTags replaces all tags for a file
func (r *TagRepository) UpdateFileTags(ctx context.Context, fileID uint, userID uint, tags []entity.Tag) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Get the file to ensure it exists and belongs to the user
		var file entity.CloudFile
		if err := tx.Where("id = ? AND user_id = ? AND deleted_at IS NULL", fileID, userID).
			First(&file).Error; err != nil {
			return err
		}

		// Clear existing tags
		if err := tx.Model(&file).Association("Tags").Clear(); err != nil {
			return fmt.Errorf("failed to clear existing tags: %w", err)
		}

		// Set new tags
		if len(tags) > 0 {
			if err := tx.Model(&file).Association("Tags").Append(tags); err != nil {
				return fmt.Errorf("failed to append new tags: %w", err)
			}
		}

		return nil
	})
}

// AddTagToFile adds a single tag to a file
func (r *TagRepository) AddTagToFile(ctx context.Context, fileID uint, userID uint, tag entity.Tag) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Get the file to ensure it exists and belongs to the user
		var file entity.CloudFile
		if err := tx.Preload("Tags").
			Where("id = ? AND user_id = ? AND deleted_at IS NULL", fileID, userID).
			First(&file).Error; err != nil {
			return err
		}

		// Check if tag already exists on this file
		for _, existingTag := range file.Tags {
			if existingTag.Name == tag.Name {
				return nil // Tag already exists, silently succeed
			}
		}

		// Add the tag
		if err := tx.Model(&file).Association("Tags").Append(&tag); err != nil {
			return fmt.Errorf("failed to add tag: %w", err)
		}

		return nil
	})
}

// RemoveTagFromFile removes a tag from a file by tag name
func (r *TagRepository) RemoveTagFromFile(ctx context.Context, fileID uint, userID uint, tagName string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Get the file to ensure it exists and belongs to the user
		var file entity.CloudFile
		if err := tx.Preload("Tags").
			Where("id = ? AND user_id = ? AND deleted_at IS NULL", fileID, userID).
			First(&file).Error; err != nil {
			return err
		}

		// Find the tag to remove
		var tagToRemove *entity.Tag
		for i := range file.Tags {
			if file.Tags[i].Name == tagName {
				tagToRemove = &file.Tags[i]
				break
			}
		}

		if tagToRemove == nil {
			return fmt.Errorf("tag not found on file: %s", tagName)
		}

		// Remove the association
		if err := tx.Model(&file).Association("Tags").Delete(tagToRemove); err != nil {
			return fmt.Errorf("failed to remove tag: %w", err)
		}

		return nil
	})
}

// FindOrCreateTag finds an existing tag or creates a new one
func (r *TagRepository) FindOrCreateTag(ctx context.Context, userID uint, tagName string) (*entity.Tag, error) {
	var tag entity.Tag
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND name = ?", userID, tagName).
		FirstOrCreate(&tag, entity.Tag{
			UserID: userID,
			Name:   tagName,
		}).Error; err != nil {
		return nil, err
	}
	return &tag, nil
}
