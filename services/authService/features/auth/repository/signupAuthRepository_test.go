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
	"golang.org/x/crypto/bcrypt"
)

func TestSignupAuthRepository_CreateUser_Success(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSignupAuthRepository(db)
	ctx := context.Background()

	email := "signup-repo-" + fmt.Sprintf("%d_%d", time.Now().UnixNano(), rand.Intn(100000)) + "@example.com"
	name := "Signup Repo User"
	password := "securepassword456"
	provider := "game"

	userID, err := repo.CreateUser(ctx, name, email, password, provider)
	require.NoError(t, err, "CreateUser should succeed")
	require.NotZero(t, userID, "UserID should be non-zero")

	t.Logf("CreateUser succeeded: userID=%d, email=%s", userID, email)

	var user mysql.Users
	err = db.WithContext(ctx).Where("id = ?", userID).First(&user).Error
	require.NoError(t, err, "User should exist in DB")
	assert.Equal(t, name, user.Name, "Name should match")
	assert.Equal(t, email, user.Email, "Email should match")
	assert.Equal(t, provider, user.Provider, "Provider should match")
	assert.NotEqual(t, password, user.Password, "Password should NOT be stored in plaintext")
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	assert.NoError(t, err, "Stored password should be valid bcrypt hash")

	t.Logf("Verified user in DB: ID=%d, Password hash stored (not plaintext)", user.ID)
}

func TestSignupAuthRepository_CreateUser_DuplicateEmailProvider(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSignupAuthRepository(db)
	ctx := context.Background()

	email := "dup-repo-" + fmt.Sprintf("%d_%d", time.Now().UnixNano(), rand.Intn(100000)) + "@example.com"

	_, err := repo.CreateUser(ctx, "First User", email, "password1", "game")
	require.NoError(t, err, "First CreateUser should succeed")

	_, err = repo.CreateUser(ctx, "Second User", email, "password2", "game")
	require.Error(t, err, "Duplicate CreateUser should fail")
	assert.Contains(t, err.Error(), "email already exists", "Error should mention email already exists")

	t.Logf("Duplicate email+provider correctly rejected: %v", err)
}

func TestSignupAuthRepository_CreateUser_PasswordStoredAsBcrypt(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSignupAuthRepository(db)
	ctx := context.Background()

	email := "bcrypt-repo-" + fmt.Sprintf("%d_%d", time.Now().UnixNano(), rand.Intn(100000)) + "@example.com"
	password := "mypassword789"

	userID, err := repo.CreateUser(ctx, "Bcrypt Test", email, password, "game")
	require.NoError(t, err, "CreateUser should succeed")

	var user mysql.Users
	err = db.WithContext(ctx).Where("id = ?", userID).First(&user).Error
	require.NoError(t, err, "User should exist")

	assert.True(t, len(user.Password) > 30, "Bcrypt hash should be at least 30 chars")
	assert.True(t, user.Password[:4] == "$2a$" || user.Password[:4] == "$2b$",
		"Password should look like bcrypt hash (starts with $2a$ or $2b$)")
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	require.NoError(t, err, "Stored hash should verify against original password")

	t.Logf("Verified bcrypt: stored=%s...", user.Password[:20])
}
