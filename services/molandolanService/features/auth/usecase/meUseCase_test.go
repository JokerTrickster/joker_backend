package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/JokerTrickster/joker_backend/services/molandolanService/features/auth/model/entity"
	_interface "github.com/JokerTrickster/joker_backend/services/molandolanService/features/auth/model/interface"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockMeAuthRepository struct {
	findByIDFunc func(ctx context.Context, userID uint) (*entity.MorandoranUser, error)
}

func (m *mockMeAuthRepository) FindUserByEmail(ctx context.Context, email string) (*entity.MorandoranUser, error) {
	return nil, nil
}

func (m *mockMeAuthRepository) FindUserByID(ctx context.Context, userID uint) (*entity.MorandoranUser, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, userID)
	}
	return nil, nil
}

func (m *mockMeAuthRepository) FindOrCreateByOAuth(ctx context.Context, email, nickname, provider string, profileImage *string) (*entity.MorandoranUser, error) {
	return nil, nil
}

func (m *mockMeAuthRepository) UpdateNickname(ctx context.Context, userID uint, nickname string) (*entity.MorandoranUser, error) {
	return nil, nil
}

var _ _interface.IAuthRepository = (*mockMeAuthRepository)(nil)

func TestMeUseCase_Me_Success(t *testing.T) {
	pic := "https://example.com/pic.jpg"
	mockRepo := &mockMeAuthRepository{
		findByIDFunc: func(ctx context.Context, userID uint) (*entity.MorandoranUser, error) {
			return &entity.MorandoranUser{
				ID:           1,
				Nickname:     "testuser",
				Email:        "user@example.com",
				Role:         "user",
				Provider:     "google",
				ProfileImage: &pic,
			}, nil
		},
	}
	uc := NewMeUseCase(mockRepo, 10*time.Second)
	ctx := context.Background()

	res, err := uc.Me(ctx, 1)
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "user-001", res.ID)
	assert.Equal(t, "testuser", res.Nickname)
	assert.Equal(t, "user@example.com", res.Email)
	assert.Equal(t, "user", res.Role)
	assert.Equal(t, "google", res.Provider)
	require.NotNil(t, res.ProfileImage)
	assert.Equal(t, pic, *res.ProfileImage)
	t.Logf("Me success: id=%s nickname=%s provider=%s profileImage=%s", res.ID, res.Nickname, res.Provider, *res.ProfileImage)
}

func TestMeUseCase_Me_UserNotFound(t *testing.T) {
	mockRepo := &mockMeAuthRepository{
		findByIDFunc: func(ctx context.Context, userID uint) (*entity.MorandoranUser, error) {
			return nil, assert.AnError
		},
	}
	uc := NewMeUseCase(mockRepo, 10*time.Second)
	ctx := context.Background()

	res, err := uc.Me(ctx, 999)
	require.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "user not found")
	t.Logf("User not found error: %v", err)
}
