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
