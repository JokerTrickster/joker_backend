package handler

import (
	"net/http"

	"github.com/JokerTrickster/joker_backend/services/molandolanService/features/upload/usecase"
	"github.com/labstack/echo/v4"
)

type UploadHandler struct {
	UseCase *usecase.UploadUseCase
}

func NewUploadHandler(uc *usecase.UploadUseCase) *UploadHandler {
	return &UploadHandler{UseCase: uc}
}

func (h *UploadHandler) Upload(c echo.Context) error {
	ctx := c.Request().Context()

	file, err := c.FormFile("file")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "file is required")
	}

	fileType := c.FormValue("type")
	if fileType == "" {
		fileType = "general"
	}

	res, err := h.UseCase.Upload(ctx, file, fileType)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, res)
}
