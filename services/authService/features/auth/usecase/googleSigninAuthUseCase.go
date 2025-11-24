package usecase

import (
	"context"
	"fmt"
	"os"
	"time"

	_interface "github.com/JokerTrickster/joker_backend/services/authService/features/auth/model/interface"
	"github.com/JokerTrickster/joker_backend/services/authService/features/auth/model/response"
	"github.com/JokerTrickster/joker_backend/shared/errors"
	"github.com/JokerTrickster/joker_backend/shared/jwt"
	"google.golang.org/api/idtoken"
)

type GoogleSigninAuthUseCase struct {
	Repository     _interface.IGoogleSigninAuthRepository
	ContextTimeout time.Duration
	GoogleClientID string
}

func NewGoogleSigninAuthUseCase(repo _interface.IGoogleSigninAuthRepository, timeout time.Duration) _interface.IGoogleSigninAuthUseCase {
	// 환경변수에서 구글 클라이언트 ID 가져오기 (없으면 빈 문자열 - 자동 검증)
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	return &GoogleSigninAuthUseCase{
		Repository:     repo,
		ContextTimeout: timeout,
		GoogleClientID: clientID,
	}
}

func (d *GoogleSigninAuthUseCase) GoogleSignin(c context.Context, idToken string) (response.ResGoogleSignin, error) {
	ctx, cancel := context.WithTimeout(c, d.ContextTimeout)
	defer cancel()

	// 구글 ID 토큰 검증
	// 클라이언트 ID가 설정되어 있으면 사용, 없으면 빈 문자열로 자동 검증
	payload, err := idtoken.Validate(ctx, idToken, d.GoogleClientID)
	if err != nil {
		return response.ResGoogleSignin{}, errors.Unauthorized("Invalid Google ID token")
	}

	// 토큰에서 이메일과 이름 추출
	email, ok := payload.Claims["email"].(string)
	if !ok || email == "" {
		return response.ResGoogleSignin{}, errors.BadRequest("Email not found in token")
	}

	name, _ := payload.Claims["name"].(string)
	if name == "" {
		// 이름이 없으면 이메일의 @ 앞부분을 이름으로 사용
		name = email
	}

	// 유저 찾기 또는 생성
	userID, err := d.Repository.FindOrCreateUserByGoogleEmail(ctx, email, name)
	if err != nil {
		return response.ResGoogleSignin{}, errors.InternalServerError("Failed to find or create user")
	}

	// JWT 토큰 발급
	accessToken, _, refreshToken, _, err := jwt.GenerateToken(email, userID)
	if err != nil {
		return response.ResGoogleSignin{}, fmt.Errorf("failed to generate tokens: %w", err)
	}

	res := response.ResGoogleSignin{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	return res, nil
}
