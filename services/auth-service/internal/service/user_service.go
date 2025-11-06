package service

import (
	"github.com/luxrobo/joker_backend/internal/model"
	"github.com/luxrobo/joker_backend/internal/repository"
	"github.com/luxrobo/joker_backend/pkg/database"
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(db *database.DB) *UserService {
	return &UserService{
		repo: repository.NewUserRepository(db),
	}
}

func (s *UserService) GetUserByID(id int64) (*model.User, error) {
	return s.repo.FindByID(id)
}

func (s *UserService) CreateUser(req *model.CreateUserRequest) (*model.User, error) {
	user := &model.User{
		Name:  req.Name,
		Email: req.Email,
	}

	if err := s.repo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}
