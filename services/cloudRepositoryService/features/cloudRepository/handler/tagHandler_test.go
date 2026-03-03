package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/entity"
	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/request"
	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/response"
	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/repository"
	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/usecase"
	"github.com/JokerTrickster/joker_backend/shared/utils"
	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const testDBPort = "3307"

func setupTestDB(t *testing.T) *gorm.DB {
	// Get DB credentials from environment or use defaults
	dbUser := os.Getenv("TEST_DB_USER")
	if dbUser == "" {
		dbUser = "root"
	}

	dbPassword := os.Getenv("TEST_DB_PASSWORD")
	if dbPassword == "" {
		dbPassword = "rootpassword"
	}

	dbPort := os.Getenv("TEST_DB_PORT")
	if dbPort == "" {
		dbPort = testDBPort
	}

	dsn := fmt.Sprintf("%s:%s@tcp(localhost:%s)/backend_dev?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPassword, dbPort)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skipf("Integration test: requires test database: %v", err)
	}

	return db
}

func createTestFile(t *testing.T, db *gorm.DB, userID uint) *entity.CloudFile {
	file := &entity.CloudFile{
		UserID:      userID,
		FileName:    fmt.Sprintf("test-file-%d.txt", time.Now().UnixNano()),
		S3Key:       fmt.Sprintf("test-key-%d", time.Now().UnixNano()),
		FileType:    entity.FileTypeImage,
		ContentType: "image/jpeg",
		FileSize:    1024,
	}

	if err := db.Create(file).Error; err != nil {
		t.Skipf("Integration test: requires matching database schema: %v", err)
	}

	return file
}

func cleanupTestFile(t *testing.T, db *gorm.DB, fileID uint) {
	// Delete file_tags associations first
	if err := db.Exec("DELETE FROM file_tags WHERE cloud_file_id = ?", fileID).Error; err != nil {
		t.Logf("Warning: Failed to delete file_tags: %v", err)
	}

	// Delete file
	if err := db.Unscoped().Delete(&entity.CloudFile{}, fileID).Error; err != nil {
		t.Logf("Warning: Failed to delete test file: %v", err)
	}
}

func cleanupTestTags(t *testing.T, db *gorm.DB, userID uint) {
	// First delete all file_tags associations for this user's tags
	if err := db.Exec("DELETE ft FROM file_tags ft INNER JOIN tags t ON ft.tag_id = t.id WHERE t.user_id = ?", userID).Error; err != nil {
		t.Logf("Warning: Failed to delete file_tags associations: %v", err)
	}

	// Then delete tags for test user
	if err := db.Where("user_id = ?", userID).Delete(&entity.Tag{}).Error; err != nil {
		t.Logf("Warning: Failed to delete test tags: %v", err)
	}
}

func createTestJWTToken(userID uint) string {
	claims := jwt.MapClaims{
		"userID": float64(userID),
		"exp":    time.Now().Add(time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte("test-secret"))
	return tokenString
}

func setupEcho() *echo.Echo {
	e := echo.New()
	e.Validator = utils.NewValidator()
	return e
}

func TestUpdateFileTags_Success(t *testing.T) {
	db := setupTestDB(t)
	testFile := createTestFile(t, db, testUserID)
	defer cleanupTestFile(t, db, testFile.ID)
	defer cleanupTestTags(t, db, testUserID)

	// Setup repositories and use cases
	tagRepo := repository.NewTagRepository(db)
	statsRepo := repository.NewUserStatsCloudRepositoryRepository(db)
	tagUC := usecase.NewTagUseCase(tagRepo, statsRepo, 30*time.Second)

	// Setup echo and handler
	e := setupEcho()
	group := e.Group("")
	NewTagHandler(group, tagUC)

	// Create request
	reqBody := request.UpdateFileTagsRequestDTO{
		Tags: []string{"work", "important", "project"},
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/files/%d/tags", testFile.ID), bytes.NewReader(bodyBytes))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+createTestJWTToken(testUserID))

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/files/:id/tags")
	c.SetParamNames("id")
	c.SetParamValues(fmt.Sprintf("%d", testFile.ID))

	// Set user ID in context (simulating JWT middleware)
	c.Set("userID", testUserID)

	// Execute
	handler := &TagHandler{UseCase: tagUC}
	err := handler.UpdateFileTags(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp response.UpdateFileTagsResponseDTO
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, testFile.ID, resp.FileID)
	assert.ElementsMatch(t, reqBody.Tags, resp.Tags)

	t.Logf("Update tags response: %+v", resp)
}

func TestUpdateFileTags_ReplaceExisting(t *testing.T) {
	db := setupTestDB(t)
	testFile := createTestFile(t, db, testUserID)
	defer cleanupTestFile(t, db, testFile.ID)
	defer cleanupTestTags(t, db, testUserID)

	tagRepo := repository.NewTagRepository(db)
	statsRepo := repository.NewUserStatsCloudRepositoryRepository(db)
	tagUC := usecase.NewTagUseCase(tagRepo, statsRepo, 30*time.Second)

	ctx := context.Background()

	// Add initial tags
	initialTags := []string{"tag1", "tag2"}
	var tags []entity.Tag
	for _, tagName := range initialTags {
		tag, _ := tagRepo.FindOrCreateTag(ctx, testUserID, tagName)
		tags = append(tags, *tag)
	}
	tagRepo.UpdateFileTags(ctx, testFile.ID, testUserID, tags)

	// Setup echo and handler
	e := setupEcho()
	group := e.Group("")
	NewTagHandler(group, tagUC)

	// Create request to replace tags
	reqBody := request.UpdateFileTagsRequestDTO{
		Tags: []string{"new-tag1", "new-tag2"},
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/files/%d/tags", testFile.ID), bytes.NewReader(bodyBytes))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+createTestJWTToken(testUserID))

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/files/:id/tags")
	c.SetParamNames("id")
	c.SetParamValues(fmt.Sprintf("%d", testFile.ID))
	c.Set("userID", testUserID)

	// Execute
	handler := &TagHandler{UseCase: tagUC}
	err := handler.UpdateFileTags(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp response.UpdateFileTagsResponseDTO
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.ElementsMatch(t, reqBody.Tags, resp.Tags)
	assert.NotContains(t, resp.Tags, "tag1") // Old tags should be removed

	t.Logf("Replace tags response: %+v", resp)
}

func TestAddTagToFile_Success(t *testing.T) {
	db := setupTestDB(t)
	testFile := createTestFile(t, db, testUserID)
	defer cleanupTestFile(t, db, testFile.ID)
	defer cleanupTestTags(t, db, testUserID)

	tagRepo := repository.NewTagRepository(db)
	statsRepo := repository.NewUserStatsCloudRepositoryRepository(db)
	tagUC := usecase.NewTagUseCase(tagRepo, statsRepo, 30*time.Second)

	e := setupEcho()
	group := e.Group("")
	NewTagHandler(group, tagUC)

	// Create request
	reqBody := request.AddFileTagRequestDTO{
		Tag: "urgent",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/files/%d/tags", testFile.ID), bytes.NewReader(bodyBytes))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+createTestJWTToken(testUserID))

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/files/:id/tags")
	c.SetParamNames("id")
	c.SetParamValues(fmt.Sprintf("%d", testFile.ID))
	c.Set("userID", testUserID)

	// Execute
	handler := &TagHandler{UseCase: tagUC}
	err := handler.AddTagToFile(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp response.AddFileTagResponseDTO
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, testFile.ID, resp.FileID)
	assert.Equal(t, "urgent", resp.Tag)

	t.Logf("Add tag response: %+v", resp)
}

func TestAddTagToFile_DuplicateTag(t *testing.T) {
	db := setupTestDB(t)
	testFile := createTestFile(t, db, testUserID)
	defer cleanupTestFile(t, db, testFile.ID)
	defer cleanupTestTags(t, db, testUserID)

	tagRepo := repository.NewTagRepository(db)
	statsRepo := repository.NewUserStatsCloudRepositoryRepository(db)
	tagUC := usecase.NewTagUseCase(tagRepo, statsRepo, 30*time.Second)

	ctx := context.Background()

	// Add initial tag
	tag, _ := tagRepo.FindOrCreateTag(ctx, testUserID, "existing")
	tagRepo.AddTagToFile(ctx, testFile.ID, testUserID, *tag)

	e := setupEcho()
	group := e.Group("")
	NewTagHandler(group, tagUC)

	// Try to add the same tag again
	reqBody := request.AddFileTagRequestDTO{
		Tag: "existing",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/files/%d/tags", testFile.ID), bytes.NewReader(bodyBytes))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+createTestJWTToken(testUserID))

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/files/:id/tags")
	c.SetParamNames("id")
	c.SetParamValues(fmt.Sprintf("%d", testFile.ID))
	c.Set("userID", testUserID)

	// Execute
	handler := &TagHandler{UseCase: tagUC}
	err := handler.AddTagToFile(c)

	// Assert - should succeed silently
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	t.Logf("Duplicate tag add succeeded silently")
}

func TestRemoveTagFromFile_Success(t *testing.T) {
	db := setupTestDB(t)
	testFile := createTestFile(t, db, testUserID)
	defer cleanupTestFile(t, db, testFile.ID)
	defer cleanupTestTags(t, db, testUserID)

	tagRepo := repository.NewTagRepository(db)
	statsRepo := repository.NewUserStatsCloudRepositoryRepository(db)
	tagUC := usecase.NewTagUseCase(tagRepo, statsRepo, 30*time.Second)

	ctx := context.Background()

	// Add a tag first
	tag, _ := tagRepo.FindOrCreateTag(ctx, testUserID, "to-remove")
	tagRepo.AddTagToFile(ctx, testFile.ID, testUserID, *tag)

	e := setupEcho()
	group := e.Group("")
	NewTagHandler(group, tagUC)

	// Create request
	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/files/%d/tags/to-remove", testFile.ID), nil)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+createTestJWTToken(testUserID))

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/files/:id/tags/:tag_name")
	c.SetParamNames("id", "tag_name")
	c.SetParamValues(fmt.Sprintf("%d", testFile.ID), "to-remove")
	c.Set("userID", testUserID)

	// Execute
	handler := &TagHandler{UseCase: tagUC}
	err := handler.RemoveTagFromFile(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, rec.Code)

	// Verify tag was removed
	file, _ := tagRepo.GetFileByID(ctx, testFile.ID, testUserID)
	assert.Empty(t, file.Tags)

	t.Logf("Tag removed successfully")
}

func TestRemoveTagFromFile_NotFound(t *testing.T) {
	db := setupTestDB(t)
	testFile := createTestFile(t, db, testUserID)
	defer cleanupTestFile(t, db, testFile.ID)
	defer cleanupTestTags(t, db, testUserID)

	tagRepo := repository.NewTagRepository(db)
	statsRepo := repository.NewUserStatsCloudRepositoryRepository(db)
	tagUC := usecase.NewTagUseCase(tagRepo, statsRepo, 30*time.Second)

	e := setupEcho()
	group := e.Group("")
	NewTagHandler(group, tagUC)

	// Try to remove a tag that doesn't exist
	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/files/%d/tags/nonexistent", testFile.ID), nil)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+createTestJWTToken(testUserID))

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/files/:id/tags/:tag_name")
	c.SetParamNames("id", "tag_name")
	c.SetParamValues(fmt.Sprintf("%d", testFile.ID), "nonexistent")
	c.Set("userID", testUserID)

	// Execute
	handler := &TagHandler{UseCase: tagUC}
	err := handler.RemoveTagFromFile(c)

	// Assert - should fail
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	t.Logf("Remove nonexistent tag returned error as expected")
}

func TestTagOperations_Unauthorized(t *testing.T) {
	db := setupTestDB(t)
	testFile := createTestFile(t, db, testUserID)
	defer cleanupTestFile(t, db, testFile.ID)

	tagRepo := repository.NewTagRepository(db)
	statsRepo := repository.NewUserStatsCloudRepositoryRepository(db)
	tagUC := usecase.NewTagUseCase(tagRepo, statsRepo, 30*time.Second)

	e := setupEcho()
	group := e.Group("")
	handler := NewTagHandler(group, tagUC).(*TagHandler)

	tests := []struct {
		name   string
		method string
		path   string
		body   interface{}
	}{
		{
			name:   "Update tags without auth",
			method: http.MethodPut,
			path:   fmt.Sprintf("/files/%d/tags", testFile.ID),
			body:   request.UpdateFileTagsRequestDTO{Tags: []string{"tag1"}},
		},
		{
			name:   "Add tag without auth",
			method: http.MethodPost,
			path:   fmt.Sprintf("/files/%d/tags", testFile.ID),
			body:   request.AddFileTagRequestDTO{Tag: "tag1"},
		},
		{
			name:   "Remove tag without auth",
			method: http.MethodDelete,
			path:   fmt.Sprintf("/files/%d/tags/tag1", testFile.ID),
			body:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			if tt.body != nil {
				bodyBytes, _ := json.Marshal(tt.body)
				req = httptest.NewRequest(tt.method, tt.path, bytes.NewReader(bodyBytes))
			} else {
				req = httptest.NewRequest(tt.method, tt.path, nil)
			}
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetPath("/files/:id/tags")
			c.SetParamNames("id")
			c.SetParamValues(fmt.Sprintf("%d", testFile.ID))
			// Don't set userID to simulate unauthorized request

			// Execute appropriate handler
			var err error
			switch tt.method {
			case http.MethodPut:
				err = handler.UpdateFileTags(c)
			case http.MethodPost:
				err = handler.AddTagToFile(c)
			case http.MethodDelete:
				c.SetPath("/files/:id/tags/:tag_name")
				c.SetParamNames("id", "tag_name")
				c.SetParamValues(fmt.Sprintf("%d", testFile.ID), "tag1")
				err = handler.RemoveTagFromFile(c)
			}

			// Assert
			assert.NoError(t, err)
			assert.Equal(t, http.StatusUnauthorized, rec.Code)

			t.Logf("Test '%s' correctly returned 401 Unauthorized", tt.name)
		})
	}
}

func TestUpdateFileTags_InvalidFileID(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestTags(t, db, testUserID)

	tagRepo := repository.NewTagRepository(db)
	statsRepo := repository.NewUserStatsCloudRepositoryRepository(db)
	tagUC := usecase.NewTagUseCase(tagRepo, statsRepo, 30*time.Second)

	e := setupEcho()
	group := e.Group("")
	NewTagHandler(group, tagUC)

	reqBody := request.UpdateFileTagsRequestDTO{
		Tags: []string{"tag1"},
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, "/files/99999/tags", bytes.NewReader(bodyBytes))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+createTestJWTToken(testUserID))

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/files/:id/tags")
	c.SetParamNames("id")
	c.SetParamValues("99999")
	c.Set("userID", testUserID)

	handler := &TagHandler{UseCase: tagUC}
	err := handler.UpdateFileTags(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)

	t.Logf("Invalid file ID correctly returned 404")
}
