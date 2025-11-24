package handler

import (
	"net/http"
	"strconv"

	_interface "github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/interface"
	"github.com/labstack/echo/v4"
)

type DeleteCloudRepositoryHandler struct {
	UseCase _interface.IDeleteCloudRepositoryUseCase
}

func NewDeleteCloudRepositoryHandler(c *echo.Group, useCase _interface.IDeleteCloudRepositoryUseCase) _interface.IDeleteCloudRepositoryHandler {
	handler := &DeleteCloudRepositoryHandler{
		UseCase: useCase,
	}
	c.DELETE("/files/:id", handler.DeleteFile)
	return handler
}

// DeleteFile handles file deletion
// @Summary Delete file
// @Description Soft delete a file and remove it from S3
// @Tags CloudRepository
// @Accept json
// @Produce json
// @Param id path int true "File ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/files/{id} [delete]
func (h *DeleteCloudRepositoryHandler) DeleteFile(c echo.Context) error {
	ctx := c.Request().Context()

	userID, err := getUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	fileID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid file ID"})
	}

	if err := h.UseCase.DeleteFile(ctx, userID, uint(fileID)); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "file deleted successfully"})
}
