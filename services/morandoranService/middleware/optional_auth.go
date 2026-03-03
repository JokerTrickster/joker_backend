package middleware

import (
	"strings"

	"github.com/JokerTrickster/joker_backend/shared/jwt"
	"github.com/labstack/echo/v4"
)

func OptionalAuth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return next(c)
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				return next(c)
			}

			tokenString := parts[1]
			if err := jwt.VerifyToken(tokenString); err != nil {
				return next(c)
			}

			userID, email, err := jwt.ParseToken(tokenString)
			if err != nil {
				return next(c)
			}

			c.Set("userID", userID)
			c.Set("email", email)
			return next(c)
		}
	}
}
