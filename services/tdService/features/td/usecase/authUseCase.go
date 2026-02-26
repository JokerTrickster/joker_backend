package usecase

import (
	"fmt"
	"strconv"
	"time"

	"github.com/JokerTrickster/joker_backend/services/tdService/features/td/model/entity"
	"github.com/JokerTrickster/joker_backend/services/tdService/features/td/model/request"
	"github.com/JokerTrickster/joker_backend/services/tdService/features/td/model/response"
	"github.com/JokerTrickster/joker_backend/services/tdService/features/td/repository"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthUseCase struct {
	playerRepo *repository.PlayerRepository
	jwtSecret  string
}

func NewAuthUseCase(playerRepo *repository.PlayerRepository, jwtSecret string) *AuthUseCase {
	return &AuthUseCase{
		playerRepo: playerRepo,
		jwtSecret:  jwtSecret,
	}
}

// Login handles user login
func (uc *AuthUseCase) Login(req *request.LoginRequest) (*response.LoginResponse, error) {
	// Note: In a real implementation, you would verify against a user table
	// For this TD service, we'll create/get player profile based on username
	// This is a simplified implementation for the Unity game

	// Generate a mock user ID from username (in production, verify against user table)
	userID := uc.generateUserID(req.Username)

	// Get or create player profile
	player, err := uc.playerRepo.GetOrCreatePlayer(userID, req.Username, "default_avatar")
	if err != nil {
		return nil, fmt.Errorf("failed to get player profile: %w", err)
	}

	// Generate JWT token
	token, err := uc.generateToken(userID, req.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Prepare response
	resp := &response.LoginResponse{
		Token:  token,
		UserID: strconv.FormatUint(uint64(userID), 10),
		Profile: response.ProfileResponse{
			Nickname:   player.Nickname,
			AvatarID:   player.AvatarID,
			Level:      player.Level,
			Experience: player.Experience,
		},
	}

	// Add stats if available
	if player.Stats.ID != 0 {
		resp.Profile.Stats = &response.StatsResponse{
			GamesPlayed:     player.Stats.GamesPlayed,
			Victories:       player.Stats.Victories,
			TotalScore:      player.Stats.TotalScore,
			HighestScore:    player.Stats.HighestScore,
			HighestWave:     player.Stats.HighestWave,
			UnitsPlaced:     player.Stats.UnitsPlaced,
			EnemiesDefeated: player.Stats.EnemiesKilled,
		}
	}

	return resp, nil
}

// RefreshToken refreshes JWT token
func (uc *AuthUseCase) RefreshToken(req *request.RefreshTokenRequest) (*response.RefreshTokenResponse, error) {
	// Parse and validate the refresh token
	claims, err := uc.validateToken(req.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	userIDStr, ok := claims["userID"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID in token")
	}

	username, ok := claims["username"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid username in token")
	}

	// Generate new token
	token, err := uc.generateToken(uint(userID), username)
	if err != nil {
		return nil, fmt.Errorf("failed to generate new token: %w", err)
	}

	return &response.RefreshTokenResponse{
		Token: token,
	}, nil
}

// generateToken generates a JWT token
func (uc *AuthUseCase) generateToken(userID uint, username string) (string, error) {
	claims := jwt.MapClaims{
		"userID":   strconv.FormatUint(uint64(userID), 10),
		"username": username,
		"exp":      time.Now().Add(24 * time.Hour).Unix(), // Token expires in 24 hours
		"iat":      time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(uc.jwtSecret))
}

// validateToken validates and parses a JWT token
func (uc *AuthUseCase) validateToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(uc.jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

// generateUserID generates a consistent user ID from username
func (uc *AuthUseCase) generateUserID(username string) uint {
	// Simple hash-based ID generation (in production, use actual user table)
	hash, _ := bcrypt.GenerateFromPassword([]byte(username), 1)
	var id uint
	for _, b := range hash[:4] {
		id = id*256 + uint(b)
	}
	return id % 1000000 // Keep it within reasonable range
}

// ValidateTokenAndGetUserID validates token and returns user ID
func (uc *AuthUseCase) ValidateTokenAndGetUserID(tokenString string) (uint, error) {
	claims, err := uc.validateToken(tokenString)
	if err != nil {
		return 0, err
	}

	userIDStr, ok := claims["userID"].(string)
	if !ok {
		return 0, fmt.Errorf("invalid token claims")
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid user ID in token")
	}

	return uint(userID), nil
}