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
)

func TestCheckEmailAuthRepository_CheckEmailExists_EmailExists(t *testing.T) {
	db := setupTestDB(t)
	repo := NewCheckEmailAuthRepository(db)
	ctx := context.Background()

	email := "exists-" + fmt.Sprintf("%d_%d", time.Now().UnixNano(), rand.Intn(100000)) + "@example.com"
	provider := "game"

	user := mysql.Users{
		Name:     "Exists User",
		Email:    email,
		Password: "hashed",
		Provider: provider,
	}
	err := db.WithContext(ctx).Create(&user).Error
	require.NoError(t, err, "Create user should succeed")

	exists, err := repo.CheckEmailExists(ctx, email, provider)
	require.NoError(t, err, "CheckEmailExists should succeed")
	assert.True(t, exists, "CheckEmailExists should return true when email+provider exists")

	t.Logf("CheckEmailExists(%s, %s) = %v (expected true)", email, provider, exists)
}

func TestCheckEmailAuthRepository_CheckEmailExists_EmailDoesNotExist(t *testing.T) {
	db := setupTestDB(t)
	repo := NewCheckEmailAuthRepository(db)
	ctx := context.Background()

	email := "nonexistent-" + fmt.Sprintf("%d_%d", time.Now().UnixNano(), rand.Intn(100000)) + "@example.com"
	provider := "game"

	exists, err := repo.CheckEmailExists(ctx, email, provider)
	require.NoError(t, err, "CheckEmailExists should succeed")
	assert.False(t, exists, "CheckEmailExists should return false when email+provider does not exist")

	t.Logf("CheckEmailExists(%s, %s) = %v (expected false)", email, provider, exists)
}

func TestCheckEmailAuthRepository_CheckEmailExists_SameEmailDifferentProvider(t *testing.T) {
	db := setupTestDB(t)
	repo := NewCheckEmailAuthRepository(db)
	ctx := context.Background()

	email := "provider-" + fmt.Sprintf("%d_%d", time.Now().UnixNano(), rand.Intn(100000)) + "@example.com"

	// Create user with provider "game"
	user := mysql.Users{
		Name:     "Provider User",
		Email:    email,
		Password: "hashed",
		Provider: "game",
	}
	err := db.WithContext(ctx).Create(&user).Error
	require.NoError(t, err, "Create user should succeed")

	// Check with same email but different provider - should return false
	exists, err := repo.CheckEmailExists(ctx, email, "google")
	require.NoError(t, err, "CheckEmailExists should succeed")
	assert.False(t, exists, "CheckEmailExists should return false for same email with different provider")

	// Check with game provider - should return true
	existsGame, err := repo.CheckEmailExists(ctx, email, "game")
	require.NoError(t, err, "CheckEmailExists should succeed")
	assert.True(t, existsGame, "CheckEmailExists should return true for email+game")

	t.Logf("Same email: game=%v, google=%v", existsGame, exists)
}
