package handler

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/request"
	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/response"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockMultipartUploadUseCase struct {
	mock.Mock
}

func (m *mockMultipartUploadUseCase) InitiateMultipartUpload(ctx context.Context, userID uint, req *request.InitiateMultipartUploadRequestDTO) (*response.InitiateMultipartUploadResponseDTO, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.InitiateMultipartUploadResponseDTO), args.Error(1)
}

func (m *mockMultipartUploadUseCase) GeneratePresignedURLs(ctx context.Context, userID uint, req *request.GeneratePresignedURLsRequestDTO) (*response.GeneratePresignedURLsResponseDTO, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.GeneratePresignedURLsResponseDTO), args.Error(1)
}

func (m *mockMultipartUploadUseCase) CompleteMultipartUpload(ctx context.Context, userID uint, req *request.CompleteMultipartUploadRequestDTO) (*response.CompleteMultipartUploadResponseDTO, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.CompleteMultipartUploadResponseDTO), args.Error(1)
}

func (m *mockMultipartUploadUseCase) AbortMultipartUpload(ctx context.Context, userID uint, req *request.AbortMultipartUploadRequestDTO) (*response.AbortMultipartUploadResponseDTO, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.AbortMultipartUploadResponseDTO), args.Error(1)
}

func TestMultipartUploadHandler_InitiateMultipartUpload_Success(t *testing.T) {
	t.Log("Running: InitiateMultipartUpload success -> 200")
	e := setupTestEcho()
	mockUC := new(mockMultipartUploadUseCase)
	handler := &MultipartUploadHandler{UseCase: mockUC}

	body := mustJSON(t, map[string]interface{}{
		"file_name":    "large-video.mp4",
		"file_size":    5242880,
		"content_type": "video/mp4",
		"file_type":    "video",
	})

	mockUC.On("InitiateMultipartUpload", mock.Anything, testUserID, mock.AnythingOfType("*request.InitiateMultipartUploadRequestDTO")).
		Return(&response.InitiateMultipartUploadResponseDTO{
			UploadID:   "upload-123",
			FileKey:    "user/1/key",
			PartSize:   5242880,
			TotalParts: 10,
		}, nil)

	req := httptest.NewRequest(http.MethodPost, "/files/multipart/initiate", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setupAuthContext(c)

	err := handler.InitiateMultipartUpload(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "upload-123")
	mockUC.AssertExpectations(t)
}

func TestMultipartUploadHandler_GeneratePresignedURLs_Success(t *testing.T) {
	t.Log("Running: GeneratePresignedURLs success -> 200")
	e := setupTestEcho()
	mockUC := new(mockMultipartUploadUseCase)
	handler := &MultipartUploadHandler{UseCase: mockUC}

	body := mustJSON(t, map[string]interface{}{
		"upload_id":     "upload-123",
		"file_key":      "user/1/key",
		"part_numbers": []int{1, 2, 3},
	})

	mockUC.On("GeneratePresignedURLs", mock.Anything, testUserID, mock.AnythingOfType("*request.GeneratePresignedURLsRequestDTO")).
		Return(&response.GeneratePresignedURLsResponseDTO{
			URLs:      []response.PresignedURLPart{{PartNumber: 1, URL: "https://url1"}},
			ExpiresIn: 3600,
		}, nil)

	req := httptest.NewRequest(http.MethodPost, "/files/multipart/presigned-urls", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setupAuthContext(c)

	err := handler.GeneratePresignedURLs(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockUC.AssertExpectations(t)
}

func TestMultipartUploadHandler_CompleteMultipartUpload_Success(t *testing.T) {
	t.Log("Running: CompleteMultipartUpload success -> 200")
	e := setupTestEcho()
	mockUC := new(mockMultipartUploadUseCase)
	handler := &MultipartUploadHandler{UseCase: mockUC}

	body := mustJSON(t, map[string]interface{}{
		"upload_id": "upload-123",
		"file_key":  "user/1/key",
		"parts": []map[string]interface{}{
			{"part_number": 1, "etag": "etag1"},
		},
	})

	mockUC.On("CompleteMultipartUpload", mock.Anything, testUserID, mock.AnythingOfType("*request.CompleteMultipartUploadRequestDTO")).
		Return(&response.CompleteMultipartUploadResponseDTO{FileID: 1, Size: 5242880}, nil)

	req := httptest.NewRequest(http.MethodPost, "/files/multipart/complete", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setupAuthContext(c)

	err := handler.CompleteMultipartUpload(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockUC.AssertExpectations(t)
}

func TestMultipartUploadHandler_AbortMultipartUpload_Success(t *testing.T) {
	t.Log("Running: AbortMultipartUpload success -> 200")
	e := setupTestEcho()
	mockUC := new(mockMultipartUploadUseCase)
	handler := &MultipartUploadHandler{UseCase: mockUC}

	body := mustJSON(t, map[string]interface{}{
		"upload_id": "upload-123",
		"file_key":  "user/1/key",
	})

	mockUC.On("AbortMultipartUpload", mock.Anything, testUserID, mock.AnythingOfType("*request.AbortMultipartUploadRequestDTO")).
		Return(&response.AbortMultipartUploadResponseDTO{Success: true, Message: "aborted"}, nil)

	req := httptest.NewRequest(http.MethodPost, "/files/multipart/abort", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setupAuthContext(c)

	err := handler.AbortMultipartUpload(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockUC.AssertExpectations(t)
}

func TestMultipartUploadHandler_ValidationError_MissingFileName(t *testing.T) {
	t.Log("Running: Validation error - missing file_name -> 400")
	e := setupTestEcho()
	mockUC := new(mockMultipartUploadUseCase)
	handler := &MultipartUploadHandler{UseCase: mockUC}

	body := mustJSON(t, map[string]interface{}{
		"file_size":    5242880,
		"content_type": "video/mp4",
		"file_type":    "video",
	})

	req := httptest.NewRequest(http.MethodPost, "/files/multipart/initiate", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setupAuthContext(c)

	err := handler.InitiateMultipartUpload(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	mockUC.AssertNotCalled(t, "InitiateMultipartUpload")
}

func TestMultipartUploadHandler_NoUserID(t *testing.T) {
	t.Log("Running: No userID -> 401")
	e := setupTestEcho()
	mockUC := new(mockMultipartUploadUseCase)
	handler := &MultipartUploadHandler{UseCase: mockUC}

	body := mustJSON(t, map[string]interface{}{
		"file_name":    "test.mp4",
		"file_size":    5242880,
		"content_type": "video/mp4",
		"file_type":    "video",
	})

	req := httptest.NewRequest(http.MethodPost, "/files/multipart/initiate", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.InitiateMultipartUpload(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	mockUC.AssertNotCalled(t, "InitiateMultipartUpload")
}
