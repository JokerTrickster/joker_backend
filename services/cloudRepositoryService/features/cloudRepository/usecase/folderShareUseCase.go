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
	ownerID int32,
	req *request.ShareFolderRequestDTO,
) (*response.ShareFolderResponseDTO, error) {
	ctx, cancel := context.WithTimeout(c, u.ContextTimeout)
	defer cancel()

	// Verify folder exists and belongs to owner
	_, err := u.FolderRepo.GetFolderByID(ctx, folderID, ownerID)
	if err != nil {
		return nil, fmt.Errorf("폴더를 찾을 수 없거나 권한이 없습니다")
	}

	// Get users by emails
	users, err := u.FolderShareRepo.GetUsersByEmails(ctx, req.UserEmails)
	if err != nil {
		return nil, fmt.Errorf("사용자 조회 실패: %w", err)
	}

	// Check which emails were not found
	userEmailMap := make(map[string]bool)
	for _, user := range users {
		userEmailMap[user.Email] = true
	}

	notFoundEmails := make([]string, 0)
	for _, email := range req.UserEmails {
		if !userEmailMap[email] {
			notFoundEmails = append(notFoundEmails, email)
		}
	}

	if len(notFoundEmails) > 0 {
		return nil, fmt.Errorf("존재하지 않는 사용자: %s", notFoundEmails[0])
	}

	if len(users) == 0 {
		return nil, fmt.Errorf("유효한 사용자를 찾을 수 없습니다")
	}

	// Get permission from request, default to "read"
	permission := entity.SharePermissionRead
	if req.Permission != "" {
		permission = entity.SharePermission(req.Permission)
	}

	// Get existing shares to prevent duplicates
	existingShares, err := u.FolderShareRepo.GetFolderSharesByFolderID(ctx, folderID)
	if err != nil {
		return nil, fmt.Errorf("기존 공유 정보 조회 실패: %w", err)
	}

	existingShareMap := make(map[int32]*entity.FolderShare)
	for _, share := range existingShares {
		existingShareMap[share.SharedWithID] = &share
	}

	// Create shares for each user
	sharedUsers := make([]response.ShareUserDTO, 0, len(users))
	newShareCount := 0

	for _, user := range users {
		// Skip if user is the owner
		if user.ID == ownerID {
			continue
		}

		// Skip if already shared (duplicate)
		if existingShare, exists := existingShareMap[user.ID]; exists {
			// Still add to response for frontend consistency
			sharedUsers = append(sharedUsers, response.ShareUserDTO{
				ID:         user.ID,
				Name:       user.Name,
				Email:      user.Email,
				Permission: string(existingShare.Permission),
			})
			continue
		}

		share := &entity.FolderShare{
			FolderID:     folderID,
			OwnerID:      ownerID,
			SharedWithID: user.ID,
			Permission:   permission,
		}

		// Create share
		err := u.FolderShareRepo.CreateFolderShare(ctx, share)
		if err != nil {
			// Log error but continue with other users
			fmt.Printf("Failed to share folder with user %d: %v\n", user.ID, err)
			continue
		}

		newShareCount++
		sharedUsers = append(sharedUsers, response.ShareUserDTO{
			ID:         user.ID,
			Name:       user.Name,
			Email:      user.Email,
			Permission: string(share.Permission),
			SharedAt:   share.CreatedAt,
		})
	}

	var message string
	if newShareCount > 0 {
		message = fmt.Sprintf("공유되었습니다")
	} else {
		message = "이미 공유된 사용자입니다"
	}

	return &response.ShareFolderResponseDTO{
		Message:     message,
		SharedUsers: sharedUsers,
	}, nil
}

// GetFolderShares retrieves all shares for a folder
func (u *FolderShareUseCase) GetFolderShares(
	c context.Context,
	folderID uint,
	ownerID int32,
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
				ID:         share.SharedWith.ID,
				Name:       share.SharedWith.Name,
				Email:      share.SharedWith.Email,
				Permission: string(share.Permission),
				SharedAt:   share.CreatedAt,
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
	sharedWithID int32,
	ownerID int32,
) (*response.RevokeShareResponseDTO, error) {
	ctx, cancel := context.WithTimeout(c, u.ContextTimeout)
	defer cancel()

	// Verify folder exists and belongs to owner
	_, err := u.FolderRepo.GetFolderByID(ctx, folderID, ownerID)
	if err != nil {
		return nil, fmt.Errorf("폴더/파일을 수정할 권한이 없습니다")
	}

	// Delete the share
	err = u.FolderShareRepo.DeleteFolderShare(ctx, folderID, sharedWithID, ownerID)
	if err != nil {
		return nil, fmt.Errorf("공유 취소 실패: %w", err)
	}

	return &response.RevokeShareResponseDTO{
		Message: "공유가 해제되었습니다",
	}, nil
}

// GetSharedWithMeFolders retrieves all folders shared with the current user
func (u *FolderShareUseCase) GetSharedWithMeFolders(
	c context.Context,
	userID int32,
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
			fileCount, err := u.FolderRepo.GetFolderFileCount(ctx, share.Folder.ID, int32(share.Folder.UserID))
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
				Permission: string(share.Permission),
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

// GetFoldersSharedByMe retrieves folders that the current user has shared with others
func (u *FolderShareUseCase) GetFoldersSharedByMe(
	c context.Context,
	ownerID int32,
) (*response.FoldersSharedByMeResponseDTO, error) {
	ctx, cancel := context.WithTimeout(c, u.ContextTimeout)
	defer cancel()

	// Get all folder shares created by this owner
	shares, err := u.FolderShareRepo.GetFoldersSharedByUserID(ctx, ownerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get folders shared by user: %w", err)
	}

	// Group shares by folder ID
	folderSharesMap := make(map[uint][]entity.FolderShare)
	for _, share := range shares {
		if share.Folder != nil && share.SharedWith != nil {
			folderSharesMap[share.Folder.ID] = append(folderSharesMap[share.Folder.ID], share)
		}
	}

	// Convert to DTO
	folders := make([]response.FolderSharedByMeDTO, 0)
	for folderID, folderShares := range folderSharesMap {
		if len(folderShares) == 0 {
			continue
		}

		folder := folderShares[0].Folder

		// Get file count for the folder
		fileCount, err := u.FolderRepo.GetFolderFileCount(ctx, folder.ID, int32(folder.UserID))
		if err != nil {
			fileCount = 0 // Default to 0 on error
		}

		// Build shared_with array
		sharedWith := make([]response.ShareUserDTO, 0, len(folderShares))
		for _, share := range folderShares {
			if share.SharedWith != nil {
				sharedWith = append(sharedWith, response.ShareUserDTO{
					ID:         share.SharedWith.ID,
					Name:       share.SharedWith.Name,
					Email:      share.SharedWith.Email,
					Permission: string(share.Permission),
					SharedAt:   share.CreatedAt,
				})
			}
		}

		folders = append(folders, response.FolderSharedByMeDTO{
			ID:         folderID,
			FolderName: folder.FolderName,
			FileCount:  fileCount,
			CreatedAt:  folder.CreatedAt,
			SharedWith: sharedWith,
		})
	}

	return &response.FoldersSharedByMeResponseDTO{
		Folders: folders,
	}, nil
}
