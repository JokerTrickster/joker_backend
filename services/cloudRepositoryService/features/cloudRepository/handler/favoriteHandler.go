package handler

import (
	"net/http"
	"strconv"
	"strings"

	_interface "github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/interface"
	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/request"
	"github.com/labstack/echo/v4"
)

type FavoriteHandler struct {
	UseCase _interface.IFavoriteUseCase
}

func NewFavoriteHandler(c *echo.Group, useCase _interface.IFavoriteUseCase) *FavoriteHandler {
	handler := &FavoriteHandler{
		UseCase: useCase,
	}
	c.POST("/favorites", handler.AddFavorite)
	c.DELETE("/favorites/:fileId", handler.RemoveFavorite)
	c.GET("/favorites", handler.ListFavorites)
	return handler
}

// AddFavorite handles adding a file to favorites
// @Summary Add file to favorites
// @Description Add a file to user's favorites list
// @Tags Favorites
// @Accept json
// @Produce json
// @Param body body request.AddFavoriteRequestDTO true "Add favorite request"
// @Success 200 {object} response.FavoriteResponseDTO
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/favorites [post]
func (h *FavoriteHandler) AddFavorite(c echo.Context) error {
	ctx := c.Request().Context()

	userID, err := getUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	var req request.AddFavoriteRequestDTO
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	resp, err := h.UseCase.AddFavorite(ctx, userID, req.FileID)
	if err != nil {
		// Determine appropriate status code based on error message
		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") {
			statusCode = http.StatusNotFound
		} else if strings.Contains(err.Error(), "access denied") || strings.Contains(err.Error(), "do not own") {
			statusCode = http.StatusForbidden
		}
		return c.JSON(statusCode, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, resp)
}

// RemoveFavorite handles removing a file from favorites
// @Summary Remove file from favorites
// @Description Remove a file from user's favorites list
// @Tags Favorites
// @Accept json
// @Produce json
// @Param fileId path int true "File ID"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/favorites/{fileId} [delete]
func (h *FavoriteHandler) RemoveFavorite(c echo.Context) error {
	ctx := c.Request().Context()

	userID, err := getUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	fileID, err := strconv.ParseUint(c.Param("fileId"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid file ID"})
	}

	if err := h.UseCase.RemoveFavorite(ctx, userID, uint(fileID)); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.NoContent(http.StatusNoContent)
}

// ListFavorites handles listing all favorited files
// @Summary List favorite files
// @Description List all files in user's favorites with filtering and pagination
// @Tags Favorites
// @Accept json
// @Produce json
// @Param page query int false "Page number (default: 1)"
// @Param size query int false "Page size (default: 20, max: 100)"
// @Param sort query string false "Sort field (uploadDate, fileName)"
// @Param order query string false "Sort order (asc, desc)"
// @Param q query string false "Filename search"
// @Param ext query string false "File extension filter"
// @Param tag query string false "Tag filter"
// @Success 200 {object} response.ListFavoritesResponseDTO
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/favorites [get]
func (h *FavoriteHandler) ListFavorites(c echo.Context) error {
	ctx := c.Request().Context()

	userID, err := getUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	var filter request.ListFavoritesRequestDTO
	if err := c.Bind(&filter); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid query parameters"})
	}

	if err := c.Validate(&filter); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	resp, err := h.UseCase.ListFavorites(ctx, userID, filter)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, resp)
}
