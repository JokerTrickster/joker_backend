package handler

import (
	"context"
	"net/http"

	_interface "github.com/JokerTrickster/joker_backend/services/authService/features/auth/model/interface"
	"github.com/JokerTrickster/joker_backend/services/authService/features/auth/model/request"
	_ "github.com/JokerTrickster/joker_backend/services/authService/features/auth/model/response"
	"github.com/labstack/echo/v4"
)

type CheckEmailAuthHandler struct {
	UseCase _interface.ICheckEmailAuthUseCase
}

func NewCheckEmailAuthHandler(c *echo.Echo, useCase _interface.ICheckEmailAuthUseCase) _interface.ICheckEmailAuthHandler {
	handler := &CheckEmailAuthHandler{
		UseCase: useCase,
	}
	c.POST("/v0.1/auth/check-email", handler.CheckEmail)
	return handler
}

// 이메일 중복 체크
// @Router /v0.1/auth/check-email [post]
// @Summary 이메일 중복 체크
// @Description 회원가입 전 이메일이 이미 등록되어 있는지 확인합니다.
// @Description
// @Description ■ errCode with 400
// @Description PARAM_BAD : 파라미터 오류
// @Description
// @Description ■ errCode with 500
// @Description INTERNAL_SERVER : 내부 로직 처리 실패
// @Description INTERNAL_DB : DB 처리 실패
// @Param json body request.ReqCheckEmail true "이메일과 제공자 타입"
// @Produce json
// @Success 200 {object} response.ResCheckEmail
// @Failure 400 {object} error
// @Failure 500 {object} error
// @Tags auth
func (h *CheckEmailAuthHandler) CheckEmail(c echo.Context) error {
	ctx := context.Background()

	req := &request.ReqCheckEmail{}
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	res, err := h.UseCase.CheckEmail(ctx, req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, res)
}
