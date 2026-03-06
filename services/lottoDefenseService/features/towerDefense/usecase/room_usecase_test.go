package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/towerDefense/model/entity"
	"github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/towerDefense/model/request"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type mockTDRoomRepository struct {
	mock.Mock
}

func (m *mockTDRoomRepository) Create(ctx context.Context, room *entity.TDRoom) error {
	args := m.Called(ctx, room)
	return args.Error(0)
}

func (m *mockTDRoomRepository) GetByID(ctx context.Context, id uint) (*entity.TDRoom, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.TDRoom), args.Error(1)
}

func (m *mockTDRoomRepository) GetByCode(ctx context.Context, code string) (*entity.TDRoom, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.TDRoom), args.Error(1)
}

func (m *mockTDRoomRepository) Update(ctx context.Context, room *entity.TDRoom) error {
	args := m.Called(ctx, room)
	return args.Error(0)
}

func (m *mockTDRoomRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockTDRoomRepository) AddPlayer(ctx context.Context, player *entity.TDRoomPlayer) error {
	args := m.Called(ctx, player)
	return args.Error(0)
}

func (m *mockTDRoomRepository) RemovePlayer(ctx context.Context, roomID, userID uint) error {
	args := m.Called(ctx, roomID, userID)
	return args.Error(0)
}

func (m *mockTDRoomRepository) GetPlayers(ctx context.Context, roomID uint) ([]entity.TDRoomPlayer, error) {
	args := m.Called(ctx, roomID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.TDRoomPlayer), args.Error(1)
}

func (m *mockTDRoomRepository) UpdatePlayerReady(ctx context.Context, roomID, userID uint, isReady bool) error {
	args := m.Called(ctx, roomID, userID, isReady)
	return args.Error(0)
}

func (m *mockTDRoomRepository) UpdatePlayerState(ctx context.Context, roomID, userID uint, stateJSON string) error {
	args := m.Called(ctx, roomID, userID, stateJSON)
	return args.Error(0)
}

func (m *mockTDRoomRepository) GetOpponentState(ctx context.Context, roomID, userID uint) (string, uint, string, error) {
	args := m.Called(ctx, roomID, userID)
	return args.String(0), args.Get(1).(uint), args.String(2), args.Error(3)
}

func TestTDRoomUseCase_CreateRoom_Success(t *testing.T) {
	t.Log("CreateRoom: success")
	mockRoomRepo := new(mockTDRoomRepository)
	mockUserRepo := new(mockTDUserRepository)
	uc := NewTDRoomUseCase(mockRoomRepo, mockUserRepo, 5*time.Second)
	ctx := context.Background()
	userID := uint(1)
	req := &request.CreateRoomRequest{RoomType: "random"}

	mockRoomRepo.On("Create", ctx, mock.AnythingOfType("*entity.TDRoom")).Run(func(args mock.Arguments) {
		r := args.Get(1).(*entity.TDRoom)
		r.ID = 1
		r.RoomCode = "ABCD"
	}).Return(nil)
	mockRoomRepo.On("AddPlayer", ctx, mock.AnythingOfType("*entity.TDRoomPlayer")).Return(nil)

	resp, err := uc.CreateRoom(ctx, userID, req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, uint(1), resp.RoomID)
	assert.Len(t, resp.RoomCode, 4)
	mockRoomRepo.AssertExpectations(t)
}

func TestTDRoomUseCase_JoinRoom_Success(t *testing.T) {
	t.Log("JoinRoom: success")
	mockRoomRepo := new(mockTDRoomRepository)
	mockUserRepo := new(mockTDUserRepository)
	uc := NewTDRoomUseCase(mockRoomRepo, mockUserRepo, 5*time.Second)
	ctx := context.Background()
	userID := uint(1)
	req := &request.JoinRoomRequest{RoomCode: "ABCD"}

	room := &entity.TDRoom{ID: 1, RoomCode: "ABCD", RoomType: "random", MaxPlayers: 2, CurrentPlayers: 1, Status: "waiting"}
	mockRoomRepo.On("GetByCode", ctx, "ABCD").Return(room, nil)
	mockRoomRepo.On("AddPlayer", ctx, mock.AnythingOfType("*entity.TDRoomPlayer")).Return(nil)
	mockRoomRepo.On("Update", ctx, mock.AnythingOfType("*entity.TDRoom")).Return(nil)

	resp, err := uc.JoinRoom(ctx, userID, req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, uint(1), resp.RoomID)
	mockRoomRepo.AssertExpectations(t)
}

func TestTDRoomUseCase_JoinRoom_RoomNotFound(t *testing.T) {
	t.Log("JoinRoom: room not found")
	mockRoomRepo := new(mockTDRoomRepository)
	mockUserRepo := new(mockTDUserRepository)
	uc := NewTDRoomUseCase(mockRoomRepo, mockUserRepo, 5*time.Second)
	ctx := context.Background()
	userID := uint(1)
	req := &request.JoinRoomRequest{RoomCode: "XXXX"}

	mockRoomRepo.On("GetByCode", ctx, "XXXX").Return(nil, gorm.ErrRecordNotFound)

	resp, err := uc.JoinRoom(ctx, userID, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "room not found")
	mockRoomRepo.AssertExpectations(t)
}

func TestTDRoomUseCase_JoinRoom_RoomFull(t *testing.T) {
	t.Log("JoinRoom: room full")
	mockRoomRepo := new(mockTDRoomRepository)
	mockUserRepo := new(mockTDUserRepository)
	uc := NewTDRoomUseCase(mockRoomRepo, mockUserRepo, 5*time.Second)
	ctx := context.Background()
	userID := uint(1)
	req := &request.JoinRoomRequest{RoomCode: "ABCD"}

	room := &entity.TDRoom{ID: 1, RoomCode: "ABCD", MaxPlayers: 2, CurrentPlayers: 2, Status: "waiting"}
	mockRoomRepo.On("GetByCode", ctx, "ABCD").Return(room, nil)

	resp, err := uc.JoinRoom(ctx, userID, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "room is full")
	mockRoomRepo.AssertExpectations(t)
	mockRoomRepo.AssertNotCalled(t, "AddPlayer")
}

func TestTDRoomUseCase_GetRoom_Success(t *testing.T) {
	t.Log("GetRoom: success")
	mockRoomRepo := new(mockTDRoomRepository)
	mockUserRepo := new(mockTDUserRepository)
	uc := NewTDRoomUseCase(mockRoomRepo, mockUserRepo, 5*time.Second)
	ctx := context.Background()
	roomID := uint(1)

	mockRoomRepo.On("GetByID", ctx, roomID).Return(&entity.TDRoom{ID: 1, RoomCode: "ABCD", Status: "waiting"}, nil)
	mockRoomRepo.On("GetPlayers", ctx, roomID).Return([]entity.TDRoomPlayer{}, nil)

	resp, err := uc.GetRoom(ctx, roomID)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, uint(1), resp.RoomID)
	mockRoomRepo.AssertExpectations(t)
}

func TestTDRoomUseCase_LeaveRoom_Success(t *testing.T) {
	t.Log("LeaveRoom: success")
	mockRoomRepo := new(mockTDRoomRepository)
	mockUserRepo := new(mockTDUserRepository)
	uc := NewTDRoomUseCase(mockRoomRepo, mockUserRepo, 5*time.Second)
	ctx := context.Background()
	userID := uint(1)
	roomID := uint(1)

	mockRoomRepo.On("RemovePlayer", ctx, roomID, userID).Return(nil)

	err := uc.LeaveRoom(ctx, userID, roomID)
	require.NoError(t, err)
	mockRoomRepo.AssertExpectations(t)
}

func TestTDRoomUseCase_SetReady_Success(t *testing.T) {
	t.Log("SetReady: success")
	mockRoomRepo := new(mockTDRoomRepository)
	mockUserRepo := new(mockTDUserRepository)
	uc := NewTDRoomUseCase(mockRoomRepo, mockUserRepo, 5*time.Second)
	ctx := context.Background()
	userID := uint(1)
	roomID := uint(1)

	mockRoomRepo.On("UpdatePlayerReady", ctx, roomID, userID, true).Return(nil)
	mockRoomRepo.On("GetByID", ctx, roomID).Return(&entity.TDRoom{ID: 1, RoomCode: "ABCD", Status: "waiting"}, nil)
	mockRoomRepo.On("GetPlayers", ctx, roomID).Return([]entity.TDRoomPlayer{}, nil)

	resp, err := uc.SetReady(ctx, userID, roomID, true)
	require.NoError(t, err)
	require.NotNil(t, resp)
	mockRoomRepo.AssertExpectations(t)
}

func TestTDRoomUseCase_JoinRoom_AlreadyStarted(t *testing.T) {
	t.Log("JoinRoom: room already started")
	mockRoomRepo := new(mockTDRoomRepository)
	mockUserRepo := new(mockTDUserRepository)
	uc := NewTDRoomUseCase(mockRoomRepo, mockUserRepo, 5*time.Second)
	ctx := context.Background()
	userID := uint(1)
	req := &request.JoinRoomRequest{RoomCode: "ABCD"}

	room := &entity.TDRoom{ID: 1, RoomCode: "ABCD", MaxPlayers: 2, CurrentPlayers: 1, Status: "playing"}
	mockRoomRepo.On("GetByCode", ctx, "ABCD").Return(room, nil)

	resp, err := uc.JoinRoom(ctx, userID, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "room already started")
	mockRoomRepo.AssertExpectations(t)
	mockRoomRepo.AssertNotCalled(t, "AddPlayer")
}

func TestTDRoomUseCase_UpdatePlayerState_Success(t *testing.T) {
	t.Log("UpdatePlayerState: success")
	mockRoomRepo := new(mockTDRoomRepository)
	mockUserRepo := new(mockTDUserRepository)
	uc := NewTDRoomUseCase(mockRoomRepo, mockUserRepo, 5*time.Second)
	ctx := context.Background()
	roomID := uint(1)
	userID := uint(1)
	req := &request.UpdateGameStateRequest{Round: 5, HP: 100, Gold: 50, Kills: 10, Timestamp: 12345}

	mockRoomRepo.On("UpdatePlayerState", ctx, roomID, userID, mock.AnythingOfType("string")).Return(nil)

	err := uc.UpdatePlayerState(ctx, roomID, userID, req)
	require.NoError(t, err)
	mockRoomRepo.AssertExpectations(t)
}

func TestTDRoomUseCase_GetOpponentState_Success(t *testing.T) {
	t.Log("GetOpponentState: success")
	mockRoomRepo := new(mockTDRoomRepository)
	mockUserRepo := new(mockTDUserRepository)
	uc := NewTDRoomUseCase(mockRoomRepo, mockUserRepo, 5*time.Second)
	ctx := context.Background()
	roomID := uint(1)
	userID := uint(1)

	stateJSON := `{"round":5,"hp":80,"gold":60,"kills":15,"timestamp":12345,"is_alive":true}`
	mockRoomRepo.On("GetOpponentState", ctx, roomID, userID).
		Return(stateJSON, uint(2), "Player2", nil)

	resp, err := uc.GetOpponentState(ctx, roomID, userID)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, uint(2), resp.OpponentID)
	assert.Equal(t, "Player2", resp.OpponentName)
	assert.Equal(t, 5, resp.Round)
	assert.Equal(t, 80, resp.HP)
	assert.True(t, resp.IsAlive)
	mockRoomRepo.AssertExpectations(t)
}

func TestTDRoomUseCase_GetRoom_WithUserLookup(t *testing.T) {
	t.Log("GetRoom: with user lookup for username")
	mockRoomRepo := new(mockTDRoomRepository)
	mockUserRepo := new(mockTDUserRepository)
	uc := NewTDRoomUseCase(mockRoomRepo, mockUserRepo, 5*time.Second)
	ctx := context.Background()
	roomID := uint(1)

	players := []entity.TDRoomPlayer{
		{RoomID: roomID, UserID: 1, PlayerSlot: 0},
	}
	mockRoomRepo.On("GetByID", ctx, roomID).Return(&entity.TDRoom{ID: 1, RoomCode: "ABCD", Status: "waiting"}, nil)
	mockRoomRepo.On("GetPlayers", ctx, roomID).Return(players, nil)
	mockUserRepo.On("GetByID", ctx, uint(1)).Return(&entity.TDUser{ID: 1, Username: "Alice"}, nil)

	resp, err := uc.GetRoom(ctx, roomID)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Players, 1)
	assert.Equal(t, "Alice", resp.Players[0].Username)
	mockRoomRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestTDRoomUseCase_GetOpponentState_InvalidJSON(t *testing.T) {
	t.Log("GetOpponentState: invalid JSON in state")
	mockRoomRepo := new(mockTDRoomRepository)
	mockUserRepo := new(mockTDUserRepository)
	uc := NewTDRoomUseCase(mockRoomRepo, mockUserRepo, 5*time.Second)
	ctx := context.Background()
	roomID := uint(1)
	userID := uint(1)

	mockRoomRepo.On("GetOpponentState", ctx, roomID, userID).
		Return(`{invalid json`, uint(2), "Player2", nil)

	resp, err := uc.GetOpponentState(ctx, roomID, userID)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "failed to parse")
	mockRoomRepo.AssertExpectations(t)
}

func TestTDRoomUseCase_GetOpponentState_OpponentNotFound(t *testing.T) {
	t.Log("GetOpponentState: opponent not found (empty state)")
	mockRoomRepo := new(mockTDRoomRepository)
	mockUserRepo := new(mockTDUserRepository)
	uc := NewTDRoomUseCase(mockRoomRepo, mockUserRepo, 5*time.Second)
	ctx := context.Background()
	roomID := uint(1)
	userID := uint(1)

	mockRoomRepo.On("GetOpponentState", ctx, roomID, userID).
		Return("", uint(0), "", nil)

	resp, err := uc.GetOpponentState(ctx, roomID, userID)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "opponent not found")
	mockRoomRepo.AssertExpectations(t)
}
