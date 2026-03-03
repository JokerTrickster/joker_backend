package handler

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/towerDefense/model/request"
	"github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/towerDefense/model/response"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockTDAuthUseCase struct {
	mock.Mock
}

func (m *mockTDAuthUseCase) Register(ctx context.Context, req *request.RegisterRequest) (*response.AuthResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.AuthResponse), args.Error(1)
}

func (m *mockTDAuthUseCase) Login(ctx context.Context, req *request.LoginRequest) (*response.AuthResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.AuthResponse), args.Error(1)
}

func (m *mockTDAuthUseCase) GetUserInfo(ctx context.Context, userID uint) (*response.UserInfoResponse, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.UserInfoResponse), args.Error(1)
}

func TestTDAuthHandler_Register_Success(t *testing.T) {
	t.Log("Register: success -> 201")
	e := setupTDTestEcho()
	mockUC := new(mockTDAuthUseCase)
	h := &TDAuthHandler{uc: mockUC}

	body := tdMustJSON(t, &request.RegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	})
	mockUC.On("Register", mock.Anything, mock.AnythingOfType("*request.RegisterRequest")).
		Return(&response.AuthResponse{
			User:  &response.UserData{ID: 1, Username: "testuser", Email: "test@example.com"},
			Token: "jwt-token-123",
		}, nil)

	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.Register(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Contains(t, rec.Body.String(), "success")
	mockUC.AssertExpectations(t)
}

func TestTDAuthHandler_Register_EmailExists(t *testing.T) {
	t.Log("Register: email exists -> 409")
	e := setupTDTestEcho()
	mockUC := new(mockTDAuthUseCase)
	h := &TDAuthHandler{uc: mockUC}

	body := tdMustJSON(t, &request.RegisterRequest{
		Username: "testuser",
		Email:    "existing@example.com",
		Password: "password123",
	})
	mockUC.On("Register", mock.Anything, mock.AnythingOfType("*request.RegisterRequest")).
		Return(nil, errors.New("email already exists"))

	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.Register(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusConflict, rec.Code)
	mockUC.AssertExpectations(t)
}

func TestTDAuthHandler_Register_UsernameExists(t *testing.T) {
	t.Log("Register: username exists -> 409")
	e := setupTDTestEcho()
	mockUC := new(mockTDAuthUseCase)
	h := &TDAuthHandler{uc: mockUC}

	body := tdMustJSON(t, &request.RegisterRequest{
		Username: "existinguser",
		Email:    "new@example.com",
		Password: "password123",
	})
	mockUC.On("Register", mock.Anything, mock.AnythingOfType("*request.RegisterRequest")).
		Return(nil, errors.New("username already exists"))

	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.Register(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusConflict, rec.Code)
	mockUC.AssertExpectations(t)
}

func TestTDAuthHandler_Register_ValidationError(t *testing.T) {
	t.Log("Register: validation error -> 400")
	e := setupTDTestEcho()
	mockUC := new(mockTDAuthUseCase)
	h := &TDAuthHandler{uc: mockUC}

	body := tdMustJSON(t, map[string]interface{}{
		"username": "ab",
		"email":    "invalid-email",
		"password": "12345",
	})
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.Register(c)
	require.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, he.Code)
	mockUC.AssertNotCalled(t, "Register")
}

func TestTDAuthHandler_Login_Success(t *testing.T) {
	t.Log("Login: success -> 200")
	e := setupTDTestEcho()
	mockUC := new(mockTDAuthUseCase)
	h := &TDAuthHandler{uc: mockUC}

	body := tdMustJSON(t, &request.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	})
	mockUC.On("Login", mock.Anything, mock.AnythingOfType("*request.LoginRequest")).
		Return(&response.AuthResponse{
			User:  &response.UserData{ID: 1, Username: "testuser", Email: "test@example.com"},
			Token: "jwt-token-456",
		}, nil)

	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.Login(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "success")
	mockUC.AssertExpectations(t)
}

func TestTDAuthHandler_Login_InvalidCredentials(t *testing.T) {
	t.Log("Login: invalid credentials -> 401")
	e := setupTDTestEcho()
	mockUC := new(mockTDAuthUseCase)
	h := &TDAuthHandler{uc: mockUC}

	body := tdMustJSON(t, &request.LoginRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	})
	mockUC.On("Login", mock.Anything, mock.AnythingOfType("*request.LoginRequest")).
		Return(nil, errors.New("invalid credentials"))

	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.Login(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	mockUC.AssertExpectations(t)
}
