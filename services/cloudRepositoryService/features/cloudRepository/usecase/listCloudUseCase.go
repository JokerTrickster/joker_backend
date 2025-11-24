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

type ListCloudRepositoryUseCase struct {
	Repo           _interface.IListCloudRepositoryRepository
	ContextTimeout time.Duration
}

func NewListCloudRepositoryUseCase(repo _interface.IListCloudRepositoryRepository, timeout time.Duration) _interface.IListCloudRepositoryUseCase {
	return &ListCloudRepositoryUseCase{
		Repo:           repo,
		ContextTimeout: timeout,
	}
}

// ListFiles lists all files for a user with filtering and pagination
func (u *ListCloudRepositoryUseCase) ListFiles(c context.Context, userID uint, req request.ListFilesRequestDTO) (*response.ListFilesResponseDTO, error) {
	ctx, cancel := context.WithTimeout(c, u.ContextTimeout)
	defer cancel()
	// Set defaults
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 20
	}

	// Validate file type if provided
	if req.FileType != "" {
		ft := entity.FileType(req.FileType)
		if ft != entity.FileTypeImage && ft != entity.FileTypeVideo {
			return nil, fmt.Errorf("invalid file type: %s", req.FileType)
		}
	}

	files, total, err := u.Repo.GetFilesByUserID(ctx, userID, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	fileInfos := make([]response.FileInfoDTO, len(files))
	for i, file := range files {
		// Map tags
		tagDTOs := make([]response.TagDTO, len(file.Tags))
		for j, tag := range file.Tags {
			tagDTOs[j] = response.TagDTO{
				ID:   tag.ID,
				Name: tag.Name,
			}
		}

		fileInfos[i] = response.FileInfoDTO{
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

	return &response.ListFilesResponseDTO{
		Files:      fileInfos,
		TotalCount: total,
		Page:       req.Page,
		PageSize:   req.PageSize,
	}, nil
}
