package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/response"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockDownloadCloudUseCase struct {
	mock.Mock
}

func (m *mockDownloadCloudUseCase) RequestDownloadURL(ctx context.Context, userID uint, fileID uint) (*response.DownloadResponseDTO, error) {
	args := m.Called(ctx, userID, fileID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.DownloadResponseDTO), args.Error(1)
}

func TestDownloadCloudHandler_RequestDownloadURL_Success(t *testing.T) {
	t.Log("Running: Success - GET /files/1/download -> 200")
	e := setupTestEcho()
	mockUC := new(mockDownloadCloudUseCase)
	handler := &DownloadCloudRepositoryHandler{UseCase: mockUC}

	mockUC.On("RequestDownloadURL", mock.Anything, testUserID, uint(1)).
		Return(&response.DownloadResponseDTO{
			DownloadURL: "https://download.url",
			FileName:    "test.jpg",
			ExpiresIn:   3600,
		}, nil)

	req := httptest.NewRequest(http.MethodGet, "/files/1/download", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("1")
	setupAuthContext(c)

	err := handler.RequestDownloadURL(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "https://download.url")
	mockUC.AssertExpectations(t)
}

func TestDownloadCloudHandler_RequestDownloadURL_NoUserID(t *testing.T) {
	t.Log("Running: No userID -> 401")
	e := setupTestEcho()
	mockUC := new(mockDownloadCloudUseCase)
	handler := &DownloadCloudRepositoryHandler{UseCase: mockUC}

	req := httptest.NewRequest(http.MethodGet, "/files/1/download", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("1")

	err := handler.RequestDownloadURL(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "unauthorized")
	mockUC.AssertNotCalled(t, "RequestDownloadURL")
}

func TestDownloadCloudHandler_RequestDownloadURL_InvalidFileID(t *testing.T) {
	t.Log("Running: Invalid file ID -> 400")
	e := setupTestEcho()
	mockUC := new(mockDownloadCloudUseCase)
	handler := &DownloadCloudRepositoryHandler{UseCase: mockUC}

	req := httptest.NewRequest(http.MethodGet, "/files/abc/download", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("abc")
	setupAuthContext(c)

	err := handler.RequestDownloadURL(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "invalid file ID")
	mockUC.AssertNotCalled(t, "RequestDownloadURL")
}

func TestDownloadCloudHandler_RequestDownloadURL_UseCaseError_FileNotFound(t *testing.T) {
	t.Log("Running: UseCase error with 'not found' -> 404")
	e := setupTestEcho()
	mockUC := new(mockDownloadCloudUseCase)
	handler := &DownloadCloudRepositoryHandler{UseCase: mockUC}

	mockUC.On("RequestDownloadURL", mock.Anything, testUserID, uint(999)).
		Return(nil, errors.New("file not found"))

	req := httptest.NewRequest(http.MethodGet, "/files/999/download", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("999")
	setupAuthContext(c)

	err := handler.RequestDownloadURL(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Contains(t, rec.Body.String(), "file not found")
	mockUC.AssertExpectations(t)
}

func TestDownloadCloudHandler_RequestDownloadURL_UseCaseError_InternalServerError(t *testing.T) {
	t.Log("Running: UseCase error without 'not found' -> 500")
	e := setupTestEcho()
	mockUC := new(mockDownloadCloudUseCase)
	handler := &DownloadCloudRepositoryHandler{UseCase: mockUC}

	mockUC.On("RequestDownloadURL", mock.Anything, testUserID, uint(1)).
		Return(nil, errors.New("failed to generate download URL"))

	req := httptest.NewRequest(http.MethodGet, "/files/1/download", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("1")
	setupAuthContext(c)

	err := handler.RequestDownloadURL(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Contains(t, rec.Body.String(), "internal server error")
	mockUC.AssertExpectations(t)
}
