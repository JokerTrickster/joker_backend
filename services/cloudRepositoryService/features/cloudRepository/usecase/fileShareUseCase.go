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

type FileShareUseCase struct {
	FileShareRepo   _interface.IFileShareRepository
	FolderShareRepo _interface.IFolderShareRepository
	S3Repo          _interface.IListCloudRepositoryRepository
	ContextTimeout  time.Duration
}

func NewFileShareUseCase(
	fileShareRepo _interface.IFileShareRepository,
	folderShareRepo _interface.IFolderShareRepository,
	s3Repo _interface.IListCloudRepositoryRepository,
	timeout time.Duration,
) _interface.IFileShareUseCase {
	return &FileShareUseCase{
		FileShareRepo:   fileShareRepo,
		FolderShareRepo: folderShareRepo,
		S3Repo:          s3Repo,
		ContextTimeout:  timeout,
	}
}

// ShareFile shares a file with multiple users
func (u *FileShareUseCase) ShareFile(
	c context.Context,
	fileID uint,
	ownerID int32,
	req *request.ShareFileRequestDTO,
) (*response.ShareFileResponseDTO, error) {
	ctx, cancel := context.WithTimeout(c, u.ContextTimeout)
	defer cancel()

	// Verify file exists and user has access
	hasAccess, err := u.FileShareRepo.HasFileAccess(ctx, ownerID, fileID)
	if err != nil || !hasAccess {
		return nil, fmt.Errorf("파일을 찾을 수 없거나 권한이 없습니다")
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

	// Get existing shares to prevent duplicates
	existingShares, err := u.FileShareRepo.GetFileSharesByFileID(ctx, fileID)
	if err != nil {
		return nil, fmt.Errorf("기존 공유 정보 조회 실패: %w", err)
	}

	existingShareMap := make(map[int32]bool)
	for _, share := range existingShares {
		existingShareMap[share.SharedWithID] = true
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
		if existingShareMap[user.ID] {
			// Still add to response for frontend consistency
			sharedUsers = append(sharedUsers, response.ShareUserDTO{
				ID:    user.ID,
				Name:  user.Name,
				Email: user.Email,
			})
			continue
		}

		share := &entity.FileShare{
			FileID:       fileID,
			OwnerID:      ownerID,
			SharedWithID: user.ID,
		}

		// Create share
		err := u.FileShareRepo.CreateFileShare(ctx, share)
		if err != nil {
			// Log error but continue with other users
			fmt.Printf("Failed to share file with user %d: %v\n", user.ID, err)
			continue
		}

		newShareCount++
		sharedUsers = append(sharedUsers, response.ShareUserDTO{
			ID:       user.ID,
			Name:     user.Name,
			Email:    user.Email,
			SharedAt: share.CreatedAt,
		})
	}

	var message string
	if newShareCount > 0 {
		message = fmt.Sprintf("공유되었습니다")
	} else {
		message = "이미 공유된 사용자입니다"
	}

	return &response.ShareFileResponseDTO{
		Message:     message,
		SharedUsers: sharedUsers,
	}, nil
}

// GetFileShares retrieves all shares for a file
func (u *FileShareUseCase) GetFileShares(
	c context.Context,
	fileID uint,
	ownerID int32,
) (*response.FileShareListResponseDTO, error) {
	ctx, cancel := context.WithTimeout(c, u.ContextTimeout)
	defer cancel()

	// Verify file exists and user has access
	hasAccess, err := u.FileShareRepo.HasFileAccess(ctx, ownerID, fileID)
	if err != nil || !hasAccess {
		return nil, fmt.Errorf("file not found or access denied")
	}

	// Get all shares for the file
	shares, err := u.FileShareRepo.GetFileSharesByFileID(ctx, fileID)
	if err != nil {
		return nil, fmt.Errorf("failed to get file shares: %w", err)
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

	fileName := ""
	if len(shares) > 0 && shares[0].File != nil {
		fileName = shares[0].File.FileName
	}

	return &response.FileShareListResponseDTO{
		FileID:      fileID,
		FileName:    fileName,
		SharedUsers: sharedUsers,
	}, nil
}

// RevokeFileShare revokes file access from a user
func (u *FileShareUseCase) RevokeFileShare(
	c context.Context,
	fileID uint,
	sharedWithID int32,
	ownerID int32,
) (*response.RevokeShareResponseDTO, error) {
	ctx, cancel := context.WithTimeout(c, u.ContextTimeout)
	defer cancel()

	// Verify file exists and user has access
	hasAccess, err := u.FileShareRepo.HasFileAccess(ctx, ownerID, fileID)
	if err != nil || !hasAccess {
		return nil, fmt.Errorf("폴더/파일을 수정할 권한이 없습니다")
	}

	// Delete the share
	err = u.FileShareRepo.DeleteFileShare(ctx, fileID, sharedWithID, ownerID)
	if err != nil {
		return nil, fmt.Errorf("공유 취소 실패: %w", err)
	}

	return &response.RevokeShareResponseDTO{
		Message: "공유가 해제되었습니다",
	}, nil
}

// GetSharedWithMeFiles retrieves all files shared with the current user
func (u *FileShareUseCase) GetSharedWithMeFiles(
	c context.Context,
	userID int32,
) (*response.SharedWithMeFilesResponseDTO, error) {
	ctx, cancel := context.WithTimeout(c, u.ContextTimeout)
	defer cancel()

	// Get all shares where user is the recipient
	shares, err := u.FileShareRepo.GetSharedFilesByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get shared files: %w", err)
	}

	// Convert to DTO with presigned URLs
	files := make([]response.SharedFileDTO, 0, len(shares))
	for _, share := range shares {
		if share.File != nil {
			// Generate presigned download URL
			downloadURL, err := u.S3Repo.GeneratePresignedDownloadURL(ctx, share.File.S3Key, 1*time.Hour)
			if err != nil {
				fmt.Printf("Failed to generate download URL for file %d: %v\n", share.File.ID, err)
				continue
			}

			var thumbnailURL *string
			if share.File.ThumbnailKey != "" {
				url, err := u.S3Repo.GeneratePresignedDownloadURL(ctx, share.File.ThumbnailKey, 1*time.Hour)
				if err == nil {
					thumbnailURL = &url
				}
			}

			ownerInfo := response.UserInfoDTO{}
			if share.Owner != nil {
				ownerInfo.ID = share.Owner.ID
				ownerInfo.Name = share.Owner.Name
				ownerInfo.Email = share.Owner.Email
			}

			files = append(files, response.SharedFileDTO{
				ID:           share.File.ID,
				FileName:     share.File.FileName,
				FileType:     string(share.File.FileType),
				ContentType:  share.File.ContentType,
				FileSize:     share.File.FileSize,
				Owner:        ownerInfo,
				DownloadURL:  downloadURL,
				ThumbnailURL: thumbnailURL,
				SharedAt:     share.CreatedAt,
				CreatedAt:    share.File.CreatedAt,
			})
		}
	}

	return &response.SharedWithMeFilesResponseDTO{
		Files: files,
	}, nil
}

// GetFilesSharedByMe retrieves files that the current user has shared with others
func (u *FileShareUseCase) GetFilesSharedByMe(
	c context.Context,
	ownerID int32,
) (*response.FilesSharedByMeResponseDTO, error) {
	ctx, cancel := context.WithTimeout(c, u.ContextTimeout)
	defer cancel()

	// Get all file shares created by this owner
	shares, err := u.FileShareRepo.GetFilesSharedByUserID(ctx, ownerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get files shared by user: %w", err)
	}

	// Group shares by file ID
	fileSharesMap := make(map[uint][]entity.FileShare)
	for _, share := range shares {
		if share.File != nil && share.SharedWith != nil {
			fileSharesMap[share.File.ID] = append(fileSharesMap[share.File.ID], share)
		}
	}

	// Convert to DTO
	files := make([]response.FileSharedByMeDTO, 0)
	for fileID, fileShares := range fileSharesMap {
		if len(fileShares) == 0 {
			continue
		}

		file := fileShares[0].File

		// Generate presigned URL for download
		downloadURL, err := u.S3Repo.GeneratePresignedDownloadURL(ctx, file.S3Key, 1*time.Hour)
		if err != nil {
			downloadURL = "" // Default to empty on error
		}

		// Generate thumbnail URL if available
		var thumbnailURL *string
		if file.ThumbnailKey != "" {
			url, err := u.S3Repo.GeneratePresignedDownloadURL(ctx, file.ThumbnailKey, 1*time.Hour)
			if err == nil {
				thumbnailURL = &url
			}
		}

		// Build shared_with array
		sharedWith := make([]response.ShareUserDTO, 0, len(fileShares))
		for _, share := range fileShares {
			if share.SharedWith != nil {
				sharedWith = append(sharedWith, response.ShareUserDTO{
					ID:       share.SharedWith.ID,
					Name:     share.SharedWith.Name,
					Email:    share.SharedWith.Email,
					SharedAt: share.CreatedAt,
				})
			}
		}

		files = append(files, response.FileSharedByMeDTO{
			ID:           fileID,
			Name:         file.FileName,
			Type:         file.ContentType,
			Size:         file.FileSize,
			URL:          downloadURL,
			ThumbnailURL: thumbnailURL,
			CreatedAt:    file.CreatedAt,
			SharedWith:   sharedWith,
		})
	}

	return &response.FilesSharedByMeResponseDTO{
		Files: files,
	}, nil
}
