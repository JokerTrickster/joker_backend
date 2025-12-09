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
	ownerID uint,
	req *request.ShareFileRequestDTO,
) (*response.ShareFileResponseDTO, error) {
	ctx, cancel := context.WithTimeout(c, u.ContextTimeout)
	defer cancel()

	// Verify file exists and user has access
	hasAccess, err := u.FileShareRepo.HasFileAccess(ctx, ownerID, fileID)
	if err != nil || !hasAccess {
		return nil, fmt.Errorf("file not found or access denied")
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

		share := &entity.FileShare{
			FileID:       fileID,
			OwnerID:      ownerID,
			SharedWithID: user.ID,
		}

		// Create share (will ignore if already exists due to unique constraint)
		err := u.FileShareRepo.CreateFileShare(ctx, share)
		if err != nil {
			// Log error but continue with other users
			fmt.Printf("Failed to share file with user %d: %v\n", user.ID, err)
			continue
		}

		sharedUsers = append(sharedUsers, response.ShareUserDTO{
			ID:       user.ID,
			Name:     user.Name,
			Email:    user.Email,
			SharedAt: share.CreatedAt,
		})
	}

	message := fmt.Sprintf("%d명의 사용자와 파일을 공유했습니다", len(sharedUsers))

	return &response.ShareFileResponseDTO{
		Message:     message,
		SharedUsers: sharedUsers,
	}, nil
}

// GetFileShares retrieves all shares for a file
func (u *FileShareUseCase) GetFileShares(
	c context.Context,
	fileID uint,
	ownerID uint,
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
	sharedWithID uint,
	ownerID uint,
) (*response.RevokeShareResponseDTO, error) {
	ctx, cancel := context.WithTimeout(c, u.ContextTimeout)
	defer cancel()

	// Verify file exists and user has access
	hasAccess, err := u.FileShareRepo.HasFileAccess(ctx, ownerID, fileID)
	if err != nil || !hasAccess {
		return nil, fmt.Errorf("file not found or access denied")
	}

	// Delete the share
	err = u.FileShareRepo.DeleteFileShare(ctx, fileID, sharedWithID, ownerID)
	if err != nil {
		return nil, fmt.Errorf("failed to revoke file share: %w", err)
	}

	return &response.RevokeShareResponseDTO{
		Message: "공유가 취소되었습니다",
	}, nil
}

// GetSharedWithMeFiles retrieves all files shared with the current user
func (u *FileShareUseCase) GetSharedWithMeFiles(
	c context.Context,
	userID uint,
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
			if share.File.ThumbnailKey != nil && *share.File.ThumbnailKey != "" {
				url, err := u.S3Repo.GeneratePresignedDownloadURL(ctx, *share.File.ThumbnailKey, 1*time.Hour)
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
				FileType:     share.File.FileType,
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
