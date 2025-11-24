package _interface

import (
	"context"

	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/request"
	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/response"
)

type IUploadCloudRepositoryUseCase interface {
	RequestUploadURL(ctx context.Context, userID uint, req *request.UploadRequestDTO) (*response.UploadResponseDTO, error)
}

type IBatchUploadCloudRepositoryUseCase interface {
	RequestBatchUploadURL(ctx context.Context, userID uint, req *request.BatchUploadRequestDTO) (*response.BatchUploadResponseDTO, error)
}

type IDownloadCloudRepositoryUseCase interface {
	RequestDownloadURL(ctx context.Context, userID uint, fileID uint) (*response.DownloadResponseDTO, error)
}

type IListCloudRepositoryUseCase interface {
	ListFiles(ctx context.Context, userID uint, req request.ListFilesRequestDTO) (*response.ListFilesResponseDTO, error)
}

type IDeleteCloudRepositoryUseCase interface {
	DeleteFile(ctx context.Context, userID uint, fileID uint) error
}
