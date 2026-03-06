package repository

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/JokerTrickster/joker_backend/shared/db/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func requireTokensTable(t *testing.T, db *gorm.DB) {
	t.Helper()
	if !db.Migrator().HasTable(&mysql.Tokens{}) {
		t.Skipf("Integration test: requires 'tokens' table in test database")
	}
}

func TestLogoutAuthRepository_DeleteTokenByUserID_Success(t *testing.T) {
	db := setupTestDB(t)
	requireTokensTable(t, db)
	repo := NewLogoutAuthRepository(db)
	ctx := context.Background()

	// Create a user and token
	user := mysql.Users{
		Name:     "Logout Test User",
		Email:    "logout-" + fmt.Sprintf("%d_%d", time.Now().UnixNano(), rand.Intn(100000)) + "@example.com",
		Password: "",
		Provider: "game",
	}
	err := db.WithContext(ctx).Create(&user).Error
	require.NoError(t, err, "Create user should succeed")

	token := mysql.Tokens{
		UserID:           uint(user.ID),
		AccessToken:      "access-token-logout",
		RefreshToken:     "refresh-token-logout",
		RefreshExpiredAt: time.Now().Add(7 * 24 * time.Hour).Unix(),
	}
	err = db.WithContext(ctx).Create(&token).Error
	require.NoError(t, err, "Create token should succeed")

	t.Logf("Created token for user ID=%d", user.ID)

	err = repo.DeleteTokenByUserID(ctx, uint(user.ID))
	require.NoError(t, err, "DeleteTokenByUserID should succeed")

	var count int64
	db.WithContext(ctx).Model(&mysql.Tokens{}).Where("user_id = ?", user.ID).Count(&count)
	assert.Equal(t, int64(0), count, "Token should be deleted from DB")

	t.Logf("Token successfully deleted for user ID=%d", user.ID)
}

func TestLogoutAuthRepository_DeleteTokenByUserID_NoTokensFound(t *testing.T) {
	db := setupTestDB(t)
	requireTokensTable(t, db)
	repo := NewLogoutAuthRepository(db)
	ctx := context.Background()

	// Use a user ID that has no tokens (we don't create any)
	userID := uint(999999)

	err := repo.DeleteTokenByUserID(ctx, userID)
	require.Error(t, err, "DeleteTokenByUserID should fail when no tokens exist")
	assert.Contains(t, err.Error(), "no tokens found", "Error should mention no tokens found")

	t.Logf("No tokens case correctly returns error: %v", err)
}
