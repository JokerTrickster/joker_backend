package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/JokerTrickster/joker_backend/services/authService/features/auth/model/request"
	"github.com/JokerTrickster/joker_backend/services/authService/features/auth/repository"
	"github.com/JokerTrickster/joker_backend/shared/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRefreshTokenUseCase_Success(t *testing.T) {
	db := setupTestDB(t)
	initJWTForTest(t)

	// Sign up a user to get initial tokens
	signupRepo := repository.NewSignupAuthRepository(db)
	signupUC := NewSignupAuthUseCase(signupRepo, 10*time.Second)

	ctx := context.Background()
	email := "refresh-test-" + time.Now().Format("20060102150405") + "@example.com"

	signupReq := &request.ReqSignUp{
		Email:       email,
		Password:    "password123",
		ServiceType: "game",
		Name:        "Refresh Test User",
	}

	signupRes, err := signupUC.Signup(ctx, signupReq)
	require.NoError(t, err, "Signup should succeed")

	t.Logf("Signup tokens: access=%s..., refresh=%s...",
		signupRes.AccessToken[:20], signupRes.RefreshToken[:20])

	// Generate a valid refresh token for the user
	_, _, refreshToken, _, err := jwt.GenerateToken(email, 1)
	require.NoError(t, err, "Token generation should succeed")

	// Use refresh token to get new tokens
	refreshRepo := repository.NewRefreshTokenAuthRepository(db)
	refreshUC := NewRefreshTokenUseCase(refreshRepo, 10*time.Second)

	refreshReq := &request.ReqRefreshToken{
		RefreshToken: refreshToken,
	}

	res, err := refreshUC.RefreshToken(ctx, refreshReq)
	require.NoError(t, err, "Refresh should succeed")
	assert.NotEmpty(t, res.AccessToken, "New AccessToken should be set")
	assert.NotEmpty(t, res.RefreshToken, "New RefreshToken should be set")
	assert.NotEqual(t, refreshToken, res.RefreshToken, "New refresh token should differ from old one")

	t.Logf("Refresh succeeded: newAccess=%s..., newRefresh=%s...",
		res.AccessToken[:20], res.RefreshToken[:20])
}

func TestRefreshTokenUseCase_InvalidToken(t *testing.T) {
	db := setupTestDB(t)
	initJWTForTest(t)

	refreshRepo := repository.NewRefreshTokenAuthRepository(db)
	refreshUC := NewRefreshTokenUseCase(refreshRepo, 10*time.Second)

	ctx := context.Background()
	refreshReq := &request.ReqRefreshToken{
		RefreshToken: "invalid.token.string",
	}

	_, err := refreshUC.RefreshToken(ctx, refreshReq)
	assert.Error(t, err, "Refresh with invalid token should fail")

	t.Logf("Invalid token correctly rejected: %v", err)
}
