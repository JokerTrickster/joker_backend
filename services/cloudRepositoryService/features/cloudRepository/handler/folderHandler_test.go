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

type mockFolderUseCase struct {
	mock.Mock
}

func (m *mockFolderUseCase) CreateFolder(ctx context.Context, userID uint, req *request.CreateFolderRequestDTO) (*response.FolderResponseDTO, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.FolderResponseDTO), args.Error(1)
}

func (m *mockFolderUseCase) GetFolders(ctx context.Context, userID uint) ([]response.FolderTreeResponseDTO, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]response.FolderTreeResponseDTO), args.Error(1)
}

func (m *mockFolderUseCase) GetFolderByID(ctx context.Context, folderID uint, userID uint) (*response.FolderResponseDTO, error) {
	args := m.Called(ctx, folderID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.FolderResponseDTO), args.Error(1)
}

func (m *mockFolderUseCase) UpdateFolder(ctx context.Context, folderID uint, userID uint, req *request.UpdateFolderRequestDTO) (*response.FolderResponseDTO, error) {
	args := m.Called(ctx, folderID, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.FolderResponseDTO), args.Error(1)
}

func (m *mockFolderUseCase) DeleteFolder(ctx context.Context, folderID uint, userID uint) error {
	args := m.Called(ctx, folderID, userID)
	return args.Error(0)
}

func (m *mockFolderUseCase) GetFolderFiles(ctx context.Context, folderID uint, userID uint) ([]response.FileInfoDTO, error) {
	args := m.Called(ctx, folderID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]response.FileInfoDTO), args.Error(1)
}

func (m *mockFolderUseCase) MoveFiles(ctx context.Context, userID uint, req *request.MoveFilesToFolderRequestDTO) (*response.MoveFilesResponseDTO, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.MoveFilesResponseDTO), args.Error(1)
}

func TestFolderHandler_CreateFolder_Success(t *testing.T) {
	t.Log("Running: CreateFolder success -> 201")
	e := setupTestEcho()
	mockUC := new(mockFolderUseCase)
	handler := &FolderHandler{UseCase: mockUC}

	body := mustJSON(t, map[string]interface{}{"folder_name": "My Folder"})

	mockUC.On("CreateFolder", mock.Anything, testUserID, mock.AnythingOfType("*request.CreateFolderRequestDTO")).
		Return(&response.FolderResponseDTO{ID: 1, FolderName: "My Folder"}, nil)

	req := httptest.NewRequest(http.MethodPost, "/folders", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setupAuthContext(c)

	err := handler.CreateFolder(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
	mockUC.AssertExpectations(t)
}

func TestFolderHandler_GetFolders_Success(t *testing.T) {
	t.Log("Running: GetFolders success -> 200")
	e := setupTestEcho()
	mockUC := new(mockFolderUseCase)
	handler := &FolderHandler{UseCase: mockUC}

	mockUC.On("GetFolders", mock.Anything, testUserID).
		Return([]response.FolderTreeResponseDTO{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/folders", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setupAuthContext(c)

	err := handler.GetFolders(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockUC.AssertExpectations(t)
}

func TestFolderHandler_GetFolderByID_Success(t *testing.T) {
	t.Log("Running: GetFolderByID success -> 200")
	e := setupTestEcho()
	mockUC := new(mockFolderUseCase)
	handler := &FolderHandler{UseCase: mockUC}

	mockUC.On("GetFolderByID", mock.Anything, uint(1), testUserID).
		Return(&response.FolderResponseDTO{ID: 1, FolderName: "Test"}, nil)

	req := httptest.NewRequest(http.MethodGet, "/folders/1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("1")
	setupAuthContext(c)

	err := handler.GetFolderByID(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockUC.AssertExpectations(t)
}

func TestFolderHandler_UpdateFolder_Success(t *testing.T) {
	t.Log("Running: UpdateFolder success -> 200")
	e := setupTestEcho()
	mockUC := new(mockFolderUseCase)
	handler := &FolderHandler{UseCase: mockUC}

	body := mustJSON(t, map[string]interface{}{"folder_name": "Updated"})

	mockUC.On("UpdateFolder", mock.Anything, uint(1), testUserID, mock.AnythingOfType("*request.UpdateFolderRequestDTO")).
		Return(&response.FolderResponseDTO{ID: 1, FolderName: "Updated"}, nil)

	req := httptest.NewRequest(http.MethodPatch, "/folders/1", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("1")
	setupAuthContext(c)

	err := handler.UpdateFolder(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockUC.AssertExpectations(t)
}

func TestFolderHandler_DeleteFolder_Success(t *testing.T) {
	t.Log("Running: DeleteFolder success -> 204")
	e := setupTestEcho()
	mockUC := new(mockFolderUseCase)
	handler := &FolderHandler{UseCase: mockUC}

	mockUC.On("DeleteFolder", mock.Anything, uint(1), testUserID).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/folders/1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("1")
	setupAuthContext(c)

	err := handler.DeleteFolder(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, rec.Code)
	mockUC.AssertExpectations(t)
}

func TestFolderHandler_GetFolderFiles_Success(t *testing.T) {
	t.Log("Running: GetFolderFiles success -> 200")
	e := setupTestEcho()
	mockUC := new(mockFolderUseCase)
	handler := &FolderHandler{UseCase: mockUC}

	mockUC.On("GetFolderFiles", mock.Anything, uint(1), testUserID).
		Return([]response.FileInfoDTO{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/folders/1/files", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("1")
	setupAuthContext(c)

	err := handler.GetFolderFiles(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockUC.AssertExpectations(t)
}

func TestFolderHandler_MoveFiles_Success(t *testing.T) {
	t.Log("Running: MoveFiles success -> 200")
	e := setupTestEcho()
	mockUC := new(mockFolderUseCase)
	handler := &FolderHandler{UseCase: mockUC}

	body := mustJSON(t, map[string]interface{}{
		"file_ids":         []float64{1, 2},
		"target_folder_id": float64(1),
	})

	mockUC.On("MoveFiles", mock.Anything, testUserID, mock.AnythingOfType("*request.MoveFilesToFolderRequestDTO")).
		Return(&response.MoveFilesResponseDTO{MovedCount: 2}, nil)

	req := httptest.NewRequest(http.MethodPost, "/files/batch/move", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setupAuthContext(c)

	err := handler.MoveFiles(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockUC.AssertExpectations(t)
}

func TestFolderHandler_NoUserID(t *testing.T) {
	t.Log("Running: No userID -> 401")
	e := setupTestEcho()
	mockUC := new(mockFolderUseCase)
	handler := &FolderHandler{UseCase: mockUC}

	req := httptest.NewRequest(http.MethodGet, "/folders", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.GetFolders(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	mockUC.AssertNotCalled(t, "GetFolders")
}

func TestFolderHandler_InvalidFolderID(t *testing.T) {
	t.Log("Running: Invalid folder ID -> 400")
	e := setupTestEcho()
	mockUC := new(mockFolderUseCase)
	handler := &FolderHandler{UseCase: mockUC}

	req := httptest.NewRequest(http.MethodGet, "/folders/abc", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("abc")
	setupAuthContext(c)

	err := handler.GetFolderByID(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "invalid folder id")
	mockUC.AssertNotCalled(t, "GetFolderByID")
}
