package repository

import (
	"context"

	_interface "github.com/JokerTrickster/joker_backend/services/authService/features/auth/model/interface"

	"gorm.io/gorm"
)

func NewSignupAuthRepository(gormDB *gorm.DB) _interface.ISignupAuthRepository {
	return &SignupAuthRepository{GormDB: gormDB}
}

// Implement the interface methods
func (r *SignupAuthRepository) SignupAuth(ctx context.Context, req interface{}) (interface{}, error) {
	// Implementation needed
	return nil, nil
}
