package handler

import (
	"net/http"

	_interface "github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/interface"
	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/request"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

type ActivityHistoryCloudRepositoryHandler struct {
	uc _interface.IActivityHistoryCloudRepositoryUseCase
}

func NewActivityHistoryCloudRepositoryHandler(g *echo.Group, uc _interface.IActivityHistoryCloudRepositoryUseCase) {
	handler := &ActivityHistoryCloudRepositoryHandler{uc: uc}
	g.GET("/user/activity", handler.GetActivityHistory)
}

// GetActivityHistory godoc
// @Summary Get user activity history
// @Description Retrieves daily activity for a specific month
// @Tags User
// @Accept json
// @Produce json
// @Param month query string false "Month in YYYY-MM format (defaults to current month)"
// @Success 200 {object} response.ActivityHistoryResponseDTO
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/user/activity [get]
// @Security Bearer
func (h *ActivityHistoryCloudRepositoryHandler) GetActivityHistory(c echo.Context) error {
	// Get userID from context
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
	}

	// Bind query parameters
	var req request.ActivityHistoryRequestDTO
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	// Get activity history
	history, err := h.uc.GetActivityHistory(c.Request().Context(), userID, &req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": errors.Wrap(err, "failed to get activity history").Error(),
		})
	}

	return c.JSON(http.StatusOK, history)
}