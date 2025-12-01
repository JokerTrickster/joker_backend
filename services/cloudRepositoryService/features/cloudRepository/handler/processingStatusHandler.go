package handler

import (
	"net/http"
	"strconv"

	_interface "github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/interface"
	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/response"
	"github.com/labstack/echo/v4"
)

type ProcessingStatusHandler struct {
	UseCase _interface.IProcessingStatusUseCase
}

func NewProcessingStatusHandler(useCase _interface.IProcessingStatusUseCase) *ProcessingStatusHandler {
	return &ProcessingStatusHandler{
		UseCase: useCase,
	}
}

// RegisterRoutes registers processing status routes
func (h *ProcessingStatusHandler) RegisterRoutes(c *echo.Group) {
	c.GET("/files/:id/processing-status", h.GetProcessingStatus)
	c.POST("/files/processing-status/batch", h.GetBatchProcessingStatus)
}

// GetProcessingStatus handles the request for file processing status
// @Summary Get file processing status
// @Description Get the current processing status of a file
// @Tags Processing Status
// @Accept json
// @Produce json
// @Param id path int true "File ID"
// @Success 200 {object} response.ProcessingStatusResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/files/{id}/processing-status [get]
func (h *ProcessingStatusHandler) GetProcessingStatus(c echo.Context) error {
	ctx := c.Request().Context()

	// Get user ID from context (set by JWT middleware)
	userID, ok := c.Get("userID").(uint)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error":   "UNAUTHORIZED",
			"message": "Invalid or expired access token",
		})
	}

	// Parse file ID
	fileIDStr := c.Param("id")
	fileID, err := strconv.ParseUint(fileIDStr, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error":   "INVALID_FILE_ID",
			"message": "Invalid file ID format",
		})
	}

	// Get processing status
	resp, err := h.UseCase.GetProcessingStatus(ctx, userID, uint(fileID))
	if err != nil {
		if err.Error()[:14] == "FILE_NOT_FOUND" {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error":   "FILE_NOT_FOUND",
				"message": err.Error()[16:],
			})
		}
		if err.Error()[:17] == "PERMISSION_DENIED" {
			return c.JSON(http.StatusForbidden, map[string]string{
				"error":   "PERMISSION_DENIED",
				"message": "You do not have permission to access this file",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error":   "INTERNAL_SERVER_ERROR",
			"message": "Failed to fetch processing status",
		})
	}

	return c.JSON(http.StatusOK, resp)
}

// GetBatchProcessingStatus handles the request for multiple file processing statuses
// @Summary Get batch file processing status
// @Description Get the processing status of multiple files
// @Tags Processing Status
// @Accept json
// @Produce json
// @Param request body response.BatchProcessingStatusRequest true "File IDs"
// @Success 200 {object} response.BatchProcessingStatusResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/v1/files/processing-status/batch [post]
func (h *ProcessingStatusHandler) GetBatchProcessingStatus(c echo.Context) error {
	ctx := c.Request().Context()

	// Get user ID from context
	userID, ok := c.Get("userID").(uint)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error":   "UNAUTHORIZED",
			"message": "Invalid or expired access token",
		})
	}

	// Parse request
	var req response.BatchProcessingStatusRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error":   "INVALID_REQUEST",
			"message": "Invalid request format",
		})
	}

	// Validate request
	if len(req.FileIDs) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error":   "INVALID_REQUEST",
			"message": "At least one file ID is required",
		})
	}

	if len(req.FileIDs) > 100 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error":   "INVALID_REQUEST",
			"message": "Maximum 100 file IDs allowed",
		})
	}

	// Get batch processing status
	resp, err := h.UseCase.GetBatchProcessingStatus(ctx, userID, req.FileIDs)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error":   "INTERNAL_SERVER_ERROR",
			"message": "Failed to fetch processing statuses",
		})
	}

	return c.JSON(http.StatusOK, resp)
}
