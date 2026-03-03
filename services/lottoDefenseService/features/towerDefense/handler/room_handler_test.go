package handler

import (
	"bytes"
	"context"
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

type mockTDRoomUseCase struct {
	mock.Mock
}

func (m *mockTDRoomUseCase) CreateRoom(ctx context.Context, userID uint, req *request.CreateRoomRequest) (*response.RoomResponse, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.RoomResponse), args.Error(1)
}

func (m *mockTDRoomUseCase) JoinRoom(ctx context.Context, userID uint, req *request.JoinRoomRequest) (*response.RoomResponse, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.RoomResponse), args.Error(1)
}

func (m *mockTDRoomUseCase) GetRoom(ctx context.Context, roomID uint) (*response.RoomDetailResponse, error) {
	args := m.Called(ctx, roomID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.RoomDetailResponse), args.Error(1)
}

func (m *mockTDRoomUseCase) LeaveRoom(ctx context.Context, userID uint, roomID uint) error {
	args := m.Called(ctx, userID, roomID)
	return args.Error(0)
}

func (m *mockTDRoomUseCase) SetReady(ctx context.Context, userID uint, roomID uint, isReady bool) (*response.RoomDetailResponse, error) {
	args := m.Called(ctx, userID, roomID, isReady)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.RoomDetailResponse), args.Error(1)
}

func (m *mockTDRoomUseCase) UpdatePlayerState(ctx context.Context, roomID, userID uint, req *request.UpdateGameStateRequest) error {
	args := m.Called(ctx, roomID, userID, req)
	return args.Error(0)
}

func (m *mockTDRoomUseCase) GetOpponentState(ctx context.Context, roomID, userID uint) (*response.OpponentStateResponse, error) {
	args := m.Called(ctx, roomID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.OpponentStateResponse), args.Error(1)
}

func TestTDRoomHandler_CreateRoom_Success(t *testing.T) {
	t.Log("CreateRoom: success -> 201")
	e := setupTDTestEcho()
	mockUC := new(mockTDRoomUseCase)
	h := &TDRoomHandler{uc: mockUC}

	body := tdMustJSON(t, &request.CreateRoomRequest{RoomType: "random"})
	mockUC.On("CreateRoom", mock.Anything, tdTestUserID, mock.AnythingOfType("*request.CreateRoomRequest")).
		Return(&response.RoomResponse{RoomID: 1, RoomCode: "ABCD", RoomType: "random", Status: "waiting", PlayerSlot: 0}, nil)

	req := httptest.NewRequest(http.MethodPost, "/coop/rooms", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setupTDAuthContext(c)

	err := h.CreateRoom(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
	mockUC.AssertExpectations(t)
}

func TestTDRoomHandler_CreateRoom_NoUserID(t *testing.T) {
	t.Log("CreateRoom: no userID -> 401")
	e := setupTDTestEcho()
	mockUC := new(mockTDRoomUseCase)
	h := &TDRoomHandler{uc: mockUC}

	body := tdMustJSON(t, &request.CreateRoomRequest{RoomType: "random"})
	req := httptest.NewRequest(http.MethodPost, "/coop/rooms", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateRoom(c)
	require.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusUnauthorized, he.Code)
	mockUC.AssertNotCalled(t, "CreateRoom")
}

func TestTDRoomHandler_JoinRoom_Success(t *testing.T) {
	t.Log("JoinRoom: success")
	e := setupTDTestEcho()
	mockUC := new(mockTDRoomUseCase)
	h := &TDRoomHandler{uc: mockUC}

	body := tdMustJSON(t, &request.JoinRoomRequest{RoomCode: "ABCD"})
	mockUC.On("JoinRoom", mock.Anything, tdTestUserID, mock.AnythingOfType("*request.JoinRoomRequest")).
		Return(&response.RoomResponse{RoomID: 1, RoomCode: "ABCD", RoomType: "random", Status: "waiting", PlayerSlot: 1}, nil)

	req := httptest.NewRequest(http.MethodPost, "/coop/rooms/join", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setupTDAuthContext(c)

	err := h.JoinRoom(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockUC.AssertExpectations(t)
}

func TestTDRoomHandler_JoinRoom_RoomNotFound(t *testing.T) {
	t.Log("JoinRoom: room not found -> 404")
	e := setupTDTestEcho()
	mockUC := new(mockTDRoomUseCase)
	h := &TDRoomHandler{uc: mockUC}

	body := tdMustJSON(t, &request.JoinRoomRequest{RoomCode: "XXXX"})
	mockUC.On("JoinRoom", mock.Anything, tdTestUserID, mock.AnythingOfType("*request.JoinRoomRequest")).
		Return(nil, errors.New("room not found"))

	req := httptest.NewRequest(http.MethodPost, "/coop/rooms/join", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setupTDAuthContext(c)

	err := h.JoinRoom(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	mockUC.AssertExpectations(t)
}

func TestTDRoomHandler_JoinRoom_RoomFull(t *testing.T) {
	t.Log("JoinRoom: room full -> 409")
	e := setupTDTestEcho()
	mockUC := new(mockTDRoomUseCase)
	h := &TDRoomHandler{uc: mockUC}

	body := tdMustJSON(t, &request.JoinRoomRequest{RoomCode: "ABCD"})
	mockUC.On("JoinRoom", mock.Anything, tdTestUserID, mock.AnythingOfType("*request.JoinRoomRequest")).
		Return(nil, errors.New("room is full"))

	req := httptest.NewRequest(http.MethodPost, "/coop/rooms/join", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setupTDAuthContext(c)

	err := h.JoinRoom(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusConflict, rec.Code)
	mockUC.AssertExpectations(t)
}

func TestTDRoomHandler_GetRoom_Success(t *testing.T) {
	t.Log("GetRoom: success")
	e := setupTDTestEcho()
	mockUC := new(mockTDRoomUseCase)
	h := &TDRoomHandler{uc: mockUC}

	mockUC.On("GetRoom", mock.Anything, uint(1)).
		Return(&response.RoomDetailResponse{RoomID: 1, RoomCode: "ABCD", Status: "waiting", CurrentPlayers: 1, MaxPlayers: 2, Players: []response.PlayerInfo{}}, nil)

	req := httptest.NewRequest(http.MethodGet, "/coop/rooms/1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/coop/rooms/:id")
	c.SetParamNames("id")
	c.SetParamValues("1")

	err := h.GetRoom(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockUC.AssertExpectations(t)
}

func TestTDRoomHandler_GetRoom_InvalidID(t *testing.T) {
	t.Log("GetRoom: invalid ID -> 400")
	e := setupTDTestEcho()
	mockUC := new(mockTDRoomUseCase)
	h := &TDRoomHandler{uc: mockUC}

	req := httptest.NewRequest(http.MethodGet, "/coop/rooms/abc", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/coop/rooms/:id")
	c.SetParamNames("id")
	c.SetParamValues("abc")

	err := h.GetRoom(c)
	require.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, he.Code)
	mockUC.AssertNotCalled(t, "GetRoom")
}

func TestTDRoomHandler_LeaveRoom_Success(t *testing.T) {
	t.Log("LeaveRoom: success")
	e := setupTDTestEcho()
	mockUC := new(mockTDRoomUseCase)
	h := &TDRoomHandler{uc: mockUC}

	mockUC.On("LeaveRoom", mock.Anything, tdTestUserID, uint(1)).Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/coop/rooms/1/leave", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/coop/rooms/:id/leave")
	c.SetParamNames("id")
	c.SetParamValues("1")
	setupTDAuthContext(c)

	err := h.LeaveRoom(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockUC.AssertExpectations(t)
}

func TestTDRoomHandler_SetReady_Success(t *testing.T) {
	t.Log("SetReady: success")
	e := setupTDTestEcho()
	mockUC := new(mockTDRoomUseCase)
	h := &TDRoomHandler{uc: mockUC}

	body := tdMustJSON(t, &request.SetReadyRequest{IsReady: true})
	mockUC.On("SetReady", mock.Anything, tdTestUserID, uint(1), true).
		Return(&response.RoomDetailResponse{RoomID: 1, RoomCode: "ABCD", Status: "waiting", Players: []response.PlayerInfo{}}, nil)

	req := httptest.NewRequest(http.MethodPost, "/coop/rooms/1/ready", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/coop/rooms/:id/ready")
	c.SetParamNames("id")
	c.SetParamValues("1")
	setupTDAuthContext(c)

	err := h.SetReady(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockUC.AssertExpectations(t)
}
