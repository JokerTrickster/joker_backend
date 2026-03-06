package repository

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/JokerTrickster/joker_backend/shared/db/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	mysqlDriver "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		dsn = "root:rootpassword@tcp(localhost:3307)/test_db?charset=utf8mb4&parseTime=True&loc=Local"
	}
	db, err := gorm.Open(mysqlDriver.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skipf("Integration test: requires test database: %v", err)
	}
	return db
}

func TestSigninAuthRepository_FindUserByEmail_Success(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSigninAuthRepository(db)
	ctx := context.Background()

	email := "signin-repo-" + fmt.Sprintf("%d_%d", time.Now().UnixNano(), rand.Intn(100000)) + "@example.com"
	password := "correctpassword123"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err, "bcrypt.GenerateFromPassword should succeed")

	user := mysql.Users{
		Name:     "Signin Repo User",
		Email:    email,
		Password: string(hashedPassword),
		Provider: "game",
	}
	err = db.WithContext(ctx).Create(&user).Error
	require.NoError(t, err, "Create user should succeed")

	t.Logf("Created test user: ID=%d, email=%s", user.ID, email)

	userID, returnedEmail, err := repo.FindUserByEmail(ctx, email, password, "game")
	require.NoError(t, err, "FindUserByEmail should succeed")
	assert.Equal(t, uint(user.ID), userID, "UserID should match")
	assert.Equal(t, email, returnedEmail, "Email should match")

	t.Logf("FindUserByEmail succeeded: userID=%d, email=%s", userID, returnedEmail)
}

func TestSigninAuthRepository_FindUserByEmail_UserNotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSigninAuthRepository(db)
	ctx := context.Background()

	email := "nonexistent-" + fmt.Sprintf("%d_%d", time.Now().UnixNano(), rand.Intn(100000)) + "@example.com"

	_, _, err := repo.FindUserByEmail(ctx, email, "anypassword", "game")
	require.Error(t, err, "FindUserByEmail should fail for nonexistent user")
	assert.Contains(t, err.Error(), "user not found", "Error should mention user not found")

	t.Logf("User not found correctly: %v", err)
}

func TestSigninAuthRepository_FindUserByEmail_WrongPassword(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSigninAuthRepository(db)
	ctx := context.Background()

	email := "signin-wrongpw-" + fmt.Sprintf("%d_%d", time.Now().UnixNano(), rand.Intn(100000)) + "@example.com"
	password := "correctpassword"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err, "bcrypt.GenerateFromPassword should succeed")

	user := mysql.Users{
		Name:     "Wrong PW User",
		Email:    email,
		Password: string(hashedPassword),
		Provider: "game",
	}
	err = db.WithContext(ctx).Create(&user).Error
	require.NoError(t, err, "Create user should succeed")

	_, _, err = repo.FindUserByEmail(ctx, email, "wrongpassword", "game")
	require.Error(t, err, "FindUserByEmail should fail for wrong password")
	assert.Contains(t, err.Error(), "password not match", "Error should mention password mismatch")

	t.Logf("Wrong password correctly rejected: %v", err)
}

func TestSigninAuthRepository_FindUserByEmail_WrongProvider(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSigninAuthRepository(db)
	ctx := context.Background()

	email := "signin-provider-" + fmt.Sprintf("%d_%d", time.Now().UnixNano(), rand.Intn(100000)) + "@example.com"
	password := "password123"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err, "bcrypt.GenerateFromPassword should succeed")

	user := mysql.Users{
		Name:     "Provider Test User",
		Email:    email,
		Password: string(hashedPassword),
		Provider: "game",
	}
	err = db.WithContext(ctx).Create(&user).Error
	require.NoError(t, err, "Create user should succeed")

	// Query with wrong provider - user exists with provider "game", we query "google"
	_, _, err = repo.FindUserByEmail(ctx, email, password, "google")
	require.Error(t, err, "FindUserByEmail should fail for wrong provider")
	assert.Contains(t, err.Error(), "user not found", "Error should mention user not found")

	t.Logf("Wrong provider correctly rejected: %v", err)
}
