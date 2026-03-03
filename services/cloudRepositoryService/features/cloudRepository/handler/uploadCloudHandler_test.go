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

type mockUploadCloudUseCase struct {
	mock.Mock
}

func (m *mockUploadCloudUseCase) RequestUploadURL(ctx context.Context, userID uint, req *request.UploadRequestDTO) (*response.UploadResponseDTO, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.UploadResponseDTO), args.Error(1)
}

func TestUploadCloudHandler_RequestUploadURL_Success(t *testing.T) {
	t.Log("Running: Success - POST /files/upload with valid JSON -> 200")
	e := setupTestEcho()
	mockUC := new(mockUploadCloudUseCase)
	handler := &UploadCloudRepositoryHandler{UseCase: mockUC}

	body := mustJSON(t, map[string]interface{}{
		"file_name":    "test.jpg",
		"content_type": "image/jpeg",
		"file_type":    "image",
		"file_size":    1024,
	})

	mockUC.On("RequestUploadURL", mock.Anything, testUserID, mock.AnythingOfType("*request.UploadRequestDTO")).
		Return(&response.UploadResponseDTO{
			FileID:    1,
			UploadURL: "https://presigned.url",
			S3Key:     "user/1/test-key",
			ExpiresIn: 3600,
		}, nil)

	req := httptest.NewRequest(http.MethodPost, "/files/upload", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setupAuthContext(c)

	err := handler.RequestUploadURL(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "https://presigned.url")
	t.Logf("Response: status=%d body=%s", rec.Code, rec.Body.String())
	mockUC.AssertExpectations(t)
}

func TestUploadCloudHandler_RequestUploadURL_NoUserID(t *testing.T) {
	t.Log("Running: No userID -> 401")
	e := setupTestEcho()
	mockUC := new(mockUploadCloudUseCase)
	handler := &UploadCloudRepositoryHandler{UseCase: mockUC}

	body := mustJSON(t, map[string]interface{}{
		"file_name":    "test.jpg",
		"content_type": "image/jpeg",
		"file_type":    "image",
		"file_size":    1024,
	})

	req := httptest.NewRequest(http.MethodPost, "/files/upload", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	// Do NOT set userID

	err := handler.RequestUploadURL(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "unauthorized")
	t.Logf("Response: status=%d body=%s", rec.Code, rec.Body.String())
	mockUC.AssertNotCalled(t, "RequestUploadURL")
}

func TestUploadCloudHandler_RequestUploadURL_InvalidBody(t *testing.T) {
	t.Log("Running: Invalid body -> 400")
	e := setupTestEcho()
	mockUC := new(mockUploadCloudUseCase)
	handler := &UploadCloudRepositoryHandler{UseCase: mockUC}

	req := httptest.NewRequest(http.MethodPost, "/files/upload", bytes.NewReader([]byte(`{"file_name": invalid`)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setupAuthContext(c)

	err := handler.RequestUploadURL(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "invalid request body")
	mockUC.AssertNotCalled(t, "RequestUploadURL")
}

func TestUploadCloudHandler_RequestUploadURL_ValidationError_MissingFileName(t *testing.T) {
	t.Log("Running: Validation error - missing file_name -> 400")
	e := setupTestEcho()
	mockUC := new(mockUploadCloudUseCase)
	handler := &UploadCloudRepositoryHandler{UseCase: mockUC}

	body := mustJSON(t, map[string]interface{}{
		"content_type": "image/jpeg",
		"file_type":    "image",
		"file_size":    1024,
	})

	req := httptest.NewRequest(http.MethodPost, "/files/upload", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setupAuthContext(c)

	err := handler.RequestUploadURL(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	mockUC.AssertNotCalled(t, "RequestUploadURL")
}

func TestUploadCloudHandler_RequestUploadURL_UseCaseError(t *testing.T) {
	t.Log("Running: UseCase error -> 500")
	e := setupTestEcho()
	mockUC := new(mockUploadCloudUseCase)
	handler := &UploadCloudRepositoryHandler{UseCase: mockUC}

	body := mustJSON(t, map[string]interface{}{
		"file_name":    "test.jpg",
		"content_type": "image/jpeg",
		"file_type":    "image",
		"file_size":    1024,
	})

	mockUC.On("RequestUploadURL", mock.Anything, testUserID, mock.AnythingOfType("*request.UploadRequestDTO")).
		Return(nil, assert.AnError)

	req := httptest.NewRequest(http.MethodPost, "/files/upload", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setupAuthContext(c)

	err := handler.RequestUploadURL(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockUC.AssertExpectations(t)
}
