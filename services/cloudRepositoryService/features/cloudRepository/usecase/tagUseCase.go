package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/entity"
	_interface "github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/interface"
	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/request"
	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/response"
)

type TagUseCase struct {
	TagRepo        _interface.ITagRepository
	StatsRepo      _interface.IUserStatsCloudRepositoryRepository
	ContextTimeout time.Duration
}

func NewTagUseCase(
	tagRepo _interface.ITagRepository,
	statsRepo _interface.IUserStatsCloudRepositoryRepository,
	timeout time.Duration,
) _interface.ITagUseCase {
	return &TagUseCase{
		TagRepo:        tagRepo,
		StatsRepo:      statsRepo,
		ContextTimeout: timeout,
	}
}

// UpdateFileTags replaces all tags for a file
func (u *TagUseCase) UpdateFileTags(
	c context.Context,
	userID uint,
	fileID uint,
	req *request.UpdateFileTagsRequestDTO,
) (*response.UpdateFileTagsResponseDTO, error) {
	ctx, cancel := context.WithTimeout(c, u.ContextTimeout)
	defer cancel()

	// Get the file to ensure it exists
	file, err := u.TagRepo.GetFileByID(ctx, fileID, userID)
	if err != nil {
		return nil, fmt.Errorf("file not found: %w", err)
	}

	// Process tags
	tags := make([]entity.Tag, 0, len(req.Tags))
	tagNames := make([]string, 0, len(req.Tags))

	for _, tagName := range req.Tags {
		if tagName == "" {
			continue
		}

		// Find or create tag
		tag, err := u.TagRepo.FindOrCreateTag(ctx, userID, tagName)
		if err != nil {
			return nil, fmt.Errorf("failed to process tag %s: %w", tagName, err)
		}

		tags = append(tags, *tag)
		tagNames = append(tagNames, tagName)

		// Log tag activity if it's a new tag on this file
		if u.StatsRepo != nil {
			activity := &entity.ActivityLog{
				UserID:       userID,
				FileID:       &fileID,
				ActivityType: entity.ActivityTypeTagAdd,
				TagName:      tagName,
			}
			_ = u.StatsRepo.LogActivity(ctx, activity) // Don't fail on logging error
		}
	}

	// Update file tags
	if err := u.TagRepo.UpdateFileTags(ctx, fileID, userID, tags); err != nil {
		return nil, fmt.Errorf("failed to update tags: %w", err)
	}

	return &response.UpdateFileTagsResponseDTO{
		FileID:    file.ID,
		Tags:      tagNames,
		UpdatedAt: time.Now(),
	}, nil
}

// AddTagToFile adds a single tag to a file
func (u *TagUseCase) AddTagToFile(
	c context.Context,
	userID uint,
	fileID uint,
	req *request.AddFileTagRequestDTO,
) (*response.AddFileTagResponseDTO, error) {
	ctx, cancel := context.WithTimeout(c, u.ContextTimeout)
	defer cancel()

	// Get the file to ensure it exists
	file, err := u.TagRepo.GetFileByID(ctx, fileID, userID)
	if err != nil {
		return nil, fmt.Errorf("file not found: %w", err)
	}

	// Find or create tag
	tag, err := u.TagRepo.FindOrCreateTag(ctx, userID, req.Tag)
	if err != nil {
		return nil, fmt.Errorf("failed to process tag: %w", err)
	}

	// Add tag to file
	if err := u.TagRepo.AddTagToFile(ctx, fileID, userID, *tag); err != nil {
		return nil, fmt.Errorf("failed to add tag: %w", err)
	}

	// Log tag activity
	if u.StatsRepo != nil {
		activity := &entity.ActivityLog{
			UserID:       userID,
			FileID:       &fileID,
			ActivityType: entity.ActivityTypeTagAdd,
			TagName:      req.Tag,
		}
		_ = u.StatsRepo.LogActivity(ctx, activity) // Don't fail on logging error
	}

	return &response.AddFileTagResponseDTO{
		FileID:    file.ID,
		Tag:       req.Tag,
		UpdatedAt: time.Now(),
	}, nil
}

// RemoveTagFromFile removes a tag from a file
func (u *TagUseCase) RemoveTagFromFile(
	c context.Context,
	userID uint,
	fileID uint,
	tagName string,
) error {
	ctx, cancel := context.WithTimeout(c, u.ContextTimeout)
	defer cancel()

	// Get the file to ensure it exists
	_, err := u.TagRepo.GetFileByID(ctx, fileID, userID)
	if err != nil {
		return fmt.Errorf("file not found: %w", err)
	}

	// Remove tag from file
	if err := u.TagRepo.RemoveTagFromFile(ctx, fileID, userID, tagName); err != nil {
		return fmt.Errorf("failed to remove tag: %w", err)
	}

	// Log tag removal activity
	if u.StatsRepo != nil {
		activity := &entity.ActivityLog{
			UserID:       userID,
			FileID:       &fileID,
			ActivityType: entity.ActivityTypeTagDel,
			TagName:      tagName,
		}
		_ = u.StatsRepo.LogActivity(ctx, activity) // Don't fail on logging error
	}

	return nil
}
