package repository

import (
	"testing"

	"github.com/JokerTrickster/joker_backend/services/tdService/features/td/model/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func createTestPlayerForSession(t *testing.T, db *gorm.DB, userID uint) *entity.Player {
	t.Helper()
	suffix := uniqueSuffix()
	player := &entity.Player{
		UserID:   userID,
		Nickname: "player_" + suffix,
		AvatarID: "avatar_" + suffix,
	}
	require.NoError(t, db.Create(player).Error)
	return player
}

func deferCleanupSessionTest(t *testing.T, db *gorm.DB, playerIDs []uint, sessionIDs []string) {
	t.Helper()
	t.Cleanup(func() {
		for _, sid := range sessionIDs {
			db.Exec("DELETE FROM game_results WHERE session_id = ?", sid)
			db.Where("session_id = ?", sid).Delete(&entity.GameSession{})
		}
		for _, pid := range playerIDs {
			db.Where("player_id = ?", pid).Delete(&entity.PlayerStats{})
			db.Delete(&entity.Player{}, pid)
		}
	})
}

func TestGameSessionRepository_CreateSession_Single(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGameSessionRepository(db)
	player := createTestPlayerForSession(t, db, uniqueUserID())

	session := &entity.GameSession{
		Mode:        entity.GameModeSingle,
		PlayerCount: 1,
		PlayerID:    player.ID,
	}

	err := repo.CreateSession(session)
	require.NoError(t, err)
	deferCleanupSessionTest(t, db, []uint{player.ID}, []string{session.SessionID})
	assert.NotEmpty(t, session.SessionID, "SessionID should be generated")
	assert.Equal(t, entity.GameStatusWaiting, session.Status)
	assert.Empty(t, session.RoomCode, "Single mode should not have room code")

	t.Logf("CreateSession single mode: SessionID=%s", session.SessionID)
}

func TestGameSessionRepository_CreateSession_Coop(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGameSessionRepository(db)
	player := createTestPlayerForSession(t, db, uniqueUserID())

	session := &entity.GameSession{
		Mode:        entity.GameModeCoop,
		PlayerCount: 1,
		PlayerID:    player.ID,
	}

	err := repo.CreateSession(session)
	require.NoError(t, err)
	deferCleanupSessionTest(t, db, []uint{player.ID}, []string{session.SessionID})
	assert.NotEmpty(t, session.SessionID)
	assert.NotEmpty(t, session.RoomCode, "Coop mode should generate room code")
	assert.Len(t, session.RoomCode, 6, "Room code should be 6 characters")
	assert.Equal(t, entity.GameStatusWaiting, session.Status)

	t.Logf("CreateSession coop mode: SessionID=%s, RoomCode=%s", session.SessionID, session.RoomCode)
}

func TestGameSessionRepository_GetSessionByID_Found(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGameSessionRepository(db)
	player := createTestPlayerForSession(t, db, uniqueUserID())

	session := &entity.GameSession{
		Mode:        entity.GameModeSingle,
		PlayerCount: 1,
		PlayerID:    player.ID,
	}
	require.NoError(t, repo.CreateSession(session))
	deferCleanupSessionTest(t, db, []uint{player.ID}, []string{session.SessionID})

	got, err := repo.GetSessionByID(session.SessionID)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, session.SessionID, got.SessionID)
	assert.Equal(t, entity.GameModeSingle, got.Mode)
	assert.NotNil(t, got.Player)
	assert.Equal(t, player.ID, got.Player.ID)

	t.Logf("GetSessionByID found session with preloaded Player")
}

func TestGameSessionRepository_GetSessionByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGameSessionRepository(db)

	got, err := repo.GetSessionByID("non-existent-" + uniqueSuffix())
	require.Error(t, err)
	assert.Nil(t, got)
	assert.Contains(t, err.Error(), "session not found")

	t.Log("GetSessionByID correctly returns error for non-existent session")
}

func TestGameSessionRepository_GetSessionByRoomCode_Found(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGameSessionRepository(db)
	player := createTestPlayerForSession(t, db, uniqueUserID())

	session := &entity.GameSession{
		Mode:        entity.GameModeCoop,
		PlayerCount: 1,
		PlayerID:    player.ID,
	}
	require.NoError(t, repo.CreateSession(session))
	deferCleanupSessionTest(t, db, []uint{player.ID}, []string{session.SessionID})

	got, err := repo.GetSessionByRoomCode(session.RoomCode)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, session.SessionID, got.SessionID)
	assert.Equal(t, session.RoomCode, got.RoomCode)

	t.Logf("GetSessionByRoomCode found session by RoomCode=%s", session.RoomCode)
}

func TestGameSessionRepository_GetSessionByRoomCode_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGameSessionRepository(db)

	got, err := repo.GetSessionByRoomCode("XX" + uniqueSuffix())
	require.Error(t, err)
	assert.Nil(t, got)
	assert.Contains(t, err.Error(), "session not found")

	t.Log("GetSessionByRoomCode correctly returns error for invalid room code")
}

func TestGameSessionRepository_JoinSession_Success(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGameSessionRepository(db)
	player1 := createTestPlayerForSession(t, db, uniqueUserID())
	player2 := createTestPlayerForSession(t, db, uniqueUserID())

	session := &entity.GameSession{
		Mode:        entity.GameModeCoop,
		PlayerCount: 1,
		PlayerID:    player1.ID,
	}
	require.NoError(t, repo.CreateSession(session))
	deferCleanupSessionTest(t, db, []uint{player1.ID, player2.ID}, []string{session.SessionID})

	err := repo.JoinSession(session.SessionID, player2.ID)
	require.NoError(t, err)

	got, err := repo.GetSessionByID(session.SessionID)
	require.NoError(t, err)
	require.NotNil(t, got.Player2ID)
	assert.Equal(t, player2.ID, *got.Player2ID)

	t.Log("JoinSession successfully set Player2ID")
}

func TestGameSessionRepository_JoinSession_NotCoop(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGameSessionRepository(db)
	player1 := createTestPlayerForSession(t, db, uniqueUserID())
	player2 := createTestPlayerForSession(t, db, uniqueUserID())

	sid := "single-" + uniqueSuffix()
	session := &entity.GameSession{
		SessionID:   sid,
		Mode:        entity.GameModeSingle,
		PlayerCount: 1,
		PlayerID:    player1.ID,
	}
	require.NoError(t, db.Create(session).Error)
	deferCleanupSessionTest(t, db, []uint{player1.ID, player2.ID}, []string{sid})

	err := repo.JoinSession(session.SessionID, player2.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not in coop mode")

	t.Log("JoinSession correctly rejects single-mode session")
}

func TestGameSessionRepository_JoinSession_AlreadyFull(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGameSessionRepository(db)
	player1 := createTestPlayerForSession(t, db, uniqueUserID())
	player2 := createTestPlayerForSession(t, db, uniqueUserID())
	player3 := createTestPlayerForSession(t, db, uniqueUserID())

	session := &entity.GameSession{
		Mode:        entity.GameModeCoop,
		PlayerCount: 1,
		PlayerID:    player1.ID,
	}
	require.NoError(t, repo.CreateSession(session))
	require.NoError(t, repo.JoinSession(session.SessionID, player2.ID))
	deferCleanupSessionTest(t, db, []uint{player1.ID, player2.ID, player3.ID}, []string{session.SessionID})

	err := repo.JoinSession(session.SessionID, player3.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already full")

	t.Log("JoinSession correctly rejects session that already has Player2")
}

func TestGameSessionRepository_JoinSession_AlreadyStarted(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGameSessionRepository(db)
	player1 := createTestPlayerForSession(t, db, uniqueUserID())
	player2 := createTestPlayerForSession(t, db, uniqueUserID())

	session := &entity.GameSession{
		Mode:        entity.GameModeCoop,
		PlayerCount: 1,
		PlayerID:    player1.ID,
	}
	require.NoError(t, repo.CreateSession(session))
	require.NoError(t, repo.StartSession(session.SessionID))
	deferCleanupSessionTest(t, db, []uint{player1.ID, player2.ID}, []string{session.SessionID})

	err := repo.JoinSession(session.SessionID, player2.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already started")

	t.Log("JoinSession correctly rejects session that has already started")
}

func TestGameSessionRepository_StartSession(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGameSessionRepository(db)
	player := createTestPlayerForSession(t, db, uniqueUserID())

	session := &entity.GameSession{
		Mode:        entity.GameModeSingle,
		PlayerCount: 1,
		PlayerID:    player.ID,
	}
	require.NoError(t, repo.CreateSession(session))
	deferCleanupSessionTest(t, db, []uint{player.ID}, []string{session.SessionID})

	err := repo.StartSession(session.SessionID)
	require.NoError(t, err)

	got, err := repo.GetSessionByID(session.SessionID)
	require.NoError(t, err)
	assert.Equal(t, entity.GameStatusPlaying, got.Status)
	assert.NotNil(t, got.StartedAt)

	t.Log("StartSession set status=playing and StartedAt")
}

func TestGameSessionRepository_EndSession(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGameSessionRepository(db)
	player := createTestPlayerForSession(t, db, uniqueUserID())

	session := &entity.GameSession{
		Mode:        entity.GameModeSingle,
		PlayerCount: 1,
		PlayerID:    player.ID,
	}
	require.NoError(t, repo.CreateSession(session))
	require.NoError(t, repo.StartSession(session.SessionID))
	deferCleanupSessionTest(t, db, []uint{player.ID}, []string{session.SessionID})

	err := repo.EndSession(session.SessionID)
	require.NoError(t, err)

	got, err := repo.GetSessionByID(session.SessionID)
	require.NoError(t, err)
	assert.Equal(t, entity.GameStatusFinished, got.Status)
	assert.NotNil(t, got.EndedAt)

	t.Log("EndSession set status=finished and EndedAt")
}

func TestGameSessionRepository_SaveGameResult(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGameSessionRepository(db)
	player := createTestPlayerForSession(t, db, uniqueUserID())

	session := &entity.GameSession{
		Mode:        entity.GameModeSingle,
		PlayerCount: 1,
		PlayerID:    player.ID,
	}
	require.NoError(t, repo.CreateSession(session))
	deferCleanupSessionTest(t, db, []uint{player.ID}, []string{session.SessionID})

	result := &entity.GameResult{
		SessionID:       session.SessionID,
		PlayerID:       player.ID,
		Score:          1500,
		WavesCompleted: 10,
		UnitsPlaced:    25,
		EnemiesDefeated: 100,
		Victory:        true,
		PlayTime:       600,
	}

	err := repo.SaveGameResult(result)
	require.NoError(t, err)
	assert.NotZero(t, result.ID)

	var count int64
	db.Model(&entity.GameResult{}).Where("session_id = ?", session.SessionID).Count(&count)
	assert.Equal(t, int64(1), count)

	t.Logf("SaveGameResult saved result ID=%d", result.ID)
}

func TestGameSessionRepository_GetGameResults(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGameSessionRepository(db)
	player := createTestPlayerForSession(t, db, uniqueUserID())

	session := &entity.GameSession{
		Mode:        entity.GameModeSingle,
		PlayerCount: 1,
		PlayerID:    player.ID,
	}
	require.NoError(t, repo.CreateSession(session))
	deferCleanupSessionTest(t, db, []uint{player.ID}, []string{session.SessionID})

	r1 := &entity.GameResult{
		SessionID:       session.SessionID,
		PlayerID:       player.ID,
		Score:          1000,
		WavesCompleted: 5,
		UnitsPlaced:    10,
		EnemiesDefeated: 50,
		Victory:        false,
		PlayTime:       300,
	}
	r2 := &entity.GameResult{
		SessionID:       session.SessionID,
		PlayerID:       player.ID,
		Score:          2000,
		WavesCompleted: 12,
		UnitsPlaced:    30,
		EnemiesDefeated: 120,
		Victory:        true,
		PlayTime:       900,
	}
	require.NoError(t, repo.SaveGameResult(r1))
	require.NoError(t, repo.SaveGameResult(r2))

	results, err := repo.GetGameResults(session.SessionID)
	require.NoError(t, err)
	assert.Len(t, results, 2)
	assert.NotNil(t, results[0].Player)

	t.Logf("GetGameResults returned %d results", len(results))
}

func TestGameSessionRepository_GetPlayerGameHistory(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGameSessionRepository(db)
	player := createTestPlayerForSession(t, db, uniqueUserID())

	session1 := &entity.GameSession{
		Mode:        entity.GameModeSingle,
		PlayerCount: 1,
		PlayerID:    player.ID,
	}
	session2 := &entity.GameSession{
		Mode:        entity.GameModeSingle,
		PlayerCount: 1,
		PlayerID:    player.ID,
	}
	require.NoError(t, repo.CreateSession(session1))
	require.NoError(t, repo.CreateSession(session2))
	deferCleanupSessionTest(t, db, []uint{player.ID}, []string{session1.SessionID, session2.SessionID})

	for i, sid := range []string{session1.SessionID, session2.SessionID} {
		r := &entity.GameResult{
			SessionID:       sid,
			PlayerID:       player.ID,
			Score:          int64(1000 * (i + 1)),
			WavesCompleted: 5 + i,
			UnitsPlaced:    10,
			EnemiesDefeated: 50,
			Victory:        i == 1,
			PlayTime:       300,
		}
		require.NoError(t, repo.SaveGameResult(r))
	}

	results, err := repo.GetPlayerGameHistory(player.ID, 1)
	require.NoError(t, err)
	assert.Len(t, results, 1, "Limit 1 should return only 1 result")
	assert.NotNil(t, results[0].Session)

	resultsAll, err := repo.GetPlayerGameHistory(player.ID, 0)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(resultsAll), 2)

	t.Logf("GetPlayerGameHistory: limit=1 returned 1, limit=0 returned %d", len(resultsAll))
}

func TestGameSessionRepository_UpdateSession(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGameSessionRepository(db)
	player := createTestPlayerForSession(t, db, uniqueUserID())

	session := &entity.GameSession{
		Mode:        entity.GameModeSingle,
		PlayerCount: 1,
		PlayerID:    player.ID,
	}
	require.NoError(t, repo.CreateSession(session))
	deferCleanupSessionTest(t, db, []uint{player.ID}, []string{session.SessionID})

	session.WebSocketURL = "wss://example.com/game"
	err := repo.UpdateSession(session)
	require.NoError(t, err)

	got, err := repo.GetSessionByID(session.SessionID)
	require.NoError(t, err)
	assert.Equal(t, "wss://example.com/game", got.WebSocketURL)
	t.Log("UpdateSession persisted WebSocketURL change")
}

func TestGameSessionRepository_GetActiveSessions(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGameSessionRepository(db)
	player := createTestPlayerForSession(t, db, uniqueUserID())

	session1 := &entity.GameSession{
		Mode:        entity.GameModeSingle,
		PlayerCount: 1,
		PlayerID:    player.ID,
	}
	session2 := &entity.GameSession{
		Mode:        entity.GameModeSingle,
		PlayerCount: 1,
		PlayerID:    player.ID,
	}
	require.NoError(t, repo.CreateSession(session1))
	require.NoError(t, repo.CreateSession(session2))
	require.NoError(t, repo.StartSession(session1.SessionID))
	deferCleanupSessionTest(t, db, []uint{player.ID}, []string{session1.SessionID, session2.SessionID})

	sessions, err := repo.GetActiveSessions(player.ID)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(sessions), 2, "Player should have waiting and playing sessions")
	statuses := make(map[entity.GameStatus]bool)
	for _, s := range sessions {
		statuses[s.Status] = true
	}
	assert.True(t, statuses[entity.GameStatusWaiting], "Should include waiting session")
	assert.True(t, statuses[entity.GameStatusPlaying], "Should include playing session")
	t.Logf("GetActiveSessions returned %d sessions", len(sessions))
}

func TestGameSessionRepository_GetSessionByRoomCode_AlreadyStarted(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGameSessionRepository(db)
	player := createTestPlayerForSession(t, db, uniqueUserID())

	session := &entity.GameSession{
		Mode:        entity.GameModeCoop,
		PlayerCount: 1,
		PlayerID:    player.ID,
	}
	require.NoError(t, repo.CreateSession(session))
	require.NoError(t, repo.StartSession(session.SessionID))
	deferCleanupSessionTest(t, db, []uint{player.ID}, []string{session.SessionID})

	got, err := repo.GetSessionByRoomCode(session.RoomCode)
	require.Error(t, err)
	assert.Nil(t, got)
	assert.Contains(t, err.Error(), "session not found or already started")
	t.Log("GetSessionByRoomCode correctly rejects started session")
}
