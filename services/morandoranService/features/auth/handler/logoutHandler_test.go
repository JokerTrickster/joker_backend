package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JokerTrickster/joker_backend/shared/utils"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupLogoutEcho() *echo.Echo {
	e := echo.New()
	e.Validator = utils.NewValidator()
	return e
}

func TestLogoutHandler_Logout(t *testing.T) {
	e := setupLogoutEcho()
	h := NewLogoutHandler()

	req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	t.Logf("Calling Logout handler (stateless, no usecase)")
	err := h.Logout(c)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec.Code, "Logout should return 200")
	assert.Contains(t, rec.Body.String(), "message", "Response should have message field")
	assert.Contains(t, rec.Body.String(), "로그아웃", "Response should contain Korean logout message")
	t.Logf("Response: status=%d body=%s", rec.Code, rec.Body.String())
}
