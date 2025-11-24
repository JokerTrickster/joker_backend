package handler

import (
	"net/http"

	_interface "github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/interface"
	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/request"
	"github.com/labstack/echo/v4"
)

type BatchUploadCloudRepositoryHandler struct {
	UseCase _interface.IBatchUploadCloudRepositoryUseCase
}

func NewBatchUploadCloudRepositoryHandler(c *echo.Group, useCase _interface.IBatchUploadCloudRepositoryUseCase) _interface.IBatchUploadCloudRepositoryHandler {
	handler := &BatchUploadCloudRepositoryHandler{
		UseCase: useCase,
	}
	c.POST("/files/upload/batch", handler.RequestBatchUploadURL)
	return handler
}

// RequestBatchUploadURL handles the request for multiple presigned upload URLs (max 30)
// @Summary Request multiple presigned upload URLs
// @Description Get presigned URLs for uploading multiple files to S3 (max 30 files)
// @Tags CloudRepository
// @Accept json
// @Produce json
// @Param body body request.BatchUploadRequestDTO true "Batch upload request"
// @Success 200 {object} response.BatchUploadResponseDTO
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/files/upload/batch [post]
func (h *BatchUploadCloudRepositoryHandler) RequestBatchUploadURL(c echo.Context) error {
	ctx := c.Request().Context()

	userID, err := getUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	var req request.BatchUploadRequestDTO
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	resp, err := h.UseCase.RequestBatchUploadURL(ctx, userID, &req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, resp)
}
