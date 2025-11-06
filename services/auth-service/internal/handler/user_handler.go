package handler

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/luxrobo/joker_backend/services/auth-service/internal/model"
	"github.com/luxrobo/joker_backend/services/auth-service/internal/service"
	"github.com/luxrobo/joker_backend/shared/database"
	customErrors "github.com/luxrobo/joker_backend/shared/errors"
	"github.com/luxrobo/joker_backend/shared/logger"
	"github.com/luxrobo/joker_backend/shared/response"
	"go.uber.org/zap"
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
		return customErrors.InvalidInput("Invalid user ID format")
	}

	if id <= 0 {
		return customErrors.InvalidInput("User ID must be a positive number")
	}

	user, err := h.service.GetUserByID(c.Request().Context(), id)
	if err != nil {
		logger.Error("Failed to get user",
			zap.Int64("user_id", id),
			zap.Error(err),
		)
		return customErrors.InternalServerError("Failed to retrieve user")
	}

	if user == nil {
		return customErrors.ResourceNotFound("User")
	}

	return c.JSON(http.StatusOK, response.Success(user, "User retrieved successfully"))
}

func (h *UserHandler) CreateUser(c echo.Context) error {
	var req model.CreateUserRequest
	if err := c.Bind(&req); err != nil {
		return customErrors.BadRequest("Invalid request data")
	}

	if req.Email == "" || req.Name == "" {
		return customErrors.ValidationError("Email and name are required")
	}

	user, err := h.service.CreateUser(c.Request().Context(), &req)
	if err != nil {
		logger.Error("Failed to create user",
			zap.String("email", req.Email),
			zap.Error(err),
		)
		return customErrors.InternalServerError("Failed to create user")
	}

	return c.JSON(http.StatusCreated, response.Success(user, "User created successfully"))
}
