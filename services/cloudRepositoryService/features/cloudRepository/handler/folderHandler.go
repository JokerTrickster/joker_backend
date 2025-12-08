package handler

import (
	"net/http"
	"strconv"

	_interface "github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/interface"
	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/request"
	"github.com/labstack/echo/v4"
)

type FolderHandler struct {
	UseCase _interface.IFolderUseCase
}

func NewFolderHandler(c *echo.Group, useCase _interface.IFolderUseCase) _interface.IFolderHandler {
	handler := &FolderHandler{
		UseCase: useCase,
	}

	// Register routes
	c.POST("/folders", handler.CreateFolder)
	c.GET("/folders", handler.GetFolders)
	c.GET("/folders/:id", handler.GetFolderByID)
	c.PATCH("/folders/:id", handler.UpdateFolder)
	c.DELETE("/folders/:id", handler.DeleteFolder)
	c.GET("/folders/:id/files", handler.GetFolderFiles)
	c.POST("/files/batch/move", handler.MoveFiles)

	return handler
}

// CreateFolder handles folder creation
// @Summary Create a new folder
// @Description Create a new folder for organizing files
// @Tags CloudRepository
// @Accept json
// @Produce json
// @Param body body request.CreateFolderRequestDTO true "Folder details"
// @Success 201 {object} response.FolderResponseDTO
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/folders [post]
func (h *FolderHandler) CreateFolder(c echo.Context) error {
	ctx := c.Request().Context()

	// Get user ID from JWT token
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	// Parse request body
	var req request.CreateFolderRequestDTO
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	// Create folder
	resp, err := h.UseCase.CreateFolder(ctx, userID, &req)
	if err != nil {
		if err.Error() == "failed to create folder: parent folder not found" {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "parent folder not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, resp)
}

// GetFolders retrieves all folders in tree structure
// @Summary Get all folders
// @Description Get all folders for the user in tree structure
// @Tags CloudRepository
// @Accept json
// @Produce json
// @Success 200 {array} response.FolderTreeResponseDTO
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/folders [get]
func (h *FolderHandler) GetFolders(c echo.Context) error {
	ctx := c.Request().Context()

	// Get user ID from JWT token
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	// Get folders
	resp, err := h.UseCase.GetFolders(ctx, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, resp)
}

// GetFolderByID retrieves a single folder by ID
// @Summary Get folder by ID
// @Description Get folder details by ID
// @Tags CloudRepository
// @Accept json
// @Produce json
// @Param id path int true "Folder ID"
// @Success 200 {object} response.FolderResponseDTO
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/folders/{id} [get]
func (h *FolderHandler) GetFolderByID(c echo.Context) error {
	ctx := c.Request().Context()

	// Get user ID from JWT token
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	// Get folder ID from path parameter
	folderIDStr := c.Param("id")
	folderIDInt, err := strconv.Atoi(folderIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid folder id"})
	}
	folderID := uint(folderIDInt)

	// Get folder
	resp, err := h.UseCase.GetFolderByID(ctx, folderID, userID)
	if err != nil {
		if err.Error() == "folder not found: record not found" {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "folder not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, resp)
}

// UpdateFolder updates folder name or parent
// @Summary Update folder
// @Description Update folder name or move to different parent
// @Tags CloudRepository
// @Accept json
// @Produce json
// @Param id path int true "Folder ID"
// @Param body body request.UpdateFolderRequestDTO true "Update details"
// @Success 200 {object} response.FolderResponseDTO
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/folders/{id} [patch]
func (h *FolderHandler) UpdateFolder(c echo.Context) error {
	ctx := c.Request().Context()

	// Get user ID from JWT token
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	// Get folder ID from path parameter
	folderIDStr := c.Param("id")
	folderIDInt, err := strconv.Atoi(folderIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid folder id"})
	}
	folderID := uint(folderIDInt)

	// Parse request body
	var req request.UpdateFolderRequestDTO
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	// Update folder
	resp, err := h.UseCase.UpdateFolder(ctx, folderID, userID, &req)
	if err != nil {
		if err.Error() == "folder not found: record not found" {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "folder not found"})
		}
		if err.Error() == "failed to update folder: parent folder not found" {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "parent folder not found"})
		}
		if err.Error() == "failed to update folder: folder cannot be its own parent" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "folder cannot be its own parent"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, resp)
}

// DeleteFolder soft deletes a folder
// @Summary Delete folder
// @Description Soft delete a folder and move its files to root
// @Tags CloudRepository
// @Accept json
// @Produce json
// @Param id path int true "Folder ID"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/folders/{id} [delete]
func (h *FolderHandler) DeleteFolder(c echo.Context) error {
	ctx := c.Request().Context()

	// Get user ID from JWT token
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	// Get folder ID from path parameter
	folderIDStr := c.Param("id")
	folderIDInt, err := strconv.Atoi(folderIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid folder id"})
	}
	folderID := uint(folderIDInt)

	// Delete folder
	err = h.UseCase.DeleteFolder(ctx, folderID, userID)
	if err != nil {
		errMsg := err.Error()
		if errMsg == "failed to delete folder: record not found" {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "folder not found"})
		}
		// Check for sub-folder error (starts with "failed to delete folder: cannot delete folder with")
		if len(errMsg) > 47 && errMsg[:47] == "failed to delete folder: cannot delete folder w" {
			// Extract the actual error message after "failed to delete folder: "
			actualError := errMsg[26:]
			return c.JSON(http.StatusBadRequest, map[string]string{"error": actualError})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": errMsg})
	}

	return c.NoContent(http.StatusNoContent)
}

// GetFolderFiles retrieves all files in a folder
// @Summary Get folder files
// @Description Get all files in a specific folder
// @Tags CloudRepository
// @Accept json
// @Produce json
// @Param id path int true "Folder ID"
// @Success 200 {array} entity.CloudFile
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/folders/{id}/files [get]
func (h *FolderHandler) GetFolderFiles(c echo.Context) error {
	ctx := c.Request().Context()

	// Get user ID from JWT token
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	// Get folder ID from path parameter
	folderIDStr := c.Param("id")
	folderIDInt, err := strconv.Atoi(folderIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid folder id"})
	}
	folderID := uint(folderIDInt)

	// Get files
	files, err := h.UseCase.GetFolderFiles(ctx, folderID, userID)
	if err != nil {
		if err.Error() == "folder not found: record not found" {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "folder not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, files)
}

// MoveFiles moves multiple files to a target folder
// @Summary Move files to folder
// @Description Move multiple files to a target folder or root
// @Tags CloudRepository
// @Accept json
// @Produce json
// @Param body body request.MoveFilesToFolderRequestDTO true "Move request"
// @Success 200 {object} response.MoveFilesResponseDTO
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/files/batch/move [post]
func (h *FolderHandler) MoveFiles(c echo.Context) error {
	ctx := c.Request().Context()

	// Get user ID from JWT token
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	// Parse request body
	var req request.MoveFilesToFolderRequestDTO
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	// Move files
	resp, err := h.UseCase.MoveFiles(ctx, userID, &req)
	if err != nil {
		if err.Error() == "failed to move files: target folder not found" {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "target folder not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, resp)
}
