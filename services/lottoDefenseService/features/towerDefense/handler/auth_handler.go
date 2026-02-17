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
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	resp, err := h.uc.Register(ctx, &req)
	if err != nil {
		if err.Error() == "email already exists" || err.Error() == "username already exists" {
			return c.JSON(http.StatusConflict, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"success": true,
		"data":    resp,
	})
}

func (h *TDAuthHandler) Login(c echo.Context) error {
	ctx := c.Request().Context()

	var req request.LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	resp, err := h.uc.Login(ctx, &req)
	if err != nil {
		if err.Error() == "invalid credentials" {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    resp,
	})
}
