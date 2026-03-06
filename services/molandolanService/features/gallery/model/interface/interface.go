package _interface

import (
	"context"

	"github.com/JokerTrickster/joker_backend/services/molandolanService/features/gallery/model/entity"
)

type IGalleryRepository interface {
	List(ctx context.Context, page, limit int) ([]entity.GalleryPost, int64, error)
	FindByID(ctx context.Context, id uint) (*entity.GalleryPost, error)
	Create(ctx context.Context, post *entity.GalleryPost) (*entity.GalleryPost, error)
	Delete(ctx context.Context, id uint) error

	IsLiked(ctx context.Context, userID, galleryID uint) (bool, error)
	ToggleLike(ctx context.Context, userID, galleryID uint) (bool, int, error)

	ListComments(ctx context.Context, galleryID uint, page, limit int) ([]entity.GalleryComment, int64, error)
	FindCommentByID(ctx context.Context, id uint) (*entity.GalleryComment, error)
	CreateComment(ctx context.Context, comment *entity.GalleryComment) (*entity.GalleryComment, error)
	DeleteComment(ctx context.Context, id uint) error

	GetAuthorNickname(ctx context.Context, userID uint) (string, error)
	GetAuthorInfo(ctx context.Context, userID uint) (string, *string, error)
	IsLikedBatch(ctx context.Context, userID uint, galleryIDs []uint) (map[uint]bool, error)
	GetUserRole(ctx context.Context, userID uint) (string, error)
}
