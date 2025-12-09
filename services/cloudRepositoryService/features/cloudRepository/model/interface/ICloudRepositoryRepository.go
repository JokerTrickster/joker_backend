package _interface

import (
	"context"
	"time"

	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/entity"
	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/request"
)

type IUploadCloudRepositoryRepository interface {
	GeneratePresignedUploadURL(ctx context.Context, s3Key, contentType string, expiration time.Duration) (string, error)
	CreateFile(ctx context.Context, file *entity.CloudFile) error
}

type IBatchUploadCloudRepositoryRepository interface {
	GeneratePresignedUploadURL(ctx context.Context, s3Key, contentType string, expiration time.Duration) (string, error)
	CreateFile(ctx context.Context, file *entity.CloudFile) error
}

type IDownloadCloudRepositoryRepository interface {
	GeneratePresignedDownloadURL(ctx context.Context, s3Key string, expiration time.Duration) (string, error)
	GeneratePresignedDownloadURLWithFilename(ctx context.Context, s3Key, filename string, expiration time.Duration) (string, error)
	GetFileByID(ctx context.Context, id uint) (*entity.CloudFile, error)
}

type IListCloudRepositoryRepository interface {
	GetFilesByUserID(ctx context.Context, userID uint, filter request.ListFilesRequestDTO) ([]entity.CloudFile, int64, error)
	GeneratePresignedDownloadURL(ctx context.Context, s3Key string, expiration time.Duration) (string, error)
}

type IDeleteCloudRepositoryRepository interface {
	DeleteFromS3(ctx context.Context, s3Key string) error
	SoftDeleteFile(ctx context.Context, id uint, userID uint) error
	GetFileByID(ctx context.Context, id uint) (*entity.CloudFile, error)
}

type IUserStatsCloudRepositoryRepository interface {
	GetTotalStorageUsed(ctx context.Context, userID uint) (int64, error)
	GetMonthlyUploadCount(ctx context.Context, userID uint, year int, month int) (int, error)
	GetMonthlyDownloadCount(ctx context.Context, userID uint, year int, month int) (int, error)
	GetMonthlyTagsCreatedCount(ctx context.Context, userID uint, year int, month int) (int, error)
	LogActivity(ctx context.Context, activity *entity.ActivityLog) error
}

type IActivityHistoryCloudRepositoryRepository interface {
	GetMonthlyActivity(ctx context.Context, userID uint, year int, month int) ([]entity.ActivityLog, error)
	GetMonthlyUsedTags(ctx context.Context, userID uint, year int, month int) (map[string][]string, error)
}

type ITagRepository interface {
	GetFileByID(ctx context.Context, id uint, userID uint) (*entity.CloudFile, error)
	UpdateFileTags(ctx context.Context, fileID uint, userID uint, tags []entity.Tag) error
	AddTagToFile(ctx context.Context, fileID uint, userID uint, tag entity.Tag) error
	RemoveTagFromFile(ctx context.Context, fileID uint, userID uint, tagName string) error
	FindOrCreateTag(ctx context.Context, userID uint, tagName string) (*entity.Tag, error)
}

type IMultipartUploadRepository interface {
	CreateMultipartUpload(ctx context.Context, bucket, key, contentType string) (string, error)
	GeneratePresignedUploadPartURL(ctx context.Context, bucket, key, uploadID string, partNumber int, expiration time.Duration) (string, error)
	CompleteMultipartUpload(ctx context.Context, bucket, key, uploadID string, parts []CompletedPart) error
	AbortMultipartUpload(ctx context.Context, bucket, key, uploadID string) error
	CreateMultipartUploadRecord(ctx context.Context, upload *entity.MultipartUpload) error
	UpdateMultipartUploadStatus(ctx context.Context, uploadID string, status entity.MultipartUploadStatus) error
	GetMultipartUpload(ctx context.Context, uploadID string, userID uint) (*entity.MultipartUpload, error)
}

type IFolderRepository interface {
	CreateFolder(ctx context.Context, folder *entity.Folder) error
	GetFolderByID(ctx context.Context, id uint, userID uint) (*entity.Folder, error)
	GetFoldersByUserID(ctx context.Context, userID uint) ([]entity.Folder, error)
	UpdateFolder(ctx context.Context, folder *entity.Folder) error
	DeleteFolder(ctx context.Context, id uint, userID uint) error
	GetFolderFileCount(ctx context.Context, folderID uint, userID uint) (int, error)
	GetFilesByFolderID(ctx context.Context, folderID *uint, userID uint) ([]entity.CloudFile, error)
	MoveFilesToFolder(ctx context.Context, fileIDs []uint, folderID *uint, userID uint) (int, error)
}

type IFolderShareRepository interface {
	CreateFolderShare(ctx context.Context, share *entity.FolderShare) error
	GetFolderSharesByFolderID(ctx context.Context, folderID uint) ([]entity.FolderShare, error)
	GetSharedFoldersByUserID(ctx context.Context, userID uint) ([]entity.FolderShare, error)
	DeleteFolderShare(ctx context.Context, folderID uint, sharedWithID uint, ownerID uint) error
	HasFolderAccess(ctx context.Context, userID uint, folderID uint) (bool, error)
	GetUsersByEmails(ctx context.Context, emails []string) ([]entity.User, error)
}

type IFileShareRepository interface {
	CreateFileShare(ctx context.Context, share *entity.FileShare) error
	GetFileSharesByFileID(ctx context.Context, fileID uint) ([]entity.FileShare, error)
	GetSharedFilesByUserID(ctx context.Context, userID uint) ([]entity.FileShare, error)
	DeleteFileShare(ctx context.Context, fileID uint, sharedWithID uint, ownerID uint) error
	HasFileAccess(ctx context.Context, userID uint, fileID uint) (bool, error)
}

type CompletedPart struct {
	PartNumber int
	ETag       string
}
