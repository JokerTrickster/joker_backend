package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/response"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockUserStatsUseCase struct {
	mock.Mock
}

func (m *mockUserStatsUseCase) GetUserStats(ctx context.Context, userID uint) (*response.UserStatsResponseDTO, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.UserStatsResponseDTO), args.Error(1)
}

func TestUserStatsHandler_GetUserStats_Success(t *testing.T) {
	t.Log("Running: Success -> 200")
	e := setupTestEcho()
	mockUC := new(mockUserStatsUseCase)
	handler := &UserStatsCloudRepositoryHandler{uc: mockUC}

	mockUC.On("GetUserStats", mock.Anything, testUserID).
		Return(&response.UserStatsResponseDTO{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/user/stats", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setupAuthContext(c)

	err := handler.GetUserStats(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockUC.AssertExpectations(t)
}

func TestUserStatsHandler_GetUserStats_NoUserID(t *testing.T) {
	t.Log("Running: No userID -> 401")
	e := setupTestEcho()
	mockUC := new(mockUserStatsUseCase)
	handler := &UserStatsCloudRepositoryHandler{uc: mockUC}

	req := httptest.NewRequest(http.MethodGet, "/user/stats", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.GetUserStats(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	mockUC.AssertNotCalled(t, "GetUserStats")
}

func TestUserStatsHandler_GetUserStats_UseCaseError(t *testing.T) {
	t.Log("Running: UseCase error -> 500")
	e := setupTestEcho()
	mockUC := new(mockUserStatsUseCase)
	handler := &UserStatsCloudRepositoryHandler{uc: mockUC}

	mockUC.On("GetUserStats", mock.Anything, testUserID).
		Return(nil, assert.AnError)

	req := httptest.NewRequest(http.MethodGet, "/user/stats", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setupAuthContext(c)

	err := handler.GetUserStats(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockUC.AssertExpectations(t)
}
