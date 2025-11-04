package handler

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/luxrobo/joker_backend/internal/model"
	"github.com/luxrobo/joker_backend/internal/service"
	"github.com/luxrobo/joker_backend/pkg/database"
	"github.com/luxrobo/joker_backend/pkg/response"
)

type UserHandler struct {
	service *service.UserService
}

func NewUserHandler(db *database.DB) *UserHandler {
	return &UserHandler{
		service: service.NewUserService(db),
	}
}

func (h *UserHandler) GetUser(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Error("INVALID_PARAMS", "Invalid user ID"))
	}

	user, err := h.service.GetUserByID(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Error("INTERNAL_ERROR", err.Error()))
	}

	if user == nil {
		return c.JSON(http.StatusNotFound, response.Error("NOT_FOUND", "User not found"))
	}

	return c.JSON(http.StatusOK, response.Success(user, "User retrieved successfully"))
}

func (h *UserHandler) CreateUser(c echo.Context) error {
	var req model.CreateUserRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Error("INVALID_DATA", "Invalid request data"))
	}

	if req.Email == "" || req.Name == "" {
		return c.JSON(http.StatusBadRequest, response.Error("INVALID_DATA", "Email and name are required"))
	}

	user, err := h.service.CreateUser(&req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Error("INTERNAL_ERROR", err.Error()))
	}

	return c.JSON(http.StatusCreated, response.Success(user, "User created successfully"))
}
