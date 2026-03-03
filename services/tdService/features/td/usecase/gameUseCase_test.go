package usecase

import (
	"errors"
	"testing"
	"time"

	"github.com/JokerTrickster/joker_backend/services/tdService/features/td/model/entity"
	"github.com/JokerTrickster/joker_backend/services/tdService/features/td/model/request"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGameUseCase_CreateSession_Success(t *testing.T) {
	t.Log("GameUseCase.CreateSession: mock repos return player, CreateSession succeeds -> verify response")
	mockSession := new(MockGameSessionRepository)
	mockPlayer := new(MockPlayerRepository)
	uc := newTestGameUseCase(mockSession, mockPlayer)

	player := &entity.Player{
		ID:       testPlayerID,
		UserID:   testUserID,
		Nickname: testUsername,
	}
	mockPlayer.On("GetPlayerByUserID", testUserID).Return(player, nil)
	mockSession.On("CreateSession", mock.AnythingOfType("*entity.GameSession")).
		Run(func(args mock.Arguments) {
			s := args.Get(0).(*entity.GameSession)
			s.SessionID = testSessionID
			s.RoomCode = testRoomCode
		}).Return(nil)

	req := &request.CreateSessionRequest{Mode: "single", PlayerCount: 1}
	resp, err := uc.CreateSession(testUserID, req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, testSessionID, resp.SessionID)
	assert.Equal(t, testRoomCode, resp.RoomCode)
	assert.Equal(t, testWsBaseURL+"/game/"+testSessionID, resp.WebSocketURL)
	mockPlayer.AssertExpectations(t)
	mockSession.AssertExpectations(t)
}

func TestGameUseCase_CreateSession_PlayerNotFound(t *testing.T) {
	t.Log("GameUseCase.CreateSession: GetPlayerByUserID returns nil -> error 'player not found'")
	mockSession := new(MockGameSessionRepository)
	mockPlayer := new(MockPlayerRepository)
	uc := newTestGameUseCase(mockSession, mockPlayer)

	mockPlayer.On("GetPlayerByUserID", testUserID).Return(nil, nil)

	req := &request.CreateSessionRequest{Mode: "single", PlayerCount: 1}
	resp, err := uc.CreateSession(testUserID, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "player not found")
	mockSession.AssertNotCalled(t, "CreateSession")
}

func TestGameUseCase_CreateSession_RepoError(t *testing.T) {
	t.Log("GameUseCase.CreateSession: GetPlayerByUserID errors -> error")
	mockSession := new(MockGameSessionRepository)
	mockPlayer := new(MockPlayerRepository)
	uc := newTestGameUseCase(mockSession, mockPlayer)

	repoErr := errors.New("db connection failed")
	mockPlayer.On("GetPlayerByUserID", testUserID).Return(nil, repoErr)

	req := &request.CreateSessionRequest{Mode: "single", PlayerCount: 1}
	resp, err := uc.CreateSession(testUserID, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "failed to get player")
	mockSession.AssertNotCalled(t, "CreateSession")
}

func TestGameUseCase_JoinSession_Success(t *testing.T) {
	t.Log("GameUseCase.JoinSession: mock repos -> verify response")
	mockSession := new(MockGameSessionRepository)
	mockPlayer := new(MockPlayerRepository)
	uc := newTestGameUseCase(mockSession, mockPlayer)

	player := &entity.Player{
		ID:       testPlayerID,
		UserID:   testUserID,
		Nickname: testUsername,
	}
	session := &entity.GameSession{
		SessionID:   testSessionID,
		RoomCode:    testRoomCode,
		Mode:        entity.GameModeCoop,
		PlayerCount: 2,
	}

	mockPlayer.On("GetPlayerByUserID", testUserID).Return(player, nil)
	mockSession.On("GetSessionByRoomCode", testRoomCode).Return(session, nil)
	mockSession.On("JoinSession", testSessionID, testPlayerID).Return(nil)

	req := &request.JoinSessionRequest{RoomCode: testRoomCode}
	resp, err := uc.JoinSession(testUserID, req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, testSessionID, resp.SessionID)
	assert.Equal(t, testRoomCode, resp.RoomCode)
	assert.Equal(t, testWsBaseURL+"/game/"+testSessionID, resp.WebSocketURL)
	mockPlayer.AssertExpectations(t)
	mockSession.AssertExpectations(t)
}

func TestGameUseCase_JoinSession_RoomNotFound(t *testing.T) {
	t.Log("GameUseCase.JoinSession: GetSessionByRoomCode errors -> error")
	mockSession := new(MockGameSessionRepository)
	mockPlayer := new(MockPlayerRepository)
	uc := newTestGameUseCase(mockSession, mockPlayer)

	player := &entity.Player{
		ID:       testPlayerID,
		UserID:   testUserID,
		Nickname: testUsername,
	}
	repoErr := errors.New("session not found")

	mockPlayer.On("GetPlayerByUserID", testUserID).Return(player, nil)
	mockSession.On("GetSessionByRoomCode", "BADCODE").Return(nil, repoErr)

	req := &request.JoinSessionRequest{RoomCode: "BADCODE"}
	resp, err := uc.JoinSession(testUserID, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, repoErr, err)
	mockSession.AssertNotCalled(t, "JoinSession")
}

func TestGameUseCase_SaveResult_Success(t *testing.T) {
	t.Log("GameUseCase.SaveResult: mock all repos -> verify GameResultResponse")
	mockSession := new(MockGameSessionRepository)
	mockPlayer := new(MockPlayerRepository)
	uc := newTestGameUseCase(mockSession, mockPlayer)

	startedAt := time.Now().Add(-60 * time.Second)
	player := &entity.Player{
		ID:         testPlayerID,
		UserID:     testUserID,
		Nickname:   testUsername,
		Level:      1,
		Experience: 0,
	}
	session := &entity.GameSession{
		SessionID:  testSessionID,
		RoomCode:   testRoomCode,
		Mode:       entity.GameModeSingle,
		StartedAt:  &startedAt,
		PlayerID:   testPlayerID,
		Player2ID:  nil,
	}

	mockPlayer.On("GetPlayerByUserID", testUserID).Return(player, nil)
	mockSession.On("GetSessionByID", testSessionID).Return(session, nil)
	mockSession.On("SaveGameResult", mock.AnythingOfType("*entity.GameResult")).Return(nil)
	mockPlayer.On("UpdatePlayerStats", testPlayerID, mock.AnythingOfType("*entity.GameResult")).Return(nil)
	mockPlayer.On("UpdatePlayer", mock.AnythingOfType("*entity.Player")).Return(nil)
	mockSession.On("GetGameResults", testSessionID).Return([]*entity.GameResult{{}}, nil)
	mockSession.On("EndSession", testSessionID).Return(nil)

	req := &request.SaveResultRequest{
		SessionID:       testSessionID,
		Score:           1000,
		WavesCompleted:  5,
		UnitsPlaced:     10,
		EnemiesDefeated: 50,
		Victory:         true,
	}
	resp, err := uc.SaveResult(testUserID, req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.True(t, resp.Success)
	assert.Equal(t, "Game result saved successfully", resp.Message)
	mockPlayer.AssertExpectations(t)
	mockSession.AssertExpectations(t)
}

func TestGameUseCase_SaveResult_PlayerNotFound(t *testing.T) {
	t.Log("GameUseCase.SaveResult: nil player -> error")
	mockSession := new(MockGameSessionRepository)
	mockPlayer := new(MockPlayerRepository)
	uc := newTestGameUseCase(mockSession, mockPlayer)

	mockPlayer.On("GetPlayerByUserID", testUserID).Return(nil, nil)

	req := &request.SaveResultRequest{
		SessionID:       testSessionID,
		Score:           1000,
		WavesCompleted:  5,
		UnitsPlaced:     10,
		EnemiesDefeated: 50,
		Victory:         false,
	}
	resp, err := uc.SaveResult(testUserID, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "player not found")
	mockSession.AssertNotCalled(t, "GetSessionByID")
}

func TestGameUseCase_GetProfile_Success(t *testing.T) {
	t.Log("GameUseCase.GetProfile: valid userID string -> profile response")
	mockSession := new(MockGameSessionRepository)
	mockPlayer := new(MockPlayerRepository)
	uc := newTestGameUseCase(mockSession, mockPlayer)

	player := &entity.Player{
		ID:         testPlayerID,
		UserID:     testUserID,
		Nickname:   testUsername,
		AvatarID:   "avatar1",
		Level:      3,
		Experience: 500,
		Stats: entity.PlayerStats{
			ID:             1,
			GamesPlayed:    5,
			Victories:      3,
			TotalScore:     20000,
			HighestScore:   6000,
			HighestWave:    12,
		},
	}
	mockPlayer.On("GetPlayerByUserID", testUserID).Return(player, nil)

	resp, err := uc.GetProfile("42")

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, testUsername, resp.Nickname)
	assert.Equal(t, "avatar1", resp.AvatarID)
	assert.Equal(t, 3, resp.Level)
	assert.Equal(t, 500, resp.Experience)
	require.NotNil(t, resp.Stats)
	assert.Equal(t, 5, resp.Stats.GamesPlayed)
	assert.Equal(t, 3, resp.Stats.Victories)
	mockSession.AssertNotCalled(t, "GetSessionByID")
}

func TestGameUseCase_GetProfile_InvalidUserID(t *testing.T) {
	t.Log("GameUseCase.GetProfile: non-numeric string -> error")
	mockSession := new(MockGameSessionRepository)
	mockPlayer := new(MockPlayerRepository)
	uc := newTestGameUseCase(mockSession, mockPlayer)

	resp, err := uc.GetProfile("not-a-number")

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "invalid user ID")
	mockPlayer.AssertNotCalled(t, "GetPlayerByUserID")
}

func TestGameUseCase_GetProfile_NotFound(t *testing.T) {
	t.Log("GameUseCase.GetProfile: nil player -> error")
	mockSession := new(MockGameSessionRepository)
	mockPlayer := new(MockPlayerRepository)
	uc := newTestGameUseCase(mockSession, mockPlayer)

	mockPlayer.On("GetPlayerByUserID", testUserID).Return(nil, nil)

	resp, err := uc.GetProfile("42")

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "player not found")
}

func TestGameUseCase_GetSessionInfo_Success(t *testing.T) {
	t.Log("GameUseCase.GetSessionInfo: mock session with players -> verify response")
	mockSession := new(MockGameSessionRepository)
	mockPlayer := new(MockPlayerRepository)
	uc := newTestGameUseCase(mockSession, mockPlayer)

	player1 := &entity.Player{
		ID:       1,
		UserID:   10,
		Nickname: "player1",
		AvatarID: "av1",
		Level:    1,
	}
	player2 := &entity.Player{
		ID:       2,
		UserID:   20,
		Nickname: "player2",
		AvatarID: "av2",
		Level:    2,
	}
	session := &entity.GameSession{
		SessionID:   testSessionID,
		RoomCode:    testRoomCode,
		Mode:        entity.GameModeCoop,
		Status:      entity.GameStatusPlaying,
		PlayerCount: 2,
		Player:      player1,
		Player2:     player2,
	}
	mockSession.On("GetSessionByID", testSessionID).Return(session, nil)

	resp, err := uc.GetSessionInfo(testSessionID)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, testSessionID, resp.SessionID)
	assert.Equal(t, testRoomCode, resp.RoomCode)
	assert.Equal(t, "coop", resp.Mode)
	assert.Equal(t, "playing", resp.Status)
	assert.Equal(t, 2, resp.PlayerCount)
	assert.Equal(t, testWsBaseURL+"/game/"+testSessionID, resp.WebSocketURL)
	require.Len(t, resp.Players, 2)
	assert.Equal(t, "10", resp.Players[0].UserID)
	assert.Equal(t, "player1", resp.Players[0].Nickname)
	assert.Equal(t, "20", resp.Players[1].UserID)
	assert.Equal(t, "player2", resp.Players[1].Nickname)
	mockPlayer.AssertNotCalled(t, "GetPlayerByUserID")
}

func TestGameUseCase_GetSessionInfo_NotFound(t *testing.T) {
	t.Log("GameUseCase.GetSessionInfo: session not found -> error")
	mockSession := new(MockGameSessionRepository)
	mockPlayer := new(MockPlayerRepository)
	uc := newTestGameUseCase(mockSession, mockPlayer)

	repoErr := errors.New("session not found")
	mockSession.On("GetSessionByID", "missing-id").Return(nil, repoErr)

	resp, err := uc.GetSessionInfo("missing-id")

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, repoErr, err)
}
