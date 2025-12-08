package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/entity"
	_interface "github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/interface"
	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/request"
	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/response"
)

type FolderUseCase struct {
	FolderRepo     _interface.IFolderRepository
	ContextTimeout time.Duration
}

func NewFolderUseCase(
	folderRepo _interface.IFolderRepository,
	timeout time.Duration,
) _interface.IFolderUseCase {
	return &FolderUseCase{
		FolderRepo:     folderRepo,
		ContextTimeout: timeout,
	}
}

// CreateFolder creates a new folder
func (u *FolderUseCase) CreateFolder(
	c context.Context,
	userID uint,
	req *request.CreateFolderRequestDTO,
) (*response.FolderResponseDTO, error) {
	ctx, cancel := context.WithTimeout(c, u.ContextTimeout)
	defer cancel()

	// Create folder entity
	folder := &entity.Folder{
		UserID:         userID,
		FolderName:     req.FolderName,
		ParentFolderID: req.ParentFolderID,
	}

	// Create folder in repository
	if err := u.FolderRepo.CreateFolder(ctx, folder); err != nil {
		return nil, fmt.Errorf("failed to create folder: %w", err)
	}

	// Get file count (should be 0 for new folder)
	fileCount, err := u.FolderRepo.GetFolderFileCount(ctx, folder.ID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get file count: %w", err)
	}

	return &response.FolderResponseDTO{
		ID:             folder.ID,
		FolderName:     folder.FolderName,
		ParentFolderID: folder.ParentFolderID,
		FileCount:      fileCount,
		CreatedAt:      folder.CreatedAt,
		UpdatedAt:      folder.UpdatedAt,
	}, nil
}

// GetFolders retrieves all folders for a user in tree structure
func (u *FolderUseCase) GetFolders(
	c context.Context,
	userID uint,
) ([]response.FolderTreeResponseDTO, error) {
	ctx, cancel := context.WithTimeout(c, u.ContextTimeout)
	defer cancel()

	// Get all folders
	folders, err := u.FolderRepo.GetFoldersByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get folders: %w", err)
	}

	// Build folder map and get file counts
	folderMap := make(map[uint]*response.FolderTreeResponseDTO)
	for _, folder := range folders {
		fileCount, err := u.FolderRepo.GetFolderFileCount(ctx, folder.ID, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to get file count: %w", err)
		}

		folderMap[folder.ID] = &response.FolderTreeResponseDTO{
			ID:             folder.ID,
			FolderName:     folder.FolderName,
			ParentFolderID: folder.ParentFolderID,
			FileCount:      fileCount,
			SubFolders:     []response.FolderTreeResponseDTO{},
			CreatedAt:      folder.CreatedAt,
			UpdatedAt:      folder.UpdatedAt,
		}
	}

	// Build tree structure
	var rootFolders []response.FolderTreeResponseDTO
	for _, folder := range folders {
		folderDTO := folderMap[folder.ID]
		if folder.ParentFolderID == nil {
			// Root level folder
			rootFolders = append(rootFolders, *folderDTO)
		} else {
			// Child folder - add to parent's SubFolders
			if parent, exists := folderMap[*folder.ParentFolderID]; exists {
				parent.SubFolders = append(parent.SubFolders, *folderDTO)
			}
		}
	}

	return rootFolders, nil
}

// GetFolderByID retrieves a single folder by ID
func (u *FolderUseCase) GetFolderByID(
	c context.Context,
	folderID uint,
	userID uint,
) (*response.FolderResponseDTO, error) {
	ctx, cancel := context.WithTimeout(c, u.ContextTimeout)
	defer cancel()

	// Get folder
	folder, err := u.FolderRepo.GetFolderByID(ctx, folderID, userID)
	if err != nil {
		return nil, fmt.Errorf("folder not found: %w", err)
	}

	// Get file count
	fileCount, err := u.FolderRepo.GetFolderFileCount(ctx, folder.ID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get file count: %w", err)
	}

	return &response.FolderResponseDTO{
		ID:             folder.ID,
		FolderName:     folder.FolderName,
		ParentFolderID: folder.ParentFolderID,
		FileCount:      fileCount,
		CreatedAt:      folder.CreatedAt,
		UpdatedAt:      folder.UpdatedAt,
	}, nil
}

// UpdateFolder updates folder name or parent folder
func (u *FolderUseCase) UpdateFolder(
	c context.Context,
	folderID uint,
	userID uint,
	req *request.UpdateFolderRequestDTO,
) (*response.FolderResponseDTO, error) {
	ctx, cancel := context.WithTimeout(c, u.ContextTimeout)
	defer cancel()

	// Get existing folder
	folder, err := u.FolderRepo.GetFolderByID(ctx, folderID, userID)
	if err != nil {
		return nil, fmt.Errorf("folder not found: %w", err)
	}

	// Update fields if provided
	if req.FolderName != nil {
		folder.FolderName = *req.FolderName
	}
	if req.ParentFolderID != nil {
		folder.ParentFolderID = req.ParentFolderID
	}

	// Update folder
	if err := u.FolderRepo.UpdateFolder(ctx, folder); err != nil {
		return nil, fmt.Errorf("failed to update folder: %w", err)
	}

	// Get file count
	fileCount, err := u.FolderRepo.GetFolderFileCount(ctx, folder.ID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get file count: %w", err)
	}

	return &response.FolderResponseDTO{
		ID:             folder.ID,
		FolderName:     folder.FolderName,
		ParentFolderID: folder.ParentFolderID,
		FileCount:      fileCount,
		CreatedAt:      folder.CreatedAt,
		UpdatedAt:      folder.UpdatedAt,
	}, nil
}

// DeleteFolder soft deletes a folder
func (u *FolderUseCase) DeleteFolder(
	c context.Context,
	folderID uint,
	userID uint,
) error {
	ctx, cancel := context.WithTimeout(c, u.ContextTimeout)
	defer cancel()

	// Delete folder (repository handles moving files to root)
	if err := u.FolderRepo.DeleteFolder(ctx, folderID, userID); err != nil {
		return fmt.Errorf("failed to delete folder: %w", err)
	}

	return nil
}

// GetFolderFiles retrieves all files in a folder
func (u *FolderUseCase) GetFolderFiles(
	c context.Context,
	folderID uint,
	userID uint,
) ([]entity.CloudFile, error) {
	ctx, cancel := context.WithTimeout(c, u.ContextTimeout)
	defer cancel()

	// Verify folder exists
	_, err := u.FolderRepo.GetFolderByID(ctx, folderID, userID)
	if err != nil {
		return nil, fmt.Errorf("folder not found: %w", err)
	}

	// Get files
	files, err := u.FolderRepo.GetFilesByFolderID(ctx, &folderID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get folder files: %w", err)
	}

	return files, nil
}

// MoveFiles moves multiple files to a target folder
func (u *FolderUseCase) MoveFiles(
	c context.Context,
	userID uint,
	req *request.MoveFilesToFolderRequestDTO,
) (*response.MoveFilesResponseDTO, error) {
	ctx, cancel := context.WithTimeout(c, u.ContextTimeout)
	defer cancel()

	// Move files
	movedCount, err := u.FolderRepo.MoveFilesToFolder(ctx, req.FileIDs, req.TargetFolderID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to move files: %w", err)
	}

	message := fmt.Sprintf("Successfully moved %d file(s)", movedCount)
	if req.TargetFolderID == nil {
		message = fmt.Sprintf("Successfully moved %d file(s) to root", movedCount)
	}

	return &response.MoveFilesResponseDTO{
		MovedCount: movedCount,
		Message:    message,
	}, nil
}
