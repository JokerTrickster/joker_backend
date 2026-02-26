package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/JokerTrickster/joker_backend/services/authService/features/auth/model/request"
	"github.com/JokerTrickster/joker_backend/services/authService/features/auth/repository"
	"github.com/JokerTrickster/joker_backend/shared/db/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogoutAuthUseCase_Success(t *testing.T) {
	db := setupTestDB(t)
	initJWTForTest(t)

	// Sign up a user
	signupRepo := repository.NewSignupAuthRepository(db)
	signupUC := NewSignupAuthUseCase(signupRepo, 10*time.Second)

	ctx := context.Background()
	email := "logout-test-" + time.Now().Format("20060102150405") + "@example.com"

	signupReq := &request.ReqSignUp{
		Email:       email,
		Password:    "password123",
		ServiceType: "game",
		Name:        "Logout Test User",
	}

	signupRes, err := signupUC.Signup(ctx, signupReq)
	require.NoError(t, err, "Signup should succeed")

	// Find user ID
	var user mysql.Users
	err = db.Where("email = ? AND provider = ?", email, "game").First(&user).Error
	require.NoError(t, err, "User should be found")
	userID := uint(user.ID)

	t.Logf("User created: id=%d, email=%s", userID, email)

	// Store a token in DB (simulating what refresh does)
	refreshRepo := repository.NewRefreshTokenAuthRepository(db)
	tokenDTO := createTokenDTO(userID, signupRes.AccessToken, signupRes.RefreshToken)
	err = refreshRepo.CreateToken(ctx, tokenDTO)
	require.NoError(t, err, "Token storage should succeed")

	// Verify token exists
	var tokenCount int64
	err = db.Model(&mysql.Tokens{}).Where("user_id = ?", userID).Count(&tokenCount).Error
	require.NoError(t, err)
	assert.Equal(t, int64(1), tokenCount, "Token should exist before logout")

	t.Logf("Token stored for user %d, count=%d", userID, tokenCount)

	// Logout
	logoutRepo := repository.NewLogoutAuthRepository(db)
	logoutUC := NewLogoutAuthUseCase(logoutRepo, 10*time.Second)

	err = logoutUC.Logout(ctx, userID)
	require.NoError(t, err, "Logout should succeed")

	// Verify token was deleted
	err = db.Model(&mysql.Tokens{}).Where("user_id = ?", userID).Count(&tokenCount).Error
	require.NoError(t, err)
	assert.Equal(t, int64(0), tokenCount, "Token should be deleted after logout")

	t.Logf("Logout succeeded: tokens for user %d = %d", userID, tokenCount)
}

func TestLogoutAuthUseCase_NoTokens(t *testing.T) {
	db := setupTestDB(t)
	initJWTForTest(t)

	logoutRepo := repository.NewLogoutAuthRepository(db)
	logoutUC := NewLogoutAuthUseCase(logoutRepo, 10*time.Second)

	ctx := context.Background()

	// Logout with a user that has no tokens - should return error
	err := logoutUC.Logout(ctx, 99999)
	assert.Error(t, err, "Logout with no tokens should return error")

	t.Logf("No-tokens logout correctly handled: %v", err)
}
