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

type FolderShareUseCase struct {
	FolderRepo      _interface.IFolderRepository
	FolderShareRepo _interface.IFolderShareRepository
	ContextTimeout  time.Duration
}

func NewFolderShareUseCase(
	folderRepo _interface.IFolderRepository,
	folderShareRepo _interface.IFolderShareRepository,
	timeout time.Duration,
) _interface.IFolderShareUseCase {
	return &FolderShareUseCase{
		FolderRepo:      folderRepo,
		FolderShareRepo: folderShareRepo,
		ContextTimeout:  timeout,
	}
}

// ShareFolder shares a folder with multiple users
func (u *FolderShareUseCase) ShareFolder(
	c context.Context,
	folderID uint,
	ownerID uint,
	req *request.ShareFolderRequestDTO,
) (*response.ShareFolderResponseDTO, error) {
	ctx, cancel := context.WithTimeout(c, u.ContextTimeout)
	defer cancel()

	// Verify folder exists and belongs to owner
	folder, err := u.FolderRepo.GetFolderByID(ctx, folderID, ownerID)
	if err != nil {
		return nil, fmt.Errorf("folder not found or access denied: %w", err)
	}

	// Get users by emails
	users, err := u.FolderShareRepo.GetUsersByEmails(ctx, req.UserEmails)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}

	if len(users) == 0 {
		return nil, fmt.Errorf("no valid users found with provided emails")
	}

	// Create shares for each user
	sharedUsers := make([]response.ShareUserDTO, 0, len(users))
	for _, user := range users {
		// Skip if user is the owner
		if user.ID == ownerID {
			continue
		}

		share := &entity.FolderShare{
			FolderID:     folderID,
			OwnerID:      ownerID,
			SharedWithID: user.ID,
		}

		// Create share (will ignore if already exists due to unique constraint)
		err := u.FolderShareRepo.CreateFolderShare(ctx, share)
		if err != nil {
			// Log error but continue with other users
			fmt.Printf("Failed to share folder with user %d: %v\n", user.ID, err)
			continue
		}

		sharedUsers = append(sharedUsers, response.ShareUserDTO{
			ID:       user.ID,
			Name:     user.Name,
			Email:    user.Email,
			SharedAt: share.CreatedAt,
		})
	}

	message := fmt.Sprintf("%d명의 사용자와 '%s' 폴더를 공유했습니다", len(sharedUsers), folder.FolderName)

	return &response.ShareFolderResponseDTO{
		Message:     message,
		SharedUsers: sharedUsers,
	}, nil
}

// GetFolderShares retrieves all shares for a folder
func (u *FolderShareUseCase) GetFolderShares(
	c context.Context,
	folderID uint,
	ownerID uint,
) (*response.FolderShareListResponseDTO, error) {
	ctx, cancel := context.WithTimeout(c, u.ContextTimeout)
	defer cancel()

	// Verify folder exists and belongs to owner
	folder, err := u.FolderRepo.GetFolderByID(ctx, folderID, ownerID)
	if err != nil {
		return nil, fmt.Errorf("folder not found or access denied: %w", err)
	}

	// Get all shares for the folder
	shares, err := u.FolderShareRepo.GetFolderSharesByFolderID(ctx, folderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get folder shares: %w", err)
	}

	// Convert to DTO
	sharedUsers := make([]response.ShareUserDTO, 0, len(shares))
	for _, share := range shares {
		if share.SharedWith != nil {
			sharedUsers = append(sharedUsers, response.ShareUserDTO{
				ID:       share.SharedWith.ID,
				Name:     share.SharedWith.Name,
				Email:    share.SharedWith.Email,
				SharedAt: share.CreatedAt,
			})
		}
	}

	return &response.FolderShareListResponseDTO{
		FolderID:    folder.ID,
		FolderName:  folder.FolderName,
		SharedUsers: sharedUsers,
	}, nil
}

// RevokeFolderShare revokes folder access from a user
func (u *FolderShareUseCase) RevokeFolderShare(
	c context.Context,
	folderID uint,
	sharedWithID uint,
	ownerID uint,
) (*response.RevokeShareResponseDTO, error) {
	ctx, cancel := context.WithTimeout(c, u.ContextTimeout)
	defer cancel()

	// Verify folder exists and belongs to owner
	_, err := u.FolderRepo.GetFolderByID(ctx, folderID, ownerID)
	if err != nil {
		return nil, fmt.Errorf("folder not found or access denied: %w", err)
	}

	// Delete the share
	err = u.FolderShareRepo.DeleteFolderShare(ctx, folderID, sharedWithID, ownerID)
	if err != nil {
		return nil, fmt.Errorf("failed to revoke folder share: %w", err)
	}

	return &response.RevokeShareResponseDTO{
		Message: "공유가 취소되었습니다",
	}, nil
}

// GetSharedWithMeFolders retrieves all folders shared with the current user
func (u *FolderShareUseCase) GetSharedWithMeFolders(
	c context.Context,
	userID uint,
) (*response.SharedWithMeFoldersResponseDTO, error) {
	ctx, cancel := context.WithTimeout(c, u.ContextTimeout)
	defer cancel()

	// Get all shares where user is the recipient
	shares, err := u.FolderShareRepo.GetSharedFoldersByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get shared folders: %w", err)
	}

	// Convert to DTO
	folders := make([]response.SharedFolderDTO, 0, len(shares))
	for _, share := range shares {
		if share.Folder != nil {
			// Get file count for the folder
			fileCount, err := u.FolderRepo.GetFolderFileCount(ctx, share.Folder.ID, share.Folder.UserID)
			if err != nil {
				fileCount = 0 // Default to 0 on error
			}

			ownerInfo := response.UserInfoDTO{}
			if share.Owner != nil {
				ownerInfo.ID = share.Owner.ID
				ownerInfo.Name = share.Owner.Name
				ownerInfo.Email = share.Owner.Email
			}

			folders = append(folders, response.SharedFolderDTO{
				ID:         share.Folder.ID,
				FolderName: share.Folder.FolderName,
				Owner:      ownerInfo,
				FileCount:  fileCount,
				SharedAt:   share.CreatedAt,
				CreatedAt:  share.Folder.CreatedAt,
			})
		}
	}

	return &response.SharedWithMeFoldersResponseDTO{
		Folders: folders,
	}, nil
}
