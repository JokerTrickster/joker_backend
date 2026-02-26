package handler

import (
	"github.com/JokerTrickster/joker_backend/shared/utils"
	"github.com/labstack/echo/v4"
)

func getUserIDFromContext(c echo.Context) (uint, error) {
	return utils.GetUserIDFromContext(c)
}
