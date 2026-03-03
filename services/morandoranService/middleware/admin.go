package middleware

import (
	"github.com/JokerTrickster/joker_backend/shared/utils"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func AdminAuth(db *gorm.DB) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userID, err := utils.GetUserIDFromContext(c)
			if err != nil {
				return echo.NewHTTPError(401, "UNAUTHORIZED")
			}

			var role string
			result := db.Raw("SELECT role FROM morandoran_users WHERE id = ? AND deleted_at IS NULL", userID).Scan(&role)
			if result.Error != nil || role != "admin" {
				return echo.NewHTTPError(403, "FORBIDDEN")
			}

			return next(c)
		}
	}
}
