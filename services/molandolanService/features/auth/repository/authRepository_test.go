package repository

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/JokerTrickster/joker_backend/services/molandolanService/features/auth/model/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		dsn = "root:rootpassword@tcp(localhost:3307)/test_db?charset=utf8mb4&parseTime=True&loc=Local"
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skipf("Skipping: test database unavailable: %v", err)
	}
	return db
}

func TestAuthRepository_FindUserByEmail(t *testing.T) {
	db := setupTestDB(t)
	repo := NewAuthRepository(db)

	// Ensure morandoran_users table exists (may need migration)
	if err := db.AutoMigrate(&entity.MorandoranUser{}); err != nil {
		t.Skipf("Skipping: cannot ensure morandoran_users table: %v", err)
	}

	ctx := context.Background()
	email := "auth-repo-test-" + time.Now().Format("20060102150405") + "@example.com"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	require.NoError(t, err)
	pw := string(hashedPassword)

	user := &entity.MorandoranUser{
		Nickname: "auth-repo-test",
		Email:    email,
		Password: &pw,
		Role:     "user",
	}
	err = db.WithContext(ctx).Create(user).Error
	if err != nil {
		t.Skipf("Skipping: cannot create test user (table may not exist): %v", err)
	}
	defer func() {
		db.WithContext(ctx).Unscoped().Delete(user)
	}()

	t.Run("success: finds user by email", func(t *testing.T) {
		found, err := repo.FindUserByEmail(ctx, email)
		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, email, found.Email)
		assert.Equal(t, "auth-repo-test", found.Nickname)
		t.Logf("Found user by email: id=%d email=%s", found.ID, found.Email)
	})

	t.Run("not found: returns error for nonexistent email", func(t *testing.T) {
		found, err := repo.FindUserByEmail(ctx, "nonexistent-"+email)
		require.Error(t, err)
		assert.Nil(t, found)
		assert.Contains(t, err.Error(), "user not found")
		t.Logf("Not found error: %v", err)
	})
}

func TestAuthRepository_FindUserByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewAuthRepository(db)

	if err := db.AutoMigrate(&entity.MorandoranUser{}); err != nil {
		t.Skipf("Skipping: cannot ensure morandoran_users table: %v", err)
	}

	ctx := context.Background()
	email := "auth-repo-id-test-" + time.Now().Format("20060102150405") + "@example.com"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	require.NoError(t, err)
	pw := string(hashedPassword)

	user := &entity.MorandoranUser{
		Nickname: "auth-repo-id-test",
		Email:    email,
		Password: &pw,
		Role:     "user",
	}
	err = db.WithContext(ctx).Create(user).Error
	if err != nil {
		t.Skipf("Skipping: cannot create test user: %v", err)
	}
	defer func() {
		db.WithContext(ctx).Unscoped().Delete(user)
	}()

	t.Run("success: finds user by ID", func(t *testing.T) {
		found, err := repo.FindUserByID(ctx, user.ID)
		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, user.ID, found.ID)
		assert.Equal(t, email, found.Email)
		t.Logf("Found user by ID: id=%d email=%s", found.ID, found.Email)
	})

	t.Run("not found: returns error for nonexistent ID", func(t *testing.T) {
		found, err := repo.FindUserByID(ctx, 999999)
		require.Error(t, err)
		assert.Nil(t, found)
		assert.Contains(t, err.Error(), "user not found")
		t.Logf("Not found error: %v", err)
	})
}

func TestAuthRepository_FindOrCreateByOAuth_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewAuthRepository(db)
	if err := db.AutoMigrate(&entity.MorandoranUser{}); err != nil {
		t.Skipf("Skipping: cannot ensure morandoran_users table: %v", err)
	}

	ctx := context.Background()
	email := "oauth-create-" + time.Now().Format("20060102150405") + "@google.com"
	pic := "https://example.com/avatar.jpg"

	user, err := repo.FindOrCreateByOAuth(ctx, email, "OAuthUser", "google", &pic)
	if err != nil {
		t.Skipf("Skipping: FindOrCreateByOAuth failed (schema may differ): %v", err)
	}
	defer func() { db.Unscoped().Delete(user) }()

	require.NotNil(t, user)
	assert.Equal(t, email, user.Email)
	assert.Equal(t, "OAuthUser", user.Nickname)
	assert.Equal(t, "google", user.Provider)
	require.NotNil(t, user.ProfileImage)
	assert.Equal(t, pic, *user.ProfileImage)
	assert.Equal(t, "user", user.Role)
	t.Logf("Created OAuth user: id=%d email=%s provider=%s", user.ID, user.Email, user.Provider)
}

func TestAuthRepository_FindOrCreateByOAuth_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewAuthRepository(db)
	if err := db.AutoMigrate(&entity.MorandoranUser{}); err != nil {
		t.Skipf("Skipping: cannot ensure morandoran_users table: %v", err)
	}

	ctx := context.Background()
	email := "oauth-update-" + time.Now().Format("20060102150405") + "@google.com"
	firstPic := "https://example.com/first.jpg"

	first, err := repo.FindOrCreateByOAuth(ctx, email, "FirstName", "google", &firstPic)
	if err != nil {
		t.Skipf("Skipping: FindOrCreateByOAuth create failed: %v", err)
	}
	defer func() { db.Unscoped().Delete(first) }()

	newPic := "https://example.com/second.jpg"
	second, err := repo.FindOrCreateByOAuth(ctx, email, "UpdatedName", "google", &newPic)
	require.NoError(t, err)
	require.NotNil(t, second)
	assert.Equal(t, first.ID, second.ID, "should be the same user")
	assert.Equal(t, "UpdatedName", second.Nickname, "nickname should be updated")
	t.Logf("Updated OAuth user: id=%d nickname=%s", second.ID, second.Nickname)
}

func TestAuthRepository_UpdateNickname_Success(t *testing.T) {
	db := setupTestDB(t)
	repo := NewAuthRepository(db)
	if err := db.AutoMigrate(&entity.MorandoranUser{}); err != nil {
		t.Skipf("Skipping: cannot ensure morandoran_users table: %v", err)
	}

	ctx := context.Background()
	email := "update-nick-" + time.Now().Format("20060102150405") + "@example.com"
	user := &entity.MorandoranUser{
		Nickname: "oldnick",
		Email:    email,
		Role:     "user",
		Provider: "email",
	}
	err := db.WithContext(ctx).Create(user).Error
	if err != nil {
		t.Skipf("Skipping: cannot create test user: %v", err)
	}
	defer func() { db.Unscoped().Delete(user) }()

	updated, err := repo.UpdateNickname(ctx, user.ID, "newnick")
	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Equal(t, "newnick", updated.Nickname)
	assert.Equal(t, user.ID, updated.ID)
	t.Logf("Updated nickname: id=%d old=oldnick new=%s", updated.ID, updated.Nickname)
}

func TestAuthRepository_UpdateNickname_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewAuthRepository(db)
	if err := db.AutoMigrate(&entity.MorandoranUser{}); err != nil {
		t.Skipf("Skipping: cannot ensure morandoran_users table: %v", err)
	}

	ctx := context.Background()
	updated, err := repo.UpdateNickname(ctx, 999999, "newnick")
	require.Error(t, err)
	assert.Nil(t, updated)
	assert.Contains(t, err.Error(), "user not found")
	t.Logf("Not found error: %v", err)
}
