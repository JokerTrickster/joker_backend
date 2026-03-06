package handler

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/JokerTrickster/joker_backend/services/molandolanService/features/auth/model/entity"
	_interface "github.com/JokerTrickster/joker_backend/services/molandolanService/features/auth/model/interface"
	"github.com/JokerTrickster/joker_backend/services/molandolanService/features/auth/usecase"
	"github.com/JokerTrickster/joker_backend/shared/utils"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockUpdateMeAuthRepository struct {
	updateNicknameFunc func(ctx context.Context, userID uint, nickname string) (*entity.MorandoranUser, error)
}

func (m *mockUpdateMeAuthRepository) FindUserByEmail(ctx context.Context, email string) (*entity.MorandoranUser, error) {
	return nil, nil
}
func (m *mockUpdateMeAuthRepository) FindUserByID(ctx context.Context, userID uint) (*entity.MorandoranUser, error) {
	return nil, nil
}
func (m *mockUpdateMeAuthRepository) FindOrCreateByOAuth(ctx context.Context, email, nickname, provider string, profileImage *string) (*entity.MorandoranUser, error) {
	return nil, nil
}
func (m *mockUpdateMeAuthRepository) UpdateNickname(ctx context.Context, userID uint, nickname string) (*entity.MorandoranUser, error) {
	if m.updateNicknameFunc != nil {
		return m.updateNicknameFunc(ctx, userID, nickname)
	}
	return nil, nil
}

var _ _interface.IAuthRepository = (*mockUpdateMeAuthRepository)(nil)

func TestUpdateMeHandler_Success(t *testing.T) {
	t.Log("UpdateMe: valid nickname -> 200 OK")
	e := echo.New()
	e.Validator = utils.NewValidator()

	pic := "https://example.com/pic.jpg"
	mockRepo := &mockUpdateMeAuthRepository{
		updateNicknameFunc: func(ctx context.Context, userID uint, nickname string) (*entity.MorandoranUser, error) {
			return &entity.MorandoranUser{
				ID:           userID,
				Nickname:     nickname,
				Email:        "user@example.com",
				Role:         "user",
				Provider:     "google",
				ProfileImage: &pic,
			}, nil
		},
	}

	uc := usecase.NewUpdateMeUseCase(mockRepo, 10*time.Second)
	h := NewUpdateMeHandler(uc)

	body := strings.NewReader(`{"nickname":"newnick"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/auth/me", body)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("userID", uint(1))

	err := h.UpdateMe(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "newnick")
	assert.Contains(t, rec.Body.String(), "google")
	assert.Contains(t, rec.Body.String(), "profileImage")
	t.Logf("Response: %s", rec.Body.String())
}

func TestUpdateMeHandler_Unauthorized(t *testing.T) {
	t.Log("UpdateMe: no userID -> 401")
	e := echo.New()
	e.Validator = utils.NewValidator()
	mockRepo := &mockUpdateMeAuthRepository{}
	uc := usecase.NewUpdateMeUseCase(mockRepo, 10*time.Second)
	h := NewUpdateMeHandler(uc)

	body := strings.NewReader(`{"nickname":"newnick"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/auth/me", body)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.UpdateMe(c)
	require.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusUnauthorized, he.Code)
	t.Logf("Error: code=%d", he.Code)
}

func TestUpdateMeHandler_BadRequest_TooShort(t *testing.T) {
	t.Log("UpdateMe: nickname too short -> 400")
	e := echo.New()
	e.Validator = utils.NewValidator()
	mockRepo := &mockUpdateMeAuthRepository{}
	uc := usecase.NewUpdateMeUseCase(mockRepo, 10*time.Second)
	h := NewUpdateMeHandler(uc)

	body := strings.NewReader(`{"nickname":"a"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/auth/me", body)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("userID", uint(1))

	err := h.UpdateMe(c)
	require.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, he.Code)
	t.Logf("Error: code=%d message=%v", he.Code, he.Message)
}

func TestUpdateMeHandler_BadRequest_Space(t *testing.T) {
	t.Log("UpdateMe: nickname with space -> 400")
	e := echo.New()
	e.Validator = utils.NewValidator()
	mockRepo := &mockUpdateMeAuthRepository{}
	uc := usecase.NewUpdateMeUseCase(mockRepo, 10*time.Second)
	h := NewUpdateMeHandler(uc)

	body := strings.NewReader(`{"nickname":"bad name"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/auth/me", body)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("userID", uint(1))

	err := h.UpdateMe(c)
	require.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, he.Code)
	assert.Contains(t, fmt.Sprintf("%v", he.Message), "공백")
	t.Logf("Error: code=%d message=%v", he.Code, he.Message)
}

func TestUpdateMeHandler_NotFound(t *testing.T) {
	t.Log("UpdateMe: usecase error -> 404")
	e := echo.New()
	e.Validator = utils.NewValidator()
	mockRepo := &mockUpdateMeAuthRepository{
		updateNicknameFunc: func(ctx context.Context, userID uint, nickname string) (*entity.MorandoranUser, error) {
			return nil, fmt.Errorf("user not found")
		},
	}
	uc := usecase.NewUpdateMeUseCase(mockRepo, 10*time.Second)
	h := NewUpdateMeHandler(uc)

	body := strings.NewReader(`{"nickname":"newnick"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/auth/me", body)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("userID", uint(999))

	err := h.UpdateMe(c)
	require.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusNotFound, he.Code)
	t.Logf("Error: code=%d", he.Code)
}

func TestUpdateMeHandler_BadRequest_InvalidJSON(t *testing.T) {
	t.Log("UpdateMe: invalid JSON body -> 400 BAD_REQUEST")
	e := echo.New()
	e.Validator = utils.NewValidator()
	mockRepo := &mockUpdateMeAuthRepository{}
	uc := usecase.NewUpdateMeUseCase(mockRepo, 10*time.Second)
	h := NewUpdateMeHandler(uc)

	body := strings.NewReader(`{invalid json}`)
	req := httptest.NewRequest(http.MethodPut, "/api/auth/me", body)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("userID", uint(1))

	err := h.UpdateMe(c)
	require.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, he.Code)
	assert.Equal(t, "BAD_REQUEST", he.Message)
	t.Logf("Error: code=%d message=%v", he.Code, he.Message)
}

func TestUpdateMeHandler_BadRequest_EmptyNickname(t *testing.T) {
	t.Log("UpdateMe: empty nickname fails validation -> 400")
	e := echo.New()
	e.Validator = utils.NewValidator()
	mockRepo := &mockUpdateMeAuthRepository{}
	uc := usecase.NewUpdateMeUseCase(mockRepo, 10*time.Second)
	h := NewUpdateMeHandler(uc)

	body := strings.NewReader(`{"nickname":""}`)
	req := httptest.NewRequest(http.MethodPut, "/api/auth/me", body)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("userID", uint(1))

	err := h.UpdateMe(c)
	require.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, he.Code)
	t.Logf("Error: code=%d message=%v", he.Code, he.Message)
}

func TestUpdateMeHandler_BadRequest_TooLong(t *testing.T) {
	t.Log("UpdateMe: nickname exceeds max length (20) -> 400")
	e := echo.New()
	e.Validator = utils.NewValidator()
	mockRepo := &mockUpdateMeAuthRepository{}
	uc := usecase.NewUpdateMeUseCase(mockRepo, 10*time.Second)
	h := NewUpdateMeHandler(uc)

	body := strings.NewReader(`{"nickname":"this_nickname_is_way_too_long"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/auth/me", body)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("userID", uint(1))

	err := h.UpdateMe(c)
	require.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, he.Code)
	t.Logf("Error: code=%d message=%v", he.Code, he.Message)
}

func TestUpdateMeHandler_BadRequest_MissingNickname(t *testing.T) {
	t.Log("UpdateMe: missing nickname field -> 400")
	e := echo.New()
	e.Validator = utils.NewValidator()
	mockRepo := &mockUpdateMeAuthRepository{}
	uc := usecase.NewUpdateMeUseCase(mockRepo, 10*time.Second)
	h := NewUpdateMeHandler(uc)

	body := strings.NewReader(`{}`)
	req := httptest.NewRequest(http.MethodPut, "/api/auth/me", body)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("userID", uint(1))

	err := h.UpdateMe(c)
	require.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, he.Code)
	t.Logf("Error: code=%d message=%v", he.Code, he.Message)
}
