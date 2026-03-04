package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/JokerTrickster/joker_backend/services/morandoranService/features/news/model/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/mock"
)

type mockDeleteNewsRepo struct {
	mock.Mock
}

func (m *mockDeleteNewsRepo) List(ctx context.Context, page, limit int, category string) ([]entity.News, int64, error) {
	args := m.Called(ctx, page, limit, category)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]entity.News), args.Get(1).(int64), args.Error(2)
}

func (m *mockDeleteNewsRepo) FindByID(ctx context.Context, id uint) (*entity.News, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.News), args.Error(1)
}

func (m *mockDeleteNewsRepo) Create(ctx context.Context, news *entity.News) (*entity.News, error) {
	args := m.Called(ctx, news)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.News), args.Error(1)
}

func (m *mockDeleteNewsRepo) Update(ctx context.Context, id uint, updates map[string]interface{}) (*entity.News, error) {
	args := m.Called(ctx, id, updates)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.News), args.Error(1)
}

func (m *mockDeleteNewsRepo) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestDeleteUseCase_Delete_Success(t *testing.T) {
	t.Log("Delete: success - calls repo and returns no error")
	mockRepo := new(mockDeleteNewsRepo)
	uc := NewDeleteUseCase(mockRepo, 5*time.Second)
	ctx := context.Background()

	mockRepo.On("Delete", mock.Anything, uint(1)).Return(nil)

	err := uc.Delete(ctx, 1)
	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestDeleteUseCase_Delete_RepoError(t *testing.T) {
	t.Log("Delete: repo error propagates")
	mockRepo := new(mockDeleteNewsRepo)
	uc := NewDeleteUseCase(mockRepo, 5*time.Second)
	ctx := context.Background()

	mockRepo.On("Delete", mock.Anything, uint(999)).Return(assert.AnError)

	err := uc.Delete(ctx, 999)
	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}
