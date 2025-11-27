package usecase

import (
	"context"
	"fmt"
	"time"

	_interface "github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/interface"
	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/request"
	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/response"
)

type FavoriteUseCase struct {
	FavoriteRepo   _interface.IFavoriteRepository
	FileRepo       _interface.IDownloadCloudRepositoryRepository // For file validation and ownership
	ListRepo       _interface.IListCloudRepositoryRepository     // For presigned URL generation
	ContextTimeout time.Duration
}

func NewFavoriteUseCase(
	favoriteRepo _interface.IFavoriteRepository,
	fileRepo _interface.IDownloadCloudRepositoryRepository,
	listRepo _interface.IListCloudRepositoryRepository,
	timeout time.Duration,
) _interface.IFavoriteUseCase {
	return &FavoriteUseCase{
		FavoriteRepo:   favoriteRepo,
		FileRepo:       fileRepo,
		ListRepo:       listRepo,
		ContextTimeout: timeout,
	}
}

// AddFavorite validates file exists and user owns it, then adds to favorites
func (u *FavoriteUseCase) AddFavorite(c context.Context, userID, fileID uint) (*response.FavoriteResponseDTO, error) {
	ctx, cancel := context.WithTimeout(c, u.ContextTimeout)
	defer cancel()

	// Validate file exists
	file, err := u.FileRepo.GetFileByID(ctx, fileID)
	if err != nil {
		return nil, fmt.Errorf("file not found")
	}

	// Verify ownership
	if file.UserID != userID {
		return nil, fmt.Errorf("access denied: you do not own this file")
	}

	// Add favorite (idempotent - won't error if already favorited)
	favorite, err := u.FavoriteRepo.AddFavorite(ctx, userID, fileID)
	if err != nil {
		return nil, fmt.Errorf("failed to add favorite: %w", err)
	}

	return &response.FavoriteResponseDTO{
		Success:     true,
		FavoritedAt: favorite.FavoritedAt,
	}, nil
}

// RemoveFavorite removes a file from favorites (idempotent)
func (u *FavoriteUseCase) RemoveFavorite(c context.Context, userID, fileID uint) error {
	ctx, cancel := context.WithTimeout(c, u.ContextTimeout)
	defer cancel()

	// Idempotent - no error if favorite doesn't exist
	return u.FavoriteRepo.RemoveFavorite(ctx, userID, fileID)
}

// ListFavorites retrieves favorited files with presigned URLs and pagination
func (u *FavoriteUseCase) ListFavorites(c context.Context, userID uint, filter request.ListFavoritesRequestDTO) (*response.ListFavoritesResponseDTO, error) {
	ctx, cancel := context.WithTimeout(c, u.ContextTimeout)
	defer cancel()

	// Set defaults
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Size < 1 {
		filter.Size = 20
	}
	if filter.Size > 100 {
		filter.Size = 100
	}

	// Get favorited files
	files, total, err := u.FavoriteRepo.GetFavoritesByUserID(ctx, userID, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list favorites: %w", err)
	}

	// Generate presigned URLs for each file
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

		// Generate presigned download URL
		downloadURL, err := u.ListRepo.GeneratePresignedDownloadURL(ctx, file.S3Key, 1*time.Hour)
		if err != nil {
			// Log error but don't fail the entire request
			downloadURL = ""
		}

		// Generate presigned thumbnail URL if available
		thumbnailURL := ""
		if file.ThumbnailKey != "" {
			thumbnailURL, err = u.ListRepo.GeneratePresignedDownloadURL(ctx, file.ThumbnailKey, 1*time.Hour)
			if err != nil {
				// Log error but don't fail the entire request
				thumbnailURL = ""
			}
		}

		fileInfos[i] = response.FileInfoDTO{
			ID:           file.ID,
			FileName:     file.FileName,
			FileType:     string(file.FileType),
			ContentType:  file.ContentType,
			FileSize:     file.FileSize,
			Duration:     file.Duration,
			Tags:         tagDTOs,
			DownloadURL:  downloadURL,
			ThumbnailURL: thumbnailURL,
			CreatedAt:    file.CreatedAt.Format(time.RFC3339),
			UpdatedAt:    file.UpdatedAt.Format(time.RFC3339),
		}
	}

	// Calculate total pages
	totalPages := int(total) / filter.Size
	if int(total)%filter.Size != 0 {
		totalPages++
	}

	return &response.ListFavoritesResponseDTO{
		Data: fileInfos,
		Pagination: response.PaginationMeta{
			Total:      total,
			Page:       filter.Page,
			Size:       filter.Size,
			TotalPages: totalPages,
		},
	}, nil
}
