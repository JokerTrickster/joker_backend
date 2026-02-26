package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JokerTrickster/joker_backend/services/authService/features/auth/usecase"
	"github.com/JokerTrickster/joker_backend/services/authService/features/auth/handler/request"
	"github.com/JokerTrickster/joker_backend/services/authService/features/auth/handler/response"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockAuthUseCase is a mock implementation of auth use case
type MockAuthUseCase struct {
	mock.Mock
}

func (m *MockAuthUseCase) Signup(ctx context.Context, req *request.ReqSignUp) (*response.ResSignUp, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.ResSignUp), args.Error(1)
}

func (m *MockAuthUseCase) Signin(ctx context.Context, req *request.ReqSignIn) (*response.ResSignIn, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.ResSignIn), args.Error(1)
}

func TestAuthHandler_SignUp(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    interface{}
		setupMock      func(*MockAuthUseCase)
		expectedStatus int
		expectedBody   map[string]interface{}
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "successful signup",
			requestBody: map[string]interface{}{
				"email":        "test@example.com",
				"password":     "password123",
				"name":         "Test User",
				"service_type": "game",
			},
			setupMock: func(m *MockAuthUseCase) {
				m.On("Signup", mock.Anything, mock.MatchedBy(func(req *request.ReqSignUp) bool {
					return req.Email == "test@example.com" &&
						req.Password == "password123" &&
						req.Name == "Test User" &&
						req.ServiceType == "game"
				})).Return(&response.ResSignUp{
					UserID:       1,
					AccessToken:  "test-access-token",
					RefreshToken: "test-refresh-token",
				}, nil)
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var resp map[string]interface{}
				err := json.Unmarshal(rec.Body.Bytes(), &resp)
				require.NoError(t, err)

				data := resp["data"].(map[string]interface{})
				assert.Equal(t, float64(1), data["user_id"])
				assert.Equal(t, "test-access-token", data["access_token"])
				assert.Equal(t, "test-refresh-token", data["refresh_token"])
			},
		},
		{
			name: "invalid email format",
			requestBody: map[string]interface{}{
				"email":        "invalid-email",
				"password":     "password123",
				"name":         "Test User",
				"service_type": "game",
			},
			setupMock:      func(m *MockAuthUseCase) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var resp map[string]interface{}
				err := json.Unmarshal(rec.Body.Bytes(), &resp)
				require.NoError(t, err)
				assert.Contains(t, resp["error"], "validation")
			},
		},
		{
			name: "missing required fields",
			requestBody: map[string]interface{}{
				"email": "test@example.com",
			},
			setupMock:      func(m *MockAuthUseCase) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var resp map[string]interface{}
				err := json.Unmarshal(rec.Body.Bytes(), &resp)
				require.NoError(t, err)
				assert.Contains(t, resp["error"], "required")
			},
		},
		{
			name: "duplicate email",
			requestBody: map[string]interface{}{
				"email":        "existing@example.com",
				"password":     "password123",
				"name":         "Test User",
				"service_type": "game",
			},
			setupMock: func(m *MockAuthUseCase) {
				m.On("Signup", mock.Anything, mock.Anything).
					Return(nil, errors.New("email already exists"))
			},
			expectedStatus: http.StatusConflict,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var resp map[string]interface{}
				err := json.Unmarshal(rec.Body.Bytes(), &resp)
				require.NoError(t, err)
				assert.Contains(t, resp["error"], "already exists")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock
			mockUseCase := new(MockAuthUseCase)
			tt.setupMock(mockUseCase)

			// Create handler
			handler := NewAuthHandler(mockUseCase)

			// Create request
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/auth/signup", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			rec := httptest.NewRecorder()

			// Create gin context
			router := gin.New()
			router.POST("/auth/signup", handler.SignUp)

			// Serve request
			router.ServeHTTP(rec, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, rec.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, rec)
			}

			// Verify mock expectations
			mockUseCase.AssertExpectations(t)
		})
	}
}

func TestAuthHandler_SignIn(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    interface{}
		setupMock      func(*MockAuthUseCase)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "successful signin",
			requestBody: map[string]interface{}{
				"email":        "test@example.com",
				"password":     "password123",
				"service_type": "game",
			},
			setupMock: func(m *MockAuthUseCase) {
				m.On("Signin", mock.Anything, mock.MatchedBy(func(req *request.ReqSignIn) bool {
					return req.Email == "test@example.com" &&
						req.Password == "password123" &&
						req.ServiceType == "game"
				})).Return(&response.ResSignIn{
					UserID:       1,
					AccessToken:  "test-access-token",
					RefreshToken: "test-refresh-token",
				}, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var resp map[string]interface{}
				err := json.Unmarshal(rec.Body.Bytes(), &resp)
				require.NoError(t, err)

				data := resp["data"].(map[string]interface{})
				assert.Equal(t, float64(1), data["user_id"])
				assert.Equal(t, "test-access-token", data["access_token"])
				assert.Equal(t, "test-refresh-token", data["refresh_token"])
			},
		},
		{
			name: "wrong password",
			requestBody: map[string]interface{}{
				"email":        "test@example.com",
				"password":     "wrongpassword",
				"service_type": "game",
			},
			setupMock: func(m *MockAuthUseCase) {
				m.On("Signin", mock.Anything, mock.Anything).
					Return(nil, errors.New("invalid credentials"))
			},
			expectedStatus: http.StatusUnauthorized,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var resp map[string]interface{}
				err := json.Unmarshal(rec.Body.Bytes(), &resp)
				require.NoError(t, err)
				assert.Contains(t, resp["error"], "invalid")
			},
		},
		{
			name: "user not found",
			requestBody: map[string]interface{}{
				"email":        "nonexistent@example.com",
				"password":     "password123",
				"service_type": "game",
			},
			setupMock: func(m *MockAuthUseCase) {
				m.On("Signin", mock.Anything, mock.Anything).
					Return(nil, errors.New("user not found"))
			},
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var resp map[string]interface{}
				err := json.Unmarshal(rec.Body.Bytes(), &resp)
				require.NoError(t, err)
				assert.Contains(t, resp["error"], "not found")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock
			mockUseCase := new(MockAuthUseCase)
			tt.setupMock(mockUseCase)

			// Create handler
			handler := NewAuthHandler(mockUseCase)

			// Create request
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/auth/signin", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			rec := httptest.NewRecorder()

			// Create gin context
			router := gin.New()
			router.POST("/auth/signin", handler.SignIn)

			// Serve request
			router.ServeHTTP(rec, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, rec.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, rec)
			}

			// Verify mock expectations
			mockUseCase.AssertExpectations(t)
		})
	}
}

// Benchmark tests
func BenchmarkAuthHandler_SignUp(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)

	mockUseCase := new(MockAuthUseCase)
	mockUseCase.On("Signup", mock.Anything, mock.Anything).
		Return(&response.ResSignUp{
			UserID:       1,
			AccessToken:  "token",
			RefreshToken: "refresh",
		}, nil)

	handler := NewAuthHandler(mockUseCase)
	router := gin.New()
	router.POST("/auth/signup", handler.SignUp)

	reqBody := map[string]interface{}{
		"email":        "bench@example.com",
		"password":     "password123",
		"name":         "Bench User",
		"service_type": "game",
	}
	body, _ := json.Marshal(reqBody)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest(http.MethodPost, "/auth/signup", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
		}
	})
}