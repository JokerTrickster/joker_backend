package handler

import (
	"net/http"
	"strconv"

	_interface "github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/interface"
	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/request"
	"github.com/labstack/echo/v4"
)

type TagHandler struct {
	UseCase _interface.ITagUseCase
}

func NewTagHandler(c *echo.Group, useCase _interface.ITagUseCase) _interface.ITagHandler {
	handler := &TagHandler{
		UseCase: useCase,
	}

	// Register routes
	c.PUT("/files/:id/tags", handler.UpdateFileTags)
	c.POST("/files/:id/tags", handler.AddTagToFile)
	c.DELETE("/files/:id/tags/:tag_name", handler.RemoveTagFromFile)

	return handler
}

// UpdateFileTags handles bulk tag update (replace all tags)
// @Summary Update file tags
// @Description Replace all tags for a file
// @Tags CloudRepository
// @Accept json
// @Produce json
// @Param id path int true "File ID"
// @Param body body request.UpdateFileTagsRequestDTO true "Tags to set"
// @Success 200 {object} response.UpdateFileTagsResponseDTO
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/files/{id}/tags [put]
func (h *TagHandler) UpdateFileTags(c echo.Context) error {
	ctx := c.Request().Context()

	// Get user ID from JWT token
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	// Get file ID from path parameter
	fileIDStr := c.Param("id")
	fileIDInt, err := strconv.Atoi(fileIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid file id"})
	}
	fileID := uint(fileIDInt)

	// Parse request body
	var req request.UpdateFileTagsRequestDTO
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	// Update tags
	resp, err := h.UseCase.UpdateFileTags(ctx, userID, fileID, &req)
	if err != nil {
		if err.Error() == "file not found: record not found" {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "file not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, resp)
}

// AddTagToFile handles adding a single tag to a file
// @Summary Add tag to file
// @Description Add a single tag to a file
// @Tags CloudRepository
// @Accept json
// @Produce json
// @Param id path int true "File ID"
// @Param body body request.AddFileTagRequestDTO true "Tag to add"
// @Success 200 {object} response.AddFileTagResponseDTO
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/files/{id}/tags [post]
func (h *TagHandler) AddTagToFile(c echo.Context) error {
	ctx := c.Request().Context()

	// Get user ID from JWT token
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	// Get file ID from path parameter
	fileIDStr := c.Param("id")
	fileIDInt, err := strconv.Atoi(fileIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid file id"})
	}
	fileID := uint(fileIDInt)

	// Parse request body
	var req request.AddFileTagRequestDTO
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	// Add tag
	resp, err := h.UseCase.AddTagToFile(ctx, userID, fileID, &req)
	if err != nil {
		if err.Error() == "file not found: record not found" {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "file not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, resp)
}

// RemoveTagFromFile handles removing a tag from a file
// @Summary Remove tag from file
// @Description Remove a tag from a file by tag name
// @Tags CloudRepository
// @Accept json
// @Produce json
// @Param id path int true "File ID"
// @Param tag_name path string true "Tag name to remove"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/files/{id}/tags/{tag_name} [delete]
func (h *TagHandler) RemoveTagFromFile(c echo.Context) error {
	ctx := c.Request().Context()

	// Get user ID from JWT token
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	// Get file ID from path parameter
	fileIDStr := c.Param("id")
	fileIDInt, err := strconv.Atoi(fileIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid file id"})
	}
	fileID := uint(fileIDInt)

	// Get tag name from path parameter
	tagName := c.Param("tag_name")
	if tagName == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "tag name is required"})
	}

	// Remove tag
	err = h.UseCase.RemoveTagFromFile(ctx, userID, fileID, tagName)
	if err != nil {
		if err.Error() == "file not found: record not found" {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "file not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.NoContent(http.StatusNoContent)
}
