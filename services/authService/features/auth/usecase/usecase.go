package usecase

import (
	"time"

	"github.com/JokerTrickster/joker_backend/shared/db/mysql"
)

func createTokenDTO(userID uint, accessToken string, refreshToken string) *mysql.Tokens {
	return &mysql.Tokens{
		UserID:       userID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		// 예시로 7일 후 만료 설정
		RefreshExpiredAt: time.Now().Add(7 * 24 * time.Hour).Unix(),
	}
}
