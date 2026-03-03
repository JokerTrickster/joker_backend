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

type mockListCloudUseCase struct {
	mock.Mock
}

func (m *mockListCloudUseCase) ListFiles(ctx context.Context, userID uint, req request.ListFilesRequestDTO) (*response.ListFilesResponseDTO, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.ListFilesResponseDTO), args.Error(1)
}

func TestListCloudHandler_ListFiles_Success(t *testing.T) {
	t.Log("Running: Success - GET /files -> 200")
	e := setupTestEcho()
	mockUC := new(mockListCloudUseCase)
	handler := &ListCloudRepositoryHandler{UseCase: mockUC}

	mockUC.On("ListFiles", mock.Anything, testUserID, mock.AnythingOfType("request.ListFilesRequestDTO")).
		Return(&response.ListFilesResponseDTO{
			Files:      []response.FileInfoDTO{},
			TotalCount: 0,
			Page:       1,
			PageSize:   20,
		}, nil)

	req := httptest.NewRequest(http.MethodGet, "/files", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setupAuthContext(c)

	err := handler.ListFiles(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockUC.AssertExpectations(t)
}

func TestListCloudHandler_ListFiles_NoUserID(t *testing.T) {
	t.Log("Running: No userID -> 401")
	e := setupTestEcho()
	mockUC := new(mockListCloudUseCase)
	handler := &ListCloudRepositoryHandler{UseCase: mockUC}

	req := httptest.NewRequest(http.MethodGet, "/files", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.ListFiles(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "unauthorized")
	mockUC.AssertNotCalled(t, "ListFiles")
}

func TestListCloudHandler_ListFiles_UseCaseError(t *testing.T) {
	t.Log("Running: UseCase error -> 500")
	e := setupTestEcho()
	mockUC := new(mockListCloudUseCase)
	handler := &ListCloudRepositoryHandler{UseCase: mockUC}

	mockUC.On("ListFiles", mock.Anything, testUserID, mock.AnythingOfType("request.ListFilesRequestDTO")).
		Return(nil, assert.AnError)

	req := httptest.NewRequest(http.MethodGet, "/files", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setupAuthContext(c)

	err := handler.ListFiles(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockUC.AssertExpectations(t)
}
