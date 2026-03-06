package handler

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/request"
	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/response"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockFavoriteUseCase struct {
	mock.Mock
}

func (m *mockFavoriteUseCase) AddFavorite(ctx context.Context, userID, fileID uint) (*response.FavoriteResponseDTO, error) {
	args := m.Called(ctx, userID, fileID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.FavoriteResponseDTO), args.Error(1)
}

func (m *mockFavoriteUseCase) RemoveFavorite(ctx context.Context, userID, fileID uint) error {
	args := m.Called(ctx, userID, fileID)
	return args.Error(0)
}

func (m *mockFavoriteUseCase) ListFavorites(ctx context.Context, userID uint, filter request.ListFavoritesRequestDTO) (*response.ListFavoritesResponseDTO, error) {
	args := m.Called(ctx, userID, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.ListFavoritesResponseDTO), args.Error(1)
}

func TestFavoriteHandler_AddFavorite_Success(t *testing.T) {
	t.Log("Running: AddFavorite success -> 200")
	e := setupTestEcho()
	mockUC := new(mockFavoriteUseCase)
	handler := &FavoriteHandler{UseCase: mockUC}

	body := mustJSON(t, map[string]interface{}{"fileId": float64(1)})

	mockUC.On("AddFavorite", mock.Anything, testUserID, uint(1)).
		Return(&response.FavoriteResponseDTO{Success: true, FavoritedAt: time.Now()}, nil)

	req := httptest.NewRequest(http.MethodPost, "/favorites", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setupAuthContext(c)

	err := handler.AddFavorite(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "success")
	mockUC.AssertExpectations(t)
}

func TestFavoriteHandler_RemoveFavorite_Success(t *testing.T) {
	t.Log("Running: RemoveFavorite success -> 204")
	e := setupTestEcho()
	mockUC := new(mockFavoriteUseCase)
	handler := &FavoriteHandler{UseCase: mockUC}

	mockUC.On("RemoveFavorite", mock.Anything, testUserID, uint(1)).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/favorites/1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("fileId")
	c.SetParamValues("1")
	setupAuthContext(c)

	err := handler.RemoveFavorite(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, rec.Code)
	mockUC.AssertExpectations(t)
}

func TestFavoriteHandler_ListFavorites_Success(t *testing.T) {
	t.Log("Running: ListFavorites success -> 200")
	e := setupTestEcho()
	mockUC := new(mockFavoriteUseCase)
	handler := &FavoriteHandler{UseCase: mockUC}

	mockUC.On("ListFavorites", mock.Anything, testUserID, mock.AnythingOfType("request.ListFavoritesRequestDTO")).
		Return(&response.ListFavoritesResponseDTO{
			Data:       []response.FileInfoDTO{},
			Pagination: response.PaginationMeta{Total: 0, Page: 1, Size: 20, TotalPages: 0},
		}, nil)

	req := httptest.NewRequest(http.MethodGet, "/favorites", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setupAuthContext(c)

	err := handler.ListFavorites(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockUC.AssertExpectations(t)
}

func TestFavoriteHandler_NoUserID(t *testing.T) {
	t.Log("Running: No userID -> 401")
	e := setupTestEcho()
	mockUC := new(mockFavoriteUseCase)
	handler := &FavoriteHandler{UseCase: mockUC}

	req := httptest.NewRequest(http.MethodGet, "/favorites", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.ListFavorites(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	mockUC.AssertNotCalled(t, "ListFavorites")
}

func TestFavoriteHandler_InvalidFileID(t *testing.T) {
	t.Log("Running: Invalid file ID -> 400")
	e := setupTestEcho()
	mockUC := new(mockFavoriteUseCase)
	handler := &FavoriteHandler{UseCase: mockUC}

	req := httptest.NewRequest(http.MethodDelete, "/favorites/xyz", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("fileId")
	c.SetParamValues("xyz")
	setupAuthContext(c)

	err := handler.RemoveFavorite(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "invalid file ID")
	mockUC.AssertNotCalled(t, "RemoveFavorite")
}

func TestFavoriteHandler_AddFavorite_ValidationError_MissingFileID(t *testing.T) {
	t.Log("Running: AddFavorite missing fileId -> 400")
	e := setupTestEcho()
	mockUC := new(mockFavoriteUseCase)
	handler := &FavoriteHandler{UseCase: mockUC}

	body := mustJSON(t, map[string]interface{}{})
	req := httptest.NewRequest(http.MethodPost, "/favorites", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setupAuthContext(c)

	err := handler.AddFavorite(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	mockUC.AssertNotCalled(t, "AddFavorite")
}

func TestFavoriteHandler_AddFavorite_UseCaseError_NotFound(t *testing.T) {
	t.Log("Running: AddFavorite file not found -> 404")
	e := setupTestEcho()
	mockUC := new(mockFavoriteUseCase)
	handler := &FavoriteHandler{UseCase: mockUC}

	body := mustJSON(t, map[string]interface{}{"fileId": float64(999)})
	mockUC.On("AddFavorite", mock.Anything, testUserID, uint(999)).
		Return(nil, fmt.Errorf("file not found"))

	req := httptest.NewRequest(http.MethodPost, "/favorites", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setupAuthContext(c)

	err := handler.AddFavorite(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	mockUC.AssertExpectations(t)
}

func TestFavoriteHandler_AddFavorite_UseCaseError_Forbidden(t *testing.T) {
	t.Log("Running: AddFavorite access denied -> 403")
	e := setupTestEcho()
	mockUC := new(mockFavoriteUseCase)
	handler := &FavoriteHandler{UseCase: mockUC}

	body := mustJSON(t, map[string]interface{}{"fileId": float64(1)})
	mockUC.On("AddFavorite", mock.Anything, testUserID, uint(1)).
		Return(nil, fmt.Errorf("you do not own this file"))

	req := httptest.NewRequest(http.MethodPost, "/favorites", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setupAuthContext(c)

	err := handler.AddFavorite(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockUC.AssertExpectations(t)
}

func TestFavoriteHandler_RemoveFavorite_UseCaseError(t *testing.T) {
	t.Log("Running: RemoveFavorite UseCase error -> 500")
	e := setupTestEcho()
	mockUC := new(mockFavoriteUseCase)
	handler := &FavoriteHandler{UseCase: mockUC}

	mockUC.On("RemoveFavorite", mock.Anything, testUserID, uint(1)).Return(assert.AnError)

	req := httptest.NewRequest(http.MethodDelete, "/favorites/1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("fileId")
	c.SetParamValues("1")
	setupAuthContext(c)

	err := handler.RemoveFavorite(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockUC.AssertExpectations(t)
}

func TestFavoriteHandler_ListFavorites_UseCaseError(t *testing.T) {
	t.Log("Running: ListFavorites UseCase error -> 500")
	e := setupTestEcho()
	mockUC := new(mockFavoriteUseCase)
	handler := &FavoriteHandler{UseCase: mockUC}

	mockUC.On("ListFavorites", mock.Anything, testUserID, mock.AnythingOfType("request.ListFavoritesRequestDTO")).
		Return(nil, assert.AnError)

	req := httptest.NewRequest(http.MethodGet, "/favorites", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setupAuthContext(c)

	err := handler.ListFavorites(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockUC.AssertExpectations(t)
}
