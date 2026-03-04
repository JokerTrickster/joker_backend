package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/JokerTrickster/joker_backend/services/molandolanService/features/auth/model/entity"
	_interface "github.com/JokerTrickster/joker_backend/services/molandolanService/features/auth/model/interface"
	"github.com/JokerTrickster/joker_backend/shared/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func initJWTKeys() {
	if len(jwt.AccessTokenSecretKey) == 0 {
		jwt.AccessTokenSecretKey = []byte("test-access-secret-key-for-unit-tests")
	}
	if len(jwt.RefreshTokenSecretKey) == 0 {
		jwt.RefreshTokenSecretKey = []byte("test-refresh-secret-key-for-unit-tests")
	}
}

type mockOAuthAuthRepository struct {
	findOrCreateByOAuthFunc func(ctx context.Context, email, nickname, provider string, profileImage *string) (*entity.MorandoranUser, error)
}

func (m *mockOAuthAuthRepository) FindUserByEmail(ctx context.Context, email string) (*entity.MorandoranUser, error) {
	return nil, nil
}
func (m *mockOAuthAuthRepository) FindUserByID(ctx context.Context, userID uint) (*entity.MorandoranUser, error) {
	return nil, nil
}
func (m *mockOAuthAuthRepository) FindOrCreateByOAuth(ctx context.Context, email, nickname, provider string, profileImage *string) (*entity.MorandoranUser, error) {
	if m.findOrCreateByOAuthFunc != nil {
		return m.findOrCreateByOAuthFunc(ctx, email, nickname, provider, profileImage)
	}
	return nil, nil
}
func (m *mockOAuthAuthRepository) UpdateNickname(ctx context.Context, userID uint, nickname string) (*entity.MorandoranUser, error) {
	return nil, nil
}

var _ _interface.IAuthRepository = (*mockOAuthAuthRepository)(nil)

func TestOAuthUseCase_HandleCallback_Success(t *testing.T) {
	t.Log("HandleCallback: creates user and returns token")
	initJWTKeys()

	mockRepo := &mockOAuthAuthRepository{
		findOrCreateByOAuthFunc: func(ctx context.Context, email, nickname, provider string, profileImage *string) (*entity.MorandoranUser, error) {
			assert.Equal(t, "user@google.com", email)
			assert.Equal(t, "GoogleUser", nickname)
			assert.Equal(t, "google", provider)
			return &entity.MorandoranUser{
				ID:       1,
				Nickname: "GoogleUser",
				Email:    "user@google.com",
				Role:     "user",
				Provider: "google",
			}, nil
		},
	}

	uc := NewOAuthUseCase(mockRepo, 10*time.Second)
	res, err := uc.HandleCallback(context.Background(), "user@google.com", "GoogleUser", "google", nil)

	require.NoError(t, err)
	require.NotNil(t, res)
	assert.NotEmpty(t, res.Token, "token should not be empty")
	assert.Equal(t, "user-001", res.User.ID)
	assert.Equal(t, "GoogleUser", res.User.Nickname)
	assert.Equal(t, "user@google.com", res.User.Email)
	assert.Equal(t, "user", res.User.Role)
	t.Logf("Token generated: %s...", res.Token[:20])
}

func TestOAuthUseCase_HandleCallback_WithProfileImage(t *testing.T) {
	t.Log("HandleCallback: with profile image")
	initJWTKeys()

	pic := "https://example.com/pic.jpg"
	var receivedPic *string
	mockRepo := &mockOAuthAuthRepository{
		findOrCreateByOAuthFunc: func(ctx context.Context, email, nickname, provider string, profileImage *string) (*entity.MorandoranUser, error) {
			receivedPic = profileImage
			return &entity.MorandoranUser{
				ID:           2,
				Nickname:     "KakaoUser",
				Email:        "user@kakao.com",
				Role:         "user",
				Provider:     "kakao",
				ProfileImage: profileImage,
			}, nil
		},
	}

	uc := NewOAuthUseCase(mockRepo, 10*time.Second)
	res, err := uc.HandleCallback(context.Background(), "user@kakao.com", "KakaoUser", "kakao", &pic)

	require.NoError(t, err)
	require.NotNil(t, res)
	require.NotNil(t, receivedPic)
	assert.Equal(t, pic, *receivedPic)
	assert.Equal(t, "user-002", res.User.ID)
	t.Logf("Profile image passed correctly: %s", *receivedPic)
}

func TestOAuthUseCase_HandleCallback_RepoError(t *testing.T) {
	t.Log("HandleCallback: repository error -> return error")
	initJWTKeys()

	mockRepo := &mockOAuthAuthRepository{
		findOrCreateByOAuthFunc: func(ctx context.Context, email, nickname, provider string, profileImage *string) (*entity.MorandoranUser, error) {
			return nil, assert.AnError
		},
	}

	uc := NewOAuthUseCase(mockRepo, 10*time.Second)
	res, err := uc.HandleCallback(context.Background(), "fail@test.com", "FailUser", "google", nil)

	require.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "failed to find or create user")
	t.Logf("Error: %v", err)
}
