package repository

import (
	"context"
	"testing"
	"time"

	"github.com/JokerTrickster/joker_backend/shared/db/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRefreshTokenAuthRepository_CreateToken(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRefreshTokenAuthRepository(db)
	ctx := context.Background()

	// Create a user first (tokens table may have FK or we need valid user_id)
	user := mysql.Users{
		Name:     "Refresh Repo User",
		Email:    "refresh-repo-" + time.Now().Format("20060102150405") + "@example.com",
		Password: "",
		Provider: "game",
	}
	err := db.WithContext(ctx).Create(&user).Error
	require.NoError(t, err, "Create user should succeed")

	tokenDTO := &mysql.Tokens{
		UserID:           uint(user.ID),
		AccessToken:      "access-token-create",
		RefreshToken:     "refresh-token-create",
		RefreshExpiredAt: time.Now().Add(7 * 24 * time.Hour).Unix(),
	}

	err = repo.CreateToken(ctx, tokenDTO)
	require.NoError(t, err, "CreateToken should succeed")

	t.Logf("CreateToken succeeded for user ID=%d", user.ID)

	var saved mysql.Tokens
	err = db.WithContext(ctx).Where("user_id = ?", user.ID).First(&saved).Error
	require.NoError(t, err, "Token should exist in DB")
	assert.Equal(t, tokenDTO.AccessToken, saved.AccessToken, "AccessToken should match")
	assert.Equal(t, tokenDTO.RefreshToken, saved.RefreshToken, "RefreshToken should match")
	assert.Equal(t, tokenDTO.RefreshExpiredAt, saved.RefreshExpiredAt, "RefreshExpiredAt should match")

	t.Logf("Verified token in DB: UserID=%d, AccessToken=%s...", saved.UserID, saved.AccessToken[:20])
}

func TestRefreshTokenAuthRepository_FindOneByUserIDAndDeleteToken_DeletesTokens(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRefreshTokenAuthRepository(db)
	ctx := context.Background()

	user := mysql.Users{
		Name:     "Refresh Delete User",
		Email:    "refresh-del-" + time.Now().Format("20060102150405") + "@example.com",
		Password: "",
		Provider: "game",
	}
	err := db.WithContext(ctx).Create(&user).Error
	require.NoError(t, err, "Create user should succeed")

	// Create multiple tokens for same user
	token1 := mysql.Tokens{UserID: uint(user.ID), AccessToken: "a1", RefreshToken: "r1", RefreshExpiredAt: time.Now().Unix()}
	token2 := mysql.Tokens{UserID: uint(user.ID), AccessToken: "a2", RefreshToken: "r2", RefreshExpiredAt: time.Now().Unix()}
	err = db.WithContext(ctx).Create(&[]mysql.Tokens{token1, token2}).Error
	require.NoError(t, err, "Create tokens should succeed")

	err = repo.FindOneByUserIDAndDeleteToken(ctx, uint(user.ID))
	require.NoError(t, err, "FindOneByUserIDAndDeleteToken should succeed")

	var count int64
	db.WithContext(ctx).Model(&mysql.Tokens{}).Where("user_id = ?", user.ID).Count(&count)
	assert.Equal(t, int64(0), count, "All tokens for user should be deleted")

	t.Logf("FindOneByUserIDAndDeleteToken deleted all tokens for user ID=%d", user.ID)
}

func TestRefreshTokenAuthRepository_FindOneByUserIDAndDeleteToken_NoErrorWhenNoTokens(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRefreshTokenAuthRepository(db)
	ctx := context.Background()

	// Use user ID with no tokens - repo does not check RowsAffected, so returns nil
	userID := uint(888888)

	err := repo.FindOneByUserIDAndDeleteToken(ctx, userID)
	require.NoError(t, err, "FindOneByUserIDAndDeleteToken should not error when no tokens (repo does not check RowsAffected)")

	t.Logf("No tokens case: err=%v (expected nil)", err)
}
