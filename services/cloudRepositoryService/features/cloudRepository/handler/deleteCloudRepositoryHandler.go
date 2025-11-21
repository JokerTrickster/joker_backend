package handler

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

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
func (h *CloudRepositoryHandler) DeleteFile(c echo.Context) error {
	ctx := c.Request().Context()
	
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	fileID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid file ID"})
	}

	if err := h.usecase.DeleteFile(ctx, userID, uint(fileID)); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "file deleted successfully"})
}
