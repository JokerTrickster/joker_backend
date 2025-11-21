package handler

import (
	"net/http"

	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model"
	"github.com/labstack/echo/v4"
)

// RequestUploadURL handles the request for a presigned upload URL
// @Summary Request presigned upload URL
// @Description Get a presigned URL for uploading a file to S3
// @Tags CloudRepository
// @Accept json
// @Produce json
// @Param body body model.UploadRequestDTO true "Upload request"
// @Success 200 {object} model.UploadResponseDTO
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/files/upload [post]
func (h *CloudRepositoryHandler) RequestUploadURL(c echo.Context) error {
	ctx := c.Request().Context()

	// Get user ID from JWT token (assuming middleware sets this)
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	var req model.UploadRequestDTO
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	resp, err := h.usecase.RequestUploadURL(ctx, userID, &req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, resp)
}

// RequestBatchUploadURL handles the request for multiple presigned upload URLs (max 30)
// @Summary Request multiple presigned upload URLs
// @Description Get presigned URLs for uploading multiple files to S3 (max 30 files)
// @Tags CloudRepository
// @Accept json
// @Produce json
// @Param body body model.BatchUploadRequestDTO true "Batch upload request"
// @Success 200 {object} model.BatchUploadResponseDTO
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/files/upload/batch [post]
func (h *CloudRepositoryHandler) RequestBatchUploadURL(c echo.Context) error {
	ctx := c.Request().Context()

	userID, err := getUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	var req model.BatchUploadRequestDTO
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	resp, err := h.usecase.RequestBatchUploadURL(ctx, userID, &req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, resp)
}
