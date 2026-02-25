package repository

import (
	"context"

	"github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/towerDefense/model/entity"
	_interface "github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/towerDefense/model/interface"
	"gorm.io/gorm"
)

type TDRoomRepository struct {
	db *gorm.DB
}

func NewTDRoomRepository(db *gorm.DB) _interface.ITDRoomRepository {
	return &TDRoomRepository{db: db}
}

func (r *TDRoomRepository) Create(ctx context.Context, room *entity.TDRoom) error {
	return r.db.WithContext(ctx).Create(room).Error
}

func (r *TDRoomRepository) GetByID(ctx context.Context, id uint) (*entity.TDRoom, error) {
	var room entity.TDRoom
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&room).Error
	if err != nil {
		return nil, err
	}
	return &room, nil
}

func (r *TDRoomRepository) GetByCode(ctx context.Context, code string) (*entity.TDRoom, error) {
	var room entity.TDRoom
	err := r.db.WithContext(ctx).Where("room_code = ?", code).First(&room).Error
	if err != nil {
		return nil, err
	}
	return &room, nil
}

func (r *TDRoomRepository) Update(ctx context.Context, room *entity.TDRoom) error {
	return r.db.WithContext(ctx).Save(room).Error
}

func (r *TDRoomRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entity.TDRoom{}, id).Error
}

func (r *TDRoomRepository) AddPlayer(ctx context.Context, player *entity.TDRoomPlayer) error {
	return r.db.WithContext(ctx).Create(player).Error
}

func (r *TDRoomRepository) RemovePlayer(ctx context.Context, roomID, userID uint) error {
	return r.db.WithContext(ctx).
		Where("room_id = ? AND user_id = ?", roomID, userID).
		Delete(&entity.TDRoomPlayer{}).Error
}

func (r *TDRoomRepository) GetPlayers(ctx context.Context, roomID uint) ([]entity.TDRoomPlayer, error) {
	var players []entity.TDRoomPlayer
	err := r.db.WithContext(ctx).Where("room_id = ?", roomID).Find(&players).Error
	return players, err
}

func (r *TDRoomRepository) UpdatePlayerReady(ctx context.Context, roomID, userID uint, isReady bool) error {
	return r.db.WithContext(ctx).Model(&entity.TDRoomPlayer{}).
		Where("room_id = ? AND user_id = ?", roomID, userID).
		Update("is_ready", isReady).Error
}

func (r *TDRoomRepository) UpdatePlayerState(ctx context.Context, roomID, userID uint, stateJSON string) error {
	// Find which player position this user is in
	var room entity.TDRoom
	if err := r.db.WithContext(ctx).First(&room, roomID).Error; err != nil {
		return err
	}

	now := gorm.Expr("NOW()")
	
	// Update the appropriate player state based on user ID
	if room.HostUserID == userID {
		// User is host (player 1)
		return r.db.WithContext(ctx).Model(&entity.TDRoom{}).
			Where("id = ?", roomID).
			Updates(map[string]interface{}{
				"player1_state":   stateJSON,
				"last_p1_update":  now,
			}).Error
	}
	
	// User is player 2
	return r.db.WithContext(ctx).Model(&entity.TDRoom{}).
		Where("id = ?", roomID).
		Updates(map[string]interface{}{
			"player2_state":   stateJSON,
			"last_p2_update":  now,
		}).Error
}

func (r *TDRoomRepository) GetOpponentState(ctx context.Context, roomID, userID uint) (string, uint, string, error) {
	var room entity.TDRoom
	err := r.db.WithContext(ctx).
		Preload("HostUser").
		Preload("Players.User").
		First(&room, roomID).Error
	if err != nil {
		return "", 0, "", err
	}

	var opponentState string
	var opponentID uint
	var opponentName string

	if room.HostUserID == userID {
		// Current user is host, get player 2 state
		opponentState = room.Player2State
		if len(room.Players) > 0 {
			opponentID = room.Players[0].UserID
			if room.Players[0].User != nil {
				opponentName = room.Players[0].User.Name
			}
		}
	} else {
		// Current user is player 2, get host state
		opponentState = room.Player1State
		opponentID = room.HostUserID
		if room.HostUser != nil {
			opponentName = room.HostUser.Name
		}
	}

	return opponentState, opponentID, opponentName, nil
}
