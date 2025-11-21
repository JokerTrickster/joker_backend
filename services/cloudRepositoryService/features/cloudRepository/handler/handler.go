package handler

import (
	"net/http"

	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/usecase"
	"github.com/labstack/echo/v4"
)

type CloudRepositoryHandler struct {
	usecase *usecase.CloudRepositoryUsecase
}

func NewCloudRepositoryHandler(uc *usecase.CloudRepositoryUsecase) *CloudRepositoryHandler {
	return &CloudRepositoryHandler{
		usecase: uc,
	}
}

// getUserIDFromContext extracts user ID from context (set by JWT middleware)
func getUserIDFromContext(c echo.Context) (uint, error) {
	// This is a placeholder - implement based on your JWT middleware
	// Usually something like:
	// claims := c.Get("user").(*jwt.Token).Claims.(jwt.MapClaims)
	// userID := uint(claims["user_id"].(float64))

	// For now, return a dummy value - replace with actual implementation
	userIDStr := c.Get("userID")
	if userIDStr == nil {
		return 0, echo.NewHTTPError(http.StatusUnauthorized, "user not found in context")
	}

	userID, ok := userIDStr.(uint)
	if !ok {
		return 0, echo.NewHTTPError(http.StatusUnauthorized, "invalid user ID type")
	}

	return userID, nil
}
