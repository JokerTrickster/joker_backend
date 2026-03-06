package repository

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/towerDefense/model/entity"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func setupQuestTestDB(t *testing.T) *gorm.DB {
	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		dsn = "root:rootpassword@tcp(localhost:3307)/test_db?charset=utf8mb4&parseTime=True&loc=Local"
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skipf("Skipping: test database unavailable: %v", err)
	}
	if err := db.AutoMigrate(&entity.TDUser{}, &entity.TDUserStats{}, &entity.TDQuest{}); err != nil {
		t.Skipf("Skipping: migration failed: %v", err)
	}
	return db
}

func createQuestTestUser(t *testing.T, db *gorm.DB) *entity.TDUser {
	unique := fmt.Sprintf("%d_%d", time.Now().UnixNano(), rand.Intn(100000))
	user := &entity.TDUser{
		Username:     "questuser_" + unique,
		Email:        "quest_" + unique + "@test.com",
		PasswordHash: "hash",
		IsActive:     true,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(user).Error)
	return user
}

func TestTDQuestRepository_Create(t *testing.T) {
	db := setupQuestTestDB(t)
	requireTable(t, db, "td_users")
	requireTable(t, db, "td_quests")
	ctx := context.Background()
	repo := NewTDQuestRepository(db)
	user := createQuestTestUser(t, db)

	quest := &entity.TDQuest{
		UserID:           user.ID,
		QuestType:        "kill",
		QuestName:        "Kill 10 monsters",
		QuestDescription: "Kill 10 monsters in a game",
		TargetCount:      10,
		CurrentCount:     0,
		RewardGold:       50,
		Status:           "active",
	}
	err := repo.Create(ctx, quest)
	require.NoError(t, err)
	require.NotZero(t, quest.ID)
}

func TestTDQuestRepository_GetByID(t *testing.T) {
	db := setupQuestTestDB(t)
	requireTable(t, db, "td_users")
	requireTable(t, db, "td_quests")
	ctx := context.Background()
	repo := NewTDQuestRepository(db)
	user := createQuestTestUser(t, db)

	quest := &entity.TDQuest{
		UserID:     user.ID,
		QuestType:  "kill",
		QuestName:  "Test Quest",
		TargetCount: 5,
		Status:    "active",
	}
	require.NoError(t, repo.Create(ctx, quest))

	got, err := repo.GetByID(ctx, quest.ID)
	require.NoError(t, err)
	require.Equal(t, quest.ID, got.ID)
	require.Equal(t, "Test Quest", got.QuestName)
}

func TestTDQuestRepository_GetActiveQuests(t *testing.T) {
	db := setupQuestTestDB(t)
	requireTable(t, db, "td_users")
	requireTable(t, db, "td_quests")
	ctx := context.Background()
	repo := NewTDQuestRepository(db)
	user := createQuestTestUser(t, db)

	q1 := &entity.TDQuest{UserID: user.ID, QuestType: "kill", QuestName: "Q1", TargetCount: 10, Status: "active"}
	q2 := &entity.TDQuest{UserID: user.ID, QuestType: "kill", QuestName: "Q2", TargetCount: 5, Status: "active"}
	require.NoError(t, repo.Create(ctx, q1))
	require.NoError(t, repo.Create(ctx, q2))

	quests, err := repo.GetActiveQuests(ctx, user.ID)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(quests), 2)
}

func TestTDQuestRepository_UpdateProgress(t *testing.T) {
	db := setupQuestTestDB(t)
	requireTable(t, db, "td_users")
	requireTable(t, db, "td_quests")
	ctx := context.Background()
	repo := NewTDQuestRepository(db)
	user := createQuestTestUser(t, db)

	quest := &entity.TDQuest{
		UserID:     user.ID,
		QuestType:  "kill",
		QuestName:  "Progress Quest",
		TargetCount: 10,
		CurrentCount: 3,
		Status:    "active",
	}
	require.NoError(t, repo.Create(ctx, quest))

	err := repo.UpdateProgress(ctx, quest.ID, 5)
	require.NoError(t, err)

	got, _ := repo.GetByID(ctx, quest.ID)
	require.Equal(t, uint(8), got.CurrentCount)
}

func TestTDQuestRepository_CompleteQuest(t *testing.T) {
	db := setupQuestTestDB(t)
	requireTable(t, db, "td_users")
	requireTable(t, db, "td_quests")
	ctx := context.Background()
	repo := NewTDQuestRepository(db)
	user := createQuestTestUser(t, db)

	quest := &entity.TDQuest{
		UserID:     user.ID,
		QuestType:  "kill",
		QuestName:  "Complete Quest",
		TargetCount: 10,
		CurrentCount: 10,
		Status:    "active",
	}
	require.NoError(t, repo.Create(ctx, quest))

	err := repo.CompleteQuest(ctx, quest.ID)
	require.NoError(t, err)

	got, _ := repo.GetByID(ctx, quest.ID)
	require.Equal(t, "completed", got.Status)
}

func TestTDQuestRepository_ClaimQuest(t *testing.T) {
	db := setupQuestTestDB(t)
	requireTable(t, db, "td_users")
	requireTable(t, db, "td_quests")
	ctx := context.Background()
	repo := NewTDQuestRepository(db)
	user := createQuestTestUser(t, db)

	quest := &entity.TDQuest{
		UserID:     user.ID,
		QuestType:  "kill",
		QuestName:  "Claim Quest",
		TargetCount: 10,
		CurrentCount: 10,
		Status:    "completed",
	}
	require.NoError(t, repo.Create(ctx, quest))

	err := repo.ClaimQuest(ctx, quest.ID)
	require.NoError(t, err)

	got, _ := repo.GetByID(ctx, quest.ID)
	require.Equal(t, "claimed", got.Status)
	require.NotNil(t, got.ClaimedAt)
}
