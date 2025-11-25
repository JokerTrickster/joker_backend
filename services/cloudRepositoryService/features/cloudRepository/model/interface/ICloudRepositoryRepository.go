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
