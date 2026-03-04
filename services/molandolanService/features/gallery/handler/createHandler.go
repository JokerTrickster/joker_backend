package handler

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/JokerTrickster/joker_backend/services/molandolanService/features/gallery/model/request"
	"github.com/JokerTrickster/joker_backend/services/molandolanService/features/gallery/usecase"
	"github.com/JokerTrickster/joker_backend/shared/aws"
	"github.com/JokerTrickster/joker_backend/shared/utils"
	"github.com/labstack/echo/v4"
)

var allowedImageExts = map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".webp": true}
var allowedVideoExts = map[string]bool{".mp4": true, ".webm": true}

const maxImageSize = 10 * 1024 * 1024  // 10MB
const maxVideoSize = 50 * 1024 * 1024  // 50MB

type CreateHandler struct {
	UseCase *usecase.CreateUseCase
}

func NewCreateHandler(uc *usecase.CreateUseCase) *CreateHandler {
	return &CreateHandler{UseCase: uc}
}

func (h *CreateHandler) Create(c echo.Context) error {
	ctx := c.Request().Context()
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "UNAUTHORIZED")
	}

	file, err := c.FormFile("file")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "file is required")
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	var mediaType string
	if allowedImageExts[ext] {
		mediaType = "image"
		if file.Size > maxImageSize {
			return echo.NewHTTPError(http.StatusRequestEntityTooLarge, "FILE_TOO_LARGE")
		}
	} else if allowedVideoExts[ext] {
		mediaType = "video"
		if file.Size > maxVideoSize {
			return echo.NewHTTPError(http.StatusRequestEntityTooLarge, "FILE_TOO_LARGE")
		}
	} else {
		return echo.NewHTTPError(http.StatusBadRequest, "unsupported file format")
	}

	caption := c.FormValue("caption")
	if len(caption) > 500 {
		return echo.NewHTTPError(http.StatusBadRequest, "caption must be 500 characters or less")
	}

	filename := fmt.Sprintf("%s%s", aws.FileNameGenerateRandom(), ext)
	contentType := file.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	mediaURL, err := aws.RawFileUpload(ctx, file, filename, contentType, aws.ImgGallery)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "INTERNAL_ERROR")
	}

	var thumbnailURL string
	if mediaType == "image" {
		thumbURL, err := aws.ThumbnailUpload(ctx, file, filename, aws.ImgGallery)
		if err != nil {
			thumbnailURL = mediaURL
		} else {
			thumbnailURL = thumbURL
		}
	} else {
		thumbnailURL = mediaURL
	}

	req := &request.ReqCreateGallery{
		MediaURL:     mediaURL,
		ThumbnailURL: thumbnailURL,
		MediaType:    mediaType,
		Caption:      caption,
	}

	res, err := h.UseCase.Create(ctx, userID, req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "INTERNAL_ERROR")
	}
	return c.JSON(http.StatusCreated, res)
}
