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

type mockUpdateMeAuthRepository struct {
	updateNicknameFunc func(ctx context.Context, userID uint, nickname string) (*entity.MorandoranUser, error)
}

func (m *mockUpdateMeAuthRepository) FindUserByEmail(ctx context.Context, email string) (*entity.MorandoranUser, error) {
	return nil, nil
}
func (m *mockUpdateMeAuthRepository) FindUserByID(ctx context.Context, userID uint) (*entity.MorandoranUser, error) {
	return nil, nil
}
func (m *mockUpdateMeAuthRepository) FindOrCreateByOAuth(ctx context.Context, email, nickname, provider string, profileImage *string) (*entity.MorandoranUser, error) {
	return nil, nil
}
func (m *mockUpdateMeAuthRepository) UpdateNickname(ctx context.Context, userID uint, nickname string) (*entity.MorandoranUser, error) {
	if m.updateNicknameFunc != nil {
		return m.updateNicknameFunc(ctx, userID, nickname)
	}
	return nil, nil
}

var _ _interface.IAuthRepository = (*mockUpdateMeAuthRepository)(nil)

func TestUpdateMeUseCase_Success(t *testing.T) {
	t.Log("UpdateNickname: success -> returns updated user")
	pic := "https://example.com/pic.jpg"
	mockRepo := &mockUpdateMeAuthRepository{
		updateNicknameFunc: func(ctx context.Context, userID uint, nickname string) (*entity.MorandoranUser, error) {
			assert.Equal(t, uint(1), userID)
			assert.Equal(t, "newnick", nickname)
			return &entity.MorandoranUser{
				ID:           1,
				Nickname:     "newnick",
				Email:        "user@example.com",
				Role:         "user",
				Provider:     "google",
				ProfileImage: &pic,
			}, nil
		},
	}

	uc := NewUpdateMeUseCase(mockRepo, 10*time.Second)
	res, err := uc.UpdateNickname(context.Background(), 1, "newnick")

	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "user-001", res.ID)
	assert.Equal(t, "newnick", res.Nickname)
	assert.Equal(t, "user@example.com", res.Email)
	assert.Equal(t, "google", res.Provider)
	require.NotNil(t, res.ProfileImage)
	assert.Equal(t, pic, *res.ProfileImage)
	t.Logf("Updated: nickname=%s provider=%s profileImage=%s", res.Nickname, res.Provider, *res.ProfileImage)
}

func TestUpdateMeUseCase_UserNotFound(t *testing.T) {
	t.Log("UpdateNickname: user not found -> error")
	mockRepo := &mockUpdateMeAuthRepository{
		updateNicknameFunc: func(ctx context.Context, userID uint, nickname string) (*entity.MorandoranUser, error) {
			return nil, assert.AnError
		},
	}

	uc := NewUpdateMeUseCase(mockRepo, 10*time.Second)
	res, err := uc.UpdateNickname(context.Background(), 999, "newnick")

	require.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "failed to update nickname")
	t.Logf("Error: %v", err)
}

func TestUpdateMeUseCase_NilProfileImage(t *testing.T) {
	t.Log("UpdateNickname: nil profile image -> returns nil in response")
	mockRepo := &mockUpdateMeAuthRepository{
		updateNicknameFunc: func(ctx context.Context, userID uint, nickname string) (*entity.MorandoranUser, error) {
			return &entity.MorandoranUser{
				ID:       1,
				Nickname: nickname,
				Email:    "user@example.com",
				Role:     "user",
				Provider: "email",
			}, nil
		},
	}

	uc := NewUpdateMeUseCase(mockRepo, 10*time.Second)
	res, err := uc.UpdateNickname(context.Background(), 1, "newnick")

	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Nil(t, res.ProfileImage)
	assert.Equal(t, "email", res.Provider)
	t.Logf("ProfileImage is nil, Provider=%s", res.Provider)
}
