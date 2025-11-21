package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model"
)

// ListFiles lists all files for a user with filtering and pagination
func (u *CloudRepositoryUsecase) ListFiles(ctx context.Context, userID uint, req model.ListFilesRequestDTO) (*model.ListFilesResponseDTO, error) {
	// Set defaults
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 20
	}

	// Validate file type if provided
	if req.FileType != "" {
		ft := model.FileType(req.FileType)
		if ft != model.FileTypeImage && ft != model.FileTypeVideo {
			return nil, fmt.Errorf("invalid file type: %s", req.FileType)
		}
	}

	files, total, err := u.repo.GetFilesByUserID(ctx, userID, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	fileInfos := make([]model.FileInfoDTO, len(files))
	for i, file := range files {
		// Map tags
		tagDTOs := make([]model.TagDTO, len(file.Tags))
		for j, tag := range file.Tags {
			tagDTOs[j] = model.TagDTO{
				ID:   tag.ID,
				Name: tag.Name,
			}
		}

		fileInfos[i] = model.FileInfoDTO{
			ID:          file.ID,
			FileName:    file.FileName,
			FileType:    string(file.FileType),
			ContentType: file.ContentType,
			FileSize:    file.FileSize,
			Tags:        tagDTOs,
			CreatedAt:   file.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   file.UpdatedAt.Format(time.RFC3339),
		}
	}

	return &model.ListFilesResponseDTO{
		Files:      fileInfos,
		TotalCount: total,
		Page:       req.Page,
		PageSize:   req.PageSize,
	}, nil
}
