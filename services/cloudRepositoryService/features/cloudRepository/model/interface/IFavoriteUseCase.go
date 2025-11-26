package _interface

import (
	"context"

	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/request"
	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/response"
)

// IFavoriteUseCase defines methods for favorite business logic
type IFavoriteUseCase interface {
	AddFavorite(ctx context.Context, userID, fileID uint) (*response.FavoriteResponseDTO, error)
	RemoveFavorite(ctx context.Context, userID, fileID uint) error
	ListFavorites(ctx context.Context, userID uint, filter request.ListFavoritesRequestDTO) (*response.ListFavoritesResponseDTO, error)
}
