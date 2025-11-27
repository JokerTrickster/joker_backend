package _interface

import (
	"context"

	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/entity"
	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/request"
)

// IFavoriteRepository defines methods for favorite operations
type IFavoriteRepository interface {
	AddFavorite(ctx context.Context, userID, fileID uint) (*entity.Favorite, error)
	RemoveFavorite(ctx context.Context, userID, fileID uint) error
	GetFavoritesByUserID(ctx context.Context, userID uint, filter request.ListFavoritesRequestDTO) ([]entity.CloudFile, int64, error)
	CheckIsFavorited(ctx context.Context, userID, fileID uint) (bool, error)
}
