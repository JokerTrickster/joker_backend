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

type mockFolderShareUseCase struct {
	mock.Mock
}

func (m *mockFolderShareUseCase) ShareFolder(ctx context.Context, folderID uint, ownerID int32, req *request.ShareFolderRequestDTO) (*response.ShareFolderResponseDTO, error) {
	args := m.Called(ctx, folderID, ownerID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.ShareFolderResponseDTO), args.Error(1)
}

func (m *mockFolderShareUseCase) GetFolderShares(ctx context.Context, folderID uint, ownerID int32) (*response.FolderShareListResponseDTO, error) {
	args := m.Called(ctx, folderID, ownerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.FolderShareListResponseDTO), args.Error(1)
}

func (m *mockFolderShareUseCase) RevokeFolderShare(ctx context.Context, folderID uint, sharedWithID int32, ownerID int32) (*response.RevokeShareResponseDTO, error) {
	args := m.Called(ctx, folderID, sharedWithID, ownerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.RevokeShareResponseDTO), args.Error(1)
}

func (m *mockFolderShareUseCase) GetSharedWithMeFolders(ctx context.Context, userID int32) (*response.SharedWithMeFoldersResponseDTO, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.SharedWithMeFoldersResponseDTO), args.Error(1)
}

func (m *mockFolderShareUseCase) GetFoldersSharedByMe(ctx context.Context, ownerID int32) (*response.FoldersSharedByMeResponseDTO, error) {
	args := m.Called(ctx, ownerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.FoldersSharedByMeResponseDTO), args.Error(1)
}

type mockFileShareUseCase struct {
	mock.Mock
}

func (m *mockFileShareUseCase) ShareFile(ctx context.Context, fileID uint, ownerID int32, req *request.ShareFileRequestDTO) (*response.ShareFileResponseDTO, error) {
	args := m.Called(ctx, fileID, ownerID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.ShareFileResponseDTO), args.Error(1)
}

func (m *mockFileShareUseCase) GetFileShares(ctx context.Context, fileID uint, ownerID int32) (*response.FileShareListResponseDTO, error) {
	args := m.Called(ctx, fileID, ownerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.FileShareListResponseDTO), args.Error(1)
}

func (m *mockFileShareUseCase) RevokeFileShare(ctx context.Context, fileID uint, sharedWithID int32, ownerID int32) (*response.RevokeShareResponseDTO, error) {
	args := m.Called(ctx, fileID, sharedWithID, ownerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.RevokeShareResponseDTO), args.Error(1)
}

func (m *mockFileShareUseCase) GetSharedWithMeFiles(ctx context.Context, userID int32) (*response.SharedWithMeFilesResponseDTO, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.SharedWithMeFilesResponseDTO), args.Error(1)
}

func (m *mockFileShareUseCase) GetFilesSharedByMe(ctx context.Context, ownerID int32) (*response.FilesSharedByMeResponseDTO, error) {
	args := m.Called(ctx, ownerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.FilesSharedByMeResponseDTO), args.Error(1)
}

func TestShareHandler_ShareFolder_Success(t *testing.T) {
	t.Log("Running: ShareFolder success -> 200")
	e := setupTestEcho()
	mockFolderUC := new(mockFolderShareUseCase)
	mockFileUC := new(mockFileShareUseCase)
	handler := &ShareHandler{FolderShareUseCase: mockFolderUC, FileShareUseCase: mockFileUC}

	body := mustJSON(t, map[string]interface{}{
		"user_emails": []string{"user@example.com"},
		"permission":  "read",
	})

	mockFolderUC.On("ShareFolder", mock.Anything, uint(1), int32(1), mock.AnythingOfType("*request.ShareFolderRequestDTO")).
		Return(&response.ShareFolderResponseDTO{Message: "shared"}, nil)

	req := httptest.NewRequest(http.MethodPost, "/folders/1/share", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("1")
	setupAuthContext(c)

	err := handler.ShareFolder(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockFolderUC.AssertExpectations(t)
}

func TestShareHandler_GetFolderShares_Success(t *testing.T) {
	t.Log("Running: GetFolderShares success -> 200")
	e := setupTestEcho()
	mockFolderUC := new(mockFolderShareUseCase)
	mockFileUC := new(mockFileShareUseCase)
	handler := &ShareHandler{FolderShareUseCase: mockFolderUC, FileShareUseCase: mockFileUC}

	mockFolderUC.On("GetFolderShares", mock.Anything, uint(1), int32(1)).
		Return(&response.FolderShareListResponseDTO{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/folders/1/shares", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("1")
	setupAuthContext(c)

	err := handler.GetFolderShares(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockFolderUC.AssertExpectations(t)
}

func TestShareHandler_RevokeFolderShare_Success(t *testing.T) {
	t.Log("Running: RevokeFolderShare success -> 200")
	e := setupTestEcho()
	mockFolderUC := new(mockFolderShareUseCase)
	mockFileUC := new(mockFileShareUseCase)
	handler := &ShareHandler{FolderShareUseCase: mockFolderUC, FileShareUseCase: mockFileUC}

	mockFolderUC.On("RevokeFolderShare", mock.Anything, uint(1), int32(2), int32(1)).
		Return(&response.RevokeShareResponseDTO{Message: "revoked"}, nil)

	req := httptest.NewRequest(http.MethodDelete, "/folders/1/shares/2", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "userId")
	c.SetParamValues("1", "2")
	setupAuthContext(c)

	err := handler.RevokeFolderShare(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockFolderUC.AssertExpectations(t)
}

func TestShareHandler_GetSharedWithMeFolders_Success(t *testing.T) {
	t.Log("Running: GetSharedWithMeFolders success -> 200")
	e := setupTestEcho()
	mockFolderUC := new(mockFolderShareUseCase)
	handler := &ShareHandler{FolderShareUseCase: mockFolderUC, FileShareUseCase: nil}

	mockFolderUC.On("GetSharedWithMeFolders", mock.Anything, int32(1)).
		Return(&response.SharedWithMeFoldersResponseDTO{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/folders/shared-with-me", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setupAuthContext(c)

	err := handler.GetSharedWithMeFolders(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockFolderUC.AssertExpectations(t)
}

func TestShareHandler_ShareFile_Success(t *testing.T) {
	t.Log("Running: ShareFile success -> 200")
	e := setupTestEcho()
	mockFolderUC := new(mockFolderShareUseCase)
	mockFileUC := new(mockFileShareUseCase)
	handler := &ShareHandler{FolderShareUseCase: mockFolderUC, FileShareUseCase: mockFileUC}

	body := mustJSON(t, map[string]interface{}{
		"user_emails": []string{"user@example.com"},
		"permission":  "read",
	})

	mockFileUC.On("ShareFile", mock.Anything, uint(1), int32(1), mock.AnythingOfType("*request.ShareFileRequestDTO")).
		Return(&response.ShareFileResponseDTO{Message: "shared"}, nil)

	req := httptest.NewRequest(http.MethodPost, "/files/1/share", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("1")
	setupAuthContext(c)

	err := handler.ShareFile(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockFileUC.AssertExpectations(t)
}

func TestShareHandler_GetFileShares_Success(t *testing.T) {
	t.Log("Running: GetFileShares success -> 200")
	e := setupTestEcho()
	mockFolderUC := new(mockFolderShareUseCase)
	mockFileUC := new(mockFileShareUseCase)
	handler := &ShareHandler{FolderShareUseCase: mockFolderUC, FileShareUseCase: mockFileUC}

	mockFileUC.On("GetFileShares", mock.Anything, uint(1), int32(1)).
		Return(&response.FileShareListResponseDTO{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/files/1/shares", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("1")
	setupAuthContext(c)

	err := handler.GetFileShares(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockFileUC.AssertExpectations(t)
}

func TestShareHandler_RevokeFileShare_Success(t *testing.T) {
	t.Log("Running: RevokeFileShare success -> 200")
	e := setupTestEcho()
	mockFolderUC := new(mockFolderShareUseCase)
	mockFileUC := new(mockFileShareUseCase)
	handler := &ShareHandler{FolderShareUseCase: mockFolderUC, FileShareUseCase: mockFileUC}

	mockFileUC.On("RevokeFileShare", mock.Anything, uint(1), int32(2), int32(1)).
		Return(&response.RevokeShareResponseDTO{Message: "revoked"}, nil)

	req := httptest.NewRequest(http.MethodDelete, "/files/1/shares/2", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "userId")
	c.SetParamValues("1", "2")
	setupAuthContext(c)

	err := handler.RevokeFileShare(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockFileUC.AssertExpectations(t)
}

func TestShareHandler_GetSharedWithMeFiles_Success(t *testing.T) {
	t.Log("Running: GetSharedWithMeFiles success -> 200")
	e := setupTestEcho()
	mockFileUC := new(mockFileShareUseCase)
	handler := &ShareHandler{FolderShareUseCase: nil, FileShareUseCase: mockFileUC}

	mockFileUC.On("GetSharedWithMeFiles", mock.Anything, int32(1)).
		Return(&response.SharedWithMeFilesResponseDTO{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/files/shared-with-me", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setupAuthContext(c)

	err := handler.GetSharedWithMeFiles(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockFileUC.AssertExpectations(t)
}

func TestShareHandler_GetFoldersSharedByMe_Success(t *testing.T) {
	t.Log("Running: GetFoldersSharedByMe success -> 200")
	e := setupTestEcho()
	mockFolderUC := new(mockFolderShareUseCase)
	handler := &ShareHandler{FolderShareUseCase: mockFolderUC, FileShareUseCase: nil}

	mockFolderUC.On("GetFoldersSharedByMe", mock.Anything, int32(1)).
		Return(&response.FoldersSharedByMeResponseDTO{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/folders/shared-by-me", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setupAuthContext(c)

	err := handler.GetFoldersSharedByMe(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockFolderUC.AssertExpectations(t)
}

func TestShareHandler_GetFilesSharedByMe_Success(t *testing.T) {
	t.Log("Running: GetFilesSharedByMe success -> 200")
	e := setupTestEcho()
	mockFileUC := new(mockFileShareUseCase)
	handler := &ShareHandler{FolderShareUseCase: nil, FileShareUseCase: mockFileUC}

	mockFileUC.On("GetFilesSharedByMe", mock.Anything, int32(1)).
		Return(&response.FilesSharedByMeResponseDTO{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/files/shared-by-me", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setupAuthContext(c)

	err := handler.GetFilesSharedByMe(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockFileUC.AssertExpectations(t)
}

func TestShareHandler_NoUserID(t *testing.T) {
	t.Log("Running: No userID -> 401")
	e := setupTestEcho()
	mockFolderUC := new(mockFolderShareUseCase)
	handler := &ShareHandler{FolderShareUseCase: mockFolderUC, FileShareUseCase: nil}

	req := httptest.NewRequest(http.MethodGet, "/folders/shared-with-me", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.GetSharedWithMeFolders(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	mockFolderUC.AssertNotCalled(t, "GetSharedWithMeFolders")
}
