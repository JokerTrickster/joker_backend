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

type IUserStatsCloudRepositoryUseCase interface {
	GetUserStats(ctx context.Context, userID uint) (*response.UserStatsResponseDTO, error)
}

type IActivityHistoryCloudRepositoryUseCase interface {
	GetActivityHistory(ctx context.Context, userID uint, req *request.ActivityHistoryRequestDTO) (*response.ActivityHistoryResponseDTO, error)
}

type IProcessingStatusUseCase interface {
	GetProcessingStatus(ctx context.Context, userID uint, fileID uint) (*response.ProcessingStatusResponse, error)
	GetBatchProcessingStatus(ctx context.Context, userID uint, fileIDs []uint) (*response.BatchProcessingStatusResponse, error)
}

type ITagUseCase interface {
	UpdateFileTags(ctx context.Context, userID uint, fileID uint, req *request.UpdateFileTagsRequestDTO) (*response.UpdateFileTagsResponseDTO, error)
	AddTagToFile(ctx context.Context, userID uint, fileID uint, req *request.AddFileTagRequestDTO) (*response.AddFileTagResponseDTO, error)
	RemoveTagFromFile(ctx context.Context, userID uint, fileID uint, tagName string) error
}

type IMultipartUploadUseCase interface {
	InitiateMultipartUpload(ctx context.Context, userID uint, req *request.InitiateMultipartUploadRequestDTO) (*response.InitiateMultipartUploadResponseDTO, error)
	GeneratePresignedURLs(ctx context.Context, userID uint, req *request.GeneratePresignedURLsRequestDTO) (*response.GeneratePresignedURLsResponseDTO, error)
	CompleteMultipartUpload(ctx context.Context, userID uint, req *request.CompleteMultipartUploadRequestDTO) (*response.CompleteMultipartUploadResponseDTO, error)
	AbortMultipartUpload(ctx context.Context, userID uint, req *request.AbortMultipartUploadRequestDTO) (*response.AbortMultipartUploadResponseDTO, error)
}

type IFolderUseCase interface {
	CreateFolder(ctx context.Context, userID uint, req *request.CreateFolderRequestDTO) (*response.FolderResponseDTO, error)
	GetFolders(ctx context.Context, userID uint) ([]response.FolderTreeResponseDTO, error)
	GetFolderByID(ctx context.Context, folderID uint, userID uint) (*response.FolderResponseDTO, error)
	UpdateFolder(ctx context.Context, folderID uint, userID uint, req *request.UpdateFolderRequestDTO) (*response.FolderResponseDTO, error)
	DeleteFolder(ctx context.Context, folderID uint, userID uint) error
	GetFolderFiles(ctx context.Context, folderID uint, userID uint) ([]response.FileInfoDTO, error)
	MoveFiles(ctx context.Context, userID uint, req *request.MoveFilesToFolderRequestDTO) (*response.MoveFilesResponseDTO, error)
}

type IFolderShareUseCase interface {
	ShareFolder(ctx context.Context, folderID uint, ownerID int32, req *request.ShareFolderRequestDTO) (*response.ShareFolderResponseDTO, error)
	GetFolderShares(ctx context.Context, folderID uint, ownerID int32) (*response.FolderShareListResponseDTO, error)
	RevokeFolderShare(ctx context.Context, folderID uint, sharedWithID int32, ownerID int32) (*response.RevokeShareResponseDTO, error)
	GetSharedWithMeFolders(ctx context.Context, userID int32) (*response.SharedWithMeFoldersResponseDTO, error)
	GetFoldersSharedByMe(ctx context.Context, ownerID int32) (*response.FoldersSharedByMeResponseDTO, error)
}

type IFileShareUseCase interface {
	ShareFile(ctx context.Context, fileID uint, ownerID int32, req *request.ShareFileRequestDTO) (*response.ShareFileResponseDTO, error)
	GetFileShares(ctx context.Context, fileID uint, ownerID int32) (*response.FileShareListResponseDTO, error)
	RevokeFileShare(ctx context.Context, fileID uint, sharedWithID int32, ownerID int32) (*response.RevokeShareResponseDTO, error)
	GetSharedWithMeFiles(ctx context.Context, userID int32) (*response.SharedWithMeFilesResponseDTO, error)
	GetFilesSharedByMe(ctx context.Context, ownerID int32) (*response.FilesSharedByMeResponseDTO, error)
}
