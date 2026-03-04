package handler

import (
	"net/http"

	"github.com/JokerTrickster/joker_backend/services/morandoranService/features/auth/model/request"
	"github.com/JokerTrickster/joker_backend/services/morandoranService/features/auth/usecase"
	"github.com/labstack/echo/v4"
)

type LoginHandler struct {
	UseCase *usecase.LoginUseCase
}

func NewLoginHandler(uc *usecase.LoginUseCase) *LoginHandler {
	return &LoginHandler{UseCase: uc}
}

func (h *LoginHandler) Login(c echo.Context) error {
	ctx := c.Request().Context()
	req := &request.ReqLogin{}
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "BAD_REQUEST")
	}
	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	res, err := h.UseCase.Login(ctx, req)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "UNAUTHORIZED")
	}
	return c.JSON(http.StatusOK, res)
}
