package handler

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/response"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockProcessingStatusUseCase struct {
	mock.Mock
}

func (m *mockProcessingStatusUseCase) GetProcessingStatus(ctx context.Context, userID uint, fileID uint) (*response.ProcessingStatusResponse, error) {
	args := m.Called(ctx, userID, fileID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.ProcessingStatusResponse), args.Error(1)
}

func (m *mockProcessingStatusUseCase) GetBatchProcessingStatus(ctx context.Context, userID uint, fileIDs []uint) (*response.BatchProcessingStatusResponse, error) {
	args := m.Called(ctx, userID, fileIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.BatchProcessingStatusResponse), args.Error(1)
}

func TestProcessingStatusHandler_GetProcessingStatus_Success(t *testing.T) {
	t.Log("Running: GetProcessingStatus success -> 200")
	e := setupTestEcho()
	mockUC := new(mockProcessingStatusUseCase)
	handler := NewProcessingStatusHandler(mockUC)

	mockUC.On("GetProcessingStatus", mock.Anything, testUserID, uint(1)).
		Return(&response.ProcessingStatusResponse{
			FileID:    1,
			Status:    "completed",
			Progress:  100,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}, nil)

	req := httptest.NewRequest(http.MethodGet, "/files/1/processing-status", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("1")
	setupAuthContext(c)

	err := handler.GetProcessingStatus(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "completed")
	mockUC.AssertExpectations(t)
}

func TestProcessingStatusHandler_GetProcessingStatus_NoUserID(t *testing.T) {
	t.Log("Running: No userID -> 401")
	e := setupTestEcho()
	mockUC := new(mockProcessingStatusUseCase)
	handler := NewProcessingStatusHandler(mockUC)

	req := httptest.NewRequest(http.MethodGet, "/files/1/processing-status", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("1")

	err := handler.GetProcessingStatus(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "UNAUTHORIZED")
	mockUC.AssertNotCalled(t, "GetProcessingStatus")
}

func TestProcessingStatusHandler_GetProcessingStatus_InvalidFileID(t *testing.T) {
	t.Log("Running: Invalid file ID -> 400")
	e := setupTestEcho()
	mockUC := new(mockProcessingStatusUseCase)
	handler := NewProcessingStatusHandler(mockUC)

	req := httptest.NewRequest(http.MethodGet, "/files/abc/processing-status", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("abc")
	setupAuthContext(c)

	err := handler.GetProcessingStatus(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "INVALID_FILE_ID")
	mockUC.AssertNotCalled(t, "GetProcessingStatus")
}

func TestProcessingStatusHandler_GetBatchProcessingStatus_Success(t *testing.T) {
	t.Log("Running: GetBatchProcessingStatus success -> 200")
	e := setupTestEcho()
	mockUC := new(mockProcessingStatusUseCase)
	handler := NewProcessingStatusHandler(mockUC)

	body := mustJSON(t, map[string]interface{}{
		"file_ids": []float64{1, 2, 3},
	})

	mockUC.On("GetBatchProcessingStatus", mock.Anything, testUserID, mock.AnythingOfType("[]uint")).
		Return(&response.BatchProcessingStatusResponse{
			Results: []response.ProcessingStatusResponse{},
		}, nil)

	req := httptest.NewRequest(http.MethodPost, "/files/processing-status/batch", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setupAuthContext(c)

	err := handler.GetBatchProcessingStatus(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockUC.AssertExpectations(t)
}

func TestProcessingStatusHandler_GetBatchProcessingStatus_EmptyFileIDs(t *testing.T) {
	t.Log("Running: Empty file IDs -> 400")
	e := setupTestEcho()
	mockUC := new(mockProcessingStatusUseCase)
	handler := NewProcessingStatusHandler(mockUC)

	body := mustJSON(t, map[string]interface{}{
		"file_ids": []interface{}{},
	})

	req := httptest.NewRequest(http.MethodPost, "/files/processing-status/batch", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setupAuthContext(c)

	err := handler.GetBatchProcessingStatus(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "At least one file ID is required")
	mockUC.AssertNotCalled(t, "GetBatchProcessingStatus")
}

func TestProcessingStatusHandler_GetBatchProcessingStatus_TooManyFileIDs(t *testing.T) {
	t.Log("Running: Too many file IDs -> 400")
	e := setupTestEcho()
	mockUC := new(mockProcessingStatusUseCase)
	handler := NewProcessingStatusHandler(mockUC)

	fileIDs := make([]float64, 101)
	for i := range fileIDs {
		fileIDs[i] = float64(i + 1)
	}
	body := mustJSON(t, map[string]interface{}{
		"file_ids": fileIDs,
	})

	req := httptest.NewRequest(http.MethodPost, "/files/processing-status/batch", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setupAuthContext(c)

	err := handler.GetBatchProcessingStatus(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "Maximum 100 file IDs allowed")
	mockUC.AssertNotCalled(t, "GetBatchProcessingStatus")
}

func TestProcessingStatusHandler_GetProcessingStatus_UseCaseError_FileNotFound(t *testing.T) {
	t.Log("Running: GetProcessingStatus FILE_NOT_FOUND -> 404")
	e := setupTestEcho()
	mockUC := new(mockProcessingStatusUseCase)
	handler := NewProcessingStatusHandler(mockUC)

	mockUC.On("GetProcessingStatus", mock.Anything, testUserID, uint(999)).
		Return(nil, fmt.Errorf("FILE_NOT_FOUND: file with ID 999 not found"))

	req := httptest.NewRequest(http.MethodGet, "/files/999/processing-status", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("999")
	setupAuthContext(c)

	err := handler.GetProcessingStatus(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Contains(t, rec.Body.String(), "FILE_NOT_FOUND")
	mockUC.AssertExpectations(t)
}

func TestProcessingStatusHandler_GetProcessingStatus_UseCaseError_Internal(t *testing.T) {
	t.Log("Running: GetProcessingStatus internal error -> 500")
	e := setupTestEcho()
	mockUC := new(mockProcessingStatusUseCase)
	handler := NewProcessingStatusHandler(mockUC)

	mockUC.On("GetProcessingStatus", mock.Anything, testUserID, uint(1)).
		Return(nil, fmt.Errorf("failed to fetch file: database error"))

	req := httptest.NewRequest(http.MethodGet, "/files/1/processing-status", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("1")
	setupAuthContext(c)

	err := handler.GetProcessingStatus(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Contains(t, rec.Body.String(), "INTERNAL_SERVER_ERROR")
	mockUC.AssertExpectations(t)
}

func TestProcessingStatusHandler_GetBatchProcessingStatus_UseCaseError(t *testing.T) {
	t.Log("Running: GetBatchProcessingStatus UseCase error -> 500")
	e := setupTestEcho()
	mockUC := new(mockProcessingStatusUseCase)
	handler := NewProcessingStatusHandler(mockUC)

	body := mustJSON(t, map[string]interface{}{"file_ids": []float64{1, 2, 3}})
	mockUC.On("GetBatchProcessingStatus", mock.Anything, testUserID, mock.AnythingOfType("[]uint")).
		Return(nil, assert.AnError)

	req := httptest.NewRequest(http.MethodPost, "/files/processing-status/batch", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setupAuthContext(c)

	err := handler.GetBatchProcessingStatus(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockUC.AssertExpectations(t)
}

func TestProcessingStatusHandler_GetBatchProcessingStatus_InvalidRequest(t *testing.T) {
	t.Log("Running: GetBatchProcessingStatus invalid request body -> 400")
	e := setupTestEcho()
	mockUC := new(mockProcessingStatusUseCase)
	handler := NewProcessingStatusHandler(mockUC)

	req := httptest.NewRequest(http.MethodPost, "/files/processing-status/batch", bytes.NewReader([]byte(`{invalid}`)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setupAuthContext(c)

	err := handler.GetBatchProcessingStatus(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	mockUC.AssertNotCalled(t, "GetBatchProcessingStatus")
}
