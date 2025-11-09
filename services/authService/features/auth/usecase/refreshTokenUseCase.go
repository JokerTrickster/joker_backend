package usecase

import (
	"context"
	"fmt"
	_interface "github.com/JokerTrickster/joker_backend/services/authService/features/auth/model/interface"
	"github.com/JokerTrickster/joker_backend/services/authService/features/auth/model/request"
	"github.com/JokerTrickster/joker_backend/services/authService/features/auth/model/response"
	"github.com/JokerTrickster/joker_backend/shared/jwt"
	"time"
)

type RefreshTokenUseCase struct {
	ContextTimeout time.Duration
}

func NewRefreshTokenUseCase(timeout time.Duration) _interface.IRefreshTokenUseCase {
	return &RefreshTokenUseCase{ContextTimeout: timeout}
}

func (d *RefreshTokenUseCase) RefreshToken(c context.Context, req *request.ReqRefreshToken) (response.ResRefreshToken, error) {
	ctx, cancel := context.WithTimeout(c, d.ContextTimeout)
	defer cancel()
	_ = ctx

	// 리프레시 토큰 검증
	userID, email, err := jwt.VerifyRefreshToken(req.RefreshToken)
	if err != nil {
		return response.ResRefreshToken{}, fmt.Errorf("invalid or expired refresh token: %w", err)
	}

	// 새로운 액세스 토큰과 리프레시 토큰 발급
	accessToken, _, refreshToken, _, err := jwt.GenerateToken(email, userID)
	if err != nil {
		return response.ResRefreshToken{}, fmt.Errorf("failed to generate tokens: %w", err)
	}

	res := response.ResRefreshToken{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	return res, nil
}
