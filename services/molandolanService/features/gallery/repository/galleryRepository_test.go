package repository

import (
	"context"
	"os"
	"testing"
	"time"

	authEntity "github.com/JokerTrickster/joker_backend/services/molandolanService/features/auth/model/entity"
	"github.com/JokerTrickster/joker_backend/services/molandolanService/features/gallery/model/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestGalleryRepository_List(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGalleryRepository(db)
	ctx := context.Background()

	if err := db.AutoMigrate(&entity.GalleryPost{}); err != nil {
		t.Skipf("Skipping: cannot ensure gallery_posts table: %v", err)
	}

	items, total, err := repo.List(ctx, 1, 30)
	require.NoError(t, err)
	assert.NotNil(t, items)
	assert.GreaterOrEqual(t, total, int64(0))
	t.Logf("List success: items=%d total=%d", len(items), total)
}

func TestGalleryRepository_FindByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGalleryRepository(db)
	ctx := context.Background()

	if err := db.AutoMigrate(&entity.GalleryPost{}, &authEntity.MorandoranUser{}); err != nil {
		t.Skipf("Skipping: cannot ensure tables: %v", err)
	}

	pass := "hashed"
	user := &authEntity.MorandoranUser{
		Nickname: "gallery-test-user",
		Email:    "gallery-repo-" + time.Now().Format("20060102150405") + "@example.com",
		Password: &pass,
		Role:     "user",
	}
	err := db.WithContext(ctx).Create(user).Error
	if err != nil {
		t.Skipf("Skipping: cannot create test user: %v", err)
	}
	defer db.WithContext(ctx).Unscoped().Delete(user)

	post := &entity.GalleryPost{
		AuthorID:     user.ID,
		MediaType:    "image",
		MediaURL:     "https://example.com/m.jpg",
		ThumbnailURL: "https://example.com/t.jpg",
		Caption:      "test",
	}
	err = db.WithContext(ctx).Create(post).Error
	if err != nil {
		t.Skipf("Skipping: cannot create test post: %v", err)
	}
	defer db.WithContext(ctx).Unscoped().Delete(post)

	found, err := repo.FindByID(ctx, post.ID)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, post.ID, found.ID)
	assert.Equal(t, "test", found.Caption)
	t.Logf("FindByID success: id=%d", found.ID)

	_, err = repo.FindByID(ctx, 999999)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "gallery post not found")
}

func TestGalleryRepository_Create_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGalleryRepository(db)
	ctx := context.Background()

	if err := db.AutoMigrate(&entity.GalleryPost{}, &authEntity.MorandoranUser{}); err != nil {
		t.Skipf("Skipping: cannot ensure tables: %v", err)
	}

	pass := "x"
	user := &authEntity.MorandoranUser{
		Nickname: "gallery-create-user",
		Email:    "gallery-create-" + time.Now().Format("20060102150405") + "@example.com",
		Password: &pass,
		Role:     "user",
	}
	err := db.WithContext(ctx).Create(user).Error
	if err != nil {
		t.Skipf("Skipping: cannot create test user: %v", err)
	}
	defer db.WithContext(ctx).Unscoped().Delete(user)

	post := &entity.GalleryPost{
		AuthorID:     user.ID,
		MediaType:    "image",
		MediaURL:     "https://example.com/create.jpg",
		ThumbnailURL: "https://example.com/thumb.jpg",
		Caption:      "create test",
	}
	created, err := repo.Create(ctx, post)
	require.NoError(t, err)
	require.NotNil(t, created)
	assert.NotZero(t, created.ID)
	t.Logf("Create success: id=%d", created.ID)

	err = repo.Delete(ctx, created.ID)
	require.NoError(t, err)
	t.Logf("Delete success: id=%d", created.ID)

	_, err = repo.FindByID(ctx, created.ID)
	require.Error(t, err)
}

func TestGalleryRepository_IsLiked_ToggleLike(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGalleryRepository(db)
	ctx := context.Background()

	if err := db.AutoMigrate(&entity.GalleryPost{}, &entity.GalleryLike{}, &authEntity.MorandoranUser{}); err != nil {
		t.Skipf("Skipping: cannot ensure tables: %v", err)
	}

	pass := "x"
	user := &authEntity.MorandoranUser{
		Nickname: "like-test-user",
		Email:    "like-test-" + time.Now().Format("20060102150405") + "@example.com",
		Password: &pass,
		Role:     "user",
	}
	err := db.WithContext(ctx).Create(user).Error
	if err != nil {
		t.Skipf("Skipping: cannot create test user: %v", err)
	}
	defer db.WithContext(ctx).Unscoped().Delete(user)

	post := &entity.GalleryPost{
		AuthorID:     user.ID,
		MediaType:    "image",
		MediaURL:     "x",
		ThumbnailURL: "y",
		Caption:      "",
	}
	err = db.WithContext(ctx).Create(post).Error
	if err != nil {
		t.Skipf("Skipping: cannot create test post: %v", err)
	}
	defer db.WithContext(ctx).Unscoped().Delete(post)

	isLiked, err := repo.IsLiked(ctx, user.ID, post.ID)
	require.NoError(t, err)
	assert.False(t, isLiked)

	isLiked, likeCount, err := repo.ToggleLike(ctx, user.ID, post.ID)
	require.NoError(t, err)
	assert.True(t, isLiked)
	assert.GreaterOrEqual(t, likeCount, 0)

	isLiked, err = repo.IsLiked(ctx, user.ID, post.ID)
	require.NoError(t, err)
	assert.True(t, isLiked)

	isLiked, _, err = repo.ToggleLike(ctx, user.ID, post.ID)
	require.NoError(t, err)
	assert.False(t, isLiked)
	t.Logf("ToggleLike success: like/unlike works")
}

func TestGalleryRepository_Comments(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGalleryRepository(db)
	ctx := context.Background()

	if err := db.AutoMigrate(&entity.GalleryPost{}, &entity.GalleryComment{}, &authEntity.MorandoranUser{}); err != nil {
		t.Skipf("Skipping: cannot ensure tables: %v", err)
	}

	pass := "x"
	user := &authEntity.MorandoranUser{
		Nickname: "comment-test-user",
		Email:    "comment-test-" + time.Now().Format("20060102150405") + "@example.com",
		Password: &pass,
		Role:     "user",
	}
	err := db.WithContext(ctx).Create(user).Error
	if err != nil {
		t.Skipf("Skipping: cannot create test user: %v", err)
	}
	defer db.WithContext(ctx).Unscoped().Delete(user)

	post := &entity.GalleryPost{
		AuthorID:     user.ID,
		MediaType:    "image",
		MediaURL:     "x",
		ThumbnailURL: "y",
		Caption:      "",
	}
	err = db.WithContext(ctx).Create(post).Error
	if err != nil {
		t.Skipf("Skipping: cannot create test post: %v", err)
	}
	defer db.WithContext(ctx).Unscoped().Delete(post)

	comment := &entity.GalleryComment{
		GalleryID: post.ID,
		AuthorID:  user.ID,
		Content:   "test comment",
	}
	created, err := repo.CreateComment(ctx, comment)
	require.NoError(t, err)
	require.NotNil(t, created)
	assert.NotZero(t, created.ID)

	found, err := repo.FindCommentByID(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, "test comment", found.Content)

	items, total, err := repo.ListComments(ctx, post.ID, 1, 20)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(items), 1)
	assert.GreaterOrEqual(t, total, int64(1))

	nickname, err := repo.GetAuthorNickname(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, "comment-test-user", nickname)

	err = repo.DeleteComment(ctx, created.ID)
	require.NoError(t, err)
	t.Logf("Comment CRUD success")
}

func TestGalleryRepository_GetAuthorInfo(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGalleryRepository(db)
	ctx := context.Background()

	if err := db.AutoMigrate(&authEntity.MorandoranUser{}); err != nil {
		t.Skipf("Skipping: cannot ensure morandoran_users table: %v", err)
	}

	pic := "https://example.com/profile.jpg"
	user := &authEntity.MorandoranUser{
		Nickname:     "author-info-test",
		Email:        "author-info-" + time.Now().Format("20060102150405") + "@example.com",
		Role:         "user",
		Provider:     "google",
		ProfileImage: &pic,
	}
	err := db.WithContext(ctx).Create(user).Error
	if err != nil {
		t.Skipf("Skipping: cannot create test user: %v", err)
	}
	defer db.WithContext(ctx).Unscoped().Delete(user)

	nickname, profileImage, err := repo.GetAuthorInfo(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, "author-info-test", nickname)
	require.NotNil(t, profileImage)
	assert.Equal(t, pic, *profileImage)
	t.Logf("GetAuthorInfo: nickname=%s profileImage=%s", nickname, *profileImage)

	nickname2, profileImage2, err2 := repo.GetAuthorInfo(ctx, 999999)
	require.Error(t, err2)
	assert.Empty(t, nickname2)
	assert.Nil(t, profileImage2)
	t.Logf("GetAuthorInfo not found: err=%v", err2)
}

func TestGalleryRepository_IsLikedBatch(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGalleryRepository(db)
	ctx := context.Background()

	if err := db.AutoMigrate(&entity.GalleryPost{}, &entity.GalleryLike{}, &authEntity.MorandoranUser{}); err != nil {
		t.Skipf("Skipping: cannot ensure tables: %v", err)
	}

	pass := "x"
	user := &authEntity.MorandoranUser{
		Nickname: "batch-like-user",
		Email:    "batch-like-" + time.Now().Format("20060102150405") + "@example.com",
		Password: &pass,
		Role:     "user",
	}
	err := db.WithContext(ctx).Create(user).Error
	if err != nil {
		t.Skipf("Skipping: cannot create test user: %v", err)
	}
	defer db.WithContext(ctx).Unscoped().Delete(user)

	post1 := &entity.GalleryPost{AuthorID: user.ID, MediaType: "image", MediaURL: "a", ThumbnailURL: "b"}
	post2 := &entity.GalleryPost{AuthorID: user.ID, MediaType: "image", MediaURL: "c", ThumbnailURL: "d"}
	db.WithContext(ctx).Create(post1)
	db.WithContext(ctx).Create(post2)
	defer db.WithContext(ctx).Unscoped().Delete(post1)
	defer db.WithContext(ctx).Unscoped().Delete(post2)

	like := &entity.GalleryLike{UserID: user.ID, GalleryID: post1.ID}
	db.WithContext(ctx).Create(like)
	defer db.WithContext(ctx).Unscoped().Delete(like)

	result, err := repo.IsLikedBatch(ctx, user.ID, []uint{post1.ID, post2.ID})
	require.NoError(t, err)
	assert.True(t, result[post1.ID], "post1 should be liked")
	assert.False(t, result[post2.ID], "post2 should not be liked")
	t.Logf("IsLikedBatch: post1=%v post2=%v", result[post1.ID], result[post2.ID])

	emptyResult, err := repo.IsLikedBatch(ctx, 0, []uint{post1.ID})
	require.NoError(t, err)
	assert.Empty(t, emptyResult)
	t.Logf("IsLikedBatch with userID=0: empty=%v", len(emptyResult) == 0)
}
