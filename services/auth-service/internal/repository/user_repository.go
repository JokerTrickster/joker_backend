package repository

import (
	"context"
	"database/sql"

	"github.com/luxrobo/joker_backend/services/auth-service/internal/model"
	"github.com/luxrobo/joker_backend/shared/database"
)

type UserRepository struct {
	db *database.DB
}

func NewUserRepository(db *database.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) FindByID(ctx context.Context, id int64) (*model.User, error) {
	query := "SELECT id, name, email, created_at, updated_at FROM users WHERE id = ?"

	user := &model.User{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	query := "INSERT INTO users (name, email) VALUES (?, ?)"

	result, err := r.db.ExecContext(ctx, query, user.Name, user.Email)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	user.ID = id
	return nil
}
