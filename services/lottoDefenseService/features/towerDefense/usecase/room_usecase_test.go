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
