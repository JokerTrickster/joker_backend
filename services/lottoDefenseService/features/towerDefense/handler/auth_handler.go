package handler

import (
	"net/http"

	_interface "github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/towerDefense/model/interface"
	"github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/towerDefense/model/request"
	"github.com/labstack/echo/v4"
)

type TDAuthHandler struct {
	uc _interface.ITDAuthUseCase
}

func NewTDAuthHandler(g *echo.Group, uc _interface.ITDAuthUseCase) {
	h := &TDAuthHandler{uc: uc}
	g.POST("/auth/register", h.Register)
	g.POST("/auth/login", h.Login)
}

func (h *TDAuthHandler) Register(c echo.Context) error {
	ctx := c.Request().Context()

	var req request.RegisterRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	resp, err := h.uc.Register(ctx, &req)
	if err != nil {
		return h.handleAuthError(c, err)
	}

	return successResponse(c, http.StatusCreated, resp)
}

func (h *TDAuthHandler) Login(c echo.Context) error {
	ctx := c.Request().Context()

	var req request.LoginRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	resp, err := h.uc.Login(ctx, &req)
	if err != nil {
		return h.handleAuthError(c, err)
	}

	return successResponse(c, http.StatusOK, resp)
}

func (h *TDAuthHandler) handleAuthError(c echo.Context, err error) error {
	errMsg := err.Error()
	
	switch errMsg {
	case "email already exists", "username already exists":
		return errorResponse(c, http.StatusConflict, errMsg)
	case "invalid credentials":
		return errorResponse(c, http.StatusUnauthorized, errMsg)
	default:
		return errorResponse(c, http.StatusInternalServerError, errMsg)
	}
}
