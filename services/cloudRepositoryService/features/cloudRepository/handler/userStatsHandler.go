package handler

import (
	"net/http"

	_interface "github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/interface"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

type UserStatsCloudRepositoryHandler struct {
	uc _interface.IUserStatsCloudRepositoryUseCase
}

func NewUserStatsCloudRepositoryHandler(g *echo.Group, uc _interface.IUserStatsCloudRepositoryUseCase) {
	handler := &UserStatsCloudRepositoryHandler{uc: uc}
	g.GET("/user/stats", handler.GetUserStats)
}

// GetUserStats godoc
// @Summary Get user statistics
// @Description Retrieves storage usage and monthly activity statistics for the authenticated user
// @Tags User
// @Accept json
// @Produce json
// @Success 200 {object} response.UserStatsResponseDTO
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/user/stats [get]
// @Security Bearer
func (h *UserStatsCloudRepositoryHandler) GetUserStats(c echo.Context) error {
	// Get userID from context
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
	}

	// Get user stats
	stats, err := h.uc.GetUserStats(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": errors.Wrap(err, "failed to get user stats").Error(),
		})
	}

	return c.JSON(http.StatusOK, stats)
}