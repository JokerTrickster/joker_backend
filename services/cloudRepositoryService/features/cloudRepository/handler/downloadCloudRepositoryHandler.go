package handler

import (
	"net/http"
	"strconv"

	_ "github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model"
	"github.com/labstack/echo/v4"
)

// RequestDownloadURL handles the request for a presigned download URL
// @Summary Request presigned download URL
// @Description Get a presigned URL for downloading a file from S3
// @Tags CloudRepository
// @Accept json
// @Produce json
// @Param id path int true "File ID"
// @Success 200 {object} model.DownloadResponseDTO
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/files/{id}/download [get]
func (h *CloudRepositoryHandler) RequestDownloadURL(c echo.Context) error {
	ctx := c.Request().Context()

	userID, err := getUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	fileID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid file ID"})
	}

	resp, err := h.usecase.RequestDownloadURL(ctx, userID, uint(fileID))
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, resp)
}
