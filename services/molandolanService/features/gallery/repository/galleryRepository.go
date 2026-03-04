package repository

import (
	"context"
	"fmt"

	authEntity "github.com/JokerTrickster/joker_backend/services/molandolanService/features/auth/model/entity"
	"github.com/JokerTrickster/joker_backend/services/molandolanService/features/gallery/model/entity"
	_interface "github.com/JokerTrickster/joker_backend/services/molandolanService/features/gallery/model/interface"
	"gorm.io/gorm"
)

type GalleryRepository struct {
	db *gorm.DB
}

func NewGalleryRepository(db *gorm.DB) _interface.IGalleryRepository {
	return &GalleryRepository{db: db}
}

func (r *GalleryRepository) List(ctx context.Context, page, limit int) ([]entity.GalleryPost, int64, error) {
	var items []entity.GalleryPost
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.GalleryPost{})
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count gallery posts: %w", err)
	}

	offset := (page - 1) * limit
	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&items).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list gallery posts: %w", err)
	}

	return items, total, nil
}

func (r *GalleryRepository) FindByID(ctx context.Context, id uint) (*entity.GalleryPost, error) {
	var post entity.GalleryPost
	if err := r.db.WithContext(ctx).First(&post, id).Error; err != nil {
		return nil, fmt.Errorf("gallery post not found")
	}
	return &post, nil
}

func (r *GalleryRepository) Create(ctx context.Context, post *entity.GalleryPost) (*entity.GalleryPost, error) {
	if err := r.db.WithContext(ctx).Create(post).Error; err != nil {
		return nil, fmt.Errorf("failed to create gallery post: %w", err)
	}
	return post, nil
}

func (r *GalleryRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&entity.GalleryPost{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete gallery post: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("gallery post not found")
	}
	return nil
}

func (r *GalleryRepository) IsLiked(ctx context.Context, userID, galleryID uint) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&entity.GalleryLike{}).
		Where("user_id = ? AND gallery_id = ?", userID, galleryID).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *GalleryRepository) ToggleLike(ctx context.Context, userID, galleryID uint) (bool, int, error) {
	var like entity.GalleryLike
	err := r.db.WithContext(ctx).Where("user_id = ? AND gallery_id = ?", userID, galleryID).First(&like).Error

	if err == gorm.ErrRecordNotFound {
		newLike := entity.GalleryLike{UserID: userID, GalleryID: galleryID}
		if err := r.db.WithContext(ctx).Create(&newLike).Error; err != nil {
			return false, 0, fmt.Errorf("failed to create like: %w", err)
		}
		r.db.WithContext(ctx).Model(&entity.GalleryPost{}).Where("id = ?", galleryID).
			UpdateColumn("like_count", gorm.Expr("like_count + 1"))
	} else if err == nil {
		if err := r.db.WithContext(ctx).Delete(&like).Error; err != nil {
			return false, 0, fmt.Errorf("failed to delete like: %w", err)
		}
		r.db.WithContext(ctx).Model(&entity.GalleryPost{}).Where("id = ?", galleryID).
			UpdateColumn("like_count", gorm.Expr("GREATEST(like_count - 1, 0)"))
	} else {
		return false, 0, fmt.Errorf("failed to toggle like: %w", err)
	}

	var post entity.GalleryPost
	r.db.WithContext(ctx).Select("like_count").First(&post, galleryID)

	isLiked := err == gorm.ErrRecordNotFound
	return isLiked, post.LikeCount, nil
}

func (r *GalleryRepository) ListComments(ctx context.Context, galleryID uint, page, limit int) ([]entity.GalleryComment, int64, error) {
	var items []entity.GalleryComment
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.GalleryComment{}).Where("gallery_id = ?", galleryID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count comments: %w", err)
	}

	offset := (page - 1) * limit
	if err := query.Order("created_at ASC").Offset(offset).Limit(limit).Find(&items).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list comments: %w", err)
	}

	return items, total, nil
}

func (r *GalleryRepository) FindCommentByID(ctx context.Context, id uint) (*entity.GalleryComment, error) {
	var comment entity.GalleryComment
	if err := r.db.WithContext(ctx).First(&comment, id).Error; err != nil {
		return nil, fmt.Errorf("comment not found")
	}
	return &comment, nil
}

func (r *GalleryRepository) CreateComment(ctx context.Context, comment *entity.GalleryComment) (*entity.GalleryComment, error) {
	if err := r.db.WithContext(ctx).Create(comment).Error; err != nil {
		return nil, fmt.Errorf("failed to create comment: %w", err)
	}
	r.db.WithContext(ctx).Model(&entity.GalleryPost{}).Where("id = ?", comment.GalleryID).
		UpdateColumn("comment_count", gorm.Expr("comment_count + 1"))
	return comment, nil
}

func (r *GalleryRepository) DeleteComment(ctx context.Context, id uint) error {
	var comment entity.GalleryComment
	if err := r.db.WithContext(ctx).First(&comment, id).Error; err != nil {
		return fmt.Errorf("comment not found")
	}

	result := r.db.WithContext(ctx).Delete(&entity.GalleryComment{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete comment: %w", result.Error)
	}

	r.db.WithContext(ctx).Model(&entity.GalleryPost{}).Where("id = ?", comment.GalleryID).
		UpdateColumn("comment_count", gorm.Expr("GREATEST(comment_count - 1, 0)"))
	return nil
}

func (r *GalleryRepository) GetAuthorNickname(ctx context.Context, userID uint) (string, error) {
	var user authEntity.MorandoranUser
	if err := r.db.WithContext(ctx).Select("nickname").First(&user, userID).Error; err != nil {
		return "", fmt.Errorf("user not found")
	}
	return user.Nickname, nil
}

func (r *GalleryRepository) GetAuthorInfo(ctx context.Context, userID uint) (string, *string, error) {
	var user authEntity.MorandoranUser
	if err := r.db.WithContext(ctx).Select("nickname", "profile_image").First(&user, userID).Error; err != nil {
		return "", nil, fmt.Errorf("user not found")
	}
	return user.Nickname, user.ProfileImage, nil
}

func (r *GalleryRepository) IsLikedBatch(ctx context.Context, userID uint, galleryIDs []uint) (map[uint]bool, error) {
	result := make(map[uint]bool)
	if userID == 0 || len(galleryIDs) == 0 {
		return result, nil
	}
	var likes []entity.GalleryLike
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND gallery_id IN ?", userID, galleryIDs).
		Find(&likes).Error; err != nil {
		return nil, err
	}
	for _, like := range likes {
		result[like.GalleryID] = true
	}
	return result, nil
}
