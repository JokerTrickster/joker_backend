package handler

import (
	"github.com/JokerTrickster/joker_backend/services/authService/features/auth/repository"
	"github.com/JokerTrickster/joker_backend/services/authService/features/auth/usecase"

	"github.com/JokerTrickster/joker_backend/shared/db/mysql"

	"github.com/labstack/echo/v4"
)

func NewAuthHandler(c *echo.Echo) {
	NewSigninAuthHandler(c, usecase.NewSigninAuthUseCase(repository.NewSigninAuthRepository(mysql.GormMysqlDB), mysql.DBTimeOut))
	NewSignupAuthHandler(c, usecase.NewSignupAuthUseCase(repository.NewSignupAuthRepository(mysql.GormMysqlDB), mysql.DBTimeOut))
	NewRefreshTokenHandler(c, usecase.NewRefreshTokenUseCase(repository.NewRefreshTokenAuthRepository(mysql.GormMysqlDB), mysql.DBTimeOut))
	NewLogoutAuthHandler(c, usecase.NewLogoutAuthUseCase(repository.NewLogoutAuthRepository(mysql.GormMysqlDB), mysql.DBTimeOut))
}
