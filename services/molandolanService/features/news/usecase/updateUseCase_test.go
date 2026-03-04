package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/JokerTrickster/joker_backend/services/morandoranService/features/news/model/entity"
	"github.com/JokerTrickster/joker_backend/services/morandoranService/features/news/model/request"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/mock"
)

type mockUpdateNewsRepo struct {
	mock.Mock
}

func (m *mockUpdateNewsRepo) List(ctx context.Context, page, limit int, category string) ([]entity.News, int64, error) {
	args := m.Called(ctx, page, limit, category)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]entity.News), args.Get(1).(int64), args.Error(2)
}

func (m *mockUpdateNewsRepo) FindByID(ctx context.Context, id uint) (*entity.News, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.News), args.Error(1)
}

func (m *mockUpdateNewsRepo) Create(ctx context.Context, news *entity.News) (*entity.News, error) {
	args := m.Called(ctx, news)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.News), args.Error(1)
}

func (m *mockUpdateNewsRepo) Update(ctx context.Context, id uint, updates map[string]interface{}) (*entity.News, error) {
	args := m.Called(ctx, id, updates)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.News), args.Error(1)
}

func (m *mockUpdateNewsRepo) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestUpdateUseCase_Update_Success(t *testing.T) {
	t.Log("Update: success - updates and returns formatted news")
	mockRepo := new(mockUpdateNewsRepo)
	uc := NewUpdateUseCase(mockRepo, 5*time.Second)
	ctx := context.Background()
	title := "New Title"
	req := &request.ReqUpdateNews{Title: &title}

	updated := &entity.News{
		ID: 1, Title: "New Title", Summary: "S", Content: "C", Thumbnail: "", Category: "cat", Date: "2024-01-15",
	}
	mockRepo.On("Update", mock.Anything, uint(1), mock.MatchedBy(func(m map[string]interface{}) bool {
		return m["title"] == "New Title"
	})).Return(updated, nil)

	res, err := uc.Update(ctx, 1, req)
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "news-001", res.ID)
	assert.Equal(t, "New Title", res.Title)
	t.Logf("Update result: %+v", res)
	mockRepo.AssertExpectations(t)
}

func TestUpdateUseCase_Update_EmptyUpdates_ReturnsFindByID(t *testing.T) {
	t.Log("Update: empty updates returns FindByID result")
	mockRepo := new(mockUpdateNewsRepo)
	uc := NewUpdateUseCase(mockRepo, 5*time.Second)
	ctx := context.Background()
	req := &request.ReqUpdateNews{}

	existing := &entity.News{
		ID: 3, Title: "Existing", Summary: "S", Content: "C", Thumbnail: "", Category: "cat", Date: "2024-01-15",
	}
	mockRepo.On("FindByID", mock.Anything, uint(3)).Return(existing, nil)

	res, err := uc.Update(ctx, 3, req)
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "news-003", res.ID)
	assert.Equal(t, "Existing", res.Title)
	mockRepo.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "Update")
}

func TestUpdateUseCase_Update_RepoError(t *testing.T) {
	t.Log("Update: repo error propagates")
	mockRepo := new(mockUpdateNewsRepo)
	uc := NewUpdateUseCase(mockRepo, 5*time.Second)
	ctx := context.Background()
	title := "X"
	req := &request.ReqUpdateNews{Title: &title}

	mockRepo.On("Update", mock.Anything, uint(999), mock.Anything).Return(nil, assert.AnError)

	res, err := uc.Update(ctx, 999, req)
	assert.Error(t, err)
	assert.Nil(t, res)
	mockRepo.AssertExpectations(t)
}
