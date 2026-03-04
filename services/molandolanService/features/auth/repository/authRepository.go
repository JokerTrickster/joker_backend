package repository

import (
	"context"
	"fmt"

	"github.com/JokerTrickster/joker_backend/services/molandolanService/features/auth/model/entity"
	_interface "github.com/JokerTrickster/joker_backend/services/molandolanService/features/auth/model/interface"
	"gorm.io/gorm"
)

type AuthRepository struct {
	db *gorm.DB
}

func NewAuthRepository(db *gorm.DB) _interface.IAuthRepository {
	return &AuthRepository{db: db}
}

func (r *AuthRepository) FindUserByEmail(ctx context.Context, email string) (*entity.MorandoranUser, error) {
	var user entity.MorandoranUser
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		return nil, fmt.Errorf("user not found")
	}
	return &user, nil
}

func (r *AuthRepository) FindUserByID(ctx context.Context, userID uint) (*entity.MorandoranUser, error) {
	var user entity.MorandoranUser
	if err := r.db.WithContext(ctx).Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, fmt.Errorf("user not found")
	}
	return &user, nil
}

func (r *AuthRepository) FindOrCreateByOAuth(ctx context.Context, email, nickname, provider string, profileImage *string) (*entity.MorandoranUser, error) {
	var user entity.MorandoranUser
	err := r.db.WithContext(ctx).Where("email = ? AND provider = ?", email, provider).First(&user).Error
	if err == nil {
		updates := map[string]interface{}{"profile_image": profileImage, "nickname": nickname}
		if err := r.db.WithContext(ctx).Model(&user).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("failed to update user: %w", err)
		}
		return &user, nil
	}
	user = entity.MorandoranUser{
		Email:        email,
		Nickname:     nickname,
		Provider:     provider,
		ProfileImage: profileImage,
		Role:         "user",
	}
	if err := r.db.WithContext(ctx).Create(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	return &user, nil
}

func (r *AuthRepository) UpdateNickname(ctx context.Context, userID uint, nickname string) (*entity.MorandoranUser, error) {
	var user entity.MorandoranUser
	if err := r.db.WithContext(ctx).First(&user, userID).Error; err != nil {
		return nil, fmt.Errorf("user not found")
	}
	user.Nickname = nickname
	if err := r.db.WithContext(ctx).Save(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to update nickname: %w", err)
	}
	return &user, nil
}
