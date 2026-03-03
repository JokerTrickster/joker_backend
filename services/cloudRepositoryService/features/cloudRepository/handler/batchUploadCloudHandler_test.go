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

type mockBatchUploadCloudUseCase struct {
	mock.Mock
}

func (m *mockBatchUploadCloudUseCase) RequestBatchUploadURL(ctx context.Context, userID uint, req *request.BatchUploadRequestDTO) (*response.BatchUploadResponseDTO, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.BatchUploadResponseDTO), args.Error(1)
}

func TestBatchUploadCloudHandler_RequestBatchUploadURL_Success(t *testing.T) {
	t.Log("Running: Success - POST /files/upload/batch -> 200")
	e := setupTestEcho()
	mockUC := new(mockBatchUploadCloudUseCase)
	handler := &BatchUploadCloudRepositoryHandler{UseCase: mockUC}

	body := mustJSON(t, map[string]interface{}{
		"files": []map[string]interface{}{
			{
				"file_name":    "test1.jpg",
				"content_type": "image/jpeg",
				"file_type":    "image",
				"file_size":    1024,
			},
		},
	})

	mockUC.On("RequestBatchUploadURL", mock.Anything, testUserID, mock.AnythingOfType("*request.BatchUploadRequestDTO")).
		Return(&response.BatchUploadResponseDTO{
			Results:      []response.UploadResponseDTO{{FileID: 1, UploadURL: "https://url1", ExpiresIn: 3600}},
			TotalCount:   1,
			SuccessCount: 1,
			FailedCount:  0,
		}, nil)

	req := httptest.NewRequest(http.MethodPost, "/files/upload/batch", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setupAuthContext(c)

	err := handler.RequestBatchUploadURL(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "https://url1")
	mockUC.AssertExpectations(t)
}

func TestBatchUploadCloudHandler_RequestBatchUploadURL_NoUserID(t *testing.T) {
	t.Log("Running: No userID -> 401")
	e := setupTestEcho()
	mockUC := new(mockBatchUploadCloudUseCase)
	handler := &BatchUploadCloudRepositoryHandler{UseCase: mockUC}

	body := mustJSON(t, map[string]interface{}{
		"files": []map[string]interface{}{
			{"file_name": "test.jpg", "content_type": "image/jpeg", "file_type": "image", "file_size": 1024},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/files/upload/batch", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.RequestBatchUploadURL(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	mockUC.AssertNotCalled(t, "RequestBatchUploadURL")
}

func TestBatchUploadCloudHandler_RequestBatchUploadURL_InvalidBody(t *testing.T) {
	t.Log("Running: Invalid body -> 400")
	e := setupTestEcho()
	mockUC := new(mockBatchUploadCloudUseCase)
	handler := &BatchUploadCloudRepositoryHandler{UseCase: mockUC}

	req := httptest.NewRequest(http.MethodPost, "/files/upload/batch", bytes.NewReader([]byte(`{"files": [invalid]`)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setupAuthContext(c)

	err := handler.RequestBatchUploadURL(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	mockUC.AssertNotCalled(t, "RequestBatchUploadURL")
}

func TestBatchUploadCloudHandler_RequestBatchUploadURL_UseCaseError(t *testing.T) {
	t.Log("Running: UseCase error -> 500")
	e := setupTestEcho()
	mockUC := new(mockBatchUploadCloudUseCase)
	handler := &BatchUploadCloudRepositoryHandler{UseCase: mockUC}

	body := mustJSON(t, map[string]interface{}{
		"files": []map[string]interface{}{
			{"file_name": "test.jpg", "content_type": "image/jpeg", "file_type": "image", "file_size": 1024},
		},
	})

	mockUC.On("RequestBatchUploadURL", mock.Anything, testUserID, mock.AnythingOfType("*request.BatchUploadRequestDTO")).
		Return(nil, assert.AnError)

	req := httptest.NewRequest(http.MethodPost, "/files/upload/batch", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setupAuthContext(c)

	err := handler.RequestBatchUploadURL(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockUC.AssertExpectations(t)
}
