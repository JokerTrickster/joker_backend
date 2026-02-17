package handler

import (
	"net/http"
	"strconv"

	_interface "github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/towerDefense/model/interface"
	"github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/towerDefense/model/request"
	"github.com/labstack/echo/v4"
)

type TDRoomHandler struct {
	uc _interface.ITDRoomUseCase
}

func NewTDRoomHandler(g *echo.Group, uc _interface.ITDRoomUseCase) {
	h := &TDRoomHandler{uc: uc}
	g.POST("/coop/rooms", h.CreateRoom)
	g.POST("/coop/rooms/join", h.JoinRoom)
	g.GET("/coop/rooms/:id", h.GetRoom)
	g.POST("/coop/rooms/:id/leave", h.LeaveRoom)
	g.POST("/coop/rooms/:id/ready", h.SetReady)
}

func (h *TDRoomHandler) CreateRoom(c echo.Context) error {
	ctx := c.Request().Context()
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return err
	}

	var req request.CreateRoomRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	resp, err := h.uc.CreateRoom(ctx, userID, &req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"success": true,
		"data":    resp,
	})
}

func (h *TDRoomHandler) JoinRoom(c echo.Context) error {
	ctx := c.Request().Context()
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return err
	}

	var req request.JoinRoomRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	resp, err := h.uc.JoinRoom(ctx, userID, &req)
	if err != nil {
		if err.Error() == "room not found" {
			return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		}
		if err.Error() == "room is full" || err.Error() == "room already started" {
			return c.JSON(http.StatusConflict, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    resp,
	})
}

func (h *TDRoomHandler) GetRoom(c echo.Context) error {
	ctx := c.Request().Context()

	idStr := c.Param("id")
	roomID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid room id"})
	}

	resp, err := h.uc.GetRoom(ctx, uint(roomID))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    resp,
	})
}

func (h *TDRoomHandler) LeaveRoom(c echo.Context) error {
	ctx := c.Request().Context()
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return err
	}

	idStr := c.Param("id")
	roomID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid room id"})
	}

	if err := h.uc.LeaveRoom(ctx, userID, uint(roomID)); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "left room successfully",
	})
}

func (h *TDRoomHandler) SetReady(c echo.Context) error {
	ctx := c.Request().Context()
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return err
	}

	idStr := c.Param("id")
	roomID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid room id"})
	}

	var req request.SetReadyRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	resp, err := h.uc.SetReady(ctx, userID, uint(roomID), req.IsReady)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    resp,
	})
}
