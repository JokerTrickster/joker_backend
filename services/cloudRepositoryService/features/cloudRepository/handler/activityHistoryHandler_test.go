package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/request"
	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/response"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockActivityHistoryUseCase struct {
	mock.Mock
}

func (m *mockActivityHistoryUseCase) GetActivityHistory(ctx context.Context, userID uint, req *request.ActivityHistoryRequestDTO) (*response.ActivityHistoryResponseDTO, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.ActivityHistoryResponseDTO), args.Error(1)
}

func TestActivityHistoryHandler_GetActivityHistory_Success(t *testing.T) {
	t.Log("Running: Success -> 200")
	e := setupTestEcho()
	mockUC := new(mockActivityHistoryUseCase)
	handler := &ActivityHistoryCloudRepositoryHandler{uc: mockUC}

	mockUC.On("GetActivityHistory", mock.Anything, testUserID, mock.AnythingOfType("*request.ActivityHistoryRequestDTO")).
		Return(&response.ActivityHistoryResponseDTO{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/user/activity", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setupAuthContext(c)

	err := handler.GetActivityHistory(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockUC.AssertExpectations(t)
}

func TestActivityHistoryHandler_GetActivityHistory_NoUserID(t *testing.T) {
	t.Log("Running: No userID -> 401")
	e := setupTestEcho()
	mockUC := new(mockActivityHistoryUseCase)
	handler := &ActivityHistoryCloudRepositoryHandler{uc: mockUC}

	req := httptest.NewRequest(http.MethodGet, "/user/activity", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.GetActivityHistory(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	mockUC.AssertNotCalled(t, "GetActivityHistory")
}

func TestActivityHistoryHandler_GetActivityHistory_UseCaseError(t *testing.T) {
	t.Log("Running: UseCase error -> 500")
	e := setupTestEcho()
	mockUC := new(mockActivityHistoryUseCase)
	handler := &ActivityHistoryCloudRepositoryHandler{uc: mockUC}

	mockUC.On("GetActivityHistory", mock.Anything, testUserID, mock.AnythingOfType("*request.ActivityHistoryRequestDTO")).
		Return(nil, assert.AnError)

	req := httptest.NewRequest(http.MethodGet, "/user/activity", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setupAuthContext(c)

	err := handler.GetActivityHistory(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockUC.AssertExpectations(t)
}

