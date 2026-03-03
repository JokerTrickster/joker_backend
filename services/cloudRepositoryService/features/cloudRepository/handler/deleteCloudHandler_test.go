package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockDeleteCloudUseCase struct {
	mock.Mock
}

func (m *mockDeleteCloudUseCase) DeleteFile(ctx context.Context, userID uint, fileID uint) error {
	args := m.Called(ctx, userID, fileID)
	return args.Error(0)
}

func TestDeleteCloudHandler_DeleteFile_Success(t *testing.T) {
	t.Log("Running: Success - DELETE /files/1 -> 200")
	e := setupTestEcho()
	mockUC := new(mockDeleteCloudUseCase)
	handler := &DeleteCloudRepositoryHandler{UseCase: mockUC}

	mockUC.On("DeleteFile", mock.Anything, testUserID, uint(1)).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/files/1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("1")
	setupAuthContext(c)

	err := handler.DeleteFile(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "file deleted successfully")
	mockUC.AssertExpectations(t)
}

func TestDeleteCloudHandler_DeleteFile_NoUserID(t *testing.T) {
	t.Log("Running: No userID -> 401")
	e := setupTestEcho()
	mockUC := new(mockDeleteCloudUseCase)
	handler := &DeleteCloudRepositoryHandler{UseCase: mockUC}

	req := httptest.NewRequest(http.MethodDelete, "/files/1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("1")

	err := handler.DeleteFile(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "unauthorized")
	mockUC.AssertNotCalled(t, "DeleteFile")
}

func TestDeleteCloudHandler_DeleteFile_InvalidFileID(t *testing.T) {
	t.Log("Running: Invalid file ID -> 400")
	e := setupTestEcho()
	mockUC := new(mockDeleteCloudUseCase)
	handler := &DeleteCloudRepositoryHandler{UseCase: mockUC}

	req := httptest.NewRequest(http.MethodDelete, "/files/xyz", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("xyz")
	setupAuthContext(c)

	err := handler.DeleteFile(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "invalid file ID")
	mockUC.AssertNotCalled(t, "DeleteFile")
}

func TestDeleteCloudHandler_DeleteFile_UseCaseError(t *testing.T) {
	t.Log("Running: UseCase error -> 500")
	e := setupTestEcho()
	mockUC := new(mockDeleteCloudUseCase)
	handler := &DeleteCloudRepositoryHandler{UseCase: mockUC}

	mockUC.On("DeleteFile", mock.Anything, testUserID, uint(999)).Return(assert.AnError)

	req := httptest.NewRequest(http.MethodDelete, "/files/999", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("999")
	setupAuthContext(c)

	err := handler.DeleteFile(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockUC.AssertExpectations(t)
}
