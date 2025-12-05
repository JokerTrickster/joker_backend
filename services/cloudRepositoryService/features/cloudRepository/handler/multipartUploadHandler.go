package handler

import (
	"net/http"

	_interface "github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/interface"
	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/request"
	"github.com/labstack/echo/v4"
)

type MultipartUploadHandler struct {
	UseCase _interface.IMultipartUploadUseCase
}

func NewMultipartUploadHandler(c *echo.Group, useCase _interface.IMultipartUploadUseCase) _interface.IMultipartUploadHandler {
	handler := &MultipartUploadHandler{
		UseCase: useCase,
	}
	c.POST("/files/multipart/initiate", handler.InitiateMultipartUpload)
	c.POST("/files/multipart/presigned-urls", handler.GeneratePresignedURLs)
	c.POST("/files/multipart/complete", handler.CompleteMultipartUpload)
	c.POST("/files/multipart/abort", handler.AbortMultipartUpload)
	return handler
}

// InitiateMultipartUpload handles the initiation of a multipart upload
// @Summary Initiate multipart upload
// @Description Initiate a multipart upload session for large files (>5GB)
// @Tags CloudRepository
// @Accept json
// @Produce json
// @Param body body request.InitiateMultipartUploadRequestDTO true "Initiate request"
// @Success 200 {object} response.InitiateMultipartUploadResponseDTO
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/files/multipart/initiate [post]
func (h *MultipartUploadHandler) InitiateMultipartUpload(c echo.Context) error {
	ctx := c.Request().Context()

	// Get user ID from JWT token
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	var req request.InitiateMultipartUploadRequestDTO
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	resp, err := h.UseCase.InitiateMultipartUpload(ctx, userID, &req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, resp)
}

// GeneratePresignedURLs handles generating presigned URLs for parts
// @Summary Generate presigned URLs for parts
// @Description Generate presigned URLs for uploading specific parts of a multipart upload
// @Tags CloudRepository
// @Accept json
// @Produce json
// @Param body body request.GeneratePresignedURLsRequestDTO true "Presigned URLs request"
// @Success 200 {object} response.GeneratePresignedURLsResponseDTO
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/files/multipart/presigned-urls [post]
func (h *MultipartUploadHandler) GeneratePresignedURLs(c echo.Context) error {
	ctx := c.Request().Context()

	// Get user ID from JWT token
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	var req request.GeneratePresignedURLsRequestDTO
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	resp, err := h.UseCase.GeneratePresignedURLs(ctx, userID, &req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, resp)
}

// CompleteMultipartUpload handles completing a multipart upload
// @Summary Complete multipart upload
// @Description Complete a multipart upload after all parts have been uploaded
// @Tags CloudRepository
// @Accept json
// @Produce json
// @Param body body request.CompleteMultipartUploadRequestDTO true "Complete request"
// @Success 200 {object} response.CompleteMultipartUploadResponseDTO
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/files/multipart/complete [post]
func (h *MultipartUploadHandler) CompleteMultipartUpload(c echo.Context) error {
	ctx := c.Request().Context()

	// Get user ID from JWT token
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	var req request.CompleteMultipartUploadRequestDTO
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	resp, err := h.UseCase.CompleteMultipartUpload(ctx, userID, &req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, resp)
}

// AbortMultipartUpload handles aborting a multipart upload
// @Summary Abort multipart upload
// @Description Abort a multipart upload and clean up resources
// @Tags CloudRepository
// @Accept json
// @Produce json
// @Param body body request.AbortMultipartUploadRequestDTO true "Abort request"
// @Success 200 {object} response.AbortMultipartUploadResponseDTO
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/files/multipart/abort [post]
func (h *MultipartUploadHandler) AbortMultipartUpload(c echo.Context) error {
	ctx := c.Request().Context()

	// Get user ID from JWT token
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	var req request.AbortMultipartUploadRequestDTO
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	resp, err := h.UseCase.AbortMultipartUpload(ctx, userID, &req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, resp)
}
