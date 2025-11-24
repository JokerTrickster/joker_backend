package handler

import (
	"net/http"

	_interface "github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/interface"
	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/request"
	"github.com/labstack/echo/v4"
)

type UploadCloudRepositoryHandler struct {
	UseCase _interface.IUploadCloudRepositoryUseCase
}

func NewUploadCloudRepositoryHandler(c *echo.Group, useCase _interface.IUploadCloudRepositoryUseCase) _interface.IUploadCloudRepositoryHandler {
	handler := &UploadCloudRepositoryHandler{
		UseCase: useCase,
	}
	c.POST("/files/upload", handler.RequestUploadURL)
	return handler
}

// RequestUploadURL handles the request for a presigned upload URL
// @Summary Request presigned upload URL
// @Description Get a presigned URL for uploading a file to S3
// @Tags CloudRepository
// @Accept json
// @Produce json
// @Param body body request.UploadRequestDTO true "Upload request"
// @Success 200 {object} response.UploadResponseDTO
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/files/upload [post]
func (h *UploadCloudRepositoryHandler) RequestUploadURL(c echo.Context) error {
	ctx := c.Request().Context()

	// Get user ID from JWT token (assuming middleware sets this)
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	var req request.UploadRequestDTO
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	resp, err := h.UseCase.RequestUploadURL(ctx, userID, &req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, resp)
}
