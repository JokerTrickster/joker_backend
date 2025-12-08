package _interface

import (
	"github.com/labstack/echo/v4"
)

type IUploadCloudRepositoryHandler interface {
	RequestUploadURL(c echo.Context) error
}

type IBatchUploadCloudRepositoryHandler interface {
	RequestBatchUploadURL(c echo.Context) error
}

type IDownloadCloudRepositoryHandler interface {
	RequestDownloadURL(c echo.Context) error
}

type IListCloudRepositoryHandler interface {
	ListFiles(c echo.Context) error
}

type IDeleteCloudRepositoryHandler interface {
	DeleteFile(c echo.Context) error
}

type ITagHandler interface {
	UpdateFileTags(c echo.Context) error
	AddTagToFile(c echo.Context) error
	RemoveTagFromFile(c echo.Context) error
}

type IMultipartUploadHandler interface {
	InitiateMultipartUpload(c echo.Context) error
	GeneratePresignedURLs(c echo.Context) error
	CompleteMultipartUpload(c echo.Context) error
	AbortMultipartUpload(c echo.Context) error
}

type IFolderHandler interface {
	CreateFolder(c echo.Context) error
	GetFolders(c echo.Context) error
	GetFolderByID(c echo.Context) error
	UpdateFolder(c echo.Context) error
	DeleteFolder(c echo.Context) error
	GetFolderFiles(c echo.Context) error
	MoveFiles(c echo.Context) error
}
