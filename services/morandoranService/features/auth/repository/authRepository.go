package repository

import (
	"context"
	"fmt"

	"github.com/JokerTrickster/joker_backend/services/morandoranService/features/auth/model/entity"
	_interface "github.com/JokerTrickster/joker_backend/services/morandoranService/features/auth/model/interface"
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
