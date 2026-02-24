package handler

import (
	"net/http"

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
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	resp, err := h.uc.CreateRoom(ctx, userID, &req)
	if err != nil {
		return errorResponse(c, http.StatusInternalServerError, err.Error())
	}

	return successResponse(c, http.StatusCreated, resp)
}

func (h *TDRoomHandler) JoinRoom(c echo.Context) error {
	ctx := c.Request().Context()
	
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return err
	}

	var req request.JoinRoomRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	resp, err := h.uc.JoinRoom(ctx, userID, &req)
	if err != nil {
		return h.handleRoomError(c, err)
	}

	return successResponse(c, http.StatusOK, resp)
}

func (h *TDRoomHandler) GetRoom(c echo.Context) error {
	ctx := c.Request().Context()

	roomID, err := getParamUint(c, "id")
	if err != nil {
		return err
	}

	resp, err := h.uc.GetRoom(ctx, roomID)
	if err != nil {
		return errorResponse(c, http.StatusInternalServerError, err.Error())
	}

	return successResponse(c, http.StatusOK, resp)
}

func (h *TDRoomHandler) LeaveRoom(c echo.Context) error {
	ctx := c.Request().Context()
	
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return err
	}

	roomID, err := getParamUint(c, "id")
	if err != nil {
		return err
	}

	if err := h.uc.LeaveRoom(ctx, userID, roomID); err != nil {
		return errorResponse(c, http.StatusInternalServerError, err.Error())
	}

	return successResponse(c, http.StatusOK, map[string]string{
		"message": "left room successfully",
	})
}

func (h *TDRoomHandler) SetReady(c echo.Context) error {
	ctx := c.Request().Context()
	
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return err
	}

	roomID, err := getParamUint(c, "id")
	if err != nil {
		return err
	}

	var req request.SetReadyRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	resp, err := h.uc.SetReady(ctx, userID, roomID, req.IsReady)
	if err != nil {
		return errorResponse(c, http.StatusInternalServerError, err.Error())
	}

	return successResponse(c, http.StatusOK, resp)
}

func (h *TDRoomHandler) handleRoomError(c echo.Context, err error) error {
	errMsg := err.Error()
	
	switch errMsg {
	case "room not found":
		return errorResponse(c, http.StatusNotFound, errMsg)
	case "room is full", "room already started":
		return errorResponse(c, http.StatusConflict, errMsg)
	default:
		return errorResponse(c, http.StatusInternalServerError, errMsg)
	}
}
