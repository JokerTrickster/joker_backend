package handler

import (
	"net/http"
	"strconv"

	_interface "github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/interface"
	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/request"
	"github.com/labstack/echo/v4"
)

type ShareHandler struct {
	FolderShareUseCase _interface.IFolderShareUseCase
	FileShareUseCase   _interface.IFileShareUseCase
}

func NewShareHandler(
	c *echo.Group,
	folderShareUseCase _interface.IFolderShareUseCase,
	fileShareUseCase _interface.IFileShareUseCase,
) *ShareHandler {
	handler := &ShareHandler{
		FolderShareUseCase: folderShareUseCase,
		FileShareUseCase:   fileShareUseCase,
	}

	// Folder share routes
	c.POST("/folders/:id/share", handler.ShareFolder)
	c.GET("/folders/:id/shares", handler.GetFolderShares)
	c.DELETE("/folders/:id/shares/:userId", handler.RevokeFolderShare)
	c.GET("/folders/shared-with-me", handler.GetSharedWithMeFolders)

	// File share routes
	c.POST("/files/:id/share", handler.ShareFile)
	c.GET("/files/:id/shares", handler.GetFileShares)
	c.DELETE("/files/:id/shares/:userId", handler.RevokeFileShare)
	c.GET("/files/shared-with-me", handler.GetSharedWithMeFiles)

	return handler
}

// ShareFolder handles folder sharing
// @Summary Share a folder with users
// @Description Share a folder with one or more users by email
// @Tags CloudRepository
// @Accept json
// @Produce json
// @Param id path int true "Folder ID"
// @Param body body request.ShareFolderRequestDTO true "User emails"
// @Success 200 {object} response.ShareFolderResponseDTO
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/folders/{id}/share [post]
func (h *ShareHandler) ShareFolder(c echo.Context) error {
	ctx := c.Request().Context()

	// Get user ID from JWT token
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	// Parse folder ID
	folderID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid folder ID"})
	}

	// Parse request body
	var req request.ShareFolderRequestDTO
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	// Share folder
	resp, err := h.FolderShareUseCase.ShareFolder(ctx, uint(folderID), int32(userID), &req)
	if err != nil {
		if err.Error() == "folder not found or access denied: record not found" {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "folder not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, resp)
}

// GetFolderShares retrieves all shares for a folder
// @Summary Get folder shares
// @Description Get list of users who have access to a folder
// @Tags CloudRepository
// @Accept json
// @Produce json
// @Param id path int true "Folder ID"
// @Success 200 {object} response.FolderShareListResponseDTO
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/folders/{id}/shares [get]
func (h *ShareHandler) GetFolderShares(c echo.Context) error {
	ctx := c.Request().Context()

	// Get user ID from JWT token
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	// Parse folder ID
	folderID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid folder ID"})
	}

	// Get shares
	resp, err := h.FolderShareUseCase.GetFolderShares(ctx, uint(folderID), int32(userID))
	if err != nil {
		if err.Error() == "folder not found or access denied: record not found" {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "folder not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, resp)
}

// RevokeFolderShare revokes folder access from a user
// @Summary Revoke folder share
// @Description Remove user's access to a folder
// @Tags CloudRepository
// @Accept json
// @Produce json
// @Param id path int true "Folder ID"
// @Param userId path int true "User ID to revoke"
// @Success 200 {object} response.RevokeShareResponseDTO
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/folders/{id}/shares/{userId} [delete]
func (h *ShareHandler) RevokeFolderShare(c echo.Context) error {
	ctx := c.Request().Context()

	// Get owner ID from JWT token
	ownerID, err := getUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	// Parse folder ID
	folderID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid folder ID"})
	}

	// Parse user ID
	sharedWithID, err := strconv.ParseUint(c.Param("userId"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid user ID"})
	}

	// Revoke share
	resp, err := h.FolderShareUseCase.RevokeFolderShare(ctx, uint(folderID), int32(sharedWithID), int32(ownerID))
	if err != nil {
		if err.Error() == "folder not found or access denied: record not found" {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "folder not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, resp)
}

// GetSharedWithMeFolders retrieves folders shared with the current user
// @Summary Get folders shared with me
// @Description Get list of folders that have been shared with the current user
// @Tags CloudRepository
// @Accept json
// @Produce json
// @Success 200 {object} response.SharedWithMeFoldersResponseDTO
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/folders/shared-with-me [get]
func (h *ShareHandler) GetSharedWithMeFolders(c echo.Context) error {
	ctx := c.Request().Context()

	// Get user ID from JWT token
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	// Get shared folders
	resp, err := h.FolderShareUseCase.GetSharedWithMeFolders(ctx, int32(userID))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, resp)
}

// ShareFile handles file sharing
// @Summary Share a file with users
// @Description Share a file with one or more users by email
// @Tags CloudRepository
// @Accept json
// @Produce json
// @Param id path int true "File ID"
// @Param body body request.ShareFileRequestDTO true "User emails"
// @Success 200 {object} response.ShareFileResponseDTO
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/files/{id}/share [post]
func (h *ShareHandler) ShareFile(c echo.Context) error {
	ctx := c.Request().Context()

	// Get user ID from JWT token
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	// Parse file ID
	fileID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid file ID"})
	}

	// Parse request body
	var req request.ShareFileRequestDTO
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	// Share file
	resp, err := h.FileShareUseCase.ShareFile(ctx, uint(fileID), int32(userID), &req)
	if err != nil {
		if err.Error() == "file not found or access denied" {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "file not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, resp)
}

// GetFileShares retrieves all shares for a file
// @Summary Get file shares
// @Description Get list of users who have access to a file
// @Tags CloudRepository
// @Accept json
// @Produce json
// @Param id path int true "File ID"
// @Success 200 {object} response.FileShareListResponseDTO
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/files/{id}/shares [get]
func (h *ShareHandler) GetFileShares(c echo.Context) error {
	ctx := c.Request().Context()

	// Get user ID from JWT token
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	// Parse file ID
	fileID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid file ID"})
	}

	// Get shares
	resp, err := h.FileShareUseCase.GetFileShares(ctx, uint(fileID), int32(userID))
	if err != nil {
		if err.Error() == "file not found or access denied" {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "file not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, resp)
}

// RevokeFileShare revokes file access from a user
// @Summary Revoke file share
// @Description Remove user's access to a file
// @Tags CloudRepository
// @Accept json
// @Produce json
// @Param id path int true "File ID"
// @Param userId path int true "User ID to revoke"
// @Success 200 {object} response.RevokeShareResponseDTO
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/files/{id}/shares/{userId} [delete]
func (h *ShareHandler) RevokeFileShare(c echo.Context) error {
	ctx := c.Request().Context()

	// Get owner ID from JWT token
	ownerID, err := getUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	// Parse file ID
	fileID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid file ID"})
	}

	// Parse user ID
	sharedWithID, err := strconv.ParseUint(c.Param("userId"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid user ID"})
	}

	// Revoke share
	resp, err := h.FileShareUseCase.RevokeFileShare(ctx, uint(fileID), int32(sharedWithID), int32(ownerID))
	if err != nil {
		if err.Error() == "file not found or access denied" {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "file not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, resp)
}

// GetSharedWithMeFiles retrieves files shared with the current user
// @Summary Get files shared with me
// @Description Get list of files that have been shared with the current user
// @Tags CloudRepository
// @Accept json
// @Produce json
// @Success 200 {object} response.SharedWithMeFilesResponseDTO
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/files/shared-with-me [get]
func (h *ShareHandler) GetSharedWithMeFiles(c echo.Context) error {
	ctx := c.Request().Context()

	// Get user ID from JWT token
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	// Get shared files
	resp, err := h.FileShareUseCase.GetSharedWithMeFiles(ctx, int32(userID))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, resp)
}
