package handler

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/towerDefense/model/request"
	"github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/towerDefense/model/response"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestTDCoopStateHandler_UpdateState_Success(t *testing.T) {
	t.Log("UpdateState: success")
	e := setupTDTestEcho()
	mockUC := new(mockTDRoomUseCase)
	h := &TDCoopStateHandler{uc: mockUC}

	body := tdMustJSON(t, &request.UpdateGameStateRequest{
		Round:     5,
		HP:        100,
		Gold:      50,
		Kills:     10,
		Timestamp: 12345,
	})
	mockUC.On("UpdatePlayerState", mock.Anything, uint(1), tdTestUserID, mock.AnythingOfType("*request.UpdateGameStateRequest")).Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/coop/rooms/1/state", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/coop/rooms/:id/state")
	c.SetParamNames("id")
	c.SetParamValues("1")
	setupTDAuthContext(c)

	err := h.UpdateState(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockUC.AssertExpectations(t)
}

func TestTDCoopStateHandler_UpdateState_NoUserID(t *testing.T) {
	t.Log("UpdateState: no userID -> 401")
	e := setupTDTestEcho()
	mockUC := new(mockTDRoomUseCase)
	h := &TDCoopStateHandler{uc: mockUC}

	body := tdMustJSON(t, &request.UpdateGameStateRequest{
		Round:     5,
		HP:        100,
		Gold:      50,
		Kills:     10,
		Timestamp: 12345,
	})
	req := httptest.NewRequest(http.MethodPost, "/coop/rooms/1/state", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/coop/rooms/:id/state")
	c.SetParamNames("id")
	c.SetParamValues("1")

	err := h.UpdateState(c)
	require.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusUnauthorized, he.Code)
	mockUC.AssertNotCalled(t, "UpdatePlayerState")
}

func TestTDCoopStateHandler_GetOpponentState_Success(t *testing.T) {
	t.Log("GetOpponentState: success")
	e := setupTDTestEcho()
	mockUC := new(mockTDRoomUseCase)
	h := &TDCoopStateHandler{uc: mockUC}

	mockUC.On("GetOpponentState", mock.Anything, uint(1), tdTestUserID).
		Return(&response.OpponentStateResponse{
			OpponentID:   2,
			OpponentName: "Player2",
			Round:        5,
			HP:           80,
			Gold:         60,
			Kills:        15,
			LastUpdate:   12345,
			IsAlive:      true,
		}, nil)

	req := httptest.NewRequest(http.MethodGet, "/coop/rooms/1/opponent-state", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/coop/rooms/:id/opponent-state")
	c.SetParamNames("id")
	c.SetParamValues("1")
	setupTDAuthContext(c)

	err := h.GetOpponentState(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockUC.AssertExpectations(t)
}

func TestTDCoopStateHandler_UpdateState_InvalidID(t *testing.T) {
	t.Log("UpdateState: invalid id -> 400")
	e := setupTDTestEcho()
	mockUC := new(mockTDRoomUseCase)
	h := &TDCoopStateHandler{uc: mockUC}

	body := tdMustJSON(t, &request.UpdateGameStateRequest{
		Round:     5,
		HP:        100,
		Gold:      50,
		Kills:     10,
		Timestamp: 12345,
	})
	req := httptest.NewRequest(http.MethodPost, "/coop/rooms/xyz/state", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/coop/rooms/:id/state")
	c.SetParamNames("id")
	c.SetParamValues("xyz")
	setupTDAuthContext(c)

	err := h.UpdateState(c)
	require.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, he.Code)
	mockUC.AssertNotCalled(t, "UpdatePlayerState")
}

func TestTDCoopStateHandler_UpdateState_ValidationError(t *testing.T) {
	t.Log("UpdateState: validation error -> 400")
	e := setupTDTestEcho()
	mockUC := new(mockTDRoomUseCase)
	h := &TDCoopStateHandler{uc: mockUC}

	body := tdMustJSON(t, map[string]interface{}{"round": "invalid"})
	req := httptest.NewRequest(http.MethodPost, "/coop/rooms/1/state", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/coop/rooms/:id/state")
	c.SetParamNames("id")
	c.SetParamValues("1")
	setupTDAuthContext(c)

	err := h.UpdateState(c)
	require.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, he.Code)
	mockUC.AssertNotCalled(t, "UpdatePlayerState")
}

func TestTDCoopStateHandler_UpdateState_UseCaseError(t *testing.T) {
	t.Log("UpdateState: usecase error -> 500")
	e := setupTDTestEcho()
	mockUC := new(mockTDRoomUseCase)
	h := &TDCoopStateHandler{uc: mockUC}

	body := tdMustJSON(t, &request.UpdateGameStateRequest{
		Round:     5,
		HP:        100,
		Gold:      50,
		Kills:     10,
		Timestamp: 12345,
	})
	mockUC.On("UpdatePlayerState", mock.Anything, uint(1), tdTestUserID, mock.AnythingOfType("*request.UpdateGameStateRequest")).
		Return(errors.New("room not found"))

	req := httptest.NewRequest(http.MethodPost, "/coop/rooms/1/state", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/coop/rooms/:id/state")
	c.SetParamNames("id")
	c.SetParamValues("1")
	setupTDAuthContext(c)

	err := h.UpdateState(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockUC.AssertExpectations(t)
}

func TestTDCoopStateHandler_GetOpponentState_InvalidID(t *testing.T) {
	t.Log("GetOpponentState: invalid id -> 400")
	e := setupTDTestEcho()
	mockUC := new(mockTDRoomUseCase)
	h := &TDCoopStateHandler{uc: mockUC}

	req := httptest.NewRequest(http.MethodGet, "/coop/rooms/abc/opponent-state", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/coop/rooms/:id/opponent-state")
	c.SetParamNames("id")
	c.SetParamValues("abc")
	setupTDAuthContext(c)

	err := h.GetOpponentState(c)
	require.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, he.Code)
	mockUC.AssertNotCalled(t, "GetOpponentState")
}

func TestTDCoopStateHandler_GetOpponentState_NoUserID(t *testing.T) {
	t.Log("GetOpponentState: no userID -> 401")
	e := setupTDTestEcho()
	mockUC := new(mockTDRoomUseCase)
	h := &TDCoopStateHandler{uc: mockUC}

	req := httptest.NewRequest(http.MethodGet, "/coop/rooms/1/opponent-state", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/coop/rooms/:id/opponent-state")
	c.SetParamNames("id")
	c.SetParamValues("1")

	err := h.GetOpponentState(c)
	require.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusUnauthorized, he.Code)
	mockUC.AssertNotCalled(t, "GetOpponentState")
}

func TestTDCoopStateHandler_GetOpponentState_InternalError(t *testing.T) {
	t.Log("GetOpponentState: internal error -> 500")
	e := setupTDTestEcho()
	mockUC := new(mockTDRoomUseCase)
	h := &TDCoopStateHandler{uc: mockUC}

	mockUC.On("GetOpponentState", mock.Anything, uint(1), tdTestUserID).
		Return(nil, errors.New("database connection failed"))

	req := httptest.NewRequest(http.MethodGet, "/coop/rooms/1/opponent-state", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/coop/rooms/:id/opponent-state")
	c.SetParamNames("id")
	c.SetParamValues("1")
	setupTDAuthContext(c)

	err := h.GetOpponentState(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockUC.AssertExpectations(t)
}

func TestTDCoopStateHandler_GetOpponentState_OpponentNotFound(t *testing.T) {
	t.Log("GetOpponentState: opponent not found -> 404")
	e := setupTDTestEcho()
	mockUC := new(mockTDRoomUseCase)
	h := &TDCoopStateHandler{uc: mockUC}

	mockUC.On("GetOpponentState", mock.Anything, uint(1), tdTestUserID).
		Return(nil, errors.New("opponent not found"))

	req := httptest.NewRequest(http.MethodGet, "/coop/rooms/1/opponent-state", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/coop/rooms/:id/opponent-state")
	c.SetParamNames("id")
	c.SetParamValues("1")
	setupTDAuthContext(c)

	err := h.GetOpponentState(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	mockUC.AssertExpectations(t)
}
