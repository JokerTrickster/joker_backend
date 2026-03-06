package handler

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/lottoDefense/model/request"
	"github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/lottoDefense/model/response"
	"github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/lottoDefense/usecase"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockGameRoundUseCase struct {
	mock.Mock
}

func (m *mockGameRoundUseCase) StartRound(ctx context.Context, userID uint) (*response.RoundResponse, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.RoundResponse), args.Error(1)
}

func (m *mockGameRoundUseCase) EndRound(ctx context.Context, userID uint, roundID uint, req *request.EndRoundRequest) (*response.RoundWithDrawResponse, error) {
	args := m.Called(ctx, userID, roundID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.RoundWithDrawResponse), args.Error(1)
}

func (m *mockGameRoundUseCase) GetMyRounds(ctx context.Context, userID uint, limit int) ([]response.RoundResponse, error) {
	args := m.Called(ctx, userID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]response.RoundResponse), args.Error(1)
}

func (m *mockGameRoundUseCase) GetRound(ctx context.Context, userID uint, roundID uint) (*response.RoundWithDrawResponse, error) {
	args := m.Called(ctx, userID, roundID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.RoundWithDrawResponse), args.Error(1)
}

func TestRoundHandler_StartRound_Success(t *testing.T) {
	t.Log("StartRound: success -> 201")
	e := setupTestEcho()
	mockUC := new(mockGameRoundUseCase)
	h := &RoundHandler{uc: mockUC}

	now := time.Now()
	mockUC.On("StartRound", mock.Anything, testUserID).
		Return(&response.RoundResponse{
			ID:        1,
			UserID:    testUserID,
			Status:    "active",
			StartedAt: &now,
			CreatedAt: now,
		}, nil)

	req := httptest.NewRequest(http.MethodPost, "/rounds", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setupAuthContext(c)

	err := h.StartRound(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Contains(t, rec.Body.String(), `"id":1`)
	mockUC.AssertExpectations(t)
}

func TestRoundHandler_StartRound_NoUserID(t *testing.T) {
	t.Log("StartRound: no userID -> error")
	e := setupTestEcho()
	mockUC := new(mockGameRoundUseCase)
	h := &RoundHandler{uc: mockUC}

	req := httptest.NewRequest(http.MethodPost, "/rounds", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.StartRound(c)
	require.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusUnauthorized, he.Code)
	mockUC.AssertNotCalled(t, "StartRound")
}

func TestRoundHandler_EndRound_Success(t *testing.T) {
	t.Log("EndRound: success -> 200")
	e := setupTestEcho()
	mockUC := new(mockGameRoundUseCase)
	h := &RoundHandler{uc: mockUC}

	body := mustJSON(t, &request.EndRoundRequest{Score: 500})
	mockUC.On("EndRound", mock.Anything, testUserID, uint(1), mock.AnythingOfType("*request.EndRoundRequest")).
		Return(&response.RoundWithDrawResponse{
			RoundResponse: response.RoundResponse{ID: 1, UserID: testUserID, Status: "completed"},
			Numbers:       []int{1, 2, 3, 4, 5, 6},
		}, nil)

	req := httptest.NewRequest(http.MethodPatch, "/rounds/1/end", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/rounds/:id/end")
	c.SetParamNames("id")
	c.SetParamValues("1")
	setupAuthContext(c)

	err := h.EndRound(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockUC.AssertExpectations(t)
}

func TestRoundHandler_EndRound_NoUserID(t *testing.T) {
	t.Log("EndRound: no userID -> 401")
	e := setupTestEcho()
	mockUC := new(mockGameRoundUseCase)
	h := &RoundHandler{uc: mockUC}

	body := mustJSON(t, &request.EndRoundRequest{Score: 500})
	req := httptest.NewRequest(http.MethodPatch, "/rounds/1/end", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/rounds/:id/end")
	c.SetParamNames("id")
	c.SetParamValues("1")

	err := h.EndRound(c)
	require.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusUnauthorized, he.Code)
	mockUC.AssertNotCalled(t, "EndRound")
}

func TestRoundHandler_EndRound_InvalidRoundID(t *testing.T) {
	t.Log("EndRound: invalid round ID -> 400")
	e := setupTestEcho()
	mockUC := new(mockGameRoundUseCase)
	h := &RoundHandler{uc: mockUC}

	body := mustJSON(t, &request.EndRoundRequest{Score: 500})
	req := httptest.NewRequest(http.MethodPatch, "/rounds/abc/end", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/rounds/:id/end")
	c.SetParamNames("id")
	c.SetParamValues("abc")
	setupAuthContext(c)

	err := h.EndRound(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	mockUC.AssertNotCalled(t, "EndRound")
}

func TestRoundHandler_EndRound_BindError(t *testing.T) {
	t.Log("EndRound: invalid JSON body -> 400")
	e := setupTestEcho()
	mockUC := new(mockGameRoundUseCase)
	h := &RoundHandler{uc: mockUC}

	req := httptest.NewRequest(http.MethodPatch, "/rounds/1/end", bytes.NewReader([]byte("{invalid}")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/rounds/:id/end")
	c.SetParamNames("id")
	c.SetParamValues("1")
	setupAuthContext(c)

	err := h.EndRound(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	mockUC.AssertNotCalled(t, "EndRound")
}

func TestRoundHandler_EndRound_ValidationError(t *testing.T) {
	t.Log("EndRound: validation error -> 400")
	e := setupTestEcho()
	mockUC := new(mockGameRoundUseCase)
	h := &RoundHandler{uc: mockUC}

	body := mustJSON(t, map[string]interface{}{"score": "invalid"})
	req := httptest.NewRequest(http.MethodPatch, "/rounds/1/end", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/rounds/:id/end")
	c.SetParamNames("id")
	c.SetParamValues("1")
	setupAuthContext(c)

	err := h.EndRound(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	mockUC.AssertNotCalled(t, "EndRound")
}

func TestRoundHandler_EndRound_UseCaseError(t *testing.T) {
	t.Log("EndRound: usecase error (round not found) -> 400")
	e := setupTestEcho()
	mockUC := new(mockGameRoundUseCase)
	h := &RoundHandler{uc: mockUC}

	body := mustJSON(t, &request.EndRoundRequest{Score: 500})
	mockUC.On("EndRound", mock.Anything, testUserID, uint(999), mock.AnythingOfType("*request.EndRoundRequest")).
		Return(nil, assert.AnError)

	req := httptest.NewRequest(http.MethodPatch, "/rounds/999/end", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/rounds/:id/end")
	c.SetParamNames("id")
	c.SetParamValues("999")
	setupAuthContext(c)

	err := h.EndRound(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockUC.AssertExpectations(t)
}

func TestRoundHandler_GetMyRounds_Success(t *testing.T) {
	t.Log("GetMyRounds: success -> 200")
	e := setupTestEcho()
	mockUC := new(mockGameRoundUseCase)
	h := &RoundHandler{uc: mockUC}

	mockUC.On("GetMyRounds", mock.Anything, testUserID, 20).
		Return([]response.RoundResponse{{ID: 1, UserID: testUserID, Status: "active", CreatedAt: time.Now()}}, nil)

	req := httptest.NewRequest(http.MethodGet, "/rounds", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setupAuthContext(c)

	err := h.GetMyRounds(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockUC.AssertExpectations(t)
}

func TestRoundHandler_GetMyRounds_NoUserID(t *testing.T) {
	t.Log("GetMyRounds: no userID -> error")
	e := setupTestEcho()
	mockUC := new(mockGameRoundUseCase)
	h := &RoundHandler{uc: mockUC}

	req := httptest.NewRequest(http.MethodGet, "/rounds", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.GetMyRounds(c)
	require.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusUnauthorized, he.Code)
	mockUC.AssertNotCalled(t, "GetMyRounds")
}

func TestRoundHandler_GetRound_Success(t *testing.T) {
	t.Log("GetRound: success -> 200")
	e := setupTestEcho()
	mockUC := new(mockGameRoundUseCase)
	h := &RoundHandler{uc: mockUC}

	mockUC.On("GetRound", mock.Anything, testUserID, uint(1)).
		Return(&response.RoundWithDrawResponse{
			RoundResponse: response.RoundResponse{ID: 1, UserID: testUserID, Status: "completed"},
			Numbers:       []int{1, 2, 3, 4, 5, 6},
		}, nil)

	req := httptest.NewRequest(http.MethodGet, "/rounds/1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/rounds/:id")
	c.SetParamNames("id")
	c.SetParamValues("1")
	setupAuthContext(c)

	err := h.GetRound(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockUC.AssertExpectations(t)
}

func TestRoundHandler_GetRound_NoUserID(t *testing.T) {
	t.Log("GetRound: no userID -> 401")
	e := setupTestEcho()
	mockUC := new(mockGameRoundUseCase)
	h := &RoundHandler{uc: mockUC}

	req := httptest.NewRequest(http.MethodGet, "/rounds/1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/rounds/:id")
	c.SetParamNames("id")
	c.SetParamValues("1")

	err := h.GetRound(c)
	require.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusUnauthorized, he.Code)
	mockUC.AssertNotCalled(t, "GetRound")
}

func TestRoundHandler_GetRound_InvalidID(t *testing.T) {
	t.Log("GetRound: invalid ID -> 400")
	e := setupTestEcho()
	mockUC := new(mockGameRoundUseCase)
	h := &RoundHandler{uc: mockUC}

	req := httptest.NewRequest(http.MethodGet, "/rounds/xyz", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/rounds/:id")
	c.SetParamNames("id")
	c.SetParamValues("xyz")
	setupAuthContext(c)

	err := h.GetRound(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	mockUC.AssertNotCalled(t, "GetRound")
}

func TestRoundHandler_StartRound_UseCaseError(t *testing.T) {
	t.Log("StartRound: usecase error -> 500")
	e := setupTestEcho()
	mockUC := new(mockGameRoundUseCase)
	h := &RoundHandler{uc: mockUC}

	mockUC.On("StartRound", mock.Anything, testUserID).
		Return(nil, assert.AnError)

	req := httptest.NewRequest(http.MethodPost, "/rounds", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setupAuthContext(c)

	err := h.StartRound(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockUC.AssertExpectations(t)
}

func TestRoundHandler_EndRound_ErrRoundNotFound(t *testing.T) {
	t.Log("EndRound: ErrRoundNotFound -> 400")
	e := setupTestEcho()
	mockUC := new(mockGameRoundUseCase)
	h := &RoundHandler{uc: mockUC}

	body := mustJSON(t, &request.EndRoundRequest{Score: 500})
	mockUC.On("EndRound", mock.Anything, testUserID, uint(999), mock.AnythingOfType("*request.EndRoundRequest")).
		Return(nil, usecase.ErrRoundNotFound)

	req := httptest.NewRequest(http.MethodPatch, "/rounds/999/end", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/rounds/:id/end")
	c.SetParamNames("id")
	c.SetParamValues("999")
	setupAuthContext(c)

	err := h.EndRound(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	mockUC.AssertExpectations(t)
}

func TestRoundHandler_EndRound_ErrRoundCompleted(t *testing.T) {
	t.Log("EndRound: ErrRoundCompleted -> 400")
	e := setupTestEcho()
	mockUC := new(mockGameRoundUseCase)
	h := &RoundHandler{uc: mockUC}

	body := mustJSON(t, &request.EndRoundRequest{Score: 500})
	mockUC.On("EndRound", mock.Anything, testUserID, uint(1), mock.AnythingOfType("*request.EndRoundRequest")).
		Return(nil, usecase.ErrRoundCompleted)

	req := httptest.NewRequest(http.MethodPatch, "/rounds/1/end", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/rounds/:id/end")
	c.SetParamNames("id")
	c.SetParamValues("1")
	setupAuthContext(c)

	err := h.EndRound(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	mockUC.AssertExpectations(t)
}

func TestRoundHandler_GetMyRounds_WithLimitParam(t *testing.T) {
	t.Log("GetMyRounds: with limit=50 param -> 200")
	e := setupTestEcho()
	mockUC := new(mockGameRoundUseCase)
	h := &RoundHandler{uc: mockUC}

	mockUC.On("GetMyRounds", mock.Anything, testUserID, 50).
		Return([]response.RoundResponse{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/rounds?limit=50", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setupAuthContext(c)

	err := h.GetMyRounds(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockUC.AssertExpectations(t)
}

func TestRoundHandler_GetMyRounds_UseCaseError(t *testing.T) {
	t.Log("GetMyRounds: usecase error -> 500")
	e := setupTestEcho()
	mockUC := new(mockGameRoundUseCase)
	h := &RoundHandler{uc: mockUC}

	mockUC.On("GetMyRounds", mock.Anything, testUserID, 20).
		Return(nil, assert.AnError)

	req := httptest.NewRequest(http.MethodGet, "/rounds", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setupAuthContext(c)

	err := h.GetMyRounds(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockUC.AssertExpectations(t)
}

func TestRoundHandler_GetRound_InternalError(t *testing.T) {
	t.Log("GetRound: internal error (non-ErrRoundNotFound) -> 500")
	e := setupTestEcho()
	mockUC := new(mockGameRoundUseCase)
	h := &RoundHandler{uc: mockUC}

	mockUC.On("GetRound", mock.Anything, testUserID, uint(1)).
		Return(nil, assert.AnError)

	req := httptest.NewRequest(http.MethodGet, "/rounds/1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/rounds/:id")
	c.SetParamNames("id")
	c.SetParamValues("1")
	setupAuthContext(c)

	err := h.GetRound(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockUC.AssertExpectations(t)
}

func TestRoundHandler_GetRound_NotFound(t *testing.T) {
	t.Log("GetRound: not found -> 404")
	e := setupTestEcho()
	mockUC := new(mockGameRoundUseCase)
	h := &RoundHandler{uc: mockUC}

	mockUC.On("GetRound", mock.Anything, testUserID, uint(999)).
		Return(nil, usecase.ErrRoundNotFound)

	req := httptest.NewRequest(http.MethodGet, "/rounds/999", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/rounds/:id")
	c.SetParamNames("id")
	c.SetParamValues("999")
	setupAuthContext(c)

	err := h.GetRound(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	mockUC.AssertExpectations(t)
}
