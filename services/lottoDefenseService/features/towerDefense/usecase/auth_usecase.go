package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/towerDefense/model/entity"
	_interface "github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/towerDefense/model/interface"
	"github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/towerDefense/model/request"
	"github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/towerDefense/model/response"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type TDAuthUseCase struct {
	userRepo  _interface.ITDUserRepository
	jwtSecret string
	timeout   time.Duration
}

func NewTDAuthUseCase(userRepo _interface.ITDUserRepository, jwtSecret string, timeout time.Duration) _interface.ITDAuthUseCase {
	return &TDAuthUseCase{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
		timeout:   timeout,
	}
}

func (u *TDAuthUseCase) Register(ctx context.Context, req *request.RegisterRequest) (*response.AuthResponse, error) {
	// Check if user exists
	existing, _ := u.userRepo.GetByEmail(ctx, req.Email)
	if existing != nil {
		return nil, errors.New("email already exists")
	}

	existing, _ = u.userRepo.GetByUsername(ctx, req.Username)
	if existing != nil {
		return nil, errors.New("username already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Create user
	user := &entity.TDUser{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		IsActive:     true,
	}

	if err := u.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// Create initial stats
	stats := &entity.TDUserStats{
		UserID: user.ID,
	}
	if err := u.userRepo.CreateStats(ctx, stats); err != nil {
		return nil, err
	}

	// Generate JWT
	token, err := u.generateJWT(user.ID)
	if err != nil {
		return nil, err
	}

	return &response.AuthResponse{
		User:  userToUserData(user),
		Token: token,
	}, nil
}

func (u *TDAuthUseCase) Login(ctx context.Context, req *request.LoginRequest) (*response.AuthResponse, error) {
	user, err := u.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid credentials")
		}
		return nil, err
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Update last login
	if err := u.userRepo.UpdateLastLogin(ctx, user.ID); err != nil {
		// Log but don't fail
	}

	// Generate JWT
	token, err := u.generateJWT(user.ID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	user.LastLogin = &now

	return &response.AuthResponse{
		User:  userToUserData(user),
		Token: token,
	}, nil
}

func (u *TDAuthUseCase) GetUserInfo(ctx context.Context, userID uint) (*response.UserInfoResponse, error) {
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	stats, err := u.userRepo.GetStats(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &response.UserInfoResponse{
		User:  userToUserData(user),
		Stats: statsToStatsData(stats),
	}, nil
}

func (u *TDAuthUseCase) generateJWT(userID uint) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(u.jwtSecret))
}

func userToUserData(user *entity.TDUser) *response.UserData {
	return &response.UserData{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		LastLogin: user.LastLogin,
	}
}

func statsToStatsData(stats *entity.TDUserStats) *response.UserStatsData {
	return &response.UserStatsData{
		SingleHighestRound: stats.SingleHighestRound,
		SingleTotalGames:   stats.SingleTotalGames,
		SingleTotalKills:   stats.SingleTotalKills,
		CoopHighestRound:   stats.CoopHighestRound,
		CoopTotalGames:     stats.CoopTotalGames,
		CoopTotalKills:     stats.CoopTotalKills,
		CoopWins:           stats.CoopWins,
		TotalGoldEarned:    stats.TotalGoldEarned,
		CurrentGold:        stats.CurrentGold,
		QuestsCompleted:    stats.QuestsCompleted,
	}
}
