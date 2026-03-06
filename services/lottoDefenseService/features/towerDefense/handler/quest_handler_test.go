package handler

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/towerDefense/model/request"
	"github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/towerDefense/model/response"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockTDQuestUseCase struct {
	mock.Mock
}

func (m *mockTDQuestUseCase) GetActiveQuests(ctx context.Context, userID uint) ([]response.QuestResponse, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]response.QuestResponse), args.Error(1)
}

func (m *mockTDQuestUseCase) UpdateQuestProgress(ctx context.Context, userID uint, questID uint, req *request.UpdateQuestProgressRequest) (*response.QuestResponse, error) {
	args := m.Called(ctx, userID, questID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.QuestResponse), args.Error(1)
}

func (m *mockTDQuestUseCase) ClaimReward(ctx context.Context, userID uint, questID uint) (*response.ClaimRewardResponse, error) {
	args := m.Called(ctx, userID, questID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.ClaimRewardResponse), args.Error(1)
}

func TestTDQuestHandler_GetActiveQuests_Success(t *testing.T) {
	t.Log("GetActiveQuests: success")
	e := setupTDTestEcho()
	mockUC := new(mockTDQuestUseCase)
	h := &TDQuestHandler{uc: mockUC}

	mockUC.On("GetActiveQuests", mock.Anything, tdTestUserID).
		Return([]response.QuestResponse{
			{QuestID: 1, QuestType: "kill", QuestName: "Kill 10", TargetCount: 10, CurrentCount: 5, Status: "active"},
		}, nil)

	req := httptest.NewRequest(http.MethodGet, "/quests", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setupTDAuthContext(c)

	err := h.GetActiveQuests(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockUC.AssertExpectations(t)
}

func TestTDQuestHandler_GetActiveQuests_NoUserID(t *testing.T) {
	t.Log("GetActiveQuests: no userID -> 401")
	e := setupTDTestEcho()
	mockUC := new(mockTDQuestUseCase)
	h := &TDQuestHandler{uc: mockUC}

	req := httptest.NewRequest(http.MethodGet, "/quests", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.GetActiveQuests(c)
	require.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusUnauthorized, he.Code)
	mockUC.AssertNotCalled(t, "GetActiveQuests")
}

func TestTDQuestHandler_UpdateProgress_InvalidID(t *testing.T) {
	t.Log("UpdateProgress: invalid id -> 400")
	e := setupTDTestEcho()
	mockUC := new(mockTDQuestUseCase)
	h := &TDQuestHandler{uc: mockUC}

	body := tdMustJSON(t, &request.UpdateQuestProgressRequest{Increment: 5})
	req := httptest.NewRequest(http.MethodPost, "/quests/abc/progress", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/quests/:id/progress")
	c.SetParamNames("id")
	c.SetParamValues("abc")
	setupTDAuthContext(c)

	err := h.UpdateProgress(c)
	require.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, he.Code)
	mockUC.AssertNotCalled(t, "UpdateQuestProgress")
}

func TestTDQuestHandler_UpdateProgress_QuestNotCompletedError(t *testing.T) {
	t.Log("UpdateProgress: quest not completed (bad request) -> 400")
	e := setupTDTestEcho()
	mockUC := new(mockTDQuestUseCase)
	h := &TDQuestHandler{uc: mockUC}

	body := tdMustJSON(t, &request.UpdateQuestProgressRequest{Increment: 5})
	mockUC.On("UpdateQuestProgress", mock.Anything, tdTestUserID, uint(1), mock.AnythingOfType("*request.UpdateQuestProgressRequest")).
		Return(nil, errors.New("quest not completed"))

	req := httptest.NewRequest(http.MethodPost, "/quests/1/progress", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/quests/:id/progress")
	c.SetParamNames("id")
	c.SetParamValues("1")
	setupTDAuthContext(c)

	err := h.UpdateProgress(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	mockUC.AssertExpectations(t)
}

func TestTDQuestHandler_UpdateProgress_InternalError(t *testing.T) {
	t.Log("UpdateProgress: internal error -> 500")
	e := setupTDTestEcho()
	mockUC := new(mockTDQuestUseCase)
	h := &TDQuestHandler{uc: mockUC}

	body := tdMustJSON(t, &request.UpdateQuestProgressRequest{Increment: 5})
	mockUC.On("UpdateQuestProgress", mock.Anything, tdTestUserID, uint(1), mock.AnythingOfType("*request.UpdateQuestProgressRequest")).
		Return(nil, errors.New("db error"))

	req := httptest.NewRequest(http.MethodPost, "/quests/1/progress", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/quests/:id/progress")
	c.SetParamNames("id")
	c.SetParamValues("1")
	setupTDAuthContext(c)

	err := h.UpdateProgress(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockUC.AssertExpectations(t)
}

func TestTDQuestHandler_ClaimReward_QuestNotFound(t *testing.T) {
	t.Log("ClaimReward: quest not found -> 404")
	e := setupTDTestEcho()
	mockUC := new(mockTDQuestUseCase)
	h := &TDQuestHandler{uc: mockUC}

	mockUC.On("ClaimReward", mock.Anything, tdTestUserID, uint(999)).
		Return(nil, errors.New("quest not found"))

	req := httptest.NewRequest(http.MethodPost, "/quests/999/claim", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/quests/:id/claim")
	c.SetParamNames("id")
	c.SetParamValues("999")
	setupTDAuthContext(c)

	err := h.ClaimReward(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	mockUC.AssertExpectations(t)
}

func TestTDQuestHandler_ClaimReward_InvalidID(t *testing.T) {
	t.Log("ClaimReward: invalid id -> 400")
	e := setupTDTestEcho()
	mockUC := new(mockTDQuestUseCase)
	h := &TDQuestHandler{uc: mockUC}

	req := httptest.NewRequest(http.MethodPost, "/quests/xyz/claim", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/quests/:id/claim")
	c.SetParamNames("id")
	c.SetParamValues("xyz")
	setupTDAuthContext(c)

	err := h.ClaimReward(c)
	require.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, he.Code)
	mockUC.AssertNotCalled(t, "ClaimReward")
}

func TestTDQuestHandler_GetActiveQuests_UseCaseError(t *testing.T) {
	t.Log("GetActiveQuests: usecase error -> 500")
	e := setupTDTestEcho()
	mockUC := new(mockTDQuestUseCase)
	h := &TDQuestHandler{uc: mockUC}

	mockUC.On("GetActiveQuests", mock.Anything, tdTestUserID).
		Return(nil, errors.New("db error"))

	req := httptest.NewRequest(http.MethodGet, "/quests", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setupTDAuthContext(c)

	err := h.GetActiveQuests(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockUC.AssertExpectations(t)
}

func TestTDQuestHandler_UpdateProgress_Success(t *testing.T) {
	t.Log("UpdateProgress: success")
	e := setupTDTestEcho()
	mockUC := new(mockTDQuestUseCase)
	h := &TDQuestHandler{uc: mockUC}

	body := tdMustJSON(t, &request.UpdateQuestProgressRequest{Increment: 5})
	mockUC.On("UpdateQuestProgress", mock.Anything, tdTestUserID, uint(1), mock.AnythingOfType("*request.UpdateQuestProgressRequest")).
		Return(&response.QuestResponse{QuestID: 1, CurrentCount: 10, TargetCount: 10, Status: "completed", CreatedAt: time.Now()}, nil)

	req := httptest.NewRequest(http.MethodPost, "/quests/1/progress", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/quests/:id/progress")
	c.SetParamNames("id")
	c.SetParamValues("1")
	setupTDAuthContext(c)

	err := h.UpdateProgress(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockUC.AssertExpectations(t)
}

func TestTDQuestHandler_UpdateProgress_QuestNotFound(t *testing.T) {
	t.Log("UpdateProgress: quest not found -> 404")
	e := setupTDTestEcho()
	mockUC := new(mockTDQuestUseCase)
	h := &TDQuestHandler{uc: mockUC}

	body := tdMustJSON(t, &request.UpdateQuestProgressRequest{Increment: 5})
	mockUC.On("UpdateQuestProgress", mock.Anything, tdTestUserID, uint(999), mock.AnythingOfType("*request.UpdateQuestProgressRequest")).
		Return(nil, errors.New("quest not found"))

	req := httptest.NewRequest(http.MethodPost, "/quests/999/progress", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/quests/:id/progress")
	c.SetParamNames("id")
	c.SetParamValues("999")
	setupTDAuthContext(c)

	err := h.UpdateProgress(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	mockUC.AssertExpectations(t)
}

func TestTDQuestHandler_ClaimReward_Success(t *testing.T) {
	t.Log("ClaimReward: success")
	e := setupTDTestEcho()
	mockUC := new(mockTDQuestUseCase)
	h := &TDQuestHandler{uc: mockUC}

	mockUC.On("ClaimReward", mock.Anything, tdTestUserID, uint(1)).
		Return(&response.ClaimRewardResponse{
			Quest:    &response.QuestResponse{QuestID: 1, Status: "claimed"},
			Rewards:  []response.Reward{{Type: "gold", Amount: uintPtr(100)}},
			NewGold:  200,
			ClaimedAt: time.Now(),
		}, nil)

	req := httptest.NewRequest(http.MethodPost, "/quests/1/claim", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/quests/:id/claim")
	c.SetParamNames("id")
	c.SetParamValues("1")
	setupTDAuthContext(c)

	err := h.ClaimReward(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockUC.AssertExpectations(t)
}

func TestTDQuestHandler_ClaimReward_QuestNotCompleted(t *testing.T) {
	t.Log("ClaimReward: quest not completed -> 400")
	e := setupTDTestEcho()
	mockUC := new(mockTDQuestUseCase)
	h := &TDQuestHandler{uc: mockUC}

	mockUC.On("ClaimReward", mock.Anything, tdTestUserID, uint(1)).
		Return(nil, errors.New("quest not completed"))

	req := httptest.NewRequest(http.MethodPost, "/quests/1/claim", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/quests/:id/claim")
	c.SetParamNames("id")
	c.SetParamValues("1")
	setupTDAuthContext(c)

	err := h.ClaimReward(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	mockUC.AssertExpectations(t)
}

func uintPtr(v uint) *uint { return &v }
